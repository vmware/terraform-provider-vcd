/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	types "github.com/ukcloud/govcloudair/types/v56"
)

type VM struct {
	VM *types.VM
	c  *Client
}

func NewVM(c *Client) *VM {
	return &VM{
		VM: new(types.VM),
		c:  c,
	}
}

func (v *VM) GetStatus() (string, error) {
	err := v.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %v", err)
	}
	return types.VAppStatuses[v.VM.Status], nil
}

func (v *VM) Refresh() error {

	if v.VM.HREF == "" {
		return fmt.Errorf("cannot refresh VM, Object is empty")
	}

	u, _ := url.ParseRequestURI(v.VM.HREF)

	req := v.c.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	v.VM = &types.VM{}

	if err = decodeBody(resp, v.VM); err != nil {
		return fmt.Errorf("error decoding task response VM: %s", err)
	}

	// The request was successful
	return nil
}

func ExecuteRequest(payload, path, type_, contentType string, client *Client) (Task, error) {
	s, _ := url.ParseRequestURI(path)

	var req *http.Request
	if type_ == "POST" {
		b := bytes.NewBufferString(xml.Header + payload)
		req = client.NewRequest(map[string]string{}, type_, *s, b)
	} else if type_ == "GET" {
		req = client.NewRequest(map[string]string{}, type_, *s, nil)

	}

	req.Header.Add("Content-Type", contentType)

	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring VM: %s", err)
	}

	task := NewTask(client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (v *VM) Reconfigure() (Task, error) {
	// err := v.Refresh()
	// if err != nil {
	// 	return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	// }

	// WORKAROUND for XML namespace support in go, see bottom of types.go
	// github.com/ukcloud/govcloudair/types/v56
	log.Printf("[TRACE] (%s) Testing variable what3: %s", v.VM.Name, v.VM.NestedHypervisorEnabled)
	v.correctAddressOnParentForNetworkHardware()

	log.Printf("[TRACE] (%s) Testing variable what4: %s", v.VM.Name, v.VM.NestedHypervisorEnabled)
	v.SetXMLNamespaces()

	ovfVirtualHardwareSection := v.VM.VirtualHardwareSection.ConvertToOVF()
	log.Printf("[TRACE] (%s) Testing variable what5: %s", v.VM.Name, v.VM.NestedHypervisorEnabled)
	virtualHardwareSection := v.VM.VirtualHardwareSection
	v.VM.VirtualHardwareSection = nil
	v.VM.OVFVirtualHardwareSection = ovfVirtualHardwareSection
	log.Printf("[TRACE] (%s) Testing variable what6: %s", v.VM.Name, v.VM.NestedHypervisorEnabled)

	output, err := xml.MarshalIndent(v.VM, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Printf("[TRACE] (%s) Testing variable what7: %s", v.VM.Name, v.VM.NestedHypervisorEnabled)

	v.VM.VirtualHardwareSection = virtualHardwareSection

	v.VM.OVFVirtualHardwareSection = nil

	log.Printf("[DEBUG] VM: %s", output)

	return ExecuteRequest(string(output),
		v.VM.HREF+"/action/reconfigureVm",
		"POST",
		"application/vnd.vmware.vcloud.vm+xml",
		v.c)
}

func (c *VCDClient) FindVMByHREF(vmhref string) (VM, error) {

	u, err := url.ParseRequestURI(vmhref)

	if err != nil {
		return VM{}, fmt.Errorf("error decoding vm HREF: %s", err)
	}

	// Querying the VApp
	req := c.Client.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return VM{}, fmt.Errorf("error retrieving VM: %s", err)
	}

	newvm := NewVM(&c.Client)

	if err = decodeBody(resp, newvm.VM); err != nil {
		return VM{}, fmt.Errorf("error decoding VM response: %s", err)
	}

	return *newvm, nil

}

func (v *VM) PowerOn() (Task, error) {
	return ExecuteRequest("",
		v.VM.HREF+"/power/action/powerOn",
		"POST",
		"application/vnd.vmware.vcloud.vm+xml",
		v.c)
}

func (v *VM) PowerOff() (Task, error) {
	return ExecuteRequest("",
		v.VM.HREF+"/power/action/powerOff",
		"POST",
		"application/vnd.vmware.vcloud.vm+xml",
		v.c)
}

func (v *VM) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               "http://www.vmware.com/vcloud/v1.5",
		UndeployPowerAction: "powerOff",
	}

	output, err := xml.MarshalIndent(vu, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return ExecuteRequest(string(output),
		v.VM.HREF+"/action/undeploy",
		"POST",
		"application/vnd.vmware.vcloud.undeployVAppParams+xml",
		v.c)
}

func (v *VM) getVirtualHardwareItemsByResourceType(resourceType int) ([]*types.VirtualHardwareItem, error) {
	err := v.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	var items []*types.VirtualHardwareItem

	for _, item := range v.VM.VirtualHardwareSection.Item {
		if item.ResourceType == resourceType {
			items = append(items, item)
		}
	}

	return items, nil
}

func (v *VM) GetCPUCount() (int, error) {
	items, err := v.getVirtualHardwareItemsByResourceType(types.ResourceTypeProcessor)
	if err != nil {
		return 0, err
	}

	// The amount of CPU items must be one
	if len(items) != 1 {
		return 0, fmt.Errorf("error: Did not find any CPU on the given vm (%s)", v.VM.Name)
	}

	return items[0].VirtualQuantity, nil
}

func (v *VM) GetMemoryCount() (int, error) {
	items, err := v.getVirtualHardwareItemsByResourceType(types.ResourceTypeMemory)
	if err != nil {
		return 0, err
	}

	// The amount of memory items must be one
	if len(items) != 1 {
		return 0, fmt.Errorf("error: Did not find any Memory on the given vm (%s)", v.VM.Name)
	}

	return items[0].VirtualQuantity, nil
}

func (v *VM) correctAddressOnParentForNetworkHardware() error {
	for index := range v.VM.VirtualHardwareSection.Item {
		if v.VM.VirtualHardwareSection.Item[index].ResourceType == types.ResourceTypeEthernet {
			v.VM.VirtualHardwareSection.Item[index].AddressOnParent = v.VM.VirtualHardwareSection.Item[index].InstanceID
		}
	}
	return nil
}

func (v *VM) SetXMLNamespaces() {
	v.VM.Xmlns = "http://www.vmware.com/vcloud/v1.5"
	v.VM.Vcloud = "http://www.vmware.com/vcloud/v1.5"
	v.VM.Ovf = "http://schemas.dmtf.org/ovf/envelope/1"
	v.VM.Vmw = "http://www.vmware.com/schema/ovf"
	v.VM.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
	v.VM.Rasd = "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData"
	v.VM.Vssd = "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData"
}

func (v *VM) SetCPUCount(count int) {
	for index := range v.VM.VirtualHardwareSection.Item {
		if v.VM.VirtualHardwareSection.Item[index].ResourceType == types.ResourceTypeProcessor {
			v.VM.VirtualHardwareSection.Item[index].ElementName = strconv.Itoa(count) + " virtual CPU(s)"
			v.VM.VirtualHardwareSection.Item[index].VirtualQuantity = count
		}
	}
}

func (v *VM) SetMemoryCount(count int) {
	for index := range v.VM.VirtualHardwareSection.Item {
		if v.VM.VirtualHardwareSection.Item[index].ResourceType == types.ResourceTypeMemory {
			v.VM.VirtualHardwareSection.Item[index].ElementName = strconv.Itoa(count) + " MB of memory"
			v.VM.VirtualHardwareSection.Item[index].VirtualQuantity = count
		}
	}
}

func (v *VM) SetNestedHypervisor(value bool) {
	v.VM.NestedHypervisorEnabled = value
}

func (v *VM) SetNestedHypervisorWithRequest(value bool) (Task, error) {
	url := ""
	if value {
		url = "/action/enableNestedHypervisor"
	} else {
		url = "/action/disableNestedHypervisor"
	}

	return ExecuteRequest("",
		v.VM.HREF+url,
		"POST",
		"application/vnd.vmware.vcloud.vm+xml",
		v.c)
}

// func (v *VM) SetStorageProfile(name string, meta interface{}) error {
// 	vcdClient := meta.(*govcloudair.VCDClient)

// }

func (v *VM) SetInitscript(value string) {
	v.VM.GuestCustomizationSection.CustomizationScript = value
}

func (v *VM) SetNetworkConnectionSection(networks *types.NetworkConnectionSection) {
	v.VM.NetworkConnectionSection = networks
}

func (v *VM) SetAdminPasswordAuto(value bool) {
	v.VM.GuestCustomizationSection.AdminPasswordAuto = value
}

func (v *VM) SetAdminPassword(value string) {
	v.VM.GuestCustomizationSection.AdminPassword = value
}

func (v *VM) SetHostName(value string) {
	v.VM.GuestCustomizationSection.ComputerName = value
}
