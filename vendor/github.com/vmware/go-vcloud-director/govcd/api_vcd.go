/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type VCDClient struct {
	Client      Client  // Client for the underlying VCD instance
	sessionHREF url.URL // HREF for the session API
	QueryHREF   url.URL // HREF for the query API
	Mutex       sync.Mutex
}

type supportedVersions struct {
	VersionInfo struct {
		Version  string `xml:"Version"`
		LoginUrl string `xml:"LoginUrl"`
	} `xml:"VersionInfo"`
}

func (vdcCli *VCDClient) vcdloginurl() error {
	apiEndpoint := vdcCli.Client.VCDHREF
	apiEndpoint.Path += "/versions"
	// No point in checking for errors here
	req := vdcCli.Client.NewRequest(map[string]string{}, "GET", apiEndpoint, nil)
	resp, err := checkResp(vdcCli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	supportedVersions := new(supportedVersions)
	err = decodeBody(resp, supportedVersions)
	if err != nil {
		return fmt.Errorf("error decoding versions response: %s", err)
	}
	loginUrl, err := url.Parse(supportedVersions.VersionInfo.LoginUrl)
	if err != nil {
		return fmt.Errorf("couldn't find a LoginUrl in versions")
	}
	vdcCli.sessionHREF = *loginUrl
	return nil
}

func (vdcCli *VCDClient) vcdauthorize(user, pass, org string) error {
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
		return fmt.Errorf("Authorization is not possible because of these missing items: %v", missing_items)
	}
	// No point in checking for errors here
	req := vdcCli.Client.NewRequest(map[string]string{}, "POST", vdcCli.sessionHREF, nil)
	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version="+vdcCli.Client.APIVersion)
	resp, err := checkResp(vdcCli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Store the authentication header
	vdcCli.Client.VCDToken = resp.Header.Get("x-vcloud-authorization")
	vdcCli.Client.VCDAuthHeader = "x-vcloud-authorization"
	vdcCli.Client.IsSysAdmin = false
	if "System" == org {
		vdcCli.Client.IsSysAdmin = true
	}
	// Get query href
	vdcCli.QueryHREF = vdcCli.Client.VCDHREF
	vdcCli.QueryHREF.Path += "/query"
	return nil
}

func NewVCDClient(vcdEndpoint url.URL, insecure bool) *VCDClient {

	return &VCDClient{
		Client: Client{
			APIVersion: "27.0", // supported by vCD 8.20, 9.0, 9.1, 9.5
			VCDHREF:    vcdEndpoint,
			Http: http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
					},
					Proxy:               http.ProxyFromEnvironment,
					TLSHandshakeTimeout: 120 * time.Second,
				},
			},
		},
	}
}

// Authenticate is an helper function that performs a login in vCloud Director.
func (vdcCli *VCDClient) Authenticate(username, password, org string) error {
	// LoginUrl
	err := vdcCli.vcdloginurl()
	if err != nil {
		return fmt.Errorf("error finding LoginUrl: %s", err)
	}
	// Authorize
	err = vdcCli.vcdauthorize(username, password, org)
	if err != nil {
		return fmt.Errorf("error authorizing: %s", err)
	}
	return nil
}

// Disconnect performs a disconnection from the vCloud Director API endpoint.
func (vdcCli *VCDClient) Disconnect() error {
	if vdcCli.Client.VCDToken == "" && vdcCli.Client.VCDAuthHeader == "" {
		return fmt.Errorf("cannot disconnect, client is not authenticated")
	}
	req := vdcCli.Client.NewRequest(map[string]string{}, "DELETE", vdcCli.sessionHREF, nil)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version="+vdcCli.Client.APIVersion)
	// Set Authorization Header
	req.Header.Add(vdcCli.Client.VCDAuthHeader, vdcCli.Client.VCDToken)
	if _, err := checkResp(vdcCli.Client.Http.Do(req)); err != nil {
		return fmt.Errorf("error processing session delete for vCloud Director: %s", err)
	}
	return nil
}
