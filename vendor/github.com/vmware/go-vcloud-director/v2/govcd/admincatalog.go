/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// AdminCatalog is a admin view of a vCloud Director Catalog
// To be able to get an AdminCatalog representation, users must have
// admin credentials to the System org. AdminCatalog is used
// for creating, updating, and deleting a Catalog.
// Definition: https://code.vmware.com/apis/220/vcloud#/doc/doc/types/AdminCatalogType.html
type AdminCatalog struct {
	AdminCatalog *types.AdminCatalog
	client       *Client
}

func NewAdminCatalog(client *Client) *AdminCatalog {
	return &AdminCatalog{
		AdminCatalog: new(types.AdminCatalog),
		client:       client,
	}
}

// Deletes the Catalog, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Catalog.html
func (adminCatalog *AdminCatalog) Delete(force, recursive bool) error {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.Delete(force, recursive)
}

// Updates the Catalog definition from current Catalog struct contents.
// Any differences that may be legally applied will be updated.
// Returns an error if the call to vCD fails. Update automatically performs
// a refresh with the admin catalog it gets back from the rest api
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Catalog.html
func (adminCatalog *AdminCatalog) Update() error {
	reqCatalog := &types.Catalog{
		Name:        adminCatalog.AdminCatalog.Catalog.Name,
		Description: adminCatalog.AdminCatalog.Description,
	}
	vcomp := &types.AdminCatalog{
		Xmlns:       types.XMLNamespaceVCloud,
		Catalog:     *reqCatalog,
		IsPublished: adminCatalog.AdminCatalog.IsPublished,
	}
	catalog := &types.AdminCatalog{}
	_, err := adminCatalog.client.ExecuteRequest(adminCatalog.AdminCatalog.HREF, http.MethodPut,
		"application/vnd.vmware.admin.catalog+xml", "error updating catalog: %s", vcomp, catalog)
	adminCatalog.AdminCatalog = catalog
	return err
}

// Uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded. Files from ova are extracted to system
// temp folder "govcd+random number" and left for inspection on error.
func (adminCatalog *AdminCatalog) UploadOvf(ovaFileName, itemName, description string, uploadPieceSize int64) (UploadTask, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.UploadOvf(ovaFileName, itemName, description, uploadPieceSize)
}

func (adminCatalog *AdminCatalog) Refresh() error {
	if *adminCatalog == (AdminCatalog{}) || adminCatalog.AdminCatalog.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty or HREF is empty")
	}

	refreshedCatalog := &types.AdminCatalog{}

	_, err := adminCatalog.client.ExecuteRequest(adminCatalog.AdminCatalog.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, refreshedCatalog)
	if err != nil {
		return err
	}
	adminCatalog.AdminCatalog = refreshedCatalog

	return nil
}

// getOrgInfo finds the organization to which the admin catalog belongs, and returns its name and ID
func (adminCatalog *AdminCatalog) getOrgInfo() (orgInfoType, error) {
	return getOrgInfo(adminCatalog.client, adminCatalog.AdminCatalog.Link, adminCatalog.AdminCatalog.ID, adminCatalog.AdminCatalog.Name, "AdminCatalog")
}
