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
	"sync"
	"time"
)

// VCDClientOption defines signature for customizing VCDClient using
// functional options pattern.
type VCDClientOption func(*VCDClient) error

type VCDClient struct {
	Client            Client  // Client for the underlying VCD instance
	sessionHREF       url.URL // HREF for the session API
	QueryHREF         url.URL // HREF for the query API
	Mutex             sync.Mutex
	supportedVersions SupportedVersions // Versions from /api/versions endpoint
}

func (vcdCli *VCDClient) vcdloginurl() error {
	if err := vcdCli.validateAPIVersion(); err != nil {
		return fmt.Errorf("could not find valid version for login: %s", err)
	}

	// find login address matching the API version
	var neededVersion VersionInfo
	for _, versionInfo := range vcdCli.supportedVersions.VersionInfos {
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

func (vcdCli *VCDClient) vcdauthorize(user, pass, org string) error {
	var missing_items []string
	if user == "" {
		missing_items = append(missing_items, "user")
	}
	if pass == "" {
		missing_items = append(missing_items, "password")
	}
	if org == "" {
		missing_items = append(missing_items, "org")
	}
	if len(missing_items) > 0 {
		return fmt.Errorf("authorization is not possible because of these missing items: %v", missing_items)
	}
	// No point in checking for errors here
	req := vcdCli.Client.NewRequest(map[string]string{}, http.MethodPost, vcdCli.sessionHREF, nil)
	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version="+vcdCli.Client.APIVersion)
	resp, err := checkResp(vcdCli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Store the authentication header
	vcdCli.Client.VCDToken = resp.Header.Get("x-vcloud-authorization")
	vcdCli.Client.VCDAuthHeader = "x-vcloud-authorization"
	vcdCli.Client.IsSysAdmin = false
	if strings.ToLower(org) == "system" {
		vcdCli.Client.IsSysAdmin = true
	}
	// Get query href
	vcdCli.QueryHREF = vcdCli.Client.VCDHREF
	vcdCli.QueryHREF.Path += "/query"
	return nil
}

// NewVCDClient initializes VMware vCloud Director client with reasonable defaults.
// It accepts functions of type VCDClientOption for adjusting defaults.
func NewVCDClient(vcdEndpoint url.URL, insecure bool, options ...VCDClientOption) *VCDClient {
	// Setting defaults
	vcdClient := &VCDClient{
		Client: Client{
			APIVersion: "27.0", // supported by vCD 8.20, 9.0, 9.1, 9.5, 9.7
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

// Authenticate is an helper function that performs a login in vCloud Director.
func (vcdCli *VCDClient) Authenticate(username, password, org string) error {

	// LoginUrl
	err := vcdCli.vcdloginurl()
	if err != nil {
		return fmt.Errorf("error finding LoginUrl: %s", err)
	}
	// Authorize
	err = vcdCli.vcdauthorize(username, password, org)
	if err != nil {
		return fmt.Errorf("error authorizing: %s", err)
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
