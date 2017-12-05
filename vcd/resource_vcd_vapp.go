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
				ForceNew: true,
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vm": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: VMSchema},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"href": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"power_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// "metadata": {
			// 	Type:     schema.TypeMap,
			// 	Optional: true,
			// },
			// "ovf": {
			// 	Type:     schema.TypeMap,
			// 	Optional: true,
			// },
		},
	}
}

func createVMDescription(vmData map[string]interface{}, vAppNetworks []string, meta interface{}) (*types.NewVMDescription, error) {
	vcdClient := meta.(*VCDClient)

	catalog, err := vcdClient.Org.FindCatalog(vmData["catalog_name"].(string))
	if err != nil {
		return nil, fmt.Errorf("Error finding catalog: %#v", err)
	}

	catalogitem, err := catalog.FindCatalogItem(vmData["template_name"].(string))
	if err != nil {
		return nil, fmt.Errorf("Error finding catalog item: %#v", err)
	}

	vapptemplate, err := catalogitem.GetVAppTemplate()
	if err != nil {
		return nil, fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}

	log.Printf("[DEBUG] VAppTemplate: %#v", vapptemplate)

	networks := vmData["network"].([]interface{})
	if err != nil {
		return nil, fmt.Errorf("Error reading networks for vm: %#v", err)
	}

	var nets []*types.NetworkOrgDescription
	for _, n := range networks {

		network := n.(map[string]interface{})
		// Check if VM network is assigned to vApp
		if !isMember(vAppNetworks, network["name"].(string)) {
			return nil, fmt.Errorf("Network (%s) assigned to VM is not assigned to vApp, vApp has the following networks: %#v", network["name"].(string), vAppNetworks)
		}

		nets = append(nets,
			&types.NetworkOrgDescription{
				Name:             network["name"].(string),
				IsPrimary:        network["is_primary"].(bool),
				IsConnected:      network["is_connected"].(bool),
				IPAllocationMode: network["ip_allocation_mode"].(string),
				AdapterType:      network["adapter_type"].(string),
			},
		)
	}

	// net, err := vcdClient.OrgVdc.FindVDCNetwork(d.Get("network_name").(string))
	// if err != nil {
	// 	return fmt.Errorf("Error finding OrgVCD Network: %#v", err)
	// }

	// storage_profile_reference := types.Reference{}

	// // Override default_storage_profile if we find the given storage profile
	// if d.Get("storage_profile").(string) != "" {
	// 	storage_profile_reference, err = vcdClient.OrgVdc.FindStorageProfileReference(d.Get("storage_profile").(string))
	// 	if err != nil {
	// 		return fmt.Errorf("Error finding storage profile %s", d.Get("storage_profile").(string))
	// 	}
	// }

	vmDescription := &types.NewVMDescription{
		Name:         vmData["name"].(string),
		VAppTemplate: vapptemplate.VAppTemplate,
		Networks:     nets,
	}

	log.Printf("[DEBUG] NewVMDescription: %#v", vmDescription)

	return vmDescription, nil

}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	networks := interfaceListToStringList(d.Get("network").([]interface{}))
	log.Printf("[TRACE] Networks from state: %#v", networks)

	// Get VMs and create descriptions for the vAppCompose
	vmData := d.Get("vm").([]interface{})

	var vmDescriptions []*types.NewVMDescription
	for _, vm := range vmData {
		vmResource := vm.(map[string]interface{})
		vmDescription, err := createVMDescription(vmResource, networks, meta)

		if err != nil {
			return err
		}

		vmDescriptions = append(vmDescriptions, vmDescription)
	}

	var orgnetworks []*types.OrgVDCNetwork
	for _, network := range networks {
		orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(network)
		if err != nil {
			return fmt.Errorf("Error finding vdc org network: %s, %#v", network, err)
		}

		orgnetworks = append(orgnetworks, orgnetwork.OrgVDCNetwork)
	}

	// See if vApp exists
	vapp, err := vcdClient.OrgVdc.FindVAppByID(d.Get("href").(string))
	log.Printf("[TRACE] Looking for existing vapp, found %#v", vapp)

	if err != nil {
		log.Printf("[TRACE] No vApp found, preparing creation")
		vapp = vcdClient.NewVApp(&vcdClient.Client)

		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.ComposeVApp(d.Get("name").(string), d.Get("description").(string), orgnetworks, vmDescriptions)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error creating vapp: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})

		if err != nil {
			return fmt.Errorf("Error creating vapp: %#v", err)
		}
	}

	log.Printf("[DEBUG] vApp created with href:  %s", vapp.VApp.HREF)
	d.Set("href", vapp.VApp.HREF)

	// err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
	// 	task, err := vapp.ChangeVMName(d.Get("name").(string))
	// 	if err != nil {
	// 		return resource.RetryableError(fmt.Errorf("Error with vm name change: %#v", err))
	// 	}

	// 	return resource.RetryableError(task.WaitTaskCompletion())
	// })
	// if err != nil {
	// 	return fmt.Errorf("Error changing vmname: %#v", err)
	// }

	// err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
	// 	task, err := vapp.ChangeNetworkConfig(d.Get("network_name").(string), d.Get("ip").(string))
	// 	if err != nil {
	// 		return resource.RetryableError(fmt.Errorf("Error with Networking change: %#v", err))
	// 	}
	// 	return resource.RetryableError(task.WaitTaskCompletion())
	// })
	// if err != nil {
	// 	return fmt.Errorf("Error changing network: %#v", err)
	// }

	// if ovf, ok := d.GetOk("ovf"); ok {
	// 	err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
	// 		task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

	// 		if err != nil {
	// 			return resource.RetryableError(fmt.Errorf("Error set ovf: %#v", err))
	// 		}
	// 		return resource.RetryableError(task.WaitTaskCompletion())
	// 	})
	// 	if err != nil {
	// 		return fmt.Errorf("Error completing tasks: %#v", err)
	// 	}
	// }

	if d.Get("power_on").(bool) == true {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.PowerOn()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error powerOn machine: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})

		if err != nil {
			return fmt.Errorf("Error completing powerOn tasks: %#v", err)
		}
	}

	// initscript := d.Get("initscript").(string)

	// err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
	// 	log.Printf("running customisation script")
	// 	task, err := vapp.RunCustomizationScript(d.Get("name").(string), initscript)
	// 	if err != nil {
	// 		return resource.RetryableError(fmt.Errorf("Error with setting init script: %#v", err))
	// 	}
	// 	return resource.RetryableError(task.WaitTaskCompletion())
	// })
	// if err != nil {
	// 	return fmt.Errorf("Error completing tasks: %#v", err)
	// }

	d.SetId(d.Get("name").(string))

	return resourceVcdVAppUpdate(d, meta)
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v, %s", err, status)
	}

	// if d.HasChange("metadata") {
	// 	oraw, nraw := d.GetChange("metadata")
	// 	metadata := oraw.(map[string]interface{})
	// 	for k := range metadata {
	// 		task, err := vapp.DeleteMetadata(k)
	// 		if err != nil {
	// 			return fmt.Errorf("Error deleting metadata: %#v", err)
	// 		}
	// 		err = task.WaitTaskCompletion()
	// 		if err != nil {
	// 			return fmt.Errorf("Error completing tasks: %#v", err)
	// 		}
	// 	}
	// 	metadata = nraw.(map[string]interface{})
	// 	for k, v := range metadata {
	// 		task, err := vapp.AddMetadata(k, v.(string))
	// 		if err != nil {
	// 			return fmt.Errorf("Error adding metadata: %#v", err)
	// 		}
	// 		err = task.WaitTaskCompletion()
	// 		if err != nil {
	// 			return fmt.Errorf("Error completing tasks: %#v", err)
	// 		}
	// 	}

	// }

	if d.HasChange("power_on") || d.HasChange("ovf") {

		if d.Get("power_on").(bool) {
			task, err := vapp.PowerOn()
			if err != nil {
				return fmt.Errorf("Error Powering Up: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}

		// if ovf, ok := d.GetOk("ovf"); ok {
		// 	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		// 		task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

		// 		if err != nil {
		// 			return resource.RetryableError(fmt.Errorf("Error set ovf: %#v", err))
		// 		}
		// 		return resource.RetryableError(task.WaitTaskCompletion())
		// 	})
		// 	if err != nil {
		// 		return fmt.Errorf("Error completing tasks: %#v", err)
		// 	}
		// }

	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	_, err = vcdClient.OrgVdc.FindVAppByName(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find vapp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	// if _, ok := d.GetOk("ip"); ok {
	// 	ip := "allocated"

	// 	oldIp, newIp := d.GetChange("ip")

	// 	log.Printf("[DEBUG] IP has changes, old: %s - new: %s", oldIp, newIp)

	// 	if newIp != "allocated" {
	// 		log.Printf("[DEBUG] IP is assigned. Lets get it (%s)", d.Get("ip"))
	// 		ip, err = getVAppIPAddress(d, meta)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		log.Printf("[DEBUG] IP is 'allocated'")
	// 	}

	// 	d.Set("ip", ip)
	// } else {
	// 	d.Set("ip", "allocated")
	// }

	return nil
}

// func getVAppIPAddresses(d *schema.ResourceData, meta interface{}) ([]string, error) {
// 	vcdClient := meta.(*VCDClient)
// 	var ips []string

// 	err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
// 		err := vcdClient.OrgVdc.Refresh()
// 		if err != nil {
// 			return resource.RetryableError(fmt.Errorf("Error refreshing vdc: %#v", err))
// 		}
// 		vapp, err := vcdClient.OrgVdc.FindVAppByID(d.Get("href").(string))
// 		if err != nil {
// 			return resource.RetryableError(fmt.Errorf("Unable to find vapp."))
// 		}

// 		// getting the IP of the specific Vm, rather than index zero.
// 		// Required as once we add more VM's, index zero doesn't guarantee the
// 		// 'first' one, and tests will fail sometimes (annoying huh?)
// 		// vm, err := vcdClient.OrgVdc.FindVM(vapp, d.Get("name").(string))

// 		for index, vm := range vapp.VApp.Children.VM {
// 			ip := vm.NetworkConnectionSection.NetworkConnection.IPAddress
// 			if ip == "" {
// 				return resource.RetryableError(fmt.Errorf("Timeout: VM (%s) did not acquire IP address", vm.Name))
// 			}
// 			ips = append(ips, ip)
// 		}

// 		return nil
// 	})

// 	return ips, err
// }

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v", err)
	}

	_ = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Undeploy()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error undeploying: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Delete()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error deleting: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	return err
}

func isMember(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}
	return false
}

func interfaceListToStringList(old []interface{}) []string {
	new := make([]string, len(old))
	for i, v := range old {
		new[i] = v.(string)
	}
	return new
}
