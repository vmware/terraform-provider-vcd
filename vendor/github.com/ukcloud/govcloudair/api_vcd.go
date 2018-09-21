package govcloudair

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	//types "github.com/ukcloud/govcloudair/types/v56"
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

func (c *VCDClient) vcdloginurl() error {

	s := c.Client.HREF
	s.Path += "/versions"

	// No point in checking for errors here
	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	supportedVersions := new(supportedVersions)

	err = decodeBody(resp, supportedVersions)

	if err != nil {
		return fmt.Errorf("error decoding versions response: %s", err)
	}

	u, err := url.Parse(supportedVersions.VersionInfo.LoginUrl)
	if err != nil {
		return fmt.Errorf("couldn't find a LoginUrl in versions")
	}
	c.sessionHREF = *u
	return nil
}

func (c *VCDClient) vcdauthorize(user, pass, org string) error {

	if user == "" {
		user = os.Getenv("VCLOUD_USERNAME")
	}

	if pass == "" {
		pass = os.Getenv("VCLOUD_PASSWORD")
	}

	if org == "" {
		org = os.Getenv("VCLOUD_ORG")
	}

	req := c.Client.NewRequest(map[string]string{}, "POST", c.sessionHREF, nil)

	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version=5.5")

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.Client.VCDToken = resp.Header.Get("x-vcloud-authorization")
	c.Client.VCDAuthHeader = "x-vcloud-authorization"

	u := c.Client.HREF
	u.Path += "/query"
	c.QueryHREF = u

	return nil
}

func NewVCDClient(vcdEndpoint url.URL, insecure bool) *VCDClient {

	return &VCDClient{
		Client: Client{
			APIVersion: "5.5",
			HREF:       vcdEndpoint,
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
func (c *VCDClient) Authenticate(username, password, org string) error {

	// LoginUrl
	err := c.vcdloginurl()
	if err != nil {
		return fmt.Errorf("error finding LoginUrl: %s", err)
	}
	// Authorize
	err = c.vcdauthorize(username, password, org)
	if err != nil {
		return fmt.Errorf("error authorizing: %s", err)
	}

	return nil
}

// Disconnect performs a disconnection from the vCloud Director API endpoint.
func (c *VCDClient) Disconnect() error {
	if c.Client.VCDToken == "" && c.Client.VCDAuthHeader == "" {
		return fmt.Errorf("cannot disconnect, client is not authenticated")
	}

	req := c.Client.NewRequest(map[string]string{}, "DELETE", c.sessionHREF, nil)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.5")

	// Set Authorization Header
	req.Header.Add(c.Client.VCDAuthHeader, c.Client.VCDToken)

	if _, err := checkResp(c.Client.Http.Do(req)); err != nil {
		return fmt.Errorf("error processing session delete for vCloud Director: %s", err)
	}
	return nil
}
