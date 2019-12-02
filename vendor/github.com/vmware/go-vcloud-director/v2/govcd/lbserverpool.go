/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLbServerPool creates a load balancer server pool based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including ID) populated or an error.
// Name and Algorithm fields must be populated.
func (egw *EdgeGateway) CreateLbServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	if err := validateCreateLbServerPool(lbPoolConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer server pool: %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/pools/pool-7]
	lbPoolID, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readPool, err := egw.GetLbServerPoolById(lbPoolID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve lb server pool with ID (%s) after creation: %s", lbPoolID, err)
	}
	return readPool, nil
}

// getLbServerPool is able to find the types.LbPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) getLbServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	if err := validateGetLbServerPool(lbPoolConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "server pool response"
	lbPoolResponse := &struct {
		LBPools []*types.LbPool `xml:"pool"`
	}{}

	// This query returns all server pools as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load lalancer server pool: %s", nil, lbPoolResponse)
	if err != nil {
		return nil, err
	}

	// Search for pool by ID or by Name
	for _, pool := range lbPoolResponse.LBPools {
		// If ID was specified for lookup - look for the same ID
		if lbPoolConfig.ID != "" && pool.ID == lbPoolConfig.ID {
			return pool, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbPoolConfig.Name != "" && pool.Name == lbPoolConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbPoolConfig.ID != "" && pool.ID != lbPoolConfig.ID {
				return nil, fmt.Errorf("load balancer server pool was found by name (%s), but its ID (%s) does not match specified ID (%s)",
					pool.Name, pool.ID, lbPoolConfig.ID)
			}
			return pool, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetLbServerPoolByName wraps getLbServerPool and needs only an ID for lookup
func (egw *EdgeGateway) GetLbServerPoolById(id string) (*types.LbPool, error) {
	return egw.getLbServerPool(&types.LbPool{ID: id})
}

// GetLbServerPoolByName wraps getLbServerPool and needs only a Name for lookup
func (egw *EdgeGateway) GetLbServerPoolByName(name string) (*types.LbPool, error) {
	return egw.getLbServerPool(&types.LbPool{Name: name})
}

// UpdateLbServerPool updates types.LbPool with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
// Name and Algorithm fields must be populated.
func (egw *EdgeGateway) UpdateLbServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	err := validateUpdateLbServerPool(lbPoolConfig, egw)
	if err != nil {
		return nil, err
	}

	lbPoolConfig.ID, err = egw.getLbServerPoolIdByNameId(lbPoolConfig.Name, lbPoolConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer server pool: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbServerPoolPath + lbPoolConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer server pool : %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readPool, err := egw.GetLbServerPoolById(lbPoolConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve server pool with ID (%s) after update: %s", lbPoolConfig.ID, err)
	}
	return readPool, nil
}

// DeleteLbServerPool is able to delete the types.LbPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) DeleteLbServerPool(lbPoolConfig *types.LbPool) error {
	err := validateDeleteLbServerPool(lbPoolConfig, egw)
	if err != nil {
		return err
	}

	lbPoolConfig.ID, err = egw.getLbServerPoolIdByNameId(lbPoolConfig.Name, lbPoolConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer server pool: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbServerPoolPath + lbPoolConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete server pool: %s", nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteLbServerPoolById wraps DeleteLbServerPool and requires only ID for deletion
func (egw *EdgeGateway) DeleteLbServerPoolById(id string) error {
	return egw.DeleteLbServerPool(&types.LbPool{ID: id})
}

// DeleteLbServerPoolByName wraps DeleteLbServerPool and requires only Name for deletion
func (egw *EdgeGateway) DeleteLbServerPoolByName(name string) error {
	return egw.DeleteLbServerPool(&types.LbPool{Name: name})
}

func validateCreateLbServerPool(lbPoolConfig *types.LbPool, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbPoolConfig.Name == "" {
		return fmt.Errorf("load balancer server pool Name cannot be empty")
	}

	if lbPoolConfig.Algorithm == "" {
		return fmt.Errorf("load balancer server pool Algorithm cannot be empty")
	}

	for _, member := range lbPoolConfig.Members {
		if member.Condition == "" {
			return fmt.Errorf("load balancer server pool Member must have Condition set")
		}
	}

	return nil
}

func validateGetLbServerPool(lbPoolConfig *types.LbPool, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbPoolConfig.ID == "" && lbPoolConfig.Name == "" {
		return fmt.Errorf("to read load balancer server pool at least one of `ID`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLbServerPool(lbPoolConfig *types.LbPool, egw *EdgeGateway) error {
	// Update and create have the same requirements for now
	return validateCreateLbServerPool(lbPoolConfig, egw)
}

func validateDeleteLbServerPool(lbPoolConfig *types.LbPool, egw *EdgeGateway) error {
	// Read and delete have the same requirements for now
	return validateGetLbServerPool(lbPoolConfig, egw)
}

// getLbServerPoolIdByNameId checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (egw *EdgeGateway) getLbServerPoolIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"server pool got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbServerPool, err := egw.GetLbServerPoolByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer server pool by name: %s", err)
	}
	return readlbServerPool.ID, nil
}
