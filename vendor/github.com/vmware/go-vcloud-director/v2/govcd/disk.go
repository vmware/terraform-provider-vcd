/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Independent disk
type Disk struct {
	Disk   *types.Disk
	client *Client
}

// Independent disk query record
type DiskRecord struct {
	Disk   *types.DiskRecordType
	client *Client
}

// Init independent disk struct
func NewDisk(cli *Client) *Disk {
	return &Disk{
		Disk:   new(types.Disk),
		client: cli,
	}
}

// Create instance with reference to types.DiskRecordType
func NewDiskRecord(cli *Client) *DiskRecord {
	return &DiskRecord{
		Disk:   new(types.DiskRecordType),
		client: cli,
	}
}

// While theoretically we can use smaller amounts, there is an issue when updating
// disks with size < 1MB
const MinimumDiskSize int64 = 1048576 // = 1Mb

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

	// Prepare the request payload
	diskCreateParams.Xmlns = types.XMLNamespaceVCloud

	disk := NewDisk(vdc.client)

	_, err = vdc.client.ExecuteRequest(createDiskLink.HREF, http.MethodPost,
		createDiskLink.Type, "error create disk: %s", diskCreateParams, disk.Disk)
	if err != nil {
		return Task{}, err
	}
	// Obtain disk task
	if disk.Disk.Tasks.Task == nil || len(disk.Disk.Tasks.Task) == 0 {
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
func (disk *Disk) Update(newDiskInfo *types.Disk) (Task, error) {
	util.Logger.Printf("[TRACE] Update disk, name: %s, size: %d, HREF: %s \n",
		newDiskInfo.Name,
		newDiskInfo.Size,
		disk.Disk.HREF,
	)

	var err error

	if newDiskInfo.Name == "" {
		return Task{}, fmt.Errorf("disk name is required")
	}

	if newDiskInfo.Size < MinimumDiskSize {
		return Task{}, fmt.Errorf("disk size should be greater than or equal to 1Mb")
	}

	// Verify the independent disk is not connected to any VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var updateDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
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

	// Prepare the request payload
	xmlPayload := &types.Disk{
		Xmlns:          types.XMLNamespaceVCloud,
		Description:    newDiskInfo.Description,
		Size:           newDiskInfo.Size,
		Name:           newDiskInfo.Name,
		StorageProfile: newDiskInfo.StorageProfile,
		Owner:          newDiskInfo.Owner,
	}

	// Return the task
	return disk.client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut,
		updateDiskLink.Type, "error updating disk: %s", xmlPayload)
}

// Remove an independent disk
// 1 Verify the independent disk is not connected to any VM
// 2 Delete the independent disk. Make a DELETE request to the URL in the rel="remove" link in the Disk
// 3 Return task of independent disk deletion
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 106 - 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (disk *Disk) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Delete disk, HREF: %s \n", disk.Disk.HREF)

	var err error

	// Verify the independent disk is not connected to any VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return Task{}, fmt.Errorf("error find attached VM: %s", err)
	}
	if vmRef != nil {
		return Task{}, errors.New("error disk is attached")
	}

	var deleteDiskLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
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

	// Return the task
	return disk.client.ExecuteTaskRequest(deleteDiskLink.HREF, http.MethodDelete,
		"", "error delete disk: %s", nil)
}

// Refresh the disk information by disk href
func (disk *Disk) Refresh() error {
	util.Logger.Printf("[TRACE] Disk refresh, HREF: %s\n", disk.Disk.HREF)

	fetchedDisk, err := FindDiskByHREF(disk.client, disk.Disk.HREF)
	if err != nil {
		return err
	}

	disk.Disk = fetchedDisk.Disk

	return nil
}

// Get a VM that is attached the disk
// An independent disk can be attached to at most one virtual machine.
// If the disk isn't attached to any VM, return empty VM reference and no error.
// Otherwise return the first VM reference and no error.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 107,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (disk *Disk) AttachedVM() (*types.Reference, error) {
	util.Logger.Printf("[TRACE] Disk attached VM, HREF: %s\n", disk.Disk.HREF)

	var attachedVMLink *types.Link

	// Find the proper link for request
	for _, diskLink := range disk.Disk.Link {
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

	// Decode request
	var vms = new(types.Vms)

	_, err := disk.client.ExecuteRequest(attachedVMLink.HREF, http.MethodGet,
		attachedVMLink.Type, "error getting attached vms: %s", nil, vms)
	if err != nil {
		return nil, err
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

	disk := NewDisk(client)

	_, err := client.ExecuteRequest(href, http.MethodGet,
		"", "error finding disk: %s", nil, disk.Disk)

	// Return the disk
	return disk, err

}

// Find independent disk using disk name. Returns VMRecord query return type
func (vdc *Vdc) QueryDisk(diskName string) (DiskRecord, error) {

	if diskName == "" {
		return DiskRecord{}, fmt.Errorf("disk name can not be empty")
	}

	typeMedia := "disk"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminDisk"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia, "filter": "name==" + url.QueryEscape(diskName)})
	if err != nil {
		return DiskRecord{}, fmt.Errorf("error querying disk %#v", err)
	}

	diskResults := results.Results.DiskRecord
	if vdc.client.IsSysAdmin {
		diskResults = results.Results.AdminDiskRecord
	}

	newDisk := NewDiskRecord(vdc.client)

	if len(diskResults) == 1 {
		newDisk.Disk = diskResults[0]
	} else {
		return DiskRecord{}, fmt.Errorf("found results %d", len(diskResults))
	}

	return *newDisk, nil
}
