package vcd

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type Config struct {
	User            string
	Password        string
	Token           string // Token used instead of user and password
	SysOrg          string // Org used for authentication
	Org             string // Default Org used for API operations
	Vdc             string // Default (optional) VDC for API operations
	Href            string
	MaxRetryTimeout int
	InsecureFlag    bool

	// UseSamlAdfs specifies if SAML auth is used for authenticating vCD instead of local login.
	// The following conditions must be met so that authentication SAML authentication works:
	// * SAML IdP (Identity Provider) is Active Directory Federation Service (ADFS)
	// * Authentication endpoint "/adfs/services/trust/13/usernamemixed" must be enabled on ADFS
	// server
	UseSamlAdfs bool
	// CustomAdfsRptId allows to set custom Relaying Party Trust identifier. By default vCD Entity
	// ID is used as Relaying Party Trust identifier.
	CustomAdfsRptId string
}

type VCDClient struct {
	*govcd.VCDClient
	SysOrg          string
	Org             string // name of default Org
	Vdc             string // name of default VDC
	MaxRetryTimeout int
	InsecureFlag    bool
}

// Type used to simplify reading resource definitions
type StringMap map[string]interface{}

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

type cacheStorage struct {
	// conMap holds cached VDC authenticated connection
	conMap map[string]cachedConnection
	// cacheClientServedCount records how many times we have cached a connection
	cacheClientServedCount int
	sync.Mutex
}

// reset clears connection cache so that next session is forced to re-authenticate
func (c *cacheStorage) reset() {
	c.Lock()
	defer c.Unlock()
	c.cacheClientServedCount = 0
	c.conMap = make(map[string]cachedConnection)
}

var (
	// Enables the caching of authenticated connections
	enableConnectionCache bool = os.Getenv("VCD_CACHE") != ""

	// Cached VDC authenticated connection
	cachedVCDClients = &cacheStorage{conMap: make(map[string]cachedConnection)}

	// Invalidates the cache after a given time (connection tokens usually expire after 20 to 30 minutes)
	maxConnectionValidity time.Duration = 20 * time.Minute

	enableDebug bool = os.Getenv("GOVCD_DEBUG") != ""
	enableTrace bool = os.Getenv("GOVCD_TRACE") != ""

	// Separation string used for import operations
	// Can be changed usin either "import_separator" property in Provider
	// or environment variable "VCD_IMPORT_SEPARATOR"
	ImportSeparator = "."
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

// This is a global MutexKV for all resources
var vcdMutexKV = mutexkv.NewMutexKV()

func (cli *VCDClient) lockVapp(d *schema.ResourceData) {
	vappName := d.Get("name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Lock(key)
}

func (cli *VCDClient) unLockVapp(d *schema.ResourceData) {
	vappName := d.Get("name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Unlock(key)
}

// locks an edge gateway resource
// Differs from lockParentEdgeGtw in the resource name. When EGW is the parent,
// it's named "edge_gateway". When it's the main resource, it's found at "name"
func (cli *VCDClient) lockEdgeGateway(d *schema.ResourceData) {
	edgeGatewayName := d.Get("name").(string)
	if edgeGatewayName == "" {
		panic("edge gateway name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|edge:%s", cli.getOrgName(d), cli.getVdcName(d), edgeGatewayName)
	vcdMutexKV.Lock(key)
}

// unlocks an edge gateway resource
// Differs from unlockParentEdgeGtw in the resource name. When EGW is the parent,
// it's named "edge_gateway". When it's the main resource, it's found at "name"
func (cli *VCDClient) unlockEdgeGateway(d *schema.ResourceData) {
	edgeGatewayName := d.Get("name").(string)
	if edgeGatewayName == "" {
		panic("edge gateway name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|edge:%s", cli.getOrgName(d), cli.getVdcName(d), edgeGatewayName)
	vcdMutexKV.Unlock(key)
}

// lockParentVappWithName locks using provided vappName.
// Parent means the resource belongs to the vApp being locked
func (cli *VCDClient) lockParentVappWithName(d *schema.ResourceData, vappName string) {
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Lock(key)
}

func (cli *VCDClient) unLockParentVappWithName(d *schema.ResourceData, vappName string) {
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Unlock(key)
}

// function lockParentVapp locks using vapp_name name existing in resource parameters.
// Parent means the resource belongs to the vApp being locked
func (cli *VCDClient) lockParentVapp(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Lock(key)
}

func (cli *VCDClient) unLockParentVapp(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.Unlock(key)
}

// lockParentVm locks using vapp_name and vm_name names existing in resource parameters.
// Parent means the resource belongs to the VM being locked
func (cli *VCDClient) lockParentVm(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	vmName := d.Get("vm_name").(string)
	if vmName == "" {
		panic("vmName name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s|vm:%s", cli.getOrgName(d), cli.getVdcName(d), vappName, vmName)
	vcdMutexKV.Lock(key)
}

func (cli *VCDClient) unLockParentVm(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	vmName := d.Get("vm_name").(string)
	if vmName == "" {
		panic("vmName name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s|vm:%s", cli.getOrgName(d), cli.getVdcName(d), vappName, vmName)
	vcdMutexKV.Unlock(key)
}

// function lockParentEdgeGtw locks using edge_gateway name existing in resource parameters.
// Parent means the resource belongs to the edge gateway being locked
func (cli *VCDClient) lockParentEdgeGtw(d *schema.ResourceData) {
	edgeGtwName := d.Get("edge_gateway").(string)
	if edgeGtwName == "" {
		panic("edge gateway not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|edge:%s", cli.getOrgName(d), cli.getVdcName(d), edgeGtwName)
	vcdMutexKV.Lock(key)
}

func (cli *VCDClient) unLockParentEdgeGtw(d *schema.ResourceData) {
	edgeGtwName := d.Get("edge_gateway").(string)
	if edgeGtwName == "" {
		panic("edge gateway not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|edge:%s", cli.getOrgName(d), cli.getVdcName(d), edgeGtwName)
	vcdMutexKV.Unlock(key)
}

func (cli *VCDClient) getOrgName(d *schema.ResourceData) string {
	orgName := d.Get("org").(string)
	if orgName == "" {
		orgName = cli.Org
	}
	return orgName
}

func (cli *VCDClient) getVdcName(d *schema.ResourceData) string {
	orgName := d.Get("vdc").(string)
	if orgName == "" {
		orgName = cli.Vdc
	}
	return orgName
}

// GetOrgAndVdc finds a pair of org and vdc using the names provided
// in the args. If the names are empty, it will use the default
// org and vdc names from the provider.
func (cli *VCDClient) GetOrgAndVdc(orgName, vdcName string) (org *govcd.Org, vdc *govcd.Vdc, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return nil, nil, fmt.Errorf("empty Org name provided")
	}
	if vdcName == "" {
		vdcName = cli.Vdc
	}
	if vdcName == "" {
		return nil, nil, fmt.Errorf("empty VDC name provided")
	}
	org, err = cli.VCDClient.GetOrgByName(orgName)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	if org.Org.Name == "" || org.Org.HREF == "" || org.Org.ID == "" {
		return nil, nil, fmt.Errorf("empty Org %s found ", orgName)
	}
	vdc, err = org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving VDC %s: %s", vdcName, err)
	}
	if vdc == nil || vdc.Vdc.ID == "" || vdc.Vdc.HREF == "" || vdc.Vdc.Name == "" {
		return nil, nil, fmt.Errorf("error retrieving VDC %s: not found", vdcName)
	}
	return org, vdc, err
}

// GetAdminOrg finds org using the names provided in the args.
// If the name is empty, it will use the default
// org name from the provider.
func (cli *VCDClient) GetAdminOrg(orgName string) (org *govcd.AdminOrg, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return nil, fmt.Errorf("empty Org name provided")
	}

	org, err = cli.VCDClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	if org.AdminOrg.Name == "" || org.AdminOrg.HREF == "" || org.AdminOrg.ID == "" {
		return nil, fmt.Errorf("empty org %s found", orgName)
	}
	return org, err
}

// Same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetOrgAndVdcFromResource(d *schema.ResourceData) (org *govcd.Org, vdc *govcd.Vdc, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	return cli.GetOrgAndVdc(orgName, vdcName)
}

// Same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetAdminOrgFromResource(d *schema.ResourceData) (org *govcd.AdminOrg, err error) {
	orgName := d.Get("org").(string)
	return cli.GetAdminOrg(orgName)
}

// Gets an edge gateway when you don't need org or vdc for other purposes
func (cli *VCDClient) GetEdgeGateway(orgName, vdcName, edgeGwName string) (eg *govcd.EdgeGateway, err error) {

	if edgeGwName == "" {
		return nil, fmt.Errorf("empty Edge Gateway name provided")
	}
	_, vdc, err := cli.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving org and vdc: %s", err)
	}
	eg, err = vdc.GetEdgeGatewayByName(edgeGwName, true)

	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf(fmt.Sprintf("(%s) [%s] ", edgeGwName, callFuncName())+errorUnableToFindEdgeGateway, err)
		}
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
	return eg, nil
}

// Same as GetEdgeGateway, but using data from the resource, if available
// edgeGatewayFieldName is the name used in the resource. It is usually "edge_gateway"
// for all resources that *use* an edge gateway, and when the resource is vcd_edgegateway, it is "name"
func (cli *VCDClient) GetEdgeGatewayFromResource(d *schema.ResourceData, edgeGatewayFieldName string) (eg *govcd.EdgeGateway, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayName := d.Get(edgeGatewayFieldName).(string)
	egw, err := cli.GetEdgeGateway(orgName, vdcName, edgeGatewayName)
	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf("(%s) [%s] : %s", edgeGatewayName, callFuncName(), err)
		}
		return nil, err
	}
	return egw, nil
}

func ProviderAuthenticate(client *govcd.VCDClient, user, password, token, org string) error {
	var err error
	if token != "" {
		err = client.SetToken(org, govcd.AuthorizationHeader, token)
		if err != nil {
			err = fmt.Errorf("error during token-based authentication: %s", err)
		}
	} else {
		err = client.Authenticate(user, password, org)
	}
	return err
}

func (c *Config) Client() (*VCDClient, error) {
	rawData := c.User + "#" +
		c.Password + "#" +
		c.Token + "#" +
		c.SysOrg + "#" +
		c.Href
	checksum := fmt.Sprintf("%x", sha1.Sum([]byte(rawData)))

	// The cached connection is served only if the variable VCD_CACHE is set
	cachedVCDClients.Lock()
	client, ok := cachedVCDClients.conMap[checksum]
	cachedVCDClients.Unlock()
	if ok && enableConnectionCache {
		cachedVCDClients.Lock()
		cachedVCDClients.cacheClientServedCount += 1
		cachedVCDClients.Unlock()
		// debugPrintf("[%s] cached connection served %d times (size:%d)\n",
		elapsed := time.Since(client.initTime)
		if elapsed > maxConnectionValidity {
			debugPrintf("cached connection invalidated after %2.0f minutes \n", maxConnectionValidity.Minutes())
			cachedVCDClients.Lock()
			delete(cachedVCDClients.conMap, checksum)
			cachedVCDClients.Unlock()
		} else {
			return client.connection, nil
		}
	}

	authUrl, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while retrieving URL: %s", err)
	}

	vcdClient := &VCDClient{
		VCDClient: govcd.NewVCDClient(*authUrl, c.InsecureFlag,
			govcd.WithMaxRetryTimeout(c.MaxRetryTimeout),
			govcd.WithSamlAdfs(c.UseSamlAdfs, c.CustomAdfsRptId)),
		SysOrg:          c.SysOrg,
		Org:             c.Org,
		Vdc:             c.Vdc,
		MaxRetryTimeout: c.MaxRetryTimeout,
		InsecureFlag:    c.InsecureFlag}

	err = ProviderAuthenticate(vcdClient.VCDClient, c.User, c.Password, c.Token, c.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during authentication: %s", err)
	}
	cachedVCDClients.Lock()
	cachedVCDClients.conMap[checksum] = cachedConnection{initTime: time.Now(), connection: vcdClient}
	cachedVCDClients.Unlock()

	return vcdClient, nil
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

func init() {
	separator := os.Getenv("VCD_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	}
}
