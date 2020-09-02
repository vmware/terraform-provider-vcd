/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// orgInfoType is the basic information about an organization (needed for tenant context)
type orgInfoType struct {
	id   string
	name string
}

// orgInfoCache is a cache to save org information, avoid repeated calls to compute the same result.
// The keys to this map are the requesting objects IDs.
var orgInfoCache = make(map[string]orgInfoType)

// GetAccessControl retrieves the access control information for the requested entity
func (client Client) GetAccessControl(href, entityType, entityName string, headerValues map[string]string) (*types.ControlAccessParams, error) {

	href += "/controlAccess"
	var controlAccess types.ControlAccessParams

	acUrl, err := url.ParseRequestURI(href)
	if err != nil {
		return nil, fmt.Errorf("[client.GetAccessControl] error parsing HREF %s: %s", href, err)
	}
	var additionalHeader = make(http.Header)

	if len(headerValues) > 0 {
		for k, v := range headerValues {
			additionalHeader.Add(k, v)
		}
	}
	req := client.newRequest(
		nil,               // params
		nil,               // notEncodedParams
		http.MethodGet,    // method
		*acUrl,            // reqUrl
		nil,               // body
		client.APIVersion, // apiVersion
		additionalHeader,  // additionalHeader
	)

	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("[client.GetAccessControl] error checking response to request %s: %s", href, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("[client.GetAccessControl] nil response received")
	}
	if err = decodeBody(types.BodyTypeXML, resp, &controlAccess); err != nil {
		return nil, fmt.Errorf("[client.GetAccessControl] error decoding response: %s", err)
	}

	return &controlAccess, nil
}

// SetAccessControl changes the access control information for this entity
// There are two ways of setting the access:
// with accessControl.IsSharedToEveryone = true we give access to everyone
// with accessControl.IsSharedToEveryone = false, accessControl.AccessSettings defines which subjects can access the vApp
// For each setting we must provide:
// * The subject (HREF and Type are mandatory)
// * The access level (one of ReadOnly, Change, FullControl)
func (client *Client) SetAccessControl(accessControl *types.ControlAccessParams, href, entityType, entityName string, headerValues map[string]string) error {

	href += "/action/controlAccess"
	// Make sure that subjects in the setting list are used only once
	if accessControl.AccessSettings != nil && len(accessControl.AccessSettings.AccessSetting) > 0 {
		if accessControl.IsSharedToEveryone {
			return fmt.Errorf("[client.SetAccessControl] can't set IsSharedToEveryone and AccessSettings at the same time for %s %s (%s)", entityType, entityName, href)
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

	accessControl.Xmlns = types.XMLNamespaceVCloud
	queryUrl, err := url.ParseRequestURI(href)
	if err != nil {
		return fmt.Errorf("[client.SetAccessControl] error parsing HREF %s: %s", href, err)
	}

	var header = make(http.Header)
	if len(headerValues) > 0 {
		for k, v := range headerValues {
			header.Add(k, v)
		}
	}

	marshaledXml, err := xml.MarshalIndent(accessControl, "  ", "    ")
	if err != nil {
		return fmt.Errorf("[client.SetAccessControl] error marshalling xml data: %s", err)
	}
	body := bytes.NewBufferString(xml.Header + string(marshaledXml))

	req := client.newRequest(
		nil,               // params
		nil,               // notEncodedParams
		http.MethodPost,   // method
		*queryUrl,         // reqUrl
		body,              // body
		client.APIVersion, // apiVersion
		header,            // additionalHeader
	)

	resp, err := checkResp(client.Http.Do(req))

	if err != nil {
		return fmt.Errorf("[client.SetAccessControl] error checking response to HREF %s: %s", href, err)
	}
	if resp == nil {
		return fmt.Errorf("[client.SetAccessControl] nil response received")
	}
	_, err = checkResp(resp, err)
	return err
}

// GetAccessControl retrieves the access control information for this vApp
func (vapp VApp) GetAccessControl(useTenantContext bool) (*types.ControlAccessParams, error) {

	if vapp.VApp.HREF == "" {
		return nil, fmt.Errorf("vApp HREF is empty")
	}
	// if useTenantContext is false, we use an empty header (= default behavior)
	// if it is true, we use a header populated with tenant context values
	accessControlHeader, err := vapp.getAccessControlHeader(useTenantContext)
	if err != nil {
		return nil, err
	}
	return vapp.client.GetAccessControl(vapp.VApp.HREF, "vApp", vapp.VApp.Name, accessControlHeader)
}

// SetAccessControl changes the access control information for this vApp
func (vapp VApp) SetAccessControl(accessControl *types.ControlAccessParams, useTenantContext bool) error {

	if vapp.VApp.HREF == "" {
		return fmt.Errorf("vApp HREF is empty")
	}

	// if useTenantContext is false, we use an empty header (= default behavior)
	// if it is true, we use a header populated with tenant context values
	accessControlHeader, err := vapp.getAccessControlHeader(useTenantContext)
	if err != nil {
		return err
	}
	return vapp.client.SetAccessControl(accessControl, vapp.VApp.HREF, "vApp", vapp.VApp.Name, accessControlHeader)

}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (vapp VApp) RemoveAccessControl(useTenantContext bool) error {
	return vapp.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false}, useTenantContext)
}

// IsShared shows whether a vApp is shared or not, regardless of the number of subjects sharing it
func (vapp VApp) IsShared(useTenantContext bool) (bool, error) {
	settings, err := vapp.GetAccessControl(useTenantContext)
	if err != nil {
		return false, err
	}
	if settings.IsSharedToEveryone {
		return true, nil
	}
	return settings.AccessSettings != nil, nil
}

// GetAccessControl retrieves the access control information for this catalog
func (adminCatalog AdminCatalog) GetAccessControl(useTenantContext bool) (*types.ControlAccessParams, error) {

	if adminCatalog.AdminCatalog.HREF == "" {
		return nil, fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(adminCatalog.AdminCatalog.HREF, "/admin/", "/", 1)

	// if useTenantContext is false, we use an empty header (= default behavior)
	// if it is true, we use a header populated with tenant context values
	accessControlHeader, err := adminCatalog.getAccessControlHeader(useTenantContext)
	if err != nil {
		return nil, err
	}
	return adminCatalog.client.GetAccessControl(href, "catalog", adminCatalog.AdminCatalog.Name, accessControlHeader)
}

// SetAccessControl changes the access control information for this catalog
func (adminCatalog AdminCatalog) SetAccessControl(accessControl *types.ControlAccessParams, useTenantContext bool) error {

	if adminCatalog.AdminCatalog.HREF == "" {
		return fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(adminCatalog.AdminCatalog.HREF, "/admin/", "/", 1)

	// if useTenantContext is false, we use an empty header (= default behavior)
	// if it is true, we use a header populated with tenant context values
	accessControlHeader, err := adminCatalog.getAccessControlHeader(useTenantContext)
	if err != nil {
		return err
	}
	return adminCatalog.client.SetAccessControl(accessControl, href, "catalog", adminCatalog.AdminCatalog.Name, accessControlHeader)
}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (adminCatalog AdminCatalog) RemoveAccessControl(useTenantContext bool) error {
	return adminCatalog.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false}, useTenantContext)
}

// IsShared shows whether a catalog is shared or not, regardless of the number of subjects sharing it
func (adminCatalog AdminCatalog) IsShared(useTenantContext bool) (bool, error) {
	settings, err := adminCatalog.GetAccessControl(useTenantContext)
	if err != nil {
		return false, err
	}
	if settings.IsSharedToEveryone {
		return true, nil
	}
	return settings.AccessSettings != nil, nil
}

// GetVappAccessControl is a convenience method to retrieve access control for a vApp
// from a VDC.
// The input variable vappIdentifier can be either the vApp name or its ID
func (vdc *Vdc) GetVappAccessControl(vappIdentifier string, useTenantContext bool) (*types.ControlAccessParams, error) {
	vapp, err := vdc.GetVAppByNameOrId(vappIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vApp %s: %s", vappIdentifier, err)
	}
	return vapp.GetAccessControl(useTenantContext)
}

// GetCatalogAccessControl is a convenience method to retrieve access control for a catalog
// from an organization.
// The input variable catalogIdentifier can be either the catalog name or its ID
func (org *AdminOrg) GetCatalogAccessControl(catalogIdentifier string, useTenantContext bool) (*types.ControlAccessParams, error) {
	catalog, err := org.GetAdminCatalogByNameOrId(catalogIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving catalog %s: %s", catalogIdentifier, err)
	}
	return catalog.GetAccessControl(useTenantContext)
}

// GetCatalogAccessControl is a convenience method to retrieve access control for a catalog
// from an organization.
// The input variable catalogIdentifier can be either the catalog name or its ID
func (org *Org) GetCatalogAccessControl(catalogIdentifier string, useTenantContext bool) (*types.ControlAccessParams, error) {
	catalog, err := org.GetCatalogByNameOrId(catalogIdentifier, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving catalog %s: %s", catalogIdentifier, err)
	}
	return catalog.GetAccessControl(useTenantContext)
}

// GetAccessControl retrieves the access control information for this catalog
func (catalog Catalog) GetAccessControl(useTenantContext bool) (*types.ControlAccessParams, error) {

	if catalog.Catalog.HREF == "" {
		return nil, fmt.Errorf("catalog HREF is empty")
	}
	href := strings.Replace(catalog.Catalog.HREF, "/admin/", "/", 1)
	accessControlHeader, err := catalog.getAccessControlHeader(useTenantContext)
	if err != nil {
		return nil, err
	}
	return catalog.client.GetAccessControl(href, "catalog", catalog.Catalog.Name, accessControlHeader)
}

// SetAccessControl changes the access control information for this catalog
func (catalog Catalog) SetAccessControl(accessControl *types.ControlAccessParams, useTenantContext bool) error {

	if catalog.Catalog.HREF == "" {
		return fmt.Errorf("catalog HREF is empty")
	}

	href := strings.Replace(catalog.Catalog.HREF, "/admin/", "/", 1)

	// if useTenantContext is false, we use an empty header (= default behavior)
	// if it is true, we use a header populated with tenant context values
	accessControlHeader, err := catalog.getAccessControlHeader(useTenantContext)
	if err != nil {
		return err
	}
	return catalog.client.SetAccessControl(accessControl, href, "catalog", catalog.Catalog.Name, accessControlHeader)
}

// RemoveAccessControl is a shortcut to SetAccessControl with all access disabled
func (catalog Catalog) RemoveAccessControl(useTenantContext bool) error {
	return catalog.SetAccessControl(&types.ControlAccessParams{IsSharedToEveryone: false}, useTenantContext)
}

// IsShared shows whether a catalog is shared or not, regardless of the number of subjects sharing it
func (catalog Catalog) IsShared(useTenantContext bool) (bool, error) {
	settings, err := catalog.GetAccessControl(useTenantContext)
	if err != nil {
		return false, err
	}
	if settings.IsSharedToEveryone {
		return true, nil
	}
	return settings.AccessSettings != nil, nil
}

// getAccessControlHeader builds the data needed to set the header when tenant context is required.
// If useTenantContext is false, it returns an empty map.
// Otherwise, it finds the Org ID and name (going up in the hierarchy through the VDC)
// and creates the header data
func (vapp *VApp) getAccessControlHeader(useTenantContext bool) (map[string]string, error) {
	if !useTenantContext {
		return map[string]string{}, nil
	}
	orgInfo, err := vapp.getOrgInfo()
	if err != nil {
		return nil, err
	}
	return map[string]string{types.HeaderTenantContext: orgInfo.id, types.HeaderAuthContext: orgInfo.name}, nil
}

// getAccessControlHeader builds the data needed to set the header when tenant context is required.
// If useTenantContext is false, it returns an empty map.
// Otherwise, it finds the Org ID and name and creates the header data
func (catalog *Catalog) getAccessControlHeader(useTenantContext bool) (map[string]string, error) {
	if !useTenantContext {
		return map[string]string{}, nil
	}
	orgInfo, err := catalog.getOrgInfo()
	if err != nil {
		return nil, err
	}
	return map[string]string{types.HeaderTenantContext: orgInfo.id, types.HeaderAuthContext: orgInfo.name}, nil
}

// getAccessControlHeader builds the data needed to set the header when tenant context is required.
// If useTenantContext is false, it returns an empty map.
// Otherwise, it finds the Org ID and name and creates the header data
func (adminCatalog *AdminCatalog) getAccessControlHeader(useTenantContext bool) (map[string]string, error) {
	if !useTenantContext {
		return map[string]string{}, nil
	}
	orgInfo, err := adminCatalog.getOrgInfo()

	if err != nil {
		return nil, err
	}
	return map[string]string{types.HeaderTenantContext: orgInfo.id, types.HeaderAuthContext: orgInfo.name}, nil
}
