/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// AdminOrg gives an admin representation of an org.
// Administrators can delete and update orgs with an admin org object.
// AdminOrg includes all members of the Org element, and adds several
// elements that can be viewed and modified only by system administrators.
// Definition: https://code.vmware.com/apis/220/vcloud#/doc/doc/types/AdminOrgType.html
type AdminOrg struct {
	AdminOrg *types.AdminOrg
	client   *Client
}

func NewAdminOrg(cli *Client) *AdminOrg {
	return &AdminOrg{
		AdminOrg: new(types.AdminOrg),
		client:   cli,
	}
}

// CreateCatalog creates a catalog with given name and description under the
// the given organization. Returns an AdminCatalog that contains a creation
// task.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-CreateCatalog.html
func (adminOrg *AdminOrg) CreateCatalog(name, description string) (AdminCatalog, error) {
	return CreateCatalog(adminOrg.client, adminOrg.AdminOrg.Link, name, description)
}

//   Deletes the org, returning an error if the vCD call fails.
//   API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Organization.html
func (adminOrg *AdminOrg) Delete(force bool, recursive bool) error {
	if force && recursive {
		//undeploys vapps
		err := adminOrg.undeployAllVApps()
		if err != nil {
			return fmt.Errorf("error could not undeploy: %s", err)
		}
		//removes vapps
		err = adminOrg.removeAllVApps()
		if err != nil {
			return fmt.Errorf("error could not remove vapp: %s", err)
		}
		//removes catalogs
		err = adminOrg.removeCatalogs()
		if err != nil {
			return fmt.Errorf("error could not remove all catalogs: %s", err)
		}
		//removes networks
		err = adminOrg.removeAllOrgNetworks()
		if err != nil {
			return fmt.Errorf("error could not remove all networks: %s", err)
		}
		//removes org vdcs
		err = adminOrg.removeAllOrgVDCs()
		if err != nil {
			return fmt.Errorf("error could not remove all vdcs: %s", err)
		}
	}
	// Disable org
	err := adminOrg.Disable()
	if err != nil {
		return fmt.Errorf("error disabling Org %s: %s", adminOrg.AdminOrg.Name, err)
	}
	// Get admin HREF
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %s", adminOrg.AdminOrg.HREF, err)
	}
	req := adminOrg.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, http.MethodDelete, *orgHREF, nil)
	resp, err := checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error deleting Org %s: %s", adminOrg.AdminOrg.ID, err)
	}

	task := NewTask(adminOrg.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}
	return task.WaitTaskCompletion()
}

// Disables the org. Returns an error if the call to vCD fails.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-DisableOrg.html
func (adminOrg *AdminOrg) Disable() error {
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %s", adminOrg.AdminOrg.HREF, err)
	}
	orgHREF.Path += "/action/disable"

	return adminOrg.client.ExecuteRequestWithoutResponse(orgHREF.String(), http.MethodPost, "", "error disabling organization: %s", nil)
}

//   Updates the Org definition from current org struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails.
//   API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Organization.html
func (adminOrg *AdminOrg) Update() (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        adminOrg.AdminOrg.Name,
		IsEnabled:   adminOrg.AdminOrg.IsEnabled,
		FullName:    adminOrg.AdminOrg.FullName,
		Description: adminOrg.AdminOrg.Description,
		OrgSettings: adminOrg.AdminOrg.OrgSettings,
	}

	// Same workaround used in Org creation, where OrgGeneralSettings properties
	// are not set unless UseServerBootSequence is also set
	if vcomp.OrgSettings.OrgGeneralSettings != nil {
		vcomp.OrgSettings.OrgGeneralSettings.UseServerBootSequence = true
	}

	// Return the task
	return adminOrg.client.ExecuteTaskRequest(adminOrg.AdminOrg.HREF, http.MethodPut,
		"application/vnd.vmware.admin.organization+xml", "error updating Org: %s", vcomp)
}

// Undeploys every vapp within an organization
func (adminOrg *AdminOrg) undeployAllVApps() error {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(vdcs.HREF)
		if err != nil {
			return err
		}
		vdc, err := adminOrg.getVdcByAdminHREF(adminVdcHREF)
		if err != nil {
			return fmt.Errorf("error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.undeployAllVdcVApps()
		if err != nil {
			return fmt.Errorf("error deleting vapp: %s", err)
		}
	}
	return nil
}

// Deletes every vapp within an organization
func (adminOrg *AdminOrg) removeAllVApps() error {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(vdcs.HREF)
		if err != nil {
			return err
		}
		vdc, err := adminOrg.getVdcByAdminHREF(adminVdcHREF)
		if err != nil {
			return fmt.Errorf("error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.removeAllVdcVApps()
		if err != nil {
			return fmt.Errorf("error deleting vapp: %s", err)
		}
	}
	return nil
}

// Given an adminorg with a valid HREF, the function refetches the adminorg
// and updates the user's adminorg data. Otherwise if the function fails,
// it returns an error.  Users should use refresh whenever they have
// a stale org due to the creation/update/deletion of a resource
// within the org or the org itself.
func (adminOrg *AdminOrg) Refresh() error {
	if *adminOrg == (AdminOrg{}) {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledAdminOrg := &types.AdminOrg{}

	_, err := adminOrg.client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet,
		"", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
	if err != nil {
		return err
	}
	adminOrg.AdminOrg = unmarshalledAdminOrg

	return nil
}

// Gets a vdc within org associated with an admin vdc url
func (adminOrg *AdminOrg) getVdcByAdminHREF(adminVdcUrl *url.URL) (*Vdc, error) {
	// get non admin vdc path
	vdcURL := adminOrg.client.VCDHREF
	vdcURL.Path += strings.Split(adminVdcUrl.Path, "/api/admin")[1] //gets id

	vdc := NewVdc(adminOrg.client)

	_, err := adminOrg.client.ExecuteRequest(vdcURL.String(), http.MethodGet,
		"", "error retrieving vdc: %s", nil, vdc.Vdc)

	return vdc, err
}

// Removes all vdcs in a org
func (adminOrg *AdminOrg) removeAllOrgVDCs() error {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {

		adminVdcUrl := adminOrg.client.VCDHREF
		splitVdcId := strings.Split(vdcs.HREF, "/api/vdc/")
		if len(splitVdcId) == 1 {
			adminVdcUrl.Path += "/admin/vdc/" + strings.Split(vdcs.HREF, "/api/admin/vdc/")[1] + "/action/disable"
		} else {
			adminVdcUrl.Path += "/admin/vdc/" + splitVdcId[1] + "/action/disable"
		}

		req := adminOrg.client.NewRequest(map[string]string{}, http.MethodPost, adminVdcUrl, nil)
		_, err := checkResp(adminOrg.client.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error disabling vdc: %s", err)
		}
		// Get admin vdc HREF for normal deletion
		adminVdcUrl.Path = strings.Split(adminVdcUrl.Path, "/action/disable")[0]
		req = adminOrg.client.NewRequest(map[string]string{
			"recursive": "true",
			"force":     "true",
		}, http.MethodDelete, adminVdcUrl, nil)
		resp, err := checkResp(adminOrg.client.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting vdc: %s", err)
		}
		task := NewTask(adminOrg.client)
		if err = decodeBody(resp, task.Task); err != nil {
			return fmt.Errorf("error decoding task response: %s", err)
		}
		if task.Task.Status == "error" {
			return fmt.Errorf("vdc not properly destroyed")
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("couldn't finish removing vdc %s", err)
		}

	}

	return nil
}

// Removes All networks in the org
func (adminOrg *AdminOrg) removeAllOrgNetworks() error {
	for _, networks := range adminOrg.AdminOrg.Networks.Networks {
		// Get Network HREF
		networkHREF := adminOrg.client.VCDHREF
		networkHREF.Path += "/admin/network/" + strings.Split(networks.HREF, "/api/admin/network/")[1] //gets id

		task, err := adminOrg.client.ExecuteTaskRequest(networkHREF.String(), http.MethodDelete,
			"", "error deleting network: %s", nil)
		if err != nil {
			return err
		}

		if task.Task.Status == "error" {
			return fmt.Errorf("network not properly destroyed")
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("couldn't finish removing network %s", err)
		}
	}
	return nil
}

// removeCatalogs force removal of all organization catalogs
func (adminOrg *AdminOrg) removeCatalogs() error {
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		isCatalogFromSameOrg, err := isCatalogFromSameOrg(adminOrg, catalog.Name)
		if err != nil {
			return fmt.Errorf("error deleting catalog: %s", err)
		}
		if isCatalogFromSameOrg {
			// Get Catalog HREF
			catalogHREF := adminOrg.client.VCDHREF
			catalogHREF.Path += "/admin/catalog/" + strings.Split(catalog.HREF, "/api/admin/catalog/")[1] //gets id
			req := adminOrg.client.NewRequest(map[string]string{
				"force":     "true",
				"recursive": "true",
			}, http.MethodDelete, catalogHREF, nil)
			_, err := checkResp(adminOrg.client.Http.Do(req))
			if err != nil {
				return fmt.Errorf("error deleting catalog: %s, %s", err, catalogHREF.Path)
			}
		}
	}
	return nil

}

// isCatalogFromSameOrg checks if catalog is in same Org. Shared catalogs from other Org are showed as normal one
// in some API responses.
func isCatalogFromSameOrg(adminOrg *AdminOrg, catalogName string) (bool, error) {
	foundCatalogs, err := adminOrg.FindAdminCatalogRecords(catalogName)
	if err != nil {
		return false, err
	}

	if len(foundCatalogs) == 1 {
		return true, nil
	}
	return false, nil
}

// FindAdminCatalogRecords uses catalog name to return AdminCatalogRecord information.
func (adminOrg *AdminOrg) FindAdminCatalogRecords(name string) ([]*types.CatalogRecord, error) {
	util.Logger.Printf("[DEBUG] FindAdminCatalogRecords with name: %s and org name: %s", name, adminOrg.AdminOrg.Name)
	results, err := adminOrg.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "adminCatalog",
		"filter":        fmt.Sprintf("name==%s;orgName==%s", url.QueryEscape(name), url.QueryEscape(adminOrg.AdminOrg.Name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] FindAdminCatalogRecords returned with : %#v and error: %s", results.Results.AdminCatalogRecord, err)
	return results.Results.AdminCatalogRecord, nil
}

// Given a valid catalog name, FindAdminCatalog returns an AdminCatalog object.
// If no catalog is found, then returns an empty AdminCatalog and no error.
// Otherwise it returns an error. Function allows user to use an AdminOrg
// to also fetch a Catalog. If user does not have proper credentials to
// perform administrator tasks then function returns an error.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/GET-Catalog-AdminView.html
// Deprecated: Use adminOrg.GetAdminCatalog instead
func (adminOrg *AdminOrg) FindAdminCatalog(catalogName string) (AdminCatalog, error) {
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if catalog.Name == catalogName {
			adminCatalog := NewAdminCatalog(adminOrg.client)
			_, err := adminOrg.client.ExecuteRequest(catalog.HREF, http.MethodGet,
				"", "error retrieving catalog: %s", nil, adminCatalog.AdminCatalog)
			// The request was successful
			return *adminCatalog, err
		}
	}
	return AdminCatalog{}, nil
}

// Given a valid catalog name, FindCatalog returns a Catalog object.
// If no catalog is found, then returns an empty catalog and no error.
// Otherwise it returns an error. Function allows user to use an AdminOrg
// to also fetch a Catalog.
// Deprecated: Use adminOrg.GetCatalogByName instead
func (adminOrg *AdminOrg) FindCatalog(catalogName string) (Catalog, error) {
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if catalog.Name == catalogName {
			catalogURL := adminOrg.client.VCDHREF
			catalogURL.Path += "/catalog/" + strings.Split(catalog.HREF, "/api/admin/catalog/")[1] //gets id

			cat := NewCatalog(adminOrg.client)

			_, err := adminOrg.client.ExecuteRequest(catalogURL.String(), http.MethodGet,
				"", "error retrieving catalog: %s", nil, cat.Catalog)

			// The request was successful
			return *cat, err
		}
	}
	return Catalog{}, nil
}

// GetCatalogByHref  finds a Catalog by HREF
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetCatalogByHref(catalogHref string) (*Catalog, error) {
	splitByAdminHREF := strings.Split(catalogHref, "/api/admin")

	// admin user and normal user will have different urls
	var catalogHREF string
	if len(splitByAdminHREF) == 1 {
		catalogHREF = catalogHref
	} else {
		catalogHREF = splitByAdminHREF[0] + "/api" + splitByAdminHREF[1]
	}

	cat := NewCatalog(adminOrg.client)

	_, err := adminOrg.client.ExecuteRequest(catalogHREF, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.Catalog)

	if err != nil {
		return nil, err
	}
	// The request was successful
	return cat, nil
}

// GetCatalogByName  finds a Catalog by Name
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetCatalogByName(catalogName string, refresh bool) (*Catalog, error) {

	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		if catalog.Name == catalogName {
			return adminOrg.GetCatalogByHref(catalog.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// Extracts an UUID from a string, regardless of surrounding text
// Returns an empty string if no UUID was found
func extractUuid(input string) string {
	reGetID := regexp.MustCompile(`([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`)
	matchListId := reGetID.FindAllStringSubmatch(input, -1)
	if len(matchListId) > 0 && len(matchListId[0]) > 0 {
		return matchListId[0][1]
	}
	return ""
}

// equalIds compares two IDs and return true if they are the same
// The comparison happens by extracting the bare UUID from both the
// wanted ID and the found one.
// When the found ID is empty, it used the HREF for such comparison,
// This function is useful when the reference structure in the parent lookup list
// may lack the ID (such as in Org.Links, AdminOrg.Catalogs) or has an ID
// that is only a UUID without prefixes (such as in CatalogItem list)
//
// wantedId is the input string to compare
// foundId is the ID field in the reference record (can be empty)
// foundHref is the HREF field in the reference record (should never be empty)
func equalIds(wantedId, foundId, foundHref string) bool {

	wantedUuid := extractUuid(wantedId)
	foundUuid := ""

	if wantedUuid == "" {
		return false
	}
	if foundId != "" {
		// In some entities, the ID is a simple UUID without prefix
		foundUuid = extractUuid(foundId)
	} else {
		foundUuid = extractUuid(foundHref)
	}
	return foundUuid == wantedUuid
}

// GetCatalogById finds a Catalog by ID
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetCatalogById(catalogId string, refresh bool) (*Catalog, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		if equalIds(catalogId, catalog.ID, catalog.HREF) {
			return adminOrg.GetCatalogByHref(catalog.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetCatalogByNameOrId finds a Catalog by name or ID
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetCatalogByNameOrId(identifier string, refresh bool) (*Catalog, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetCatalogByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return adminOrg.GetCatalogById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*Catalog), err
}

// GetAdminCatalogByHref  finds an AdminCatalog by HREF
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminCatalogByHref(catalogHref string) (*AdminCatalog, error) {
	adminCatalog := NewAdminCatalog(adminOrg.client)

	_, err := adminOrg.client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, adminCatalog.AdminCatalog)

	if err != nil {
		return nil, err
	}

	// The request was successful
	return adminCatalog, nil
}

// GetCatalogByName finds an AdminCatalog by Name
// On success, returns a pointer to the AdminCatalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminCatalogByName(catalogName string, refresh bool) (*AdminCatalog, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if catalog.Name == catalogName {
			return adminOrg.GetAdminCatalogByHref(catalog.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetCatalogById finds an AdminCatalog by ID
// On success, returns a pointer to the AdminCatalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminCatalogById(catalogId string, refresh bool) (*AdminCatalog, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if equalIds(catalogId, catalog.ID, catalog.HREF) {
			return adminOrg.GetAdminCatalogByHref(catalog.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetAdminCatalogByNameOrId finds an AdminCatalog by name or ID
// On success, returns a pointer to the AdminCatalog structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminCatalogByNameOrId(identifier string, refresh bool) (*AdminCatalog, error) {
	getByName := func(name string, refresh bool) (interface{}, error) {
		return adminOrg.GetAdminCatalogByName(name, refresh)
	}
	getById := func(id string, refresh bool) (interface{}, error) {
		return adminOrg.GetAdminCatalogById(id, refresh)
	}
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*AdminCatalog), err
}

// GetVDCByHref retrieves a VDC using a direct call with the HREF
func (adminOrg *AdminOrg) GetVDCByHref(vdcHref string) (*Vdc, error) {
	splitByAdminHREF := strings.Split(vdcHref, "/api/admin")

	// admin user and normal user will have different urls
	var vdcHREF string
	if len(splitByAdminHREF) == 1 {
		vdcHREF = vdcHref
	} else {
		vdcHREF = splitByAdminHREF[0] + "/api" + splitByAdminHREF[1]
	}

	vdc := NewVdc(adminOrg.client)

	_, err := adminOrg.client.ExecuteRequestWithApiVersion(vdcHREF, http.MethodGet,
		"", "error getting vdc: %s", nil, vdc.Vdc,
		adminOrg.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))

	if err != nil {
		return nil, err
	}

	return vdc, nil
}

// GetVDCByName finds a VDC by Name
// On success, returns a pointer to the Vdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetVDCByName(vdcName string, refresh bool) (*Vdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdc.Name == vdcName {
			return adminOrg.GetVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetVDCById finds a VDC by ID
// On success, returns a pointer to the Vdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetVDCById(vdcId string, refresh bool) (*Vdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if equalIds(vdcId, vdc.ID, vdc.HREF) {
			return adminOrg.GetVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetVDCByNameOrId finds a VDC by name or ID
// On success, returns a pointer to the VDC structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetVDCByNameOrId(identifier string, refresh bool) (*Vdc, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetVDCByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return adminOrg.GetVDCById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*Vdc), err
}

// If user specifies valid vdc name then this returns a vdc object.
// If no vdc is found, then it returns an empty vdc and no error.
// Otherwise it returns an empty vdc and an error. This function
// allows users to use an AdminOrg to fetch a vdc as well.
// Deprecated: Use adminOrg.GetVDCByName instead
func (adminOrg *AdminOrg) GetVdcByName(vdcname string) (Vdc, error) {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdcs.Name == vdcname {
			splitByAdminHREF := strings.Split(vdcs.HREF, "/api/admin")

			// admin user and normal user will have different urls
			var vdcHREF string
			if len(splitByAdminHREF) == 1 {
				vdcHREF = vdcs.HREF
			} else {
				vdcHREF = splitByAdminHREF[0] + "/api" + splitByAdminHREF[1]
			}

			vdc := NewVdc(adminOrg.client)

			_, err := adminOrg.client.ExecuteRequest(vdcHREF, http.MethodGet,
				"", "error getting vdc: %s", nil, vdc.Vdc)

			return *vdc, err
		}
	}
	return Vdc{}, nil
}

// QueryCatalogList returns a list of catalogs for this organization
func (adminOrg *AdminOrg) QueryCatalogList() ([]*types.CatalogRecord, error) {
	util.Logger.Printf("[DEBUG] QueryCatalogList with org name %s", adminOrg.AdminOrg.Name)
	queryType := types.QtCatalog
	if adminOrg.client.IsSysAdmin {
		queryType = types.QtAdminCatalog
	}
	results, err := adminOrg.client.cumulativeQuery(queryType, nil, map[string]string{
		"type":          queryType,
		"filter":        fmt.Sprintf("orgName==%s", url.QueryEscape(adminOrg.AdminOrg.Name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	var catalogs []*types.CatalogRecord

	if adminOrg.client.IsSysAdmin {
		catalogs = results.Results.AdminCatalogRecord
	} else {
		catalogs = results.Results.CatalogRecord
	}
	util.Logger.Printf("[DEBUG] QueryCatalogList returned with : %#v and error: %s", catalogs, err)
	return catalogs, nil
}
