/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type EdgeGateway struct {
	EdgeGateway *types.EdgeGateway
	client      *Client
}

// Simplified structure used to list networks connected to an edge gateway
type SimpleNetworkIdentifier struct {
	Name          string
	InterfaceType string
}

var reErrorBusy = regexp.MustCompile(`is busy completing an operation.$`)

func NewEdgeGateway(cli *Client) *EdgeGateway {
	return &EdgeGateway{
		EdgeGateway: new(types.EdgeGateway),
		client:      cli,
	}
}

// Struct which covers NAT rule fields
type NatRule struct {
	NatType      string
	NetworkHref  string
	ExternalIP   string
	ExternalPort string
	InternalIP   string
	InternalPort string
	Protocol     string
	IcmpSubType  string
	Description  string
}

// AddDhcpPool adds (or updates) the DHCP pool connected to a specific network.
// TODO: this is legacy code from 2015, which requires a Terraform structure to work. It may need some re-thinking.
func (egw *EdgeGateway) AddDhcpPool(network *types.OrgVDCNetwork, dhcppool []interface{}) (Task, error) {
	newEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	util.Logger.Printf("[DEBUG] EDGE GATEWAY: %#v", newEdgeConfig)
	util.Logger.Printf("[DEBUG] EDGE GATEWAY SERVICE: %#v", newEdgeConfig.GatewayDhcpService)
	newDchpService := &types.GatewayDhcpService{}
	if newEdgeConfig.GatewayDhcpService.Pool == nil {
		newDchpService.IsEnabled = true
	} else {
		newDchpService.IsEnabled = newEdgeConfig.GatewayDhcpService.IsEnabled

		for _, dhcpPoolService := range newEdgeConfig.GatewayDhcpService.Pool {

			// Kludgy IF to avoid deleting DNAT rules not created by us.
			// If matches, let's skip it and continue the loop
			// Note: a simple comparison of HREF fields may fail if one of them is
			// from a tenant object and the other from a provider object. They may have the
			// same ID but different paths. Using 'equalIds' we determine equality even with
			// different paths
			if equalIds(network.HREF, "", dhcpPoolService.Network.HREF) {
				continue
			}

			newDchpService.Pool = append(newDchpService.Pool, dhcpPoolService)
		}
	}

	for _, item := range dhcppool {
		data := item.(map[string]interface{})

		if data["default_lease_time"] == nil {
			data["default_lease_time"] = 3600
		}

		if data["max_lease_time"] == nil {
			data["max_lease_time"] = 7200
		}

		dhcpRule := &types.DhcpPoolService{
			IsEnabled: true,
			Network: &types.Reference{
				HREF: network.HREF,
				Name: network.Name,
			},
			DefaultLeaseTime: data["default_lease_time"].(int),
			MaxLeaseTime:     data["max_lease_time"].(int),
			LowIPAddress:     data["start_address"].(string),
			HighIPAddress:    data["end_address"].(string),
		}
		newDchpService.Pool = append(newDchpService.Pool, dhcpRule)
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:              types.XMLNamespaceVCloud,
		GatewayDhcpService: newDchpService,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := egw.client.NewRequest(map[string]string{}, http.MethodPost, *apiEndpoint, buffer)
		util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(egw.client.Http.Do(req))
		if err != nil {
			if reErrorBusy.MatchString(err.Error()) {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(egw.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

// Deprecated: use one of RemoveNATRuleAsync, RemoveNATRule
func (egw *EdgeGateway) RemoveNATMapping(natType, externalIP, internalIP, port string) (Task, error) {
	return egw.RemoveNATPortMapping(natType, externalIP, port, internalIP, port)
}

// Deprecated: use one of RemoveNATRuleAsync, RemoveNATRule
func (egw *EdgeGateway) RemoveNATPortMapping(natType, externalIP, externalPort, internalIP, internalPort string) (Task, error) {
	// Find uplink interface
	var uplink types.Reference
	for _, gi := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gi.InterfaceType != "uplink" {
			continue
		}
		uplink = *gi.Network
	}

	newEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newNatService := &types.NatService{}

	newNatService.IsEnabled = newEdgeConfig.NatService.IsEnabled
	newNatService.NatType = newEdgeConfig.NatService.NatType
	newNatService.Policy = newEdgeConfig.NatService.Policy
	newNatService.ExternalIP = newEdgeConfig.NatService.ExternalIP

	for _, natRule := range newEdgeConfig.NatService.NatRule {

		if natRule.RuleType == natType &&
			natRule.GatewayNatRule.OriginalIP == externalIP &&
			natRule.GatewayNatRule.OriginalPort == externalPort &&
			natRule.GatewayNatRule.Interface.HREF == uplink.HREF {
			util.Logger.Printf("[DEBUG] REMOVING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
			continue
		}
		util.Logger.Printf("[DEBUG] KEEPING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
		newNatService.NatRule = append(newNatService.NatRule, natRule)
	}

	newEdgeConfig.NatService = newNatService

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      types.XMLNamespaceVCloud,
		NatService: newNatService,
	}

	apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newRules)

}

// RemoveNATRule removes NAT removes NAT rule identified by ID and handles task. Returns error if issues rise.
// Old functions RemoveNATPortMapping and RemoveNATMapping removed using rule details
// and expected interface to be of external network type.
func (egw *EdgeGateway) RemoveNATRule(id string) error {
	task, err := egw.RemoveNATRuleAsync(id)
	if err != nil {
		return fmt.Errorf("error removing DNAT rule: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	return nil
}

// RemoveNATRuleAsync removes NAT rule or returns an error.
// Old functions RemoveNATPortMapping and RemoveNATMapping removed using rule details
// and expected interface to be of external network type.
func (egw *EdgeGateway) RemoveNATRuleAsync(id string) (Task, error) {
	if id == "" {
		return Task{}, fmt.Errorf("provided id is empty")
	}

	err := egw.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing edge gateway: %s", err)
	}

	natServiceToUpdate := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService
	ruleIndex := -1
	if natServiceToUpdate != nil {
		for n, existingNatRule := range natServiceToUpdate.NatRule {
			if existingNatRule.ID == id {
				ruleIndex = n
				break
			}
		}
	} else {
		return Task{}, fmt.Errorf("edge gateway doesn't have NAT rules")
	}

	if ruleIndex == -1 {
		return Task{}, fmt.Errorf("edge gateway doesn't have rule with such ID")
	}

	if len(natServiceToUpdate.NatRule) > 1 {
		natServiceToUpdate.NatRule = append(natServiceToUpdate.NatRule[:ruleIndex], natServiceToUpdate.NatRule[ruleIndex+1:]...)
	} else {
		natServiceToUpdate.NatRule = nil
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      types.XMLNamespaceVCloud,
		NatService: natServiceToUpdate,
	}

	egwConfigureHref, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	egwConfigureHref.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(egwConfigureHref.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newRules)
}

// AddDNATRule creates DNAT rule and returns the NAT struct that was created or an error.
// Allows assigning a specific Org VDC or an external network.
// When edge gateway is advanced vCD API uses element <tag> to map with NSX edge gateway ID. A known issue is
// that updating rule using User interface resets <tag> and as result mapping is lost.
// Getting using NatRule.ID won't be valid anymore.
// Old functions AddNATPortMapping and AddNATMapping assigned rule only to first external network
func (egw *EdgeGateway) AddDNATRule(ruleDetails NatRule) (*types.NatRule, error) {
	mappingId, err := getPseudoUuid()
	if err != nil {
		return nil, err
	}
	originalDescription := ruleDetails.Description
	ruleDetails.Description = mappingId

	ruleDetails.NatType = "DNAT"
	task, err := egw.AddNATRuleAsync(ruleDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating DNAT rule: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	var createdNatRule *types.NatRule

	err = egw.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing edge gateway: %s", err)
	}

	for _, natRule := range egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
		if natRule.Description == mappingId {
			createdNatRule = natRule
			break
		}
	}

	if createdNatRule == nil {
		return nil, fmt.Errorf("error creating DNAT rule, didn't match created rule")
	}

	createdNatRule.Description = originalDescription

	return egw.UpdateNatRule(createdNatRule)
}

// AddSNATRule creates SNAT rule and returns created NAT rule or error.
// Allows assigning a specific Org VDC or an external network.
// Old functions AddNATPortMapping and AddNATMapping aren't correct as assigned rule only to first external network
func (egw *EdgeGateway) AddSNATRule(networkHref, externalIP, internalIP, description string) (*types.NatRule, error) {

	// As vCD API doesn't return rule ID we get it manually:
	//  * create rule with description which value is our generated ID
	//  * find rule which has description with our generated ID
	//  * get the real (vCD's) rule ID
	//  * update description with real value and return nat rule

	mappingId, err := getPseudoUuid()
	if err != nil {
		return nil, err
	}

	task, err := egw.AddNATRuleAsync(NatRule{NetworkHref: networkHref, NatType: "SNAT", ExternalIP: externalIP,
		ExternalPort: "any", InternalIP: internalIP, InternalPort: "any",
		IcmpSubType: "", Protocol: "any", Description: mappingId})
	if err != nil {
		return nil, fmt.Errorf("error creating SNAT rule: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	var createdNatRule *types.NatRule

	err = egw.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing edge gateway: %s", err)
	}

	for _, natRule := range egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
		if natRule.Description == mappingId {
			createdNatRule = natRule
			break
		}
	}

	if createdNatRule == nil {
		return nil, fmt.Errorf("error creating SNAT rule, didn't match created rule")
	}

	createdNatRule.Description = description

	return egw.UpdateNatRule(createdNatRule)
}

// getPseudoUuid creates unique ID/UUID
func getPseudoUuid() (string, error) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid, nil
}

// UpdateNatRule updates NAT rule and handles task. Returns updated NAT rule or error.
func (egw *EdgeGateway) UpdateNatRule(natRule *types.NatRule) (*types.NatRule, error) {
	task, err := egw.UpdateNatRuleAsync(natRule)
	if err != nil {
		return nil, fmt.Errorf("error updating NAT rule: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("%s", combinedTaskErrorMessage(task.Task, err))
	}

	return egw.GetNatRule(natRule.ID)
}

// UpdateNatRuleAsync updates NAT rule and returns task or error.
func (egw *EdgeGateway) UpdateNatRuleAsync(natRule *types.NatRule) (Task, error) {
	if natRule.GatewayNatRule.Protocol != "" && !isValidProtocol(natRule.GatewayNatRule.Protocol) {
		return Task{}, fmt.Errorf("provided protocol is not one of TCP, UDP, TCPUDP, ICMP, ANY")
	}

	if strings.ToUpper(natRule.GatewayNatRule.Protocol) == "ICMP" && !isValidIcmpSubType(natRule.GatewayNatRule.IcmpSubType) {
		return Task{}, fmt.Errorf("provided icmp sub type is not correct")
	}

	err := egw.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing edge gateway: %s", err)
	}

	natServiceToUpdate := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService

	if natServiceToUpdate != nil {
		for n, existingNatRule := range natServiceToUpdate.NatRule {
			if existingNatRule.ID == natRule.ID {
				natServiceToUpdate.NatRule[n] = natRule
			}
		}
	} else {
		return Task{}, fmt.Errorf("edge gateway doesn't have such nat rule")
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      types.XMLNamespaceVCloud,
		NatService: natServiceToUpdate,
	}

	egwConfigureHref, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	egwConfigureHref.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(egwConfigureHref.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newRules)
}

// GetNatRule returns NAT rule or error.
func (egw *EdgeGateway) GetNatRule(id string) (*types.NatRule, error) {
	err := egw.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing edge gateway: %s", err)
	}

	if egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService != nil {
		for _, natRule := range egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if natRule.ID == id {
				return natRule, nil
			}
		}
	}

	return nil, ErrorEntityNotFound
}

// AddNATRuleAsync creates NAT rule and return task or err
// Allows assigning specific network Org VDC or external. Old function AddNATPortMapping and
// AddNATMapping function shouldn't be used because assigns rule to first external network
func (egw *EdgeGateway) AddNATRuleAsync(ruleDetails NatRule) (Task, error) {
	if !isValidProtocol(ruleDetails.Protocol) {
		return Task{}, fmt.Errorf("provided protocol is not one of TCP, UDP, TCPUDP, ICMP, ANY")
	}

	if strings.ToUpper(ruleDetails.Protocol) == "ICMP" && !isValidIcmpSubType(ruleDetails.IcmpSubType) {
		return Task{}, fmt.Errorf("provided icmp sub type is not correct")
	}

	currentEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newNatService := &types.NatService{}

	if currentEdgeConfig.NatService == nil {
		newNatService.IsEnabled = true
	} else {
		newNatService.IsEnabled = currentEdgeConfig.NatService.IsEnabled
		newNatService.NatType = currentEdgeConfig.NatService.NatType
		newNatService.Policy = currentEdgeConfig.NatService.Policy
		newNatService.ExternalIP = currentEdgeConfig.NatService.ExternalIP
		newNatService.NatRule = currentEdgeConfig.NatService.NatRule
	}

	//construct new rule
	natRule := &types.NatRule{
		RuleType:    ruleDetails.NatType,
		IsEnabled:   takeBoolPointer(true),
		Description: ruleDetails.Description,
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: ruleDetails.NetworkHref,
			},
			OriginalIP:     ruleDetails.ExternalIP,
			OriginalPort:   ruleDetails.ExternalPort,
			TranslatedIP:   ruleDetails.InternalIP,
			TranslatedPort: ruleDetails.InternalPort,
			Protocol:       ruleDetails.Protocol,
			IcmpSubType:    ruleDetails.IcmpSubType,
		},
	}

	newNatService.NatRule = append(newNatService.NatRule, natRule)
	currentEdgeConfig.NatService = newNatService
	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      types.XMLNamespaceVCloud,
		NatService: newNatService,
	}

	egwConfigureHref, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	egwConfigureHref.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(egwConfigureHref.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newRules)
}

// Deprecated: Use eGW.AddSNATRule() or eGW.AddDNATRule()
func (egw *EdgeGateway) AddNATRule(network *types.OrgVDCNetwork, natType, externalIP, internalIP string) (Task, error) {
	return egw.AddNATPortMappingWithUplink(network, natType, externalIP, "any", internalIP, "any", "any", "")
}

// Deprecated: Use eGW.AddNATRule()
func (egw *EdgeGateway) AddNATMapping(natType, externalIP, internalIP string) (Task, error) {
	return egw.AddNATPortMapping(natType, externalIP, "any", internalIP, "any", "any", "")
}

// Deprecated: Use eGW.AddNATPortMappingWithUplink()
func (egw *EdgeGateway) AddNATPortMapping(natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType string) (Task, error) {
	return egw.AddNATPortMappingWithUplink(nil, natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType)
}

// Deprecated: creates not good behaviour of functionality
func (egw *EdgeGateway) getFirstUplink() types.Reference {
	var uplink types.Reference
	for _, gi := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gi.InterfaceType != "uplink" {
			continue
		}
		uplink = *gi.Network
	}
	return uplink
}

// Values are matched with VCD UI when creating DNAT for edge gateway.
func isValidProtocol(protocol string) bool {
	switch strings.ToUpper(protocol) {
	case
		"TCP",
		"UDP",
		"TCPUDP",
		"ICMP",
		"ANY":
		return true
	}
	return false
}

// Used values are named here https://code.vmware.com/apis/287/vcloud#/doc/doc/types/GatewayNatRuleType.html
// Also can be matched in VCD UI when creating DNAT for edge gateway.
func isValidIcmpSubType(protocol string) bool {
	switch strings.ToLower(protocol) {
	case
		"address-mask-request",
		"address-mask-reply",
		"destination-unreachable",
		"echo-request",
		"echo-reply",
		"parameter-problem",
		"redirect",
		"router-advertisement",
		"router-solicitation",
		"source-quench",
		"time-exceeded",
		"timestamp-request",
		"timestamp-reply",
		"any":
		return true
	}
	return false
}

// Deprecated: Use eGW.AddDNATRule() or eGW.CreateNsxvNatRule() for NSX-V
func (egw *EdgeGateway) AddNATPortMappingWithUplink(network *types.OrgVDCNetwork, natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType string) (Task, error) {
	// if a network is provided take it, otherwise find first uplink on the edge gateway
	var uplinkRef string

	if network != nil {
		uplinkRef = network.HREF
	} else {
		// TODO: remove when method used this removed
		uplinkRef = egw.getFirstUplink().HREF
	}

	if !isValidProtocol(protocol) {
		return Task{}, fmt.Errorf("provided protocol is not one of TCP, UDP, TCPUDP, ICMP, ANY")
	}

	if strings.ToUpper(protocol) == "ICMP" && !isValidIcmpSubType(icmpSubType) {
		return Task{}, fmt.Errorf("provided icmp sub type is not correct")
	}

	newEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newNatService := &types.NatService{}

	if newEdgeConfig.NatService == nil {
		newNatService.IsEnabled = true
	} else {
		newNatService.IsEnabled = newEdgeConfig.NatService.IsEnabled
		newNatService.NatType = newEdgeConfig.NatService.NatType
		newNatService.Policy = newEdgeConfig.NatService.Policy
		newNatService.ExternalIP = newEdgeConfig.NatService.ExternalIP

		for _, natRule := range newEdgeConfig.NatService.NatRule {

			// Kludgy IF to avoid deleting DNAT rules not created by us.
			// If matches, let's skip it and continue the loop
			if natRule.RuleType == natType &&
				natRule.GatewayNatRule.OriginalIP == externalIP &&
				natRule.GatewayNatRule.OriginalPort == externalPort &&
				natRule.GatewayNatRule.TranslatedIP == internalIP &&
				natRule.GatewayNatRule.TranslatedPort == internalPort &&
				natRule.GatewayNatRule.Interface.HREF == uplinkRef {
				continue
			}

			newNatService.NatRule = append(newNatService.NatRule, natRule)
		}
	}

	//add rule
	natRule := &types.NatRule{
		RuleType:  natType,
		IsEnabled: takeBoolPointer(true),
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkRef,
			},
			OriginalIP:     externalIP,
			OriginalPort:   externalPort,
			TranslatedIP:   internalIP,
			TranslatedPort: internalPort,
			Protocol:       protocol,
			IcmpSubType:    icmpSubType,
		},
	}
	newNatService.NatRule = append(newNatService.NatRule, natRule)

	newEdgeConfig.NatService = newNatService

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      types.XMLNamespaceVCloud,
		NatService: newNatService,
	}

	apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newRules)
}

func (egw *EdgeGateway) CreateFirewallRules(defaultAction string, rules []*types.FirewallRule) (Task, error) {
	err := egw.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error: %s", err)
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns: types.XMLNamespaceVCloud,
		FirewallService: &types.FirewallService{
			IsEnabled:        true,
			DefaultAction:    defaultAction,
			LogDefaultAction: true,
			FirewallRule:     rules,
		},
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error: %s", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := egw.client.NewRequest(map[string]string{}, http.MethodPost, *apiEndpoint, buffer)
		util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(egw.client.Http.Do(req))
		if err != nil {
			if reErrorBusy.MatchString(err.Error()) {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(egw.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (egw *EdgeGateway) Refresh() error {

	if egw.EdgeGateway == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url := egw.EdgeGateway.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	egw.EdgeGateway = &types.EdgeGateway{}

	_, err := egw.client.ExecuteRequest(url, http.MethodGet,
		"", "error retrieving Edge Gateway: %s", nil, egw.EdgeGateway)

	return err
}

func (egw *EdgeGateway) Remove1to1Mapping(internal, external string) (Task, error) {

	// Refresh EdgeGateway rules
	err := egw.Refresh()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	var uplinkif string
	for _, gifs := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newNatService := &types.NatService{}

	// Copy over the NAT configuration
	newNatService.IsEnabled = newEdgeConfig.NatService.IsEnabled
	newNatService.NatType = newEdgeConfig.NatService.NatType
	newNatService.Policy = newEdgeConfig.NatService.Policy
	newNatService.ExternalIP = newEdgeConfig.NatService.ExternalIP

	for i, natRule := range newEdgeConfig.NatService.NatRule {

		// Kludgy IF to avoid deleting DNAT rules not created by us.
		// If matches, let's skip it and continue the loop
		if natRule.RuleType == "DNAT" &&
			natRule.GatewayNatRule.OriginalIP == external &&
			natRule.GatewayNatRule.TranslatedIP == internal &&
			natRule.GatewayNatRule.OriginalPort == "any" &&
			natRule.GatewayNatRule.TranslatedPort == "any" &&
			natRule.GatewayNatRule.Protocol == "any" &&
			natRule.GatewayNatRule.Interface.HREF == uplinkif {
			continue
		}

		// Kludgy IF to avoid deleting SNAT rules not created by us.
		// If matches, let's skip it and continue the loop
		if natRule.RuleType == "SNAT" &&
			natRule.GatewayNatRule.OriginalIP == internal &&
			natRule.GatewayNatRule.TranslatedIP == external &&
			natRule.GatewayNatRule.Interface.HREF == uplinkif {
			continue
		}

		// If doesn't match the above IFs, it's something we need to preserve,
		// let's add it to the new NatService struct
		newNatService.NatRule = append(newNatService.NatRule, newEdgeConfig.NatService.NatRule[i])

	}

	// Fill the new NatService Section
	newEdgeConfig.NatService = newNatService

	// Take care of the Firewall service
	newFwService := &types.FirewallService{}

	// Copy over the firewall configuration
	newFwService.IsEnabled = newEdgeConfig.FirewallService.IsEnabled
	newFwService.DefaultAction = newEdgeConfig.FirewallService.DefaultAction
	newFwService.LogDefaultAction = newEdgeConfig.FirewallService.LogDefaultAction

	for i, firewallRule := range newEdgeConfig.FirewallService.FirewallRule {

		// Kludgy IF to avoid deleting inbound FW rules not created by us.
		// If matches, let's skip it and continue the loop
		if firewallRule.Policy == "allow" &&
			firewallRule.Protocols.Any &&
			firewallRule.DestinationPortRange == "Any" &&
			firewallRule.SourcePortRange == "Any" &&
			firewallRule.SourceIP == "Any" &&
			firewallRule.DestinationIP == external {
			continue
		}

		// Kludgy IF to avoid deleting outbound FW rules not created by us.
		// If matches, let's skip it and continue the loop
		if firewallRule.Policy == "allow" &&
			firewallRule.Protocols.Any &&
			firewallRule.DestinationPortRange == "Any" &&
			firewallRule.SourcePortRange == "Any" &&
			firewallRule.SourceIP == internal &&
			firewallRule.DestinationIP == "Any" {
			continue
		}

		// If doesn't match the above IFs, it's something we need to preserve,
		// let's add it to the new FirewallService struct
		newFwService.FirewallRule = append(newFwService.FirewallRule, newEdgeConfig.FirewallService.FirewallRule[i])

	}

	// Fill the new FirewallService Section
	newEdgeConfig.FirewallService = newFwService

	// Fix
	newEdgeConfig.NatService.IsEnabled = true

	apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newEdgeConfig)

}

func (egw *EdgeGateway) Create1to1Mapping(internal, external, description string) (Task, error) {

	// Refresh EdgeGateway rules
	err := egw.Refresh()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	var uplinkif string
	for _, gifs := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newEdgeConfig := egw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	snat := &types.NatRule{
		Description: description,
		RuleType:    "SNAT",
		IsEnabled:   takeBoolPointer(true),
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkif,
			},
			OriginalIP:   internal,
			TranslatedIP: external,
			Protocol:     "any",
		},
	}

	if newEdgeConfig.NatService == nil {
		newEdgeConfig.NatService = &types.NatService{}
	}
	newEdgeConfig.NatService.NatRule = append(newEdgeConfig.NatService.NatRule, snat)

	dnat := &types.NatRule{
		Description: description,
		RuleType:    "DNAT",
		IsEnabled:   takeBoolPointer(true),
		GatewayNatRule: &types.GatewayNatRule{
			Interface: &types.Reference{
				HREF: uplinkif,
			},
			OriginalIP:     external,
			OriginalPort:   "any",
			TranslatedIP:   internal,
			TranslatedPort: "any",
			Protocol:       "any",
		},
	}

	newEdgeConfig.NatService.NatRule = append(newEdgeConfig.NatService.NatRule, dnat)

	fwin := &types.FirewallRule{
		Description: description,
		IsEnabled:   true,
		Policy:      "allow",
		Protocols: &types.FirewallRuleProtocols{
			Any: true,
		},
		DestinationPortRange: "Any",
		DestinationIP:        external,
		SourcePortRange:      "Any",
		SourceIP:             "Any",
		EnableLogging:        false,
	}

	newEdgeConfig.FirewallService.FirewallRule = append(newEdgeConfig.FirewallService.FirewallRule, fwin)

	fwout := &types.FirewallRule{
		Description: description,
		IsEnabled:   true,
		Policy:      "allow",
		Protocols: &types.FirewallRuleProtocols{
			Any: true,
		},
		DestinationPortRange: "Any",
		DestinationIP:        "Any",
		SourcePortRange:      "Any",
		SourceIP:             internal,
		EnableLogging:        false,
	}

	newEdgeConfig.FirewallService.FirewallRule = append(newEdgeConfig.FirewallService.FirewallRule, fwout)

	apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", newEdgeConfig)

}

func (egw *EdgeGateway) AddIpsecVPN(ipsecVPNConfig *types.EdgeGatewayServiceConfiguration) (Task, error) {

	err := egw.Refresh()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	ipsecVPNConfig.Xmlns = types.XMLNamespaceVCloud

	apiEndpoint, _ := url.ParseRequestURI(egw.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	// Return the task
	return egw.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml", "error reconfiguring Edge Gateway: %s", ipsecVPNConfig)

}

// Removes an Edge Gateway VPN, by passing an empty configuration
func (egw *EdgeGateway) RemoveIpsecVPN() (Task, error) {
	err := egw.Refresh()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	ipsecVPNConfig := &types.EdgeGatewayServiceConfiguration{
		Xmlns: types.XMLNamespaceVCloud,
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: false,
		},
	}
	return egw.AddIpsecVPN(ipsecVPNConfig)
}

// Deletes the edge gateway, returning a task and an error with the operation result.
// https://code.vmware.com/apis/442/vcloud-director/doc/doc/operations/DELETE-EdgeGateway.html
func (egw *EdgeGateway) DeleteAsync(force bool, recursive bool) (Task, error) {
	util.Logger.Printf("[TRACE] EdgeGateway.Delete - deleting edge gateway with force: %t, recursive: %t", force, recursive)

	if egw.EdgeGateway.HREF == "" {
		return Task{}, fmt.Errorf("cannot delete, HREF is missing")
	}

	egwUrl, err := url.ParseRequestURI(egw.EdgeGateway.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing edge gateway url: %s", err)
	}

	req := egw.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, http.MethodDelete, *egwUrl, nil)
	resp, err := checkResp(egw.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting edge gateway: %s", err)
	}
	task := NewTask(egw.client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, err
}

// Deletes the edge gateway, returning an error with the operation result.
// https://code.vmware.com/apis/442/vcloud-director/doc/doc/operations/DELETE-EdgeGateway.html
func (egw *EdgeGateway) Delete(force bool, recursive bool) error {

	task, err := egw.DeleteAsync(force, recursive)
	if err != nil {
		return err
	}
	if task.Task.Status == "error" {
		return fmt.Errorf(combinedTaskErrorMessage(task.Task, fmt.Errorf("edge gateway not properly destroyed")))
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf(combinedTaskErrorMessage(task.Task, err))
	}

	return nil
}

// GetNetworks returns the list of networks associated with an edge gateway
// In the return structure, an interfaceType of "uplink" indicates an external network,
// while "internal" is for Org VDC routed networks
func (egw *EdgeGateway) GetNetworks() ([]SimpleNetworkIdentifier, error) {
	var networks []SimpleNetworkIdentifier
	err := egw.Refresh()
	if err != nil {
		return networks, err
	}
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		netIdentifier := SimpleNetworkIdentifier{
			Name:          net.Name,
			InterfaceType: net.InterfaceType,
		}
		networks = append(networks, netIdentifier)
	}

	return networks, nil
}

// HasDefaultGateway returns true if the edge gateway uses one of the external
// networks as default gateway
func (egw *EdgeGateway) HasDefaultGateway() bool {
	if egw.EdgeGateway.Configuration != nil &&
		egw.EdgeGateway.Configuration.GatewayInterfaces != nil {
		for _, gw := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
			// Check if the interface is used for default route
			if gw.UseForDefaultRoute {
				// Look for a specific subnet which is used as a default route
				for _, subnetParticipation := range gw.SubnetParticipation {
					if subnetParticipation.UseForDefaultRoute &&
						subnetParticipation.Gateway != "" &&
						subnetParticipation.Netmask != "" {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasAdvancedNetworking returns true if the edge gateway has advanced network configuration enabled
func (egw *EdgeGateway) HasAdvancedNetworking() bool {
	return egw.EdgeGateway.Configuration != nil &&
		egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled != nil &&
		*egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled
}

// buildProxiedEdgeEndpointURL helps to get root endpoint for Edge Gateway using the
// NSX API Proxy and can append optionalSuffix which must have its own leading /
func (egw *EdgeGateway) buildProxiedEdgeEndpointURL(optionalSuffix string) (string, error) {
	apiEndpoint, err := url.ParseRequestURI(egw.EdgeGateway.HREF)
	if err != nil {
		return "", fmt.Errorf("unable to process edge gateway URL: %s", err)
	}
	edgeID := strings.Split(egw.EdgeGateway.ID, ":")
	if len(edgeID) != 4 {
		return "", fmt.Errorf("unable to find edge gateway id: %s", egw.EdgeGateway.ID)
	}
	hostname := apiEndpoint.Scheme + "://" + apiEndpoint.Host + "/network/edges/" + edgeID[3]

	if optionalSuffix != "" {
		return hostname + optionalSuffix, nil
	}

	return hostname, nil
}

// GetLBGeneralParams retrieves load balancer configuration of `&types.LoadBalancer` and can be used
// to access global configuration options. These are 4 fields only:
// LoadBalancer.Enabled, LoadBalancer.AccelerationEnabled, LoadBalancer.Logging.Enable,
// LoadBalancer.Logging.LogLevel
func (egw *EdgeGateway) GetLBGeneralParams() (*types.LbGeneralParamsWithXml, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateway supports load balancing")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	loadBalancerConfig := &types.LbGeneralParamsWithXml{}
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer configuration: %s", nil, loadBalancerConfig)

	if err != nil {
		return nil, err
	}

	return loadBalancerConfig, nil
}

// UpdateLBGeneralParams allows to update global load balancer configuration.
// It accepts four fields (Enabled, AccelerationEnabled, Logging.Enable, Logging.LogLevel) and uses
// them to construct types.LbGeneralParamsWithXml without altering other options to prevent config
// corruption.
// They are represented in load balancer global configuration tab in the UI.
func (egw *EdgeGateway) UpdateLBGeneralParams(enabled, accelerationEnabled, loggingEnabled bool, logLevel string) (*types.LbGeneralParamsWithXml, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateway supports load balancing")
	}

	if err := validateUpdateLBGeneralParams(logLevel); err != nil {
		return nil, err
	}
	// Retrieve load balancer to work on latest configuration
	currentLb, err := egw.GetLBGeneralParams()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve load balancer before update: %s", err)
	}

	// Check if change is needed. If not - return early.
	if currentLb.Logging != nil &&
		currentLb.Enabled == enabled && currentLb.AccelerationEnabled == accelerationEnabled &&
		currentLb.Logging.Enable == loggingEnabled && currentLb.Logging.LogLevel == logLevel {
		return currentLb, nil
	}

	// Modify only the global configuration settings
	currentLb.Enabled = enabled
	currentLb.AccelerationEnabled = accelerationEnabled
	currentLb.Logging = &types.LbLogging{
		Enable:   loggingEnabled,
		LogLevel: logLevel,
	}
	// Omit the version as it is updated automatically with each put
	currentLb.Version = ""

	// Push updated configuration
	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer config: %s", currentLb, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Retrieve configuration after update
	updatedLb, err := egw.GetLBGeneralParams()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve load balancer config after update: %s", err)
	}

	return updatedLb, nil
}

// GetFirewallConfig retrieves firewall configuration and can be used
// to alter master configuration options. These are 3 fields only:
// FirewallConfigWithXml.Enabled, FirewallConfigWithXml.DefaultPolicy.LoggingEnabled and
// FirewallConfigWithXml.DefaultPolicy.Action
func (egw *EdgeGateway) GetFirewallConfig() (*types.FirewallConfigWithXml, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateway support firewall configuration")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	firewallConfig := &types.FirewallConfigWithXml{}
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read firewall configuration: %s", nil, firewallConfig)

	if err != nil {
		return nil, err
	}

	return firewallConfig, nil
}

// UpdateFirewallConfig allows to update firewall configuration.
// It accepts three fields (Enabled, DefaultLoggingEnabled, DefaultAction) and uses
// them to construct types.FirewallConfigWithXml without altering other options to prevent config
// corruption.
// They are represented in firewall configuration page in the UI.
func (egw *EdgeGateway) UpdateFirewallConfig(enabled, defaultLoggingEnabled bool, defaultAction string) (*types.FirewallConfigWithXml, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateway supports load balancing")
	}

	if defaultAction != "accept" && defaultAction != "deny" {
		return nil, fmt.Errorf("default action must be either 'accept' or 'deny'")
	}

	// Retrieve firewall latest configuration
	currentFw, err := egw.GetFirewallConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve firewall config before update: %s", err)
	}

	// Check if change is needed. If not - return early.
	if currentFw.Enabled == enabled && currentFw.DefaultPolicy.LoggingEnabled == defaultLoggingEnabled &&
		currentFw.DefaultPolicy.Action == defaultAction {
		return currentFw, nil
	}

	// Modify only the global configuration settings
	currentFw.Enabled = enabled
	currentFw.DefaultPolicy.LoggingEnabled = defaultLoggingEnabled
	currentFw.DefaultPolicy.Action = defaultAction

	// Omit the version as it is updated automatically with each put
	currentFw.Version = ""

	// Push updated configuration
	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating firewall configuration : %s", currentFw, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Retrieve configuration after update
	updatedFw, err := egw.GetFirewallConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve firewall after update: %s", err)
	}

	return updatedFw, nil
}

// validateUpdateLoadBalancer validates mandatory fields for global load balancer configuration
// settings
func validateUpdateLBGeneralParams(logLevel string) error {
	if logLevel == "" {
		return fmt.Errorf("field Logging.LogLevel must be set to update load balancer")
	}

	return nil
}

// getVdcNetworks retrieves a structure of type EdgeGatewayInterfaces which contains network
// interfaces available in Edge Gateway (uses "/vdcNetworks" endpoint)
func (egw *EdgeGateway) getVdcNetworks() (*types.EdgeGatewayInterfaces, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateway supports vNics")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL("/vdcNetworks")
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	vnicConfig := &types.EdgeGatewayInterfaces{}
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to edge gateway vnic configuration: %s", nil, vnicConfig)

	if err != nil {
		return nil, err
	}

	return vnicConfig, nil
}

// GetVnicIndexByNetworkNameAndType returns *int of vNic index for specified network name and network type
// networkType one of: 'internal', 'uplink', 'trunk', 'subinterface'
// networkName cannot be empty
func (egw *EdgeGateway) GetVnicIndexByNetworkNameAndType(networkName, networkType string) (*int, error) {
	vnics, err := egw.getVdcNetworks()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve vNic configuration: %s", err)
	}
	return getVnicIndexByNetworkNameAndType(networkName, networkType, vnics)
}

// GetAnyVnicIndexByNetworkName parses XML structure of vNic mapping to networks in edge gateway XML
// and returns *int of vNic index and network type by network name
// networkName cannot be empty
// networkType will be one of: 'internal', 'uplink', 'trunk', 'subinterface'
//
// Warning: this function assumes that there are no duplicate network names attached. If it is so
// this function will return the first network
func (egw *EdgeGateway) GetAnyVnicIndexByNetworkName(networkName string) (*int, string, error) {
	vnics, err := egw.getVdcNetworks()
	if err != nil {
		return nil, "", fmt.Errorf("cannot retrieve vNic configuration: %s", err)
	}

	var foundVnicIndex *int
	var foundVnicType string

	possibleNicTypes := []string{types.EdgeGatewayVnicTypeUplink, types.EdgeGatewayVnicTypeInternal,
		types.EdgeGatewayVnicTypeTrunk, types.EdgeGatewayVnicTypeSubinterface}

	for _, nicType := range possibleNicTypes {
		vNicIndex, err := getVnicIndexByNetworkNameAndType(networkName, nicType, vnics)
		if err == nil { // nil error means we have found nic
			foundVnicIndex = vNicIndex
			foundVnicType = nicType
			break
		}
	}

	if foundVnicIndex == nil && foundVnicType == "" {
		return nil, "", ErrorEntityNotFound
	}
	return foundVnicIndex, foundVnicType, nil
}

// GetNetworkNameAndTypeByVnicIndex returns network name and network type for given vNic index
// returned networkType can be one of: 'internal', 'uplink', 'trunk', 'subinterface'
func (egw *EdgeGateway) GetNetworkNameAndTypeByVnicIndex(vNicIndex int) (string, string, error) {
	vnics, err := egw.getVdcNetworks()
	if err != nil {
		return "", "", fmt.Errorf("cannot retrieve vNic configuration: %s", err)
	}
	return getNetworkNameAndTypeByVnicIndex(vNicIndex, vnics)
}

// getVnicIndexByNetworkNameAndType is wrapped and used by public function GetVnicIndexByNetworkNameAndType
func getVnicIndexByNetworkNameAndType(networkName, networkType string, vnics *types.EdgeGatewayInterfaces) (*int, error) {
	if networkName == "" {
		return nil, fmt.Errorf("network name cannot be empty")
	}
	if networkType != types.EdgeGatewayVnicTypeUplink &&
		networkType != types.EdgeGatewayVnicTypeInternal &&
		networkType != types.EdgeGatewayVnicTypeTrunk &&
		networkType != types.EdgeGatewayVnicTypeSubinterface {
		return nil, fmt.Errorf("networkType must be one of 'uplink', 'internal', 'trunk', 'subinterface'")
	}

	var foundIndex *int
	foundCount := 0

	for _, vnic := range vnics.EdgeInterface {
		// Look for matching portgroup name and network type. If the PortgroupName is not empty -
		// check that it contains network name as well.
		if vnic.Name == networkName && vnic.Type == networkType &&
			(vnic.PortgroupName == networkName || vnic.PortgroupName == "") {
			foundIndex = vnic.Index
			foundCount++
		}
	}

	if foundCount > 1 {
		return nil, fmt.Errorf("more than one (%d) networks of type '%s' with name '%s' found",
			foundCount, networkType, networkName)
	}

	if foundCount == 0 {
		return nil, ErrorEntityNotFound
	}

	return foundIndex, nil
}

// getNetworkNameAndTypeByVnicIndex looks up network type and name in list of edge gateway interfaces
func getNetworkNameAndTypeByVnicIndex(vNicIndex int, vnics *types.EdgeGatewayInterfaces) (string, string, error) {
	if vNicIndex < 0 {
		return "", "", fmt.Errorf("vNic index cannot be negative")
	}

	foundCount := 0
	var networkName, networkType string

	for _, vnic := range vnics.EdgeInterface {
		if vnic.Index != nil && *vnic.Index == vNicIndex {
			foundCount++
			networkName = vnic.Name
			networkType = vnic.Type
		}
	}

	if foundCount > 1 {
		return "", "", fmt.Errorf("more than one networks found for vNic %d", vNicIndex)
	}

	if foundCount == 0 {
		return "", "", ErrorEntityNotFound
	}

	return networkName, networkType, nil
}

// UpdateAsync updates the edge gateway in place with the information contained in the internal structure
func (egw *EdgeGateway) UpdateAsync() (Task, error) {

	egw.EdgeGateway.Xmlns = types.XMLNamespaceVCloud
	egw.EdgeGateway.Configuration.Xmlns = types.XMLNamespaceVCloud
	egw.EdgeGateway.Tasks = nil

	// Return the task
	return egw.client.ExecuteTaskRequest(egw.EdgeGateway.HREF, http.MethodPut,
		types.MimeEdgeGateway, "error updating Edge Gateway: %s", egw.EdgeGateway)
}

// Update is a wrapper around UpdateAsync
// The pointer receiver is refreshed after update
func (egw *EdgeGateway) Update() error {

	task, err := egw.UpdateAsync()
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return egw.Refresh()
}
