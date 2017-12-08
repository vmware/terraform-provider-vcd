package helper

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	types "github.com/ukcloud/govcloudair/types/v56"
	"github.com/terraform-providers/terraform-provider-vcd/vcd
)

func CreateVMDescription(vmData map[string]interface{}, vAppNetworks []string, meta interface{}) (*types.NewVMDescription, error) {
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

	nets := make([]*types.NetworkOrgDescription, len(networks))
	for index, n := range networks {

		network := n.(map[string]interface{})
		// Check if VM network is assigned to vApp
		if !isMember(vAppNetworks, network["name"].(string)) {
			return nil, fmt.Errorf("Network (%s) assigned to VM is not assigned to vApp, vApp has the following networks: %#v", network["name"].(string), vAppNetworks)
		}

		nets[index] = &types.NetworkOrgDescription{
			Name:             network["name"].(string),
			IsPrimary:        network["is_primary"].(bool),
			IsConnected:      network["is_connected"].(bool),
			IPAllocationMode: network["ip_allocation_mode"].(string),
			AdapterType:      network["adapter_type"].(string),
		}
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

func ConfigureVM(vmResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	vcdClient := meta.(*VCDClient)

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(vmResource["href"].(string))

	if err != nil {
		return nil, fmt.Errorf("Could not find VM (%s) in VCD", vmResource["href"].(string))
	}

	// TODO: Detect change in subResourceData
	// TODO: Power off/on logic

	// TODO: Use partial state
	//d.Partial(true)

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
	// name
	// ip
	// ip_allocation_mode
	// is_primary
	// is_connected
	// adapter_type

	// d.Partial(false)

	log.Printf("[TRACE] Done configureing %s, vmresource before reread: %#v", vmResource["name"], vmResource)
	vmResourceAfterReRead, err := readVM(vmResource, meta)

	if err != nil {
		return nil, err
	}

	return vmResourceAfterReRead, nil
}

func ReadVM(vmResource map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	vcdClient := meta.(*VCDClient)

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(vmResource["href"].(string))

	if err != nil {
		return nil, fmt.Errorf("Could not find VM (%s) in VCD", vmResource["href"].(string))
	}

	err = vm.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	log.Printf("[TRACE] Reading inforamtion of VM (%s)", vm.VM.Name)

	// Read network information
	log.Printf("[TRACE] Reading network information for vm (%s)", vm.VM.Name)
	networkConnections := vm.VM.NetworkConnectionSection.NetworkConnection
	readNetworks := make([]map[string]interface{}, len(networkConnections))

	primaryInterfaceIndex := vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex

	for index, networkConnection := range networkConnections {

		readNetwork := readNetwork(networkConnection, primaryInterfaceIndex)

		readNetworks[index] = readNetwork
	}

	// Read cpu count
	cpuCount, err := vm.GetCPUCount()
	if err != nil {
		return nil, err
	}

	// Read memory count
	memoryCount, err := vm.GetMemoryCount()
	if err != nil {
		return nil, err
	}

	// vmResource["vapp_href"] = vm.VM.VAppParent.HREF
	vmResource["name"] = vm.VM.Name
	vmResource["memory"] = memoryCount
	vmResource["cpus"] = cpuCount
	vmResource["network"] = readNetworks
	vmResource["nested_hypervisor_enabled"] = vm.VM.NestedHypervisorEnabled
	vmResource["href"] = vm.VM.HREF

	return vmResource, nil
}

func isMember(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}
	return false
}
