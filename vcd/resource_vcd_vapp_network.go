package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/govcd"
	"github.com/vmware/go-vcloud-director/types/v56"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
			},
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
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
				Default:  false,
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
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("error finding vApp. %#v", err)
	}

	ipRanges := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())

	task, err := vapp.AddIsolatedNetwork(&govcd.VappNetworkSettings{
		Name:             d.Get("name").(string),
		Gateway:          d.Get("gateway").(string),
		NetMask:          d.Get("netmask").(string),
		DNS1:             d.Get("dns1").(string),
		DNS2:             d.Get("dns2").(string),
		DNSSuffix:        d.Get("dns_suffix").(string),
		IPRange:          ipRanges.IPRange,
		GuestVLANAllowed: d.Get("guest_vlan_allowed").(bool)})
	if err != nil {
		return fmt.Errorf("error creating vApp network. %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting from task to complete: %+v", err)
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
			d.Set("gateway", c.IPScopes.IPScope.Gateway)
			d.Set("netmask", c.IPScopes.IPScope.Netmask)
			d.Set("dns1", c.IPScopes.IPScope.DNS1)
			d.Set("dns2", c.IPScopes.IPScope.DNS2)
		}
	}

	return nil
}

func resourceVappNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

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
		return fmt.Errorf("error waiting from task to complete: %+v", err)
	}

	d.SetId("")

	return nil
}
