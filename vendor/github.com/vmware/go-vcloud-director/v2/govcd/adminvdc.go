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

// vdcVersionedFuncs is a generic representation of VDC CRUD operations across multiple versions
type vdcVersionedFuncs struct {
	SupportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
	UpdateVdc        func(adminVdc *AdminVdc) (*AdminVdc, error)
	UpdateVdcAsync   func(adminVdc *AdminVdc) (Task, error)
}

// VDC function mapping for API version 31.0 (from vCD 9.5)
var vdcVersionedFuncsV95 = vdcVersionedFuncs{
	SupportedVersion: "31.0",
	CreateVdc:        createVdc,
	CreateVdcAsync:   createVdcAsync,
	UpdateVdc:        updateVdc,
	UpdateVdcAsync:   updateVdcAsync,
}

// VDC function mapping for API version 32.0 (from vCD 9.7)
var vdcVersionedFuncsV97 = vdcVersionedFuncs{
	SupportedVersion: "32.0",
	CreateVdc:        createVdcV97,
	CreateVdcAsync:   createVdcAsyncV97,
	UpdateVdc:        updateVdcV97,
	UpdateVdcAsync:   updateVdcAsyncV97,
}

// TODO: add a wrapper function to use newest available method when version is higher than currently handled
// VDC function mapping by vDC version
var vdcVersionedFuncsByVcdVersion = map[string]vdcVersionedFuncs{
	"vdc9.5":  vdcVersionedFuncsV95,
	"vdc9.7":  vdcVersionedFuncsV97,
	"vdc10.0": vdcVersionedFuncsV97,
	"vdc10.1": vdcVersionedFuncsV97,
}

// GetAdminVdcByName function uses a valid VDC name and returns a admin VDC object.
// If no VDC is found, then it returns an empty VDC and no error.
// Otherwise it returns an empty VDC and an error.
// Deprecated: Use adminOrg.GetAdminVDCByName
func (adminOrg *AdminOrg) GetAdminVdcByName(vdcname string) (AdminVdc, error) {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdcs.Name == vdcname {
			adminVdc := NewAdminVdc(adminOrg.client)
			_, err := adminOrg.client.ExecuteRequest(vdcs.HREF, http.MethodGet,
				"", "error getting vdc: %s", nil, adminVdc.AdminVdc)
			return *adminVdc, err
		}
	}
	return AdminVdc{}, nil
}

// GetAdminVDCByHref retrieves a VDC using a direct call with the HREF
func (adminOrg *AdminOrg) GetAdminVDCByHref(vdcHref string) (*AdminVdc, error) {

	adminVdc := NewAdminVdc(adminOrg.client)

	// We are executing below request with a specific API version in the header, because we want to retrieve the most
	// available fields in AdminVdc which vCD provides, but also which our code understands. As we can't blindly use
	// the latest version, we're limiting the highest used version to the one we support with
	// the GetSpecificApiVersionOnCondition(...) function. Specifically, the API version 32 returns
	// two additional fields: IncludeMemoryOverhead and IsElastic for Flex allocation
	_, err := adminOrg.client.ExecuteRequestWithApiVersion(vdcHref, http.MethodGet,
		"", "error getting vdc: %s", nil, adminVdc.AdminVdc, adminVdc.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))

	if err != nil {
		return nil, err
	}
	return adminVdc, nil
}

// GetAdminVDCByName finds an Admin VDC by Name
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCByName(vdcName string, refresh bool) (*AdminVdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdc.Name == vdcName {
			return adminOrg.GetAdminVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetAdminVDCById finds an Admin VDC by ID
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCById(vdcId string, refresh bool) (*AdminVdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if equalIds(vdcId, vdc.ID, vdc.HREF) {
			return adminOrg.GetAdminVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetAdminVDCByNameOrId finds an Admin VDC by Name Or ID
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCByNameOrId(identifier string, refresh bool) (*AdminVdc, error) {
	getByName := func(name string, refresh bool) (interface{}, error) {
		return adminOrg.GetAdminVDCByName(name, refresh)
	}
	getById := func(id string, refresh bool) (interface{}, error) { return adminOrg.GetAdminVDCById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*AdminVdc), err
}

// CreateVdc creates a VDC with the given params under the given organization.
// Returns an AdminVdc.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-VdcConfiguration.html
// Deprecated in favor of adminOrg.CreateOrgVdcAsync
func (adminOrg *AdminOrg) CreateVdc(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	err := validateVdcConfiguration(vdcConfiguration)
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

	_, err = adminOrg.client.ExecuteRequest(vdcCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.createVdcParams+xml", "error creating VDC: %s", vdcConfiguration, adminVdc.AdminVdc)
	if err != nil {
		return Task{}, err
	}

	// Return the task
	task := NewTask(adminOrg.client)
	task.Task = adminVdc.AdminVdc.Tasks.Task[0]
	return *task, nil
}

// Creates the VDC and waits for the asynchronous task to complete.
// Deprecated in favor of adminOrg.CreateOrgVdc
func (adminOrg *AdminOrg) CreateVdcWait(vdcDefinition *types.VdcConfiguration) error {
	task, err := adminOrg.CreateVdc(vdcDefinition)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("couldn't finish creating VDC %s", err)
	}
	return nil
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

	// We are executing below request with a specific API version in the header, because we want to retrieve the most
	// available fields in AdminVdc which vCD provides, but also which our code understands. As we can't blindly use
	// the latest version, we're limiting the highest used version to the one we support with
	// the GetSpecificApiVersionOnCondition(...) function. Specifically, the API version 32 returns
	// two additional fields: IncludeMemoryOverhead and IsElastic for Flex allocation
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
	vdcFunctions, ok := vdcVersionedFuncsByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return Task{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if vdcFunctions.UpdateVdcAsync == nil {
		return Task{}, fmt.Errorf("function UpdateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}
	util.Logger.Printf("[DEBUG] UpdateAsync call function for version %s", vdcFunctions.SupportedVersion)

	return vdcFunctions.UpdateVdcAsync(adminVdc)

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

	vdcFunctions, ok := vdcVersionedFuncsByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return AdminVdc{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if vdcFunctions.UpdateVdc == nil {
		return AdminVdc{}, fmt.Errorf("function UpdateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] Update call function for version %s", vdcFunctions.SupportedVersion)

	updatedAdminVdc, err := vdcFunctions.UpdateVdc(adminVdc)
	if err != nil {
		return AdminVdc{}, err
	}
	return *updatedAdminVdc, err
}

// CreateOrgVdc creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete.
// Returns an AdminVdc pointer and an error.
func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return nil, err
	}
	vdcFunctions, ok := vdcVersionedFuncsByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return nil, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if vdcFunctions.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdc call function for version %s", vdcFunctions.SupportedVersion)
	return vdcFunctions.CreateVdc(adminOrg, vdcConfiguration)
}

// CreateOrgVdcAsync creates a VDC with the given params under the given organization.
// Returns a Task and an error.
func (adminOrg *AdminOrg) CreateOrgVdcAsync(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	vdcFunctions, ok := vdcVersionedFuncsByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return Task{}, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if vdcFunctions.CreateVdcAsync == nil {
		return Task{}, fmt.Errorf("function CreateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdcAsync call function for version %s", vdcFunctions.SupportedVersion)

	return vdcFunctions.CreateVdcAsync(adminOrg, vdcConfiguration)
}

// createVdc creates a VDC with the given params under the given organization.
// Returns a Vdc pointer and an error
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

// updateVdcAsync updates a VDC with the given params. Returns a Task and error
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
// Needs vCD 9.7+ to work. Returns a Task and an error.
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
// Needs vCD 9.7+ to work. Returns an AdminVdc pointer and an error.
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
// Returns a Task and an error
func createVdcAsync(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	util.Logger.Printf("[TRACE] createVdcAsync called %#v", *vdcConfiguration)
	return adminOrg.CreateVdc(vdcConfiguration)
}

// createVdcV97 creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns a Vdc pointer and error.
func createVdcV97(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	util.Logger.Printf("[TRACE] createVdcV97 called %#v", *vdcConfiguration)
	task, err := createVdcAsyncV97(adminOrg, vdcConfiguration)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("couldn't finish creating VDC %s", err)
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

// createVdcAsyncV97 creates a VDC with the given params under the given organization. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns a Task and an error
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
		"application/vnd.vmware.admin.createVdcParams+xml", "error creating VDC: %s",
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
