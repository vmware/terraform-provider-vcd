/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
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

func (cli *Client) FindVMByHREF(vmHREF string) (VM, error) {

	findUrl, err := url.ParseRequestURI(vmHREF)

	if err != nil {
		return VM{}, fmt.Errorf("error decoding vm HREF: %s", err)
	}

	// Querying the VApp
	req := cli.NewRequest(map[string]string{}, "GET", *findUrl, nil)

	resp, err := checkResp(cli.Http.Do(req))
	if err != nil {
		return VM{}, fmt.Errorf("error retrieving VM: %s", err)
	}

	newVm := NewVM(cli)

	if err = decodeBody(resp, newVm.VM); err != nil {
		return VM{}, fmt.Errorf("error decoding VM response: %s", err)
	}

	return *newVm, nil

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

// Helper function which finds media and calls InsertMedia
func (vm *VM) HandleInsertMedia(org *Org, catalogName, mediaName string) (Task, error) {

	media, err := FindMediaAsCatalogItem(org, catalogName, mediaName)
	if err != nil || media == (CatalogItem{}) {
		return Task{}, err
	}

	task, err := vm.InsertMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.CatalogItem.Entity.HREF,
			Name: media.CatalogItem.Entity.Name,
			ID:   media.CatalogItem.Entity.ID,
			Type: media.CatalogItem.Entity.Type,
		},
	})

	return task, err
}

// Helper function which finds media and calls EjectMedia
func (vm *VM) HandleEjectMedia(org *Org, catalogName, mediaName string) (Task, error) {
	media, err := FindMediaAsCatalogItem(org, catalogName, mediaName)
	if err != nil || media == (CatalogItem{}) {
		return Task{}, err
	}

	task, err := vm.EjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.CatalogItem.Entity.HREF,
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
func (vm *VM) EjectMedia(mediaParams *types.MediaInsertOrEjectParams) (Task, error) {
	util.Logger.Printf("[TRACE] Detach disk, HREF: %s\n", mediaParams.Media.HREF)

	err := validateMediaParams(mediaParams)
	if err != nil {
		return Task{}, err
	}

	vmStatus, err := vm.GetStatus()
	if err != nil {
		return Task{}, err
	}

	if vmStatus != types.VAppStatuses[8] {
		return Task{}, fmt.Errorf("to eject media, vm has to be in power off state")
	}

	return vm.insertOrEjectMedia(mediaParams, types.RelMediaEjectMedia)
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

	var err error
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

	reqUrl, err := url.ParseRequestURI(insertOrEjectMediaLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("could not parse request URL for insert or eject media. Error: %#v", err)
	}

	mediaParams.Xmlns = types.NsVCloud
	xmlPayload, err := xml.Marshal(mediaParams)
	if err != nil {
		return Task{}, fmt.Errorf("error marshal xml: %s", err)
	}

	reqPayload := bytes.NewBufferString(xml.Header + string(xmlPayload))
	req := vm.client.NewRequest(nil, http.MethodPost, *reqUrl, reqPayload)
	req.Header.Add("Content-Type", insertOrEjectMediaLink.Type)
	resp, err := checkResp(vm.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error insert or eject disk: %s", err)
	}

	task := NewTask(vm.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	return *task, nil
}
