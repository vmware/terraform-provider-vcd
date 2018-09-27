/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	types "github.com/vmware/go-vcloud-director/types/v56"
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

func (ci *CatalogItem) GetVAppTemplate() (VAppTemplate, error) {
	catalogItemUrl, err := url.ParseRequestURI(ci.CatalogItem.Entity.HREF)

	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding catalogitem response: %s", err)
	}

	req := ci.client.NewRequest(map[string]string{}, "GET", *catalogItemUrl, nil)

	resp, err := checkResp(ci.client.Http.Do(req))
	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error retreiving vapptemplate: %s", err)
	}

	cat := NewVAppTemplate(ci.client)

	if err = decodeBody(resp, cat.VAppTemplate); err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding vapptemplate response: %s", err)
	}

	// The request was successful
	return *cat, nil

}
