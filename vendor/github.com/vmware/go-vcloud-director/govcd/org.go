/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
	"strconv"
	"strings"
)

// Interface for methods in common for Org and AdminOrg
type OrgOperations interface {
	FindCatalog(catalog string) (Catalog, error)
	GetVdcByName(vdcname string) (Vdc, error)
	Refresh() error
}

type Org struct {
	Org    *types.Org
	client *Client
}

func NewOrg(client *Client) *Org {
	return &Org{
		Org:    new(types.Org),
		client: client,
	}
}

// Given an org with a valid HREF, the function refetches the org
// and updates the user's org data. Otherwise if the function fails,
// it returns an error. Users should use refresh whenever they have
// a stale org due to the creation/update/deletion of a resource
// within the org or the org itself.
func (org *Org) Refresh() error {
	if *org == (Org{}) {
		return fmt.Errorf("cannot refresh, Object is empty")
	}
	orgHREF, _ := url.ParseRequestURI(org.Org.HREF)
	req := org.client.NewRequest(map[string]string{}, "GET", *orgHREF, nil)
	resp, err := checkResp(org.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error performing request: %s", err)
	}
	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledOrg := &types.Org{}
	if err = decodeBody(resp, unmarshalledOrg); err != nil {
		return fmt.Errorf("error decoding org response: %s", err)
	}
	org.Org = unmarshalledOrg
	// The request was successful
	return nil
}

// Given a valid catalog name, FindCatalog returns a Catalog object.
// If no catalog is found, then returns an empty catalog and no error.
// Otherwise it returns an error.
func (org *Org) FindCatalog(catalogName string) (Catalog, error) {

	for _, link := range org.Org.Link {
		if link.Rel == "down" && link.Type == "application/vnd.vmware.vcloud.catalog+xml" && link.Name == catalogName {
			orgHREF, err := url.ParseRequestURI(link.HREF)

			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding org response: %s", err)
			}

			req := org.client.NewRequest(map[string]string{}, "GET", *orgHREF, nil)

			resp, err := checkResp(org.client.Http.Do(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}

			cat := NewCatalog(org.client)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil

		}
	}

	return Catalog{}, nil
}

// If user specifies valid vdc name then this returns a vdc object.
// If no vdc is found, then it returns an empty vdc and no error.
// Otherwise it returns an empty vdc and an error.
func (org *Org) GetVdcByName(vdcname string) (Vdc, error) {
	for _, link := range org.Org.Link {
		if link.Name == vdcname {
			vdcHREF, err := url.ParseRequestURI(link.HREF)
			if err != nil {
				return Vdc{}, fmt.Errorf("Error parsing url: %v", err)
			}
			req := org.client.NewRequest(map[string]string{}, "GET", *vdcHREF, nil)
			resp, err := checkResp(org.client.Http.Do(req))
			if err != nil {
				return Vdc{}, fmt.Errorf("error getting vdc: %s", err)
			}

			vdc := NewVdc(org.client)
			if err = decodeBody(resp, vdc.Vdc); err != nil {
				return Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
			}
			// The request was successful
			return *vdc, nil
		}
	}
	return Vdc{}, nil
}

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

// Given an adminorg with a valid HREF, the function refetches the adminorg
// and updates the user's adminorg data. Otherwise if the function fails,
// it returns an error.  Users should use refresh whenever they have
// a stale org due to the creation/update/deletion of a resource
// within the org or the org itself.
func (adminOrg *AdminOrg) Refresh() error {
	if *adminOrg == (AdminOrg{}) {
		return fmt.Errorf("cannot refresh, Object is empty")
	}
	adminOrgHREF, _ := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	req := adminOrg.client.NewRequest(map[string]string{}, "GET", *adminOrgHREF, nil)
	resp, err := checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error performing request: %s", err)
	}
	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledAdminOrg := &types.AdminOrg{}
	if err = decodeBody(resp, unmarshalledAdminOrg); err != nil {
		return fmt.Errorf("error decoding org response: %s", err)
	}
	adminOrg.AdminOrg = unmarshalledAdminOrg
	// The request was successful
	return nil
}

// CreateCatalog creates a catalog with given name and description under the
// the given organization. Returns an AdminCatalog that contains a creation
// task.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-CreateCatalog.html
func (adminOrg *AdminOrg) CreateCatalog(Name, Description string, isPublished bool) (AdminCatalog, error) {

	vcomp := &types.AdminCatalog{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        Name,
		Description: Description,
		IsPublished: isPublished,
	}

	catalogHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return AdminCatalog{}, fmt.Errorf("error parsing admin org's href: %v", err)
	}
	catalogHREF.Path += "/catalogs"

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	xmlData := bytes.NewBufferString(xml.Header + string(output))

	req := adminOrg.client.NewRequest(map[string]string{}, "POST", *catalogHREF, xmlData)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.catalog+xml")

	resp, err := checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return AdminCatalog{}, fmt.Errorf("error creating catalog: %s : %s", err, catalogHREF.Path)
	}

	catalog := NewAdminCatalog(adminOrg.client)
	if err = decodeBody(resp, catalog.AdminCatalog); err != nil {
		return AdminCatalog{}, fmt.Errorf("error decoding task response: %s", err)
	}
	// Task is within the catalog
	return *catalog, nil

}

// If user specifies valid vdc name then this returns a vdc object.
// If no vdc is found, then it returns an empty vdc and no error.
// Otherwise it returns an empty vdc and an error. This function
// allows users to use an AdminOrg to fetch a vdc as well.
func (adminOrg *AdminOrg) GetVdcByName(vdcname string) (Vdc, error) {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdcs.Name == vdcname {
			splitbyAdminHREF := strings.Split(vdcs.HREF, "/admin")
			vdcHREF := splitbyAdminHREF[0] + splitbyAdminHREF[1]
			vdcURL, err := url.ParseRequestURI(vdcHREF)
			if err != nil {
				return Vdc{}, fmt.Errorf("Error parsing url: %v", err)
			}
			req := adminOrg.client.NewRequest(map[string]string{}, "GET", *vdcURL, nil)
			resp, err := checkResp(adminOrg.client.Http.Do(req))
			if err != nil {
				return Vdc{}, fmt.Errorf("error getting vdc: %s", err)
			}

			vdc := NewVdc(adminOrg.client)
			if err = decodeBody(resp, vdc.Vdc); err != nil {
				return Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
			}
			// The request was successful
			return *vdc, nil
		}
	}
	return Vdc{}, nil
}

//   Deletes the org, returning an error if the vCD call fails.
//   API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Organization.html
func (adminOrg *AdminOrg) Delete(force bool, recursive bool) error {
	if force && recursive {
		//undeploys vapps
		err := adminOrg.undeployAllVApps()
		if err != nil {
			return fmt.Errorf("error could not undeploy: %#v", err)
		}
		//removes vapps
		err = adminOrg.removeAllVApps()
		if err != nil {
			return fmt.Errorf("error could not remove vapp: %#v", err)
		}
		//removes catalogs
		err = adminOrg.removeCatalogs()
		if err != nil {
			return fmt.Errorf("error could not remove all catalogs: %#v", err)
		}
		//removes networks
		err = adminOrg.removeAllOrgNetworks()
		if err != nil {
			return fmt.Errorf("error could not remove all networks: %#v", err)
		}
		//removes org vdcs
		err = adminOrg.removeAllOrgVDCs()
		if err != nil {
			return fmt.Errorf("error could not remove all vdcs: %#v", err)
		}
	}
	// Disable org
	err := adminOrg.Disable()
	if err != nil {
		return fmt.Errorf("error disabling Org %s: %s", adminOrg.AdminOrg.ID, err)
	}
	// Get admin HREF
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	req := adminOrg.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", *orgHREF, nil)
	_, err = checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error deleting Org %s: %s", adminOrg.AdminOrg.ID, err)
	}
	return nil
}

// Disables the org. Returns an error if the call to vCD fails.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-DisableOrg.html
func (adminOrg *AdminOrg) Disable() error {
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	orgHREF.Path += "/action/disable"
	req := adminOrg.client.NewRequest(map[string]string{}, "POST", *orgHREF, nil)
	_, err = checkResp(adminOrg.client.Http.Do(req))
	return err
}

//   Updates the Org definition from current org struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails.
//   API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Organization.html
func (adminOrg *AdminOrg) Update() (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        adminOrg.AdminOrg.Name,
		IsEnabled:   adminOrg.AdminOrg.IsEnabled,
		FullName:    adminOrg.AdminOrg.FullName,
		OrgSettings: adminOrg.AdminOrg.OrgSettings,
	}
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	xmlData := bytes.NewBufferString(xml.Header + string(output))
	// Update org
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	req := adminOrg.client.NewRequest(map[string]string{}, "PUT", *orgHREF, xmlData)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error updating Org: %s", err)
	}
	// Create Return object
	task := NewTask(adminOrg.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
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
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.undeployAllVdcVApps()
		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
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
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.removeAllVdcVApps()
		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
		}
	}
	return nil
}

// Gets a vdc within org associated with an admin vdc url
func (adminOrg *AdminOrg) getVdcByAdminHREF(adminVdcUrl *url.URL) (*Vdc, error) {
	// get non admin vdc path
	non_admin := strings.Split(adminVdcUrl.Path, "/admin")
	adminVdcUrl.Path = non_admin[0] + non_admin[1]
	req := adminOrg.client.NewRequest(map[string]string{}, "GET", *adminVdcUrl, nil)
	resp, err := checkResp(adminOrg.client.Http.Do(req))
	if err != nil {
		return &Vdc{}, fmt.Errorf("error retreiving vdc: %s", err)
	}

	vdc := NewVdc(adminOrg.client)
	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return &Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	return vdc, nil
}

// Removes all vdcs in a org
func (adminOrg *AdminOrg) removeAllOrgVDCs() error {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		// Get admin Vdc HREF
		adminVdcUrl := adminOrg.client.VCDHREF
		adminVdcUrl.Path += "/admin/vdc/" + strings.Split(vdcs.HREF, "/vdc/")[1] + "/action/disable"
		req := adminOrg.client.NewRequest(map[string]string{}, "POST", adminVdcUrl, nil)
		_, err := checkResp(adminOrg.client.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error disabling vdc: %s", err)
		}
		// Get admin vdc HREF for normal deletion
		adminVdcUrl.Path = strings.Split(adminVdcUrl.Path, "/action/disable")[0]
		req = adminOrg.client.NewRequest(map[string]string{
			"recursive": "true",
			"force":     "true",
		}, "DELETE", adminVdcUrl, nil)
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
			return fmt.Errorf("Couldn't finish removing vdc %#v", err)
		}

	}

	return nil
}

// Removes All networks in the org
func (adminOrg *AdminOrg) removeAllOrgNetworks() error {
	for _, networks := range adminOrg.AdminOrg.Networks.Networks {
		// Get Network HREF
		networkHREF := adminOrg.client.VCDHREF
		networkHREF.Path += "/admin/network/" + strings.Split(networks.HREF, "/network/")[1] //gets id
		req := adminOrg.client.NewRequest(map[string]string{}, "DELETE", networkHREF, nil)
		resp, err := checkResp(adminOrg.client.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting newtork: %s, %s", err, networkHREF.Path)
		}

		task := NewTask(adminOrg.client)
		if err = decodeBody(resp, task.Task); err != nil {
			return fmt.Errorf("error decoding task response: %s", err)
		}
		if task.Task.Status == "error" {
			return fmt.Errorf("network not properly destroyed")
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Couldn't finish removing network %#v", err)
		}
	}
	return nil
}

// Forced removal of all organization catalogs
func (adminOrg *AdminOrg) removeCatalogs() error {
	for _, catalogs := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		catalogHREF := adminOrg.client.VCDHREF
		catalogHREF.Path += "/admin/catalog/" + strings.Split(catalogs.HREF, "/catalog/")[1] //gets id
		req := adminOrg.client.NewRequest(map[string]string{
			"force":     "true",
			"recursive": "true",
		}, "DELETE", catalogHREF, nil)
		_, err := checkResp(adminOrg.client.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting catalog: %s, %s", err, catalogHREF.Path)
		}
	}
	return nil

}

// Given a valid catalog name, FindCatalog returns an AdminCatalog object.
// If no catalog is found, then returns an empty AdminCatalog and no error.
// Otherwise it returns an error. Function allows user to use an AdminOrg
// to also fetch a Catalog. If user does not have proper credentials to
// perform administrator tasks then function returns an error.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/GET-Catalog-AdminView.html
func (adminOrg *AdminOrg) FindAdminCatalog(catalogName string) (AdminCatalog, error) {
	for _, adminCatalog := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if adminCatalog.Name == catalogName {
			catalogURL, err := url.ParseRequestURI(adminCatalog.HREF)
			if err != nil {
				return AdminCatalog{}, fmt.Errorf("error decoding catalog url: %s", err)
			}
			req := adminOrg.client.NewRequest(map[string]string{}, "GET", *catalogURL, nil)
			resp, err := checkResp(adminOrg.client.Http.Do(req))
			if err != nil {
				return AdminCatalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}
			adminCatalog := NewAdminCatalog(adminOrg.client)
			if err = decodeBody(resp, adminCatalog.AdminCatalog); err != nil {
				return AdminCatalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}
			// The request was successful
			return *adminCatalog, nil
		}
	}
	return AdminCatalog{}, nil
}

// Given a valid catalog name, FindCatalog returns a Catalog object.
// If no catalog is found, then returns an empty catalog and no error.
// Otherwise it returns an error. Function allows user to use an AdminOrg
// to also fetch a Catalog.
func (adminOrg *AdminOrg) FindCatalog(catalogName string) (Catalog, error) {
	for _, catalogs := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if catalogs.Name == catalogName {
			splitbyAdminHREF := strings.Split(catalogs.HREF, "/admin")
			catalogHREF := splitbyAdminHREF[0] + splitbyAdminHREF[1]
			catalogURL, err := url.ParseRequestURI(catalogHREF)
			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog url: %s", err)
			}
			req := adminOrg.client.NewRequest(map[string]string{}, "GET", *catalogURL, nil)
			resp, err := checkResp(adminOrg.client.Http.Do(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}
			cat := NewCatalog(adminOrg.client)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil
		}
	}
	return Catalog{}, nil
}
