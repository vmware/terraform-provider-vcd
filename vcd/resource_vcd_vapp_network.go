package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
)

func resourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVappNetworkCreate,
		Read:   resourceVappNetworkRead,
		Update: resourceVappNetworkUpdate,
		Delete: resourceVappNetworkDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
		if err != nil {
			return err
		}
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

	d.Set("name", vAppNetwork.NetworkName)
	d.Set("description", vAppNetwork.Description)
	d.Set("href", vAppNetwork.HREF)
	if c := vAppNetwork.Configuration; c != nil {
		if c.IPScopes != nil {
			d.Set("gateway", c.IPScopes.IPScope[0].Gateway)
			d.Set("netmask", c.IPScopes.IPScope[0].Netmask)
			d.Set("dns1", c.IPScopes.IPScope[0].DNS1)
			d.Set("dns2", c.IPScopes.IPScope[0].DNS2)
			d.Set("dnsSuffix", c.IPScopes.IPScope[0].DNSSuffix)
		}
		// TODO adjust when we have option to switch between API versions or upgrade the default version
		// API does not return GuestVlanAllowed if API client version is 27.0 (default at the moment) therefore we rely
		// on updating statefile only if the field was returned. In API v31.0 - the field is returned.
		if c.GuestVlanAllowed != nil {
			err = d.Set("guest_vlan_allowed", *c.GuestVlanAllowed)
			if err != nil {
				return err
			}
		}
		if vAppNetwork.Configuration.ParentNetwork != nil {
			d.Set("org_network", vAppNetwork.Configuration.ParentNetwork.Name)
		}
		d.Set("retain_ip_mac_enabled", vAppNetwork.Configuration.RetainNetInfoAcrossDeployments)
		if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.FirewallService != nil {
			d.Set("firewall_enabled", vAppNetwork.Configuration.Features.FirewallService.IsEnabled)
		}
		if vAppNetwork.Configuration.Features != nil && vAppNetwork.Configuration.Features.NatService != nil {
			d.Set("nat_enabled", vAppNetwork.Configuration.Features.NatService.IsEnabled)
		}
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
