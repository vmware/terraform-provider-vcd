/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetAccessControl retrieves the access control information for the requested entity
func (client Client) GetAccessControl(href, entityType, entityName string) (*types.ControlAccessParams, error) {

	href += "/controlAccess"
	var vappControlAccess types.ControlAccessParams
	errorMessage := fmt.Sprintf("error retrieving access control for %s %s", entityType, entityName)
	resp, err := client.ExecuteRequest(href, http.MethodGet,
		types.MimeControlAccess, errorMessage+": %s", nil, &vappControlAccess)

	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("[client.GetAccessControl] nil response received")
	}
	return &vappControlAccess, nil
}

// SetAccessControl changes the access control information for this entity
//There are two ways of setting the access:
// with accessControl.IsSharedToEveryone = true we give access to everyone
// with accessControl.IsSharedToEveryone = false, accessControl.AccessSettings defines which subjects can access the vApp
// For each setting we must provide:
// * The subject (HREF and Type are mandatory)
// * The access level (one of ReadOnly, Change, FullControl)
func (client *Client) SetAccessControl(accessControl *types.ControlAccessParams, href, entityType, entityName string) error {

	href += "/action/controlAccess"
	// Make sure that subjects in the setting list are used only once
	if accessControl.AccessSettings != nil && len(accessControl.AccessSettings.AccessSetting) > 0 {
		if accessControl.IsSharedToEveryone {
			return fmt.Errorf("[client.SetAccessControl] can't set IsSharedToEveryone and AccessSettings at the same time for %s %s", entityType, entityName)
		}
		var used = make(map[string]bool)
		for _, setting := range accessControl.AccessSettings.AccessSetting {
			_, seen := used[setting.Subject.HREF]
			if seen {
				return fmt.Errorf("[client.SetAccessControl] subject %s (%s) used more than once", setting.Subject.Name, setting.Subject.HREF)
			}
			used[setting.Subject.HREF] = true
			if setting.Subject.Type == "" {
				return fmt.Errorf("[client.SetAccessControl] subject %s (%s) has no type defined", setting.Subject.Name, setting.Subject.HREF)
			}
		}
	}
	errorMessage := fmt.Sprintf("[client.SetAccessControl] error setting access control for %s %s", entityType, entityName)

	accessControl.Xmlns = types.XMLNamespaceVCloud
	resp, err := client.ExecuteRequest(href, http.MethodPost, types.MimeControlAccess,
		errorMessage+": %s", accessControl, nil)

	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("[client.SetAccessControl] nil response received")
	}
	return nil
}

// GetAccessControl retrieves the access control information for this vApp
func (vapp VApp) GetAccessControl() (*types.ControlAccessParams, error) {

	if vapp.VApp.HREF == "" {
		return nil, fmt.Errorf("vApp HREF is empty")
	}

	return vapp.client.GetAccessControl(vapp.VApp.HREF, "vApp", vapp.VApp.Name)
}

// SetAccessControl changes the access control information for this vApp
func (vapp VApp) SetAccessControl(accessControl *types.ControlAccessParams) error {

	if vapp.VApp.HREF == "" {
		return fmt.Errorf("vApp HREF is empty")
	}

	return vapp.client.SetAccessControl(accessControl, vapp.VApp.HREF, "vApp", vapp.VApp.Name)

}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (vapp VApp) RemoveAccessControl() error {
	return vapp.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false})
}

// IsShared shows whether a vApp is shared or not, regardless of the number of subjects sharing it
func (vapp VApp) IsShared() bool {
	settings, err := vapp.GetAccessControl()
	if err != nil {
		return false
	}
	if settings.IsSharedToEveryone {
		return true
	}
	return settings.AccessSettings != nil
}

// GetAccessControl retrieves the access control information for this catalog
func (catalog AdminCatalog) GetAccessControl() (*types.ControlAccessParams, error) {

	if catalog.AdminCatalog.HREF == "" {
		return nil, fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(catalog.AdminCatalog.HREF, "/admin/", "/", 1)

	return catalog.client.GetAccessControl(href, "catalog", catalog.AdminCatalog.Name)
}

// SetAccessControl changes the access control information for this catalog
func (catalog AdminCatalog) SetAccessControl(accessControl *types.ControlAccessParams) error {

	if catalog.AdminCatalog.HREF == "" {
		return fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(catalog.AdminCatalog.HREF, "/admin/", "/", 1)

	return catalog.client.SetAccessControl(accessControl, href, "catalog", catalog.AdminCatalog.Name)
}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (catalog AdminCatalog) RemoveAccessControl() error {
	return catalog.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false})
}

// IsShared shows whether a vApp is shared or not, regardless of the number of subjects sharing it
func (catalog AdminCatalog) IsShared() bool {
	settings, err := catalog.GetAccessControl()
	if err != nil {
		return false
	}
	if settings.IsSharedToEveryone {
		return true
	}
	return settings.AccessSettings != nil
}

// GetVappAccessControl is a convenience method to retrieve access control for a vApp
// from a VDC.
// The input variable vappIdentifier can be either the vApp name or its ID
func (vdc *Vdc) GetVappAccessControl(vappIdentifier string) (*types.ControlAccessParams, error) {
	vapp, err := vdc.GetVAppByNameOrId(vappIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vapp %s: %s", vappIdentifier, err)
	}
	return vapp.GetAccessControl()
}

// GetCatalogAccessControl is a convenience method to retrieve access control for a vApp
// from an organization.
// The input variable catalogIdentifier can be either the catalog name or its ID
func (org *AdminOrg) GetCatalogAccessControl(catalogIdentifier string) (*types.ControlAccessParams, error) {
	catalog, err := org.GetAdminCatalogByNameOrId(catalogIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving catalog %s: %s", catalogIdentifier, err)
	}
	return catalog.GetAccessControl()
}

// GetCatalogAccessControl is a convenience method to retrieve access control for a vApp
// from an organization.
// The input variable catalogIdentifier can be either the catalog name or its ID
func (org *Org) GetCatalogAccessControl(catalogIdentifier string) (*types.ControlAccessParams, error) {
	catalog, err := org.GetCatalogByNameOrId(catalogIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving catalog %s: %s", catalogIdentifier, err)
	}
	return catalog.GetAccessControl()
}

// GetAccessControl retrieves the access control information for this catalog
func (catalog Catalog) GetAccessControl() (*types.ControlAccessParams, error) {

	if catalog.Catalog.HREF == "" {
		return nil, fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(catalog.Catalog.HREF, "/admin/", "/", 1)

	return catalog.client.GetAccessControl(href, "catalog", catalog.Catalog.Name)
}

// SetAccessControl changes the access control information for this catalog
func (catalog Catalog) SetAccessControl(accessControl *types.ControlAccessParams) error {

	if catalog.Catalog.HREF == "" {
		return fmt.Errorf("catalog HREF is empty")
	}

	href := strings.Replace(catalog.Catalog.HREF, "/admin/", "/", 1)

	// When we set IsSharedToEveryone in the UI, what happens behind the scenes is that the request is changed
	// to a read/only share to all visible organizations

	if accessControl.IsSharedToEveryone {
		return fmt.Errorf("share to everyone is not allowed for catalogs. You should set the sharing for all orgs that you need")
	}

	return catalog.client.SetAccessControl(accessControl, href, "catalog", catalog.Catalog.Name)
}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (catalog Catalog) RemoveAccessControl() error {
	return catalog.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false})
}

// IsShared shows whether a vApp is shared or not, regardless of the number of subjects sharing it
func (catalog Catalog) IsShared() bool {
	settings, err := catalog.GetAccessControl()
	if err != nil {
		return false
	}
	if settings.IsSharedToEveryone {
		return true
	}
	return settings.AccessSettings != nil
}
