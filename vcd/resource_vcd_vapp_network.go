package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVappNetworkCreate,
		Read:   resourceVappNetworkRead,
		Update: resourceVappNetworkUpdate,
		Delete: resourceVappNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVappNetworkImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vapp_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "255.255.255.0",
			},
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dns1": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "8.8.8.8",
			},

			"dns2": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "8.8.4.4",
			},

			"dns_suffix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"guest_vlan_allowed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"org_network": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "org network name to which vapp network is connected",
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
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "NAT service enabled or disabled. Default - true",
			},
			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"default_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Default:  3600,
							Optional: true,
						},

						"max_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Default:  7200,
							Optional: true,
						},

						"enabled": &schema.Schema{
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
					},
				},
				Set: resourceVcdNetworkIPAddressHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
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
				Set: resourceVcdNetworkIPAddressHash,
			},
		},
	}
}

func resourceVappNetworkCreate(d *schema.ResourceData, meta interface{}) error {
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

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	natEnabled := d.Get("nat_enabled").(bool)
	fwEnabled := d.Get("firewall_enabled").(bool)
	retainIpMacEnabled := d.Get("retain_ip_mac_enabled").(bool)

	vappNetworkSettings := &govcd.VappNetworkSettings{
		Name:               d.Get("name").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		NatEnabled:         &natEnabled,
		FirewallEnabled:    &fwEnabled,
		Description:        d.Get("description").(string),
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok && len(dhcp.(*schema.Set).List()) > 0 {
		for _, item := range dhcp.(*schema.Set).List() {
			data := item.(map[string]interface{})
			vappNetworkSettings.DhcpSettings = &govcd.DhcpSettings{
				IsEnabled:        data["enabled"].(bool),
				DefaultLeaseTime: data["default_lease_time"].(int),
				MaxLeaseTime:     data["max_lease_time"].(int),
				IPRange: &types.IPRange{StartAddress: data["start_address"].(string),
					EndAddress: data["end_address"].(string)}}
		}
	}
	var orgVdcNetwork *types.OrgVDCNetwork
	if networkId, ok := d.GetOk("org_network"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkId.(string), true)
		if err != nil {
			return err
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}
	vAppNetworkConfig, err := vapp.AddNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return fmt.Errorf("error creating vApp network. %#v", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == d.Get("name").(string) {
			vAppNetwork = networkConfig
		}
	}

	// Parsing UUID from 'https://bos1-vcloud-static-170-210.eng.vmware.com/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset' or similar
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF)
	if err != nil {
		return fmt.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(networkId)

	return resourceVappNetworkRead(d, meta)
}

func resourceVappNetworkRead(d *schema.ResourceData, meta interface{}) error {
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
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		networkId, err := govcd.GetUuidFromHref(networkConfig.Link.HREF)
		if err != nil {
			return fmt.Errorf("unable to get network ID from HREF: %s", err)
		}
		// name check needed to support old resource Id's which was names
		if d.Id() == networkId || networkConfig.NetworkName == d.Get("name").(string) {
			vAppNetwork = networkConfig
			break
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}

	d.Set("description", vAppNetwork.Description)
	if config := vAppNetwork.Configuration; config != nil {
		if config.IPScopes != nil {
			d.Set("gateway", config.IPScopes.IPScope[0].Gateway)
			d.Set("netmask", config.IPScopes.IPScope[0].Netmask)
			d.Set("dns1", config.IPScopes.IPScope[0].DNS1)
			d.Set("dns2", config.IPScopes.IPScope[0].DNS2)
			d.Set("dns_suffix", config.IPScopes.IPScope[0].DNSSuffix)
		}

		if config.Features != nil && config.Features.DhcpService != nil {
			transformed := schema.NewSet(resourceVcdNetworkIPAddressHash, []interface{}{})
			newValues := map[string]interface{}{
				"enabled":            config.Features.DhcpService.IsEnabled,
				"max_lease_time":     config.Features.DhcpService.MaxLeaseTime,
				"default_lease_time": config.Features.DhcpService.DefaultLeaseTime,
			}
			if config.Features.DhcpService.IPRange != nil {
				newValues["start_address"] = config.Features.DhcpService.IPRange.StartAddress
				newValues["end_address"] = config.Features.DhcpService.IPRange.EndAddress
			}
			transformed.Add(newValues)
			d.Set("dhcp_pool", transformed)
		}

		if config.IPScopes != nil && config.IPScopes.IPScope[0].IPRanges != nil {
			staticIpRanges := schema.NewSet(resourceVcdNetworkIPAddressHash, []interface{}{})
			for _, ipRange := range config.IPScopes.IPScope[0].IPRanges.IPRange {
				newValues := map[string]interface{}{
					"start_address": ipRange.StartAddress,
					"end_address":   ipRange.EndAddress,
				}
				staticIpRanges.Add(newValues)
			}
			d.Set("static_ip_pool", staticIpRanges)
		}

		// TODO adjust when we have option to switch between API versions or upgrade the default version
		// API does not return GuestVlanAllowed if API client version is 27.0 (default at the moment) therefore we rely
		// on updating statefile only if the field was returned. In API v31.0 - the field is returned.
		if config.GuestVlanAllowed != nil {
			err = d.Set("guest_vlan_allowed", *config.GuestVlanAllowed)
			if err != nil {
				return err
			}
		}
		if vAppNetwork.Configuration.ParentNetwork != nil {
			d.Set("org_network", vAppNetwork.Configuration.ParentNetwork.Name)
		}
		if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.FirewallService != nil {
			d.Set("firewall_enabled", vAppNetwork.Configuration.Features.FirewallService.IsEnabled)
		}
		if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.NatService != nil {
			d.Set("nat_enabled", vAppNetwork.Configuration.Features.NatService.IsEnabled)
		}
		d.Set("retain_ip_mac_enabled", vAppNetwork.Configuration.RetainNetInfoAcrossDeployments)
	}

	return nil
}

func resourceVappNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
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

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	natEnabled := d.Get("nat_enabled").(bool)
	fwEnabled := d.Get("firewall_enabled").(bool)
	retainIpMacEnabled := d.Get("retain_ip_mac_enabled").(bool)

	// we can't change network name as this results ID(href) change
	vappNetworkSettings := &govcd.VappNetworkSettings{
		Id:                 d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		NatEnabled:         &natEnabled,
		FirewallEnabled:    &fwEnabled,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok && len(dhcp.(*schema.Set).List()) > 0 {
		for _, item := range dhcp.(*schema.Set).List() {
			data := item.(map[string]interface{})
			vappNetworkSettings.DhcpSettings = &govcd.DhcpSettings{
				IsEnabled:        data["enabled"].(bool),
				DefaultLeaseTime: data["default_lease_time"].(int),
				MaxLeaseTime:     data["max_lease_time"].(int),
				IPRange: &types.IPRange{StartAddress: data["start_address"].(string),
					EndAddress: data["end_address"].(string)}}
		}
	}

	var orgVdcNetwork *types.OrgVDCNetwork
	if networkId, ok := d.GetOk("org_network"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkId.(string), true)
		if err != nil {
			return err
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}

	_, err = vapp.UpdateNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return fmt.Errorf("error creating vApp network. %#v", err)
	}
	return resourceVappNetworkRead(d, meta)
}

func resourceVappNetworkDelete(d *schema.ResourceData, meta interface{}) error {
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

// resourceVcdVappNetworkImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_network.network_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name.network-name
func resourceVcdVappNetworkImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("[vApp network import] resource name must be specified as org-name.vdc-name.vapp-name.network-name")
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
		return nil, fmt.Errorf("didn't find vApp network: %s", networkName)
	}

	networkId, err := govcd.GetUuidFromHref(vappNetworkToImport.Link.HREF)
	if err != nil {
		return nil, fmt.Errorf("unable to get network ID from HREF: %s", err)
	}

	d.SetId(networkId)

	if vcdClient.Org != orgName {
		d.Set("org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		d.Set("vdc", vdcName)
	}
	_ = d.Set("name", networkName)
	_ = d.Set("vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}
