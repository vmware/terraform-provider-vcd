/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
	"net/http"
	"net/url"
)

// Independent disk
type Disk struct {
	Disk   *types.Disk
	client *Client
}

// Init independent disk struct
func NewDisk(cli *Client) *Disk {
	return &Disk{
		Disk:   new(types.Disk),
		client: cli,
	}
}

// While theoretically we can use smaller amounts, there is an issue when updating
// disks with size < 1MB
const MinimumDiskSize int = 1048576 // = 1Mb

// Create an independent disk in VDC
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 102 - 103,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vdc *Vdc) CreateDisk(diskCreateParams *types.DiskCreateParams) (Task, error) {
	util.Logger.Printf("[TRACE] Create disk, name: %s, size: %d \n",
		diskCreateParams.Disk.Name,
		diskCreateParams.Disk.Size,
	)

	if diskCreateParams.Disk.Name == "" {
		return Task{}, fmt.Errorf("disk name is required")
	}

	if diskCreateParams.Disk.Size < MinimumDiskSize {
		return Task{}, fmt.Errorf("disk size should be greater than or equal to 1Mb")
	}

	var err error
	var createDiskLink *types.Link

	// Find the proper link for request
	for _, vdcLink := range vdc.Vdc.Link {
		if vdcLink.Rel == types.RelAdd && vdcLink.Type == types.MimeDiskCreateParams {
			util.Logger.Printf("[TRACE] Create disk - found the proper link for request, HREF: %s, name: %s, type: %s, id: %s, rel: %s \n",
				vdcLink.HREF,
				vdcLink.Name,
				vdcLink.Type,
				vdcLink.ID,
				vdcLink.Rel)
			createDiskLink = vdcLink
			break
		}
	}

	if createDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for create disk in vdc Link")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(createDiskLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse URI: %s", err)
	}

	// Prepare the request payload
	diskCreateParams.Xmlns = types.NsVCloud

	xmlPayload, err := xml.Marshal(diskCreateParams)
	if err != nil {
		return Task{}, fmt.Errorf("error xml.Marshal: %s", err)
	}

	// Send Request
	reqPayload := bytes.NewBufferString(xml.Header + string(xmlPayload))
	req := vdc.client.NewRequest(nil, http.MethodPost, *reqUrl, reqPayload)
	req.Header.Add("Content-Type", createDiskLink.Type)
	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error create disk: %s", err)
	}

	// Decode response
	disk := NewDisk(vdc.client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return Task{}, fmt.Errorf("error decoding create disk params response: %s", err)
	}

	// Obtain disk task
	if disk.Disk.Tasks.Task == nil || len(disk.Disk.Tasks.Task) <= 0 {
		return Task{}, errors.New("error cannot find disk creation task in API response")
	}
	task := NewTask(vdc.client)
	task.Task = disk.Disk.Tasks.Task[0]

	util.Logger.Printf("[TRACE] AFTER CREATE DISK\n %s\n", prettyDisk(*disk.Disk))
	// Return the disk
	return *task, nil
}

// Update an independent disk
// 1 Verify the independent disk is not connected to any VM
// 2 Use newDiskInfo to change update the independent disk
// 3 Return task of independent disk update
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 104 - 106,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) Update(newDiskInfo *types.Disk) (Task, error) {
	util.Logger.Printf("[TRACE] Update disk, name: %s, size: %d, HREF: %s \n",
		newDiskInfo.Name,
		newDiskInfo.Size,
		d.Disk.HREF,
	)

	var err error

	if newDiskInfo.Name == "" {
		return Task{}, fmt.Errorf("disk name is required")
	}

	if newDiskInfo.Size < MinimumDiskSize {
		return Task{}, fmt.Errorf("disk size should be greater than or equal to 1Mb")
	}

	// Verify the independent disk is not connected to any VM
	vmRef, err := d.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var updateDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelEdit && diskLink.Type == types.MimeDisk {
			util.Logger.Printf("[TRACE] Update disk - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)
			updateDiskLink = diskLink
			break
		}
	}

	if updateDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for update disk in disk Link")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(updateDiskLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse URI: %s", err)
	}

	// Prepare the request payload
	xmlPayload, err := xml.Marshal(&types.Disk{
		Xmlns:          types.NsVCloud,
		Description:    newDiskInfo.Description,
		Size:           newDiskInfo.Size,
		Name:           newDiskInfo.Name,
		StorageProfile: newDiskInfo.StorageProfile,
		Owner:          newDiskInfo.Owner,
	})
	if err != nil {
		return Task{}, fmt.Errorf("error xml.Marshal: %s", err)
	}
	util.Logger.Printf("[TRACE] BEFORE UPDATE DISK\n %s\n", prettyDisk(*d.Disk))

	// Send request
	reqPayload := bytes.NewBufferString(xml.Header + string(xmlPayload))
	req := d.client.NewRequest(nil, http.MethodPut, *reqUrl, reqPayload)
	req.Header.Add("Content-Type", updateDiskLink.Type)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error update disk: %s", err)
	}

	// Decode response
	task := NewTask(d.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// Return the task
	return *task, nil
}

// Remove an independent disk
// 1 Verify the independent disk is not connected to any VM
// 2 Delete the independent disk. Make a DELETE request to the URL in the rel="remove" link in the Disk
// 3 Return task of independent disk deletion
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 106 - 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Delete disk, HREF: %s \n", d.Disk.HREF)

	var err error

	// Verify the independent disk is not connected to any VM
	vmRef, err := d.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var deleteDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Rel == types.RelRemove {
			util.Logger.Printf("[TRACE] Delete disk - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)
			deleteDiskLink = diskLink
			break
		}
	}

	if deleteDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for delete disk in disk Link")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(deleteDiskLink.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parse uri: %s", err)
	}

	// Make request
	req := d.client.NewRequest(nil, http.MethodDelete, *reqUrl, nil)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error delete disk: %s", err)
	}

	// Decode response
	task := NewTask(d.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding delete disk params response: %s", err)
	}

	// Return the task
	return *task, nil
}

// Refresh the disk information by disk href
func (d *Disk) Refresh() error {
	util.Logger.Printf("[TRACE] Disk refresh, HREF: %s\n", d.Disk.HREF)

	disk, err := FindDiskByHREF(d.client, d.Disk.HREF)
	if err != nil {
		return err
	}

	d.Disk = disk.Disk

	return nil
}

// Get a VM that is attached the disk
// An independent disk can be attached to at most one virtual machine.
// If the disk isn't attached to any VM, return empty VM reference and no error.
// Otherwise return the first VM reference and no error.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (d *Disk) AttachedVM() (*types.Reference, error) {
	util.Logger.Printf("[TRACE] Disk attached VM, HREF: %s\n", d.Disk.HREF)

	var attachedVMLink *types.Link
	var err error

	// Find the proper link for request
	for _, diskLink := range d.Disk.Link {
		if diskLink.Type == types.MimeVMs {
			util.Logger.Printf("[TRACE] Disk attached VM - found the proper link for request, HREF: %s, name: %s, type: %s,id: %s, rel: %s \n",
				diskLink.HREF,
				diskLink.Name,
				diskLink.Type,
				diskLink.ID,
				diskLink.Rel)

			attachedVMLink = diskLink
			break
		}
	}

	if attachedVMLink == nil {
		return nil, fmt.Errorf("could not find request URL for attached vm in disk Link")
	}

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(attachedVMLink.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parse uri: %s", err)
	}

	// Send request
	req := d.client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	req.Header.Add("Content-Type", attachedVMLink.Type)
	resp, err := checkResp(d.client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error attached vms: %s", err)
	}

	// Decode request
	var vms = new(types.Vms)
	if err = decodeBody(resp, vms); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// If disk is not attached to any VM
	if vms.VmReference == nil {
		return nil, nil
	}

	// An independent disk can be attached to at most one virtual machine so return the first result of VM reference
	return vms.VmReference, nil
}

// Find an independent disk by disk href in VDC
func (vdc *Vdc) FindDiskByHREF(href string) (*Disk, error) {
	util.Logger.Printf("[TRACE] VDC find disk By HREF: %s\n", href)

	return FindDiskByHREF(vdc.client, href)
}

// Find an independent disk by VDC client and disk href
func FindDiskByHREF(client *Client, href string) (*Disk, error) {
	util.Logger.Printf("[TRACE] Find disk By HREF: %s\n", href)

	// Parse request URI
	reqUrl, err := url.ParseRequestURI(href)
	if err != nil {
		return nil, fmt.Errorf("error parse URI: %s", err)
	}

	// Send request
	req := client.NewRequest(nil, http.MethodGet, *reqUrl, nil)
	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error find disk: %s", err)
	}

	// Decode response
	disk := NewDisk(client)
	if err = decodeBody(resp, disk.Disk); err != nil {
		return nil, fmt.Errorf("error decoding find disk response: %s", err)
	}

	// Return the disk
	return disk, nil
}
