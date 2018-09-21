/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	"os"

	types "github.com/vmware/go-vcloud-director/types/v56"
	"strconv"
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

// Returns the vdc where the vapp resides in.
func (v *VApp) getParentVDC() (Vdc, error) {
	for _, a := range v.VApp.Link {
		if a.Type == "application/vnd.vmware.vcloud.vdc+xml" {
			u, err := url.ParseRequestURI(a.HREF)
			if err != nil {
				return Vdc{}, fmt.Errorf("Cannot parse HREF : %v", err)
			}
			req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)
			resp, err := checkResp(v.c.Http.Do(req))

			vdc := NewVdc(v.c)
			if err = decodeBody(resp, vdc.Vdc); err != nil {
				return Vdc{}, fmt.Errorf("error decoding task response: %s", err)
			}
			return *vdc, nil
		}
	}
	return Vdc{}, fmt.Errorf("Could not find a parent Vdc")
}

func (v *VApp) Refresh() error {

	if v.VApp.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VApp.HREF)

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
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

func (v *VApp) AddVM(orgvdcnetworks []*types.OrgVDCNetwork, vapptemplate VAppTemplate, name string, acceptalleulas bool) (Task, error) {

	vcomp := &types.ReComposeVAppParams{
		Ovf:         "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:         "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Deploy:      false,
		Name:        v.VApp.Name,
		PowerOn:     false,
		Description: v.VApp.Description,
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
				Name: name,
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
		AllEULAsAccepted: acceptalleulas,
	}

	for index, orgvdcnetwork := range orgvdcnetworks {
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

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/recomposeVApp"

	log.Printf("[TRACE] Recompose XML: %s", string(output))

	b := bytes.NewBufferString(xml.Header + string(output))

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.recomposeVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}

	return *task, nil
}

func (v *VApp) RemoveVM(vm VM) error {

	v.Refresh()
	task := NewTask(v.c)
	if v.VApp.Tasks != nil {
		for _, t := range v.VApp.Tasks.Task {
			task.Task = t
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error performing task: %#v", err)
			}
		}
	}

	vcomp := &types.ReComposeVAppParams{
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:   "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		DeleteItem: &types.DeleteItem{
			HREF: vm.VM.HREF,
		},
	}

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/recomposeVApp"

	b := bytes.NewBufferString(xml.Header + string(output))

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.recomposeVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error instantiating a new vApp: %s", err)
	}

	task = NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("Error performing task: %#v", err)
	}

	return nil
}

func (v *VApp) PowerOn() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/powerOn"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering on vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) PowerOff() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/powerOff"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error powering off vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Reboot() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/reboot"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error rebooting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Reset() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/reset"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error resetting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Suspend() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/suspend"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error suspending vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Shutdown() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/power/action/shutdown"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error shutting down vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

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

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/undeploy"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.undeployVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error undeploy vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

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

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/action/deploy"

	req := v.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.deployVAppParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error undeploy vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) Delete() (Task, error) {

	s, _ := url.ParseRequestURI(v.VApp.HREF)

	req := v.c.NewRequest(map[string]string{}, "DELETE", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting vApp: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) RunCustomizationScript(computername, script string) (Task, error) {
	return v.Customize(computername, script, false)
}

func (v *VApp) Customize(computername, script string, changeSid bool) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Xsi:   "http://www.w3.org/2001/XMLSchema-instance",
		Xmlns: "http://www.vmware.com/vcloud/v1.5",

		HREF:                v.VApp.Children.VM[0].HREF,
		Type:                "application/vnd.vmware.vcloud.guestCustomizationSection+xml",
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             true,
		ComputerName:        computername,
		CustomizationScript: script,
		ChangeSid:           false,
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] VCD Client configuration: %s", output)

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/guestCustomizationSection/"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.guestCustomizationSection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (v *VApp) GetStatus() (string, error) {
	err := v.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing vapp: %v", err)
	}
	return types.VAppStatuses[v.VApp.Status], nil
}

func (v *VApp) GetNetworkConnectionSection() (*types.NetworkConnectionSection, error) {

	networkConnectionSection := &types.NetworkConnectionSection{}

	if v.VApp.Children.VM[0].HREF == "" {
		return networkConnectionSection, fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF + "/networkConnectionSection/")

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkConnectionSection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return networkConnectionSection, fmt.Errorf("error retrieving task: %s", err)
	}

	if err = decodeBody(resp, networkConnectionSection); err != nil {
		return networkConnectionSection, fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return networkConnectionSection, nil
}

func (v *VApp) ChangeCPUcount(size int) (Task, error) {

	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newcpu := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(size) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    3,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newcpu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/virtualHardwareSection/cpu"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) ChangeStorageProfile(name string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	vdc, err := v.getParentVDC()
	storageprofileref, err := vdc.FindStorageProfileReference(name)

	newprofile := &types.VM{
		Name:           v.VApp.Children.VM[0].Name,
		StorageProfile: &storageprofileref,
		Xmlns:          "http://www.vmware.com/vcloud/v1.5",
	}

	output, err := xml.MarshalIndent(newprofile, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] VCD Client configuration: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.vm+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) ChangeVMName(name string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newname := &types.VM{
		Name:  name,
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
	}

	output, err := xml.MarshalIndent(newname, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] VCD Client configuration: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.vm+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) DeleteMetadata(key string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/metadata/" + key

	req := v.c.NewRequest(map[string]string{}, "DELETE", *s, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting Metadata: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (v *VApp) AddMetadata(key, value string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newmetadata := &types.MetadataValue{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Xsi:   "http://www.w3.org/2001/XMLSchema-instance",
		TypedValue: &types.TypedValue{
			XsiType: "MetadataStringValue",
			Value:   value,
		},
	}

	output, err := xml.MarshalIndent(newmetadata, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] NetworkXML: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/metadata/" + key

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.metadata.value+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM Network: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) SetOvf(parameters map[string]string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	if v.VApp.Children.VM[0].ProductSection == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children with ProductSection, aborting customization")
	}

	for key, value := range parameters {
		for _, ovf_value := range v.VApp.Children.VM[0].ProductSection.Property {
			if ovf_value.Key == key {
				ovf_value.Value = &types.Value{Value: value}
				break
			}
		}
	}

	newmetadata := &types.ProductSectionList{
		Xmlns:          "http://www.vmware.com/vcloud/v1.5",
		Ovf:            "http://schemas.dmtf.org/ovf/envelope/1",
		ProductSection: v.VApp.Children.VM[0].ProductSection,
	}

	output, err := xml.MarshalIndent(newmetadata, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] NetworkXML: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/productSections"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.productSections+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM Network: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) ChangeNetworkConfig(networks []map[string]interface{}, ip string) (Task, error) {
	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	networksection, err := v.GetNetworkConnectionSection()

	for index, network := range networks {
		// Determine what type of address is requested for the vApp
		ipAllocationMode := "NONE"
		ipAddress := "Any"

		// TODO: Review current behaviour of using DHCP when left blank
		if ip == "" || ip == "dhcp" || network["ip"] == "dhcp" {
			ipAllocationMode = "DHCP"
		} else if ip == "allocated" || network["ip"] == "allocated" {
			ipAllocationMode = "POOL"
		} else if ip == "none" || network["ip"] == "none" {
			ipAllocationMode = "NONE"
		} else if ip != "" || network["ip"] != "" {
			ipAllocationMode = "MANUAL"
			// TODO: Check a valid IP has been given
			ipAddress = ip
		}

		log.Printf("[DEBUG] Function ChangeNetworkConfig() for %s invoked", network["orgnetwork"])

		networksection.Xmlns = "http://www.vmware.com/vcloud/v1.5"
		networksection.Ovf = "http://schemas.dmtf.org/ovf/envelope/1"
		networksection.Info = "Specifies the available VM network connections"

		networksection.NetworkConnection[index].NeedsCustomization = true
		networksection.NetworkConnection[index].IPAddress = ipAddress
		networksection.NetworkConnection[index].IPAddressAllocationMode = ipAllocationMode
		networksection.NetworkConnection[index].MACAddress = ""

		if network["is_primary"] == true {
			networksection.PrimaryNetworkConnectionIndex = index
		}

	}

	output, err := xml.MarshalIndent(networksection, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] NetworkXML: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/networkConnectionSection/"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkConnectionSection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM Network: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (v *VApp) ChangeMemorySize(size int) (Task, error) {

	err := v.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vapp before running customization: %v", err)
	}

	// Check if VApp Children is populated
	if v.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newmem := &types.OVFItem{
		XmlnsRasd:       "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData",
		XmlnsVCloud:     "http://www.vmware.com/vcloud/v1.5",
		XmlnsXsi:        "http://www.w3.org/2001/XMLSchema-instance",
		VCloudHREF:      v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
		VCloudType:      "application/vnd.vmware.vcloud.rasdItem+xml",
		AllocationUnits: "byte * 2^20",
		Description:     "Memory Size",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    4,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: v.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: "application/vnd.vmware.vcloud.rasdItem+xml",
		},
	}

	output, err := xml.MarshalIndent(newmem, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	debug := os.Getenv("GOVCLOUDAIR_DEBUG")

	if debug == "true" {
		fmt.Printf("\n\nXML DEBUG: %s\n\n", string(output))
	}

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.Children.VM[0].HREF)
	s.Path += "/virtualHardwareSection/memory"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.rasdItem+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error customizing VM: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VApp) GetNetworkConfig() (*types.NetworkConfigSection, error) {

	networkConfig := &types.NetworkConfigSection{}

	if v.VApp.HREF == "" {
		return networkConfig, fmt.Errorf("cannot refresh, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VApp.HREF + "/networkConfigSection/")

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkConfigSection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return networkConfig, fmt.Errorf("error retrieving task: %s", err)
	}

	if err = decodeBody(resp, networkConfig); err != nil {
		return networkConfig, fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return networkConfig, nil
}

func (v *VApp) AddRAWNetworkConfig(orgvdcnetworks []*types.OrgVDCNetwork) (Task, error) {

	networkConfig := &types.NetworkConfigSection{
		Info:  "Configuration parameters for logical networks",
		Ovf:   "http://schemas.dmtf.org/ovf/envelope/1",
		Type:  "application/vnd.vmware.vcloud.networkConfigSection+xml",
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
	}

	for _, network := range orgvdcnetworks {
		networkConfig.NetworkConfig = append(networkConfig.NetworkConfig,
			types.VAppNetworkConfiguration{
				NetworkName: network.Name,
				Configuration: &types.NetworkConfiguration{
					ParentNetwork: &types.Reference{
						HREF: network.HREF,
					},
					FenceMode: "bridged",
				},
			},
		)
	}

	output, err := xml.MarshalIndent(networkConfig, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[DEBUG] RAWNETWORK Config NetworkXML: %s", output)

	b := bytes.NewBufferString(xml.Header + string(output))

	s, _ := url.ParseRequestURI(v.VApp.HREF)
	s.Path += "/networkConfigSection/"

	req := v.c.NewRequest(map[string]string{}, "PUT", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.networkconfigsection+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error adding vApp Network: %s", err)
	}

	task := NewTask(v.c)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}
