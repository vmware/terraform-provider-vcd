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
	errorRetrievingOrgAndVdc     = "error retrieving Org and VDC: %s"
	errorRetrievingVdcFromOrg    = "error retrieving VDC %s from Org %s: %s"
	errorUnableToFindEdgeGateway = "unable to find edge gateway: %s"
	errorCompletingTask          = "error completing tasks: %s"
)

func debugPrintf(format string, args ...interface{}) {
	if os.Getenv("GOVCD_DEBUG") != "" {
		fmt.Printf(format, args...)
	}
}

// Cache values for vCD connection.
// When the Client() function is called with the same parameters, it will return
// a cached value instead of connecting again.
// This makes the Client() function both deterministic and fast.
type cachedConnection struct {
	initTime   time.Time
	connection *VCDClient
}

var cachedVCDClients = make(map[string]cachedConnection)
var cacheClientServedCount int = 0

// Invalidates the cache after a given time (connection tokens usually expire after 20 to 30 minutes)
var maxConnectionValidity time.Duration = 20 * time.Minute

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
	caller := filepath.Base(callFuncName())
	client, ok := cachedVCDClients[checksum]
	if ok && os.Getenv("VCD_CACHE") != "" {
		cacheClientServedCount += 1
		debugPrintf("[%s] cached connection served %d times (size:%d)\n",
			caller, cacheClientServedCount, len(cachedVCDClients))
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
	debugPrintf("[%s] %#v\n", caller, c)
	err = vcdclient.Authenticate(c.User, c.Password, c.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during authentication: %s", err)
	}
	cachedVCDClients[checksum] = cachedConnection{initTime: time.Now(), connection: vcdclient}

	return vcdclient, nil
}

// Returns the name of the function that called the
// current function.
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
