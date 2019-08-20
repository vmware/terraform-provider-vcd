/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// setGuestProperties is a shared function for
func setGuestProperties(client *Client, href string, properties *types.ProductSectionList) error {
	if href == "" {
		return fmt.Errorf("href cannot be empty to set guest properties")
	}

	properties.Xmlns = types.XMLNamespaceVCloud
	properties.Ovf = types.XMLNamespaceOVF

	task, err := client.ExecuteTaskRequest(href+"/productSections", http.MethodPut,
		types.MimeProductSection, "error setting guest properties: %s", properties)

	if err != nil {
		return fmt.Errorf("unable to set guest properties: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("task for setting guest properties failed: %s", err)
	}

	return nil
}

// getGuestProperties is a shared function for both vApp and VM
func getGuestProperties(client *Client, href string) (*types.ProductSectionList, error) {
	if href == "" {
		return nil, fmt.Errorf("href cannot be empty to set guest properties")
	}
	properties := &types.ProductSectionList{}

	_, err := client.ExecuteRequest(href+"/productSections", http.MethodGet,
		types.MimeProductSection, "error retrieving guest properties: %s", nil, properties)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve guest properties: %s", err)
	}

	return properties, nil
}
