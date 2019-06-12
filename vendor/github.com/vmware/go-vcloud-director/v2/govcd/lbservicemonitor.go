/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBServiceMonitor creates a load balancer service monitor based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including ID) populated or an error.
func (eGW *EdgeGateway) CreateLBServiceMonitor(lbMonitorConfig *types.LBMonitor) (*types.LBMonitor, error) {
	if err := validateCreateLBServiceMonitor(lbMonitorConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer service monitor: %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}
	location := resp.Header.Get("Location")

	// Last element in location header is the service monitor ID
	// i.e. Location: [/network/edges/edge-3/loadbalancer/config/monitors/monitor-5]
	// The code below extracts that ID from the last segment
	if location == "" {
		return nil, fmt.Errorf("unable to retrieve ID for new load balancer service monitor with name %s", lbMonitorConfig.Name)
	}
	splitLocation := strings.Split(location, "/")
	lbMonitorID := splitLocation[len(splitLocation)-1]

	readMonitor, err := eGW.ReadLBServiceMonitor(&types.LBMonitor{ID: lbMonitorID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with ID (%s) after creation: %s", readMonitor.ID, err)
	}
	return readMonitor, nil
}

// ReadLBServiceMonitor is able to find the types.LBMonitor type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBServiceMonitor(lbMonitorConfig *types.LBMonitor) (*types.LBMonitor, error) {
	if err := validateReadLBServiceMonitor(lbMonitorConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBMonitorPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "monitor response"
	lbMonitorResponse := &struct {
		LBMonitors []*types.LBMonitor `xml:"monitor"`
	}{}

	// This query returns all service monitors as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime, "unable to read Load Balancer monitor: %s", nil, lbMonitorResponse)
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

	return nil, fmt.Errorf("could not find load balancer service monitor (name: %s, ID: %s)",
		lbMonitorConfig.Name, lbMonitorConfig.ID)
}

// UpdateLBServiceMonitor
func (eGW *EdgeGateway) UpdateLBServiceMonitor(lbMonitorConfig *types.LBMonitor) (*types.LBMonitor, error) {
	if err := validateUpdateLBServiceMonitor(lbMonitorConfig); err != nil {
		return nil, err
	}

	// if only name was specified for update, ID must be found, because ID is mandatory for update
	if lbMonitorConfig.ID == "" {
		readLBMonitor, err := eGW.ReadLBServiceMonitor(&types.LBMonitor{Name: lbMonitorConfig.Name})
		if err != nil {
			return nil, err
		}
		lbMonitorConfig.ID = readLBMonitor.ID
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBMonitorPath + lbMonitorConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer service monitor : %s", lbMonitorConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readMonitor, err := eGW.ReadLBServiceMonitor(&types.LBMonitor{ID: lbMonitorConfig.ID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve monitor with ID (%s) after update: %s", readMonitor.ID, err)
	}
	return readMonitor, nil
}

// DeleteLBServiceMonitor is able to delete the types.LBMonitor type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBServiceMonitor(lbMonitorConfig *types.LBMonitor) error {
	if err := validateDeleteLBServiceMonitor(lbMonitorConfig); err != nil {
		return err
	}

	lbMonitorID := lbMonitorConfig.ID
	// if only name was specified for deletion, ID must be found, because only ID can be used for deletion
	if lbMonitorConfig.ID == "" {
		readLBMonitor, err := eGW.ReadLBServiceMonitor(&types.LBMonitor{Name: lbMonitorConfig.Name})
		if err != nil {
			return fmt.Errorf("unable to find load balancer monitor by name for deletion: %s", err)
		}
		lbMonitorID = readLBMonitor.ID
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBMonitorPath + lbMonitorID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete Service Monitor: %s", nil)
}

func validateCreateLBServiceMonitor(lbMonitorConfig *types.LBMonitor) error {
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

func validateReadLBServiceMonitor(lbMonitorConfig *types.LBMonitor) error {
	if lbMonitorConfig.ID == "" && lbMonitorConfig.Name == "" {
		return fmt.Errorf("to read load balancer service monitor at least one of `ID`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLBServiceMonitor(lbMonitorConfig *types.LBMonitor) error {
	// Update and create have the same requirements for now
	return validateCreateLBServiceMonitor(lbMonitorConfig)
}

func validateDeleteLBServiceMonitor(lbMonitorConfig *types.LBMonitor) error {
	// Read and delete have the same requirements for now
	return validateReadLBServiceMonitor(lbMonitorConfig)
}
