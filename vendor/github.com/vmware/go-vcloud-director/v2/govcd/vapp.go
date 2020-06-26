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
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type VApp struct {
	VApp   *types.VApp
	client *Client
}

func NewVApp(cli *Client) *VApp {
	return &VApp{
		VApp:   new(types.VApp),
		client: cli,
	}
}

func (vcdCli *VCDClient) NewVApp(client *Client) VApp {
	newvapp := NewVApp(client)
	return *newvapp
}

// struct type used to pass information for vApp network creation
type VappNetworkSettings struct {
	ID                 string
	Name               string
	Description        string
	Gateway            string
	NetMask            string
	DNS1               string
	DNS2               string
	DNSSuffix          string
	GuestVLANAllowed   *bool
	StaticIPRanges     []*types.IPRange
	DhcpSettings       *DhcpSettings
	RetainIpMacEnabled *bool
	VappFenceEnabled   *bool
}

// struct type used to pass information for vApp network DHCP
type DhcpSettings struct {
	IsEnabled        bool
	MaxLeaseTime     int
	DefaultLeaseTime int
	IPRange          *types.IPRange
}

// Returns the vdc where the vapp resides in.
func (vapp *VApp) getParentVDC() (Vdc, error) {
	for _, link := range vapp.VApp.Link {
		if link.Type == "application/vnd.vmware.vcloud.vdc+xml" {

			vdc := NewVdc(vapp.client)

			_, err := vapp.client.ExecuteRequest(link.HREF, http.MethodGet,
				"", "error retrieving paren vdc: %s", nil, vdc.Vdc)
			if err != nil {
				return Vdc{}, err
			}

			return *vdc, nil
		}
	}
	return Vdc{}, fmt.Errorf("could not find a parent Vdc")
}

func (vapp *VApp) Refresh() error {

	if vapp.VApp.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url := vapp.VApp.HREF
	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	vapp.VApp = &types.VApp{}

	_, err := vapp.client.ExecuteRequest(url, http.MethodGet,
		"", "error refreshing vApp: %s", nil, vapp.VApp)

	// The request was successful
	return err
}

// AddVM create vm in vApp using vApp template
// orgVdcNetworks - adds org VDC networks to be available for vApp. Can be empty.
// vappNetworkName - adds vApp network to be available for vApp. Can be empty.
// vappTemplate - vApp Template which will be used for VM creation.
// name - name for VM.
// acceptAllEulas - setting allows to automatically accept or not Eulas.
//
// Deprecated: Use vapp.AddNewVM instead for more sophisticated network handling
func (vapp *VApp) AddVM(orgVdcNetworks []*types.OrgVDCNetwork, vappNetworkName string, vappTemplate VAppTemplate, name string, acceptAllEulas bool) (Task, error) {
	util.Logger.Printf("[INFO] vapp.AddVM() is deprecated in favor of vapp.AddNewVM()")
	if vappTemplate == (VAppTemplate{}) || vappTemplate.VAppTemplate == nil {
		return Task{}, fmt.Errorf("vApp Template can not be empty")
	}

	// primaryNetworkConnectionIndex will be inherited from template or defaulted to 0
	// if the template does not have any NICs assigned.
	primaryNetworkConnectionIndex := 0
	if vappTemplate.VAppTemplate.Children != nil && len(vappTemplate.VAppTemplate.Children.VM) > 0 &&
		vappTemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection != nil {
		primaryNetworkConnectionIndex = vappTemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.PrimaryNetworkConnectionIndex
	}

	networkConnectionSection := types.NetworkConnectionSection{
		Info:                          "Network config for sourced item",
		PrimaryNetworkConnectionIndex: primaryNetworkConnectionIndex,
	}

	for index, orgVdcNetwork := range orgVdcNetworks {
		networkConnectionSection.NetworkConnection = append(networkConnectionSection.NetworkConnection,
			&types.NetworkConnection{
				Network:                 orgVdcNetwork.Name,
				NetworkConnectionIndex:  index,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
			},
		)
	}

	if vappNetworkName != "" {
		networkConnectionSection.NetworkConnection = append(networkConnectionSection.NetworkConnection,
			&types.NetworkConnection{
				Network:                 vappNetworkName,
				NetworkConnectionIndex:  len(orgVdcNetworks),
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
			},
		)
	}

	return vapp.AddNewVM(name, vappTemplate, &networkConnectionSection, acceptAllEulas)
}

// AddNewVM adds VM from vApp template with custom NetworkConnectionSection
func (vapp *VApp) AddNewVM(name string, vappTemplate VAppTemplate, network *types.NetworkConnectionSection, acceptAllEulas bool) (Task, error) {
	return vapp.AddNewVMWithStorageProfile(name, vappTemplate, network, nil, acceptAllEulas)
}

// AddNewVMWithStorageProfile adds VM from vApp template with custom NetworkConnectionSection and optional storage profile
func (vapp *VApp) AddNewVMWithStorageProfile(name string, vappTemplate VAppTemplate,
	network *types.NetworkConnectionSection,
	storageProfileRef *types.Reference, acceptAllEulas bool) (Task, error) {

	if vappTemplate == (VAppTemplate{}) || vappTemplate.VAppTemplate == nil {
		return Task{}, fmt.Errorf("vApp Template can not be empty")
	}

	templateHref := vappTemplate.VAppTemplate.HREF
	if vappTemplate.VAppTemplate.Children != nil && len(vappTemplate.VAppTemplate.Children.VM) != 0 {
		templateHref = vappTemplate.VAppTemplate.Children.VM[0].HREF
	}

	// Status 8 means The object is resolved and powered off.
	// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/94b8bd8d-74ff-4fe3-b7a4-41ae31516ed7/1b42f3b5-8b31-4279-8b3f-547f6c7c5aa8/doc/GUID-843BE3AD-5EF6-4442-B864-BCAE44A51867.html
	if vappTemplate.VAppTemplate.Status != 8 {
		return Task{}, fmt.Errorf("vApp Template shape is not ok (status: %d)", vappTemplate.VAppTemplate.Status)
	}

	// Validate network config only if it was supplied
	if network != nil && network.NetworkConnection != nil {
		for _, nic := range network.NetworkConnection {
			if nic.Network == "" {
				return Task{}, fmt.Errorf("missing mandatory attribute Network: %s", nic.Network)
			}
			if nic.IPAddressAllocationMode == "" {
				return Task{}, fmt.Errorf("missing mandatory attribute IPAddressAllocationMode: %s", nic.IPAddressAllocationMode)
			}
		}
	}

	vAppComposition := &types.ReComposeVAppParams{
		Ovf:         types.XMLNamespaceOVF,
		Xsi:         types.XMLNamespaceXSI,
		Xmlns:       types.XMLNamespaceVCloud,
		Deploy:      false,
		Name:        vapp.VApp.Name,
		PowerOn:     false,
		Description: vapp.VApp.Description,
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: templateHref,
				Name: name,
			},
			InstantiationParams: &types.InstantiationParams{}, // network config is injected below
		},
		AllEULAsAccepted: acceptAllEulas,
	}

	// Add storage profile
	if storageProfileRef != nil && storageProfileRef.HREF != "" {
		vAppComposition.SourcedItem.StorageProfile = storageProfileRef
	}

	// Inject network config
	vAppComposition.SourcedItem.InstantiationParams.NetworkConnectionSection = network

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/recomposeVApp"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeRecomposeVappParams, "error instantiating a new VM: %s", vAppComposition)

}

// ========================= issue#252 ==================================
// TODO: To be refactored, handling networks better. See issue#252 for details
// https://github.com/vmware/go-vcloud-director/issues/252
// ======================================================================
func (vapp *VApp) RemoveVM(vm VM) error {

	err := vapp.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vApp before removing VM: %s", err)
	}
	task := NewTask(vapp.client)
	if vapp.VApp.Tasks != nil {
		for _, taskItem := range vapp.VApp.Tasks.Task {
			task.Task = taskItem
			// Leftover tasks may have unhandled errors that can be dismissed at this stage
			// we complete any incomplete tasks at this stage, to finish the refresh.
			if task.Task.Status != "error" && task.Task.Status != "success" {
				err := task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("error performing task: %s", err)
				}
			}
		}
	}

	vcomp := &types.ReComposeVAppParams{
		Ovf:   types.XMLNamespaceOVF,
		Xsi:   types.XMLNamespaceXSI,
		Xmlns: types.XMLNamespaceVCloud,
		DeleteItem: &types.DeleteItem{
			HREF: vm.VM.HREF,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/recomposeVApp"

	deleteTask, err := vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeRecomposeVappParams, "error removing VM: %s", vcomp)
	if err != nil {
		return err
	}

	err = deleteTask.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error performing removing VM task: %s", err)
	}

	return nil
}

func (vapp *VApp) PowerOn() (Task, error) {

	err := vapp.BlockWhileStatus("UNRESOLVED", vapp.client.MaxRetryTimeout)
	if err != nil {
		return Task{}, fmt.Errorf("error powering on vApp: %s", err)
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/powerOn"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering on vApp: %s", nil)
}

func (vapp *VApp) PowerOff() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/powerOff"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering off vApp: %s", nil)

}

func (vapp *VApp) Reboot() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/reboot"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error rebooting vApp: %s", nil)
}

func (vapp *VApp) Reset() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/reset"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error resetting vApp: %s", nil)
}

func (vapp *VApp) Suspend() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/suspend"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error suspending vApp: %s", nil)
}

func (vapp *VApp) Shutdown() (Task, error) {

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/power/action/shutdown"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error shutting down vApp: %s", nil)
}

func (vapp *VApp) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               types.XMLNamespaceVCloud,
		UndeployPowerAction: "powerOff",
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/undeploy"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeUndeployVappParams, "error undeploy vApp: %s", vu)
}

func (vapp *VApp) Deploy() (Task, error) {

	vu := &types.DeployVAppParams{
		Xmlns:   types.XMLNamespaceVCloud,
		PowerOn: false,
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/deploy"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeDeployVappParams, "error deploy vApp: %s", vu)
}

func (vapp *VApp) Delete() (Task, error) {

	// Return the task
	return vapp.client.ExecuteTaskRequest(vapp.VApp.HREF, http.MethodDelete,
		"", "error deleting vApp: %s", nil)
}

func (vapp *VApp) RunCustomizationScript(computername, script string) (Task, error) {
	return vapp.Customize(computername, script, false)
}

// Customize applies customization to first child VM
//
// Deprecated: Use vm.SetGuestCustomizationSection()
func (vapp *VApp) Customize(computername, script string, changeSid bool) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	// Check if VApp Children is populated
	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   types.XMLNamespaceOVF,
		Xsi:   types.XMLNamespaceXSI,
		Xmlns: types.XMLNamespaceVCloud,

		HREF:                vapp.VApp.Children.VM[0].HREF,
		Type:                types.MimeGuestCustomizationSection,
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             takeBoolPointer(true),
		ComputerName:        computername,
		CustomizationScript: script,
		ChangeSid:           takeBoolPointer(changeSid),
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.Children.VM[0].HREF)
	apiEndpoint.Path += "/guestCustomizationSection/"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeGuestCustomizationSection, "error customizing VM: %s", vu)
}

func (vapp *VApp) GetStatus() (string, error) {
	err := vapp.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing vApp: %s", err)
	}
	// Trying to make this function future-proof:
	// If a new status is added to a future vCD API and the status map in types.go
	// is not updated, we may get a panic.
	// Using the ", ok" construct we take control of the data lookup and are able to fail
	// gracefully.
	statusText, ok := types.VAppStatuses[vapp.VApp.Status]
	if ok {
		return statusText, nil
	}
	return "", fmt.Errorf("status %d does not have a description in types.VappStatuses", vapp.VApp.Status)
}

// BlockWhileStatus blocks until the status of vApp exits unwantedStatus.
// It sleeps 200 milliseconds between iterations and times out after timeOutAfterSeconds
// of seconds.
func (vapp *VApp) BlockWhileStatus(unwantedStatus string, timeOutAfterSeconds int) error {
	timeoutAfter := time.After(time.Duration(timeOutAfterSeconds) * time.Second)
	tick := time.NewTicker(200 * time.Millisecond)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timed out waiting for vApp to exit state %s after %d seconds",
				unwantedStatus, timeOutAfterSeconds)
		case <-tick.C:
			currentStatus, err := vapp.GetStatus()

			if err != nil {
				return fmt.Errorf("could not get vApp status %s", err)
			}
			if currentStatus != unwantedStatus {
				return nil
			}
		}
	}
}

func (vapp *VApp) GetNetworkConnectionSection() (*types.NetworkConnectionSection, error) {

	networkConnectionSection := &types.NetworkConnectionSection{}

	if vapp.VApp.Children.VM[0].HREF == "" {
		return networkConnectionSection, fmt.Errorf("cannot refresh, Object is empty")
	}

	_, err := vapp.client.ExecuteRequest(vapp.VApp.Children.VM[0].HREF+"/networkConnectionSection/", http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving network connection: %s", nil, networkConnectionSection)

	// The request was successful
	return networkConnectionSection, err
}

// Sets number of available virtual logical processors
// (i.e. CPUs x cores per socket)
// https://communities.vmware.com/thread/576209
// Deprecated: Use vm.ChangeCPUcount()
func (vapp *VApp) ChangeCPUCount(virtualCpuCount int) (Task, error) {
	return vapp.ChangeCPUCountWithCore(virtualCpuCount, nil)
}

// Sets number of available virtual logical processors
// (i.e. CPUs x cores per socket) and cores per socket.
// Socket count is a result of: virtual logical processors/cores per socket
// https://communities.vmware.com/thread/576209
// Deprecated: Use vm.ChangeCPUCountWithCore()
func (vapp *VApp) ChangeCPUCountWithCore(virtualCpuCount int, coresPerSocket *int) (Task, error) {

	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	// Check if VApp Children is populated
	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newcpu := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		XmlnsVmw:        types.XMLNamespaceVMW,
		VCloudHREF:      vapp.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(virtualCpuCount) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    types.ResourceTypeProcessor,
		VirtualQuantity: virtualCpuCount,
		Weight:          0,
		CoresPerSocket:  coresPerSocket,
		Link: &types.Link{
			HREF: vapp.VApp.Children.VM[0].HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.Children.VM[0].HREF)
	apiEndpoint.Path += "/virtualHardwareSection/cpu"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing CPU count: %s", newcpu)
}

func (vapp *VApp) ChangeStorageProfile(name string) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	if vapp.VApp.Children == nil || len(vapp.VApp.Children.VM) == 0 {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	vdc, err := vapp.getParentVDC()
	if err != nil {
		return Task{}, fmt.Errorf("error retrieving parent VDC for vApp %s", vapp.VApp.Name)
	}
	storageProfileRef, err := vdc.FindStorageProfileReference(name)
	if err != nil {
		return Task{}, fmt.Errorf("error retrieving storage profile %s for vApp %s", name, vapp.VApp.Name)
	}

	newProfile := &types.VM{
		Name:           vapp.VApp.Children.VM[0].Name,
		StorageProfile: &storageProfileRef,
		Xmlns:          types.XMLNamespaceVCloud,
	}

	// Return the task
	return vapp.client.ExecuteTaskRequest(vapp.VApp.Children.VM[0].HREF, http.MethodPut,
		types.MimeVM, "error changing CPU count: %s", newProfile)
}

// Deprecated as it changes only first VM's name
func (vapp *VApp) ChangeVMName(name string) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newName := &types.VM{
		Name:  name,
		Xmlns: types.XMLNamespaceVCloud,
	}

	// Return the task
	return vapp.client.ExecuteTaskRequest(vapp.VApp.Children.VM[0].HREF, http.MethodPut,
		types.MimeVM, "error changing VM name: %s", newName)
}

// SetOvf sets guest properties for the first child VM in vApp
//
// Deprecated: Use vm.SetProductSectionList()
func (vapp *VApp) SetOvf(parameters map[string]string) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	if vapp.VApp.Children.VM[0].ProductSection == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children with ProductSection, aborting customization")
	}

	for key, value := range parameters {
		for _, ovf_value := range vapp.VApp.Children.VM[0].ProductSection.Property {
			if ovf_value.Key == key {
				ovf_value.Value = &types.Value{Value: value}
				break
			}
		}
	}

	ovf := &types.ProductSectionList{
		Xmlns:          types.XMLNamespaceVCloud,
		Ovf:            types.XMLNamespaceOVF,
		ProductSection: vapp.VApp.Children.VM[0].ProductSection,
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.Children.VM[0].HREF)
	apiEndpoint.Path += "/productSections"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeProductSection, "error setting ovf: %s", ovf)
}

func (vapp *VApp) ChangeNetworkConfig(networks []map[string]interface{}, ip string) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	networksection, err := vapp.GetNetworkConnectionSection()
	if err != nil {
		return Task{}, err
	}

	for index, network := range networks {
		// Determine what type of address is requested for the vApp
		ipAllocationMode := types.IPAllocationModeNone
		ipAddress := "Any"

		// TODO: Review current behaviour of using DHCP when left blank
		if ip == "" || ip == "dhcp" || network["ip"] == "dhcp" {
			ipAllocationMode = types.IPAllocationModeDHCP
		} else if ip == "allocated" || network["ip"] == "allocated" {
			ipAllocationMode = types.IPAllocationModePool
		} else if ip == "none" || network["ip"] == "none" {
			ipAllocationMode = types.IPAllocationModeNone
		} else if ip != "" || network["ip"] != "" {
			ipAllocationMode = types.IPAllocationModeManual
			// TODO: Check a valid IP has been given
			ipAddress = ip
		}

		util.Logger.Printf("[DEBUG] Function ChangeNetworkConfig() for %s invoked", network["orgnetwork"])

		networksection.Xmlns = types.XMLNamespaceVCloud
		networksection.Ovf = types.XMLNamespaceOVF
		networksection.Info = "Specifies the available VM network connections"

		networksection.NetworkConnection[index].NeedsCustomization = true
		networksection.NetworkConnection[index].IPAddress = ipAddress
		networksection.NetworkConnection[index].IPAddressAllocationMode = ipAllocationMode
		networksection.NetworkConnection[index].MACAddress = ""

		if network["is_primary"] == true {
			networksection.PrimaryNetworkConnectionIndex = index
		}

	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.Children.VM[0].HREF)
	apiEndpoint.Path += "/networkConnectionSection/"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeNetworkConnectionSection, "error changing network config: %s", networksection)
}

// Deprecated as it changes only first VM's memory
func (vapp *VApp) ChangeMemorySize(size int) (Task, error) {

	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %s", err)
	}

	// Check if VApp Children is populated
	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	newMem := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		VCloudHREF:      vapp.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "byte * 2^20",
		Description:     "Memory Size",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    types.ResourceTypeMemory,
		VirtualQuantity: size,
		Weight:          0,
		Link: &types.Link{
			HREF: vapp.VApp.Children.VM[0].HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.Children.VM[0].HREF)
	apiEndpoint.Path += "/virtualHardwareSection/memory"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing memory size: %s", newMem)
}

func (vapp *VApp) GetNetworkConfig() (*types.NetworkConfigSection, error) {

	networkConfig := &types.NetworkConfigSection{}

	if vapp.VApp.HREF == "" {
		return networkConfig, fmt.Errorf("cannot refresh, Object is empty")
	}

	_, err := vapp.client.ExecuteRequest(vapp.VApp.HREF+"/networkConfigSection/", http.MethodGet,
		types.MimeNetworkConfigSection, "error retrieving network config: %s", nil, networkConfig)

	// The request was successful
	return networkConfig, err
}

// AddRAWNetworkConfig adds existing VDC network to vApp
// Deprecated: in favor of vapp.AddOrgNetwork
func (vapp *VApp) AddRAWNetworkConfig(orgvdcnetworks []*types.OrgVDCNetwork) (Task, error) {

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return Task{}, fmt.Errorf("error getting vApp networks: %s", err)
	}
	networkConfigurations := vAppNetworkConfig.NetworkConfig

	for _, network := range orgvdcnetworks {
		networkConfigurations = append(networkConfigurations,
			types.VAppNetworkConfiguration{
				NetworkName: network.Name,
				Configuration: &types.NetworkConfiguration{
					ParentNetwork: &types.Reference{
						HREF: network.HREF,
					},
					FenceMode: types.FenceModeBridged,
				},
			},
		)
	}

	return updateNetworkConfigurations(vapp, networkConfigurations)
}

// Function allows to create isolated network for vApp. This is equivalent to vCD UI function - vApp network creation.
// Deprecated: in favor of vapp.CreateVappNetwork
func (vapp *VApp) AddIsolatedNetwork(newIsolatedNetworkSettings *VappNetworkSettings) (Task, error) {

	err := validateNetworkConfigSettings(newIsolatedNetworkSettings)
	if err != nil {
		return Task{}, err
	}

	// for case when range is one ip address
	if newIsolatedNetworkSettings.DhcpSettings != nil && newIsolatedNetworkSettings.DhcpSettings.IPRange != nil && newIsolatedNetworkSettings.DhcpSettings.IPRange.EndAddress == "" {
		newIsolatedNetworkSettings.DhcpSettings.IPRange.EndAddress = newIsolatedNetworkSettings.DhcpSettings.IPRange.StartAddress
	}

	// only add values if available. Won't be send to API if not provided
	var networkFeatures *types.NetworkFeatures
	if newIsolatedNetworkSettings.DhcpSettings != nil {
		networkFeatures = &types.NetworkFeatures{DhcpService: &types.DhcpService{
			IsEnabled:        newIsolatedNetworkSettings.DhcpSettings.IsEnabled,
			DefaultLeaseTime: newIsolatedNetworkSettings.DhcpSettings.DefaultLeaseTime,
			MaxLeaseTime:     newIsolatedNetworkSettings.DhcpSettings.MaxLeaseTime,
			IPRange:          newIsolatedNetworkSettings.DhcpSettings.IPRange}}
	}

	networkConfigurations := vapp.VApp.NetworkConfigSection.NetworkConfig
	networkConfigurations = append(networkConfigurations,
		types.VAppNetworkConfiguration{
			NetworkName: newIsolatedNetworkSettings.Name,
			Description: newIsolatedNetworkSettings.Description,
			Configuration: &types.NetworkConfiguration{
				FenceMode:        types.FenceModeIsolated,
				GuestVlanAllowed: newIsolatedNetworkSettings.GuestVLANAllowed,
				Features:         networkFeatures,
				IPScopes: &types.IPScopes{IPScope: []*types.IPScope{&types.IPScope{IsInherited: false, Gateway: newIsolatedNetworkSettings.Gateway,
					Netmask: newIsolatedNetworkSettings.NetMask, DNS1: newIsolatedNetworkSettings.DNS1,
					DNS2: newIsolatedNetworkSettings.DNS2, DNSSuffix: newIsolatedNetworkSettings.DNSSuffix, IsEnabled: true,
					IPRanges: &types.IPRanges{IPRange: newIsolatedNetworkSettings.StaticIPRanges}}}},
			},
			IsDeployed: false,
		})

	return updateNetworkConfigurations(vapp, networkConfigurations)

}

// CreateVappNetwork creates isolated or nat routed(connected to Org VDC network) network for vApp.
// Returns pointer to types.NetworkConfigSection or error
// If orgNetwork is nil, then isolated network created.
func (vapp *VApp) CreateVappNetwork(newNetworkSettings *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork) (*types.NetworkConfigSection, error) {
	task, err := vapp.CreateVappNetworkAsync(newNetworkSettings, orgNetwork)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	return vAppNetworkConfig, nil
}

// CreateVappNetworkAsync creates asynchronously isolated or nat routed network for vApp. Returns Task or error
// If orgNetwork is nil, then isolated network created.
func (vapp *VApp) CreateVappNetworkAsync(newNetworkSettings *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork) (Task, error) {

	err := validateNetworkConfigSettings(newNetworkSettings)
	if err != nil {
		return Task{}, err
	}

	// for case when range is one ip address
	if newNetworkSettings.DhcpSettings != nil && newNetworkSettings.DhcpSettings.IPRange != nil && newNetworkSettings.DhcpSettings.IPRange.EndAddress == "" {
		newNetworkSettings.DhcpSettings.IPRange.EndAddress = newNetworkSettings.DhcpSettings.IPRange.StartAddress
	}

	// only add values if available. Won't be send to API if not provided
	var networkFeatures *types.NetworkFeatures
	if newNetworkSettings.DhcpSettings != nil {
		networkFeatures = &types.NetworkFeatures{DhcpService: &types.DhcpService{
			IsEnabled:        newNetworkSettings.DhcpSettings.IsEnabled,
			DefaultLeaseTime: newNetworkSettings.DhcpSettings.DefaultLeaseTime,
			MaxLeaseTime:     newNetworkSettings.DhcpSettings.MaxLeaseTime,
			IPRange:          newNetworkSettings.DhcpSettings.IPRange},
		}
	}

	networkConfigurations := vapp.VApp.NetworkConfigSection.NetworkConfig
	vappConfiguration := types.VAppNetworkConfiguration{
		NetworkName: newNetworkSettings.Name,
		Description: newNetworkSettings.Description,
		Configuration: &types.NetworkConfiguration{
			FenceMode:        types.FenceModeIsolated,
			GuestVlanAllowed: newNetworkSettings.GuestVLANAllowed,
			Features:         networkFeatures,
			IPScopes: &types.IPScopes{IPScope: []*types.IPScope{&types.IPScope{IsInherited: false, Gateway: newNetworkSettings.Gateway,
				Netmask: newNetworkSettings.NetMask, DNS1: newNetworkSettings.DNS1,
				DNS2: newNetworkSettings.DNS2, DNSSuffix: newNetworkSettings.DNSSuffix, IsEnabled: true,
				IPRanges: &types.IPRanges{IPRange: newNetworkSettings.StaticIPRanges}}}},
			RetainNetInfoAcrossDeployments: newNetworkSettings.RetainIpMacEnabled,
		},
		IsDeployed: false,
	}
	if orgNetwork != nil {
		vappConfiguration.Configuration.ParentNetwork = &types.Reference{
			HREF: orgNetwork.HREF,
		}
		vappConfiguration.Configuration.FenceMode = types.FenceModeNAT
	}

	networkConfigurations = append(networkConfigurations,
		vappConfiguration)

	return updateNetworkConfigurations(vapp, networkConfigurations)
}

// AddOrgNetwork adds Org VDC network as vApp network.
// Returns pointer to types.NetworkConfigSection or error
func (vapp *VApp) AddOrgNetwork(newNetworkSettings *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork, isFenced bool) (*types.NetworkConfigSection, error) {
	task, err := vapp.AddOrgNetworkAsync(newNetworkSettings, orgNetwork, isFenced)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	return vAppNetworkConfig, nil
}

// AddOrgNetworkAsync adds asynchronously Org VDC network as vApp network. Returns Task or error
func (vapp *VApp) AddOrgNetworkAsync(newNetworkSettings *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork, isFenced bool) (Task, error) {

	if orgNetwork == nil {
		return Task{}, errors.New("org VDC network is missing")
	}

	fenceMode := types.FenceModeBridged
	if isFenced {
		fenceMode = types.FenceModeNAT
	}

	networkConfigurations := vapp.VApp.NetworkConfigSection.NetworkConfig
	vappConfiguration := types.VAppNetworkConfiguration{
		NetworkName: orgNetwork.Name,
		Configuration: &types.NetworkConfiguration{
			FenceMode: fenceMode,
			ParentNetwork: &types.Reference{
				HREF: orgNetwork.HREF,
			},
			RetainNetInfoAcrossDeployments: newNetworkSettings.RetainIpMacEnabled,
		},
		IsDeployed: false,
	}
	networkConfigurations = append(networkConfigurations,
		vappConfiguration)

	return updateNetworkConfigurations(vapp, networkConfigurations)

}

// UpdateNetwork updates vApp networks (isolated or connected to Org VDC network)
// Returns pointer to types.NetworkConfigSection or error
func (vapp *VApp) UpdateNetwork(newNetworkSettings *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork) (*types.NetworkConfigSection, error) {
	task, err := vapp.UpdateNetworkAsync(newNetworkSettings, orgNetwork)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	return vAppNetworkConfig, nil
}

// UpdateNetworkAsync asynchronously updates vApp networks (isolated or connected to Org VDC network).
// Returns task or error
func (vapp *VApp) UpdateNetworkAsync(networkSettingsToUpdate *VappNetworkSettings, orgNetwork *types.OrgVDCNetwork) (Task, error) {
	util.Logger.Printf("[TRACE] UpdateNetworkAsync with values: %#v and connect to org network: %#v", networkSettingsToUpdate, orgNetwork)
	currentNetworkConfiguration, err := vapp.GetNetworkConfig()
	if err != nil {
		return Task{}, err
	}
	var networkToUpdate types.VAppNetworkConfiguration
	var networkToUpdateIndex int
	for index, networkConfig := range currentNetworkConfiguration.NetworkConfig {
		if networkConfig.Link != nil {
			uuid, err := GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return Task{}, err
			}
			if uuid == extractUuid(networkSettingsToUpdate.ID) {
				networkToUpdate = networkConfig
				networkToUpdateIndex = index
				break
			}
		}
	}

	if networkToUpdate == (types.VAppNetworkConfiguration{}) {
		return Task{}, fmt.Errorf("not found network to update with Id %s", networkSettingsToUpdate.ID)
	}
	networkToUpdate.Configuration.RetainNetInfoAcrossDeployments = networkSettingsToUpdate.RetainIpMacEnabled
	// new network to connect
	if networkToUpdate.Configuration.ParentNetwork == nil && orgNetwork != nil {
		networkToUpdate.Configuration.FenceMode = types.FenceModeNAT
		networkToUpdate.Configuration.ParentNetwork = &types.Reference{HREF: orgNetwork.HREF}
	}
	// change network to connect
	if networkToUpdate.Configuration.ParentNetwork != nil && orgNetwork != nil && networkToUpdate.Configuration.ParentNetwork.HREF != orgNetwork.HREF {
		networkToUpdate.Configuration.ParentNetwork = &types.Reference{HREF: orgNetwork.HREF}
	}
	// remove network to connect
	if orgNetwork == nil {
		networkToUpdate.Configuration.FenceMode = types.FenceModeIsolated
		networkToUpdate.Configuration.ParentNetwork = nil
	}
	networkToUpdate.Description = networkSettingsToUpdate.Description
	networkToUpdate.NetworkName = networkSettingsToUpdate.Name
	networkToUpdate.Configuration.GuestVlanAllowed = networkSettingsToUpdate.GuestVLANAllowed
	networkToUpdate.Configuration.IPScopes.IPScope[0].Gateway = networkSettingsToUpdate.Gateway
	networkToUpdate.Configuration.IPScopes.IPScope[0].Netmask = networkSettingsToUpdate.NetMask
	networkToUpdate.Configuration.IPScopes.IPScope[0].DNS1 = networkSettingsToUpdate.DNS1
	networkToUpdate.Configuration.IPScopes.IPScope[0].DNS2 = networkSettingsToUpdate.DNS2
	networkToUpdate.Configuration.IPScopes.IPScope[0].DNSSuffix = networkSettingsToUpdate.DNSSuffix
	networkToUpdate.Configuration.IPScopes.IPScope[0].IPRanges = &types.IPRanges{IPRange: networkSettingsToUpdate.StaticIPRanges}

	// for case when range is one ip address
	if networkSettingsToUpdate.DhcpSettings != nil && networkSettingsToUpdate.DhcpSettings.IPRange != nil && networkSettingsToUpdate.DhcpSettings.IPRange.EndAddress == "" {
		networkSettingsToUpdate.DhcpSettings.IPRange.EndAddress = networkSettingsToUpdate.DhcpSettings.IPRange.StartAddress
	}

	// remove DHCP config
	if networkSettingsToUpdate.DhcpSettings == nil {
		networkToUpdate.Configuration.Features.DhcpService = nil
	}

	// create DHCP config
	if networkSettingsToUpdate.DhcpSettings != nil && networkToUpdate.Configuration.Features.DhcpService == nil {
		networkToUpdate.Configuration.Features.DhcpService = &types.DhcpService{
			IsEnabled:        networkSettingsToUpdate.DhcpSettings.IsEnabled,
			DefaultLeaseTime: networkSettingsToUpdate.DhcpSettings.DefaultLeaseTime,
			MaxLeaseTime:     networkSettingsToUpdate.DhcpSettings.MaxLeaseTime,
			IPRange:          networkSettingsToUpdate.DhcpSettings.IPRange}
	}

	// update DHCP config
	if networkSettingsToUpdate.DhcpSettings != nil && networkToUpdate.Configuration.Features.DhcpService != nil {
		networkToUpdate.Configuration.Features.DhcpService.IsEnabled = networkSettingsToUpdate.DhcpSettings.IsEnabled
		networkToUpdate.Configuration.Features.DhcpService.DefaultLeaseTime = networkSettingsToUpdate.DhcpSettings.DefaultLeaseTime
		networkToUpdate.Configuration.Features.DhcpService.MaxLeaseTime = networkSettingsToUpdate.DhcpSettings.MaxLeaseTime
		networkToUpdate.Configuration.Features.DhcpService.IPRange = networkSettingsToUpdate.DhcpSettings.IPRange
	}

	currentNetworkConfiguration.NetworkConfig[networkToUpdateIndex] = networkToUpdate

	return updateNetworkConfigurations(vapp, currentNetworkConfiguration.NetworkConfig)
}

// UpdateOrgNetwork updates Org VDC network which is part of a vApp
// Returns pointer to types.NetworkConfigSection or error
func (vapp *VApp) UpdateOrgNetwork(newNetworkSettings *VappNetworkSettings, isFenced bool) (*types.NetworkConfigSection, error) {
	task, err := vapp.UpdateOrgNetworkAsync(newNetworkSettings, isFenced)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	return vAppNetworkConfig, nil
}

// UpdateOrgNetworkAsync asynchronously updates Org VDC network which is part of a vApp
// Returns task or error
func (vapp *VApp) UpdateOrgNetworkAsync(networkSettingsToUpdate *VappNetworkSettings, isFenced bool) (Task, error) {
	util.Logger.Printf("[TRACE] UpdateOrgNetworkAsync with values: %#v ", networkSettingsToUpdate)
	currentNetworkConfiguration, err := vapp.GetNetworkConfig()
	if err != nil {
		return Task{}, err
	}
	var networkToUpdate types.VAppNetworkConfiguration
	var networkToUpdateIndex int

	for index, networkConfig := range currentNetworkConfiguration.NetworkConfig {
		if networkConfig.Link != nil {
			uuid, err := GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return Task{}, err
			}

			if uuid == extractUuid(networkSettingsToUpdate.ID) {
				networkToUpdate = networkConfig
				networkToUpdateIndex = index
				break
			}
		}
	}

	if networkToUpdate == (types.VAppNetworkConfiguration{}) {
		return Task{}, fmt.Errorf("not found network to update with Id %s", networkSettingsToUpdate.ID)
	}

	fenceMode := types.FenceModeBridged
	if isFenced {
		fenceMode = types.FenceModeNAT
	}

	networkToUpdate.Configuration.RetainNetInfoAcrossDeployments = networkSettingsToUpdate.RetainIpMacEnabled
	networkToUpdate.Configuration.FenceMode = fenceMode

	currentNetworkConfiguration.NetworkConfig[networkToUpdateIndex] = networkToUpdate

	return updateNetworkConfigurations(vapp, currentNetworkConfiguration.NetworkConfig)
}

func validateNetworkConfigSettings(networkSettings *VappNetworkSettings) error {
	if networkSettings.Name == "" {
		return errors.New("network name is missing")
	}

	if networkSettings.Gateway == "" {
		return errors.New("network gateway IP is missing")
	}

	if networkSettings.NetMask == "" {
		return errors.New("network mask config is missing")
	}

	if networkSettings.NetMask == "" {
		return errors.New("network mask config is missing")
	}

	if networkSettings.DhcpSettings != nil && networkSettings.DhcpSettings.IPRange == nil {
		return errors.New("network DHCP ip range config is missing")
	}

	if networkSettings.DhcpSettings != nil && networkSettings.DhcpSettings.IPRange.StartAddress == "" {
		return errors.New("network DHCP ip range start address is missing")
	}

	return nil
}

// RemoveNetwork removes any network (be it isolated or connected to an Org Network) from vApp
// Returns pointer to types.NetworkConfigSection or error
func (vapp *VApp) RemoveNetwork(identifier string) (*types.NetworkConfigSection, error) {
	task, err := vapp.RemoveNetworkAsync(identifier)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	return vAppNetworkConfig, nil
}

// RemoveNetworkAsync asynchronously removes any network (be it isolated or connected to an Org Network) from vApp
// Accepts network ID or name
func (vapp *VApp) RemoveNetworkAsync(identifier string) (Task, error) {

	if identifier == "" {
		return Task{}, fmt.Errorf("network ID/name can't be empty")
	}

	networkConfigurations := vapp.VApp.NetworkConfigSection.NetworkConfig
	for _, networkConfig := range networkConfigurations {
		networkId, err := GetUuidFromHref(networkConfig.Link.HREF, false)
		if err != nil {
			return Task{}, fmt.Errorf("unable to get network ID from HREF: %s", err)
		}
		if networkId == extractUuid(identifier) || networkConfig.NetworkName == identifier {
			deleteUrl := vapp.client.VCDHREF.String() + "/network/" + networkId
			errMessage := fmt.Sprintf("detaching vApp network %s (id '%s'): %%s", networkConfig.NetworkName, networkId)
			task, err := vapp.client.ExecuteTaskRequest(deleteUrl, http.MethodDelete, types.AnyXMLMime, errMessage, nil)
			if err != nil {
				return Task{}, err
			}

			return task, nil
		}
	}

	return Task{}, fmt.Errorf("network to remove %s, wasn't found", identifier)

}

// Removes vApp isolated network
// Deprecated: in favor vapp.RemoveNetwork
func (vapp *VApp) RemoveIsolatedNetwork(networkName string) (Task, error) {

	if networkName == "" {
		return Task{}, fmt.Errorf("network name can't be empty")
	}

	networkConfigurations := vapp.VApp.NetworkConfigSection.NetworkConfig
	isNetworkFound := false
	for index, networkConfig := range networkConfigurations {
		if networkConfig.NetworkName == networkName {
			isNetworkFound = true
			networkConfigurations = append(networkConfigurations[:index], networkConfigurations[index+1:]...)
		}
	}

	if !isNetworkFound {
		return Task{}, fmt.Errorf("network to remove %s, wasn't found", networkName)
	}

	return updateNetworkConfigurations(vapp, networkConfigurations)
}

// Function allows to update vApp network configuration. This works for updating, deleting and adding.
// Network configuration has to be full with new, changed elements and unchanged.
// https://opengrok.eng.vmware.com/source/xref/cloud-sp-main.perforce-shark.1700/sp-main/dev-integration/system-tests/SystemTests/src/main/java/com/vmware/cloud/systemtests/util/VAppNetworkUtils.java#createVAppNetwork
// http://pubs.vmware.com/vcloud-api-1-5/wwhelp/wwhimpl/js/html/wwhelp.htm#href=api_prog/GUID-92622A15-E588-4FA1-92DA-A22A4757F2A0.html#1_14_12_10_1
func updateNetworkConfigurations(vapp *VApp, networkConfigurations []types.VAppNetworkConfiguration) (Task, error) {
	util.Logger.Printf("[TRACE] updateNetworkConfigurations for vAPP: %#v and network config: %#v", vapp, networkConfigurations)
	networkConfig := &types.NetworkConfigSection{
		Info:          "Configuration parameters for logical networks",
		Ovf:           types.XMLNamespaceOVF,
		Type:          types.MimeNetworkConfigSection,
		Xmlns:         types.XMLNamespaceVCloud,
		NetworkConfig: networkConfigurations,
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/networkConfigSection/"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeNetworkConfigSection, "error updating vApp Network: %s", networkConfig)
}

// RemoveAllNetworks detaches all networks from vApp
func (vapp *VApp) RemoveAllNetworks() (Task, error) {
	return updateNetworkConfigurations(vapp, []types.VAppNetworkConfiguration{})
}

// SetProductSectionList sets product section for a vApp. It allows to change vApp guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vapp *VApp) SetProductSectionList(productSection *types.ProductSectionList) (*types.ProductSectionList, error) {
	err := setProductSectionList(vapp.client, vapp.VApp.HREF, productSection)
	if err != nil {
		return nil, fmt.Errorf("unable to set vApp product section: %s", err)
	}

	return vapp.GetProductSectionList()
}

// GetProductSectionList retrieves product section for a vApp. It allows to read vApp guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vapp *VApp) GetProductSectionList() (*types.ProductSectionList, error) {
	return getProductSectionList(vapp.client, vapp.VApp.HREF)
}

// GetVMByHref returns a VM reference by running a vCD API call
// If no valid VM is found, it returns a nil VM reference and an error
// Note that the pointer receiver here is a Client instead of a VApp, because
// there are cases where we know the VM HREF but not which VApp it belongs to.
func (client *Client) GetVMByHref(vmHref string) (*VM, error) {

	newVm := NewVM(client)

	_, err := client.ExecuteRequestWithApiVersion(vmHref, http.MethodGet,
		"", "error retrieving vm: %s", nil, newVm.VM, client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))

	if err != nil {

		return nil, err
	}

	return newVm, nil
}

// GetVMByName returns a VM reference if the VM name matches an existing one.
// If no valid VM is found, it returns a nil VM reference and an error
func (vapp *VApp) GetVMByName(vmName string, refresh bool) (*VM, error) {
	if refresh {
		err := vapp.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing vapp: %s", err)
		}
	}

	//vApp Might Not Have Any VMs
	if vapp.VApp.Children == nil {
		return nil, ErrorEntityNotFound
	}

	util.Logger.Printf("[TRACE] Looking for VM: %s", vmName)
	for _, child := range vapp.VApp.Children.VM {

		util.Logger.Printf("[TRACE] Looking at: %s", child.Name)
		if child.Name == vmName {
			return vapp.client.GetVMByHref(child.HREF)
		}

	}
	util.Logger.Printf("[TRACE] Couldn't find VM: %s", vmName)
	return nil, ErrorEntityNotFound
}

// GetVMById returns a VM reference if the VM ID matches an existing one.
// If no valid VM is found, it returns a nil VM reference and an error
func (vapp *VApp) GetVMById(id string, refresh bool) (*VM, error) {
	if refresh {
		err := vapp.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing vapp: %s", err)
		}
	}

	//vApp Might Not Have Any VMs
	if vapp.VApp.Children == nil {
		return nil, ErrorEntityNotFound
	}

	util.Logger.Printf("[TRACE] Looking for VM: %s", id)
	for _, child := range vapp.VApp.Children.VM {

		util.Logger.Printf("[TRACE] Looking at: %s", child.Name)
		if equalIds(id, child.ID, child.HREF) {
			return vapp.client.GetVMByHref(child.HREF)
		}
	}
	util.Logger.Printf("[TRACE] Couldn't find VM: %s", id)
	return nil, ErrorEntityNotFound
}

// GetVMByNameOrId returns a VM reference if either the VM name or ID matches an existing one.
// If no valid VM is found, it returns a nil VM reference and an error
func (vapp *VApp) GetVMByNameOrId(identifier string, refresh bool) (*VM, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vapp.GetVMByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return vapp.GetVMById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*VM), err
}
