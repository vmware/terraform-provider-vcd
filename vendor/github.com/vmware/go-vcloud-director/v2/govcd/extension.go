/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
)

func GetExternalNetworkByName(vcdClient *VCDClient, networkName string) (*types.ExternalNetworkReference, error) {
	extNetworkRefs := &types.ExternalNetworkReferences{}

	extNetworkHREF, err := getExternalNetworkHref(vcdClient)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
	}

	_, err = vcdClient.Client.ExecuteRequest(extNetworkHREF, http.MethodGet,
		"", "error retrieving external networks: %s", nil, extNetworkRefs)
	if err != nil {
		return &types.ExternalNetworkReference{}, err
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

	_, err := vcdClient.Client.ExecuteRequest(extensionHREF.String(), http.MethodGet,
		"", "error retrieving extension: %s", nil, extensions)

	return extensions, err
}
