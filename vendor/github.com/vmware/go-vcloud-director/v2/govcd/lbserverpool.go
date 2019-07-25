/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBServerPool creates a load balancer server pool based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including ID) populated or an error.
// Name and Algorithm fields must be populated.
func (eGW *EdgeGateway) CreateLBServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	if err := validateCreateLBServerPool(lbPoolConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer server pool: %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// Location: [/network/edges/edge-3/loadbalancer/config/pools/pool-7]
	lbPoolID, err := extractNSXObjectIDFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readPool, err := eGW.ReadLBServerPoolByID(lbPoolID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve lb server pool with ID (%s) after creation: %s", lbPoolID, err)
	}
	return readPool, nil
}

// ReadLBServerPool is able to find the types.LBPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	if err := validateReadLBServerPool(lbPoolConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "server pool response"
	lbPoolResponse := &struct {
		LBPools []*types.LbPool `xml:"pool"`
	}{}

	// This query returns all server pools as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load lalancer server pool: %s", nil, lbPoolResponse)
	if err != nil {
		return nil, err
	}

	// Search for pool by ID or by Name
	for _, pool := range lbPoolResponse.LBPools {
		// If ID was specified for lookup - look for the same ID
		if lbPoolConfig.Id != "" && pool.Id == lbPoolConfig.Id {
			return pool, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbPoolConfig.Name != "" && pool.Name == lbPoolConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbPoolConfig.Id != "" && pool.Id != lbPoolConfig.Id {
				return nil, fmt.Errorf("load balancer server pool was found by name (%s), but it's ID (%s) does not match specified ID (%s)",
					pool.Name, pool.Id, lbPoolConfig.Id)
			}
			return pool, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBServerPoolByName wraps ReadLBServerPool and needs only an ID for lookup
func (eGW *EdgeGateway) ReadLBServerPoolByID(id string) (*types.LbPool, error) {
	return eGW.ReadLBServerPool(&types.LbPool{Id: id})
}

// ReadLBServerPoolByName wraps ReadLBServerPool and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBServerPoolByName(name string) (*types.LbPool, error) {
	return eGW.ReadLBServerPool(&types.LbPool{Name: name})
}

// UpdateLBServerPool updates types.LBPool with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
// Name and Algorithm fields must be populated.
func (eGW *EdgeGateway) UpdateLBServerPool(lbPoolConfig *types.LbPool) (*types.LbPool, error) {
	err := validateUpdateLBServerPool(lbPoolConfig)
	if err != nil {
		return nil, err
	}

	lbPoolConfig.Id, err = eGW.getLBServerPoolIDByNameID(lbPoolConfig.Name, lbPoolConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer server pool: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbServerPoolPath + lbPoolConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer server pool : %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readPool, err := eGW.ReadLBServerPoolByID(lbPoolConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve server pool with ID (%s) after update: %s", lbPoolConfig.Id, err)
	}
	return readPool, nil
}

// DeleteLBServerPool is able to delete the types.LBPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBServerPool(lbPoolConfig *types.LbPool) error {
	err := validateDeleteLBServerPool(lbPoolConfig)
	if err != nil {
		return err
	}

	lbPoolConfig.Id, err = eGW.getLBServerPoolIDByNameID(lbPoolConfig.Name, lbPoolConfig.Id)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer server pool: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LbServerPoolPath + lbPoolConfig.Id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete Server Pool: %s", nil)
}

// DeleteLBServerPoolByID wraps DeleteLBServerPool and requires only ID for deletion
func (eGW *EdgeGateway) DeleteLBServerPoolByID(id string) error {
	return eGW.DeleteLBServerPool(&types.LbPool{Id: id})
}

// DeleteLBServerPoolByName wraps DeleteLBServerPool and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBServerPoolByName(name string) error {
	return eGW.DeleteLBServerPool(&types.LbPool{Name: name})
}

func validateCreateLBServerPool(lbPoolConfig *types.LbPool) error {
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

func validateReadLBServerPool(lbPoolConfig *types.LbPool) error {
	if lbPoolConfig.Id == "" && lbPoolConfig.Name == "" {
		return fmt.Errorf("to read load balancer server pool at least one of `ID`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLBServerPool(lbPoolConfig *types.LbPool) error {
	// Update and create have the same requirements for now
	return validateCreateLBServerPool(lbPoolConfig)
}

func validateDeleteLBServerPool(lbPoolConfig *types.LbPool) error {
	// Read and delete have the same requirements for now
	return validateReadLBServerPool(lbPoolConfig)
}

// getLBServerPoolIDByNameID checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (eGW *EdgeGateway) getLBServerPoolIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"server pool got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbServerPool, err := eGW.ReadLBServerPoolByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer server pool by name: %s", err)
	}
	return readlbServerPool.Id, nil
}
