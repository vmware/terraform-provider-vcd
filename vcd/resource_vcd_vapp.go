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
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vm": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: VirtualMachineSubresourceSchema()},
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
		},
	}
}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	networks := d.Get("network").([]interface{})
	log.Printf("[TRACE] Networks from state: %#v", networks)

	// Get VMs and create descriptions for the vAppCompose
	oldState, newState := d.GetChange("vm")

	oldStateListOfVms := interfaceListToMapStringInterface(oldState.([]interface{}))
	newStateListOfVms := interfaceListToMapStringInterface(newState.([]interface{}))

	newVMDescriptions := make([]*types.NewVMDescription, len(newStateListOfVms))
	for index := range newStateListOfVms {
		vmDescription, err := createVMDescription(newStateListOfVms[index], interfaceListToStringList(networks), meta)

		if err != nil {
			return err
		}

		newVMDescriptions = append(newVMDescriptions, vmDescription)
	}

	orgnetworks := make([]*types.OrgVDCNetwork, len(networks))
	for index, network := range networks {
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
			task, err := vapp.ComposeVApp(d.Get("name").(string), d.Get("description").(string), orgnetworks, newVMDescriptions)
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
	log.Printf("[TRACE] Refreshing vcd")
	err = vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	log.Printf("[TRACE] Refreshing vApp (%s)", vapp.VApp.Name)
	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vApp: %#v", err)
	}

	// Start configuring the machines
	log.Printf("[TRACE] (%s) Updating virtual machines", vapp.VApp.Name)
	// newVMResources := make([]map[string]interface{}, len(vmResources))

	for index := range newStateListOfVms {
		vm, err := vapp.GetVmByName(newStateListOfVms[index]["name"].(string))

		if err != nil {
			return err
		}

		href := vm.HREF
		newStateListOfVms[index]["href"] = href
		vmSubResource := NewVirtualMachineSubresource(newStateListOfVms[index], oldStateListOfVms[index], 0)

		err = configureVM(vmSubResource, meta)

		if err != nil {
			return err
		}
		// newVMResources[index] = vmResourceAfterConfiguration
		newStateListOfVms[index] = vmSubResource.Data()
	}

	d.Set("vm", newStateListOfVms)

	// This should be HREF, but FindVAppByHREF is buggy
	d.SetId(d.Get("name").(string))

	return nil
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// TODO: HERE WE MUST CHECK AND ADD NEW OR REMOVE VMs
	// new/remove vm
	// new/remove vapp networks
	// changes to vms...?

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
		networks := d.Get("network").([]interface{})
		orgnetworks := make([]*types.OrgVDCNetwork, len(networks))
		for index, network := range networks {
			orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(network.(string))
			if err != nil {
				return fmt.Errorf("Error finding vdc org network: %s, %#v", network, err)
			}

			orgnetworks[index] = orgnetwork.OrgVDCNetwork
		}
	}

	// Updates VMs
	if d.HasChange("vm") {
		oldState, newState := d.GetChange("vm")

		oldStateListOfVms := interfaceListToMapStringInterface(oldState.([]interface{}))
		newStateListOfVms := interfaceListToMapStringInterface(newState.([]interface{}))

		newVms := make([]map[string]interface{}, 0)
		removedVms := make([]map[string]interface{}, 0)

		for index := range oldStateListOfVms {
			if !isVMMapStringInterfaceMember(newStateListOfVms, oldStateListOfVms[index]) {
				removedVms = append(removedVms, oldStateListOfVms[index])
			}
		}
		log.Printf("[TRACE] (%s) VMs to remove: %#v", vapp.VApp.Name, removedVms)

		for index := range newStateListOfVms {
			if !isVMMapStringInterfaceMember(oldStateListOfVms, newStateListOfVms[index]) {
				newVms = append(newVms, newStateListOfVms[index])
			}
		}
		log.Printf("[TRACE] (%s) VMs to add: %#v", vapp.VApp.Name, newVms)
		// log.Printf("[TRACE] (%s) VMs to change: %#v", vapp.VApp.Name, changedVms)

		// Delete VMs
		removedVmsAsVMType := make([]*types.VM, 0)
		for index := range removedVms {
			vm, err := vapp.GetVmByHREF(removedVms[index]["href"].(string))

			if err != nil {
				return err
			}

			if vm != nil {
				removedVmsAsVMType = append(removedVmsAsVMType, vm)
			}
		}

		// Send delete request to vApp
		log.Printf("[TRACE] (%s) Removing VMs", vapp.VApp.Name)
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.RemoveVMs(removedVmsAsVMType)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error deleting VMs: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("Error completing task: %#v", err)
		}

		// Add VMs
		networks := d.Get("network").([]interface{})

		newVmDescriptions := make([]*types.NewVMDescription, len(newVms))
		for index := range newVms {
			vmDescription, err := createVMDescription(newVms[index], interfaceListToStringList(networks), meta)

			if err != nil {
				return err
			}

			log.Printf("[TRACE] VMDescription order: %d %s", index, vmDescription.Name)

			newVmDescriptions[index] = vmDescription
		}

		// Send add request to vApp
		log.Printf("[TRACE] (%s) Adding VMs", vapp.VApp.Name)
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.AddVMs(newVmDescriptions)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error adding VMs: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("Error completing task: %#v", err)
		}

		// Find HREF for new virtual machines
		log.Printf("[TRACE] (%s) Adding HREF for new VMs", vapp.VApp.Name)
		// vmResources := make([]map[string]interface{}, len(changedVms))
		for index := range newVms {
			log.Printf("[TRACE] (%s) VM found with no HREF: %s", vapp.VApp.Name, newVms[index]["name"].(string))
		}

		// Start configuring the machines
		log.Printf("[TRACE] (%s) Updating virtual machines", vapp.VApp.Name)
		// newVMResources := make([]map[string]interface{}, len(vmResources))

		for index := range newStateListOfVms {
			if newStateListOfVms[index]["href"] != "" {
				vm, err := vapp.GetVmByName(newVms[index]["name"].(string))
				if err != nil {
					return err

				}
				newStateListOfVms[index]["href"] = vm.HREF
			}

			vmSubResource := NewVirtualMachineSubresource(newStateListOfVms[index], oldStateListOfVms[index], 0)

			err := configureVM(vmSubResource, meta)

			if err != nil {
				return err
			}
			// newVMResources[index] = vmResourceAfterConfiguration
			newStateListOfVms[index] = vmSubResource.Data()
		}

		d.Set("vm", newStateListOfVms)
	}

	// TODO: MAybe remove this coupling
	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
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

	// Get VMs and create descriptions for the vAppCompose
	oldState, newState := d.GetChange("vm")

	oldStateMap := mapStringInterfaceToMapStringMapStringInterface(oldState.(map[string]interface{}))
	newStateMap := mapStringInterfaceToMapStringMapStringInterface(newState.(map[string]interface{}))

	for key := range newStateMap {

		vmSubResource := NewVirtualMachineSubresource(newStateMap[key], oldStateMap[key], 0)

		err := readVM(vmSubResource, meta)

		if err != nil {
			return err
		}
		// newVMResources[index] = vmResourceAfterConfiguration
		newStateMap[key] = vmSubResource.Data()
	}

	d.Set("vm", newStateMap)
	return nil
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

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

	return err
}
