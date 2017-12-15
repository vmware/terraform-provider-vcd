package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	govcd "github.com/ukcloud/govcloudair"
	types "github.com/ukcloud/govcloudair/types/v56"
)

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

	nets := make([]*types.NetworkOrgDescription, len(networks))
	for index, n := range networks {

		network := n.(map[string]interface{})
		// Check if VM network is assigned to vApp
		if !isStringMember(vAppNetworks, network["name"].(string)) {
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

func configureVM(vmResource *VirtualMachineSubresource, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(vmResource.Get("href").(string))

	if err != nil {
		return fmt.Errorf("Could not find VM (%s) in VCD", vmResource.Get("href").(string))
	}

	// TODO: Detect change in subResourceData
	// TODO: Power off/on logic

	// TODO: Use partial state
	//d.Partial(true)

	// Configure VM with initscript
	log.Printf("[TRACE] (%s) Configuring vm with initscript", vmResource.Get("name").(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.RunCustomizationScript(vmResource.Get("name").(string), vmResource.Get("initscript").(string))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error with setting init script: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error completing tasks: %#v", err)
	}

	// Change CPU count of VM
	log.Printf("[TRACE] (%s) Changing CPU", vmResource.Get("name").(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeCPUcount(vmResource.Get("cpus").(int))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing cpu count: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error completing task: %#v", err)
	}

	// Change Memory of VM
	log.Printf("[TRACE] (%s) Changing memory", vmResource.Get("name").(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeMemorySize(vmResource.Get("memory").(int))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing memory size: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return err
	}

	// Change nested hypervisor setting of VM
	log.Printf("[TRACE] (%s) Changing nested hypervisor setting", vmResource.Get("name").(string))
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeNestedHypervisor(vmResource.Get("nested_hypervisor_enabled").(bool))
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing nested hypervisor setting count: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error completing task: %#v", err)
	}

	// Change networks setting of VM
	log.Printf("[TRACE] (%s) Changing network settings", vmResource.Get("name").(string))

	networks := interfaceListToMapStringInterface(vmResource.Get("network").([]interface{}))

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := configureVmNetwork(networks, vm)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error changing nested hypervisor setting count: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error completing task: %#v", err)
	}

	// d.Partial(false)

	log.Printf("[TRACE] (%s) Done configuring %s, vmresource before reread: %#v", vmResource.Get("name").(string), vmResource.Get("href").(string), vmResource)
	err = readVM(vmResource, meta)

	if err != nil {
		return err
	}

	return nil
}

func readVM(vmResource *VirtualMachineSubresource, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	log.Printf("[TRACE] (%s) readVM got vmResource with href %s", vmResource.Get("name").(string), vmResource.Get("href").(string))

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(vmResource.Get("href").(string))

	if err != nil {
		return fmt.Errorf("Could not find VM (%s) in VCD", vmResource.Get("href").(string))
	}

	err = vm.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	log.Printf("[TRACE] (%s) Reading information of VM struct, href: (%s)", vm.VM.Name, vm.VM.HREF)
	log.Printf("[TRACE] (%s) Reading information of vmResource, href: (%s)", vmResource.Get("name").(string), vmResource.Get("href").(string))

	// Read network information
	log.Printf("[TRACE] Reading network information for vm (%s)", vm.VM.Name)
	networkConnections := vm.VM.NetworkConnectionSection.NetworkConnection
	readNetworks := make([]map[string]interface{}, len(networkConnections))

	primaryInterfaceIndex := vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex

	for index, networkConnection := range networkConnections {

		readNetwork := readVmNetwork(networkConnection, primaryInterfaceIndex)

		readNetworks[index] = readNetwork
	}

	// Read cpu count
	cpuCount, err := vm.GetCPUCount()
	if err != nil {
		return err
	}

	// Read memory count
	memoryCount, err := vm.GetMemoryCount()
	if err != nil {
		return err
	}

	// vmResource.Set("vapp_href", vm.VM.VAppParent.HREF)
	vmResource.Set("name", vm.VM.Name)
	vmResource.Set("memory", memoryCount)
	vmResource.Set("cpus", cpuCount)
	vmResource.Set("network", readNetworks)
	vmResource.Set("nested_hypervisor_enabled", vm.VM.NestedHypervisorEnabled)
	vmResource.Set("href", vm.VM.HREF)

	return nil
}

func configureVmNetwork(networkConnections []map[string]interface{}, vm govcd.VM) (govcd.Task, error) {

	err := vm.Refresh()
	if err != nil {
		return govcd.Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	var primaryNetworkConnectionIndex int
	newNetworkConnections := make([]*types.NetworkConnection, len(networkConnections))
	for index, network := range networkConnections {

		if network["is_primary"].(bool) {
			primaryNetworkConnectionIndex = index
		}

		newNetworkConnections[index] = &types.NetworkConnection{
			Network:                 network["name"].(string),
			NetworkConnectionIndex:  index,
			IsConnected:             network["is_connected"].(bool),
			IPAddressAllocationMode: network["ip_allocation_mode"].(string),
			NetworkAdapterType:      network["adapter_type"].(string),
		}
	}

	return vm.ChangeNetworkConfig(newNetworkConnections, primaryNetworkConnectionIndex)
}

func readVmNetwork(networkConnection *types.NetworkConnection, primaryInterfaceIndex int) map[string]interface{} {
	readNetwork := make(map[string]interface{})

	readNetwork["name"] = networkConnection.Network
	readNetwork["ip"] = networkConnection.IPAddress
	readNetwork["ip_allocation_mode"] = networkConnection.IPAddressAllocationMode
	readNetwork["is_primary"] = (primaryInterfaceIndex == networkConnection.NetworkConnectionIndex)
	readNetwork["is_connected"] = networkConnection.IsConnected
	readNetwork["adapter_type"] = networkConnection.NetworkAdapterType

	return readNetwork
}

func isStringMember(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}
	return false
}

func isVMMapStringInterfaceMember(list []map[string]interface{}, vm map[string]interface{}) bool {
	for _, item := range list {
		if item["href"] == vm["href"] {
			return true
		}
	}
	return false
}

func mapStringInterfaceToMapStringMapStringInterface(m map[string]interface{}) map[string]map[string]interface{} {
	newMap := make(map[string]map[string]interface{})
	for key, value := range m {
		newMap[key] = value.(map[string]interface{})
	}
	return newMap
}

func getKeys(m map[string]map[string]interface{}) []string {
	keys := make([]string, 0)
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
