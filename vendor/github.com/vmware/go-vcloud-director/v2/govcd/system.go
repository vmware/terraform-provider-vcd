/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Creates an Organization based on settings, network, and org name.
// The Organization created will have these settings specified in the
// settings parameter. The settings variable is defined in types.go.
// Method will fail unless user has an admin token.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-CreateOrganization.html
// Organization creation in vCD has two bugs BZ 2177355, BZ 2228936 (fixes are in 9.1.0.3 and 9.5.0.2) which require
// organization settings to be provided as workarounds.
// At least one element among DelayAfterPowerOnSeconds, DeployedVMQuota, StoredVmQuota, UseServerBootSequence, getVdcQuota
// should be set when providing generalOrgSettings.
// If either VAppLeaseSettings or VAppTemplateLeaseSettings is provided then all elements need to have values, otherwise don't provide them at all.
// Overall elements must be in the correct order.
func CreateOrg(vcdClient *VCDClient, name string, fullName string, description string, settings *types.OrgSettings, isEnabled bool) (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        name,
		IsEnabled:   isEnabled,
		FullName:    fullName,
		Description: description,
		OrgSettings: settings,
	}

	orgCreateHREF := vcdClient.Client.VCDHREF
	orgCreateHREF.Path += "/admin/orgs"

	// Return the task
	return vcdClient.Client.ExecuteTaskRequest(orgCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.organization+xml", "error instantiating a new Org: %s", vcomp)

}

// If user specifies a valid organization name, then this returns a
// organization object. If no valid org is found, it returns an empty
// org and no error. Otherwise it returns an error and an empty
// Org object
func GetOrgByName(vcdClient *VCDClient, orgName string) (Org, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgName)
	if err != nil {
		return Org{}, fmt.Errorf("organization '%s' fetch failed: %#v", orgName, err)
	}
	org := NewOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgUrl, http.MethodGet,
		"", "error retrieving org list: %s", nil, org.Org)
	if err != nil {
		return Org{}, err
	}

	return *org, nil
}

// If user specifies valid organization name,
// then this returns an admin organization object.
// If no valid org is found, it returns an empty
// org and no error. Otherwise returns an empty AdminOrg
// and an error.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/GET-Organization-AdminView.html
func GetAdminOrgByName(vcdClient *VCDClient, orgName string) (AdminOrg, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgName)
	if err != nil {
		return AdminOrg{}, err
	}
	orgHREF := vcdClient.Client.VCDHREF
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/api/org/")[1]

	org := NewAdminOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgHREF.String(), http.MethodGet,
		"", "error retrieving org: %s", nil, org.AdminOrg)
	if err != nil {
		return AdminOrg{}, err
	}

	return *org, nil
}

// Returns the HREF of the org with the name orgName
func getOrgHREF(vcdClient *VCDClient, orgName string) (string, error) {
	orgListHREF := vcdClient.Client.VCDHREF
	orgListHREF.Path += "/org"

	orgList := new(types.OrgList)

	_, err := vcdClient.Client.ExecuteRequest(orgListHREF.String(), http.MethodGet,
		"", "error retrieving org list: %s", nil, orgList)
	if err != nil {
		return "", err
	}

	// Look for orgName within OrgList
	for _, org := range orgList.Org {
		if org.Name == orgName {
			return org.HREF, nil
		}
	}
	return "", fmt.Errorf("couldn't find org with name: %s. Please check Org name as it is case sensitive", orgName)
}
