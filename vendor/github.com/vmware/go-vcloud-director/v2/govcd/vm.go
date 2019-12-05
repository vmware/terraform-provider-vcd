/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type VM struct {
	VM     *types.VM
	client *Client
}

type VMRecord struct {
	VM     *types.QueryResultVMRecordType
	client *Client
}

func NewVM(cli *Client) *VM {
	return &VM{
		VM:     new(types.VM),
		client: cli,
	}
}

// create instance with reference to types.QueryResultVMRecordType
func NewVMRecord(cli *Client) *VMRecord {
	return &VMRecord{
		VM:     new(types.QueryResultVMRecordType),
		client: cli,
	}
}

func (vm *VM) GetStatus() (string, error) {
	err := vm.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %s", err)
	}
	return types.VAppStatuses[vm.VM.Status], nil
}

// IsDeployed checks if the VM is deployed or not
func (vm *VM) IsDeployed() (bool, error) {
	err := vm.Refresh()
	if err != nil {
		return false, fmt.Errorf("error refreshing VM: %s", err)
	}
	return vm.VM.Deployed, nil
}

func (vm *VM) Refresh() error {

	if vm.VM.HREF == "" {
		return fmt.Errorf("cannot refresh VM, Object is empty")
	}

	refreshUrl := vm.VM.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	vm.VM = &types.VM{}

	_, err := vm.client.ExecuteRequest(refreshUrl, http.MethodGet,
		"", "error refreshing VM: %s", nil, vm.VM)

	// The request was successful
	return err
}

// GetVirtualHardwareSection returns the virtual hardware items attached to a VM
func (vm *VM) GetVirtualHardwareSection() (*types.VirtualHardwareSection, error) {

	virtualHardwareSection := &types.VirtualHardwareSection{}

	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot refresh, invalid reference url")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/virtualHardwareSection/", http.MethodGet,
		types.MimeVirtualHardwareSection, "error retrieving virtual hardware: %s", nil, virtualHardwareSection)

	// The request was successful
	return virtualHardwareSection, err
}

// GetNetworkConnectionSection returns current networks attached to VM
//
// The slice of NICs is not necessarily ordered by NIC index
func (vm *VM) GetNetworkConnectionSection() (*types.NetworkConnectionSection, error) {

	networkConnectionSection := &types.NetworkConnectionSection{}

	if vm.VM.HREF == "" {
		return networkConnectionSection, fmt.Errorf("cannot retrieve network when VM HREF is unset")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/networkConnectionSection/", http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving network connection: %s", nil, networkConnectionSection)

	// The request was successful
	return networkConnectionSection, err
}

// UpdateNetworkConnectionSection applies network configuration of types.NetworkConnectionSection for the VM
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) UpdateNetworkConnectionSection(networks *types.NetworkConnectionSection) error {
	if vm.VM.HREF == "" {
		return fmt.Errorf("cannot update network connection when VM HREF is unset")
	}

	// Retrieve current network configuration so that we are not altering any other internal fields
	updateNetwork, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return fmt.Errorf("cannot read network section for update: %s", err)
	}
	updateNetwork.PrimaryNetworkConnectionIndex = networks.PrimaryNetworkConnectionIndex
	updateNetwork.NetworkConnection = networks.NetworkConnection
	updateNetwork.Ovf = types.XMLNamespaceOVF

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/networkConnectionSection/", http.MethodPut,
		types.MimeNetworkConnectionSection, "error updating network connection: %s", updateNetwork)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task completion after network update for vm %s: %s", vm.VM.Name, err)
	}

	return nil
}

// Deprecated: use client.GetVMByHref instead
func (cli *Client) FindVMByHREF(vmHREF string) (VM, error) {

	newVm := NewVM(cli)

	_, err := cli.ExecuteRequest(vmHREF, http.MethodGet,
		"", "error retrieving VM: %s", nil, newVm.VM)

	return *newVm, err

}

func (vm *VM) PowerOn() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOn"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering on VM: %s", nil)

}

// PowerOnAndForceCustomization is a synchronous function which is equivalent to the functionality
// one has in UI. It triggers customization which may be useful in some cases (like altering NICs)
//
// The VM _must_ be un-deployed for this action to actually work.
func (vm *VM) PowerOnAndForceCustomization() error {
	// PowerOnAndForceCustomization only works if the VM was previously un-deployed
	vmIsDeployed, err := vm.IsDeployed()
	if err != nil {
		return fmt.Errorf("unable to check if VM %s is un-deployed forcing customization: %s",
			vm.VM.Name, err)
	}

	if vmIsDeployed {
		return fmt.Errorf("VM %s must be undeployed before forcing customization", vm.VM.Name)
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/deploy"

	powerOnAndCustomize := &types.DeployVAppParams{
		Xmlns:              types.XMLNamespaceVCloud,
		PowerOn:            true,
		ForceCustomization: true,
	}

	task, err := vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering on VM with customization: %s", powerOnAndCustomize)

	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task completion after power on with customization %s: %s", vm.VM.Name, err)
	}

	return nil
}

func (vm *VM) PowerOff() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOff"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering off VM: %s", nil)
}

// Sets number of available virtual logical processors
// (i.e. CPUs x cores per socket)
// Cpu cores count is inherited from template.
// https://communities.vmware.com/thread/576209
func (vm *VM) ChangeCPUCount(virtualCpuCount int) (Task, error) {
	return vm.ChangeCPUCountWithCore(virtualCpuCount, nil)
}

// Sets number of available virtual logical processors
// (i.e. CPUs x cores per socket) and cores per socket.
// Socket count is a result of: virtual logical processors/cores per socket
// https://communities.vmware.com/thread/576209
func (vm *VM) ChangeCPUCountWithCore(virtualCpuCount int, coresPerSocket *int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	newCpu := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		XmlnsVmw:        types.XMLNamespaceVMW,
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/cpu",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(virtualCpuCount) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    types.ResourceTypeProcessor,
		VirtualQuantity: virtualCpuCount,
		Weight:          0,
		CoresPerSocket:  coresPerSocket,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/cpu"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing CPU count: %s", newCpu)

}

func (vm *VM) updateNicParameters(networks []map[string]interface{}, networkSection *types.NetworkConnectionSection) error {
	for tfNicSlot, network := range networks {
		for loopIndex := range networkSection.NetworkConnection {
			// Change network config only if we have the same virtual slot number as in .tf config
			if tfNicSlot == networkSection.NetworkConnection[loopIndex].NetworkConnectionIndex {

				// Determine what type of address is requested for the vApp
				var ipAllocationMode string
				ipAddress := "Any"

				var ipFieldString string
				ipField, ipIsSet := network["ip"]
				if ipIsSet {
					ipFieldString = ipField.(string)
				}

				switch {
				// TODO v3.0 remove from here when deprecated `ip` and `network_name` attributes are removed
				case ipIsSet && ipFieldString == "dhcp": // Deprecated ip="dhcp" mode
					ipAllocationMode = types.IPAllocationModeDHCP
				case ipIsSet && ipFieldString == "allocated": // Deprecated ip="allocated" mode
					ipAllocationMode = types.IPAllocationModePool
				case ipIsSet && ipFieldString == "none": // Deprecated ip="none" mode
					ipAllocationMode = types.IPAllocationModeNone

				// Deprecated ip="valid_ip" mode (currently it is hit by ip_allocation_mode=MANUAL as well)
				case ipIsSet && net.ParseIP(ipFieldString) != nil:
					ipAllocationMode = types.IPAllocationModeManual
					ipAddress = ipFieldString
				case ipIsSet && ipFieldString != "": // Deprecated ip="something_invalid" we default to DHCP. This is odd but backwards compatible.
					ipAllocationMode = types.IPAllocationModeDHCP
					// TODO v3.0 remove until here when deprecated `ip` and `network_name` attributes are removed

				case ipIsSet && net.ParseIP(ipFieldString) != nil && (network["ip_allocation_mode"].(string) == types.IPAllocationModeManual):
					ipAllocationMode = types.IPAllocationModeManual
					ipAddress = ipFieldString
				default: // New networks functionality. IP was not set and we're defaulting to provided ip_allocation_mode (only manual requires the IP)
					ipAllocationMode = network["ip_allocation_mode"].(string)
				}

				networkSection.NetworkConnection[loopIndex].NeedsCustomization = true
				networkSection.NetworkConnection[loopIndex].IsConnected = true
				networkSection.NetworkConnection[loopIndex].IPAddress = ipAddress
				networkSection.NetworkConnection[loopIndex].IPAddressAllocationMode = ipAllocationMode

				// for IPAllocationModeNone we hardcode special network name used by vcd 'none'
				if ipAllocationMode == types.IPAllocationModeNone {
					networkSection.NetworkConnection[loopIndex].Network = types.NoneNetwork
				} else {
					if _, ok := network["network_name"]; !ok {
						return fmt.Errorf("could not identify network name")
					}
					networkSection.NetworkConnection[loopIndex].Network = network["network_name"].(string)
				}

				// If we have one NIC only then it is primary by default, otherwise we check for "is_primary" key
				if (len(networks) == 1) || (network["is_primary"] != nil && network["is_primary"].(bool)) {
					networkSection.PrimaryNetworkConnectionIndex = tfNicSlot
				}
			}
		}
	}
	return nil
}

// ChangeNetworkConfig allows to update existing VM NIC configuration.f
func (vm *VM) ChangeNetworkConfig(networks []map[string]interface{}) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	networkSection, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return Task{}, fmt.Errorf("could not retrieve network connection for VM: %s", err)
	}

	err = vm.updateNicParameters(networks, networkSection)
	if err != nil {
		return Task{}, fmt.Errorf("failed processing NIC parameters: %s", err)
	}

	networkSection.Xmlns = types.XMLNamespaceVCloud
	networkSection.Ovf = types.XMLNamespaceOVF
	networkSection.Info = "Specifies the available VM network connections"

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/networkConnectionSection/"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeNetworkConnectionSection, "error changing network config: %s", networkSection)
}

func (vm *VM) ChangeMemorySize(size int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	newMem := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/memory",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "byte * 2^20",
		Description:     "Memory Size",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    types.ResourceTypeMemory,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/memory"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing memory size: %s", newMem)
}

func (vm *VM) RunCustomizationScript(computername, script string) (Task, error) {
	return vm.Customize(computername, script, false)
}

// GetGuestCustomizationStatus retrieves guest customization status.
// It can be one of "GC_PENDING", "REBOOT_PENDING", "GC_FAILED", "POST_GC_PENDING", "GC_COMPLETE"
func (vm *VM) GetGuestCustomizationStatus() (string, error) {
	guestCustomizationStatus := &types.GuestCustomizationStatusSection{}

	if vm.VM.HREF == "" {
		return "", fmt.Errorf("cannot retrieve guest customization, VM HREF is empty")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/guestcustomizationstatus", http.MethodGet,
		types.MimeGuestCustomizationStatus, "error retrieving guest customization status: %s", nil, guestCustomizationStatus)

	// The request was successful
	return guestCustomizationStatus.GuestCustStatus, err
}

// BlockWhileGuestCustomizationStatus blocks until the customization status of VM exits unwantedStatus.
// It sleeps 3 seconds between iterations and times out after timeOutAfterSeconds of seconds.
//
// timeOutAfterSeconds must be more than 4 and less than 2 hours (60s*120)
func (vm *VM) BlockWhileGuestCustomizationStatus(unwantedStatus string, timeOutAfterSeconds int) error {
	if timeOutAfterSeconds < 5 || timeOutAfterSeconds > 60*120 {
		return fmt.Errorf("timeOutAfterSeconds must be in range 4<X<7200")
	}

	timeoutAfter := time.After(time.Duration(timeOutAfterSeconds) * time.Second)
	tick := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timed out waiting for VM guest customization status to exit state %s after %d seconds",
				unwantedStatus, timeOutAfterSeconds)
		case <-tick.C:
			currentStatus, err := vm.GetGuestCustomizationStatus()
			if err != nil {
				return fmt.Errorf("could not get VM customization status %s", err)
			}
			if currentStatus != unwantedStatus {
				return nil
			}
		}
	}
}

func (vm *VM) Customize(computername, script string, changeSid bool) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   types.XMLNamespaceOVF,
		Xsi:   types.XMLNamespaceXSI,
		Xmlns: types.XMLNamespaceVCloud,

		HREF:                vm.VM.HREF,
		Type:                types.MimeGuestCustomizationSection,
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             true,
		ComputerName:        computername,
		CustomizationScript: script,
		ChangeSid:           false,
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/guestCustomizationSection/"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeGuestCustomizationSection, "error customizing VM: %s", vu)
}

// Undeploy triggers a VM undeploy and power off action. "Power off" action in UI behaves this way.
func (vm *VM) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               types.XMLNamespaceVCloud,
		UndeployPowerAction: "powerOff",
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/undeploy"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeUndeployVappParams, "error undeploy VM: %s", vu)
}

// Attach or detach an independent disk
// Use the disk/action/attach or disk/action/detach links in a Vm to attach or detach an independent disk.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) attachOrDetachDisk(diskParams *types.DiskAttachOrDetachParams, rel string) (Task, error) {
	util.Logger.Printf("[TRACE] Attach or detach disk, href: %s, rel: %s \n", diskParams.Disk.HREF, rel)

	var attachOrDetachDiskLink *types.Link
	for _, link := range vm.VM.Link {
		if link.Rel == rel && link.Type == types.MimeDiskAttachOrDetachParams {
			util.Logger.Printf("[TRACE] Attach or detach disk - found the proper link for request, HREF: %s, name: %s, type: %s, id: %s, rel: %s \n",
				link.HREF,
				link.Name,
				link.Type,
				link.ID,
				link.Rel)
			attachOrDetachDiskLink = link
		}
	}

	if attachOrDetachDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for attach or detach disk in disk Link")
	}

	diskParams.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return vm.client.ExecuteTaskRequest(attachOrDetachDiskLink.HREF, http.MethodPost,
		attachOrDetachDiskLink.Type, "error attach or detach disk: %s", diskParams)
}

// Attach an independent disk
// Call attachOrDetachDisk with disk and types.RelDiskAttach to attach an independent disk.
// Please verify the independent disk is not connected to any VM before calling this function.
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) AttachDisk(diskParams *types.DiskAttachOrDetachParams) (Task, error) {
	util.Logger.Printf("[TRACE] Attach disk, HREF: %s\n", diskParams.Disk.HREF)

	if diskParams.Disk == nil {
		return Task{}, fmt.Errorf("could not find disk info for attach")
	}

	return vm.attachOrDetachDisk(diskParams, types.RelDiskAttach)
}

// Detach an independent disk
// Call attachOrDetachDisk with disk and types.RelDiskDetach to detach an independent disk.
// Please verify the independent disk is connected the VM before calling this function.
// If the independent disk is not connected to the VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) DetachDisk(diskParams *types.DiskAttachOrDetachParams) (Task, error) {
	util.Logger.Printf("[TRACE] Detach disk, HREF: %s\n", diskParams.Disk.HREF)

	if diskParams.Disk == nil {
		return Task{}, fmt.Errorf("could not find disk info for detach")
	}

	return vm.attachOrDetachDisk(diskParams, types.RelDiskDetach)
}

// Helper function which finds media and calls InsertMedia
func (vm *VM) HandleInsertMedia(org *Org, catalogName, mediaName string) (Task, error) {

	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return Task{}, err
	}

	media, err := catalog.GetMediaByName(mediaName, false)
	if err != nil {
		return Task{}, err
	}

	return vm.InsertMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
			Name: media.Media.Name,
			ID:   media.Media.ID,
			Type: media.Media.Type,
		},
	})
}

// HandleEjectMediaAndAnswer helper function which finds media, calls EjectMedia, waits for task to complete and answer question.
// Also waits until VM status refreshes - this added as 9.7-10.0 vCD versions has lag in status update.
// answerYes - handles question risen when VM is running. True value enforces ejection.
func (vm *VM) HandleEjectMediaAndAnswer(org *Org, catalogName, mediaName string, answerYes bool) (*VM, error) {
	task, err := vm.HandleEjectMedia(org, catalogName, mediaName)
	if err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	err = task.WaitTaskCompletion(answerYes)
	if err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	for i := 0; i < 10; i++ {
		err = vm.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
		if !isMediaInjected(vm.VM.VirtualHardwareSection.Item) {
			return vm, nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil, fmt.Errorf("eject media executed but waiting for state update failed")
}

// check resource subtype for specific value which means media is injected
func isMediaInjected(items []*types.VirtualHardwareItem) bool {
	for _, hardwareItem := range items {
		if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
			return true
		}
	}
	return false
}

// Helper function which finds media and calls EjectMedia
func (vm *VM) HandleEjectMedia(org *Org, catalogName, mediaName string) (EjectTask, error) {
	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return EjectTask{}, err
	}

	media, err := catalog.GetMediaByName(mediaName, false)
	if err != nil {
		return EjectTask{}, err
	}

	task, err := vm.EjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
		},
	})

	return task, err
}

// Insert media for VM
// Call insertOrEjectMedia with media and types.RelMediaInsertMedia to insert media from VM.
func (vm *VM) InsertMedia(mediaParams *types.MediaInsertOrEjectParams) (Task, error) {
	util.Logger.Printf("[TRACE] Insert media, HREF: %s\n", mediaParams.Media.HREF)

	err := validateMediaParams(mediaParams)
	if err != nil {
		return Task{}, err
	}

	return vm.insertOrEjectMedia(mediaParams, types.RelMediaInsertMedia)
}

// Eject media from VM
// Call insertOrEjectMedia with media and types.RelMediaEjectMedia to eject media from VM.
// If media isn't inserted then task still will be successful.
func (vm *VM) EjectMedia(mediaParams *types.MediaInsertOrEjectParams) (EjectTask, error) {
	util.Logger.Printf("[TRACE] Detach disk, HREF: %s\n", mediaParams.Media.HREF)

	err := validateMediaParams(mediaParams)
	if err != nil {
		return EjectTask{}, err
	}

	task, err := vm.insertOrEjectMedia(mediaParams, types.RelMediaEjectMedia)
	if err != nil {
		return EjectTask{}, err
	}

	return *NewEjectTask(&task, vm), nil
}

// validates that media and media.href isn't empty
func validateMediaParams(mediaParams *types.MediaInsertOrEjectParams) error {
	if mediaParams.Media == nil {
		return fmt.Errorf("could not find media info for eject")
	}
	if mediaParams.Media.HREF == "" {
		return fmt.Errorf("could not find media HREF which is required for insert")
	}
	return nil
}

// Insert or eject a media for VM
// Use the vm/action/insert or vm/action/eject links in a Vm to insert or eject media.
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-InsertCdRom.html
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-EjectCdRom.html
func (vm *VM) insertOrEjectMedia(mediaParams *types.MediaInsertOrEjectParams, linkRel string) (Task, error) {
	util.Logger.Printf("[TRACE] Insert or eject media, href: %s, name: %s, , linkRel: %s \n", mediaParams.Media.HREF, mediaParams.Media.Name, linkRel)

	var insertOrEjectMediaLink *types.Link
	for _, link := range vm.VM.Link {
		if link.Rel == linkRel && link.Type == types.MimeMediaInsertOrEjectParams {
			util.Logger.Printf("[TRACE] Insert or eject media - found the proper link for request, HREF: %s, "+
				"name: %s, type: %s, id: %s, rel: %s \n", link.HREF, link.Name, link.Type, link.ID, link.Rel)
			insertOrEjectMediaLink = link
		}
	}

	if insertOrEjectMediaLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for insert or eject media")
	}

	mediaParams.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return vm.client.ExecuteTaskRequest(insertOrEjectMediaLink.HREF, http.MethodPost,
		insertOrEjectMediaLink.Type, "error insert or eject media: %s", mediaParams)
}

// Use the get existing VM question for operation which need additional response
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/GET-VmPendingQuestion.html
func (vm *VM) GetQuestion() (types.VmPendingQuestion, error) {

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/question"

	req := vm.client.NewRequest(map[string]string{}, http.MethodGet, *apiEndpoint, nil)

	resp, err := vm.client.Http.Do(req)

	// vCD security feature - on no question return 403 access error
	if http.StatusForbidden == resp.StatusCode {
		util.Logger.Printf("No question found for VM: %s\n", vm.VM.ID)
		return types.VmPendingQuestion{}, nil
	}

	if err != nil {
		return types.VmPendingQuestion{}, fmt.Errorf("error getting question: %s", err)
	}

	if http.StatusOK != resp.StatusCode {
		return types.VmPendingQuestion{}, fmt.Errorf("error getting question: %s", ParseErr(resp, &types.Error{}))
	}

	question := &types.VmPendingQuestion{}

	if err = decodeBody(resp, question); err != nil {
		return types.VmPendingQuestion{}, fmt.Errorf("error decoding question response: %s", err)
	}

	// The request was successful
	return *question, nil

}

// Use the provide answer to existing VM question for operation which need additional response
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-AnswerVmPendingQuestion.html
func (vm *VM) AnswerQuestion(questionId string, choiceId int) error {

	//validate input
	if questionId == "" {
		return fmt.Errorf("questionId can not be empty")
	}

	answer := &types.VmQuestionAnswer{
		Xmlns:      types.XMLNamespaceVCloud,
		QuestionId: questionId,
		ChoiceId:   choiceId,
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/question/action/answer"

	return vm.client.ExecuteRequestWithoutResponse(apiEndpoint.String(), http.MethodPost,
		"", "error asnwering question: %s", answer)
}

// ToggleHardwareVirtualization allows to either enable or disable hardware assisted
// CPU virtualization for the VM. It can only be performed on a powered off VM and
// will return an error otherwise. This is mainly useful for hypervisor nesting.
func (vm *VM) ToggleHardwareVirtualization(isEnabled bool) (Task, error) {
	vmStatus, err := vm.GetStatus()
	if err != nil {
		return Task{}, fmt.Errorf("unable to toggle hardware virtualization: %s", err)
	}
	if vmStatus != "POWERED_OFF" {
		return Task{}, fmt.Errorf("hardware virtualization can be changed from powered off state, status: %s", vmStatus)
	}

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	if isEnabled {
		apiEndpoint.Path += "/action/enableNestedHypervisor"
	} else {
		apiEndpoint.Path += "/action/disableNestedHypervisor"
	}
	errMessage := fmt.Sprintf("error toggling hypervisor nesting feature to %t for VM: %%s", isEnabled)
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", errMessage, nil)
}

// SetProductSectionList sets product section for a VM. It allows to change VM guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vm *VM) SetProductSectionList(productSection *types.ProductSectionList) (*types.ProductSectionList, error) {
	err := setProductSectionList(vm.client, vm.VM.HREF, productSection)
	if err != nil {
		return nil, fmt.Errorf("unable to set VM product section: %s", err)
	}

	return vm.GetProductSectionList()
}

// GetProductSectionList retrieves product section for a VM. It allows to read VM guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vm *VM) GetProductSectionList() (*types.ProductSectionList, error) {
	return getProductSectionList(vm.client, vm.VM.HREF)
}

// GetGuestCustomizationSection retrieves guest customization section for a VM. It allows to read VM guest customization properties.
func (vm *VM) GetGuestCustomizationSection() (*types.GuestCustomizationSection, error) {
	if vm == nil || vm.VM.HREF == "" {
		return nil, fmt.Errorf("vm or href cannot be empty to get  guest customization section")
	}
	guestCustomizationSection := &types.GuestCustomizationSection{}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/guestCustomizationSection", http.MethodGet,
		types.MimeGuestCustomizationSection, "error retrieving guest customization section : %s", nil, guestCustomizationSection)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve guest customization section: %s", err)
	}

	return guestCustomizationSection, nil
}

// SetGuestCustomizationSection sets guest customization section for a VM. It allows to change VM guest customization properties.
func (vm *VM) SetGuestCustomizationSection(guestCustomizationSection *types.GuestCustomizationSection) (*types.GuestCustomizationSection, error) {
	if vm == nil || vm.VM.HREF == "" {
		return nil, fmt.Errorf("vm or href cannot be empty to get  guest customization section")
	}

	guestCustomizationSection.Xmlns = types.XMLNamespaceVCloud
	guestCustomizationSection.Ovf = types.XMLNamespaceOVF

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/guestCustomizationSection", http.MethodPut,
		types.MimeGuestCustomizationSection, "error setting product section: %s", guestCustomizationSection)

	if err != nil {
		return nil, fmt.Errorf("unable to set guest customization section: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("task for setting guest customization section failed: %s", err)
	}

	return vm.GetGuestCustomizationSection()
}

// AddInternalDisk creates disk type *types.DiskSettings to the VM.
// Returns new disk ID and error.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) AddInternalDisk(diskData *types.DiskSettings) (string, error) {
	err := vm.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %s", err)
	}

	err = vm.validateInternalDiskInput(diskData)
	if err != nil {
		return "", err
	}

	var diskSettings []*types.DiskSettings
	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.DiskSection != nil && vm.VM.VmSpecSection.DiskSection.DiskSettings != nil {
		diskSettings = vm.VM.VmSpecSection.DiskSection.DiskSettings
	}

	diskSettings = append(diskSettings, diskData)
	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings

	vmSpecSection, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return "", err
	}

	for _, diskSetting := range vmSpecSection.DiskSection.DiskSettings {
		if diskSetting.AdapterType == diskData.AdapterType &&
			diskSetting.BusNumber == diskData.BusNumber &&
			diskSetting.UnitNumber == diskData.UnitNumber {
			return diskSetting.DiskId, nil
		}
	}

	return "", fmt.Errorf("created disk wasn't in list of returned VM internal disks")
}

func (vm *VM) validateInternalDiskInput(diskData *types.DiskSettings) error {
	if diskData.AdapterType == "" {
		return fmt.Errorf("disk settings missing required field: adapter type")
	}

	if diskData.BusNumber < 0 {
		return fmt.Errorf("disk settings bus number has to be 0 or higher")
	}

	if diskData.UnitNumber < 0 {
		return fmt.Errorf("disk settings unit number has to be 0 or higher")
	}

	if diskData.SizeMb < int64(0) {
		return fmt.Errorf("disk settings size MB has to be 0 or higher")
	}

	if diskData.Iops != nil && *diskData.Iops < int64(0) {
		return fmt.Errorf("disk settings iops has to be 0 or higher")
	}

	if diskData.ThinProvisioned == nil {
		return fmt.Errorf("disk settings missing required field: thin provisioned")
	}

	if diskData.StorageProfile == nil {
		return fmt.Errorf("disk settings missing required field: storage profile")
	}

	return nil
}

// GetInternalDiskId returns a valid *types.DiskSettings if it exists.
// If it doesn't, returns nil and ErrorEntityNotFound or other err.
func (vm *VM) GetInternalDiskId(diskId string, refresh bool) (*types.DiskSettings, error) {
	if refresh {
		err := vm.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing VM: %s", err)
		}
	}

	if diskId == "" {
		return nil, fmt.Errorf("cannot get internal disk - provided disk Id is empty")
	}

	if vm.VM.VmSpecSection.DiskSection == nil || vm.VM.VmSpecSection.DiskSection.DiskSettings == nil ||
		len(vm.VM.VmSpecSection.DiskSection.DiskSettings) == 0 {
		return nil, fmt.Errorf("cannot get internal disk - VM doesn't have internal disks")
	}

	for _, diskSetting := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if diskSetting.DiskId == diskId {
			return diskSetting, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// DeleteInternalDiskById delete disk using provided disk ID.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) DeleteInternalDiskById(diskId string) error {
	err := vm.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing VM: %s", err)
	}

	diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings
	if diskSettings == nil {
		diskSettings = []*types.DiskSettings{}
	}

	diskPlacement := -1
	for i, diskSetting := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if diskSetting.DiskId == diskId {
			diskPlacement = i
		}
	}

	if diskPlacement == -1 {
		return ErrorEntityNotFound
	}

	// remove disk in slice
	diskSettings = append(diskSettings[:diskPlacement], diskSettings[diskPlacement+1:]...)

	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings

	_, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return fmt.Errorf("error deleting VM %s internal disk %s: %s", vm.VM.Name, diskId, err)
	}

	return nil
}

// UpdateInternalDisks applies disks configuration for the VM.
// types.VmSpecSection requires consist of all disk entities which exist and not updated,
// as also new ones or changed ones. Returns new disk ID and error.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) UpdateInternalDisks(disksSettingToUpdate *types.VmSpecSection) (*types.VmSpecSection, error) {
	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot update internal disks - VM HREF is unset")
	}

	task, err := vm.UpdateInternalDisksAsync(disksSettingToUpdate)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting for task completion after internal disks update for VM %s: %s", vm.VM.Name, err)
	}
	err = vm.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing VM %s: %s", vm.VM.Name, err)
	}
	return vm.VM.VmSpecSection, nil
}

// UpdateInternalDisksAsync applies disks configuration and return task or err
// types.VmSpecSection requires consist of all disk entities which exist and not updated,
// as also new ones or changed ones.
func (vm *VM) UpdateInternalDisksAsync(disksSettingToUpdate *types.VmSpecSection) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update disks, VM HREF is unset")
	}

	disksSettingToUpdate.Info = "Virtual hardware requirements (simplified)"
	vmSpecSectionModified := true
	disksSettingToUpdate.Modified = &vmSpecSectionModified

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM disks: %s", &types.VMDiskChange{
			XMLName:       xml.Name{},
			Xmlns:         types.XMLNamespaceVCloud,
			Ovf:           types.XMLNamespaceOVF,
			Name:          vm.VM.Name,
			VmSpecSection: disksSettingToUpdate,
		})
}
