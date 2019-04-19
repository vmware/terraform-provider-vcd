/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/http"
)

type CatalogItem struct {
	CatalogItem *types.CatalogItem
	client      *Client
}

func NewCatalogItem(cli *Client) *CatalogItem {
	return &CatalogItem{
		CatalogItem: new(types.CatalogItem),
		client:      cli,
	}
}

func (catalogItem *CatalogItem) GetVAppTemplate() (VAppTemplate, error) {

	cat := NewVAppTemplate(catalogItem.client)

	_, err := catalogItem.client.ExecuteRequest(catalogItem.CatalogItem.Entity.HREF, http.MethodGet,
		"", "error retrieving vApp template: %s", nil, cat.VAppTemplate)

	// The request was successful
	return *cat, err

}

// Deletes the Catalog Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-CatalogItem.html
func (catalogItem *CatalogItem) Delete() error {
	util.Logger.Printf("[TRACE] Deleting catalog item: %#v", catalogItem.CatalogItem)
	catalogItemHREF := catalogItem.client.VCDHREF
	catalogItemHREF.Path += "/catalogItem/" + catalogItem.CatalogItem.ID[23:]

	util.Logger.Printf("[TRACE] Url for deleting catalog item: %#v and name: %s", catalogItemHREF, catalogItem.CatalogItem.Name)

	return catalogItem.client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete,
		"", "error deleting Catalog item: %s", nil)
}
