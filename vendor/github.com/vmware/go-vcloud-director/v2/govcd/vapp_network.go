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
func (vapp *VApp) UpdateNetworkFirewallRules(networkId string, firewallRules []*types.FirewallRule, enabled bool, defaultAction string, logDefaultAction bool) (*types.VAppNetwork, error) {
	task, err := vapp.UpdateNetworkFirewallRulesAsync(networkId, firewallRules, enabled, defaultAction, logDefaultAction)
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
func (vapp *VApp) UpdateNetworkFirewallRulesAsync(networkId string, firewallRules []*types.FirewallRule, enabled bool, defaultAction string, logDefaultAction bool) (Task, error) {
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
	networkToUpdate.Configuration.Features.FirewallService.IsEnabled = enabled
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
		// Break early for empty network interfaces. They don't have all information
		if vappNetwork.NetworkName == types.NoneNetwork {
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
func (vapp *VApp) UpdateNetworkNatRules(networkId string, natRules []*types.NatRule, enabled bool, natType, policy string) (*types.VAppNetwork, error) {
	task, err := vapp.UpdateNetworkNatRulesAsync(networkId, natRules, enabled, natType, policy)
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
func (vapp *VApp) UpdateNetworkNatRulesAsync(networkId string, natRules []*types.NatRule, enabled bool, natType, policy string) (Task, error) {
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

	// if services are empty return by API, then we can deduce that network isn't connected to Org network or fenced
	if networkToUpdate.Configuration.Features.NatService == nil && networkToUpdate.Configuration.Features.FirewallService == nil {
		return Task{}, fmt.Errorf("provided network isn't connected to org network or isn't fenced")
	}
	if networkToUpdate.Configuration.Features.NatService == nil {
		networkToUpdate.Configuration.Features.NatService = &types.NatService{}
	}
	networkToUpdate.Configuration.Features.NatService.IsEnabled = enabled
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
	task, err := vapp.UpdateNetworkNatRulesAsync(networkId, []*types.NatRule{}, false, "ipTranslation", "allowTraffic")
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}
	return nil
}

// RemoveAllNetworkFirewallRules removes all network firewall rules from a vApp network.
// Returns error
func (vapp *VApp) RemoveAllNetworkFirewallRules(networkId string) error {
	networkToUpdate, err := vapp.GetVappNetworkById(networkId, true)
	if err != nil {
		return err
	}
	task, err := vapp.UpdateNetworkFirewallRulesAsync(networkId, []*types.FirewallRule{}, false,
		networkToUpdate.Configuration.Features.FirewallService.DefaultAction, networkToUpdate.Configuration.Features.FirewallService.LogDefaultAction)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}
	return nil
}

// UpdateNetworkStaticRouting updates vApp network static routes.
// Returns pointer to types.VAppNetwork or error
func (vapp *VApp) UpdateNetworkStaticRouting(networkId string, staticRoutes []*types.StaticRoute, enabled bool) (*types.VAppNetwork, error) {
	task, err := vapp.UpdateNetworkStaticRoutingAsync(networkId, staticRoutes, enabled)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	return vapp.GetVappNetworkById(networkId, false)
}

// UpdateNetworkStaticRoutingAsync asynchronously updates vApp network static routes.
// Returns task or error
func (vapp *VApp) UpdateNetworkStaticRoutingAsync(networkId string, staticRoutes []*types.StaticRoute, enabled bool) (Task, error) {
	util.Logger.Printf("[TRACE] UpdateNetworkStaticRoutingAsync with values: id: %s and staticRoutes: %#v, enable: %t", networkId, staticRoutes, enabled)

	uuid := extractUuid(networkId)
	networkToUpdate, err := vapp.GetVappNetworkById(uuid, true)
	if err != nil {
		return Task{}, err
	}

	if !IsVappNetwork(networkToUpdate.Configuration) {
		return Task{}, fmt.Errorf("network static routing can be applied only for vapp network, not vapp org network")
	}

	if networkToUpdate.Configuration.Features == nil {
		networkToUpdate.Configuration.Features = &types.NetworkFeatures{}
	}
	networkToUpdate.Xmlns = types.XMLNamespaceVCloud

	networkToUpdate.Configuration.Features.StaticRoutingService = &types.StaticRoutingService{IsEnabled: enabled, StaticRoute: staticRoutes}

	// here we use `PUT /network/{id}` which allow to change vApp network.
	// But `GET /network/{id}` can return org VDC network or vApp network.
	apiEndpoint := vapp.client.VCDHREF
	apiEndpoint.Path += "/network/" + uuid

	return vapp.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeVappNetwork, "error updating vApp Network static routes: %s", networkToUpdate)
}

// IsVappNetwork allows to identify if given network config is a vApp network and not a vApp Org network
func IsVappNetwork(networkConfig *types.NetworkConfiguration) bool {
	if networkConfig.FenceMode == types.FenceModeIsolated ||
		(networkConfig.FenceMode == types.FenceModeNAT && networkConfig.IPScopes != nil &&
			networkConfig.IPScopes.IPScope != nil && len(networkConfig.IPScopes.IPScope) > 0 &&
			!networkConfig.IPScopes.IPScope[0].IsInherited) {
		return true
	}
	return false
}

// RemoveAllNetworkStaticRoutes removes all static routes from a vApp network
// Returns error
func (vapp *VApp) RemoveAllNetworkStaticRoutes(networkId string) error {
	task, err := vapp.UpdateNetworkStaticRoutingAsync(networkId, []*types.StaticRoute{}, false)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}
	return nil
}
