/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// OrgVDCNetwork an org vdc network client
type OrgVDCNetwork struct {
	OrgVDCNetwork *types.OrgVDCNetwork
	client        *Client
}

var reErrorBusy2 = regexp.MustCompile("is busy, cannot proceed with the operation.$")

// NewOrgVDCNetwork creates an org vdc network client
func NewOrgVDCNetwork(cli *Client) *OrgVDCNetwork {
	return &OrgVDCNetwork{
		OrgVDCNetwork: new(types.OrgVDCNetwork),
		client:        cli,
	}
}

func (orgVdcNet *OrgVDCNetwork) Refresh() error {
	if orgVdcNet.OrgVDCNetwork.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	refreshUrl := orgVdcNet.OrgVDCNetwork.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	orgVdcNet.OrgVDCNetwork = &types.OrgVDCNetwork{}

	_, err := orgVdcNet.client.ExecuteRequest(refreshUrl, http.MethodGet,
		"", "error retrieving vDC network: %s", nil, orgVdcNet.OrgVDCNetwork)

	return err
}

// Delete a network. Fails if the network is busy.
// Returns a task to monitor the deletion.
func (orgVdcNet *OrgVDCNetwork) Delete() (Task, error) {
	err := orgVdcNet.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing network: %s", err)
	}
	pathArr := strings.Split(orgVdcNet.OrgVDCNetwork.HREF, "/")
	apiEndpoint, _ := url.ParseRequestURI(orgVdcNet.OrgVDCNetwork.HREF)
	apiEndpoint.Path = "/api/admin/network/" + pathArr[len(pathArr)-1]

	var resp *http.Response
	for {
		req := orgVdcNet.client.NewRequest(map[string]string{}, http.MethodDelete, *apiEndpoint, nil)
		resp, err = checkResp(orgVdcNet.client.Http.Do(req))
		if err != nil {
			if reErrorBusy2.MatchString(err.Error()) {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error deleting Network: %s", err)
		}
		break
	}

	task := NewTask(orgVdcNet.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

// Looks for an Org Vdc network and, if found, will delete it.
func RemoveOrgVdcNetworkIfExists(vdc Vdc, networkName string) error {
	network, err := vdc.GetOrgVdcNetworkByName(networkName, true)

	if IsNotFound(err) {
		// Network not found. No action needed
		return nil
	}
	if err != nil {
		// Some other error happened during retrieval. We pass it along
		return err
	}
	// The network was found. We attempt deletion
	task, err := network.Delete()
	if err != nil {
		return fmt.Errorf("error deleting network [phase 1] %s", networkName)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error deleting network [task] %s", networkName)
	}
	return nil
}

// A wrapper call around CreateOrgVDCNetwork.
// Creates a network and then uses the associated task to monitor its configuration
func (vdc *Vdc) CreateOrgVDCNetworkWait(networkConfig *types.OrgVDCNetwork) error {

	task, err := vdc.CreateOrgVDCNetwork(networkConfig)
	if err != nil {
		return fmt.Errorf("error creating the network: %s", err)
	}
	if task == (Task{}) {
		return fmt.Errorf("NULL task retrieved after network creation")

	}
	err = task.WaitTaskCompletion()
	// err = task.WaitInspectTaskCompletion(InspectTask, 10)
	if err != nil {
		return fmt.Errorf("error performing task: %s", err)
	}
	return nil
}

// Fine tuning network creation function.
// Return an error (the result of the network creation) and a task (used to monitor
// the network configuration)
// This function can create any type of Org Vdc network. The exact type is determined by
// the combination of properties given with the network configuration structure.
func (vdc *Vdc) CreateOrgVDCNetwork(networkConfig *types.OrgVDCNetwork) (Task, error) {
	for _, av := range vdc.Vdc.Link {
		if av.Rel == "add" && av.Type == "application/vnd.vmware.vcloud.orgVdcNetwork+xml" {
			createUrl, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return Task{}, fmt.Errorf("error decoding vdc response: %s", err)
			}

			networkConfig.Xmlns = types.XMLNamespaceVCloud

			output, err := xml.MarshalIndent(networkConfig, "  ", "    ")
			if err != nil {
				return Task{}, fmt.Errorf("error marshaling OrgVDCNetwork compose: %s", err)
			}

			var resp *http.Response
			for {
				b := bytes.NewBufferString(xml.Header + string(output))
				util.Logger.Printf("[DEBUG] VCD Client configuration: %s", b)
				req := vdc.client.NewRequest(map[string]string{}, http.MethodPost, *createUrl, b)
				req.Header.Add("Content-Type", av.Type)
				resp, err = checkResp(vdc.client.Http.Do(req))
				if err != nil {
					if reErrorBusy2.MatchString(err.Error()) {
						time.Sleep(3 * time.Second)
						continue
					}
					return Task{}, fmt.Errorf("error instantiating a new OrgVDCNetwork: %s", err)
				}
				break
			}
			orgVDCNetwork := NewOrgVDCNetwork(vdc.client)
			if err = decodeBody(resp, orgVDCNetwork.OrgVDCNetwork); err != nil {
				return Task{}, fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
			}
			activeTasks := 0
			// Makes sure that there is only one active task for this network.
			for _, taskItem := range orgVDCNetwork.OrgVDCNetwork.Tasks.Task {
				if taskItem.HREF != "" {
					activeTasks += 1
					if os.Getenv("GOVCD_DEBUG") != "" {
						fmt.Printf("task %s (%s) is active\n", taskItem.HREF, taskItem.Status)
					}
				}
			}
			if activeTasks > 1 {
				// By my understanding of the implementation, there should not be more than one task for this operation.
				// If there is, we will need to change the logic of this function, as we can only return one task. (GM)
				return Task{}, fmt.Errorf("found %d active tasks instead of one", activeTasks)
			}
			for _, taskItem := range orgVDCNetwork.OrgVDCNetwork.Tasks.Task {
				return Task{taskItem, vdc.client}, nil
			}
			return Task{}, fmt.Errorf("[%s] no suitable task found", util.CurrentFuncName())
		}
	}
	return Task{}, fmt.Errorf("network creation failed: no operational link found")
}

// GetNetworkList returns a list of networks for the VDC
func (vdc *Vdc) GetNetworkList() ([]*types.QueryResultOrgVdcNetworkRecordType, error) {
	// Find the list of networks with the wanted name
	result, err := vdc.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "orgVdcNetwork",
		"filter": fmt.Sprintf("vdc==%s", url.QueryEscape(vdc.Vdc.ID)),
	})
	if err != nil {
		return nil, fmt.Errorf("[findEdgeGatewayConnection] error returning the list of networks for VDC: %s", err)
	}
	return result.Results.OrgVdcNetworkRecord, nil
}

// FindEdgeGatewayNameByNetwork searches the VDC for a connection between an edge gateway and a given network.
// On success, returns the name of the edge gateway
func (vdc *Vdc) FindEdgeGatewayNameByNetwork(networkName string) (string, error) {

	// Find the list of networks with the wanted name
	result, err := vdc.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "orgVdcNetwork",
		"filter": fmt.Sprintf("name==%s;vdc==%s", url.QueryEscape(networkName), url.QueryEscape(vdc.Vdc.ID)),
	})
	if err != nil {
		return "", fmt.Errorf("[findEdgeGatewayConnection] error returning the list of networks for VDC: %s", err)
	}
	netList := result.Results.OrgVdcNetworkRecord

	for _, net := range netList {
		if net.Name == networkName {
			// linkType is not well documented, but empiric tests show that:
			// 0 = direct
			// 1 = routed
			// 2 = isolated
			if net.ConnectedTo != "" && net.LinkType == 1 { // We only want routed networks
				return net.ConnectedTo, nil
			}
		}
	}
	return "", fmt.Errorf("no edge gateway connection found")
}
