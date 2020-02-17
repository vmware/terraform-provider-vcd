/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type AdminVdc struct {
	AdminVdc *types.AdminVdc
	client   *Client
}

func NewAdminVdc(cli *Client) *AdminVdc {
	return &AdminVdc{
		AdminVdc: new(types.AdminVdc),
		client:   cli,
	}
}

// Given an adminVdc with a valid HREF, the function refresh the adminVdc
// and updates the adminVdc data. Returns an error on failure
// Users should use refresh whenever they suspect
// a stale VDC due to the creation/update/deletion of a resource
// within the the VDC itself.
func (adminVdc *AdminVdc) Refresh() error {
	if *adminVdc == (AdminVdc{}) || adminVdc.AdminVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty or HREF is empty")
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledAdminVdc := &types.AdminVdc{}

	_, err := adminVdc.client.ExecuteRequest(adminVdc.AdminVdc.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, unmarshalledAdminVdc)
	if err != nil {
		return err
	}
	adminVdc.AdminVdc = unmarshalledAdminVdc

	return nil
}

// UpdateAsync updates VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) UpdateAsync() (Task, error) {

	adminVdc.AdminVdc.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return adminVdc.client.ExecuteTaskRequest(adminVdc.AdminVdc.HREF, http.MethodPut,
		types.MimeAdminVDC, "error updating VDC: %s", adminVdc.AdminVdc)
}

// Update function updates an Admin VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an empty AdminVdc struct and error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) Update() (AdminVdc, error) {
	task, err := adminVdc.UpdateAsync()
	if err != nil {
		return AdminVdc{}, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return AdminVdc{}, err
	}

	err = adminVdc.Refresh()
	if err != nil {
		return AdminVdc{}, err
	}

	return *adminVdc, nil
}
