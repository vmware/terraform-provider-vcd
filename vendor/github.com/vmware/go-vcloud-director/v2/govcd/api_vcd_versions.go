/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	semver "github.com/hashicorp/go-version"

	"github.com/vmware/go-vcloud-director/v2/util"
)

type VersionInfo struct {
	Version    string `xml:"Version"`
	LoginUrl   string `xml:"LoginUrl"`
	Deprecated bool   `xml:"deprecated,attr,omitempty"`
}

type VersionInfos []VersionInfo

type SupportedVersions struct {
	VersionInfos `xml:"VersionInfo"`
}

// apiVersionToVcdVersion gets the vCD version from max supported API version
var apiVersionToVcdVersion = map[string]string{
	"29.0": "9.0",
	"30.0": "9.1",
	"31.0": "9.5",
	"32.0": "9.7",
	"33.0": "10.0",
	"34.0": "10.1",
}

// vcdVersionToApiVersion gets the max supported API version from vCD version
var vcdVersionToApiVersion = map[string]string{
	"9.0":  "29.0",
	"9.1":  "30.0",
	"9.5":  "31.0",
	"9.7":  "32.0",
	"10.0": "33.0",
	"10.1": "34.0",
}

// to make vcdVersionToApiVersion used
var _ = vcdVersionToApiVersion

// APIVCDMaxVersionIs compares against maximum vCD supported API version from /api/versions (not necessarily
// the currently used one). This allows to check what is the maximum API version that vCD instance
// supports and can be used to guess vCD product version. API 31.0 support was first introduced in
// vCD 9.5 (as per https://code.vmware.com/doc/preview?id=8072). Therefore APIMaxVerIs(">= 31.0")
// implies that you have vCD 9.5 or newer running inside.
// It does not require for the client to be authenticated.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (cli *Client) APIVCDMaxVersionIs(versionConstraint string) bool {
	err := cli.vcdFetchSupportedVersions()
	if err != nil {
		util.Logger.Printf("[ERROR] could not retrieve supported versions: %s", err)
		return false
	}

	util.Logger.Printf("[TRACE] checking max API version against constraints '%s'", versionConstraint)
	maxVersion, err := cli.maxSupportedVersion()
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	isSupported, err := cli.apiVersionMatchesConstraint(maxVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	return isSupported
}

// APIClientVersionIs allows to compare against currently used API version VCDClient.Client.APIVersion.
// Can be useful to validate if a certain feature can be used or not.
// It does not require for the client to be authenticated.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (cli *Client) APIClientVersionIs(versionConstraint string) bool {

	util.Logger.Printf("[TRACE] checking current API version against constraints '%s'", versionConstraint)

	isSupported, err := cli.apiVersionMatchesConstraint(cli.APIVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find cur supported version : %s", err)
		return false
	}

	return isSupported
}

// vcdFetchSupportedVersions retrieves list of supported versions from
// /api/versions endpoint and stores them in VCDClient for future uses.
// It only does it once.
func (cli *Client) vcdFetchSupportedVersions() error {
	// Only fetch /versions if it is not stored already
	numVersions := len(cli.supportedVersions.VersionInfos)
	if numVersions > 0 {
		util.Logger.Printf("[TRACE] skipping fetch of versions because %d are stored", numVersions)
		return nil
	}

	apiEndpoint := cli.VCDHREF
	apiEndpoint.Path += "/versions"

	suppVersions := new(SupportedVersions)
	_, err := cli.ExecuteRequest(apiEndpoint.String(), http.MethodGet,
		"", "error fetching versions: %s", nil, suppVersions)

	cli.supportedVersions = *suppVersions

	// Log all supported API versions in one line to help identify vCD version from logs
	allApiVersions := make([]string, len(cli.supportedVersions.VersionInfos))
	for versionIndex, version := range cli.supportedVersions.VersionInfos {
		allApiVersions[versionIndex] = version.Version
	}
	util.Logger.Printf("[DEBUG] supported API versions : %s", strings.Join(allApiVersions, ","))

	return err
}

// maxSupportedVersion parses supported version list and returns the highest version in string format.
func (cli *Client) maxSupportedVersion() (string, error) {
	versions := make([]*semver.Version, len(cli.supportedVersions.VersionInfos))
	for index, versionInfo := range cli.supportedVersions.VersionInfos {
		version, _ := semver.NewVersion(versionInfo.Version)
		versions[index] = version
	}
	// Sort supported versions in order lowest-highest
	sort.Sort(semver.Collection(versions))

	switch {
	case len(versions) > 1:
		return versions[len(versions)-1].Original(), nil
	case len(versions) == 1:
		return versions[0].Original(), nil
	default:
		return "", fmt.Errorf("could not identify supported versions")
	}
}

// vcdCheckSupportedVersion checks if there is at least one specified version exactly matching listed ones.
// Format example "27.0"
func (cli *Client) vcdCheckSupportedVersion(version string) (bool, error) {
	return cli.checkSupportedVersionConstraint(fmt.Sprintf("= %s", version))
}

// Checks if there is at least one specified version matching the list returned by vCD.
// Constraint format can be in format ">= 27.0, < 32",">= 30" ,"= 27.0".
func (cli *Client) checkSupportedVersionConstraint(versionConstraint string) (bool, error) {
	for _, versionInfo := range cli.supportedVersions.VersionInfos {
		versionMatch, err := cli.apiVersionMatchesConstraint(versionInfo.Version, versionConstraint)
		if err != nil {
			return false, fmt.Errorf("cannot match version: %s", err)
		}

		if versionMatch {
			return true, nil
		}
	}
	return false, fmt.Errorf("version %s is not supported", versionConstraint)
}

func (cli *Client) apiVersionMatchesConstraint(version, versionConstraint string) (bool, error) {

	checkVer, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse version %s : %s", version, err)
	}
	// Create a provided constraint to check against current max version
	constraints, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse given version constraint '%s' : %s", versionConstraint, err)
	}
	if constraints.Check(checkVer) {
		util.Logger.Printf("[INFO] API version %s satisfies constraints '%s'", checkVer, constraints)
		return true, nil
	}

	util.Logger.Printf("[TRACE] API version %s does not satisfy constraints '%s'", checkVer, constraints)
	return false, nil
}

// validateAPIVersion fetches API versions
func (cli *Client) validateAPIVersion() error {
	err := cli.vcdFetchSupportedVersions()
	if err != nil {
		return fmt.Errorf("could not retrieve supported versions: %s", err)
	}

	// Check if version is supported
	if ok, err := cli.vcdCheckSupportedVersion(cli.APIVersion); !ok || err != nil {
		return fmt.Errorf("API version %s is not supported: %s", cli.APIVersion, err)
	}

	return nil
}

// GetSpecificApiVersionOnCondition returns default version or wantedApiVersion if it is connected to version
// described in vcdApiVersionCondition
// f.e. values ">= 32.0", "32.0" returns 32.0 if vCD version is above or 9.7
func (cli *Client) GetSpecificApiVersionOnCondition(vcdApiVersionCondition, wantedApiVersion string) string {
	apiVersion := cli.APIVersion
	if cli.APIVCDMaxVersionIs(vcdApiVersionCondition) {
		apiVersion = wantedApiVersion
	}
	return apiVersion
}
