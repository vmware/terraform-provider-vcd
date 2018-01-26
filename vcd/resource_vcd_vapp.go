package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func resourceVcdVApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppCreate,
		Update: resourceVcdVAppUpdate,
		Read:   resourceVcdVAppRead,
		Delete: resourceVcdVAppDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vapp_network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: VAppNetworkSubresourceSchema(),
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"href": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// vApp Network
	vAppNetworksInterfaceList := d.Get("vapp_network").([]interface{})
	vAppNetworks := interfaceListToMapStringInterface(vAppNetworksInterfaceList)

	vAppNetworkConfigurations := make([]*types.VAppNetworkConfiguration, len(vAppNetworks))
	for index := range vAppNetworks {
		vAppNetwork := NewVAppNetworkSubresource(vAppNetworks[index], nil)

		configuration := &types.NetworkConfiguration{
			FenceMode: types.FenceModeIsolated,
			Features:  &types.NetworkFeatures{},
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
					IsInherited: false,
					Gateway:     vAppNetwork.Get("gateway").(string),
					Netmask:     vAppNetwork.Get("netmask").(string),
					DNS1:        vAppNetwork.Get("dns1").(string),
					DNS2:        vAppNetwork.Get("dns2").(string),
					IsEnabled:   true,
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{&types.IPRange{
							StartAddress: vAppNetwork.Get("start").(string),
							EndAddress:   vAppNetwork.Get("end").(string),
						}},
					},
				},
			},
		}

		// if vAppNetworks.Get("dhcp").(bool) {
		// 	configuration.Features = &types.NetworkFeatures{
		// 		DhcpService: &types.DhcpService{
		// 			IsEnabled: True,

		// 		}
		// 	}
		// }

		if vAppNetwork.Get("nat").(bool) {
			configuration.Features.NatService = &types.NatService{
				IsEnabled: true,
				NatType:   "ipTranslation",
				Policy:    "allowTrafficIn",
				// We need to set parent
			}
			configuration.FenceMode = types.FenceModeNAT
		}

		vAppNetworkConfigurations[index] = &types.VAppNetworkConfiguration{
			Configuration: configuration,
			NetworkName:   vAppNetwork.Get("name").(string),
		}
	}

	// Organization Network
	organizationNetworks := d.Get("organization_network").([]interface{})
	log.Printf("[TRACE] Networks from state: %#v", organizationNetworks)

	orgnetworks := make([]*types.OrgVDCNetwork, len(organizationNetworks))
	for index, network := range organizationNetworks {
		orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(network.(string))
		if err != nil {
			return fmt.Errorf("Error finding vdc org network: %s, %#v", network, err)
		}
		orgnetworks[index] = orgnetwork.OrgVDCNetwork
	}

	// See if vApp exists
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())
	log.Printf("[TRACE] Looking for existing vapp, found %#v", vapp)

	if err != nil {
		log.Printf("[TRACE] No vApp found, preparing creation")
		vapp = vcdClient.NewVApp(&vcdClient.Client)

		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.ComposeVApp(d.Get("name").(string), d.Get("description").(string), orgnetworks, vAppNetworkConfigurations)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error creating vapp: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})

		if err != nil {
			return fmt.Errorf("Error creating vapp: %#v", err)
		}
	}

	log.Printf("[DEBUG] vApp created with href:  %s", vapp.VApp.HREF)
	d.Set("href", vapp.VApp.HREF)

	// Refresh vcd and vApp to get the new versions
	log.Printf("[TRACE] Updating state from VCD")
	err = vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	log.Printf("[TRACE] Updateing vApp (%s) state", vapp.VApp.Name)
	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vApp: %#v", err)
	}

	// This should be HREF, but FindVAppByHREF is buggy
	d.SetId(d.Get("name").(string))

	return nil
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] Updating state from VCD")
	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v, %s", err, status)
	}

	// Update networks
	if d.HasChange("network") {
		log.Printf("[TRACE] (%s) Updating vApp networks", vapp.VApp.Name)
		networks := d.Get("organization_network").([]interface{})
		orgnetworks := make([]*types.OrgVDCNetwork, len(networks))
		for index, network := range networks {
			orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(network.(string))
			if err != nil {
				return fmt.Errorf("Error finding vdc org network: %s, %#v", network, err)
			}

			orgnetworks[index] = orgnetwork.OrgVDCNetwork
		}
	}

	return nil
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] Updating state from VCD")
	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	// Should be fetched by ID/HREF
	_, err = vcdClient.OrgVdc.FindVAppByName(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find vapp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	err = readVApp(d, meta)

	if err != nil {
		return err
	}

	return nil
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] Updating state from VCD")
	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v, %s", err, status)
	}

	_ = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Undeploy()
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error undeploying: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Delete()
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error deleting: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	if err != nil {
		return err
	}

	return nil
}
