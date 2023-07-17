package vcd

//lint:file-ignore SA1019 ignore deprecated functions
// This code relies on using `d.GetOkExists` which is provided by terraform-plugin-sdk.
// The use case for using this function is to have tristate boolean (to separate between an `empty`
// and `false` values)
// This field is deprecated in SDK to discourage its usage, but is not going to be removed without a
// proper solution. Only an experimental option exists now.
// More information in https://github.com/hashicorp/terraform-plugin-sdk/issues/817
import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// lookupvAppTemplateforVm will do the following
// evaluate if optional parameter `vm_name_in_template` was specified.
//
// If `vm_name_in_template` was specified
// * It will look up the exact VM with given `vm_name_in_template` inside `vapp_template_id` (or deprecated `template_name` and catalog
// `catalog_name`)
//
// If `vm_name_in_template` was not specified:
// * It will look up vApp template with ID `vapp_template_id` (or deprecated `template_name` in catalog `catalog_name`)
// * After it is found - it will pick the first child VM template
func lookupvAppTemplateforVm(d *schema.ResourceData, vcdClient *VCDClient, org *govcd.Org, vdc *govcd.Vdc) (govcd.VAppTemplate, error) {
	vAppTemplateId, vAppTemplateIdSet := d.GetOk("vapp_template_id")
	if vAppTemplateIdSet {
		// Lookup of vApp Template using URN
		vAppTemplate, err := vcdClient.GetVAppTemplateById(vAppTemplateId.(string))
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error finding vApp Template with URN %s: %s", vAppTemplateId.(string), err)
		}

		if vmNameInTemplate, ok := d.GetOk("vm_name_in_template"); ok { // specific VM name in template is given
			vmInTemplateRecord, err := vcdClient.QuerySynchronizedVmInVAppTemplateByHref(vAppTemplate.VAppTemplate.HREF, vmNameInTemplate.(string))
			if err != nil {
				return govcd.VAppTemplate{}, fmt.Errorf("error obtaining VM '%s' inside vApp Template: %s", vmNameInTemplate, err)
			}
			returnedVAppTemplate, err := vcdClient.GetVAppTemplateByHref(vmInTemplateRecord.HREF)
			if err != nil {
				return govcd.VAppTemplate{}, fmt.Errorf("error getting vApp template from inner VM %s: %s", vmNameInTemplate, err)
			}
			return *returnedVAppTemplate, err
		} else {
			if vAppTemplate.VAppTemplate == nil || vAppTemplate.VAppTemplate.Children == nil || len(vAppTemplate.VAppTemplate.Children.VM) == 0 {
				return govcd.VAppTemplate{}, fmt.Errorf("the vApp Template %s doesn't contain any usable VM inside", vAppTemplateId)
			}
			returnedVAppTemplate := govcd.NewVAppTemplate(&vcdClient.Client)
			returnedVAppTemplate.VAppTemplate = vAppTemplate.VAppTemplate.Children.VM[0]
			return *returnedVAppTemplate, nil
		}
	} else {
		// Deprecated way of looking up the vApp Template
		catalogName := d.Get("catalog_name").(string)
		templateName := d.Get("template_name").(string)

		catalog, err := org.GetCatalogByName(catalogName, false)
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog %s: %s", catalogName, err)
		}

		var vappTemplateHref string
		if vmNameInTemplate, ok := d.GetOk("vm_name_in_template"); ok { // specific VM name in template is given
			vmInTemplateRecord, err := vdc.QueryVappSynchronizedVmTemplate(catalogName, templateName, vmNameInTemplate.(string))
			if err != nil {
				return govcd.VAppTemplate{}, fmt.Errorf("error querying VM template %s: %s", vmNameInTemplate, err)
			}

			vappTemplateHref = vmInTemplateRecord.HREF
		} else { // No specific `vm_name_in_template` was given - will pick first item in vApp template
			catalogItem, err := catalog.GetCatalogItemByName(templateName, false)
			if err != nil {
				return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog item %s: %s", templateName, err)
			}
			vappTemplate, err := catalogItem.GetVAppTemplate()
			if err != nil {
				return govcd.VAppTemplate{}, fmt.Errorf("[VM create] error finding VAppTemplate %s: %s", templateName, err)
			}

			if vappTemplate.VAppTemplate == nil || vappTemplate.VAppTemplate.Children == nil || len(vappTemplate.VAppTemplate.Children.VM) == 0 {
				return govcd.VAppTemplate{}, fmt.Errorf("error finding VM template")
			}
			vappTemplateHref = vappTemplate.VAppTemplate.Children.VM[0].HREF

		}

		returnedVappTemplate, err := catalog.GetVappTemplateByHref(vappTemplateHref)
		if err != nil || returnedVappTemplate == nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error retrieving Catalog Item by HREF: '%s'", err)
		}

		return *returnedVappTemplate, nil
	}
}

func lookupStorageProfile(d *schema.ResourceData, vdc *govcd.Vdc) (*types.Reference, error) {
	// If no storage profile lookup was requested - bail out early and return nil reference
	storageProfileName := d.Get("storage_profile").(string)
	if storageProfileName == "" {
		return nil, nil
	}

	storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
	if err != nil {
		return nil, fmt.Errorf("[vm creation] error retrieving storage profile %s : %s", storageProfileName, err)
	}

	return &storageProfile, nil

}

// lookupComputePolicy returns the Compute Policy associated to the value of the given Compute Policy attribute. If the
// attribute is not set, the returned policy will be nil. If the obtained policy is incorrect, it will return an error.
func lookupComputePolicy(d *schema.ResourceData, vcdClient *VCDClient, computePolicyAttribute string) (*govcd.VdcComputePolicyV2, error) {
	if value, ok := d.GetOk(computePolicyAttribute); ok {
		computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(value.(string))
		if err != nil {
			return nil, fmt.Errorf("error getting compute policy %s: %s", value.(string), err)
		}
		if computePolicy.Href == "" {
			return nil, fmt.Errorf("empty compute policy HREF detected")
		}
		return computePolicy, nil
	}
	return nil, nil
}

// getCpuMemoryValues returns CPU, CPU core count and Memory variables. Priority comes from HCL
// schema configuration and then whatever is present in compute policy (if it was specified at all)
func getCpuMemoryValues(d *schema.ResourceData, vdcComputePolicy *types.VdcComputePolicyV2) (*int, *int, *int64, error) {

	var (
		setCpu    *int
		setCores  *int
		setMemory *int64
	)

	// If VDC Compute policy is not set - we're specifying CPU and Memory directly
	// if vdcComputePolicy == nil {
	if memory, isMemorySet := d.GetOk("memory"); isMemorySet {
		memInt := int64(memory.(int))
		setMemory = &memInt
	}

	if cpus, isCpusSet := d.GetOk("cpus"); isCpusSet {
		cpuInt := cpus.(int)
		setCpu = &cpuInt
	}

	if cpuCores, isCpuCoresSet := d.GetOk("cpu_cores"); isCpuCoresSet {
		cpuCoresInt := cpuCores.(int)
		setCores = &cpuCoresInt
	}

	// Check if sizing policy has any settings settings and override VM configuration with it
	if vdcComputePolicy != nil {
		if vdcComputePolicy.Memory != nil {
			mem := int64(*vdcComputePolicy.Memory)
			setMemory = &mem
		}

		if vdcComputePolicy.CPUCount != nil {
			setCpu = vdcComputePolicy.CPUCount
		}

		if vdcComputePolicy.CoresPerSocket != nil {
			setCores = vdcComputePolicy.CoresPerSocket
		}
	}

	return setCpu, setCores, setMemory, nil
}

func updateAdvancedComputeSettings(d *schema.ResourceData, vm *govcd.VM) error {
	vmSpecSection := vm.VM.VmSpecSection
	description := vm.VM.Description

	// DiskSection must be nil because leaving it with actual configuration values will try to
	// update and fail
	vmSpecSection.DiskSection = nil

	updateNeeded := false

	if memorySharesLevel, ok := d.GetOk("memory_priority"); ok {
		vmSpecSection.MemoryResourceMb.SharesLevel = memorySharesLevel.(string)
		updateNeeded = true
	}

	if memoryLimit, ok := d.GetOk("memory_limit"); ok {
		vmSpecSection.MemoryResourceMb.Limit = addrOf(int64(memoryLimit.(int)))
		updateNeeded = true
	}

	if memoryShares, ok := d.GetOk("memory_shares"); ok {
		vmSpecSection.MemoryResourceMb.Shares = addrOf(memoryShares.(int))
		updateNeeded = true
	}

	if memoryReservation, ok := d.GetOk("memory_reservation"); ok {
		vmSpecSection.MemoryResourceMb.Reservation = addrOf(int64(memoryReservation.(int)))
		updateNeeded = true
	}

	if memorySharesLevel, ok := d.GetOk("cpu_priority"); ok {
		vmSpecSection.CpuResourceMhz.SharesLevel = memorySharesLevel.(string)
		updateNeeded = true
	}

	if memoryLimit, ok := d.GetOk("cpu_limit"); ok {
		vmSpecSection.CpuResourceMhz.Limit = addrOf(int64(memoryLimit.(int)))
		updateNeeded = true
	}

	if memoryShares, ok := d.GetOk("cpu_shares"); ok {
		vmSpecSection.CpuResourceMhz.Shares = addrOf(memoryShares.(int))
		updateNeeded = true
	}

	if memoryReservation, ok := d.GetOk("cpu_reservation"); ok {
		vmSpecSection.CpuResourceMhz.Reservation = addrOf(int64(memoryReservation.(int)))
		updateNeeded = true
	}

	if updateNeeded {
		err := updateVmSpecSection(vmSpecSection, vm, description)
		if err != nil {
			return fmt.Errorf("error updating advanced compute settings: %s", err)
		}
	}

	// Refresh VM to ensure that latest VM structure is used in other function calls
	err := vm.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing VM after updating advanced compute settings: %s", err)
	}

	return nil
}

// isItVappNetwork checks if it is a vApp network (not vApp Org Network)
func isItVappNetwork(vAppNetworkName string, vapp govcd.VApp) (bool, error) {
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error getting vApp networks: %s", err)
	}

	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vAppNetworkName &&
			govcd.IsVappNetwork(networkConfig.Configuration) {
			log.Printf("[TRACE] vApp network found: %s", vAppNetworkName)
			return true, nil
		}
	}

	return false, fmt.Errorf("configured vApp network isn't found: %s", vAppNetworkName)
}

type diskParams struct {
	name       string
	busNumber  *int
	unitNumber *int
}

func expandDisksProperties(v interface{}) ([]diskParams, error) {
	v = v.(*schema.Set).List()
	l := v.([]interface{})
	diskParamsArray := make([]diskParams, 0, len(l))

	for _, raw := range l {
		if raw == nil {
			continue
		}
		original := raw.(map[string]interface{})
		addParams := diskParams{name: original["name"].(string)}

		busNumber := original["bus_number"].(string)
		if busNumber != "" {
			convertedBusNumber, err := strconv.Atoi(busNumber)
			if err != nil {
				return nil, fmt.Errorf("value `%s` bus_number is not number. err: %s", busNumber, err)
			}
			addParams.busNumber = &convertedBusNumber
		}

		unitNumber := original["unit_number"].(string)
		if unitNumber != "" {
			convertedUnitNumber, err := strconv.Atoi(unitNumber)
			if err != nil {
				return nil, fmt.Errorf("value `%s` unit_number is not number. err: %s", unitNumber, err)
			}
			addParams.unitNumber = &convertedUnitNumber
		}

		diskParamsArray = append(diskParamsArray, addParams)
	}
	return diskParamsArray, nil
}

// getVmIndependentDisks iterates over VirtualHardwareSection of VM and returns slice of independent (named) disk references
//
// item.ResourceType == 17 is a Hard disk
// item.HostResource[0].Disk provides the distinction between VM disks and Independent (named)
// disks. For VM disks the `item.HostResource[0].Disk` is empty, while for independent (named) disk
// it contains a reference to that named disk.
// Sample HostResource value for:
// * VM disk: <rasd:HostResource ns10:storageProfileHref="https://HOST/api/vdcStorageProfile/ef9fd5e2-a417-4e63-9c30-60bc8a8f53d0" ns10:busType="6" ns10:busSubType="VirtualSCSI" ns10:capacity="16384" ns10:iops="0" ns10:storageProfileOverrideVmDefault="false"></rasd:HostResource>
// * Independent (named) disk: <rasd:HostResource ns10:storageProfileHref="https://HOST/api/vdcStorageProfile/ef9fd5e2-a417-4e63-9c30-60bc8a8f53d0" ns10:disk="https://HOST/api/disk/55da11ad-967a-43ba-a744-81e3c86b4b9e" ns10:busType="6" ns10:busSubType="VirtualSCSI" ns10:capacity="100" ns10:iops="0" ns10:storageProfileOverrideVmDefault="true"></rasd:HostResource>
func getVmIndependentDisks(vm govcd.VM) []string {

	var disks []string
	// We use VirtualHardwareSection because in time of implementation we didn't have access to VmSpecSection which we used for internal disks.
	for _, item := range vm.VM.VirtualHardwareSection.Item {
		// disk resource type is 17
		if item.ResourceType == 17 && item.HostResource[0].Disk != "" {
			disks = append(disks, item.HostResource[0].Disk)
		}
	}
	return disks
}

func addRemoveGuestProperties(d *schema.ResourceData, vm *govcd.VM) error {
	if d.HasChange("guest_properties") {
		vmProperties, err := getGuestProperties(d)
		if err != nil {
			return fmt.Errorf("unable to convert guest properties to data structure")
		}

		log.Printf("[TRACE] Updating VM guest properties")
		_, err = vm.SetProductSectionList(vmProperties)
		if err != nil {
			return fmt.Errorf("error setting guest properties: %s", err)
		}
	}
	return nil
}

// isPrimaryNicRemoved checks if the updated schema has a primary NIC at all
func isPrimaryNicRemoved(d *schema.ResourceData) bool {
	_, newNetworkRaw := d.GetChange("network")
	newNetworks := newNetworkRaw.([]interface{})

	var foundPrimaryNic bool
	for _, newNet := range newNetworks {
		netMap := newNet.(map[string]interface{})
		isPrimary := netMap["is_primary"].(bool)
		if isPrimary {
			foundPrimaryNic = true
			break
		}
	}

	return !foundPrimaryNic
}

func updateVmSpecSection(vmSpecSection *types.VmSpecSection, vm *govcd.VM, description string) error {
	// add missing values if not inherited from template, otherwise API throws error if some value is nil
	if vmSpecSection.MemoryResourceMb.Reservation == nil {
		vmSpecSection.MemoryResourceMb.Reservation = addrOf(int64(0))
	}
	if vmSpecSection.MemoryResourceMb.Limit == nil {
		vmSpecSection.MemoryResourceMb.Limit = addrOf(int64(-1))
	}
	if vmSpecSection.MemoryResourceMb.SharesLevel == "" {
		vmSpecSection.MemoryResourceMb.SharesLevel = "NORMAL"
	}
	if vmSpecSection.CpuResourceMhz.Reservation == nil {
		vmSpecSection.CpuResourceMhz.Reservation = addrOf(int64(0))
	}
	if vmSpecSection.CpuResourceMhz.Limit == nil {
		vmSpecSection.CpuResourceMhz.Limit = addrOf(int64(-1))
	}
	if vmSpecSection.CpuResourceMhz.SharesLevel == "" {
		vmSpecSection.CpuResourceMhz.SharesLevel = "NORMAL"
	}
	_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
	if err != nil {
		return fmt.Errorf("error updating Vm Spec Section: %s", err)
	}
	return nil
}

// getVmFromResource retrieves a VM by using HCL schema configuration
// It returns VM and its parent structures:
// * VCDClient
// * Org
// * VDC
// * vApp
// * VM
func getVmFromResource(d *schema.ResourceData, meta interface{}, vmType typeOfVm) (*VCDClient, *govcd.Org, *govcd.Vdc, *govcd.VApp, string, *govcd.VM, error) {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, nil, nil, nil, "", nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)

	if err != nil {
		additionalMessage := ""
		if vmType == standaloneVmType {
			additionalMessage = fmt.Sprintf("\nAdding a vApp name to a standalone VM is not allowed." +
				"Please use 'vcd_vapp_vm' resource to specify vApp")
			dSet(d, "vapp_name", "")
		}

		return nil, nil, nil, nil, "", nil, fmt.Errorf("[getVmFromResource] error finding vApp '%s': %s%s", vappName, err, additionalMessage)
	}

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, nil, nil, nil, "", nil, fmt.Errorf("[VM update] neither name or ID was set")
	}

	vm, err := vapp.GetVMByNameOrId(identifier, false)

	if err != nil {
		d.SetId("")
		return nil, nil, nil, nil, "", nil, fmt.Errorf("[VM update] error getting VM %s: %s", identifier, err)
	}
	return vcdClient, org, vdc, vapp, identifier, vm, nil
}

// attachDetachIndependentDisks updates attached disks to latest state, removes not needed, and adds
// new ones
func attachDetachIndependentDisks(d *schema.ResourceData, vm govcd.VM, vdc *govcd.Vdc) error {
	oldValues, newValues := d.GetChange("disk")

	attachDisks := newValues.(*schema.Set).Difference(oldValues.(*schema.Set))
	detachDisks := oldValues.(*schema.Set).Difference(newValues.(*schema.Set))

	removeDiskProperties, err := expandDisksProperties(detachDisks)
	if err != nil {
		return err
	}

	for _, diskData := range removeDiskProperties {
		disk, err := vdc.QueryDisk(diskData.name)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", diskData.name, err)
		}

		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}
		if diskData.unitNumber != nil {
			attachParams.UnitNumber = diskData.unitNumber
		}
		if diskData.busNumber != nil {
			attachParams.BusNumber = diskData.busNumber
		}

		task, err := vm.DetachDisk(attachParams)
		if err != nil {
			return fmt.Errorf("error detaching disk `%s` to vm %s", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete detaching disk `%s` to vm %s", diskData.name, err)
		}
	}

	// attach new independent disks
	newDiskProperties, err := expandDisksProperties(attachDisks)
	if err != nil {
		return err
	}

	sort.SliceStable(newDiskProperties, func(i, j int) bool {
		if newDiskProperties[i].busNumber == newDiskProperties[j].busNumber {
			return *newDiskProperties[i].unitNumber > *newDiskProperties[j].unitNumber
		}
		return *newDiskProperties[i].busNumber > *newDiskProperties[j].busNumber
	})

	for _, diskData := range newDiskProperties {
		disk, err := vdc.QueryDisk(diskData.name)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", diskData.name, err)
		}

		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}
		if diskData.unitNumber != nil {
			attachParams.UnitNumber = diskData.unitNumber
		}
		if diskData.busNumber != nil {
			attachParams.BusNumber = diskData.busNumber
		}

		task, err := vm.AttachDisk(attachParams)
		if err != nil {
			return fmt.Errorf("error attaching disk `%s` to vm %s", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete attaching disk `%s` to vm %s", diskData.name, err)
		}
	}
	return nil
}

func updateStateOfAttachedIndependentDisks(d *schema.ResourceData, vm govcd.VM) error {

	existingDisks := getVmIndependentDisks(vm)
	transformed := schema.NewSet(resourceVcdVmIndependentDiskHash, []interface{}{})

	for _, existingDiskHref := range existingDisks {
		diskSettings, err := getIndependentDiskFromVmDisks(vm, existingDiskHref)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", existingDiskHref, err)
		}
		newValues := map[string]interface{}{
			"name":        diskSettings.Disk.Name,
			"bus_number":  strconv.Itoa(diskSettings.BusNumber),
			"unit_number": strconv.Itoa(diskSettings.UnitNumber),
			"size_in_mb":  diskSettings.SizeMb,
		}

		transformed.Add(newValues)
	}

	return d.Set("disk", transformed)
}

// getIndependentDiskFromVmDisks finds independent disk in VM disk list.
func getIndependentDiskFromVmDisks(vm govcd.VM, diskHref string) (*types.DiskSettings, error) {
	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return nil, govcd.ErrorEntityNotFound
	}
	for _, disk := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if disk.Disk != nil && disk.Disk.HREF == diskHref {
			return disk, nil
		}
	}
	return nil, govcd.ErrorEntityNotFound
}

func updateStateOfInternalDisks(d *schema.ResourceData, vm govcd.VM) error {
	err := vm.Refresh()
	if err != nil {
		return err
	}

	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return fmt.Errorf("[updateStateOfInternalDisks] VmSpecSection part is missing")
	}
	existingInternalDisks := vm.VM.VmSpecSection.DiskSection.DiskSettings
	var internalDiskList []map[string]interface{}
	for _, internalDisk := range existingInternalDisks {
		// API shows internal disk and independent disks in one list. If disk.Disk != nil then it's independent disk
		// We use VmSpecSection as it is newer type than VirtualHardwareSection. It is used by HTML5 vCD client, has easy understandable structure.
		// VirtualHardwareSection has undocumented relationships between elements and very hard to use without issues for internal disks.
		if internalDisk.Disk == nil {
			newValue := map[string]interface{}{
				"disk_id":         internalDisk.DiskId,
				"bus_type":        internalDiskBusTypesFromValues[internalDisk.AdapterType],
				"size_in_mb":      int(internalDisk.SizeMb),
				"bus_number":      internalDisk.BusNumber,
				"unit_number":     internalDisk.UnitNumber,
				"storage_profile": internalDisk.StorageProfile.Name,
			}

			// There have been real cases where these values were `nil` and caused panic of plugin.
			if internalDisk.Iops != nil {
				newValue["iops"] = int(*internalDisk.Iops)
			}
			if internalDisk.ThinProvisioned != nil {
				newValue["thin_provisioned"] = *internalDisk.ThinProvisioned
			}

			internalDiskList = append(internalDiskList, newValue)
		}
	}

	return d.Set("internal_disk", internalDiskList)
}

func updateTemplateInternalDisks(d *schema.ResourceData, meta interface{}, vm govcd.VM) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return fmt.Errorf("[updateTemplateInternalDisks] VmSpecSection part is missing")
	}

	diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings

	var storageProfilePrt *types.Reference
	var overrideVmDefault bool

	internalDisksList := d.Get("override_template_disk").(*schema.Set).List()

	if len(internalDisksList) == 0 {
		return nil
	}

	for _, internalDisk := range internalDisksList {
		internalDiskProvidedConfig := internalDisk.(map[string]interface{})
		diskCreatedByTemplate := getMatchedDisk(internalDiskProvidedConfig, diskSettings)

		storageProfileName := internalDiskProvidedConfig["storage_profile"].(string)
		if storageProfileName != "" {
			storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
			if err != nil {
				return fmt.Errorf("[vm creation] error retrieving storage profile %s : %s", storageProfileName, err)
			}
			storageProfilePrt = &storageProfile
			overrideVmDefault = true
		} else {
			storageProfilePrt = vm.VM.StorageProfile
			overrideVmDefault = false
		}

		if diskCreatedByTemplate == nil {
			return fmt.Errorf("[vm creation] disk with bus type %s, bus number %d and unit number %d not found",
				internalDiskProvidedConfig["bus_type"].(string), internalDiskProvidedConfig["bus_number"].(int), internalDiskProvidedConfig["unit_number"].(int))
		}

		// Update details of internal disk for disk existing in template
		if value, ok := internalDiskProvidedConfig["iops"]; ok {
			iops := int64(value.(int))
			diskCreatedByTemplate.Iops = &iops
		}

		// value is required but not treated.
		isThinProvisioned := true
		diskCreatedByTemplate.ThinProvisioned = &isThinProvisioned

		diskCreatedByTemplate.SizeMb = int64(internalDiskProvidedConfig["size_in_mb"].(int))
		diskCreatedByTemplate.StorageProfile = storageProfilePrt
		diskCreatedByTemplate.OverrideVmDefault = overrideVmDefault
	}

	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings
	_, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return fmt.Errorf("error updating VM disks: %s", err)
	}

	return nil
}

// getMatchedDisk returns matched disk by adapter type, bus number and unit number
func getMatchedDisk(internalDiskProvidedConfig map[string]interface{}, diskSettings []*types.DiskSettings) *types.DiskSettings {
	for _, diskSetting := range diskSettings {
		if diskSetting.AdapterType == internalDiskBusTypes[internalDiskProvidedConfig["bus_type"].(string)] &&
			diskSetting.BusNumber == internalDiskProvidedConfig["bus_number"].(int) &&
			diskSetting.UnitNumber == internalDiskProvidedConfig["unit_number"].(int) {
			return diskSetting
		}
	}
	return nil
}

// getGuestProperties returns a struct for setting guest properties
func getGuestProperties(d *schema.ResourceData) (*types.ProductSectionList, error) {
	guestProperties := d.Get("guest_properties")
	guestProp := convertToStringMap(guestProperties.(map[string]interface{}))
	vmProperties := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info:     "Custom properties",
			Property: []*types.Property{},
		},
	}
	for key, value := range guestProp {
		log.Printf("[TRACE] Adding guest property: key=%s, value=%s to object", key, value)
		oneProp := &types.Property{
			UserConfigurable: true,
			Type:             "string",
			Key:              key,
			Label:            key,
			Value:            &types.Value{Value: value},
		}
		vmProperties.ProductSection.Property = append(vmProperties.ProductSection.Property, oneProp)
	}

	return vmProperties, nil
}

// setGuestProperties sets guest properties into state
func setGuestProperties(d *schema.ResourceData, properties *types.ProductSectionList) error {
	data := make(map[string]string)

	// if properties object does not have actual properties - do not set it at all (leave Terraform 'null')
	log.Printf("[TRACE] Setting empty properties into statefile because no properties were specified")
	if properties == nil || properties.ProductSection == nil || len(properties.ProductSection.Property) == 0 {
		return nil
	}

	for _, prop := range properties.ProductSection.Property {
		// if a value was set - use it
		if prop.Value != nil {
			data[prop.Key] = prop.Value.Value
		}
	}

	log.Printf("[TRACE] Setting properties into statefile")
	return d.Set("guest_properties", data)
}

// isForcedCustomization checks "customization" block in resource and checks if the value of field "force"
// is set to "true". It returns false if the value is not set or is set to false
func isForcedCustomization(customizationBlock interface{}) bool {
	customizationSlice := customizationBlock.([]interface{})

	if len(customizationSlice) != 1 {
		return false
	}

	cust := customizationSlice[0]
	fc := cust.(map[string]interface{})
	forceCust, ok := fc["force"]
	forceCustBool := forceCust.(bool)

	if !ok || !forceCustBool {
		return false
	}

	return true
}

func resourceVcdVmIndependentDiskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	_, err := buf.WriteString(fmt.Sprintf("%s-",
		m["name"].(string)))
	// We use the name and no other identifier to calculate the hash
	// With the VM resource, we assume that disks have a unique name.
	// In the event that this is not true, we return an error
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	return hashcodeString(buf.String())
}

// networksToConfig converts terraform schema for 'network' to types.NetworkConnectionSection
// which is used for creating new VM
//
// The `vapp` parameter does not play critical role in the code, but adds additional validations:
// * `org` type of networks will be checked if they are already attached to the vApp
// * `vapp` type networks will be checked for existence inside the vApp
func networksToConfig(d *schema.ResourceData, vapp *govcd.VApp) (types.NetworkConnectionSection, error) {
	networks := d.Get("network").([]interface{})

	isStandaloneVm := vapp == nil || (vapp != nil && vapp.VApp.IsAutoNature)
	networkConnectionSection := types.NetworkConnectionSection{}

	// sets existing primary network connection index. Further code changes index only if change is
	// found
	for index, singleNetwork := range networks {
		nic := singleNetwork.(map[string]interface{})
		isPrimary := nic["is_primary"].(bool)
		if isPrimary {
			networkConnectionSection.PrimaryNetworkConnectionIndex = index
		}
	}

	for index, singleNetwork := range networks {

		nic := singleNetwork.(map[string]interface{})
		netConn := &types.NetworkConnection{}

		networkName := nic["name"].(string)
		if networkName == "" {
			networkName = "none"
		}
		ipAllocationMode := nic["ip_allocation_mode"].(string)
		ip := nic["ip"].(string)
		macAddress, macIsSet := nic["mac"].(string)

		isPrimary := nic["is_primary"].(bool)
		nicHasPrimaryChange := d.HasChange("network." + strconv.Itoa(index) + ".is_primary")
		if nicHasPrimaryChange && isPrimary {
			networkConnectionSection.PrimaryNetworkConnectionIndex = index
		}

		networkType := nic["type"].(string)
		if networkType == "org" && !isStandaloneVm {
			isVappOrgNetwork, err := isItVappOrgNetwork(networkName, *vapp)
			if err != nil {
				return types.NetworkConnectionSection{}, err
			}
			if !isVappOrgNetwork {
				return types.NetworkConnectionSection{}, fmt.Errorf("vApp Org network : %s is not found", networkName)
			}
		}
		if networkType == "vapp" && !isStandaloneVm {
			isVappNetwork, err := isItVappNetwork(networkName, *vapp)
			if err != nil {
				return types.NetworkConnectionSection{}, fmt.Errorf("unable to find vApp network %s: %s", networkName, err)
			}
			if !isVappNetwork {
				return types.NetworkConnectionSection{}, fmt.Errorf("vApp network : %s is not found", networkName)
			}
		}

		netConn.IsConnected = nic["connected"].(bool)
		netConn.IPAddressAllocationMode = ipAllocationMode
		netConn.NetworkConnectionIndex = index
		netConn.Network = networkName
		if macIsSet {
			netConn.MACAddress = macAddress
		}

		if ipAllocationMode == types.IPAllocationModeNone {
			netConn.Network = types.NoneNetwork
		}

		if net.ParseIP(ip) != nil {
			netConn.IPAddress = ip
		}

		adapterType, isSetAdapterType := nic["adapter_type"]
		if isSetAdapterType {
			netConn.NetworkAdapterType = adapterType.(string)
		}

		networkConnectionSection.NetworkConnection = append(networkConnectionSection.NetworkConnection, netConn)
	}
	return networkConnectionSection, nil
}

// isItVappOrgNetwork checks if it is a vApp Org network (not vApp Network)
func isItVappOrgNetwork(vAppNetworkName string, vapp govcd.VApp) (bool, error) {
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error getting vApp networks: %s", err)
	}

	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vAppNetworkName &&
			!govcd.IsVappNetwork(networkConfig.Configuration) {
			log.Printf("[TRACE] vApp Org network found: %s", vAppNetworkName)
			return true, nil
		}
	}

	return false, fmt.Errorf("configured vApp Org network isn't found: %s", vAppNetworkName)
}

// getVmNicIndexesWithDhcpEnabled loops over VMs NICs and returns list of indexes for the ones using
// DHCP
func getVmNicIndexesWithDhcpEnabled(networkConnectionSection *types.NetworkConnectionSection) []int {

	var nicIndexes []int

	// Sort NIC cards by their virtual slot numbers as the API returns them in random order
	sort.SliceStable(networkConnectionSection.NetworkConnection, func(i, j int) bool {
		return networkConnectionSection.NetworkConnection[i].NetworkConnectionIndex <
			networkConnectionSection.NetworkConnection[j].NetworkConnectionIndex
	})

	for nicIndex, singleNic := range networkConnectionSection.NetworkConnection {

		// validate if the NIC is suitable for DHCP waiting (has DHCP interface)
		if singleNic.IPAddressAllocationMode != types.IPAllocationModeDHCP {
			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] NIC '%d' is not using DHCP in 'ip_allocation_mode'. Skipping IP wait", nicIndex)
			continue
		}
		log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] NIC '%d' is using DHCP in 'ip_allocation_mode'.", nicIndex)
		nicIndexes = append(nicIndexes, singleNic.NetworkConnectionIndex)

	}
	return nicIndexes
}

// getVmByName returns a VM by the given name if found unequivocally
// If there are more than one instance by the wanted name, it also returns a list of
// matching VMs with sample information (ID, guest OS, network name, IP address)
func getVmByName(client *VCDClient, vdc *govcd.Vdc, name string) (*govcd.VM, string, error) {

	vmList, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
	if err != nil {
		return nil, "", err
	}

	var foundList []*types.QueryResultVMRecordType
	for _, vm := range vmList {
		if vm.Name == name {
			foundList = append(foundList, vm)
		}
	}

	if len(foundList) == 0 {
		return nil, "", govcd.ErrorEntityNotFound
	}
	if len(foundList) == 1 {
		vm, err := client.Client.GetVMByHref(foundList[0].HREF)
		if err != nil {
			return nil, "", err
		}
		return vm, "", nil
	}
	// More than one element found for the given name. Returning the list
	listStr := fmt.Sprintf("%-50s %-30s %s\n", "ID", "Guest OS", "Network")
	listStr += fmt.Sprintf("%-50s %-30s %s\n", strings.Repeat("-", 50), strings.Repeat("-", 30), strings.Repeat("-", 20))
	for _, vm := range foundList {
		id := extractUuid(vm.HREF)
		networkInfo := ""
		if vm.NetworkName != "" || vm.IpAddress != "" {
			networkInfo = fmt.Sprintf("(%s - %s)", vm.NetworkName, vm.IpAddress)
		}
		listStr += fmt.Sprintf("urn:vcloud:vm:%s %-30s %s\n", id, vm.GuestOS, networkInfo)
	}
	return nil, listStr, fmt.Errorf("more than one VM found with name %s", name)
}

// updateGuestCustomizationSetting is responsible for setting all the data related to VM customization
func updateGuestCustomizationSetting(d *schema.ResourceData, vm *govcd.VM) error {
	// Retrieve existing customization section to only customize what was throughout this function
	customizationSection, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return fmt.Errorf("error getting existing customization section before changing: %s", err)
	}

	// for back compatibility we allow to set computer name from `name` if computer_name isn't provided
	var computerName string
	if cName, ok := d.GetOk("computer_name"); ok {
		computerName = cName.(string)
	} else {
		computerName = d.Get("name").(string)
	}

	if _, isSetComputerName := d.GetOk("computer_name"); isSetComputerName {
		customizationSection.ComputerName = computerName
	}

	// Process parameters from 'customization' block
	updateCustomizationSection(d.Get("customization"), d, customizationSection)

	// Apply any of the settings we have set
	if _, err = vm.SetGuestCustomizationSection(customizationSection); err != nil {
		return fmt.Errorf("error applying guest customization details: %s", err)
	}

	return nil
}

func updateCustomizationSection(customizationInterface interface{}, d *schema.ResourceData, customizationSection *types.GuestCustomizationSection) {
	customizationSlice := customizationInterface.([]interface{})
	if len(customizationSlice) == 1 {
		cust := customizationSlice[0]
		if cust != nil {

			if enabled, isSetEnabled := d.GetOkExists("customization.0.enabled"); isSetEnabled {
				customizationSection.Enabled = addrOf(enabled.(bool))
			}
			if initScript, isSetInitScript := d.GetOkExists("customization.0.initscript"); isSetInitScript {
				customizationSection.CustomizationScript = initScript.(string)
			}

			if changeSid, isSetChangeSid := d.GetOkExists("customization.0.change_sid"); isSetChangeSid {
				customizationSection.ChangeSid = addrOf(changeSid.(bool))
			}

			if allowLocalAdminPasswd, isSetAllowLocalAdminPasswd := d.GetOkExists("customization.0.allow_local_admin_password"); isSetAllowLocalAdminPasswd {
				customizationSection.AdminPasswordEnabled = addrOf(allowLocalAdminPasswd.(bool))

			}

			if mustChangeOnFirstLogin, isSetMustChangeOnFirstLogin := d.GetOkExists("customization.0.must_change_password_on_first_login"); isSetMustChangeOnFirstLogin {
				customizationSection.ResetPasswordRequired = addrOf(mustChangeOnFirstLogin.(bool))
			}

			if autoGeneratePasswd, isSetAutoGeneratePasswd := d.GetOkExists("customization.0.auto_generate_password"); isSetAutoGeneratePasswd {
				customizationSection.AdminPasswordAuto = addrOf(autoGeneratePasswd.(bool))
			}

			if adminPasswd, isSetAdminPasswd := d.GetOkExists("customization.0.admin_password"); isSetAdminPasswd {
				customizationSection.AdminPassword = adminPasswd.(string)
				// customizationSection.AdminPasswordEnabled = addrOf(true)
			}

			if nrTimesForLogin, isSetNrTimesForLogin := d.GetOkExists("customization.0.number_of_auto_logons"); isSetNrTimesForLogin {
				// The AdminAutoLogonEnabled is "hidden" from direct user input to behave exactly like UI does. UI sets
				// the value of this field behind the scenes based on number_of_auto_logons count.
				// AdminAutoLogonEnabled=false if number_of_auto_logons == 0
				// AdminAutoLogonEnabled=true if number_of_auto_logons > 0
				isMoreThanZero := nrTimesForLogin.(int) > 0
				customizationSection.AdminAutoLogonEnabled = &isMoreThanZero

				customizationSection.AdminAutoLogonCount = nrTimesForLogin.(int)
			}

			if joinDomain, isSetJoinDomain := d.GetOkExists("customization.0.join_domain"); isSetJoinDomain {
				customizationSection.JoinDomainEnabled = addrOf(joinDomain.(bool))
			}

			if joinOrgDomain, isSetJoinOrgDomain := d.GetOkExists("customization.0.join_org_domain"); isSetJoinOrgDomain {
				customizationSection.UseOrgSettings = addrOf(joinOrgDomain.(bool))
			}

			if joinDomainName, isSetJoinDomainName := d.GetOkExists("customization.0.join_domain_name"); isSetJoinDomainName {
				customizationSection.DomainName = joinDomainName.(string)
			}

			if joinDomainUser, isSetJoinDomainUser := d.GetOkExists("customization.0.join_domain_user"); isSetJoinDomainUser {
				customizationSection.DomainUserName = joinDomainUser.(string)
			}

			if joinDomainPasswd, isSetJoinDomainPasswd := d.GetOkExists("customization.0.join_domain_password"); isSetJoinDomainPasswd {
				customizationSection.DomainUserPassword = joinDomainPasswd.(string)
			}

			if joinDomainOu, isSetJoinDomainOu := d.GetOkExists("customization.0.join_domain_account_ou"); isSetJoinDomainOu {
				customizationSection.MachineObjectOU = joinDomainOu.(string)
			}

		}
	}
}

// setGuestCustomizationData is responsible for persisting all guest customization details into statefile
func setGuestCustomizationData(d *schema.ResourceData, vm *govcd.VM) error {
	customizationSection, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return fmt.Errorf("unable to get guest customization section: %s", err)
	}

	dSet(d, "computer_name", customizationSection.ComputerName)

	customizationBlock := make([]interface{}, 1)
	customizationBlockAttributes := make(map[string]interface{})

	customizationBlockAttributes["enabled"] = customizationSection.Enabled
	customizationBlockAttributes["change_sid"] = customizationSection.ChangeSid
	customizationBlockAttributes["allow_local_admin_password"] = customizationSection.AdminPasswordEnabled
	customizationBlockAttributes["must_change_password_on_first_login"] = customizationSection.ResetPasswordRequired
	customizationBlockAttributes["auto_generate_password"] = customizationSection.AdminPasswordAuto
	customizationBlockAttributes["admin_password"] = customizationSection.AdminPassword
	customizationBlockAttributes["number_of_auto_logons"] = customizationSection.AdminAutoLogonCount
	customizationBlockAttributes["join_domain"] = customizationSection.JoinDomainEnabled
	customizationBlockAttributes["join_org_domain"] = customizationSection.UseOrgSettings
	customizationBlockAttributes["join_domain_name"] = customizationSection.DomainName
	customizationBlockAttributes["join_domain_user"] = customizationSection.DomainUserName
	customizationBlockAttributes["join_domain_password"] = customizationSection.DomainUserPassword
	customizationBlockAttributes["join_domain_account_ou"] = customizationSection.MachineObjectOU
	customizationBlockAttributes["initscript"] = customizationSection.CustomizationScript

	customizationBlock[0] = customizationBlockAttributes

	err = d.Set("customization", customizationBlock)
	if err != nil {
		return fmt.Errorf("")
	}

	return nil
}

// handleExposeHardwareVirtualization toggles hardware virtualization according to
// `expose_hardware_virtualization` field value.
func handleExposeHardwareVirtualization(d *schema.ResourceData, newVm *govcd.VM) error {
	// The operation below assumes the VM is powered off and does not check for status because the
	// VM is being powered on in the last stage of create/update cycle
	if d.Get("expose_hardware_virtualization").(bool) {
		task, err := newVm.ToggleHardwareVirtualization(true)
		if err != nil {
			return fmt.Errorf("error enabling hardware assisted virtualization: %s", err)
		}
		err = task.WaitTaskCompletion()

		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}
	}
	return nil
}

// readNetworks returns network configuration for saving into statefile
func readNetworks(d *schema.ResourceData, vm govcd.VM, vapp govcd.VApp, vdc *govcd.Vdc) ([]map[string]interface{}, error) {
	// Determine type for all networks in vApp
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return []map[string]interface{}{}, fmt.Errorf("error getting vApp networks: %s", err)
	}
	// If vApp network is "isolated" and has no ParentNetwork - it is a vApp network.
	// https://code.vmware.com/apis/72/vcloud/doc/doc/types/NetworkConfigurationType.html
	vAppNetworkTypes := make(map[string]string)
	for _, netConfig := range vAppNetworkConfig.NetworkConfig {
		switch {
		case netConfig.NetworkName == types.NoneNetwork:
			vAppNetworkTypes[netConfig.NetworkName] = types.NoneNetwork
		case govcd.IsVappNetwork(netConfig.Configuration):
			vAppNetworkTypes[netConfig.NetworkName] = "vapp"
		default:
			vAppNetworkTypes[netConfig.NetworkName] = "org"
		}
	}

	var nets []map[string]interface{}
	// Sort NIC cards by their virtual slot numbers as the API returns them in random order
	sort.SliceStable(vm.VM.NetworkConnectionSection.NetworkConnection, func(i, j int) bool {
		return vm.VM.NetworkConnectionSection.NetworkConnection[i].NetworkConnectionIndex <
			vm.VM.NetworkConnectionSection.NetworkConnection[j].NetworkConnectionIndex
	})

	for _, vmNet := range vm.VM.NetworkConnectionSection.NetworkConnection {
		singleNIC := make(map[string]interface{})
		singleNIC["ip_allocation_mode"] = vmNet.IPAddressAllocationMode
		singleNIC["ip"] = vmNet.IPAddress
		singleNIC["mac"] = vmNet.MACAddress
		singleNIC["adapter_type"] = vmNet.NetworkAdapterType
		singleNIC["connected"] = vmNet.IsConnected
		if vmNet.Network != types.NoneNetwork {
			singleNIC["name"] = vmNet.Network
		}

		singleNIC["is_primary"] = false
		if vmNet.NetworkConnectionIndex == vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex {
			singleNIC["is_primary"] = true
		}

		var ok bool
		if singleNIC["type"], ok = vAppNetworkTypes[vmNet.Network]; !ok {
			// Prior vCD 10.1 used to return a placeholder for none networks. It allowed to identify
			// NIC type for types.NoneNetwork. This was removed in 10.1 therefore when vApp network
			// type has no details - the NIC network type is types.NoneNetwork
			singleNIC["type"] = types.NoneNetwork
		}

		nets = append(nets, singleNIC)
	}

	vmStatus, err := vm.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("unablet to check if VM is powered on: %s", err)
	}

	// If at least one`network_dhcp_wait_seconds` was defined
	if maxDhcpWaitSeconds, ok := d.GetOk("network_dhcp_wait_seconds"); ok && vmStatus == "POWERED_ON" {
		maxDhcpWaitSecondsInt := maxDhcpWaitSeconds.(int)

		// look up NIC indexes which have DHCP enabled
		dhcpNicIndexes := getVmNicIndexesWithDhcpEnabled(vm.VM.NetworkConnectionSection)
		log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] '%s' DHCP is used on NICs %v with wait time '%d seconds'",
			vm.VM.Name, dhcpNicIndexes, maxDhcpWaitSecondsInt)
		if len(dhcpNicIndexes) == 0 {
			logForScreen("vcd_vapp_vm", "INFO: Using 'network_dhcp_wait_seconds' only "+
				"makes sense if at least one NIC is using 'ip_allocation_mode=DHCP'\n")
		}

		if len(dhcpNicIndexes) > 0 { // at least one NIC uses DHCP for IP allocation mode
			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] '%s' waiting for DHCP IPs up to '%d' seconds on NICs %v",
				vm.VM.Name, maxDhcpWaitSeconds, dhcpNicIndexes)

			start := time.Now()

			// Only use DHCP lease check if it is NSX-V as NSX-T Edge Gateway does not expose it and errors on such query
			useNsxvDhcpLeaseCheck := vdc.IsNsxv()
			nicIps, timeout, err := vm.WaitForDhcpIpByNicIndexes(dhcpNicIndexes, maxDhcpWaitSecondsInt, useNsxvDhcpLeaseCheck)
			if err != nil {
				return nil, fmt.Errorf("unable to to look up DHCP IPs for VM NICs '%v': %s", dhcpNicIndexes, err)
			}

			if timeout {
				log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM %s timed out waiting %d seconds "+
					"to report DHCP IPs. You may want to increase 'network_dhcp_wait_seconds' or ensure "+
					"your DHCP settings are correct.\n", vm.VM.Name, maxDhcpWaitSeconds)
				logForScreen("vcd_vapp_vm", fmt.Sprintf("WARNING: VM %s timed out waiting %d seconds "+
					"to report DHCP IPs. You may want to increase 'network_dhcp_wait_seconds' or ensure "+
					"your DHCP settings are correct.", vm.VM.Name, maxDhcpWaitSeconds))
			}

			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM '%s' waiting for DHCP IPs took '%s' (of '%ds')",
				vm.VM.Name, time.Since(start), maxDhcpWaitSeconds)

			for sliceIndex, nicIndex := range dhcpNicIndexes {
				log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM '%s' NIC %d reported IP %s",
					vm.VM.Name, nicIndex, nicIps[sliceIndex])
				nets[nicIndex]["ip"] = nicIps[sliceIndex]
			}
		}
	}

	return nets, nil
}

func updateHardwareVersionAndOsType(d *schema.ResourceData, vm *govcd.VM) error {
	var err error
	var osTypeOrHardwareVersionChanged bool

	vmSpecSection := vm.VM.VmSpecSection
	if hardwareVersion := d.Get("hardware_version").(string); hardwareVersion != "" {
		vmSpecSection.HardwareVersion = &types.HardwareVersion{Value: hardwareVersion}
		osTypeOrHardwareVersionChanged = true
	}

	if osType := d.Get("os_type").(string); osType != "" {
		vmSpecSection.OsType = osType
		osTypeOrHardwareVersionChanged = true
	}

	if osTypeOrHardwareVersionChanged {
		_, err = vm.UpdateVmSpecSection(vmSpecSection, d.Get("description").(string))
		if err != nil {
			return fmt.Errorf("error changing VM spec section: %s", err)
		}
	}
	return nil
}

func createOrUpdateVmSecurityTags(d *schema.ResourceData, vm *govcd.VM) error {
	var err error
	entitySecurityTags := &types.EntitySecurityTags{}

	entitySecurityTagsFromSchema := d.Get("security_tags")
	entitySecurityTagsSlice := convertSchemaSetToSliceOfStrings(entitySecurityTagsFromSchema.(*schema.Set))
	entitySecurityTags.Tags = entitySecurityTagsSlice
	log.Printf("[DEBUG] Setting security_tags %s", entitySecurityTags)
	_, err = vm.UpdateVMSecurityTags(entitySecurityTags)

	if err != nil {
		return err
	}

	return nil
}
