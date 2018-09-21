/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"

	types "github.com/ukcloud/govcloudair/types/v56"
	"strings"
)

type AdminOrg struct {
	AdminOrg *types.AdminOrg
	Org
}

type Org struct {
	Org *types.Org
	c   *Client
}

func NewAdminOrg(c *Client) *AdminOrg {
	return &AdminOrg{
		AdminOrg: new(types.AdminOrg),
		Org: Org{
			c: c,
		},
	}
}

func NewOrg(c *Client) *Org {
	return &Org{
		Org: new(types.Org),
		c:   c,
	}
}

//If user specifies valid organization name and vdc name, then this returns a vdc object
func (o *Org) GetVDCFromName(vdcname string) (Vdc, error) {

	HREF := ""
	for _, a := range o.Org.Link {
		if a.Type == "application/vnd.vmware.vcloud.vdc+xml" && a.Name == vdcname {
			HREF = a.HREF
			break
		}
	}

	if HREF == "" {
		return Vdc{}, fmt.Errorf("Error finding VDC from VDCName")
	}

	u, err := url.ParseRequestURI(HREF)

	if err != nil {
		return Vdc{}, fmt.Errorf("Error retrieving VDC: %v", err)
	}
	req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return Vdc{}, fmt.Errorf("error retreiving vdc: %s", err)
	}

	vdc := NewVdc(o.c)

	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}

	// The request was successful
	return *vdc, nil
}

//   Refetches the underlying org resource so that all resource URLs are
//   current. You must call this if you wish to ensure Org state is current,
//   e.g., after adding a VDC or catalog to the org. Returns an error if the
//   call to vCD fails.
func (o *Org) Refresh() error {
	if o.Org.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(o.Org.HREF)

	req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(o.c.Http.Do(req))

	if resp.StatusCode == 404 && err != nil {
		return fmt.Errorf("Org does not exist")
	}

	if err != nil {
		return fmt.Errorf("error finding Org: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledOrg := &types.Org{}

	if err = decodeBody(resp, unmarshalledOrg); err != nil {
		return fmt.Errorf("error decoding org response: %s", err)
	}

	o.Org = unmarshalledOrg

	// The request was successful
	return nil
}

//   Deletes the org, returning an error if the vCD call fails.
//   This call is idempotent and may be safely called on a non-existent org.
func (o *AdminOrg) Delete(force bool, recursive bool) error {

	if force && recursive {

		//undeploys vapps
		err := o.undeployAllVApps()

		if err != nil {
			return fmt.Errorf("Could not undeploy with error %#v", err)
		}

		//removes vapps
		err = o.removeAllVApps()

		if err != nil {
			return fmt.Errorf("Could not remove vapp with error %#v", err)
		}

		//removes catalogs
		err = o.removeCatalogs()

		if err != nil {
			return fmt.Errorf("Could not remove all catalogs %#v", err)
		}

		//removes networks
		err = o.removeAllOrgNetworks()
		if err != nil {
			return fmt.Errorf("Could not remove all networks %#v", err)
		}

		//removes org vdcs
		err = o.removeAllOrgVDCs()
		if err != nil {
			return fmt.Errorf("Could not remove all vdcs %#v", err)
		}

	}

	err := o.Disable()

	if err != nil {
		return fmt.Errorf("error disabling Org %s: %s", o.AdminOrg.ID, err)
	}

	s := o.c.HREF
	s.Path += "/admin/org/" + o.AdminOrg.ID[15:]

	req := o.c.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", s, nil)

	_, err = checkResp(o.c.Http.Do(req))

	if err != nil {
		return fmt.Errorf("error deleting Org %s: %s", o.AdminOrg.ID, err)
	}

	return nil
}

// Disables the org. Returns an error if the call to vCD fails.
func (o *AdminOrg) Disable() error {
	s := o.c.HREF
	s.Path += "/admin/org/" + o.AdminOrg.ID[15:] + "/action/disable"

	req := o.c.NewRequest(map[string]string{}, "POST", s, nil)

	_, err := checkResp(o.c.Http.Do(req))
	return err
}

//   Updates the Org definition from current org struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails.
func (o *AdminOrg) Update() (Task, error) {

	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        o.AdminOrg.Name,
		IsEnabled:   o.AdminOrg.IsEnabled,
		FullName:    o.AdminOrg.FullName,
		OrgSettings: o.AdminOrg.OrgSettings,
	}

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")

	s := o.c.HREF
	s.Path += "/admin/org/" + o.AdminOrg.ID[15:]

	b := bytes.NewBufferString(xml.Header + string(output))

	req := o.c.NewRequest(map[string]string{}, "PUT", s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")

	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error updating Org: %s", err)
	}

	task := NewTask(o.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}

	return *task, nil
}

//undeploys every vapp within an organization
func (o *AdminOrg) undeployAllVApps() error {
	for _, a := range o.AdminOrg.Vdcs.Vdcs {

		u, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}

		vdc, err := o.getOrgVdc(u)

		if err != nil {
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", u.Path, err)
		}

		err = vdc.undeployAllVdcVApps()

		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
		}

	}

	return nil
}

//deletes every vapp within an organization
func (o *AdminOrg) removeAllVApps() error {

	for _, a := range o.AdminOrg.Vdcs.Vdcs {

		u, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}

		vdc, err := o.getOrgVdc(u)

		if err != nil {
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", u.Path, err)
		}

		err = vdc.removeAllVdcVApps()

		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
		}

	}

	return nil
}

//gets a vdc within org with associated with a url
func (o *AdminOrg) getOrgVdc(u *url.URL) (*Vdc, error) {

	non_admin := strings.Split(u.Path, "/admin")
	u.Path = non_admin[0] + non_admin[1]
	req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return &Vdc{}, fmt.Errorf("error retreiving vdc: %s", err)
	}

	vdc := NewVdc(o.c)

	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return &Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	return vdc, nil
}

//removes all vdcs in a org
func (o *AdminOrg) removeAllOrgVDCs() error {

	for _, a := range o.AdminOrg.Vdcs.Vdcs {
		u, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}

		vdc, err := o.getOrgVdc(u)
		if err != nil {
			return err
		}

		//split into different private functions
		s := o.c.HREF
		s.Path += "/admin/vdc/" + vdc.Vdc.ID[15:]

		copyPath := s.Path

		s.Path += "/action/disable"

		req := o.c.NewRequest(map[string]string{}, "POST", s, nil)

		_, err = checkResp(o.c.Http.Do(req))

		if err != nil {
			return fmt.Errorf("error disabling vdc: %s", err)
		}

		s.Path = copyPath

		req = o.c.NewRequest(map[string]string{
			"recursive": "true",
			"force":     "true",
		}, "DELETE", s, nil)

		resp, err := checkResp(o.c.Http.Do(req))

		if err != nil {
			return fmt.Errorf("error deleting vdc: %s", err)
		}

		task := NewTask(o.c)

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

//removes All networks in the org
func (o *AdminOrg) removeAllOrgNetworks() error {

	for _, a := range o.AdminOrg.Networks.Networks {
		u, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}

		s := o.c.HREF
		s.Path += "/admin/network/" + strings.Split(u.Path, "/network/")[1] //gets id

		req := o.c.NewRequest(map[string]string{}, "DELETE", s, nil)

		resp, err := checkResp(o.c.Http.Do(req))

		if err != nil {
			return fmt.Errorf("error deleting newtork: %s, %s", err, u.Path)
		}

		task := NewTask(o.c)

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

//forced removal of all organization catalogs
func (o *AdminOrg) removeCatalogs() error {

	for _, a := range o.AdminOrg.Catalogs.Catalog {
		u, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}

		s := o.c.HREF
		s.Path += "/admin/catalog/" + strings.Split(u.Path, "/catalog/")[1] //gets id

		req := o.c.NewRequest(map[string]string{
			"force":     "true",
			"recursive": "true",
		}, "DELETE", s, nil)

		_, err = checkResp(o.c.Http.Do(req))

		if err != nil {
			return fmt.Errorf("error deleting catalog: %s, %s", err, u.Path)
		}

	}

	return nil

}

func (o *Org) FindCatalog(catalog string) (Catalog, error) {

	for _, av := range o.Org.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.catalog+xml" && av.Name == catalog {
			u, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding org response: %s", err)
			}

			req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)

			resp, err := checkResp(o.c.Http.Do(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}

			cat := NewCatalog(o.c)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil

		}
	}

	return Catalog{}, fmt.Errorf("can't find catalog: %s", catalog)
}
