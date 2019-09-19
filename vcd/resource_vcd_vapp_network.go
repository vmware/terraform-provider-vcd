package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVappNetworkCreate,
		Read:   resourceVappNetworkRead,
		Delete: resourceVappNetworkDelete,

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

			"guest_vlan_allowed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeSet,
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
				Set: resourceVcdNetworkIPAddressHash,
			},
		},
	}
}

func resourceVcdVappNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("error finding vApp. %#v", err)
	}

	staticIpRanges := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())

	vappNetworkSettings := &govcd.VappNetworkSettings{
		Name:           d.Get("name").(string),
		Gateway:        d.Get("gateway").(string),
		NetMask:        d.Get("netmask").(string),
		DNS1:           d.Get("dns1").(string),
		DNS2:           d.Get("dns2").(string),
		DNSSuffix:      d.Get("dns_suffix").(string),
		StaticIPRanges: staticIpRanges.IPRange,
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

	task, err := vapp.AddIsolatedNetwork(vappNetworkSettings)

	if err != nil {
		return fmt.Errorf("error creating vApp network. %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task to complete: %+v", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceVappNetworkRead(d, meta)
}

func resourceVappNetworkRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("error finding Vapp: %#v", err)
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return fmt.Errorf("error getting vApp networks: %#v", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == d.Get("name").(string) {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}

	d.Set("name", vAppNetwork.NetworkName)
	d.Set("href", vAppNetwork.HREF)
	if c := vAppNetwork.Configuration; c != nil {
		d.Set("fence_mode", c.FenceMode)
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
			d.Set("guest_vlan_allowed", &c.GuestVlanAllowed)
		}
	}

	return nil
}

func resourceVappNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("error finding vApp: %#v", err)
	}

	task, err := vapp.RemoveIsolatedNetwork(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("error removing vApp network: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task to complete: %+v", err)
	}

	d.SetId("")

	return nil
}
