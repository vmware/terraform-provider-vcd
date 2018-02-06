package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	govcd "github.com/ukcloud/govcloudair" // Forked from vmware/govcloudair
	types "github.com/ukcloud/govcloudair/types/v56"
)

// func createVMDescription() (*types.NewVMDescription, error) {

// 	log.Printf("[DEBUG] VAppTemplate: %#v", vapptemplate)

// 	vmDescription := &types.NewVMDescription{
// 		Name:         d.Get("name").(string),
// 		VAppTemplate: vapptemplate.VAppTemplate,
// 		Networks:     nets,
// 	}

// 	log.Printf("[DEBUG] NewVMDescription: %#v", vmDescription)

// 	return vmDescription, nil
// }

func composeSourceItem(d *schema.ResourceData, meta interface{}) (*types.SourcedCompositionItemParam, error) {
	vcdClient := meta.(*VCDClient)

	catalog, err := vcdClient.Org.FindCatalog(d.Get("catalog_name").(string))
	if err != nil {
		return nil, fmt.Errorf("Error finding catalog: %#v", err)
	}

	catalogitem, err := catalog.FindCatalogItem(d.Get("template_name").(string))
	if err != nil {
		return nil, fmt.Errorf("Error finding catalog item: %#v", err)
	}

	vapptemplate, err := catalogitem.GetVAppTemplate()
	if err != nil {
		return nil, fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}

	vm := govcd.NewVM(&vcdClient.Client)
	vm.VM = vapptemplate.VAppTemplate.Children.VM[0]

	// Remove the Network connections from the template
	vm.VM.NetworkConnectionSection.NetworkConnection = []*types.NetworkConnection{}

	configureVM(d, vm)

	sourceItem := &types.SourcedCompositionItemParam{
		Source: &types.Reference{
			HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
			Name: d.Get("name").(string),
		},
		InstantiationParams: &types.InstantiationParams{
			NetworkConnectionSection: vm.VM.NetworkConnectionSection,
			// GuestCustomizationSection: vm.VM.GuestCustomizationSection,
			OVFVirtualHardwareSection: vm.VM.VirtualHardwareSection.ConvertToOVF(),
		},
	}

	var storageProfile types.Reference

	if d.Get("storage_profile").(string) != "" {
		storageProfile, err = vcdClient.OrgVdc.FindStorageProfileReference(d.Get("storage_profile").(string))
		if err != nil {
			return nil, err
		}
	} else {
		storageProfile, err = vcdClient.OrgVdc.GetDefaultStorageProfileReference()
		if err != nil {
			return nil, err
		}
	}

	sourceItem.StorageProfile = &storageProfile

	return sourceItem, nil
}

func configureVM(d *schema.ResourceData, vm *govcd.VM) error {
	// vcdClient := meta.(*VCDClient)

	// Remove network hardware from virtualhw list.
	vm.RemoveVirtualHardwareItemByResourceType(types.ResourceTypeEthernet)

	// // Some changes requires the VM to be off or restarted
	// if d.HasChange("cpus") ||
	// 	d.HasChange("memory") ||
	// 	d.HasChange("nested_hypervisor_enabled") ||
	// 	d.HasChange("network") ||
	// 	d.HasChange("storage_profile") {

	log.Printf("[TRACE] (%s) Changing settings that require power off or restart", d.Get("name").(string))

	// Change CPU count of VM
	if d.HasChange("cpus") {
		log.Printf("[TRACE] (%s) Changing CPU", d.Get("name").(string))

		vm.SetCPUCount(d.Get("cpus").(int))
	}

	// Change Memory of VM
	if d.HasChange("memory") {
		log.Printf("[TRACE] (%s) Changing memory", d.Get("name").(string))

		vm.SetMemoryCount(d.Get("memory").(int))
	}

	// See func configureVMWorkaround
	// // Change nested hypervisor setting of VM
	// if d.HasChange("nested_hypervisor_enabled") {
	// 	log.Printf("[TRACE] (%s) Changing nested hypervisor setting", d.Get("name").(string))

	// 	// This cannot be reconfigured with reconfigureVM until vCloud 9.0
	// 	vm.SetNestedHypervisor(d.Get("nested_hypervisor_enabled").(bool))

	// }

	// Change networks setting of VM
	if d.HasChange("network") {
		log.Printf("[TRACE] (%s) Changing network settings", d.Get("name").(string))

		networks := interfaceListToMapStringInterface(d.Get("network").([]interface{}))
		vm.SetNetworkConnectionSection(createNetworkConnectionSection(networks))
	}

	// See func configureVMWorkaround
	// Change storage profile of VM
	if d.HasChange("storage_profile") {
		log.Printf("[TRACE] (%s) Changing storage profile", d.Get("name").(string))
		err := vm.SetStorageProfile(d.Get("storage_profile").(string))
		if err != nil {
			return err
		}
	}

	// vm.SetNeedsCustomization(true)
	// }

	// Here we need a powered on VM

	// Some changes requires the VM to be off or restarted
	// if d.HasChange("description") ||
	// 	d.HasChange("name") ||
	// 	d.HasChange("initscript") ||
	// 	d.HasChange("admin_password_auto") ||
	// 	d.HasChange("admin_password") {

	if d.HasChange("description") {
		log.Printf("[TRACE] (%s) Changing description", d.Get("name").(string))

		vm.SetDescription(d.Get("description").(string))
	}

	// Change hostname of VM
	if d.HasChange("name") {
		log.Printf("[TRACE] (%s) Changing hostname", d.Get("name").(string))

		vm.SetName(d.Get("name").(string))
		vm.SetHostName(d.Get("name").(string))
	}

	// Change nested hypervisor setting of VM
	if d.HasChange("initscript") {
		log.Printf("[TRACE] (%s) Changing initscript", d.Get("name").(string))

		vm.SetInitscript(d.Get("initscript").(string))
	}

	// Change nested hypervisor setting of VM
	if d.HasChange("admin_password_auto") {
		log.Printf("[TRACE] (%s) Changing admin_password_auto", d.Get("name").(string))

		vm.SetAdminPasswordAuto(d.Get("admin_password_auto").(bool))
	}

	// Change nested hypervisor setting of VM
	if d.HasChange("admin_password") {
		log.Printf("[TRACE] (%s) Changing admin_password", d.Get("name").(string))

		vm.SetAdminPassword(d.Get("admin_password").(string))
	}

	vm.SetNeedsCustomization(true)
	// }

	// log.Printf("[TRACE] (%s) Done configuring %s, d before reread: %#v", d.Get("name").(string), d.Get("href").(string), d)

	return nil
}

// Before vCloud 9.0, some elements cannot be configured by reconfigureVM,
// nestedhypervisor and storage profile, this has to be done in seperate calls
func configureVMWorkaround(d *schema.ResourceData, vm *govcd.VM, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Change nested hypervisor setting of VM
	if d.HasChange("nested_hypervisor_enabled") {
		log.Printf("[TRACE] (%s) Changing nested hypervisor setting", d.Get("name").(string))

		// This cannot be reconfigured with reconfigureVM until vCloud 9.0
		// vm.SetNestedHypervisor(d.Get("nested_hypervisor_enabled").(bool))

		// vCloud 8.2 and older
		err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vm.SetNestedHypervisorWithRequest(d.Get("nested_hypervisor_enabled").(bool))
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error setting nested hyperv VM: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("Error completing task: %#v", err)
		}
	}

	// // Change storage profile of VM
	// if d.HasChange("storage_profile") {
	// 	log.Printf("[TRACE] (%s) Changing storage profile", d.Get("name").(string))

	// }

	return nil
}

func readVM(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	log.Printf("[TRACE] (%s) readVM got d with href %s", d.Get("name").(string), d.Get("href").(string))

	// Get VM object from VCD
	vm, err := vcdClient.OrgVdc.GetVMByHREF(d.Get("href").(string))

	if err != nil {
		return fmt.Errorf("Could not find VM (%s)(%s) in VCD", d.Get("name").(string), d.Get("href").(string))
	}

	err = vm.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	log.Printf("[TRACE] (%s) Reading information of VM struct, href: (%s)", vm.VM.Name, vm.VM.HREF)
	log.Printf("[TRACE] (%s) Reading information of d, href: (%s)", d.Get("name").(string), d.Get("href").(string))

	// Read network information
	log.Printf("[TRACE] Reading network information for vm (%s)", vm.VM.Name)
	networkConnections := vm.VM.NetworkConnectionSection.NetworkConnection
	readNetworks := make([]map[string]interface{}, len(networkConnections))

	primaryInterfaceIndex := vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex

	for index, networkConnection := range networkConnections {

		readNetwork := readVMNetwork(networkConnection, primaryInterfaceIndex)

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

	// Read power state
	status, err := vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VM status: %#v", err)
	}

	// Check that the VM is powered off, and turn off if not.
	if status == types.VAppStatuses[4] {
		d.Set("power_on", true)
	}

	// d.Set("vapp_href", vm.VM.VAppParent.HREF)
	d.Set("name", vm.VM.Name)
	d.Set("memory", memoryCount)
	d.Set("cpus", cpuCount)
	d.Set("network", readNetworks)
	d.Set("nested_hypervisor_enabled", vm.VM.NestedHypervisorEnabled)
	d.Set("href", vm.VM.HREF)

	return nil
}

func createNetworkConnectionSection(networkConnections []map[string]interface{}) *types.NetworkConnectionSection {

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

	newNetwork := &types.NetworkConnectionSection{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Info:  "Specifies the available VM network connections",
		PrimaryNetworkConnectionIndex: primaryNetworkConnectionIndex,
		NetworkConnection:             newNetworkConnections,
	}

	return newNetwork
}

func readVMNetwork(networkConnection *types.NetworkConnection, primaryInterfaceIndex int) map[string]interface{} {
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

func getVMResourcebyHrefFromList(href string, list []map[string]interface{}) map[string]interface{} {
	for _, vm := range list {
		if vm["href"] == href {
			return vm
		}
	}
	return nil
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
