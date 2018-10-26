package vcd

import (
	"crypto/sha1"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/vmware/go-vcloud-director/govcd"
)

type Config struct {
	User            string
	Password        string
	SysOrg          string // Org used for authentication
	Org             string // Default Org used for API operations
	Vdc             string // Default (optional) VDC for API operations
	Href            string
	MaxRetryTimeout int
	InsecureFlag    bool
}

type VCDClient struct {
	*govcd.VCDClient
	SysOrg          string
	Org             string // name of default Org
	Vdc             string // name of default Vdc
	MaxRetryTimeout int
	InsecureFlag    bool
}

const (
	// Most common error messages in the library

	// Used when a call to GetOrgAndVdc fails. The placeholder is for the error
	errorRetrievingOrgAndVdc = "error retrieving Org and VDC: %s"

	// Used when a call to GetOrgAndVdc fails. The placeholders are for vdc, org, and the error
	errorRetrievingVdcFromOrg = "error retrieving VDC %s from Org %s: %s"

	// Used when we can't get a valid edge gateway. The placeholder is for the error
	errorUnableToFindEdgeGateway = "unable to find edge gateway: %s"

  // Used when a task fails. The placeholder is for the error
	errorCompletingTask = "error completing tasks: %s"
  
  // Used when a call to GetAdminOrgFromResource fails. The placeholder is for the error
  errorRetrievingOrg = "error retrieving Org: %s"
)

// Cache values for vCD connection.
// When the Client() function is called with the same parameters, it will return
// a cached value instead of connecting again.
// This makes the Client() function both deterministic and fast.
type cachedConnection struct {
	initTime   time.Time
	connection *VCDClient
}

var (
	// Enables the caching of authenticated connections
	enableConnectionCache bool = os.Getenv("VCD_CACHE") != ""

	// Cached VDC authenticated connection
	cachedVCDClients = make(map[string]cachedConnection)

	// Records how many times we have cached a connection
	cacheClientServedCount int = 0

	// Invalidates the cache after a given time (connection tokens usually expire after 20 to 30 minutes)
	maxConnectionValidity time.Duration = 20 * time.Minute

	enableDebug bool = os.Getenv("GOVCD_DEBUG") != ""
	enableTrace bool = os.Getenv("GOVCD_TRACE") != ""
)

// Displays conditional messages
func debugPrintf(format string, args ...interface{}) {
	// When GOVCD_TRACE is enabled, we also display the function that generated the message
	if enableTrace {
		format = fmt.Sprintf("[%s] %s", filepath.Base(callFuncName()), format)
	}
	// The formatted message passed to this function is displayed only when GOVCD_DEBUG is enabled.
	if enableDebug {
		fmt.Printf(format, args...)
	}
}

// GetOrgAndVdc finds a pair of org and vdc using the names provided
// in the args. If the names are empty, it will use the default
// org and vdc names from the provider.
func (cli *VCDClient) GetOrgAndVdc(orgName, vdcName string) (org govcd.Org, vdc govcd.Vdc, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return govcd.Org{}, govcd.Vdc{}, fmt.Errorf("empty Org name provided")
	}
	if vdcName == "" {
		vdcName = cli.Vdc
	}
	if vdcName == "" {
		return govcd.Org{}, govcd.Vdc{}, fmt.Errorf("empty VDC name provided")
	}
	org, err = govcd.GetOrgByName(cli.VCDClient, orgName)
	if err != nil {
		return govcd.Org{}, govcd.Vdc{}, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	vdc, err = org.GetVdcByName(vdcName)
	if err != nil {
		return govcd.Org{}, govcd.Vdc{}, fmt.Errorf("error retrieving VDC %s: %s", vdcName, err)
	}
	return org, vdc, err
}

// GetAdminOrg finds org using the names provided in the args.
// If the name is empty, it will use the default
// org name from the provider.
func (cli *VCDClient) GetAdminOrg(orgName string) (org govcd.AdminOrg, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return govcd.AdminOrg{}, fmt.Errorf("empty Org name provided")
	}

	org, err = govcd.GetAdminOrgByName(cli.VCDClient, orgName)
	if err != nil {
		return govcd.AdminOrg{}, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	return org, err
}

// Same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetOrgAndVdcFromResource(d *schema.ResourceData) (org govcd.Org, vdc govcd.Vdc, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	return cli.GetOrgAndVdc(orgName, vdcName)
}

// Same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetAdminOrgFromResource(d *schema.ResourceData) (org govcd.AdminOrg, err error) {
	orgName := d.Get("org").(string)
	return cli.GetAdminOrg(orgName)
}

// Gets an edge gateway when you don't need org or vdc for other purposes
func (cli *VCDClient) GetEdgeGateway(orgName, vdcName, edgeGwName string) (eg govcd.EdgeGateway, err error) {

	if edgeGwName == "" {
		return govcd.EdgeGateway{}, fmt.Errorf("empty Edge Gateway name provided")
	}
	_, vdc, err := cli.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return govcd.EdgeGateway{}, fmt.Errorf("error retrieving org and vdc: %s", err)
	}
	eg, err = vdc.FindEdgeGateway(edgeGwName)

	if err != nil {
		return govcd.EdgeGateway{}, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
	return eg, err
}

// Same as GetEdgeGateway, but using data from the resource, if available
func (cli *VCDClient) GetEdgeGatewayFromResource(d *schema.ResourceData) (eg govcd.EdgeGateway, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayName := d.Get("edge_gateway").(string)
	return cli.GetEdgeGateway(orgName, vdcName, edgeGatewayName)

}

func (c *Config) Client() (*VCDClient, error) {

	rawData := c.User + "#" +
		c.Password + "#" +
		c.SysOrg + "#" +
		c.Href
	checksum := fmt.Sprintf("%x", sha1.Sum([]byte(rawData)))

	// The cached connection is served only if the variable VCD_CACHE is set
	client, ok := cachedVCDClients[checksum]
	if ok && enableConnectionCache {
		cacheClientServedCount += 1
		// debugPrintf("[%s] cached connection served %d times (size:%d)\n",
		elapsed := time.Since(client.initTime)
		if elapsed > maxConnectionValidity {
			debugPrintf("cached connection invalidated after %2.0f minutes \n", maxConnectionValidity.Minutes())
			delete(cachedVCDClients, checksum)
		} else {
			return client.connection, nil
		}
	}

	authUrl, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while retrieving URL: %s", err)
	}

	vcdclient := &VCDClient{
		VCDClient:       govcd.NewVCDClient(*authUrl, c.InsecureFlag),
		SysOrg:          c.SysOrg,
		Org:             c.Org,
		Vdc:             c.Vdc,
		MaxRetryTimeout: c.MaxRetryTimeout,
		InsecureFlag:    c.InsecureFlag}
	err = vcdclient.Authenticate(c.User, c.Password, c.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during authentication: %s", err)
	}
	cachedVCDClients[checksum] = cachedConnection{initTime: time.Now(), connection: vcdclient}

	return vcdclient, nil
}

// Returns the name of the function that called the
// current function.
// It is used for tracing
func callFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}
