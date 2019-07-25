/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBAppRule creates a load balancer application rule based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including ID) populated or an error.
func (eGW *EdgeGateway) CreateLBAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateCreateLBAppRule(lbAppRuleConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer application rule: %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-3/loadbalancer/config/applicationrules/applicationRule-4]
	lbAppRuleId, err := extractNSXObjectIDFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppRule, err := eGW.ReadLBAppRule(&types.LbAppRule{ID: lbAppRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with ID (%s) after creation: %s",
			readAppRule.ID, err)
	}
	return readAppRule, nil
}

// ReadLBAppRule is able to find the types.LBAppRule type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateReadLBAppRule(lbAppRuleConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap response
	lbAppRuleResponse := &struct {
		LbAppRules []*types.LbAppRule `xml:"applicationRule"`
	}{}

	// This query returns all application rules as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
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

// ReadLBAppRuleById wraps ReadLBAppRule and needs only an ID for lookup
func (eGW *EdgeGateway) ReadLBAppRuleByID(id string) (*types.LbAppRule, error) {
	return eGW.ReadLBAppRule(&types.LbAppRule{ID: id})
}

// ReadLBAppRuleByName wraps ReadLBAppRule and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBAppRuleByName(name string) (*types.LbAppRule, error) {
	return eGW.ReadLBAppRule(&types.LbAppRule{Name: name})
}

// UpdateLBAppRule updates types.LBAppRule with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) UpdateLBAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	err := validateUpdateLBAppRule(lbAppRuleConfig)
	if err != nil {
		return nil, err
	}

	lbAppRuleConfig.ID, err = eGW.getLBAppRuleIDByNameID(lbAppRuleConfig.Name, lbAppRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer application rule : %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readAppRule, err := eGW.ReadLBAppRule(&types.LbAppRule{ID: lbAppRuleConfig.ID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with ID (%s) after update: %s",
			readAppRule.ID, err)
	}
	return readAppRule, nil
}

// DeleteLBAppRule is able to delete the types.LBAppRule type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBAppRule(lbAppRuleConfig *types.LbAppRule) error {
	err := validateDeleteLBAppRule(lbAppRuleConfig)
	if err != nil {
		return err
	}

	lbAppRuleConfig.ID, err = eGW.getLBAppRuleIDByNameID(lbAppRuleConfig.Name, lbAppRuleConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, "application/xml",
		"unable to delete application rule: %s", nil)
}

// DeleteLBAppRuleById wraps DeleteLBAppRule and requires only ID for deletion
func (eGW *EdgeGateway) DeleteLBAppRuleByID(id string) error {
	return eGW.DeleteLBAppRule(&types.LbAppRule{ID: id})
}

// DeleteLBAppRuleByName wraps DeleteLBAppRule and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBAppRuleByName(name string) error {
	return eGW.DeleteLBAppRule(&types.LbAppRule{Name: name})
}

func validateCreateLBAppRule(lbAppRuleConfig *types.LbAppRule) error {
	if lbAppRuleConfig.Name == "" {
		return fmt.Errorf("load balancer application rule Name cannot be empty")
	}

	return nil
}

func validateReadLBAppRule(lbAppRuleConfig *types.LbAppRule) error {
	if lbAppRuleConfig.ID == "" && lbAppRuleConfig.Name == "" {
		return fmt.Errorf("to read load balancer application rule at least one of `ID`, `Name`" +
			" fields must be specified")
	}

	return nil
}

func validateUpdateLBAppRule(lbAppRuleConfig *types.LbAppRule) error {
	// Update and create have the same requirements for now
	return validateCreateLBAppRule(lbAppRuleConfig)
}

func validateDeleteLBAppRule(lbAppRuleConfig *types.LbAppRule) error {
	// Read and delete have the same requirements for now
	return validateReadLBAppRule(lbAppRuleConfig)
}

// getLBAppRuleIDByNameID checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (eGW *EdgeGateway) getLBAppRuleIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"application rule got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbAppRule, err := eGW.ReadLBAppRuleByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer application rule by name: %s", err)
	}
	return readlbAppRule.ID, nil
}
