/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// requestEdgeFirewallRules nests EdgeFirewallRule as a convenience for unmarshalling POST requests
type requestEdgeFirewallRules struct {
	XMLName           xml.Name                  `xml:"firewallRules"`
	EdgeFirewallRules []*types.EdgeFirewallRule `xml:"firewallRule"`
}

// responseEdgeFirewallRules is used to unwrap response when retrieving
type responseEdgeFirewallRules struct {
	XMLName           xml.Name                 `xml:"firewall"`
	Version           string                   `xml:"version"`
	EdgeFirewallRules requestEdgeFirewallRules `xml:"firewallRules"`
}

// CreateNsxvFirewall creates firewall rule using proxied NSX-V API. It is a synchronuous operation.
// It returns an object with all fields populated (including ID)
// If aboveRuleId is not empty, it will send a query parameter aboveRuleId= which instructs NSX to
// place this rule above the specified rule ID
func (egw *EdgeGateway) CreateNsxvFirewall(firewallRuleConfig *types.EdgeFirewallRule, aboveRuleId string) (*types.EdgeFirewallRule, error) {
	if err := validateCreateNsxvFirewall(firewallRuleConfig, egw); err != nil {
		return nil, err
	}

	params := make(map[string]string)
	if aboveRuleId != "" {
		params["aboveRuleId"] = aboveRuleId
	}

	// Wrap the provided rule for POST request
	firewallRuleRequest := requestEdgeFirewallRules{
		EdgeFirewallRules: []*types.EdgeFirewallRule{firewallRuleConfig},
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	// The query must be wrapped differently, depending if it mus specify the "aboveRuleId" parameter
	var resp *http.Response
	if aboveRuleId == "" {
		resp, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
			"error creating firewall rule: %s", firewallRuleRequest, &types.NSXError{})
	} else {
		errString := fmt.Sprintf("error creating firewall rule (aboveRuleId: %s): %%s", aboveRuleId)
		resp, err = egw.client.ExecuteParamRequestWithCustomError(httpPath, params, http.MethodPost, types.AnyXMLMime,
			errString, firewallRuleConfig, &types.NSXError{})
	}

	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-1/firewall/config/rules/197157]
	firewallRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readFirewallRule, err := egw.GetNsxvFirewallById(firewallRuleId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve firewall rule with ID (%s) after creation: %s",
			firewallRuleId, err)
	}
	return readFirewallRule, nil
}

// UpdateNsxvFirewall updates types.EdgeFirewallRule with all fields using proxied NSX-V API. ID is
// mandatory to perform the update.
func (egw *EdgeGateway) UpdateNsxvFirewall(firewallRuleConfig *types.EdgeFirewallRule) (*types.EdgeFirewallRule, error) {
	err := validateUpdateNsxvFirewall(firewallRuleConfig, egw)
	if err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath + "/" + firewallRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating firewall rule : %s", firewallRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readFirewallRule, err := egw.GetNsxvFirewallById(firewallRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve firewall rule with ID (%s) after update: %s",
			readFirewallRule.ID, err)
	}
	return readFirewallRule, nil
}

// GetNsxvFirewallById retrieves types.EdgeFirewallRule by firewall rule ID as shown in the UI using proxied
// NSX-V API.
// It returns and error `ErrorEntityNotFound` if the firewall rule is not found
func (egw *EdgeGateway) GetNsxvFirewallById(id string) (*types.EdgeFirewallRule, error) {
	if err := validateGetNsxvFirewall(id, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	firewallRuleResponse := &responseEdgeFirewallRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read firewall rule: %s", nil, firewallRuleResponse)
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] Searching for firewall rule with ID: %s", id)
	for _, rule := range firewallRuleResponse.EdgeFirewallRules.EdgeFirewallRules {
		util.Logger.Printf("[DEBUG] Checking rule: %#+v", rule)
		if rule.ID != "" && rule.ID == id {
			return rule, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetAllNsxvFirewallRules retrieves all firewall rules and returns []*types.EdgeFirewallRule or an
// error of type ErrorEntityNotFound if there are no firewall rules
func (egw *EdgeGateway) GetAllNsxvFirewallRules() ([]*types.EdgeFirewallRule, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support firewall rules")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	firewallRuleResponse := &responseEdgeFirewallRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read firewall rule: %s", nil, firewallRuleResponse)
	if err != nil {
		return nil, err
	}

	if len(firewallRuleResponse.EdgeFirewallRules.EdgeFirewallRules) == 0 {
		return nil, ErrorEntityNotFound
	}

	return firewallRuleResponse.EdgeFirewallRules.EdgeFirewallRules, nil
}

// DeleteNsxvFirewallById deletes types.EdgeFirewallRule by firewall rule ID as shown in the UI using proxied
// NSX-V API.
// It returns and error `ErrorEntityNotFound` if the firewall rule is not found.
func (egw *EdgeGateway) DeleteNsxvFirewallById(id string) error {
	err := validateDeleteNsxvFirewall(id, egw)
	if err != nil {
		return err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath + "/" + id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// check if the rule exists and pass back the error at it may be 'ErrorEntityNotFound'
	_, err = egw.GetNsxvFirewallById(id)
	if err != nil {
		return err
	}

	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete firewall rule: %s", nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

func validateCreateNsxvFirewall(firewallRuleConfig *types.EdgeFirewallRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support firewall rules")
	}

	if firewallRuleConfig.Action == "" {
		return fmt.Errorf("firewall rule must have action specified")
	}

	return nil
}

func validateUpdateNsxvFirewall(firewallRuleConfig *types.EdgeFirewallRule, egw *EdgeGateway) error {
	if firewallRuleConfig.ID == "" {
		return fmt.Errorf("firewall rule ID must be set for update")
	}

	return validateCreateNsxvFirewall(firewallRuleConfig, egw)
}

func validateGetNsxvFirewall(id string, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support firewall rules")
	}

	if id == "" {
		return fmt.Errorf("unable to retrieve firewall rule without ID")
	}

	return nil
}

func validateDeleteNsxvFirewall(id string, egw *EdgeGateway) error {
	return validateGetNsxvFirewall(id, egw)
}
