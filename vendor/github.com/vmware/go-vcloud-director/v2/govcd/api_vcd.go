/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VCDClientOption defines signature for customizing VCDClient using
// functional options pattern.
type VCDClientOption func(*VCDClient) error

type VCDClient struct {
	Client      Client  // Client for the underlying VCD instance
	sessionHREF url.URL // HREF for the session API
	QueryHREF   url.URL // HREF for the query API
}

func (vcdCli *VCDClient) vcdloginurl() error {
	if err := vcdCli.Client.validateAPIVersion(); err != nil {
		return fmt.Errorf("could not find valid version for login: %s", err)
	}

	// find login address matching the API version
	var neededVersion VersionInfo
	for _, versionInfo := range vcdCli.Client.supportedVersions.VersionInfos {
		if versionInfo.Version == vcdCli.Client.APIVersion {
			neededVersion = versionInfo
			break
		}
	}

	loginUrl, err := url.Parse(neededVersion.LoginUrl)
	if err != nil {
		return fmt.Errorf("couldn't find a LoginUrl for version %s", vcdCli.Client.APIVersion)
	}
	vcdCli.sessionHREF = *loginUrl
	return nil
}

// vcdAuthorize authorizes the client and returns a http response
func (vcdCli *VCDClient) vcdAuthorize(user, pass, org string) (*http.Response, error) {
	var missingItems []string
	if user == "" {
		missingItems = append(missingItems, "user")
	}
	if pass == "" {
		missingItems = append(missingItems, "password")
	}
	if org == "" {
		missingItems = append(missingItems, "org")
	}
	if len(missingItems) > 0 {
		return nil, fmt.Errorf("authorization is not possible because of these missing items: %v", missingItems)
	}
	// No point in checking for errors here
	req := vcdCli.Client.NewRequest(map[string]string{}, http.MethodPost, vcdCli.sessionHREF, nil)
	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version="+vcdCli.Client.APIVersion)
	resp, err := checkResp(vcdCli.Client.Http.Do(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Store the authorization header
	vcdCli.Client.VCDToken = resp.Header.Get(AuthorizationHeader)
	vcdCli.Client.VCDAuthHeader = AuthorizationHeader
	vcdCli.Client.IsSysAdmin = strings.EqualFold(org, "system")
	// Get query href
	vcdCli.QueryHREF = vcdCli.Client.VCDHREF
	vcdCli.QueryHREF.Path += "/query"
	return resp, nil
}

// NewVCDClient initializes VMware vCloud Director client with reasonable defaults.
// It accepts functions of type VCDClientOption for adjusting defaults.
func NewVCDClient(vcdEndpoint url.URL, insecure bool, options ...VCDClientOption) *VCDClient {
	// Setting defaults
	vcdClient := &VCDClient{
		Client: Client{
			APIVersion: "31.0", // supported by 9.5, 9.7, 10.0, 10.1
			VCDHREF:    vcdEndpoint,
			Http: http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
					},
					Proxy:               http.ProxyFromEnvironment,
					TLSHandshakeTimeout: 120 * time.Second, // Default timeout for TSL hand shake
				},
				Timeout: 600 * time.Second, // Default value for http request+response timeout
			},
			MaxRetryTimeout: 60, // Default timeout in seconds for retries calls in functions
		},
	}

	// Override defaults with functional options
	for _, option := range options {
		err := option(vcdClient)
		if err != nil {
			// We do not have error in return of this function signature.
			// To avoid breaking API the only thing we can do is panic.
			panic(fmt.Sprintf("unable to initialize vCD client: %s", err))
		}
	}
	return vcdClient
}

// Authenticate is a helper function that performs a login in vCloud Director.
func (vcdCli *VCDClient) Authenticate(username, password, org string) error {
	_, err := vcdCli.GetAuthResponse(username, password, org)
	return err
}

// GetAuthResponse performs authentication and returns the full HTTP response
// The purpose of this function is to preserve information that is useful
// for token-based authentication
func (vcdCli *VCDClient) GetAuthResponse(username, password, org string) (*http.Response, error) {
	// LoginUrl
	err := vcdCli.vcdloginurl()
	if err != nil {
		return nil, fmt.Errorf("error finding LoginUrl: %s", err)
	}

	// Choose correct auth mechanism based on what type of authentication is used. The end result
	// for each of the below functions is to set authorization token vcdCli.Client.VCDToken.
	var resp *http.Response
	switch {
	case vcdCli.Client.UseSamlAdfs:
		err = vcdCli.authorizeSamlAdfs(username, password, org, vcdCli.Client.CustomAdfsRptId)
		if err != nil {
			return nil, fmt.Errorf("error authorizing SAML: %s", err)
		}
	default:
		// Authorize
		resp, err = vcdCli.vcdAuthorize(username, password, org)
		if err != nil {
			return nil, fmt.Errorf("error authorizing: %s", err)
		}
	}

	return resp, nil
}

// SetToken will set the authorization token in the client, without using other credentials
// Up to version 29, token authorization uses the the header key x-vcloud-authorization
// In version 30+ it also uses X-Vmware-Vcloud-Access-Token:TOKEN coupled with
// X-Vmware-Vcloud-Token-Type:"bearer"
// TODO: when enabling version 30+ for SDK, add ability of using bearer token
func (vcdCli *VCDClient) SetToken(org, authHeader, token string) error {
	vcdCli.Client.VCDAuthHeader = authHeader
	vcdCli.Client.VCDToken = token

	err := vcdCli.vcdloginurl()
	if err != nil {
		return fmt.Errorf("error finding LoginUrl: %s", err)
	}

	vcdCli.Client.IsSysAdmin = strings.EqualFold(org, "system")
	// Get query href
	vcdCli.QueryHREF = vcdCli.Client.VCDHREF
	vcdCli.QueryHREF.Path += "/query"

	// The client is now ready to connect using the token, but has not communicated with the vCD yet.
	// To make sure that it is working, we run a request for the org list.
	// This list should work always: when run as system administrator, it retrieves all organizations.
	// When run as org user, it only returns the organization the user is authorized to.
	// In both cases, we discard the list, as we only use it to certify that the token works.
	orgListHREF := vcdCli.Client.VCDHREF
	orgListHREF.Path += "/org"

	orgList := new(types.OrgList)

	_, err = vcdCli.Client.ExecuteRequest(orgListHREF.String(), http.MethodGet,
		"", "error connecting to vCD using token: %s", nil, orgList)
	if err != nil {
		return err
	}
	return nil
}

// Disconnect performs a disconnection from the vCloud Director API endpoint.
func (vcdCli *VCDClient) Disconnect() error {
	if vcdCli.Client.VCDToken == "" && vcdCli.Client.VCDAuthHeader == "" {
		return fmt.Errorf("cannot disconnect, client is not authenticated")
	}
	req := vcdCli.Client.NewRequest(map[string]string{}, http.MethodDelete, vcdCli.sessionHREF, nil)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version="+vcdCli.Client.APIVersion)
	// Set Authorization Header
	req.Header.Add(vcdCli.Client.VCDAuthHeader, vcdCli.Client.VCDToken)
	if _, err := checkResp(vcdCli.Client.Http.Do(req)); err != nil {
		return fmt.Errorf("error processing session delete for vCloud Director: %s", err)
	}
	return nil
}

// WithMaxRetryTimeout allows default vCDClient MaxRetryTimeout value override
func WithMaxRetryTimeout(timeoutSeconds int) VCDClientOption {
	return func(vcdClient *VCDClient) error {
		vcdClient.Client.MaxRetryTimeout = timeoutSeconds
		return nil
	}
}

// WithAPIVersion allows to override default API version. Please be cautious
// about changing the version as the default specified is the most tested.
func WithAPIVersion(version string) VCDClientOption {
	return func(vcdClient *VCDClient) error {
		vcdClient.Client.APIVersion = version
		return nil
	}
}

// WithHttpTimeout allows to override default http timeout
func WithHttpTimeout(timeout int64) VCDClientOption {
	return func(vcdClient *VCDClient) error {
		vcdClient.Client.Http.Timeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithSamlAdfs specifies if SAML auth is used for authenticating to vCD instead of local login.
// The following conditions must be met so that SAML authentication works:
// * SAML IdP (Identity Provider) is Active Directory Federation Service (ADFS)
// * WS-Trust authentication endpoint "/adfs/services/trust/13/usernamemixed" must be enabled on
// ADFS server
// By default vCD SAML Entity ID will be used as Relaying Party Trust Identifier unless
// customAdfsRptId is specified
func WithSamlAdfs(useSaml bool, customAdfsRptId string) VCDClientOption {
	return func(vcdClient *VCDClient) error {
		vcdClient.Client.UseSamlAdfs = useSaml
		vcdClient.Client.CustomAdfsRptId = customAdfsRptId
		return nil
	}
}
