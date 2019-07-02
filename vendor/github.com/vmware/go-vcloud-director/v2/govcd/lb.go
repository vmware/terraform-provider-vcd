/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"path"
	"strings"
)

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
