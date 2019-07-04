/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"path"
	"strings"
)

// extractNSXObjectIDfromPath parses proxied NSX API response Location header and
// extracts ID for newly created object from it. The object ID is the last element in path.
// It expects the path to have at least one "/" to be a valid path and cleans up the trailing slash
// if there is one.
//
// Sample locationPath from API: /network/edges/edge-3/loadbalancer/config/monitors/monitor-5
// Expected ID to be returned: monitor-5
func extractNSXObjectIDfromPath(locationPath string) (string, error) {
	if locationPath == "" {
		return "", fmt.Errorf("unable to get ID from empty path")
	}

	cleanPath := path.Clean(locationPath) // Removes trailing slash if there is one
	splitPath := strings.Split(cleanPath, "/")

	if len(splitPath) < 2 {
		return "", fmt.Errorf("path does not contain url path: %s", splitPath)
	}

	objectID := splitPath[len(splitPath)-1]

	return objectID, nil
}
