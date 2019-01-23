/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

func GetExternalNetworkByName(vcdClient *VCDClient, networkName string) (*types.ExternalNetworkReference, error) {
	extNetworkRefs := &types.ExternalNetworkReferences{}

	extNetworkHREF, err := getExternalNetworkHref(vcdClient)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	extNetworkURL, err := url.ParseRequestURI(extNetworkHREF)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", *extNetworkURL, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving external networks: %s", err)
		return &types.ExternalNetworkReference{}, fmt.Errorf("error retrieving external networks: %s", err)
	}

	if err = decodeBody(resp, extNetworkRefs); err != nil {
		util.Logger.Printf("[TRACE] error retrieving  external networks: %s", err)
		return &types.ExternalNetworkReference{}, fmt.Errorf("error decoding extension  external networks: %s", err)
	}

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			return netRef, nil
		}
	}

	return &types.ExternalNetworkReference{}, nil
}

func getExternalNetworkHref(vcdClient *VCDClient) (string, error) {
	extensions, err := getExtension(vcdClient)
	if err != nil {
		return "", err
	}

	for _, extensionLink := range extensions.Link {
		if extensionLink.Type == "application/vnd.vmware.admin.vmwExternalNetworkReferences+xml" {
			return extensionLink.HREF, nil
		}
	}

	return "", errors.New("external network link isn't found")
}

func getExtension(vcdClient *VCDClient) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := vcdClient.Client.VCDHREF
	extensionHREF.Path += "/admin/extension/"
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", extensionHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[TRACE] error retrieving extension: %s", err)
		return extensions, fmt.Errorf("error retrieving extension: %s", err)
	}

	if err = decodeBody(resp, extensions); err != nil {
		util.Logger.Printf("[TRACE] error retrieving extension list: %s", err)
		return extensions, fmt.Errorf("error decoding extension list response: %s", err)
	}

	return extensions, nil
}
