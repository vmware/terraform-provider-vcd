package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkDirectCreate,
		Read:   resourceVcdNetworkDirectRead,
		Update: resourceVcdNetworkDirectUpdate,
		Delete: resourceVcdNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNetworkDirectImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name for this network",
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
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description for the network",
			},
			"external_network": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the external network",
			},
			"external_network_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the external network",
			},
			"external_network_netmask": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Net mask of the external network",
			},
			"external_network_dns1": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Main DNS of the external network",
			},
			"external_network_dns2": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary DNS of the external network",
			},
			"external_network_dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix of the external network",
			},
			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hypertext Reference",
			},
			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},
		},
	}
}

func resourceVcdNetworkDirectCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return fmt.Errorf("creation of a vcd_network_direct requires system administrator privileges")
	}
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	externalNetworkName := d.Get("external_network").(string)
	networkName := d.Get("name").(string)
	externalNetwork, err := vcdClient.GetExternalNetworkByName(externalNetworkName)
	if err != nil {
		return fmt.Errorf("unable to find external network %s (%s)", externalNetworkName, err)
	}

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        networkName,
		Description: d.Get("description").(string),
		Configuration: &types.NetworkConfiguration{
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.ExternalNetwork.HREF,
				Type: externalNetwork.ExternalNetwork.Type,
				Name: externalNetwork.ExternalNetwork.Name,
			},
			FenceMode:                 "bridged",
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
		return fmt.Errorf("error retrieving network %s after creation", networkName)
	}
	d.SetId(network.OrgVDCNetwork.ID)
	return resourceVcdNetworkDirectRead(d, meta)
}

func resourceVcdNetworkDirectRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdNetworkDirectRead(d, meta, "resource")
}

func genericVcdNetworkDirectRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf("[direct network read] "+errorRetrievingOrgAndVdc, err)
	}

	network, err := getNetwork(d, vcdClient, origin == "datasource", "direct")
	if err != nil {
		if origin == "resource" {
			networkName := d.Get("name").(string)
			log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", networkName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[direct network read] network not found: %s", err)
	}

	_ = d.Set("name", network.OrgVDCNetwork.Name)
	_ = d.Set("href", network.OrgVDCNetwork.HREF)
	_ = d.Set("shared", network.OrgVDCNetwork.IsShared)

	// Getting external network data through network list, as a direct call to external network
	// structure requires system admin privileges.
	// Org Users can't create a direct network, but should be able to see the connection info.
	networkList, err := vdc.GetNetworkList()
	if err != nil {
		return fmt.Errorf("error retrieving network list for VDC %s : %s", vdc.Vdc.Name, err)
	}
	var currentNetwork *types.QueryResultOrgVdcNetworkRecordType
	for _, net := range networkList {
		if net.Name == network.OrgVDCNetwork.Name {
			currentNetwork = net
		}
	}
	if currentNetwork == nil {
		return fmt.Errorf("error retrieving network %s from network list", network.OrgVDCNetwork.Name)
	}
	_ = d.Set("external_network", currentNetwork.ConnectedTo)
	_ = d.Set("external_network_netmask", currentNetwork.Netmask)
	_ = d.Set("external_network_dns1", currentNetwork.Dns1)
	_ = d.Set("external_network_dns2", currentNetwork.Dns2)
	_ = d.Set("external_network_dns_suffix", currentNetwork.DnsSuffix)
	// Fixes issue #450
	_ = d.Set("external_network_gateway", currentNetwork.DefaultGateway)

	_ = d.Set("description", network.OrgVDCNetwork.Description)

	d.SetId(network.OrgVDCNetwork.ID)

	return nil
}

func resourceVcdNetworkDirectUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return fmt.Errorf("update of a vcd_network_direct requires system administrator privileges")
	}
	network, err := getNetwork(d, vcdClient, false, "direct")
	if err != nil {
		return fmt.Errorf("[direct network update] error getting network: %s", err)
	}

	networkName := d.Get("name").(string)
	network.OrgVDCNetwork.Name = networkName
	network.OrgVDCNetwork.Description = d.Get("description").(string)
	network.OrgVDCNetwork.IsShared = d.Get("shared").(bool)

	err = network.Update()
	if err != nil {
		return fmt.Errorf("[direct network update] error updating network %s: %s", network.OrgVDCNetwork.Name, err)
	}

	return resourceVcdNetworkDirectRead(d, meta)
}

func getNetwork(d *schema.ResourceData, vcdClient *VCDClient, isDataSource bool, wanted string) (*govcd.OrgVDCNetwork, error) {

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	var network *govcd.OrgVDCNetwork
	if isDataSource {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_network_"+wanted)
		}
		filter, hasFilter := d.GetOk("filter")

		if hasFilter {
			network, err = getNetworkByFilter(vdc, filter, wanted)
			if err != nil {
				return nil, err
			}
			return network, nil
		}
	}

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, fmt.Errorf("[get network] no identifier found for network")
	}
	network, err = vdc.GetOrgVdcNetworkByNameOrId(identifier, false)
	if err != nil {
		return nil, fmt.Errorf("[get network] error getting network %s: %s", identifier, err)
	}

	return network, nil
}

// resourceVcdNetworkDirectImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_network_direct.my-network
// Example import path (_the_id_string_): org.vdc.my-network
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdNetworkDirectImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[direct network import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[direct network import] unable to find VDC %s: %s ", vdcName, err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[direct network import] error retrieving network %s: %s", networkName, err)
	}
	parentNetwork := network.OrgVDCNetwork.Configuration.ParentNetwork
	if parentNetwork == nil || parentNetwork.Name == "" {
		return nil, fmt.Errorf("[direct network import] no parent network found for %s", network.OrgVDCNetwork.Name)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("external_network", parentNetwork.Name)
	d.SetId(network.OrgVDCNetwork.ID)
	return []*schema.ResourceData{d}, nil
}
