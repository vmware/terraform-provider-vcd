/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Simple structure to pass Edge Gateway creation parameters.
type EdgeGatewayCreation struct {
	ExternalNetworks           []string // List of external networks to be linked to this gateway
	DefaultGateway             string   // Which network should be used as default gateway (empty name = no default gateway)
	OrgName                    string   // parent Org
	VdcName                    string   // parent VDC
	Name                       string   // edge gateway name
	Description                string   // Optional description
	BackingConfiguration       string   // Type of backing configuration (compact, full)
	AdvancedNetworkingEnabled  bool     // enable advanced gateway
	HAEnabled                  bool     // enable HA
	UseDefaultRouteForDNSRelay bool     // True if the default gateway should be used as the DNS relay
	DistributedRoutingEnabled  bool     // If advanced networking enabled, also enable distributed routing
}

// Creates an Admin Organization based on settings, description, and org name.
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

	// There is a bug in the settings of CanPublishCatalogs.
	// If UseServerBootSequence is not set, CanPublishCatalogs is always false
	// regardless of the value passed during creation.
	if settings != nil {
		if settings.OrgGeneralSettings != nil {
			settings.OrgGeneralSettings.UseServerBootSequence = true
		}
	}
	orgCreateHREF := vcdClient.Client.VCDHREF
	orgCreateHREF.Path += "/admin/orgs"

	// Return the task
	return vcdClient.Client.ExecuteTaskRequest(orgCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.organization+xml", "error instantiating a new Org: %s", vcomp)

}

// Returns the UUID part of an entity ID
// From "urn:vcloud:vdc:72fefde7-4fed-45b8-a774-79b72c870325",
// will return "72fefde7-4fed-45b8-a774-79b72c870325"
// From "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0"
// will return "97384890-180c-4563-b9b7-0dc50a2430b0"
func getBareEntityUuid(entityId string) (string, error) {
	// Regular expression to match an ID:
	//     3 strings (alphanumeric + "-") separated by a colon (:)
	//     1 group of 8 hexadecimal digits
	//     3 groups of 4 hexadecimal digits
	//     1 group of 12 hexadecimal digits
	reGetID := regexp.MustCompile(`^[\w-]+:[\w-]+:[\w-]+:([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})$`)
	matchList := reGetID.FindAllStringSubmatch(entityId, -1)

	// matchList has the format
	// [][]string{[]string{"TOTAL MATCHED STRING", "CAPTURED TEXT"}}
	// such as
	// [][]string{[]string{"urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0", "97384890-180c-4563-b9b7-0dc50a2430b0"}}
	if len(matchList) == 0 || len(matchList[0]) < 2 {
		return "", fmt.Errorf("error extracting ID from '%s'", entityId)
	}
	return matchList[0][1], nil
}

// CreateEdgeGatewayAsync creates an edge gateway using a simplified configuration structure
// https://code.vmware.com/apis/442/vcloud-director/doc/doc/operations/POST-CreateEdgeGateway.html
//
// Note. This function does not allow to pick exact subnet in external network to use for edge
// gateway. It will pick first one instead.
func CreateEdgeGatewayAsync(vcdClient *VCDClient, egwc EdgeGatewayCreation) (Task, error) {

	distributed := egwc.DistributedRoutingEnabled
	if !egwc.AdvancedNetworkingEnabled {
		distributed = false
	}
	// This is the main configuration structure
	egwConfiguration := &types.EdgeGateway{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        egwc.Name,
		Description: egwc.Description,
		Configuration: &types.GatewayConfiguration{
			UseDefaultRouteForDNSRelay: &egwc.UseDefaultRouteForDNSRelay,
			HaEnabled:                  &egwc.HAEnabled,
			GatewayBackingConfig:       egwc.BackingConfiguration,
			AdvancedNetworkingEnabled:  &egwc.AdvancedNetworkingEnabled,
			DistributedRoutingEnabled:  &distributed,
			GatewayInterfaces: &types.GatewayInterfaces{
				GatewayInterface: []*types.GatewayInterface{},
			},
			EdgeGatewayServiceConfiguration: &types.GatewayFeatures{},
		},
	}

	if len(egwc.ExternalNetworks) == 0 {
		return Task{}, fmt.Errorf("no external networks provided. At least one is needed")
	}

	// If the user has indicated a default gateway, we make sure that it matches
	// a name in the list of external networks
	if egwc.DefaultGateway != "" {
		defaultGatewayFound := false
		for _, name := range egwc.ExternalNetworks {
			if egwc.DefaultGateway == name {
				defaultGatewayFound = true
			}
		}
		if !defaultGatewayFound {
			return Task{}, fmt.Errorf("default gateway (%s) selected, but its name is not among the external networks (%v)", egwc.DefaultGateway, egwc.ExternalNetworks)
		}
	}
	// Add external networks inside the configuration structure
	for _, extNetName := range egwc.ExternalNetworks {
		extNet, err := vcdClient.GetExternalNetworkByName(extNetName)
		if err != nil {
			return Task{}, err
		}

		// Populate the subnet participation only if default gateway was set
		var subnetParticipation *types.SubnetParticipation
		if egwc.DefaultGateway != "" && extNet.ExternalNetwork.Name == egwc.DefaultGateway {
			for _, net := range extNet.ExternalNetwork.Configuration.IPScopes.IPScope {
				if net.IsEnabled {
					subnetParticipation = &types.SubnetParticipation{
						Gateway: net.Gateway,
						Netmask: net.Netmask,
					}
					break
				}
			}
		}
		networkConf := &types.GatewayInterface{
			Name:          extNet.ExternalNetwork.Name,
			DisplayName:   extNet.ExternalNetwork.Name,
			InterfaceType: "uplink",
			Network: &types.Reference{
				HREF: extNet.ExternalNetwork.HREF,
				ID:   extNet.ExternalNetwork.ID,
				Type: "application/vnd.vmware.admin.network+xml",
				Name: extNet.ExternalNetwork.Name,
			},
			UseForDefaultRoute:  egwc.DefaultGateway == extNet.ExternalNetwork.Name,
			SubnetParticipation: []*types.SubnetParticipation{subnetParticipation},
		}

		egwConfiguration.Configuration.GatewayInterfaces.GatewayInterface =
			append(egwConfiguration.Configuration.GatewayInterfaces.GatewayInterface, networkConf)
	}

	// Once the configuration structure has been filled using the simplified data, we delegate
	// the edge gateway creation to the main configuration function.
	return CreateAndConfigureEdgeGatewayAsync(vcdClient, egwc.OrgName, egwc.VdcName, egwc.Name, egwConfiguration)
}

// CreateAndConfigureEdgeGatewayAsync creates an edge gateway using a full configuration structure
func CreateAndConfigureEdgeGatewayAsync(vcdClient *VCDClient, orgName, vdcName, egwName string, egwConfiguration *types.EdgeGateway) (Task, error) {

	if egwConfiguration.Name != egwName {
		return Task{}, fmt.Errorf("name mismatch: '%s' used as parameter but '%s' in the configuration structure", egwName, egwConfiguration.Name)
	}

	egwConfiguration.Xmlns = types.XMLNamespaceVCloud

	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return Task{}, err
	}
	vdc, err := adminOrg.GetVDCByName(vdcName, false)
	if err != nil {
		return Task{}, err
	}

	egwCreateHREF := vcdClient.Client.VCDHREF

	vdcId, err := getBareEntityUuid(vdc.Vdc.ID)
	if err != nil {
		return Task{}, fmt.Errorf("error retrieving ID from Vdc %s: %s", vdcName, err)
	}
	if vdcId == "" {
		return Task{}, fmt.Errorf("error retrieving ID from Vdc %s - empty ID returned", vdcName)
	}
	egwCreateHREF.Path += fmt.Sprintf("/admin/vdc/%s/edgeGateways", vdcId)

	// The first task is the creation task. It is quick, and does only create the vCD entity,
	// but not yet deploy the underlying VM
	creationTask, err := vcdClient.Client.ExecuteTaskRequest(egwCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGateway+xml", "error instantiating a new Edge Gateway: %s", egwConfiguration)

	if err != nil {
		return Task{}, err
	}

	err = creationTask.WaitTaskCompletion()

	if err != nil {
		return Task{}, err
	}

	// After creation, there is a build task that supervises the gateway deployment
	for _, innerTask := range creationTask.Task.Tasks.Task {
		if innerTask.OperationName == "networkEdgeGatewayCreate" {
			deployTask := Task{
				Task:   innerTask,
				client: &vcdClient.Client,
			}
			return deployTask, nil
		}
	}
	return Task{}, fmt.Errorf("no deployment task found for edge gateway %s - The edge gateway might have been created, but not deployed properly", egwName)
}

// Private convenience function used by CreateAndConfigureEdgeGateway and CreateEdgeGateway to
// process the task and return the object that was created.
// It should not be invoked directly.
func createEdgeGateway(vcdClient *VCDClient, egwc EdgeGatewayCreation, egwConfiguration *types.EdgeGateway) (EdgeGateway, error) {
	var task Task
	var err error
	if egwConfiguration != nil {
		task, err = CreateAndConfigureEdgeGatewayAsync(vcdClient, egwc.OrgName, egwc.VdcName, egwc.Name, egwConfiguration)
	} else {
		task, err = CreateEdgeGatewayAsync(vcdClient, egwc)
	}

	if err != nil {
		return EdgeGateway{}, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return EdgeGateway{}, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	// The edge gateway is created. Now we retrieve it from the server
	org, err := vcdClient.GetAdminOrgByName(egwc.OrgName)
	if err != nil {
		return EdgeGateway{}, err
	}
	vdc, err := org.GetVDCByName(egwc.VdcName, false)
	if err != nil {
		return EdgeGateway{}, err
	}
	egw, err := vdc.GetEdgeGatewayByName(egwc.Name, false)
	if err != nil {
		return EdgeGateway{}, err
	}
	return *egw, nil
}

// CreateAndConfigureEdgeGateway creates an edge gateway using a full configuration structure
func CreateAndConfigureEdgeGateway(vcdClient *VCDClient, orgName, vdcName, egwName string, egwConfiguration *types.EdgeGateway) (EdgeGateway, error) {
	return createEdgeGateway(vcdClient, EdgeGatewayCreation{OrgName: orgName, VdcName: vdcName, Name: egwName}, egwConfiguration)
}

// CreateEdgeGateway creates an edge gateway using a simplified configuration structure
func CreateEdgeGateway(vcdClient *VCDClient, egwc EdgeGatewayCreation) (EdgeGateway, error) {
	return createEdgeGateway(vcdClient, egwc, nil)
}

// If user specifies a valid organization name, then this returns a
// organization object. If no valid org is found, it returns an empty
// org and no error. Otherwise it returns an error and an empty
// Org object
// Deprecated: Use vcdClient.GetOrgByName instead
func GetOrgByName(vcdClient *VCDClient, orgName string) (Org, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgName)
	if err != nil {
		return Org{}, fmt.Errorf("organization '%s' fetch failed: %s", orgName, err)
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
// Deprecated: Use vcdClient.GetAdminOrgByName instead
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

// Returns the HREF of the org from the org ID
func getOrgHREFById(vcdClient *VCDClient, orgId string) (string, error) {
	orgListHREF := vcdClient.Client.VCDHREF
	orgListHREF.Path += "/org"

	orgList := new(types.OrgList)

	_, err := vcdClient.Client.ExecuteRequest(orgListHREF.String(), http.MethodGet,
		"", "error retrieving org list: %s", nil, orgList)
	if err != nil {
		return "", err
	}

	orgUuid, err := getBareEntityUuid(orgId)
	if err != nil {
		return "", err
	}
	// Look for org UUID within OrgList
	for _, org := range orgList.Org {
		// ID in orgList is usually empty. We extract the UUID from HREF to make the comparison
		uuidFromHref, err := GetUuidFromHref(org.HREF, true)
		if err != nil {
			return "", err
		}
		if uuidFromHref == orgUuid {
			return org.HREF, nil
		}
	}
	return "", fmt.Errorf("couldn't find org with ID: %s", orgId)
}

// Find a list of Virtual Centers matching the filter parameter.
// Filter constructing guide: https://pubs.vmware.com/vcloud-api-1-5/wwhelp/wwhimpl/js/html/wwhelp.htm#href=api_prog/GUID-CDF04296-5EB5-47E1-9BEC-228837C584CE.html
// Possible parameters are any attribute from QueryResultVirtualCenterRecordType struct
// E.g. filter could look like: name==vC1
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
	return QueryPortGroups(vcdCli, fmt.Sprintf("name==%s;portgroupType==%s", url.QueryEscape(name), "NETWORK"))
}

// Find a Distributed port group by name
func QueryDistributedPortGroup(vcdCli *VCDClient, name string) ([]*types.PortGroupRecordType, error) {
	return QueryPortGroups(vcdCli, fmt.Sprintf("name==%s;portgroupType==%s", url.QueryEscape(name), "DV_PORTGROUP"))
}

// Find a list of Port groups matching the filter parameter.
func QueryPortGroups(vcdCli *VCDClient, filter string) ([]*types.PortGroupRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "portgroup",
		"filter":        filter,
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.PortGroupRecord, nil
}

// GetExternalNetwork returns an ExternalNetwork reference if the network name matches an existing one.
// If no valid external network is found, it returns an empty ExternalNetwork reference and an error
// Deprecated: use vcdClient.GetExternalNetworkByName instead
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

	found := false
	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			err = externalNetwork.Refresh()
			found = true
			if err != nil {
				return &ExternalNetwork{}, err
			}
		}
	}

	if found {
		return externalNetwork, nil
	}
	return externalNetwork, fmt.Errorf("could not find external network named %s", networkName)

}

// GetExternalNetworks returns a list of available external networks
func (vcdClient *VCDClient) GetExternalNetworks() (*types.ExternalNetworkReferences, error) {

	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires system administrator privileges")
	}

	extNetworkHREF, err := getExternalNetworkHref(&vcdClient.Client)
	if err != nil {
		return nil, err
	}

	extNetworkRefs := &types.ExternalNetworkReferences{}
	_, err = vcdClient.Client.ExecuteRequest(extNetworkHREF, http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving external networks: %s", nil, extNetworkRefs)
	if err != nil {
		return nil, err
	}

	return extNetworkRefs, nil
}

// GetExternalNetworkByName returns an ExternalNetwork reference if the network name matches an existing one.
// If no valid external network is found, it returns a nil ExternalNetwork reference and an error
func (vcdClient *VCDClient) GetExternalNetworkByName(networkName string) (*ExternalNetwork, error) {

	extNetworkRefs, err := vcdClient.GetExternalNetworks()

	if err != nil {
		return nil, err
	}

	externalNetwork := NewExternalNetwork(&vcdClient.Client)

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		if netRef.Name == networkName {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			err = externalNetwork.Refresh()
			if err != nil {
				return nil, err
			}
			return externalNetwork, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetExternalNetworkById returns an ExternalNetwork reference if the network ID matches an existing one.
// If no valid external network is found, it returns a nil ExternalNetwork reference and an error
func (vcdClient *VCDClient) GetExternalNetworkById(id string) (*ExternalNetwork, error) {

	extNetworkRefs, err := vcdClient.GetExternalNetworks()

	if err != nil {
		return nil, err
	}

	externalNetwork := NewExternalNetwork(&vcdClient.Client)

	for _, netRef := range extNetworkRefs.ExternalNetworkReference {
		// ExternalNetworkReference items don't have ID
		// We compare using the UUID from HREF
		if equalIds(id, "", netRef.HREF) {
			externalNetwork.ExternalNetwork.HREF = netRef.HREF
			err = externalNetwork.Refresh()
			if err != nil {
				return nil, err
			}
			return externalNetwork, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetExternalNetworkByNameOrId returns an ExternalNetwork reference if either the network name or ID matches an existing one.
// If no valid external network is found, it returns a nil ExternalNetwork reference and an error
func (vcdClient *VCDClient) GetExternalNetworkByNameOrId(identifier string) (*ExternalNetwork, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vcdClient.GetExternalNetworkByName(name) }
	getById := func(id string, refresh bool) (interface{}, error) { return vcdClient.GetExternalNetworkById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*ExternalNetwork), err
}

// CreateExternalNetwork allows create external network and returns Task or error.
// types.ExternalNetwork struct is general and used for various types of networks. But for external network
// fence mode is always isolated, isInherited is false, parentNetwork is empty.
func CreateExternalNetwork(vcdClient *VCDClient, externalNetworkData *types.ExternalNetwork) (Task, error) {

	if !vcdClient.Client.IsSysAdmin {
		return Task{}, fmt.Errorf("functionality requires system administrator privileges")
	}

	err := validateExternalNetwork(externalNetworkData)
	if err != nil {
		return Task{}, err
	}

	// Type: VimObjectRefType
	// Namespace: http://www.vmware.com/vcloud/extension/v1.5
	// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VimObjectRefsType.html
	// Description: Represents the Managed Object Reference (MoRef) and the type of a vSphere object.
	// Since: 0.9
	type vimObjectRefCreate struct {
		VimServerRef  *types.Reference `xml:"vmext:VimServerRef"`
		MoRef         string           `xml:"vmext:MoRef"`
		VimObjectType string           `xml:"vmext:VimObjectType"`
	}

	// Type: VimObjectRefsType
	// Namespace: http://www.vmware.com/vcloud/extension/v1.5
	// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VimObjectRefsType.html
	// Description: List of VimObjectRef elements.
	// Since: 0.9
	type vimObjectRefsCreate struct {
		VimObjectRef []*vimObjectRefCreate `xml:"vmext:VimObjectRef"`
	}

	// Type: VMWExternalNetworkType
	// Namespace: http://www.vmware.com/vcloud/extension/v1.5
	// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VMWExternalNetworkType.html
	// Description: External network type.
	// Since: 1.0
	type externalNetworkCreate struct {
		XMLName          xml.Name                    `xml:"vmext:VMWExternalNetwork"`
		XmlnsVmext       string                      `xml:"xmlns:vmext,attr,omitempty"`
		XmlnsVcloud      string                      `xml:"xmlns:vcloud,attr,omitempty"`
		HREF             string                      `xml:"href,attr,omitempty"`
		Type             string                      `xml:"type,attr,omitempty"`
		ID               string                      `xml:"id,attr,omitempty"`
		OperationKey     string                      `xml:"operationKey,attr,omitempty"`
		Name             string                      `xml:"name,attr"`
		Link             []*types.Link               `xml:"Link,omitempty"`
		Description      string                      `xml:"vcloud:Description,omitempty"`
		Tasks            *types.TasksInProgress      `xml:"Tasks,omitempty"`
		Configuration    *types.NetworkConfiguration `xml:"vcloud:Configuration,omitempty"`
		VimPortGroupRef  *vimObjectRefCreate         `xml:"VimPortGroupRef,omitempty"`
		VimPortGroupRefs *vimObjectRefsCreate        `xml:"vmext:VimPortGroupRefs,omitempty"`
		VCloudExtension  *types.VCloudExtension      `xml:"VCloudExtension,omitempty"`
	}

	// Specific struct is used as two different name spaces needed for vCD API and return struct has diff name spaces
	externalNetwork := &externalNetworkCreate{}
	externalNetwork.HREF = externalNetworkData.HREF
	externalNetwork.Description = externalNetworkData.Description
	externalNetwork.Name = externalNetworkData.Name
	externalNetwork.Type = externalNetworkData.Type
	externalNetwork.ID = externalNetworkData.ID
	externalNetwork.OperationKey = externalNetworkData.OperationKey
	externalNetwork.Link = externalNetworkData.Link
	externalNetwork.Configuration = externalNetworkData.Configuration
	if externalNetwork.Configuration != nil {
		externalNetwork.Configuration.Xmlns = types.XMLNamespaceVCloud
	}
	externalNetwork.VCloudExtension = externalNetworkData.VCloudExtension
	externalNetwork.XmlnsVmext = types.XMLNamespaceExtension
	externalNetwork.XmlnsVcloud = types.XMLNamespaceVCloud
	externalNetwork.Type = types.MimeExternalNetwork
	if externalNetworkData.VimPortGroupRefs != nil {
		externalNetwork.VimPortGroupRefs = &vimObjectRefsCreate{}
		for _, vimObjRef := range externalNetworkData.VimPortGroupRefs.VimObjectRef {
			externalNetwork.VimPortGroupRefs.VimObjectRef = append(externalNetwork.VimPortGroupRefs.VimObjectRef, &vimObjectRefCreate{
				VimServerRef:  vimObjRef.VimServerRef,
				MoRef:         vimObjRef.MoRef,
				VimObjectType: vimObjRef.VimObjectType,
			})
		}
	}
	if externalNetworkData.VimPortGroupRef != nil {
		externalNetwork.VimPortGroupRef = &vimObjectRefCreate{
			VimServerRef:  externalNetworkData.VimPortGroupRef.VimServerRef,
			MoRef:         externalNetworkData.VimPortGroupRef.MoRef,
			VimObjectType: externalNetworkData.VimPortGroupRef.VimObjectType,
		}
	}

	externalNetHREF := vcdClient.Client.VCDHREF
	externalNetHREF.Path += "/admin/extension/externalnets"

	externalNetwork.Configuration.FenceMode = "isolated"

	// Return the task
	task, err := vcdClient.Client.ExecuteTaskRequest(externalNetHREF.String(), http.MethodPost,
		types.MimeExternalNetwork, "error instantiating a new ExternalNetwork: %s", externalNetwork)

	// Real task in task array
	if err == nil {
		if task.Task != nil && task.Task.Tasks != nil && len(task.Task.Tasks.Task) == 0 {
			return Task{}, fmt.Errorf("create external network task wasn't found")
		}
		task.Task = task.Task.Tasks.Task[0]
	}

	return task, err
}

func getExtension(client *Client) (*types.Extension, error) {
	extensions := &types.Extension{}

	extensionHREF := client.VCDHREF
	extensionHREF.Path += "/admin/extension/"

	_, err := client.ExecuteRequest(extensionHREF.String(), http.MethodGet,
		"", "error retrieving extension: %s", nil, extensions)

	return extensions, err
}

// GetStorageProfileByHref fetches storage profile using provided HREF.
func GetStorageProfileByHref(vcdClient *VCDClient, url string) (*types.VdcStorageProfile, error) {

	vdcStorageProfile := &types.VdcStorageProfile{}

	_, err := vcdClient.Client.ExecuteRequest(url, http.MethodGet,
		"", "error retrieving storage profile: %s", nil, vdcStorageProfile)
	if err != nil {
		return nil, err
	}

	return vdcStorageProfile, nil
}

// QueryProviderVdcStorageProfileByName finds a provider VDC storage profile by name
func QueryProviderVdcStorageProfileByName(vcdCli *VCDClient, name string) ([]*types.QueryResultProviderVdcStorageProfileRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "providerVdcStorageProfile",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.ProviderVdcStorageProfileRecord, nil
}

// QueryNetworkPoolByName finds a network pool by name
func QueryNetworkPoolByName(vcdCli *VCDClient, name string) ([]*types.QueryResultNetworkPoolRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "networkPool",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.NetworkPoolRecord, nil
}

// QueryProviderVdcByName finds a provider VDC by name
func QueryProviderVdcByName(vcdCli *VCDClient, name string) ([]*types.QueryResultVMWProviderVdcRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "providerVdc",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.VMWProviderVdcRecord, nil
}

// QueryProviderVdcs gets the list of available provider VDCs
func (vcdClient *VCDClient) QueryProviderVdcs() ([]*types.QueryResultVMWProviderVdcRecordType, error) {
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type": "providerVdc",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.VMWProviderVdcRecord, nil
}

// QueryNetworkPools gets the list of network pools
func (vcdClient *VCDClient) QueryNetworkPools() ([]*types.QueryResultNetworkPoolRecordType, error) {
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type": "networkPool",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.NetworkPoolRecord, nil
}

// QueryProviderVdcStorageProfiles gets the list of provider VDC storage profiles
func (vcdClient *VCDClient) QueryProviderVdcStorageProfiles() ([]*types.QueryResultProviderVdcStorageProfileRecordType, error) {
	results, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type": "providerVdcStorageProfile",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.ProviderVdcStorageProfileRecord, nil
}

// GetNetworkPoolByHREF functions fetches an network pool using VDC client and network pool href
func GetNetworkPoolByHREF(client *VCDClient, href string) (*types.VMWNetworkPool, error) {
	util.Logger.Printf("[TRACE] Get network pool by HREF: %s\n", href)

	networkPool := &types.VMWNetworkPool{}

	_, err := client.Client.ExecuteRequest(href, http.MethodGet,
		"", "error fetching network pool: %s", nil, networkPool)

	// Return the disk
	return networkPool, err

}

// QueryOrgVdcNetworkByName finds a org VDC network by name which has edge gateway as reference
func QueryOrgVdcNetworkByName(vcdCli *VCDClient, name string) ([]*types.QueryResultOrgVdcNetworkRecordType, error) {
	results, err := vcdCli.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "orgVdcNetwork",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	return results.Results.OrgVdcNetworkRecord, nil
}

// GetOrgByName finds an Organization by name
// On success, returns a pointer to the Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetOrgByName(orgName string) (*Org, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgName)
	if err != nil {
		// Since this operation is a lookup from a list, we return the standard ErrorEntityNotFound
		return nil, ErrorEntityNotFound
	}
	org := NewOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgUrl, http.MethodGet,
		"", "error retrieving org: %s", nil, org.Org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// GetOrgById finds an Organization by ID
// On success, returns a pointer to the Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetOrgById(orgId string) (*Org, error) {
	orgUrl, err := getOrgHREFById(vcdClient, orgId)
	if err != nil {
		// Since this operation is a lookup from a list, we return the standard ErrorEntityNotFound
		return nil, ErrorEntityNotFound
	}
	org := NewOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgUrl, http.MethodGet,
		"", "error retrieving org list: %s", nil, org.Org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// GetOrgByNameOrId finds an Organization by name or ID
// On success, returns a pointer to the Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetOrgByNameOrId(identifier string) (*Org, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vcdClient.GetOrgByName(name) }
	getById := func(id string, refresh bool) (interface{}, error) { return vcdClient.GetOrgById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*Org), err
}

// GetAdminOrgByName finds an Admin Organization by name
// On success, returns a pointer to the Admin Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetAdminOrgByName(orgName string) (*AdminOrg, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgName)
	if err != nil {
		return nil, ErrorEntityNotFound
	}
	orgHREF := vcdClient.Client.VCDHREF
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/api/org/")[1]

	adminOrg := NewAdminOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgHREF.String(), http.MethodGet,
		"", "error retrieving org: %s", nil, adminOrg.AdminOrg)
	if err != nil {
		return nil, err
	}

	return adminOrg, nil
}

// GetAdminOrgById finds an Admin Organization by ID
// On success, returns a pointer to the Admin Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetAdminOrgById(orgId string) (*AdminOrg, error) {
	orgUrl, err := getOrgHREFById(vcdClient, orgId)
	if err != nil {
		return nil, ErrorEntityNotFound
	}
	orgHREF := vcdClient.Client.VCDHREF
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/api/org/")[1]

	adminOrg := NewAdminOrg(&vcdClient.Client)

	_, err = vcdClient.Client.ExecuteRequest(orgHREF.String(), http.MethodGet,
		"", "error retrieving org: %s", nil, adminOrg.AdminOrg)
	if err != nil {
		return nil, err
	}

	return adminOrg, nil
}

// GetAdminOrgByNameOrId finds an Admin Organization by name or ID
// On success, returns a pointer to the Admin Org structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetAdminOrgByNameOrId(identifier string) (*AdminOrg, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vcdClient.GetAdminOrgByName(name) }
	getById := func(id string, refresh bool) (interface{}, error) { return vcdClient.GetAdminOrgById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*AdminOrg), err
}

// Returns the UUID part of an HREF
// Similar to getBareEntityUuid, but tailored to HREF
func GetUuidFromHref(href string, idAtEnd bool) (string, error) {
	util.Logger.Printf("[TRACE] GetUuidFromHref got href: %s with idAtEnd: %t", href, idAtEnd)
	// Regular expression to match an ID:
	//     1 string starting by 'https://' and ending with a '/',
	//     followed by
	//        1 group of 8 hexadecimal digits
	//        3 groups of 4 hexadecimal digits
	//        1 group of 12 hexadecimal digits

	searchExpression := `^https://.+/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`
	if idAtEnd {
		searchExpression += `$`
	} else {
		searchExpression += `.*$`
	}
	reGetID := regexp.MustCompile(searchExpression)
	matchList := reGetID.FindAllStringSubmatch(href, -1)

	if len(matchList) == 0 || len(matchList[0]) < 2 {
		return "", fmt.Errorf("error extracting UUID from '%s'", href)
	}
	util.Logger.Printf("[TRACE] GetUuidFromHref returns UUID : %s", matchList[0][1])
	return matchList[0][1], nil
}
