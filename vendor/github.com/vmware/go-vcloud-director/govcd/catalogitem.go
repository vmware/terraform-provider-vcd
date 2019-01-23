/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
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
	catalogItemUrl, err := url.ParseRequestURI(catalogItem.CatalogItem.Entity.HREF)

	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding catalogitem response: %s", err)
	}

	req := catalogItem.client.NewRequest(map[string]string{}, "GET", *catalogItemUrl, nil)

	resp, err := checkResp(catalogItem.client.Http.Do(req))
	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error retrieving vapptemplate: %s", err)
	}

	cat := NewVAppTemplate(catalogItem.client)

	if err = decodeBody(resp, cat.VAppTemplate); err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding vapptemplate response: %s", err)
	}

	// The request was successful
	return *cat, nil

}

// Deletes the Catalog Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-CatalogItem.html
func (catalogItem *CatalogItem) Delete() error {
	util.Logger.Printf("[TRACE] Deleting catalog item: %#v", catalogItem.CatalogItem)
	catalogItemHREF := catalogItem.client.VCDHREF
	catalogItemHREF.Path += "/catalogItem/" + catalogItem.CatalogItem.ID[23:]

	util.Logger.Printf("[TRACE] Url for deleting catalog item: %#v and name: %s", catalogItemHREF, catalogItem.CatalogItem.Name)

	req := catalogItem.client.NewRequest(map[string]string{}, "DELETE", catalogItemHREF, nil)

	_, err := checkResp(catalogItem.client.Http.Do(req))

	if err != nil {
		return fmt.Errorf("error deleting Catalog item %s: %s", catalogItem.CatalogItem.ID, err)
	}

	return nil
}
