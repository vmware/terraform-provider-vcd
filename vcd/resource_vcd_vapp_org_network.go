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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"org_network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
				Description: "NAT service enabled or disabled. Default - false",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "firewall service enabled or disabled. Default - true",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "NAT service enabled or disabled. Default - true",
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

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %#v", err)
	}

	natEnabled := d.Get("nat_enabled").(bool)
	fwEnabled := d.Get("firewall_enabled").(bool)
	retainIpMacEnabled := d.Get("retain_ip_mac_enabled").(bool)

	vappNetworkSettings := &govcd.VappNetworkSettings{
		NatEnabled:         &natEnabled,
		FirewallEnabled:    &fwEnabled,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(d.Get("org_network").(string), true)
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

	// we need not changeable value
	// Parsing UUID from 'https://bos1-vcloud-static-170-210.eng.vmware.com/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset'
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF)
	if err != nil {
		return fmt.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(networkId)

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
		return fmt.Errorf("error finding Vapp: %#v", err)
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return fmt.Errorf("error getting vApp networks: %#v", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	var networkId string
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		networkId, err = govcd.GetUuidFromHref(networkConfig.Link.HREF)
		if err != nil {
			return fmt.Errorf("unable to get network ID from HREF: %s", err)
		}
		// name check needed for datasource to find network as don't have ID
		if d.Id() == networkId || networkConfig.NetworkName == d.Get("org_network").(string) {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		if origin == "resource" {
			log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[VAPP org network read] didn't find vApp org network: %s", d.Get("org_network").(string))
	}

	// needs to set for datasource
	if d.Id() == "" {
		d.SetId(networkId)
	}

	_ = d.Set("retain_ip_mac_enabled", *vAppNetwork.Configuration.RetainNetInfoAcrossDeployments)

	isFenced := false
	if vAppNetwork.Configuration.FenceMode == types.FenceModeNAT {
		isFenced = true
	}
	_ = d.Set("is_fenced", isFenced)
	if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.FirewallService != nil {
		_ = d.Set("firewall_enabled", vAppNetwork.Configuration.Features.FirewallService.IsEnabled)
	}
	if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.NatService != nil {
		_ = d.Set("nat_enabled", vAppNetwork.Configuration.Features.NatService.IsEnabled)
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

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %#v", err)
	}

	retainIpMacEnabled := d.Get("retain_ip_mac_enabled").(bool)
	natEnabled := d.Get("nat_enabled").(bool)
	fwEnabled := d.Get("firewall_enabled").(bool)

	vappNetworkSettings := &govcd.VappNetworkSettings{
		Id:                 d.Id(),
		RetainIpMacEnabled: &retainIpMacEnabled,
		NatEnabled:         &natEnabled,
		FirewallEnabled:    &fwEnabled,
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
		return fmt.Errorf("error removing vApp network: %#v", err)
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
		return nil, fmt.Errorf("[VM import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[VM import] error retrieving vapp %s: %s", vappName, err)
	}
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("[VM import] error retrieving vApp network configuration %s: %s", networkName, err)
	}

	vappNetworkToImport := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		// name check needed to support old resource Id's which was names
		if networkConfig.NetworkName == networkName {
			vappNetworkToImport = networkConfig
			break
		}
	}

	if vappNetworkToImport == (types.VAppNetworkConfiguration{}) {
		return nil, fmt.Errorf("didn't find vApp org network: %s", networkName)
	}

	networkId, err := govcd.GetUuidFromHref(vappNetworkToImport.Link.HREF)
	if err != nil {
		return nil, fmt.Errorf("unable to get network ID from HREF: %s", err)
	}

	d.SetId(networkId)

	if vcdClient.Org != orgName {
		_ = d.Set("org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		_ = d.Set("vdc", vdcName)
	}
	_ = d.Set("org_network", networkName)
	_ = d.Set("vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}
