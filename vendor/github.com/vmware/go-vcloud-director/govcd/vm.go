/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"

	types "github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
	"net/http"
)

type VM struct {
	VM     *types.VM
	client *Client
}

func NewVM(cli *Client) *VM {
	return &VM{
		VM:     new(types.VM),
		client: cli,
	}
}

func (vm *VM) GetStatus() (string, error) {
	err := vm.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %v", err)
	}
	return types.VAppStatuses[vm.VM.Status], nil
}

func (vm *VM) Refresh() error {

	if vm.VM.HREF == "" {
		return fmt.Errorf("cannot refresh VM, Object is empty")
	}

	refreshUrl, _ := url.ParseRequestURI(vm.VM.HREF)

	req := vm.client.NewRequest(map[string]string{}, "GET", *refreshUrl, nil)

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	vm.VM = &types.VM{}

	if err = decodeBody(resp, vm.VM); err != nil {
		return fmt.Errorf("error decoding task response VM: %s", err)
	}

	// The request was successful
	return nil
}

func (vm *VM) GetNetworkConnectionSection() (*types.NetworkConnectionSection, error) {

	networkConnectionSection := &types.NetworkConnectionSection{}

	if vm.VM.HREF == "" {
		return networkConnectionSection, fmt.Errorf("cannot refresh, Object is empty")
	}

	getNetworkUrl, _ := url.ParseRequestURI(vm.VM.HREF + "/networkConnectionSection/")

	req := vm.client.NewRequest(map[string]string{}, "GET", *getNetworkUrl, nil)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkConnectionSection+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return networkConnectionSection, fmt.Errorf("error retrieving task: %s", err)
	}

	if err = decodeBody(resp, networkConnectionSection); err != nil {
		return networkConnectionSection, fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return networkConnectionSection, nil
}

func (vdcCli *VCDClient) FindVMByHREF(vmhref string) (VM, error) {

	findUrl, err := url.ParseRequestURI(vmhref)

	if err != nil {
		return VM{}, fmt.Errorf("error decoding vm HREF: %s", err)
	}

	// Querying the VApp
	req := vdcCli.Client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

	resp, err := checkResp(vdcCli.Client.Http.Do(req))
	if err != nil {
		return VM{}, fmt.Errorf("error retrieving VM: %s", err)
	}

	newvm := NewVM(&vdcCli.Client)

	if err = decodeBody(resp, newvm.VM); err != nil {
		return VM{}, fmt.Errorf("error decoding VM response: %s", err)
	}

	return *newvm, nil

}

func (vm *VM) PowerOn() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOn"

	req := vm.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, nil)

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering on VM: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (vm *VM) PowerOff() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOff"

	req := vm.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, nil)

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering off VM: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (vm *VM) ChangeCPUcount(size int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	newcpu := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/cpu",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(size) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    3,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newcpu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/cpu"

	req := vm.client.NewRequest(map[string]string{}, "PUT", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (vm *VM) ChangeNetworkConfig(networks []map[string]interface{}, ip string) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	networksection, err := vm.GetNetworkConnectionSection()

	for index, network := range networks {
		// Determine what type of address is requested for the vApp
		ipAllocationMode := "NONE"
		ipAddress := "Any"

		// TODO: Review current behaviour of using DHCP when left blank
		if ip == "dhcp" || network["ip"].(string) == "dhcp" {
			ipAllocationMode = "DHCP"
		} else if ip == "allocated" || network["ip"].(string) == "allocated" {
			ipAllocationMode = "POOL"
		} else if ip == "none" || network["ip"].(string) == "none" {
			ipAllocationMode = "NONE"
		} else if ip != "" {
			ipAllocationMode = "MANUAL"
			// TODO: Check a valid IP has been given
			ipAddress = ip
		} else if network["ip"].(string) != "" {
			ipAllocationMode = "MANUAL"
			// TODO: Check a valid IP has been given
			ipAddress = network["ip"].(string)
		} else if ip == "" {
			ipAllocationMode = "DHCP"
		}

		util.Logger.Printf("[DEBUG] Function ChangeNetworkConfig() for %s invoked", network["orgnetwork"])

		networksection.Xmlns = "http://www.vmware.com/vcloud/v1.5"
		networksection.Ovf = "http://schemas.dmtf.org/ovf/envelope/1"
		networksection.Info = "Specifies the available VM network connections"

		networksection.NetworkConnection[index].NeedsCustomization = true
		networksection.NetworkConnection[index].IPAddress = ipAddress
		networksection.NetworkConnection[index].IPAddressAllocationMode = ipAllocationMode
		networksection.NetworkConnection[index].MACAddress = ""

		if network["is_primary"] == true {
			networksection.PrimaryNetworkConnectionIndex = index
		}

	}

	output, err := xml.MarshalIndent(networksection, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("[DEBUG] NetworkXML: %s", output)

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/networkConnectionSection/"

	req := vm.client.NewRequest(map[string]string{}, "PUT", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkConnectionSection+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM Network: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (vm *VM) ChangeMemorySize(size int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	newmem := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/memory",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "byte * 2^20",
		Description:     "Memory Size",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    4,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newmem, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/memory"

	req := vm.client.NewRequest(map[string]string{}, "PUT", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (vm *VM) RunCustomizationScript(computername, script string) (Task, error) {
	return vm.Customize(computername, script, false)
}

func (vm *VM) Customize(computername, script string, changeSid bool) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:   "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns: "http://www.vmware.com/vcloud/v1.5",

		HREF:                vm.VM.HREF,
		Type:                "application/vnd.vmware.vcloud.guestCustomizationSection+xml",
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             true,
		ComputerName:        computername,
		CustomizationScript: script,
		ChangeSid:           false,
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("[DEBUG] VCD Client configuration: %s", output)

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/guestCustomizationSection/"

	req := vm.client.NewRequest(map[string]string{}, "PUT", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.guestCustomizationSection+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (vm *VM) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               "http://www.vmware.com/vcloud/v1.5",
		UndeployPowerAction: "powerOff",
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/undeploy"

	req := vm.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.undeployVAppParams+xml")

	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error undeploy vApp: %s", err)
	}

	task := NewTask(vm.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

// Attach or detach an independent disk
// Use the disk/action/attach or disk/action/detach links in a Vm to attach or detach an independent disk.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) attachOrDetachDisk(diskParams *types.DiskAttachOrDetachParams, rel string) (Task, error) {
	util.Logger.Printf("[TRACE] Attach or detach disk, href: %s, rel: %s \n", diskParams.Disk.HREF, rel)

	var err error
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

	reqUrl, err := url.ParseRequestURI(attachOrDetachDiskLink.HREF)

	diskParams.Xmlns = types.NsVCloud

	xmlPayload, err := xml.Marshal(diskParams)
	if err != nil {
		return Task{}, fmt.Errorf("error marshal xml: %s", err)
	}

	// Send request
	reqPayload := bytes.NewBufferString(xml.Header + string(xmlPayload))
	req := vm.client.NewRequest(nil, http.MethodPost, *reqUrl, reqPayload)
	req.Header.Add("Content-Type", attachOrDetachDiskLink.Type)
	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error attach or detach disk: %s", err)
	}

	// Decode response
	task := NewTask(vm.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
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
