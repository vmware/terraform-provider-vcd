package vcd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkRouted() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkRoutedCreate,
		Read:   resourceVcdNetworkRoutedRead,
		Delete: resourceVcdNetworkDeleteLocked,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNetworkRoutedImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "255.255.255.0",
			},

			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"dns1": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "8.8.8.8",
			},

			"dns2": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "8.8.4.4",
			},

			"dns_suffix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"href": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"default_lease_time": &schema.Schema{
							Type:             schema.TypeInt,
							Removed:          "vCD doesn't process this input. It sets the value to max_lease_time",
							Default:          3600,
							Optional:         true,
							DiffSuppressFunc: suppressAlways(),
						},

						"max_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Default:  7200,
							Optional: true,
						},
					},
				},
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// suppressAlways suppresses the processing of the property unconditionally
// Used when we want to remove a property that should not have been
// added in the first place, but we want to keep compatibility
func suppressAlways() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return true
	}
}

func resourceVcdNetworkRoutedCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	edgeGatewayName := d.Get("edge_gateway").(string)
	edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	gatewayName := d.Get("gateway").(string)
	networkName := d.Get("name").(string)

	ipRanges := expandIPRange(d.Get("static_ip_pool").([]interface{}))

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  networkName,

		EdgeGateway: &types.Reference{
			HREF: edgeGateway.EdgeGateway.HREF,
		},
		Configuration: &types.NetworkConfiguration{
			FenceMode: "natRouted",
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     d.Get("netmask").(string),
					DNS1:        d.Get("dns1").(string),
					DNS2:        d.Get("dns2").(string),
					DNSSuffix:   d.Get("dns_suffix").(string),
					IPRanges:    &ipRanges,
				}},
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, true)
	if err != nil {
		return fmt.Errorf("error finding network: %s", err)
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok {
		task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.([]interface{}))
		if err != nil {
			return fmt.Errorf("error adding DHCP pool: %s", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}

	}

	d.SetId(network.OrgVDCNetwork.ID)

	return resourceVcdNetworkRoutedRead(d, meta)
}

func resourceVcdNetworkRoutedRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}
	identifier := d.Id()

	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	network, err := vdc.GetOrgVdcNetworkByNameOrId(identifier, false)
	if err != nil {
		log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", identifier)
		return fmt.Errorf("[network routed read] error retrieving network %s: %s", identifier, err)
	}
	edgeGatewayName := d.Get("edge_gateway").(string)

	edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
	if err != nil {
		log.Printf("[DEBUG] error retrieving edge gateway")
		return fmt.Errorf("[network routed read] error retrieving edge gateway %s: %s", edgeGatewayName, err)
	}

	_ = d.Set("name", network.OrgVDCNetwork.Name)
	_ = d.Set("href", network.OrgVDCNetwork.HREF)
	_ = d.Set("shared", network.OrgVDCNetwork.IsShared)
	if c := network.OrgVDCNetwork.Configuration; c != nil {
		_ = d.Set("fence_mode", c.FenceMode)
		if c.IPScopes != nil {
			_ = d.Set("gateway", c.IPScopes.IPScope[0].Gateway)
			_ = d.Set("netmask", c.IPScopes.IPScope[0].Netmask)
			_ = d.Set("dns1", c.IPScopes.IPScope[0].DNS1)
			_ = d.Set("dns2", c.IPScopes.IPScope[0].DNS2)
			_ = d.Set("dns_suffix", c.IPScopes.IPScope[0].DNSSuffix)
		}
	}

	dhcp := getDhcpFromEdgeGateway(network.OrgVDCNetwork.HREF, edgeGateway)
	if len(dhcp) > 0 {
		err := d.Set("dhcp_pool", dhcp)
		if err != nil {
			return err
		}
	}

	staticIpPool := getStaticIpPool(network)
	if len(staticIpPool) > 0 {
		err := d.Set("static_ip_pool", staticIpPool)
		if err != nil {
			return err
		}
	}

	d.SetId(network.OrgVDCNetwork.ID)
	return nil
}

func getStaticIpPool(network *govcd.OrgVDCNetwork) []StringMap {
	var staticIpPool []StringMap
	if network.OrgVDCNetwork.Configuration.IPScopes == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope) == 0 ||
		network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange) == 0 {
		return staticIpPool
	}
	for _, sip := range network.OrgVDCNetwork.Configuration.IPScopes.IPScope {
		if sip.IsEnabled {
			for _, iprange := range sip.IPRanges.IPRange {
				staticIp := StringMap{
					"start_address": iprange.StartAddress,
					"end_address":   iprange.EndAddress,
				}
				staticIpPool = append(staticIpPool, staticIp)
			}
		}
	}

	return staticIpPool
}

// hasSameUuid compares two IDs (or HREF)
// and returns true if the UUID part of the two input strings are the same.
// This is useful when comparing a HREF to a ID, or a HREF from an anmin path
// to a HREF from a regular user path.
func haveSameUuid(s1, s2 string) bool {
	reUuid := regexp.MustCompile(`/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})$`)
	s1List := reUuid.FindAllStringSubmatch(s1, -1)
	s2List := reUuid.FindAllStringSubmatch(s2, -1)
	if len(s1List) > 0 && len(s1List[0]) > 0 && len(s2List) > 0 && len(s2List[0]) > 0 {
		return s1List[0][1] == s2List[0][1]
	}
	return false
}

// getDhcpFromEdgeGateway examines the edge gateway for a DHCP service
// that is registered to the given network HREF.
// Returns an array of string maps suitable to be passed to d.Set("dhcp_pool", value)
func getDhcpFromEdgeGateway(networkHref string, edgeGateway *govcd.EdgeGateway) []StringMap {

	var dhcpConfig []StringMap

	gwConf := edgeGateway.EdgeGateway.Configuration

	if gwConf == nil ||
		gwConf.EdgeGatewayServiceConfiguration == nil ||
		gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService == nil ||
		len(gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool) == 0 {
		return dhcpConfig
	}
	for _, dhcp := range gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool {
		if haveSameUuid(dhcp.Network.HREF, networkHref) {
			dhcpRec := StringMap{
				"start_address":      dhcp.LowIPAddress,
				"end_address":        dhcp.HighIPAddress,
				"default_lease_time": dhcp.MaxLeaseTime,
				"max_lease_time":     dhcp.MaxLeaseTime,
			}
			dhcpConfig = append(dhcpConfig, dhcpRec)
		}
	}
	return dhcpConfig
}

func resourceVcdNetworkDeleteLocked(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	return resourceVcdNetworkDelete(d, meta)
}

func resourceVcdNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.GetOrgVdcNetworkByNameOrId(d.Id(), false)
	if err != nil {
		return fmt.Errorf("[delete] error retrieving network: %s", err)
	}

	task, err := network.Delete()
	if err != nil {
		return fmt.Errorf("error deleting network: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	return nil
}

func resourceVcdNetworkIPAddressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["start_address"].(string))))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["end_address"].(string))))

	return hashcode.String(buf.String())
}

// findEdgeGatewayConnection scans the VDC for a connection between an edge gateway and a given network.
// On success, returns the name of the edge gateway
func findEdgeGatewayConnection(client *VCDClient, vdc *govcd.Vdc, network *govcd.OrgVDCNetwork) (string, error) {

	for _, av := range vdc.Vdc.Link {
		if av.Rel == "edgeGateways" && av.Type == "application/vnd.vmware.vcloud.query.records+xml" {

			edgeGatewayRecordsType := new(types.QueryResultEdgeGatewayRecordsType)

			_, err := client.Client.ExecuteRequest(av.HREF, http.MethodGet,
				"", "error querying edge gateways: %s", nil, edgeGatewayRecordsType)
			if err != nil {
				return "", err
			}
			for _, eg := range edgeGatewayRecordsType.EdgeGatewayRecord {
				edgeGateway, err := vdc.GetEdgeGatewayByName(eg.Name, false)
				if err != nil {
					return "", err
				}
				for _, gi := range edgeGateway.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
					if gi.Network.Name == network.OrgVDCNetwork.Name {
						return edgeGateway.EdgeGateway.Name, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("no edge gateway connection found")
}

// resourceVcdNetworkRoutedImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_network_isolated.my-network
// Example import path (_the_id_string_): org.vdc.edge-gateway.my-network
func resourceVcdNetworkRoutedImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[network routed import] resource name must be specified as org.vdc.network")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[network routed import] unable to find VDC %s: %s ", vdcName, err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[network_routed import] error retrieving network %s: %s", networkName, err)
	}

	edgeGatewayName, err := findEdgeGatewayConnection(vcdClient, vdc, network)
	if err != nil {
		return nil, fmt.Errorf("[network_routed import] no edge gateway connection found for network %s: %s", network.OrgVDCNetwork.Name, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("edge_gateway", edgeGatewayName)
	d.SetId(network.OrgVDCNetwork.ID)

	return []*schema.ResourceData{d}, nil
}
