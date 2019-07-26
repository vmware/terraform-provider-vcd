/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLbServiceMonitor creates a load balancer service monitor based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including ID) populated or an error.
func (egw *EdgeGateway) CreateLbServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	if err := validateCreateLbServiceMonitor(lbMonitorConfig, egw); err != nil {
		return nil, err
	}

	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("edge gateway does not have advanced networking enabled")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer service monitor: %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/monitors/monitor-5]
	lbMonitorID, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readMonitor, err := egw.GetLbServiceMonitorById(lbMonitorID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with ID (%s) after creation: %s", lbMonitorID, err)
	}
	return readMonitor, nil
}

// getLbServiceMonitor is able to find the types.LbMonitor type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) getLbServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	if err := validateGetLbServiceMonitor(lbMonitorConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "monitor response"
	lbMonitorResponse := &struct {
		LBMonitors []*types.LbMonitor `xml:"monitor"`
	}{}

	// This query returns all service monitors as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime, "unable to read Load Balancer monitor: %s", nil, lbMonitorResponse)
	if err != nil {
		return nil, err
	}

	// Search for monitor by ID or by Name
	for _, monitor := range lbMonitorResponse.LBMonitors {
		// If ID was specified for lookup - look for the same ID
		if lbMonitorConfig.ID != "" && monitor.ID == lbMonitorConfig.ID {
			return monitor, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbMonitorConfig.Name != "" && monitor.Name == lbMonitorConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbMonitorConfig.ID != "" && monitor.ID != lbMonitorConfig.ID {
				return nil, fmt.Errorf("load balancer monitor was found by name (%s), but it's ID (%s) does not match specified ID (%s)",
					monitor.Name, monitor.ID, lbMonitorConfig.ID)
			}
			return monitor, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetLbServiceMonitorById wraps getLbServiceMonitor and needs only an ID for lookup
func (egw *EdgeGateway) GetLbServiceMonitorById(id string) (*types.LbMonitor, error) {
	return egw.getLbServiceMonitor(&types.LbMonitor{ID: id})
}

// GetLbServiceMonitorByName wraps getLbServiceMonitor and needs only a Name for lookup
func (egw *EdgeGateway) GetLbServiceMonitorByName(name string) (*types.LbMonitor, error) {
	return egw.getLbServiceMonitor(&types.LbMonitor{Name: name})
}

// UpdateLbServiceMonitor updates types.LbMonitor with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) UpdateLbServiceMonitor(lbMonitorConfig *types.LbMonitor) (*types.LbMonitor, error) {
	err := validateUpdateLbServiceMonitor(lbMonitorConfig, egw)
	if err != nil {
		return nil, err
	}

	lbMonitorConfig.ID, err = egw.getLbServiceMonitorIdByNameId(lbMonitorConfig.Name, lbMonitorConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer service monitor: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath + lbMonitorConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer service monitor : %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readMonitor, err := egw.GetLbServiceMonitorById(lbMonitorConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with ID (%s) after update: %s", lbMonitorConfig.ID, err)
	}
	return readMonitor, nil
}

// DeleteLbServiceMonitor is able to delete the types.LbMonitor type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) DeleteLbServiceMonitor(lbMonitorConfig *types.LbMonitor) error {
	err := validateDeleteLbServiceMonitor(lbMonitorConfig, egw)
	if err != nil {
		return err
	}

	lbMonitorConfig.ID, err = egw.getLbServiceMonitorIdByNameId(lbMonitorConfig.Name, lbMonitorConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer service monitor: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbMonitorPath + lbMonitorConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete service monitor: %s", nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteLbServiceMonitorById wraps DeleteLbServiceMonitor and requires only ID for deletion
func (egw *EdgeGateway) DeleteLbServiceMonitorById(id string) error {
	return egw.DeleteLbServiceMonitor(&types.LbMonitor{ID: id})
}

// DeleteLbServiceMonitorByName wraps DeleteLbServiceMonitor and requires only Name for deletion
func (egw *EdgeGateway) DeleteLbServiceMonitorByName(name string) error {
	return egw.DeleteLbServiceMonitor(&types.LbMonitor{Name: name})
}

func validateCreateLbServiceMonitor(lbMonitorConfig *types.LbMonitor, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbMonitorConfig.Name == "" {
		return fmt.Errorf("load balancer monitor Name cannot be empty")
	}

	if lbMonitorConfig.Timeout == 0 {
		return fmt.Errorf("load balancer monitor Timeout cannot be 0")
	}

	if lbMonitorConfig.Interval == 0 {
		return fmt.Errorf("load balancer monitor Interval cannot be 0")
	}

	if lbMonitorConfig.MaxRetries == 0 {
		return fmt.Errorf("load balancer monitor MaxRetries cannot be 0")
	}

	if lbMonitorConfig.Type == "" {
		return fmt.Errorf("load balancer monitor Type cannot be empty")
	}

	return nil
}

func validateGetLbServiceMonitor(lbMonitorConfig *types.LbMonitor, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbMonitorConfig.ID == "" && lbMonitorConfig.Name == "" {
		return fmt.Errorf("to read load balancer service monitor at least one of `ID`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLbServiceMonitor(lbMonitorConfig *types.LbMonitor, egw *EdgeGateway) error {
	// Update and create have the same requirements for now
	return validateCreateLbServiceMonitor(lbMonitorConfig, egw)
}

func validateDeleteLbServiceMonitor(lbMonitorConfig *types.LbMonitor, egw *EdgeGateway) error {
	// Read and delete have the same requirements for now
	return validateGetLbServiceMonitor(lbMonitorConfig, egw)
}

// getLbServiceMonitorIdByNameId checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (egw *EdgeGateway) getLbServiceMonitorIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"service monitor got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbServiceMonitor, err := egw.GetLbServiceMonitorByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer service monitor by name: %s", err)
	}
	return readlbServiceMonitor.ID, nil
}
