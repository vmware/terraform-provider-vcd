package govcd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBServerPool creates a load balancer server pool based on mandatory fields. It is a synchronous
// operation. It returns created object with all fields (including ID) populated or an error.
func (eGW *EdgeGateway) CreateLBServerPool(lbPoolConfig *types.LBPool) (*types.LBPool, error) {
	if err := validateCreateLBServerPool(lbPoolConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer server pool: %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}
	location := resp.Header.Get("Location")

	// Last element in location header is the server pool ID
	// i.e. Location: [/network/edges/edge-3/loadbalancer/config/pools/pool-7]
	if location == "" {
		return nil, fmt.Errorf("unable to retrieve ID for new load balancer server pool with name %s", lbPoolConfig.Name)
	}
	splitLocation := strings.Split(location, "/")
	lbPoolID := splitLocation[len(splitLocation)-1]
	readPool, err := eGW.ReadLBServerPool(&types.LBPool{ID: lbPoolID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve lb server pool with ID (%s) after creation: %s", readPool.ID, err)
	}
	return readPool, nil
}

// ReadLBServerPool is able to find the types.LBPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBServerPool(lbPoolConfig *types.LBPool) (*types.LBPool, error) {
	if err := validateReadLBServerPool(lbPoolConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBServerPoolPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap "server pool response"
	lbPoolResponse := &struct {
		LBPools []*types.LBPool `xml:"pool"`
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
		if lbPoolConfig.ID != "" && pool.ID == lbPoolConfig.ID {
			return pool, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbPoolConfig.Name != "" && pool.Name == lbPoolConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbPoolConfig.ID != "" && pool.ID != lbPoolConfig.ID {
				return nil, fmt.Errorf("load balancer server pool was found by name (%s), but it's ID (%s) does not match specified ID (%s)",
					pool.Name, pool.ID, lbPoolConfig.ID)
			}
			return pool, nil
		}
	}

	return nil, fmt.Errorf("could not find load balancer server pool (name: %s, ID: %s)",
		lbPoolConfig.Name, lbPoolConfig.ID)
}

// UpdateLBServerPool updates types.LBPool with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) UpdateLBServerPool(lbPoolConfig *types.LBPool) (*types.LBPool, error) {
	if err := validateUpdateLBServerPool(lbPoolConfig); err != nil {
		return nil, err
	}

	// if only name was specified for update, ID must be found, because ID is mandatory for update
	if lbPoolConfig.ID == "" {
		readLBPool, err := eGW.ReadLBServerPool(&types.LBPool{Name: lbPoolConfig.Name})
		if err != nil {
			return nil, fmt.Errorf("unable to find load balancer pool by name for update: %s", err)
		}
		lbPoolConfig.ID = readLBPool.ID
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBServerPoolPath + lbPoolConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer server pool : %s", lbPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readPool, err := eGW.ReadLBServerPool(&types.LBPool{ID: lbPoolConfig.ID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve server pool with ID (%s) after update: %s", readPool.ID, err)
	}
	return readPool, nil
}

// DeleteLBServerPool is able to delete the types.LBPool type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBServerPool(lbPoolConfig *types.LBPool) error {
	if err := validateDeleteLBServerPool(lbPoolConfig); err != nil {
		return err
	}

	lbPoolID := lbPoolConfig.ID
	// if only name was specified for deletion, ID must be found, because only ID can be used for deletion
	if lbPoolConfig.ID == "" {
		readLBPool, err := eGW.ReadLBServerPool(&types.LBPool{Name: lbPoolConfig.Name})
		if err != nil {
			return fmt.Errorf("unable to find load balancer pool by name for deletion: %s", err)
		}
		lbPoolID = readLBPool.ID
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBServerPoolPath + lbPoolID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete Server Pool: %s", nil)
}

func validateCreateLBServerPool(lbPoolConfig *types.LBPool) error {
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

func validateReadLBServerPool(lbPoolConfig *types.LBPool) error {
	if lbPoolConfig.ID == "" && lbPoolConfig.Name == "" {
		return fmt.Errorf("to read load balancer server pool at least one of `ID`, `Name` fields must be specified")
	}

	return nil
}

func validateUpdateLBServerPool(lbPoolConfig *types.LBPool) error {
	// Update and create have the same requirements for now
	return validateCreateLBServerPool(lbPoolConfig)
}

func validateDeleteLBServerPool(lbPoolConfig *types.LBPool) error {
	// Read and delete have the same requirements for now
	return validateReadLBServerPool(lbPoolConfig)
}
