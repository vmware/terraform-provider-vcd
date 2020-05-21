/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/http"
)

// UpdateNetworkFirewallRules updates vApp networks firewall rules.
// Returns pointer to types.VAppNetwork or error
func (vapp *VApp) UpdateNetworkFirewallRules(networkId string, firewallRules []*types.FirewallRule, defaultAction string, logDefaultAction bool) (*types.VAppNetwork, error) {
	task, err := vapp.UpdateNetworkFirewallRulesAsync(networkId, firewallRules, defaultAction, logDefaultAction)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	return vapp.GetVappNetworkById(networkId, false)
}

// UpdateNetworkFirewallRulesAsync asynchronously updates vApp networks firewall rules.
// Returns task or error
func (vapp *VApp) UpdateNetworkFirewallRulesAsync(networkId string, firewallRules []*types.FirewallRule, defaultAction string, logDefaultAction bool) (Task, error) {
	util.Logger.Printf("[TRACE] UpdateNetworkFirewallRulesAsync with values: id: %s and firewallServiceConfiguration: %#v", networkId, firewallRules)
	uuid := extractUuid(networkId)
	networkToUpdate, err := vapp.GetVappNetworkById(uuid, true)
	if err != nil {
		return Task{}, err
	}

	if networkToUpdate.Configuration.Features == nil {
		networkToUpdate.Configuration.Features = &types.NetworkFeatures{}
	}
	networkToUpdate.Xmlns = types.XMLNamespaceVCloud

	if networkToUpdate.Configuration.Features.FirewallService == nil {
		return Task{}, fmt.Errorf("provided network isn't connecd to org network or isn't fenced")
	}
	networkToUpdate.Configuration.Features.FirewallService.LogDefaultAction = logDefaultAction
	networkToUpdate.Configuration.Features.FirewallService.DefaultAction = defaultAction
	networkToUpdate.Configuration.Features.FirewallService.FirewallRule = firewallRules

	apiEndpoint := vapp.client.VCDHREF
	apiEndpoint.Path += "/network/" + uuid

	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeVappNetwork, "error updating vApp Network firewall rules: %s", networkToUpdate)
}

// GetVappNetworkById returns a VApp network reference if the vApp network ID matches an existing one.
// If no valid VApp network is found, it returns a nil VApp network reference and an error
func (vapp *VApp) GetVappNetworkById(id string, refresh bool) (*types.VAppNetwork, error) {
	util.Logger.Printf("[TRACE] [GetVappNetworkById] getting vApp Network: %s", id)

	if refresh {
		err := vapp.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing vapp: %s", err)
		}
	}

	//vApp Might Not Have Any networks
	if vapp.VApp.NetworkConfigSection == nil || len(vapp.VApp.NetworkConfigSection.NetworkConfig) == 0 {
		return nil, ErrorEntityNotFound
	}

	util.Logger.Printf("[TRACE] Looking for networks: %s --- %d", id, len(vapp.VApp.NetworkConfigSection.NetworkConfig))
	for _, vappNetwork := range vapp.VApp.NetworkConfigSection.NetworkConfig {
		util.Logger.Printf("[TRACE] Looking at: %s", vappNetwork.Link.HREF)
		if equalIds(id, vappNetwork.ID, vappNetwork.Link.HREF) {
			vappNetwork := &types.VAppNetwork{}

			apiEndpoint := vapp.client.VCDHREF
			apiEndpoint.Path += "/network/" + extractUuid(id)

			_, err := vapp.client.ExecuteRequest(apiEndpoint.String(), http.MethodGet,
				types.MimeVappNetwork, "error getting vApp network: %s", nil, vappNetwork)
			if err != nil {
				return nil, err
			}
			return vappNetwork, nil
		}
	}
	util.Logger.Printf("[TRACE] GetVappNetworkById returns not found entity")
	return nil, ErrorEntityNotFound
}

// GetVMByName returns a VM reference if the VM name matches an existing one.
// If no valid VM is found, it returns a nil VM reference and an error
func (vapp *VApp) GetVappNetworkByName(vappNetworkName string, refresh bool) (*types.VAppNetwork, error) {
	if refresh {
		err := vapp.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing vapp: %s", err)
		}
	}

	//vApp Might Not Have Any networks
	if vapp.VApp.NetworkConfigSection == nil || len(vapp.VApp.NetworkConfigSection.NetworkConfig) == 0 {
		return nil, ErrorEntityNotFound
	}

	util.Logger.Printf("[TRACE] Looking for networks: %s", vappNetworkName)
	for _, vappNetwork := range vapp.VApp.NetworkConfigSection.NetworkConfig {

		util.Logger.Printf("[TRACE] Looking at: %s", vappNetwork.NetworkName)
		if vappNetwork.NetworkName == vappNetworkName {
			return vapp.GetVappNetworkById(extractUuid(vappNetwork.Link.HREF), refresh)
		}

	}
	util.Logger.Printf("[TRACE] Couldn't find vApp network: %s", vappNetworkName)
	return nil, ErrorEntityNotFound
}

// GetVappNetworkByNameOrId returns a types.VAppNetwork reference if either the vApp network name or ID matches an existing one.
// If no valid vApp network is found, it returns a nil types.VAppNetwork reference and an error
func (vapp *VApp) GetVappNetworkByNameOrId(identifier string, refresh bool) (*types.VAppNetwork, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vapp.GetVappNetworkByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return vapp.GetVappNetworkById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, false)
	if entity == nil {
		return nil, err
	}
	return entity.(*types.VAppNetwork), err
}
