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
	Name             string
	Gateway          string
	NetMask          string
	DNS1             string
	DNS2             string
	DNSSuffix        string
	GuestVLANAllowed *bool
	StaticIPRanges   []*types.IPRange
	DhcpSettings     *DhcpSettings
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
	return Vdc{}, fmt.Errorf("Could not find a parent Vdc")
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

// Function create vm in vApp using vApp template
// orgVdcNetworks - adds org VDC networks to be available for vApp. Can be empty.
// vappNetworkName - adds vApp network to be available for vApp. Can be empty.
// vappTemplate - vApp Template which will be used for VM creation.
// name - name for VM.
// acceptAllEulas - setting allows to automatically accept or not Eulas.
func (vapp *VApp) AddVM(orgVdcNetworks []*types.OrgVDCNetwork, vappNetworkName string, vappTemplate VAppTemplate, name string, acceptAllEulas bool) (Task, error) {

	if vappTemplate == (VAppTemplate{}) || vappTemplate.VAppTemplate == nil {
		return Task{}, fmt.Errorf("vApp Template can not be empty")
	}

	// Status 8 means The object is resolved and powered off.
	// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/94b8bd8d-74ff-4fe3-b7a4-41ae31516ed7/1b42f3b5-8b31-4279-8b3f-547f6c7c5aa8/doc/GUID-843BE3AD-5EF6-4442-B864-BCAE44A51867.html
	if vappTemplate.VAppTemplate.Status != 8 {
		return Task{}, fmt.Errorf("vApp Template shape is not ok")
	}

	vcomp := &types.ReComposeVAppParams{
		Ovf:         types.XMLNamespaceOVF,
		Xsi:         types.XMLNamespaceXSI,
		Xmlns:       types.XMLNamespaceVCloud,
		Deploy:      false,
		Name:        vapp.VApp.Name,
		PowerOn:     false,
		Description: vapp.VApp.Description,
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vappTemplate.VAppTemplate.Children.VM[0].HREF,
				Name: name,
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{
					Type:                          vappTemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.Type,
					HREF:                          vappTemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.HREF,
					Info:                          "Network config for sourced item",
					PrimaryNetworkConnectionIndex: vappTemplate.VAppTemplate.Children.VM[0].NetworkConnectionSection.PrimaryNetworkConnectionIndex,
				},
			},
		},
		AllEULAsAccepted: acceptAllEulas,
	}

	for index, orgVdcNetwork := range orgVdcNetworks {
		vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection = append(vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection,
			&types.NetworkConnection{
				Network:                 orgVdcNetwork.Name,
				NetworkConnectionIndex:  index,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
			},
		)
		vcomp.SourcedItem.NetworkAssignment = append(vcomp.SourcedItem.NetworkAssignment,
			&types.NetworkAssignment{
				InnerNetwork:     orgVdcNetwork.Name,
				ContainerNetwork: orgVdcNetwork.Name,
			},
		)
	}

	if vappNetworkName != "" {
		vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection = append(vcomp.SourcedItem.InstantiationParams.NetworkConnectionSection.NetworkConnection,
			&types.NetworkConnection{
				Network:                 vappNetworkName,
				NetworkConnectionIndex:  len(orgVdcNetworks),
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModePool,
			},
		)
		vcomp.SourcedItem.NetworkAssignment = append(vcomp.SourcedItem.NetworkAssignment,
			&types.NetworkAssignment{
				InnerNetwork:     vappNetworkName,
				ContainerNetwork: vappNetworkName,
			},
		)
	}

	apiEndpoint, _ := url.ParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/recomposeVApp"

	// Return the task
	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeRecomposeVappParams, "error instantiating a new VM: %s", vcomp)

}

func (vapp *VApp) RemoveVM(vm VM) error {

	vapp.Refresh()
	task := NewTask(vapp.client)
	if vapp.VApp.Tasks != nil {
		for _, taskItem := range vapp.VApp.Tasks.Task {
			task.Task = taskItem
			err := task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error performing task: %#v", err)
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
		return fmt.Errorf("error performing removing VM task: %#v", err)
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

func (vapp *VApp) Customize(computername, script string, changeSid bool) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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
		Enabled:             true,
		ComputerName:        computername,
		CustomizationScript: script,
		ChangeSid:           false,
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
		return "", fmt.Errorf("error refreshing vApp: %v", err)
	}
	return types.VAppStatuses[vapp.VApp.Status], nil
}

// BlockWhileStatus blocks until the status of vApp exits unwantedStatus.
// It sleeps 200 milliseconds between iterations and times out after timeOutAfterSeconds
// of seconds.
func (vapp *VApp) BlockWhileStatus(unwantedStatus string, timeOutAfterSeconds int) error {
	timeoutAfter := time.After(time.Duration(timeOutAfterSeconds) * time.Second)
	tick := time.Tick(200 * time.Millisecond)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timed out waiting for vApp to exit state %s after %d seconds",
				unwantedStatus, timeOutAfterSeconds)
		case <-tick:
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
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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

// GetMetadata() function calls private function getMetadata() with vapp.client and vapp.VApp.HREF
// which returns a *types.Metadata struct for provided vapp input.
func (vapp *VApp) GetMetadata() (*types.Metadata, error) {
	return getMetadata(vapp.client, vapp.VApp.HREF)
}

func getMetadata(client *Client, requestUri string) (*types.Metadata, error) {
	metadata := &types.Metadata{}

	_, err := client.ExecuteRequest(requestUri+"/metadata/", http.MethodGet,
		types.MimeMetaData, "error retrieving metadata: %s", nil, metadata)

	return metadata, err
}

// DeleteMetadata() function calls private function deleteMetadata() with vapp.client and vapp.VApp.HREF
// which deletes metadata depending on key provided as input from vApp.
func (vapp *VApp) DeleteMetadata(key string) (Task, error) {
	return deleteMetadata(vapp.client, key, vapp.VApp.HREF)
}

// Deletes metadata (type MetadataStringValue) from the vApp
// TODO: Support all MetadataTypedValue types with this function
func deleteMetadata(client *Client, key string, requestUri string) (Task, error) {
	apiEndpoint, _ := url.ParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodDelete,
		"", "error deleting metadata: %s", nil)
}

// AddMetadata() function calls private function addMetadata() with vapp.client and vapp.VApp.HREF
// which adds metadata key, value pair provided as input.
func (vapp *VApp) AddMetadata(key string, value string) (Task, error) {
	return addMetadata(vapp.client, key, value, vapp.VApp.HREF)
}

// Adds metadata (type MetadataStringValue) to the vApp
// TODO: Support all MetadataTypedValue types with this function
func addMetadata(client *Client, key string, value string, requestUri string) (Task, error) {
	newMetadata := &types.MetadataValue{
		Xmlns: types.XMLNamespaceVCloud,
		Xsi:   types.XMLNamespaceXSI,
		TypedValue: &types.TypedValue{
			XsiType: "MetadataStringValue",
			Value:   value,
		},
	}

	apiEndpoint, _ := url.ParseRequestURI(requestUri)
	apiEndpoint.Path += "/metadata/" + key

	// Return the task
	return client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeMetaDataValue, "error adding metadata: %s", newMetadata)
}

func (vapp *VApp) SetOvf(parameters map[string]string) (Task, error) {
	err := vapp.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %v", err)
	}

	if vapp.VApp.Children == nil {
		return Task{}, fmt.Errorf("vApp doesn't contain any children, aborting customization")
	}

	networksection, err := vapp.GetNetworkConnectionSection()

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
		return Task{}, fmt.Errorf("error refreshing vApp before running customization: %v", err)
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

// Function adds existing VDC network to vApp
func (vapp *VApp) AddRAWNetworkConfig(orgvdcnetworks []*types.OrgVDCNetwork) (Task, error) {

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return Task{}, fmt.Errorf("error getting vApp networks: %#v", err)
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
func (vapp *VApp) AddIsolatedNetwork(newIsolatedNetworkSettings *VappNetworkSettings) (Task, error) {

	err := validateNetworkConfigSettings(newIsolatedNetworkSettings)
	if err != nil {
		return Task{}, err
	}

	// for case when range is one ip address
	if newIsolatedNetworkSettings.DhcpSettings != nil && newIsolatedNetworkSettings.DhcpSettings.IPRange != nil && newIsolatedNetworkSettings.DhcpSettings.IPRange.EndAddress == "" {
		newIsolatedNetworkSettings.DhcpSettings.IPRange.EndAddress = newIsolatedNetworkSettings.DhcpSettings.IPRange.StartAddress
	}

	// explicitly check if to add data, to not send any values
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
			Configuration: &types.NetworkConfiguration{
				FenceMode:        types.FenceModeIsolated,
				GuestVlanAllowed: newIsolatedNetworkSettings.GuestVLANAllowed,
				Features:         networkFeatures,
				IPScopes: &types.IPScopes{IPScope: types.IPScope{IsInherited: false, Gateway: newIsolatedNetworkSettings.Gateway,
					Netmask: newIsolatedNetworkSettings.NetMask, DNS1: newIsolatedNetworkSettings.DNS1,
					DNS2: newIsolatedNetworkSettings.DNS2, DNSSuffix: newIsolatedNetworkSettings.DNSSuffix, IsEnabled: true,
					IPRanges: &types.IPRanges{IPRange: newIsolatedNetworkSettings.StaticIPRanges}}},
			},
			IsDeployed: false,
		})

	return updateNetworkConfigurations(vapp, networkConfigurations)

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

// Removes vApp isolated network
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
