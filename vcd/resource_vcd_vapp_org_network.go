package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdVappOrgNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVappOrgNetworkCreate,
		Read:   resourceVappOrgNetworkRead,
		Update: resourceVappOrgNetworkUpdate,
		Delete: resourceVappOrgNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVappOrgNetworkImport,
		},

		Schema: map[string]*schema.Schema{
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network name",
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"org_network_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization network name to which vApp network is connected to",
			},
			"is_fenced": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp",
			},
			"retain_ip_mac_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "firewall service enabled or disabled.",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "NAT service enabled or disabled.",
			},
		},
	}
}

func resourceVappOrgNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return fmt.Errorf("error finding vApp: %s and err: %s", vappName, err)
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		NatEnabled:         takeBoolPointer(d.Get("nat_enabled").(bool)),
		FirewallEnabled:    takeBoolPointer(d.Get("firewall_enabled").(bool)),
		RetainIpMacEnabled: takeBoolPointer(d.Get("retain_ip_mac_enabled").(bool)),
	}

	orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(d.Get("org_network_name").(string), true)
	if err != nil {
		return err
	}

	vAppNetworkConfig, err := vapp.AddOrgNetwork(vappNetworkSettings, orgNetwork.OrgVDCNetwork, d.Get("is_fenced").(bool))
	if err != nil {
		return fmt.Errorf("error creating vApp org network. %#v", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == orgNetwork.OrgVDCNetwork.Name {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		return fmt.Errorf("didn't find vApp network: %s", d.Get("name").(string))
	}

	// Parsing UUID from 'https://bos1-vcloud-static-170-210.eng.vmware.com/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset'
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF, false)
	if err != nil {
		return fmt.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	return resourceVappOrgNetworkRead(d, meta)
}

func resourceVappOrgNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return genericVappOrgNetworkRead(d, meta, "resource")
}

func genericVappOrgNetworkRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding Vapp: %s", err)
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return fmt.Errorf("error getting vApp networks: %s", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	var networkId string
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.Link != nil {
			networkId, err = govcd.GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return fmt.Errorf("unable to get network ID from HREF: %s", err)
			}
			// name check needed for datasource to find network as don't have ID
			if d.Id() == networkId || networkConfig.NetworkName == d.Get("org_network_name").(string) {
				vAppNetwork = networkConfig
			}
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		if origin == "resource" {
			log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[VAPP org network read] %s : %s", govcd.ErrorEntityNotFound, d.Get("org_network_name").(string))
	}

	// needs to set for datasource
	if d.Id() == "" {
		d.SetId(normalizeId("urn:vcloud:network:", networkId))
	}

	_ = d.Set("retain_ip_mac_enabled", *vAppNetwork.Configuration.RetainNetInfoAcrossDeployments)

	isFenced := false
	if vAppNetwork.Configuration.FenceMode == types.FenceModeNAT {
		isFenced = true
	}
	_ = d.Set("is_fenced", isFenced)
	if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.FirewallService != nil {
		_ = d.Set("firewall_enabled", vAppNetwork.Configuration.Features.FirewallService.IsEnabled)
	} else {
		_ = d.Set("firewall_enabled", nil)
	}
	if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.NatService != nil {
		_ = d.Set("nat_enabled", vAppNetwork.Configuration.Features.NatService.IsEnabled)
	} else {
		_ = d.Set("nat_enabled", nil)
	}
	return nil
}

func resourceVappOrgNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return fmt.Errorf("error finding vApp: %s and err:  %s", vappName, err)
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		ID:                 d.Id(),
		NatEnabled:         takeBoolPointer(d.Get("nat_enabled").(bool)),
		FirewallEnabled:    takeBoolPointer(d.Get("firewall_enabled").(bool)),
		RetainIpMacEnabled: takeBoolPointer(d.Get("retain_ip_mac_enabled").(bool)),
	}

	_, err = vapp.UpdateOrgNetwork(vappNetworkSettings, d.Get("is_fenced").(bool))
	if err != nil {
		return fmt.Errorf("error creating vApp network. %#v", err)
	}

	return resourceVappOrgNetworkRead(d, meta)
}

func resourceVappOrgNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp: %#v", err)
	}

	_, err = vapp.RemoveNetwork(d.Id())
	if err != nil {
		return fmt.Errorf("error removing vApp network: %s", err)
	}

	d.SetId("")

	return nil
}

// resourceVcdVappOrgNetworkImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_org_network.org_network_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name.org-network-name
func resourceVcdVappOrgNetworkImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("[vApp org network import] resource name must be specified as org-name.vdc-name.vapp-name.org-network-name")
	}
	orgName, vdcName, vappName, networkName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] error retrieving vapp %s: %s", vappName, err)
	}
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] error retrieving vApp network configuration %s: %s", networkName, err)
	}

	vappNetworkToImport := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			vappNetworkToImport = networkConfig
			break
		}
	}

	if vappNetworkToImport == (types.VAppNetworkConfiguration{}) {
		return nil, fmt.Errorf("didn't find vApp org network: %s", networkName)
	}

	if isVappNetwork(&vappNetworkToImport) {
		return nil, fmt.Errorf("found vApp network, not vApp org network: %s", networkName)
	}

	networkId, err := govcd.GetUuidFromHref(vappNetworkToImport.Link.HREF, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get network ID from HREF: %s", err)
	}

	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	if vcdClient.Org != orgName {
		_ = d.Set("org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		_ = d.Set("vdc", vdcName)
	}
	_ = d.Set("org_network_name", networkName)
	_ = d.Set("vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}

// Allows to identify if given network config is a vApp network and not a vApp Org network
func isVappNetwork(networkConfig *types.VAppNetworkConfiguration) bool {
	if networkConfig.Configuration.FenceMode == types.FenceModeIsolated ||
		(networkConfig.Configuration.FenceMode == types.FenceModeNAT && networkConfig.Configuration.IPScopes != nil &&
			networkConfig.Configuration.IPScopes.IPScope != nil && len(networkConfig.Configuration.IPScopes.IPScope) > 0 &&
			!networkConfig.Configuration.IPScopes.IPScope[0].IsInherited) {
		return true
	}
	return false
}
