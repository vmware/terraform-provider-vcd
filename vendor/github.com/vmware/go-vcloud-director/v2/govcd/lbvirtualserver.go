/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLbVirtualServer creates a load balancer virtual server based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including ID) populated
// or an error.
// Name, Protocol, Port and IpAddress fields must be populated
func (egw *EdgeGateway) CreateLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	if err := validateCreateLbVirtualServer(lbVirtualServerConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer virtual server: %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/virtualservers/virtualServer-10]
	lbVirtualServerId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readVirtualServer, err := egw.GetLbVirtualServerById(lbVirtualServerId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve load balancer virtual server with ID (%s) after creation: %s",
			lbVirtualServerId, err)
	}
	return readVirtualServer, nil
}

// getLbVirtualServer is able to find the types.LbVirtualServer type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) getLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	if err := validateGetLbVirtualServer(lbVirtualServerConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "virtual server response"
	lbVirtualServerResponse := &struct {
		LBVirtualServers []*types.LbVirtualServer `xml:"virtualServer"`
	}{}

	// This query returns all virtual servers as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer virtual server: %s", nil, lbVirtualServerResponse)
	if err != nil {
		return nil, err
	}

	// Search for virtual server by ID or by Name
	for _, virtualServer := range lbVirtualServerResponse.LBVirtualServers {
		// If ID was specified for lookup - look for the same ID
		if lbVirtualServerConfig.ID != "" && virtualServer.ID == lbVirtualServerConfig.ID {
			return virtualServer, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbVirtualServerConfig.Name != "" && virtualServer.Name == lbVirtualServerConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbVirtualServerConfig.ID != "" && virtualServer.ID != lbVirtualServerConfig.ID {
				return nil, fmt.Errorf("load balancer virtual server was found by name (%s), "+
					"but its ID (%s) does not match specified ID (%s)",
					virtualServer.Name, virtualServer.ID, lbVirtualServerConfig.ID)
			}
			return virtualServer, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetLbVirtualServerById wraps getLbVirtualServer and needs only an ID for lookup
func (egw *EdgeGateway) GetLbVirtualServerById(id string) (*types.LbVirtualServer, error) {
	return egw.getLbVirtualServer(&types.LbVirtualServer{ID: id})
}

// GetLbVirtualServerByName wraps getLbVirtualServer and needs only a Name for lookup
func (egw *EdgeGateway) GetLbVirtualServerByName(name string) (*types.LbVirtualServer, error) {
	return egw.getLbVirtualServer(&types.LbVirtualServer{Name: name})
}

// UpdateLbVirtualServer updates types.LbVirtualServer with all fields. At least name or ID must be
// specified. If both - Name and ID are specified it performs a lookup by ID and returns an error if
// the specified name and found name do not match.
// Name, Protocol, Port and IpAddress fields must be populated
func (egw *EdgeGateway) UpdateLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) (*types.LbVirtualServer, error) {
	err := validateUpdateLbVirtualServer(lbVirtualServerConfig, egw)
	if err != nil {
		return nil, err
	}

	lbVirtualServerConfig.ID, err = egw.getLbVirtualServerIdByNameId(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer virtual server: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer virtual server : %s", lbVirtualServerConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readVirtualServer, err := egw.GetLbVirtualServerById(lbVirtualServerConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve virtual server with ID (%s) after update: %s",
			lbVirtualServerConfig.ID, err)
	}
	return readVirtualServer, nil
}

// DeleteLbVirtualServer is able to delete the types.LbVirtualServer type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the
// specified name and found name do not match.
func (egw *EdgeGateway) DeleteLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer) error {
	err := validateDeleteLbVirtualServer(lbVirtualServerConfig, egw)
	if err != nil {
		return err
	}

	lbVirtualServerConfig.ID, err = egw.getLbVirtualServerIdByNameId(lbVirtualServerConfig.Name, lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer virtual server: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbVirtualServerPath + lbVirtualServerConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return egw.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete load balancer virtual server: %s", nil)
}

// DeleteLbVirtualServerById wraps DeleteLbVirtualServer and requires only ID for deletion
func (egw *EdgeGateway) DeleteLbVirtualServerById(id string) error {
	return egw.DeleteLbVirtualServer(&types.LbVirtualServer{ID: id})
}

// DeleteLbVirtualServerByName wraps DeleteLbVirtualServer and requires only Name for deletion
func (egw *EdgeGateway) DeleteLbVirtualServerByName(name string) error {
	return egw.DeleteLbVirtualServer(&types.LbVirtualServer{Name: name})
}

func validateCreateLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

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

func validateGetLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbVirtualServerConfig.ID == "" && lbVirtualServerConfig.Name == "" {
		return fmt.Errorf("to read load balancer virtual server at least one of `ID`, `Name` " +
			"fields must be specified")
	}

	return nil
}

func validateUpdateLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer, egw *EdgeGateway) error {
	// Update and create have the same requirements for now
	return validateCreateLbVirtualServer(lbVirtualServerConfig, egw)
}

func validateDeleteLbVirtualServer(lbVirtualServerConfig *types.LbVirtualServer, egw *EdgeGateway) error {
	// Read and delete have the same requirements for now
	return validateGetLbVirtualServer(lbVirtualServerConfig, egw)
}

// getLbVirtualServerIdByNameId checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (egw *EdgeGateway) getLbVirtualServerIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"virtual server got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readLbVirtualServer, err := egw.GetLbVirtualServerByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer virtual server by name: %s", err)
	}
	return readLbVirtualServer.ID, nil
}
