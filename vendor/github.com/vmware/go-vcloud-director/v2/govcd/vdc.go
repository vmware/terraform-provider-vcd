/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type Vdc struct {
	Vdc    *types.Vdc
	client *Client
	VApp   *types.VApp
}

func NewVdc(cli *Client) *Vdc {
	return &Vdc{
		Vdc:    new(types.Vdc),
		client: cli,
	}
}

type AdminVdc struct {
	AdminVdc *types.AdminVdc
	client   *Client
	VApp     *types.VApp
}

func NewAdminVdc(cli *Client) *AdminVdc {
	return &AdminVdc{
		AdminVdc: new(types.AdminVdc),
		client:   cli,
	}
}

// Gets a vapp with a specific url vappHREF
func (vdc *Vdc) getVdcVAppbyHREF(vappHREF *url.URL) (*VApp, error) {
	vapp := NewVApp(vdc.client)

	_, err := vdc.client.ExecuteRequest(vappHREF.String(), http.MethodGet,
		"", "error retrieving VApp: %s", nil, vapp.VApp)

	return vapp, err
}

// Undeploys every vapp in the vdc
func (vdc *Vdc) undeployAllVdcVApps() error {
	err := vdc.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, resents := range vdc.Vdc.ResourceEntities {
		for _, resent := range resents.ResourceEntity {
			if resent.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vappHREF, err := url.Parse(resent.HREF)
				if err != nil {
					return err
				}
				vapp, err := vdc.getVdcVAppbyHREF(vappHREF)
				if err != nil {
					return fmt.Errorf("error retrieving vapp with url: %s and with error %s", vappHREF.Path, err)
				}
				task, err := vapp.Undeploy()
				if err != nil {
					return err
				}
				if task == (Task{}) {
					continue
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Removes all vapps in the vdc
func (vdc *Vdc) removeAllVdcVApps() error {
	err := vdc.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, resents := range vdc.Vdc.ResourceEntities {
		for _, resent := range resents.ResourceEntity {
			if resent.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vappHREF, err := url.Parse(resent.HREF)
				if err != nil {
					return err
				}
				vapp, err := vdc.getVdcVAppbyHREF(vappHREF)
				if err != nil {
					return fmt.Errorf("error retrieving vapp with url: %s and with error %s", vappHREF.Path, err)
				}
				task, err := vapp.Delete()
				if err != nil {
					return fmt.Errorf("error deleting vapp: %s", err)
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("couldn't finish removing vapp %#v", err)
				}
			}
		}
	}
	return nil
}

func (vdc *Vdc) Refresh() error {

	if vdc.Vdc.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledVdc := &types.Vdc{}

	_, err := vdc.client.ExecuteRequest(vdc.Vdc.HREF, http.MethodGet,
		"", "error refreshing vDC: %s", nil, unmarshalledVdc)
	if err != nil {
		return err
	}

	vdc.Vdc = unmarshalledVdc

	// The request was successful
	return nil
}

// Deletes the vdc, returning an error of the vCD call fails.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Vdc.html
func (vdc *Vdc) Delete(force bool, recursive bool) (Task, error) {
	util.Logger.Printf("[TRACE] Vdc.Delete - deleting VDC with force: %t, recursive: %t", force, recursive)

	if vdc.Vdc.HREF == "" {
		return Task{}, fmt.Errorf("cannot delete, Object is empty")
	}

	vdcUrl, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing vdc url: %s", err)
	}

	req := vdc.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, http.MethodDelete, *vdcUrl, nil)
	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting vdc: %s", err)
	}
	task := NewTask(vdc.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	if task.Task.Status == "error" {
		return Task{}, fmt.Errorf("vdc not properly destroyed")
	}
	return *task, nil
}

// Deletes the vdc and waits for the asynchronous task to complete.
func (vdc *Vdc) DeleteWait(force bool, recursive bool) error {
	task, err := vdc.Delete(force, recursive)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("couldn't finish removing vdc %#v", err)
	}
	return nil
}

func (vdc *Vdc) FindVDCNetwork(network string) (OrgVDCNetwork, error) {

	err := vdc.Refresh()
	if err != nil {
		return OrgVDCNetwork{}, fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, an := range vdc.Vdc.AvailableNetworks {
		for _, reference := range an.Network {
			if reference.Name == network {
				orgNet := NewOrgVDCNetwork(vdc.client)

				_, err := vdc.client.ExecuteRequest(reference.HREF, http.MethodGet,
					"", "error retrieving org vdc network: %s", nil, orgNet.OrgVDCNetwork)

				// The request was successful
				return *orgNet, err

			}
		}
	}

	return OrgVDCNetwork{}, fmt.Errorf("can't find VDC Network: %s", network)
}

func (vdc *Vdc) FindStorageProfileReference(name string) (types.Reference, error) {

	err := vdc.Refresh()
	if err != nil {
		return types.Reference{}, fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, sps := range vdc.Vdc.VdcStorageProfiles {
		for _, sp := range sps.VdcStorageProfile {
			if sp.Name == name {
				return types.Reference{HREF: sp.HREF, Name: sp.Name}, nil
			}
		}
	}
	return types.Reference{}, fmt.Errorf("can't find any VDC Storage_profiles")
}

func (vdc *Vdc) GetDefaultStorageProfileReference(storageprofiles *types.QueryResultRecordsType) (types.Reference, error) {

	err := vdc.Refresh()
	if err != nil {
		return types.Reference{}, fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, spr := range storageprofiles.OrgVdcStorageProfileRecord {
		if spr.IsDefaultStorageProfile {
			return types.Reference{HREF: spr.HREF, Name: spr.Name}, nil
		}
	}
	return types.Reference{}, fmt.Errorf("can't find Default VDC Storage_profile")
}

func (vdc *Vdc) FindEdgeGateway(edgegateway string) (EdgeGateway, error) {

	err := vdc.Refresh()
	if err != nil {
		return EdgeGateway{}, fmt.Errorf("error refreshing vdc: %s", err)
	}
	for _, av := range vdc.Vdc.Link {
		if av.Rel == "edgeGateways" && av.Type == "application/vnd.vmware.vcloud.query.records+xml" {

			query := new(types.QueryResultEdgeGatewayRecordsType)

			_, err := vdc.client.ExecuteRequest(av.HREF, http.MethodGet,
				"", "error quering edge gateways: %s", nil, query)
			if err != nil {
				return EdgeGateway{}, err
			}

			var href string

			for _, edge := range query.EdgeGatewayRecord {
				if edge.Name == edgegateway {
					href = edge.HREF
				}
			}

			if href == "" {
				return EdgeGateway{}, fmt.Errorf("can't find edge gateway with name: %s", edgegateway)
			}

			edge := NewEdgeGateway(vdc.client)

			_, err = vdc.client.ExecuteRequest(href, http.MethodGet,
				"", "error retrieving edge gateway: %s", nil, edge.EdgeGateway)

			return *edge, err

		}
	}
	return EdgeGateway{}, fmt.Errorf("can't find Edge Gateway")

}

func (vdc *Vdc) ComposeRawVApp(name string) error {
	vcomp := &types.ComposeVAppParams{
		Ovf:     types.XMLNamespaceOVF,
		Xsi:     types.XMLNamespaceXSI,
		Xmlns:   types.XMLNamespaceVCloud,
		Deploy:  false,
		Name:    name,
		PowerOn: false,
	}

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/composeVApp"

	task, err := vdc.client.ExecuteTaskRequest(vdcHref.String(), http.MethodPost,
		types.MimeComposeVappParams, "error instantiating a new vApp:: %s", vcomp)
	if err != nil {
		return fmt.Errorf("error executing task request: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error performing task: %#v", err)
	}

	return nil
}

// ComposeVApp creates a vapp with the given template, name, and description
// that uses the storageprofile and networks given. If you want all eulas
// to be accepted set acceptalleulas to true. Returns a successful task
// if completed successfully, otherwise returns an error and an empty task.
func (vdc *Vdc) ComposeVApp(orgvdcnetworks []*types.OrgVDCNetwork, vapptemplate VAppTemplate, storageprofileref types.Reference, name string, description string, acceptalleulas bool) (Task, error) {
	if vapptemplate.VAppTemplate.Children == nil || orgvdcnetworks == nil {
		return Task{}, fmt.Errorf("can't compose a new vApp, objects passed are not valid")
	}

	// Determine primary network connection index number. We normally depend on it being inherited from vApp template
	// but in the case when vApp template does not have network card it would fail on the index being undefined. We
	// set the value to 0 (first NIC instead)
	primaryNetworkConnectionIndex := 0
	if vapptemplate.VAppTemplate.Children != nil && len(vapptemplate.VAppTemplate.Children.VM) > 0 &&
		vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection != nil {
		primaryNetworkConnectionIndex = vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.PrimaryNetworkConnectionIndex
	}

	// Build request XML
	vcomp := &types.ComposeVAppParams{
		Ovf:         types.XMLNamespaceOVF,
		Xsi:         types.XMLNamespaceXSI,
		Xmlns:       types.XMLNamespaceVCloud,
		Deploy:      false,
		Name:        name,
		PowerOn:     false,
		Description: description,
		InstantiationParams: &types.InstantiationParams{
			NetworkConfigSection: &types.NetworkConfigSection{
				Info: "Configuration parameters for logical networks",
			},
		},
		AllEULAsAccepted: acceptalleulas,
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
				Name: vapptemplate.VAppTemplate.Children.VM[0].Name,
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{
					Info:                          "Network config for sourced item",
					PrimaryNetworkConnectionIndex: primaryNetworkConnectionIndex,
				},
			},
		},
	}
	for index, orgvdcnetwork := range orgvdcnetworks {
		vcomp.InstantiationParams.NetworkConfigSection.NetworkConfig = append(vcomp.InstantiationParams.NetworkConfigSection.NetworkConfig,
			types.VAppNetworkConfiguration{
				NetworkName: orgvdcnetwork.Name,
				Configuration: &types.NetworkConfiguration{
					FenceMode: types.FenceModeBridged,
					ParentNetwork: &types.Reference{
						HREF: orgvdcnetwork.HREF,
						Name: orgvdcnetwork.Name,
						Type: orgvdcnetwork.Type,
					},
				},
			},
		)
		vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection = append(vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection,
			&types.NetworkConnection{
				Network:                 orgvdcnetwork.Name,
				NetworkConnectionIndex:  index,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
			},
		)
		vcomp.SourcedItem.NetworkAssignment = append(vcomp.SourcedItem.NetworkAssignment,
			&types.NetworkAssignment{
				InnerNetwork:     orgvdcnetwork.Name,
				ContainerNetwork: orgvdcnetwork.Name,
			},
		)
	}
	if storageprofileref.HREF != "" {
		vcomp.SourcedItem.StorageProfile = &storageprofileref
	}

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/composeVApp"

	return vdc.client.ExecuteTaskRequest(vdcHref.String(), http.MethodPost,
		types.MimeComposeVappParams, "error instantiating a new vApp: %s", vcomp)
}

func (vdc *Vdc) FindVAppByName(vapp string) (VApp, error) {

	err := vdc.Refresh()
	if err != nil {
		return VApp{}, fmt.Errorf("error refreshing vdc: %s", err)
	}

	for _, resents := range vdc.Vdc.ResourceEntities {
		for _, resent := range resents.ResourceEntity {

			if resent.Name == vapp && resent.Type == "application/vnd.vmware.vcloud.vApp+xml" {

				newVapp := NewVApp(vdc.client)

				_, err := vdc.client.ExecuteRequest(resent.HREF, http.MethodGet,
					"", "error retrieving vApp: %s", nil, newVapp.VApp)

				return *newVapp, err

			}
		}
	}
	return VApp{}, fmt.Errorf("can't find vApp: %s", vapp)
}

func (vdc *Vdc) FindVMByName(vapp VApp, vm string) (VM, error) {

	err := vdc.Refresh()
	if err != nil {
		return VM{}, fmt.Errorf("error refreshing vdc: %s", err)
	}

	err = vapp.Refresh()
	if err != nil {
		return VM{}, fmt.Errorf("error refreshing vapp: %s", err)
	}

	//vApp Might Not Have Any VMs

	if vapp.VApp.Children == nil {
		return VM{}, fmt.Errorf("VApp Has No VMs")
	}

	util.Logger.Printf("[TRACE] Looking for VM: %s", vm)
	for _, child := range vapp.VApp.Children.VM {

		util.Logger.Printf("[TRACE] Found: %s", child.Name)
		if child.Name == vm {

			newVm := NewVM(vdc.client)

			_, err := vdc.client.ExecuteRequest(child.HREF, http.MethodGet,
				"", "error retrieving vm: %s", nil, newVm.VM)

			return *newVm, err
		}

	}
	util.Logger.Printf("[TRACE] Couldn't find VM: %s", vm)
	return VM{}, fmt.Errorf("can't find vm: %s", vm)
}

// Find vm using vApp name and VM name. Returns VMRecord query return type
func (vdc *Vdc) QueryVM(vappName, vmName string) (VMRecord, error) {

	if vmName == "" {
		return VMRecord{}, errors.New("error querying vm name is empty")
	}

	if vappName == "" {
		return VMRecord{}, errors.New("error querying vapp name is empty")
	}

	typeMedia := "vm"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminVM"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia,
		"filter": "(name==" + url.QueryEscape(vmName) + ";containerName==" + url.QueryEscape(vappName) + ")"})
	if err != nil {
		return VMRecord{}, fmt.Errorf("error querying vm %#v", err)
	}

	vmResults := results.Results.VMRecord
	if vdc.client.IsSysAdmin {
		vmResults = results.Results.AdminVMRecord
	}

	newVM := NewVMRecord(vdc.client)

	if len(vmResults) == 1 {
		newVM.VM = vmResults[0]
	} else {
		return VMRecord{}, fmt.Errorf("found results %d", len(vmResults))
	}

	return *newVM, nil
}

func (vdc *Vdc) FindVAppByID(vappid string) (VApp, error) {

	// Horrible hack to fetch a vapp with its id.
	// urn:vcloud:vapp:00000000-0000-0000-0000-000000000000

	err := vdc.Refresh()
	if err != nil {
		return VApp{}, fmt.Errorf("error refreshing vdc: %s", err)
	}

	urnslice := strings.SplitAfter(vappid, ":")
	urnid := urnslice[len(urnslice)-1]

	for _, resents := range vdc.Vdc.ResourceEntities {
		for _, resent := range resents.ResourceEntity {

			hrefslice := strings.SplitAfter(resent.HREF, "/")
			hrefslice = strings.SplitAfter(hrefslice[len(hrefslice)-1], "-")
			res := strings.Join(hrefslice[1:], "")

			if res == urnid && resent.Type == "application/vnd.vmware.vcloud.vApp+xml" {

				newVapp := NewVApp(vdc.client)

				_, err := vdc.client.ExecuteRequest(resent.HREF, http.MethodGet,
					"", "error retrieving vApp: %s", nil, newVapp.VApp)

				return *newVapp, err

			}
		}
	}
	return VApp{}, fmt.Errorf("can't find vApp")

}

func (vdc *Vdc) FindMediaImage(mediaName string) (MediaItem, error) {
	util.Logger.Printf("[TRACE] Querying medias by name\n")

	mediaResults, err := queryMediaItemsWithFilter(vdc, "name=="+url.QueryEscape(mediaName))
	if err != nil {
		return MediaItem{}, err
	}

	newMediaItem := NewMediaItem(vdc.client)

	if len(mediaResults) == 1 {
		newMediaItem.MediaItem = mediaResults[0]
	}

	if len(mediaResults) == 0 {
		return MediaItem{}, nil
	}

	if len(mediaResults) > 1 {
		return MediaItem{}, errors.New("found more than result")
	}

	util.Logger.Printf("[TRACE] Found media record by name: %#v \n", mediaResults)
	return *newMediaItem, nil
}
