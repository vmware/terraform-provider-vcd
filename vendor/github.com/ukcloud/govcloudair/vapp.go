/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"

	types "github.com/ukcloud/govcloudair/types/v56"
)

type VApp struct {
	VApp *types.VApp
	c    *Client
}

func NewVApp(c *Client) *VApp {
	return &VApp{
		VApp: new(types.VApp),
		c:    c,
	}
}

func (v *VCDClient) NewVApp(c *Client) VApp {
	newvapp := NewVApp(c)
	return *newvapp
}

func (v *VApp) Refresh() error {

	if v.VApp.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VApp.HREF)

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		log.Printf("[DEBUG] Error from HTTP Request: %#v", err)
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	v.VApp = &types.VApp{}

	if err = decodeBody(resp, v.VApp); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return nil
}

func composeSourceItems(vms []*types.NewVMDescription) []*types.SourcedCompositionItemParam {
	sourceItems := make([]*types.SourcedCompositionItemParam, len(vms))
	for i, vm := range vms {

		networkConnections := make([]*types.NetworkConnection, len(vm.Networks))
		networkAssignments := make([]*types.NetworkAssignment, len(vm.Networks))
		var primeryNetworkConnectionIndex int

		for index, orgnetwork := range vm.Networks {
			if orgnetwork.IsPrimary {
				primeryNetworkConnectionIndex = index
			}

			networkConnections[index] =
				&types.NetworkConnection{
					Network:                 orgnetwork.Name,
					NetworkConnectionIndex:  index,
					IsConnected:             orgnetwork.IsConnected,
					IPAddressAllocationMode: orgnetwork.IPAllocationMode,
					NetworkAdapterType:      orgnetwork.AdapterType,
				}

			networkAssignments[index] =
				&types.NetworkAssignment{
					InnerNetwork:     orgnetwork.Name,
					ContainerNetwork: orgnetwork.Name,
				}
		}

		sourceItems[i] = &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vm.VAppTemplate.Children.VM[0].HREF,
				Name: vm.Name,
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{
					Type: vm.VAppTemplate.Children.VM[0].NetworkConnectionSection.Type,
					HREF: vm.VAppTemplate.Children.VM[0].NetworkConnectionSection.HREF,
					// Info: "Network config for sourced item",
					PrimaryNetworkConnectionIndex: primeryNetworkConnectionIndex,
					NetworkConnection:             networkConnections,
				},
			},
			NetworkAssignment: networkAssignments,
		}

		// Add storage profile if it is providedpolation
		if vm.StorageProfile != nil {
			sourceItems[i].StorageProfile = vm.StorageProfile
		}

	}

	return sourceItems
}

func composeNetworkConfigs(orgnetworks []*types.OrgVDCNetwork) []*types.VAppNetworkConfiguration {

	networkConfigs := make([]*types.VAppNetworkConfiguration, len(orgnetworks))

	for index, orgnetwork := range orgnetworks {
		networkConfigs[index] = &types.VAppNetworkConfiguration{
			NetworkName: orgnetwork.Name,
			Configuration: &types.NetworkConfiguration{
				FenceMode: types.FenceModeBridged,
				ParentNetwork: &types.Reference{
					HREF: orgnetwork.HREF,
					Name: orgnetwork.Name,
					Type: orgnetwork.Type,
				},
			},
		}
	}

	return networkConfigs
}

func (v *VApp) AddVMs(vms []*types.NewVMDescription) (Task, error) {

	v.Refresh()
	task := NewTask(v.c)
	if v.VApp.Tasks != nil {
		for _, t := range v.VApp.Tasks.Task {
			task.Task = t
			err := task.WaitTaskCompletion()
			if err != nil {
				return Task{}, fmt.Errorf("Error performing task: %#v", err)
			}
		}
	}

	sourceItems := composeSourceItems(vms)

	vcomp := &types.ReComposeVAppParams{
		Ovf:         "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:         "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Deploy:      false,
		Name:        v.VApp.Name,
		PowerOn:     false,
		Description: v.VApp.Description,
		SourcedItem: sourceItems,
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	return ExecuteRequest(string(output),
		v.VApp.HREF+"/action/recomposeVApp",
		"POST",
		"application/vnd.vmware.vcloud.recomposeVAppParams+xml",
		v.c)
}

func (v *VApp) RemoveVMs(vms []*types.VM) (Task, error) {

	v.Refresh()
	task := NewTask(v.c)
	if v.VApp.Tasks != nil {
		for _, t := range v.VApp.Tasks.Task {
			task.Task = t
			err := task.WaitTaskCompletion()
			if err != nil {
				return Task{}, fmt.Errorf("Error performing task: %#v", err)
			}
		}
	}

	deleteItems := make([]*types.DeleteItem, len(vms))
	for index, vm := range vms {
		deleteItems[index] = &types.DeleteItem{
			HREF: vm.HREF,
		}
	}

	vcomp := &types.ReComposeVAppParams{
		Ovf:        "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:        "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:      "http://www.vmware.com/vcloud/v1.5",
		DeleteItem: deleteItems,
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	return ExecuteRequest(string(output),
		v.VApp.HREF+"/action/recomposeVApp",
		"POST",
		"application/vnd.vmware.vcloud.recomposeVAppParams+xml",
		v.c)
}

func (v *VApp) ChangeNetworks(orgnetworks []*types.OrgVDCNetwork) (Task, error) {

	networkConfigs := composeNetworkConfigs(orgnetworks)

	vcomp := &types.ReComposeVAppParams{
		Ovf:         "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:         "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Deploy:      false,
		Name:        v.VApp.Name,
		PowerOn:     false,
		Description: v.VApp.Description,
		InstantiationParams: &types.InstantiationParams{
			NetworkConfigSection: &types.NetworkConfigSection{
				// Info:          "Configuration parameters for logical networks",
				NetworkConfig: networkConfigs,
			},
		},
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	return ExecuteRequest(string(output),
		v.VApp.HREF+"/action/recomposeVApp",
		"POST",
		"application/vnd.vmware.vcloud.recomposeVAppParams+xml",
		v.c)
}

func (v *VApp) ComposeVApp(name string, description string, orgnetworks []*types.OrgVDCNetwork, vAppNetworkConfigurations []*types.VAppNetworkConfiguration) (Task, error) {

	// if vapptemplate.VAppTemplate.Children == nil || orgvdcnetwork.OrgVDCNetwork == nil {
	// 	return Task{}, fmt.Errorf("can't compose a new vApp, objects passed are not valid")
	// }

	networkConfigs := composeNetworkConfigs(orgnetworks)
	networkConfigs = append(networkConfigs, vAppNetworkConfigurations...)
	// networkConfigs := vAppNetworkConfigurations

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
				// Info:          "Configuration parameters for logical networks",
				NetworkConfig: networkConfigs,
			},
		},
	}

	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling vapp compose: %s", err)
	}

	log.Printf("[TRACE] XML vApp: %s\n", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s := v.c.VCDVDCHREF
	s.Path += "/action/composeVApp"

	req := v.c.NewRequest(map[string]string{}, "POST", s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.composeVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		log.Printf("[DEBUG] Error from HTTP Request: %#v", err)
		return Task{}, fmt.Errorf("error instantiating a new vApp: %s", err)
	}

	if err = decodeBody(resp, v.VApp); err != nil {
		return Task{}, fmt.Errorf("error decoding vApp response: %s", err)
	}

	task := NewTask(v.c)
	task.Task = v.VApp.Tasks.Task[0]

	// The request was successful
	return *task, nil
}

func (v *VApp) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               "http://www.vmware.com/vcloud/v1.5",
		UndeployPowerAction: "powerOff",
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return ExecuteRequest(string(output),
		v.VApp.HREF+"/action/undeploy",
		"POST",
		"application/vnd.vmware.vcloud.undeployVAppParams+xml",
		v.c)

}

func (v *VApp) Deploy() (Task, error) {

	vu := &types.DeployVAppParams{
		Xmlns:   "http://www.vmware.com/vcloud/v1.5",
		PowerOn: false,
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return ExecuteRequest(string(output),
		v.VApp.HREF+"/action/deploy",
		"POST",
		"application/vnd.vmware.vcloud.deployVAppParams+xml",
		v.c)
}

func (v *VApp) Delete() (Task, error) {
	return ExecuteRequest("",
		v.VApp.HREF,
		"DELETE",
		"",
		v.c)
}

func (v *VApp) GetStatus() (string, error) {
	err := v.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing vapp: %v", err)
	}
	return types.VAppStatuses[v.VApp.Status], nil
}

func (v *VApp) GetVmByName(name string) (*types.VM, error) {
	err := v.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vapp: %v", err)
	}
	for _, vm := range v.VApp.Children.VM {
		if vm.Name == name {
			return vm, nil
		}
	}
	return nil, nil
}

func (v *VApp) GetVmByHREF(href string) (*types.VM, error) {
	err := v.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vapp: %v", err)
	}
	for _, vm := range v.VApp.Children.VM {
		if vm.HREF == href {
			return vm, nil
		}
	}
	return nil, nil
}

func (v *VApp) GetNetworkByName(name string) (*types.VAppNetworkConfiguration, error) {
	err := v.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vapp: %v", err)
	}
	for _, networkConfig := range v.VApp.NetworkConfigSection.NetworkConfig {
		if networkConfig.NetworkName == name {
			return networkConfig, nil
		}
	}
	return nil, nil
}

func (v *VApp) GetNetworkNames() ([]string, error) {
	err := v.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vapp: %v", err)
	}

	networks := make([]string, len(v.VApp.NetworkConfigSection.NetworkConfig))
	for index := range v.VApp.NetworkConfigSection.NetworkConfig {
		networks[index] = v.VApp.NetworkConfigSection.NetworkConfig[index].NetworkName
	}

	return networks, nil

}
