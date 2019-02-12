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

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

type EdgeGateway struct {
	EdgeGateway *types.EdgeGateway
	client      *Client
}

func NewEdgeGateway(cli *Client) *EdgeGateway {
	return &EdgeGateway{
		EdgeGateway: new(types.EdgeGateway),
		client:      cli,
	}
}

func (eGW *EdgeGateway) AddDhcpPool(network *types.OrgVDCNetwork, dhcppool []interface{}) (Task, error) {
	newEdgeConfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
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
			if dhcpPoolService.Network.HREF == network.HREF {
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
		Xmlns:              "http://www.vmware.com/vcloud/v1.5",
		GatewayDhcpService: newDchpService,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
		util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(eGW.client.Http.Do(req))
		if err != nil {
			if match, _ := regexp.MatchString("is busy completing an operation.$", err.Error()); match {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) RemoveNATMapping(natType, externalIP, internalIP, port string) (Task, error) {
	return eGW.RemoveNATPortMapping(natType, externalIP, port, internalIP, port)
}

func (eGW *EdgeGateway) RemoveNATPortMapping(natType, externalIP, externalPort string, internalIP, internalPort string) (Task, error) {
	// Find uplink interface
	var uplink types.Reference
	for _, gi := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gi.InterfaceType != "uplink" {
			continue
		}
		uplink = *gi.Network
	}

	newEdgeConfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	// Take care of the NAT service
	newNatService := &types.NatService{}

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
			natRule.GatewayNatRule.Interface.HREF == uplink.HREF {
			util.Logger.Printf("[DEBUG] REMOVING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
			continue
		}
		util.Logger.Printf("[DEBUG] KEEPING %s Rule: %#v", natRule.RuleType, natRule.GatewayNatRule)
		newNatService.NatRule = append(newNatService.NatRule, natRule)
	}

	newEdgeConfig.NatService = newNatService

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns:      "http://www.vmware.com/vcloud/v1.5",
		NatService: newNatService,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
	util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
	util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[DEBUG] Error is: %#v", err)
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) AddNATMapping(natType, externalIP, internalIP string) (Task, error) {
	return eGW.AddNATPortMapping(natType, externalIP, "any", internalIP, "any", "any", "")
}

func (eGW *EdgeGateway) AddNATPortMapping(natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType string) (Task, error) {
	return eGW.AddNATPortMappingWithUplink(nil, natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType)
}

func (eGW *EdgeGateway) getFirstUplink() types.Reference {
	var uplink types.Reference
	for _, gi := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
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

func (eGW *EdgeGateway) AddNATPortMappingWithUplink(network *types.OrgVDCNetwork, natType, externalIP, externalPort, internalIP, internalPort, protocol, icmpSubType string) (Task, error) {
	// if a network is provided take it, otherwise find first uplink on the edge gateway
	var uplinkRef string

	if network != nil {
		uplinkRef = network.HREF
	} else {
		uplinkRef = eGW.getFirstUplink().HREF
	}

	if !isValidProtocol(protocol) {
		return Task{}, fmt.Errorf("provided protocol is not one of TCP, UDP, TCPUDP, ICMP, ANY")
	}

	if strings.ToUpper(protocol) == "ICMP" && !isValidIcmpSubType(icmpSubType) {
		return Task{}, fmt.Errorf("provided icmp sub type is not correct")
	}

	newEdgeConfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

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
		IsEnabled: true,
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
		Xmlns:      "http://www.vmware.com/vcloud/v1.5",
		NatService: newNatService,
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
	util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
	util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		util.Logger.Printf("[DEBUG] Error is: %#v", err)
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) CreateFirewallRules(defaultAction string, rules []*types.FirewallRule) (Task, error) {
	err := eGW.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error: %v\n", err)
	}

	newRules := &types.EdgeGatewayServiceConfiguration{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		FirewallService: &types.FirewallService{
			IsEnabled:        true,
			DefaultAction:    defaultAction,
			LogDefaultAction: true,
			FirewallRule:     rules,
		},
	}

	output, err := xml.MarshalIndent(newRules, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error: %v\n", err)
	}

	var resp *http.Response
	for {
		buffer := bytes.NewBufferString(xml.Header + string(output))

		apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
		apiEndpoint.Path += "/action/configureServices"

		req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)
		util.Logger.Printf("[DEBUG] POSTING TO URL: %s", apiEndpoint.Path)
		util.Logger.Printf("[DEBUG] XML TO SEND:\n%s", buffer)

		req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

		resp, err = checkResp(eGW.client.Http.Do(req))
		if err != nil {
			if match, _ := regexp.MatchString("is busy completing an operation.$", err.Error()); match {
				time.Sleep(3 * time.Second)
				continue
			}
			return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
		}
		break
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func (eGW *EdgeGateway) Refresh() error {

	if eGW.EdgeGateway == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)

	req := eGW.client.NewRequest(map[string]string{}, "GET", *apiEndpoint, nil)

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving Edge Gateway: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	eGW.EdgeGateway = &types.EdgeGateway{}

	if err = decodeBody(resp, eGW.EdgeGateway); err != nil {
		return fmt.Errorf("error decoding Edge Gateway response: %s", err)
	}

	// The request was successful
	return nil
}

func (eGW *EdgeGateway) Remove1to1Mapping(internal, external string) (Task, error) {

	// Refresh EdgeGateway rules
	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	var uplinkif string
	for _, gifs := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newEdgeConfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

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
			firewallRule.Protocols.Any == true &&
			firewallRule.DestinationPortRange == "Any" &&
			firewallRule.SourcePortRange == "Any" &&
			firewallRule.SourceIP == "Any" &&
			firewallRule.DestinationIP == external {
			continue
		}

		// Kludgy IF to avoid deleting outbound FW rules not created by us.
		// If matches, let's skip it and continue the loop
		if firewallRule.Policy == "allow" &&
			firewallRule.Protocols.Any == true &&
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

	output, err := xml.MarshalIndent(newEdgeConfig, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) Create1to1Mapping(internal, external, description string) (Task, error) {

	// Refresh EdgeGateway rules
	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	var uplinkif string
	for _, gifs := range eGW.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gifs.InterfaceType == "uplink" {
			uplinkif = gifs.Network.HREF
		}
	}

	newEdgeConfig := eGW.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration

	snat := &types.NatRule{
		Description: description,
		RuleType:    "SNAT",
		IsEnabled:   true,
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
		IsEnabled:   true,
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

	output, err := xml.MarshalIndent(newEdgeConfig, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

func (eGW *EdgeGateway) AddIpsecVPN(ipsecVPNConfig *types.EdgeGatewayServiceConfiguration) (Task, error) {

	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	output, err := xml.MarshalIndent(ipsecVPNConfig, "  ", "    ")
	if err != nil {
		return Task{}, fmt.Errorf("error marshaling ipsecVPNConfig compose: %s", err)
	}

	util.Logger.Printf("\n\nXML DEBUG: %s\n\n", string(output))

	buffer := bytes.NewBufferString(xml.Header + string(output))
	util.Logger.Printf("[DEBUG] ipsecVPN configuration: %s", buffer)

	apiEndpoint, _ := url.ParseRequestURI(eGW.EdgeGateway.HREF)
	apiEndpoint.Path += "/action/configureServices"

	req := eGW.client.NewRequest(map[string]string{}, "POST", *apiEndpoint, buffer)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml")

	resp, err := checkResp(eGW.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error reconfiguring Edge Gateway: %s", err)
	}

	if os.Getenv("GOVCD_DEBUG") != "" {
		util.Logger.Printf("Edge Gateway Service Configuration: %s\n", prettyEdgeGatewayServiceConfiguration(ipsecVPNConfig))
	}

	task := NewTask(eGW.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}

// Removes an Edge Gateway VPN, by passing an empty configuration
func (eGW *EdgeGateway) RemoveIpsecVPN() (Task, error) {
	err := eGW.Refresh()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	ipsecVPNConfig := &types.EdgeGatewayServiceConfiguration{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: false,
		},
	}
	return eGW.AddIpsecVPN(ipsecVPNConfig)
}
