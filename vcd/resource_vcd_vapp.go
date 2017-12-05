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

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	networks := interfaceListToStringList(d.Get("network").([]interface{}))
	log.Printf("[TRACE] Networks from state: %#v", networks)

	// Get VMs and create descriptions for the vAppCompose
	vmResources := d.Get("vm").([]interface{})

	var vmDescriptions []*types.NewVMDescription
	for _, vm := range vmResources {
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

	var newVMResources []map[string]interface{}

	for index, vmData := range vmResources {
		vmResource := vmData.(map[string]interface{})

		href := vapp.VApp.Children.VM[index].HREF

		vmResourceAfterConfiguration, err := configureVM(vmResource, meta)

		if err != nil {
			return err
		}

		// Update vmResourceMap
		vmResource["href"] = href
		// vmResource["network"] = readVmNetwork()
		newVMResources = append(newVMResources, vmResourceAfterConfiguration)
	}

	d.SetId(d.Get("href").(string))
	d.Set("vm", newVMResources)

	// TODO: Remove this coupling
	// return resourceVcdVAppUpdate(d, meta)
	return nil
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// TODO: HERE WE MUST CHECK AND ADD NEW OR REMOVE VMs

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("name").(string))

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v, %s", err, status)
	}

	vmResources := d.Get("vm").([]interface{})
	var newVMResources []map[string]interface{}

	for _, vmData := range vmResources {
		vmResource := vmData.(map[string]interface{})

		vmResourceAfterConfiguration, err := configureVM(vmResource, meta)

		if err != nil {
			return err
		}

		newVMResources = append(newVMResources, vmResourceAfterConfiguration)
	}

	d.Set("vm", newVMResources)

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
	_, err = vcdClient.OrgVdc.FindVAppByName(d.Get("name").(string))
	if err != nil {
		log.Printf("[DEBUG] Unable to find vapp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	return nil
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// TODO: Delete logic
	// * Whole vApp
	// * List of VMs
	// * List of Network of VMs

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v", err)
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

func configureVM(vmResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	vcdClient := meta.(*VCDClient)

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(vmResource["href"].(string))

	if err != nil {
		return nil, fmt.Errorf("Could not find VM (%s) in VCD", vmResource["href"].(string))
	}

	// TODO: Detect change in subResourceData
	// TODO: Power off/on logic

	// Configure VM with initscript
	log.Printf("[TRACE] Configuring vm (%s) with initscript", vmResource["name"].(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.RunCustomizationScript(vmResource["name"].(string), vmResource["initscript"].(string))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error with setting init script: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return nil, fmt.Errorf("Error completing tasks: %#v", err)
	}

	// Change CPU count of VM
	log.Printf("[TRACE] Changing CPU of vm (%s)", vmResource["name"].(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeCPUcount(vmResource["cpus"].(int))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing cpu count: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return nil, fmt.Errorf("Error completing task: %#v", err)
	}

	// Change Memory of VM
	log.Printf("[TRACE] Changing memory of vm (%s)", vmResource["name"].(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeMemorySize(vmResource["memory"].(int))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing memory size: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return nil, err
	}

	// Change nested hypervisor setting of VM
	log.Printf("[TRACE] Changing nested hypervisor setting of vm (%s)", vmResource["name"].(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeNestedHypervisor(vmResource["nested_hypervisor_enabled"].(bool))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing nested hypervisor setting count: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return nil, fmt.Errorf("Error completing task: %#v", err)
	}

	// TODO: Network changes

	return vmResource, nil
}

func readVM(vmResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {

	return nil, nil
}

func configureNetwork(networkResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func readNetwork(networkResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return nil, nil
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
