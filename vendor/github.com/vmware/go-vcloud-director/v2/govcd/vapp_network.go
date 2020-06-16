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

// UpdateNetworkFirewallRules updates vApp networks firewall rules. It will overwrite existing ones as there is
// no 100% way to identify them separately.
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

// UpdateNetworkFirewallRulesAsync asynchronously updates vApp networks firewall rules. It will overwrite existing ones
// as there is no 100% way to identify them separately.
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

	// If API didn't return Firewall service XML part, that means vApp network isn't connected to org network or not fenced.
	// In other words there isn't firewall when you connected directly or isolated.
	if networkToUpdate.Configuration.Features.FirewallService == nil {
		return Task{}, fmt.Errorf("provided network isn't connecd to org network or isn't fenced")
	}
	networkToUpdate.Configuration.Features.FirewallService.LogDefaultAction = logDefaultAction
	networkToUpdate.Configuration.Features.FirewallService.DefaultAction = defaultAction
	networkToUpdate.Configuration.Features.FirewallService.FirewallRule = firewallRules

	// here we use `PUT /network/{id}` which allow to change vApp network.
	// But `GET /network/{id}` can return org VDC network or vApp network.
	apiEndpoint := vapp.client.VCDHREF
	apiEndpoint.Path += "/network/" + uuid

	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeVappNetwork, "error updating vApp Network firewall rules: %s", networkToUpdate)
}

// GetVappNetworkById returns a VApp network reference if the vApp network ID matches an existing one.
// If no valid VApp network is found, it returns a nil VApp network reference and an error
func (vapp *VApp) GetVappNetworkById(id string, refresh bool) (*types.VAppNetwork, error) {
	util.Logger.Printf("[TRACE] [GetVappNetworkById] getting vApp Network: %s and refresh %t", id, refresh)

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
		// break early for disconnected network interfaces. They don't have all information
		if vappNetwork.NetworkName == "none" {
			continue
		}
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

// GetVappNetworkByName returns a VAppNetwork reference if the vApp network name matches an existing one.
// If no valid vApp network is found, it returns a nil VAppNetwork reference and an error
func (vapp *VApp) GetVappNetworkByName(vappNetworkName string, refresh bool) (*types.VAppNetwork, error) {
	util.Logger.Printf("[TRACE] [GetVappNetworkByName] getting vApp Network: %s and refresh %t", vappNetworkName, refresh)
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

// UpdateNetworkNatRules updates vApp networks NAT rules.
// Returns pointer to types.VAppNetwork or error
func (vapp *VApp) UpdateNetworkNatRules(networkId string, natRules []*types.NatRule, natType, policy string) (*types.VAppNetwork, error) {
	task, err := vapp.UpdateNetworkNatRulesAsync(networkId, natRules, natType, policy)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	return vapp.GetVappNetworkById(networkId, false)
}

// UpdateNetworkNatRulesAsync asynchronously updates vApp NAT rules.
// Returns task or error
func (vapp *VApp) UpdateNetworkNatRulesAsync(networkId string, natRules []*types.NatRule, natType, policy string) (Task, error) {
	util.Logger.Printf("[TRACE] UpdateNetworkNatRulesAsync with values: id: %s and natRules: %#v", networkId, natRules)

	uuid := extractUuid(networkId)
	networkToUpdate, err := vapp.GetVappNetworkById(uuid, true)
	if err != nil {
		return Task{}, err
	}

	if networkToUpdate.Configuration.Features == nil {
		networkToUpdate.Configuration.Features = &types.NetworkFeatures{}
	}
	networkToUpdate.Xmlns = types.XMLNamespaceVCloud

	if networkToUpdate.Configuration.Features.NatService == nil {
		return Task{}, fmt.Errorf("provided network isn't connected to org network or isn't fenced")
	}
	networkToUpdate.Configuration.Features.NatService.NatType = natType
	networkToUpdate.Configuration.Features.NatService.Policy = policy
	networkToUpdate.Configuration.Features.NatService.NatRule = natRules

	// here we use `PUT /network/{id}` which allow to change vApp network.
	// But `GET /network/{id}` can return org VDC network or vApp network.
	apiEndpoint := vapp.client.VCDHREF
	apiEndpoint.Path += "/network/" + uuid

	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeVappNetwork, "error updating vApp Network NAT rules: %s", networkToUpdate)
}

// RemoveAllNetworkNatRules removes all NAT rules from a vApp network
// Returns error
func (vapp *VApp) RemoveAllNetworkNatRules(networkId string) error {
	task, err := vapp.UpdateNetworkNatRulesAsync(networkId, []*types.NatRule{}, "ipTranslation", "allowTraffic")
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}
	return nil
}

// RemoveAllNetworkFirewallRules removes all network all firewall rules from a vApp network.
// Returns error
func (vapp *VApp) RemoveAllNetworkFirewallRules(networkId string) error {
	task, err := vapp.UpdateNetworkFirewallRulesAsync(networkId, []*types.FirewallRule{}, "allow", false)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}
	return nil
}
