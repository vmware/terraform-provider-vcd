/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLbAppRule creates a load balancer application rule based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including ID) populated or an error.
func (egw *EdgeGateway) CreateLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateCreateLbAppRule(lbAppRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer application rule: %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-3/loadbalancer/config/applicationrules/applicationRule-4]
	lbAppRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppRule, err := egw.getLbAppRule(&types.LbAppRule{ID: lbAppRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with ID (%s) after creation: %s",
			readAppRule.ID, err)
	}
	return readAppRule, nil
}

// getLbAppRule is able to find the types.LbAppRule type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) getLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateGetLbAppRule(lbAppRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap response
	lbAppRuleResponse := &struct {
		LbAppRules []*types.LbAppRule `xml:"applicationRule"`
	}{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer application rule: %s", nil, lbAppRuleResponse)
	if err != nil {
		return nil, err
	}

	// Search for application rule by ID or by Name
	for _, rule := range lbAppRuleResponse.LbAppRules {
		// If ID was specified for lookup - look for the same ID
		if lbAppRuleConfig.ID != "" && rule.ID == lbAppRuleConfig.ID {
			return rule, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbAppRuleConfig.Name != "" && rule.Name == lbAppRuleConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbAppRuleConfig.ID != "" && rule.ID != lbAppRuleConfig.ID {
				return nil, fmt.Errorf("load balancer application rule was found by name (%s)"+
					", but its ID (%s) does not match specified ID (%s)",
					rule.Name, rule.ID, lbAppRuleConfig.ID)
			}
			return rule, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBAppRuleById wraps getLbAppRule and needs only an ID for lookup
func (egw *EdgeGateway) GetLbAppRuleById(id string) (*types.LbAppRule, error) {
	return egw.getLbAppRule(&types.LbAppRule{ID: id})
}

// GetLbAppRuleByName wraps getLbAppRule and needs only a Name for lookup
func (egw *EdgeGateway) GetLbAppRuleByName(name string) (*types.LbAppRule, error) {
	return egw.getLbAppRule(&types.LbAppRule{Name: name})
}

// UpdateLbAppRule updates types.LbAppRule with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) UpdateLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	err := validateUpdateLbAppRule(lbAppRuleConfig, egw)
	if err != nil {
		return nil, err
	}

	lbAppRuleConfig.ID, err = egw.getLbAppRuleIdByNameId(lbAppRuleConfig.Name, lbAppRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer application rule : %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readAppRule, err := egw.getLbAppRule(&types.LbAppRule{ID: lbAppRuleConfig.ID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with ID (%s) after update: %s",
			readAppRule.ID, err)
	}
	return readAppRule, nil
}

// DeleteLbAppRule is able to delete the types.LbAppRule type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) DeleteLbAppRule(lbAppRuleConfig *types.LbAppRule) error {
	err := validateDeleteLbAppRule(lbAppRuleConfig, egw)
	if err != nil {
		return err
	}

	lbAppRuleConfig.ID, err = egw.getLbAppRuleIdByNameId(lbAppRuleConfig.Name, lbAppRuleConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	return egw.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, "application/xml",
		"unable to delete application rule: %s", nil)
}

// DeleteLBAppRuleById wraps DeleteLbAppRule and requires only ID for deletion
func (egw *EdgeGateway) DeleteLbAppRuleById(id string) error {
	return egw.DeleteLbAppRule(&types.LbAppRule{ID: id})
}

// DeleteLbAppRuleByName wraps DeleteLbAppRule and requires only Name for deletion
func (egw *EdgeGateway) DeleteLbAppRuleByName(name string) error {
	return egw.DeleteLbAppRule(&types.LbAppRule{Name: name})
}

func validateCreateLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbAppRuleConfig.Name == "" {
		return fmt.Errorf("load balancer application rule Name cannot be empty")
	}

	return nil
}

func validateGetLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbAppRuleConfig.ID == "" && lbAppRuleConfig.Name == "" {
		return fmt.Errorf("to read load balancer application rule at least one of `ID`, `Name`" +
			" fields must be specified")
	}

	return nil
}

func validateUpdateLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	// Update and create have the same requirements for now
	return validateCreateLbAppRule(lbAppRuleConfig, egw)
}

func validateDeleteLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	// Read and delete have the same requirements for now
	return validateGetLbAppRule(lbAppRuleConfig, egw)
}

// getLbAppRuleIdByNameId checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (egw *EdgeGateway) getLbAppRuleIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"application rule got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbAppRule, err := egw.GetLbAppRuleByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer application rule by name: %s", err)
	}
	return readlbAppRule.ID, nil
}
