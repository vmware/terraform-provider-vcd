/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
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
		Xmlns:       types.XMLNamespaceVCloud,
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

// Find a list of Virtual Centers matching the filter parameter.
// Filter constructing guide: https://pubs.vmware.com/vcloud-api-1-5/wwhelp/wwhimpl/js/html/wwhelp.htm#href=api_prog/GUID-CDF04296-5EB5-47E1-9BEC-228837C584CE.html
// Possible parameters are any attribute from QueryResultVirtualCenterRecordType struct
// E.g. filter could look like: (name==vC1)
func QueryVirtualCenters(vcdClient *VCDClient, filter string) ([]*types.QueryResultVirtualCenterRecordType, error) {
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "virtualCenter",
		"filter": filter,
	})
	if err != nil {
		return nil, err
	}

	return results.Results.VirtualCenterRecord, nil
}

// Find a Network port group by name
func QueryNetworkPortGroup(vcdCli *VCDClient, name string) ([]*types.PortGroupRecordType, error) {
	return QueryPortGroups(vcdCli, fmt.Sprintf("(name==%s;portgroupType==%s)", url.QueryEscape(name), "NETWORK"))
}

// Find a Distributed port group by name
func QueryDistributedPortGroup(vcdCli *VCDClient, name string) ([]*types.PortGroupRecordType, error) {
	return QueryPortGroups(vcdCli, fmt.Sprintf("(name==%s;portgroupType==%s)", url.QueryEscape(name), "DV_PORTGROUP"))
}

// Find a list of Port groups matching the filter parameter.
func QueryPortGroups(vcdCli *VCDClient, filter string) ([]*types.PortGroupRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "portgroup",
		"filter": filter,
	})
	if err != nil {
		return nil, err
	}

	return results.Results.PortGroupRecord, nil
}

// GetExternalNetwork returns ExternalNetwork object if user specifies a valid external network name.
// If no valid external network is found, it returns an empty
// ExternalNetwork and no error. Otherwise it returns an error and an empty
// ExternalNetwork object
func GetExternalNetwork(vcdClient *VCDClient, networkName string) (*ExternalNetwork, error) {

	if !vcdClient.Client.IsSysAdmin {
		return &ExternalNetwork{}, fmt.Errorf("functionality requires system administrator privileges")
	}

	extNetworkHREF, err := getExternalNetworkHref(&vcdClient.Client)
	if err != nil {
		return &ExternalNetwork{}, err
	}

	extNetworkRefs := &types.ExternalNetworkReferences{}
	_, err = vcdClient.Client.ExecuteRequest(extNetworkHREF, http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving external networks: %s", nil, extNetworkRefs)
	if err != nil {
		return &ExternalNetwork{}, err
	}

	externalNetwork := NewExternalNetwork(&vcdClient.Client)

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			err = externalNetwork.Refresh()
			if err != nil {
				return &ExternalNetwork{}, err
			}
		}
	}

	return externalNetwork, nil

}

// CreateExternalNetwork allows create external network and returns Task or error.
// types.ExternalNetwork struct is general and used for various types of networks. But for external network
// fence mode is always isolated, isInherited is false, parentNetwork is empty.
func CreateExternalNetwork(vcdClient *VCDClient, externalNetwork *types.ExternalNetwork) (Task, error) {

	if !vcdClient.Client.IsSysAdmin {
		return Task{}, fmt.Errorf("functionality requires system administrator privileges")
	}

	err := validateExternalNetwork(externalNetwork)
	if err != nil {
		return Task{}, err
	}

	externalNetHREF := vcdClient.Client.VCDHREF
	externalNetHREF.Path += "/admin/extension/externalnets"

	externalNetwork.Configuration.FenceMode = "isolated"

	// Return the task
	return vcdClient.Client.ExecuteTaskRequest(externalNetHREF.String(), http.MethodPost,
		types.MimeExternalNetwork, "error instantiating a new ExternalNetwork: %s", externalNetwork)
}

func getExtension(client *Client) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := client.VCDHREF
	extensionHREF.Path += "/admin/extension/"

	_, err := client.ExecuteRequest(extensionHREF.String(), http.MethodGet,
		"", "error retrieving extension: %s", nil, extensions)

	return extensions, err
}
