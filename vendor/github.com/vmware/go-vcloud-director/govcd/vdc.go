/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
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
	req := vdc.client.NewRequest(map[string]string{}, "GET", *vappHREF, nil)
	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return &VApp{}, fmt.Errorf("error retrieving VApp: %s", err)
	}

	vapp := NewVApp(vdc.client)

	if err = decodeBody(resp, vapp.VApp); err != nil {
		return &VApp{}, fmt.Errorf("error decoding VApp response: %s", err)
	}
	return vapp, nil
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
					return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", vappHREF.Path, err)
				}
				task, err := vapp.Undeploy()
				if task == (Task{}) {
					continue
				}
				err = task.WaitTaskCompletion()
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

	refreshUrl, _ := url.ParseRequestURI(vdc.Vdc.HREF)

	req := vdc.client.NewRequest(map[string]string{}, "GET", *refreshUrl, nil)

	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving Edge Gateway: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledVdc := &types.Vdc{}

	if err = decodeBody(resp, unmarshalledVdc); err != nil {
		return fmt.Errorf("error decoding vdc response: %s", err)
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
	}, "DELETE", *vdcUrl, nil)
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
				findUrl, err := url.ParseRequestURI(reference.HREF)
				if err != nil {
					return OrgVDCNetwork{}, fmt.Errorf("error decoding vdc response: %s", err)
				}

				req := vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

				resp, err := checkResp(vdc.client.Http.Do(req))
				if err != nil {
					return OrgVDCNetwork{}, fmt.Errorf("error retrieving orgvdcnetwork: %s", err)
				}

				orgnet := NewOrgVDCNetwork(vdc.client)

				if err = decodeBody(resp, orgnet.OrgVDCNetwork); err != nil {
					return OrgVDCNetwork{}, fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
				}

				// The request was successful
				return *orgnet, nil

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
		return types.Reference{}, fmt.Errorf("can't find VDC Storage_profile: %s", name)
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
			findUrl, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return EdgeGateway{}, fmt.Errorf("error decoding vdc response: %s", err)
			}

			// Querying the Result list
			req := vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

			resp, err := checkResp(vdc.client.Http.Do(req))
			if err != nil {
				return EdgeGateway{}, fmt.Errorf("error retrieving edge gateway records: %s", err)
			}

			query := new(types.QueryResultEdgeGatewayRecordsType)

			if err = decodeBody(resp, query); err != nil {
				return EdgeGateway{}, fmt.Errorf("error decoding edge gateway query response: %s", err)
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

			findUrl, err = url.ParseRequestURI(href)
			if err != nil {
				return EdgeGateway{}, fmt.Errorf("error decoding edge gateway query response: %s", err)
			}

			// Querying the Result list
			req = vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

			resp, err = checkResp(vdc.client.Http.Do(req))
			if err != nil {
				return EdgeGateway{}, fmt.Errorf("error retrieving edge gateway: %s", err)
			}

			edge := NewEdgeGateway(vdc.client)

			if err = decodeBody(resp, edge.EdgeGateway); err != nil {
				return EdgeGateway{}, fmt.Errorf("error decoding edge gateway response: %s", err)
			}

			return *edge, nil

		}
	}
	return EdgeGateway{}, fmt.Errorf("can't find Edge Gateway")

}

func (vdc *Vdc) ComposeRawVApp(name string) error {
	vcomp := &types.ComposeVAppParams{
		Ovf:     "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:     "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:   "http://www.vmware.com/vcloud/v1.5",
		Deploy:  false,
		Name:    name,
		PowerOn: false,
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	requestData := bytes.NewBufferString(xml.Header + string(output))

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/composeVApp"

	req := vdc.client.NewRequest(map[string]string{}, "POST", *vdcHref, requestData)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.composeVAppParams+xml")

	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error instantiating a new vApp: %s", err)
	}

	task := NewTask(vdc.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("Error performing task: %#v", err)
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
	// Build request XML
	vcomp := &types.ComposeVAppParams{
		Ovf:         "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:         "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
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
					Type: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.Type,
					HREF: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.HREF,
					Info: "Network config for sourced item",
					PrimaryNetworkConnectionIndex: vapptemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.PrimaryNetworkConnectionIndex,
				},
			},
		},
	}
	for index, orgvdcnetwork := range orgvdcnetworks {
		vcomp.InstantiationParams.NetworkConfigSection.NetworkConfig = append(vcomp.InstantiationParams.NetworkConfigSection.NetworkConfig,
			types.VAppNetworkConfiguration{
				NetworkName: orgvdcnetwork.Name,
				Configuration: &types.NetworkConfiguration{
					FenceMode: "bridged",
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
				IPAddressAllocationMode: "POOL",
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

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}
	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	requestData := bytes.NewBufferString(xml.Header + string(output))

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/composeVApp"

	req := vdc.client.NewRequest(map[string]string{}, "POST", *vdcHref, requestData)
	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.composeVAppParams+xml")
	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new vApp: %s", err)
	}

	vapp := NewVApp(vdc.client)
	if err = decodeBody(resp, vapp.VApp); err != nil {
		return Task{}, fmt.Errorf("error decoding vApp response: %s", err)
	}

	task := NewTask(vdc.client)
	task.Task = vapp.VApp.Tasks.Task[0]
	// The request was successful
	return *task, nil
}

func (vdc *Vdc) FindVAppByName(vapp string) (VApp, error) {

	err := vdc.Refresh()
	if err != nil {
		return VApp{}, fmt.Errorf("error refreshing vdc: %s", err)
	}

	for _, resents := range vdc.Vdc.ResourceEntities {
		for _, resent := range resents.ResourceEntity {

			if resent.Name == vapp && resent.Type == "application/vnd.vmware.vcloud.vApp+xml" {

				findUrl, err := url.ParseRequestURI(resent.HREF)

				if err != nil {
					return VApp{}, fmt.Errorf("error decoding vdc response: %s", err)
				}

				// Querying the VApp
				req := vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

				resp, err := checkResp(vdc.client.Http.Do(req))
				if err != nil {
					return VApp{}, fmt.Errorf("error retrieving vApp: %s", err)
				}

				newvapp := NewVApp(vdc.client)

				if err = decodeBody(resp, newvapp.VApp); err != nil {
					return VApp{}, fmt.Errorf("error decoding vApp response: %s", err.Error())
				}

				return *newvapp, nil

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

			findUrl, err := url.ParseRequestURI(child.HREF)

			if err != nil {
				return VM{}, fmt.Errorf("error decoding vdc response: %s", err)
			}

			// Querying the VApp
			req := vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

			resp, err := checkResp(vdc.client.Http.Do(req))
			if err != nil {
				return VM{}, fmt.Errorf("error retrieving vm: %s", err)
			}

			newvm := NewVM(vdc.client)

			//body, err := ioutil.ReadAll(resp.Body)
			//fmt.Println(string(body))

			if err = decodeBody(resp, newvm.VM); err != nil {
				return VM{}, fmt.Errorf("error decoding vm response: %s", err.Error())
			}

			return *newvm, nil

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

				findUrl, err := url.ParseRequestURI(resent.HREF)

				if err != nil {
					return VApp{}, fmt.Errorf("error decoding vdc response: %s", err)
				}

				// Querying the VApp
				req := vdc.client.NewRequest(map[string]string{}, "GET", *findUrl, nil)

				resp, err := checkResp(vdc.client.Http.Do(req))
				if err != nil {
					return VApp{}, fmt.Errorf("error retrieving vApp: %s", err)
				}

				newvapp := NewVApp(vdc.client)

				if err = decodeBody(resp, newvapp.VApp); err != nil {
					return VApp{}, fmt.Errorf("error decoding vApp response: %s", err)
				}

				return *newvapp, nil

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
