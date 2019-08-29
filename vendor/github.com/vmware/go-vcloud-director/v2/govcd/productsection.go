/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// setProductSectionList is a shared function for both vApp and VM
func setProductSectionList(client *Client, href string, productSection *types.ProductSectionList) error {
	if href == "" {
		return fmt.Errorf("href cannot be empty to set product section")
	}

	productSection.Xmlns = types.XMLNamespaceVCloud
	productSection.Ovf = types.XMLNamespaceOVF

	task, err := client.ExecuteTaskRequest(href+"/productSections", http.MethodPut,
		types.MimeProductSection, "error setting product section: %s", productSection)

	if err != nil {
		return fmt.Errorf("unable to set product section: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("task for setting product section failed: %s", err)
	}

	return nil
}

// getProductSectionList is a shared function for both vApp and VM
func getProductSectionList(client *Client, href string) (*types.ProductSectionList, error) {
	if href == "" {
		return nil, fmt.Errorf("href cannot be empty to get product section")
	}
	productSection := &types.ProductSectionList{}

	_, err := client.ExecuteRequest(href+"/productSections", http.MethodGet,
		types.MimeProductSection, "error retrieving product section : %s", nil, productSection)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve product section: %s", err)
	}

	return productSection, nil
}
