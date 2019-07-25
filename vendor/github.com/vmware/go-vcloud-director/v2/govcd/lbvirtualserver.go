/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBVirtualServer creates a load balancer virtual server based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including Id) populated
// or an error.
// Name, Protocol, Port and IpAddress fields must be populated
func (eGW *EdgeGateway) CreateLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	if err := validateCreateLBVirtualServer(lbVirtualServerConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer virtual server: %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/virtualservers/virtualServer-10]
	lbVirtualServerId, err := extractNSXObjectIDFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readVirtualServer, err := eGW.ReadLBVirtualServerById(lbVirtualServerId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve load balancer virtual server with Id (%s) after creation: %s",
			lbVirtualServerId, err)
	}
	return readVirtualServer, nil
}

// ReadLBVirtualServer is able to find the types.LBVirtualServer type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	if err := validateReadLBVirtualServer(lbVirtualServerConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "virtual server response"
	lbVirtualServerResponse := &struct {
		LBVirtualServers []*types.LbVirtualServer `xml:"virtualServer"`
	}{}

	// This query returns all virtual servers as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer virtual server: %s", nil, lbVirtualServerResponse)
	if err != nil {
		return nil, err
	}

	// Search for virtual server by Id or by Name
	for _, virtualServer := range lbVirtualServerResponse.LBVirtualServers {
		// If Id was specified for lookup - look for the same Id
		if lbVirtualServerConfig.ID != "" && virtualServer.ID == lbVirtualServerConfig.ID {
			return virtualServer, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbVirtualServerConfig.Name != "" && virtualServer.Name == lbVirtualServerConfig.Name {
			// We found it by name. Let's verify if search Id was specified and it matches the lookup object
			if lbVirtualServerConfig.ID != "" && virtualServer.ID != lbVirtualServerConfig.ID {
				return nil, fmt.Errorf("load balancer virtual server was found by name (%s), "+
					"but its Id (%s) does not match specified Id (%s)",
					virtualServer.Name, virtualServer.ID, lbVirtualServerConfig.ID)
			}
			return virtualServer, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBVirtualServerById wraps ReadLBVirtualServer and needs only an Id for lookup
func (eGW *EdgeGateway) ReadLBVirtualServerById(id string) (*types.LbVirtualServer, error) {
	return eGW.ReadLBVirtualServer(&types.LbVirtualServer{ID: id})
}

// ReadLBVirtualServerByName wraps ReadLBVirtualServer and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBVirtualServerByName(name string) (*types.LbVirtualServer, error) {
	return eGW.ReadLBVirtualServer(&types.LbVirtualServer{Name: name})
}

// UpdateLBVirtualServer updates types.LBVirtualServer with all fields. At least name or Id must be
// specified. If both - Name and Id are specified it performs a lookup by Id and returns an error if
// the specified name and found name do not match.
// Name, Protocol, Port and IpAddress fields must be populated
func (eGW *EdgeGateway) UpdateLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	err := validateUpdateLBVirtualServer(lbVirtualServerConfig)
	if err != nil {
		return nil, err
	}

	lbVirtualServerConfig.ID, err = eGW.getLBVirtualServerIdByNameId(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer virtual server: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer virtual server : %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readVirtualServer, err := eGW.ReadLBVirtualServerById(lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve virtual server with Id (%s) after update: %s",
			lbVirtualServerConfig.ID, err)
	}
	return readVirtualServer, nil
}

// DeleteLBVirtualServer is able to delete the types.LBVirtualServer type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the
// specified name and found name do not match.
func (eGW *EdgeGateway) DeleteLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	err := validateDeleteLBVirtualServer(lbVirtualServerConfig)
	if err != nil {
		return err
	}

	lbVirtualServerConfig.ID, err = eGW.getLBVirtualServerIdByNameId(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer virtual server: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete load balancer virtual server: %s", nil)
}

// DeleteLBVirtualServerById wraps DeleteLBVirtualServer and requires only Id for deletion
func (eGW *EdgeGateway) DeleteLBVirtualServerById(id string) error {
	return eGW.DeleteLBVirtualServer(&types.LbVirtualServer{ID: id})
}

// DeleteLBVirtualServerByName wraps DeleteLBVirtualServer and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBVirtualServerByName(name string) error {
	return eGW.DeleteLBVirtualServer(&types.LbVirtualServer{Name: name})
}

func validateCreateLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	if lbVirtualServerConfig.Name == "" {
		return fmt.Errorf("load balancer virtual server Name cannot be empty")
	}

	if lbVirtualServerConfig.IpAddress == "" {
		return fmt.Errorf("load balancer virtual server IpAddress cannot be empty")
	}

	if lbVirtualServerConfig.Protocol == "" {
		return fmt.Errorf("load balancer virtual server Protocol cannot be empty")
	}

	if lbVirtualServerConfig.Port == 0 {
		return fmt.Errorf("load balancer virtual server Port cannot be empty")
	}

	return nil
}

func validateReadLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	if lbVirtualServerConfig.ID == "" && lbVirtualServerConfig.Name == "" {
		return fmt.Errorf("to read load balancer virtual server at least one of `Id`, `Name` " +
			"fields must be specified")
	}

	return nil
}

func validateUpdateLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	// Update and create have the same requirements for now
	return validateCreateLBVirtualServer(lbVirtualServerConfig)
}

func validateDeleteLBVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	// Read and delete have the same requirements for now
	return validateReadLBVirtualServer(lbVirtualServerConfig)
}

// getLBVirtualServerIdByNameId checks if at least name or Id is set and returns the Id.
// If the Id is specified - it passes through the Id. If only name was specified
// it will lookup the object by name and return the Id.
func (eGW *EdgeGateway) getLBVirtualServerIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or Id must be specific to find load balancer "+
			"virtual server got name (%s) Id (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, Id must be found, because only Id can be used in request path
	readLbVirtualServer, err := eGW.ReadLBVirtualServerByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer virtual server by name: %s", err)
	}
	return readLbVirtualServer.ID, nil
}
