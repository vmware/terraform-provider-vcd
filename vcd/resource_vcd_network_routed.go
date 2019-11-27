package vcd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A unique name for the network",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},

			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the edge gateway",
			},

			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional description for the network",
			},

			"interface_type": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "internal",
				ForceNew:         true,
				Description:      "Which interface to use (one of `internal`, `subinterface`, `distributed`)",
				ValidateFunc:     validation.StringInSlice([]string{"internal", "subinterface", "distributed"}, true),
				DiffSuppressFunc: suppressNetworkUpgradedInterface(),
			},

			"netmask": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "255.255.255.0",
				Description:  "The netmask for the new network",
				ValidateFunc: validation.SingleIP(),
			},

			"gateway": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The gateway of this network",
				ValidateFunc: validation.SingleIP(),
			},

			"dns1": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "8.8.8.8",
				Description:  "First DNS server to use",
				ValidateFunc: validation.SingleIP(),
			},

			"dns2": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "8.8.4.4",
				Description:  "Second DNS server to use",
				ValidateFunc: validation.SingleIP(),
			},

			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A FQDN for the virtual machines on this network",
			},

			"href": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Network Hypertext Reference",
			},

			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},

			"dhcp_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.SingleIP(),
						},

						"end_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.SingleIP(),
						},

						"default_lease_time": &schema.Schema{
							Type:             schema.TypeInt,
							Removed:          "vCD doesn't process this input. It sets the value to max_lease_time",
							Default:          3600,
							Optional:         true,
							DiffSuppressFunc: suppressAlways(),
							Description:      "The default DHCP lease time to use",
						},

						"max_lease_time": &schema.Schema{
							Type:        schema.TypeInt,
							Default:     7200,
							Optional:    true,
							Description: "HThe maximum DHCP lease time to use",
						},
					},
				},
				Set: resourceVcdNetworkIPAddressHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.SingleIP(),
						},

						"end_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.SingleIP(),
						},
					},
				},
				Set: resourceVcdNetworkIPAddressHash,
			},
		},
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
	netMask := d.Get("netmask").(string)
	dns1 := d.Get("dns1").(string)
	dns2 := d.Get("dns2").(string)

	ipRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        networkName,
		Description: d.Get("description").(string),

		EdgeGateway: &types.Reference{
			HREF: edgeGateway.EdgeGateway.HREF,
		},
		Configuration: &types.NetworkConfiguration{
			FenceMode: "natRouted",
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     netMask,
					DNS1:        dns1,
					DNS2:        dns2,
					DNSSuffix:   d.Get("dns_suffix").(string),
					IPRanges:    &ipRanges,
				}},
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}
	distributedAllowed := false
	if edgeGateway.EdgeGateway.Configuration.DistributedRoutingEnabled != nil {
		if *edgeGateway.EdgeGateway.Configuration.DistributedRoutingEnabled {
			distributedAllowed = true
		}
	}
	networkInterface := d.Get("interface_type").(string)
	trueValue := true
	switch strings.ToLower(networkInterface) {
	case "internal":
		// default: no configuration is needed
		orgVDCNetwork.Configuration.SubInterface = nil
	case "subinterface":
		orgVDCNetwork.Configuration.SubInterface = &trueValue
	case "distributed":
		if distributedAllowed {
			orgVDCNetwork.Configuration.DistributedInterface = &trueValue
		} else {
			return fmt.Errorf("interface 'distributed' requested, but edge gateway '%s' not enabled", edgeGateway.EdgeGateway.Name)
		}
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
		task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.(*schema.Set).List())
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
	return genericVcdNetworkRoutedRead(d, meta, "resource")
}

func genericVcdNetworkRoutedRead(d *schema.ResourceData, meta interface{}, origin string) error {
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
		if origin == "resource" {
			log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", identifier)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[network routed read] error retrieving Org VDC network %s: %s", identifier, err)
	}
	edgeGatewayName := d.Get("edge_gateway").(string)

	// When this function is called from the data source
	if edgeGatewayName == "" {
		edgeGatewayName, err = findEdgeGatewayConnection(vcdClient, vdc, network)
		if err != nil {
			return fmt.Errorf("[network_routed read] no edge gateway connection found for network %s: %s", network.OrgVDCNetwork.Name, err)
		}
		_ = d.Set("edge_gateway", edgeGatewayName)
	}
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
		newSet := &schema.Set{
			F: resourceVcdNetworkIPAddressHash,
		}
		for _, element := range dhcp {
			newSet.Add(element)
		}
		err := d.Set("dhcp_pool", newSet)
		if err != nil {
			return fmt.Errorf("[network routed read] dhcp set: %s", err)
		}
	}

	staticIpPool := getStaticIpPool(network)
	if len(staticIpPool) > 0 {
		newSet := &schema.Set{
			F: resourceVcdNetworkIPAddressHash,
		}
		for _, element := range staticIpPool {
			newSet.Add(element)
		}
		err := d.Set("static_ip_pool", newSet)
		if err != nil {
			return fmt.Errorf("[network routed read] static_ip set: %s", err)
		}
	}

	if network.OrgVDCNetwork.Configuration.SubInterface == nil {
		d.Set("interface_type", "internal")
	} else {
		if *network.OrgVDCNetwork.Configuration.SubInterface {
			d.Set("interface_type", "subinterface")
		} else {
			if *network.OrgVDCNetwork.Configuration.DistributedInterface {
				d.Set("interface_type", "distributed")
			}
		}
	}
	d.Set("description", network.OrgVDCNetwork.Description)

	d.SetId(network.OrgVDCNetwork.ID)
	return nil
}

func getStaticIpPool(network *govcd.OrgVDCNetwork) []map[string]interface{} {
	var staticIpPool []map[string]interface{}
	if network.OrgVDCNetwork.Configuration.IPScopes == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope) == 0 ||
		network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange) == 0 {
		return staticIpPool
	}
	for _, sip := range network.OrgVDCNetwork.Configuration.IPScopes.IPScope {
		if sip.IsEnabled {
			for _, iprange := range sip.IPRanges.IPRange {
				staticIp := map[string]interface{}{
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
// This is useful when comparing a HREF to a ID, or a HREF from an admin path
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
func getDhcpFromEdgeGateway(networkHref string, edgeGateway *govcd.EdgeGateway) []map[string]interface{} {

	var dhcpConfig []map[string]interface{}

	gwConf := edgeGateway.EdgeGateway.Configuration

	if gwConf == nil ||
		gwConf.EdgeGatewayServiceConfiguration == nil ||
		gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService == nil ||
		len(gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool) == 0 {
		return dhcpConfig
	}
	for _, dhcp := range gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool {
		if haveSameUuid(dhcp.Network.HREF, networkHref) {
			dhcpRec := map[string]interface{}{
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
		return fmt.Errorf("[network delete] error retrieving Org VDC network: %s", err)
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

	// Find the list of networks with the wanted name
	result, err := client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "orgVdcNetwork",
		"filter": fmt.Sprintf("name==%s;vdc==%s", url.QueryEscape(network.OrgVDCNetwork.Name), url.QueryEscape(vdc.Vdc.ID)),
	})
	if err != nil {
		return "", fmt.Errorf("[findEdgeGatewayConnection] error returning the list of networks for VDC: %s", err)
	}
	netList := result.Results.OrgVdcNetworkRecord

	// Find the list of edge gateways
	edgeGatewayResult := new(types.QueryResultEdgeGatewayRecordsType)
	for _, av := range vdc.Vdc.Link {
		if av.Rel == "edgeGateways" && av.Type == "application/vnd.vmware.vcloud.query.records+xml" {

			_, err := client.Client.ExecuteRequest(av.HREF, http.MethodGet,
				"", "error querying edge gateways: %s", nil, edgeGatewayResult)
			if err != nil {
				return "", err
			}
		}
	}

	var isAnEdgeGateway = func(gwName string) bool {
		for _, gw := range edgeGatewayResult.EdgeGatewayRecord {
			if gw.Name == gwName {
				return true
			}
		}
		return false
	}

	for _, net := range netList {
		if net.Name == network.OrgVDCNetwork.Name {
			if net.ConnectedTo != "" && isAnEdgeGateway(net.ConnectedTo) {
				return net.ConnectedTo, nil
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
// Example resource name (_resource_name_): vcd_network_routed.my-network
// Example import path (_the_id_string_): org.vdc.my-network
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdNetworkRoutedImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[network routed import] resource name must be specified as org-name.vdc-name.network-name")
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
