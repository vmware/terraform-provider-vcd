/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/http"
	"net/url"
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

	_, err := adminVdc.client.ExecuteRequestWithApiVersion(adminVdc.AdminVdc.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, unmarshalledAdminVdc, adminVdc.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))
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
	apiVersion, err := adminVdc.client.maxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	realFunction, ok := vdcProducerByVersion["vdc"+vcdVersionToApiVersion[apiVersion]]
	if !ok {
		return Task{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if realFunction.UpdateVdcAsync == nil {
		return Task{}, fmt.Errorf("function UpdateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}
	util.Logger.Printf("[DEBUG] UpdateAsync call function for version %s", realFunction.SupportedVersion)

	return realFunction.UpdateVdcAsync(adminVdc)

}

// Update function updates an Admin VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an empty AdminVdc struct and error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) Update() (AdminVdc, error) {
	apiVersion, err := adminVdc.client.maxSupportedVersion()
	if err != nil {
		return AdminVdc{}, err
	}

	realFunction, ok := vdcProducerByVersion["vdc"+vcdVersionToApiVersion[apiVersion]]
	if !ok {
		return AdminVdc{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if realFunction.UpdateVdc == nil {
		return AdminVdc{}, fmt.Errorf("function UpdateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] Update call function for version %s", realFunction.SupportedVersion)

	updatedAdminVdc, err := realFunction.UpdateVdc(adminVdc)
	if err != nil {
		return AdminVdc{}, err
	}
	return *updatedAdminVdc, err
}

type vdcProducer struct {
	SupportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
	UpdateVdc        func(adminVdc *AdminVdc) (*AdminVdc, error)
	UpdateVdcAsync   func(adminVdc *AdminVdc) (Task, error)
}

var vdcCrudV90 = vdcProducer{
	SupportedVersion: "29.0",
	CreateVdc:        createVdc,
	CreateVdcAsync:   createVdcAsync,
	UpdateVdc:        updateVdc,
	UpdateVdcAsync:   updateVdcAsync,
}

var vdcCrudV97 = vdcProducer{
	SupportedVersion: "32.0",
	CreateVdc:        createVdcV97,
	CreateVdcAsync:   createVdcAsyncV97,
	UpdateVdc:        updateVdcV97,
	UpdateVdcAsync:   updateVdcAsyncV97,
}

var vdcProducerByVersion = map[string]vdcProducer{
	"vdc9.0":  vdcCrudV90,
	"vdc9.1":  vdcCrudV90,
	"vdc9.5":  vdcCrudV90,
	"vdc9.7":  vdcCrudV97,
	"vdc10.0": vdcCrudV97,
}

// CreateOrgVdc creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete.
// Returns an AdminVdc.
func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return nil, err
	}
	realFunction, ok := vdcProducerByVersion["vdc"+vcdVersionToApiVersion[apiVersion]]
	if !ok {
		return nil, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if realFunction.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdc call function for version %s", realFunction.SupportedVersion)
	return realFunction.CreateVdc(adminOrg, vdcConfiguration)
}

// CreateOrgVdcAsync creates a VDC with the given params under the given organization.
// Returns an Task.
func (adminOrg *AdminOrg) CreateOrgVdcAsync(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	realFunction, ok := vdcProducerByVersion["vdc"+vcdVersionToApiVersion[apiVersion]]
	if !ok {
		return Task{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if realFunction.CreateVdcAsync == nil {
		return Task{}, fmt.Errorf("function CreateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdcAsync call function for version %s", realFunction.SupportedVersion)

	return realFunction.CreateVdcAsync(adminOrg, vdcConfiguration)
}

// createVdc creates a VDC with the given params under the given organization.
// Returns an Vdc.
func createVdc(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	util.Logger.Printf("[TRACE] createVdc called %#v", *vdcConfiguration)
	err := adminOrg.CreateVdcWait(vdcConfiguration)
	if err != nil {
		return nil, err
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

// updateVdcAsync updates a VDC with the given params. Returns an Task.
func updateVdcAsync(adminVdc *AdminVdc) (Task, error) {
	util.Logger.Printf("[TRACE] updateVdcAsync called %#v", *adminVdc)
	adminVdc.AdminVdc.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return adminVdc.client.ExecuteTaskRequest(adminVdc.AdminVdc.HREF, http.MethodPut,
		types.MimeAdminVDC, "error updating VDC: %s", adminVdc.AdminVdc)
}

// updateVdc updates a VDC with the given params. Returns an AdminVdc.
func updateVdc(adminVdc *AdminVdc) (*AdminVdc, error) {
	util.Logger.Printf("[TRACE] updateVdc called %#v", *adminVdc)
	task, err := updateVdcAsync(adminVdc)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = adminVdc.Refresh()
	if err != nil {
		return nil, err
	}

	return adminVdc, nil
}

// updateVdcAsyncV97 updates a VDC with the given params. Supports Flex type allocation.
// Needs vCD 9.7 to work. Returns an Task.
func updateVdcAsyncV97(adminVdc *AdminVdc) (Task, error) {
	util.Logger.Printf("[TRACE] updateVdcAsyncV97 called %#v", *adminVdc)
	adminVdc.AdminVdc.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return adminVdc.client.ExecuteTaskRequestWithApiVersion(adminVdc.AdminVdc.HREF, http.MethodPut,
		types.MimeAdminVDC, "error updating VDC: %s", adminVdc.AdminVdc,
		adminVdc.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))
}

// updateVdcV97 updates a VDC with the given params
// and waits for the asynchronous task to complete. Supports Flex type allocation.
// Needs vCD 9.7 to work. Returns an AdminVdc.
func updateVdcV97(adminVdc *AdminVdc) (*AdminVdc, error) {
	util.Logger.Printf("[TRACE] updateVdcV97 called %#v", *adminVdc)
	task, err := updateVdcAsyncV97(adminVdc)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}
	err = adminVdc.Refresh()
	if err != nil {
		return nil, err
	}
	return adminVdc, nil
}

// createVdcAsync creates a VDC with the given params under the given organization.
// Returns an Task.
func createVdcAsync(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	util.Logger.Printf("[TRACE] createVdcAsync called %#v", *vdcConfiguration)
	return adminOrg.CreateVdc(vdcConfiguration)
}

// createVdcAsyncV97 creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete. Supports Flex type allocation.
// Needs vCD 9.7 to work. Returns an Vdc.
func createVdcV97(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	util.Logger.Printf("[TRACE] createVdcV97 called %#v", *vdcConfiguration)
	task, err := createVdcAsyncV97(adminOrg, vdcConfiguration)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("couldn't finish creating vdc %s", err)
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

// createVdcAsyncV97 creates a VDC with the given params under the given organization. Supports Flex type allocation.
// Needs vCD 9.7 to work. Returns an Task.
func createVdcAsyncV97(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	util.Logger.Printf("[TRACE] createVdcAsyncV97 called %#v", *vdcConfiguration)
	err := validateVdcConfigurationV97(*vdcConfiguration)
	if err != nil {
		return Task{}, err
	}

	vdcConfiguration.Xmlns = types.XMLNamespaceVCloud

	vdcCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing admin org url: %s", err)
	}
	vdcCreateHREF.Path += "/vdcsparams"

	adminVdc := NewAdminVdc(adminOrg.client)

	_, err = adminOrg.client.ExecuteRequestWithApiVersion(vdcCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.createVdcParams+xml", "error retrieving vdc: %s",
		vdcConfiguration, adminVdc.AdminVdc,
		adminOrg.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))
	if err != nil {
		return Task{}, err
	}

	// Return the task
	task := NewTask(adminOrg.client)
	task.Task = adminVdc.AdminVdc.Tasks.Task[0]
	return *task, nil
}

// validateVdcConfigurationV97 uses validateVdcConfiguration and additionally checks Flex dependent values
func validateVdcConfigurationV97(vdcDefinition types.VdcConfiguration) error {
	err := validateVdcConfiguration(&vdcDefinition)
	if err != nil {
		return err
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IsElastic == nil {
		return errors.New("VdcConfiguration missing required field: IsElastic")
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IncludeMemoryOverhead == nil {
		return errors.New("VdcConfiguration missing required field: IncludeMemoryOverhead")
	}
	return nil
}
