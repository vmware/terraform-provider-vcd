/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
	"strings"
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
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	xmlData := bytes.NewBufferString(xml.Header + string(output))
	// Make Request
	orgCreateHREF := vcdClient.Client.VCDHREF
	orgCreateHREF.Path += "/admin/orgs"
	req := vcdClient.Client.NewRequest(map[string]string{}, "POST", orgCreateHREF, xmlData)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new Org: %s", err)
	}

	task := NewTask(&vcdClient.Client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
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
	orgHREF, err := url.ParseRequestURI(orgUrl)
	if err != nil {
		return Org{}, fmt.Errorf("error parsing org href: %v", err)
	}
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", *orgHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error retrieving org: %s", err)
	}

	org := NewOrg(&vcdClient.Client)
	if err = decodeBody(resp, org.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
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
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/org/")[1]
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", orgHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error retrieving org: %s", err)
	}
	org := NewAdminOrg(&vcdClient.Client)
	if err = decodeBody(resp, org.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *org, nil
}

// Returns the HREF of the org with the name orgName
func getOrgHREF(vcdClient *VCDClient, orgName string) (string, error) {
	orgListHREF := vcdClient.Client.VCDHREF
	orgListHREF.Path += "/org"
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", orgListHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("error retrieving org list: %s", err)
	}
	orgList := new(types.OrgList)
	if err = decodeBody(resp, orgList); err != nil {
		return "", fmt.Errorf("error decoding response: %s", err)
	}
	// Look for orgName within OrgList
	for _, org := range orgList.Org {
		if org.Name == orgName {
			return org.HREF, nil
		}
	}
	return "", fmt.Errorf("couldn't find org with name: %s", orgName)
}
