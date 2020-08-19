package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// endpointMinApiVersions holds mapping of OpenAPI endpoints and API versions they were introduced in.
var endpointMinApiVersions = map[string]string{
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles: "31.0",
}

// checkOpenApiEndpointCompatibility checks if VCD version (to which the client is connected) is sufficient to work with
// specified OpenAPI endpoint and returns either error, either Api version to use for calling that endpoint. This Api
// version can then be supplied to low level OpenAPI client functions.
func (client *Client) checkOpenApiEndpointCompatibility(endpoint string) (string, error) {
	minimumApiVersion, ok := endpointMinApiVersions[endpoint]
	if !ok {
		return "", fmt.Errorf("minimum API version for endopoint '%s' is not defined", endpoint)
	}

	if client.APIVCDMaxVersionIs("< " + minimumApiVersion) {
		maxSupportedVersion, err := client.maxSupportedVersion()
		if err != nil {
			return "", fmt.Errorf("error reading maximum supported API version: %s", err)
		}
		return "", fmt.Errorf("endpoint '%s' requires API version to support at least '%s'. Maximum supported version in this instance: '%s'",
			endpoint, minimumApiVersion, maxSupportedVersion)
	}

	return minimumApiVersion, nil
}
