/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLBAppProfile creates a load balancer application profile based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including ID) populated or an error.
func (eGW *EdgeGateway) CreateLBAppProfile(lbAppProfileConfig *types.LBAppProfile) (*types.LBAppProfile, error) {
	if err := validateCreateLBAppProfile(lbAppProfileConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBAppProfilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer application profile: %s", lbAppProfileConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-3/loadbalancer/config/applicationprofiles/applicationProfile-4]
	lbAppProfileID, err := extractNSXObjectIDfromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppProfile, err := eGW.ReadLBAppProfileByID(lbAppProfileID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application profile with ID (%s) after creation: %s",
			readAppProfile.ID, err)
	}
	return readAppProfile, nil
}

// ReadLBAppProfile is able to find the types.LBAppProfile type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) ReadLBAppProfile(lbAppProfileConfig *types.LBAppProfile) (*types.LBAppProfile, error) {
	if err := validateReadLBAppProfile(lbAppProfileConfig); err != nil {
		return nil, err
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBAppProfilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap response
	lbAppProfileResponse := &struct {
		LbAppProfiles []*types.LBAppProfile `xml:"applicationProfile"`
	}{}

	// This query returns all application profiles as the API does not have filtering options
	_, err = eGW.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer application profile: %s", nil, lbAppProfileResponse)
	if err != nil {
		return nil, err
	}

	// Search for application profile by ID or by Name
	for _, profile := range lbAppProfileResponse.LbAppProfiles {
		// If ID was specified for lookup - look for the same ID
		if lbAppProfileConfig.ID != "" && profile.ID == lbAppProfileConfig.ID {
			return profile, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbAppProfileConfig.Name != "" && profile.Name == lbAppProfileConfig.Name {
			// We found it by name. Let's verify if search ID was specified and it matches the lookup object
			if lbAppProfileConfig.ID != "" && profile.ID != lbAppProfileConfig.ID {
				return nil, fmt.Errorf("load balancer application profile was found by name (%s)"+
					", but its ID (%s) does not match specified ID (%s)",
					profile.Name, profile.ID, lbAppProfileConfig.ID)
			}
			return profile, nil
		}
	}

	return nil, fmt.Errorf("could not find load balancer application profile (name: %s, ID: %s)",
		lbAppProfileConfig.Name, lbAppProfileConfig.ID)
}

// ReadLBAppProfileByID wraps ReadLBAppProfile and needs only an ID for lookup
func (eGW *EdgeGateway) ReadLBAppProfileByID(id string) (*types.LBAppProfile, error) {
	return eGW.ReadLBAppProfile(&types.LBAppProfile{ID: id})
}

// ReadLBAppProfileByName wraps ReadLBAppProfile and needs only a Name for lookup
func (eGW *EdgeGateway) ReadLBAppProfileByName(name string) (*types.LBAppProfile, error) {
	return eGW.ReadLBAppProfile(&types.LBAppProfile{Name: name})
}

// UpdateLBAppProfile updates types.LBAppProfile with all fields. At least name or ID must be specified.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) UpdateLBAppProfile(lbAppProfileConfig *types.LBAppProfile) (*types.LBAppProfile, error) {
	err := validateUpdateLBAppProfile(lbAppProfileConfig)
	if err != nil {
		return nil, err
	}

	lbAppProfileConfig.ID, err = eGW.getLBAppProfileIDByNameID(lbAppProfileConfig.Name, lbAppProfileConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer application profile: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBAppProfilePath + lbAppProfileConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = eGW.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer application profile : %s", lbAppProfileConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readAppProfile, err := eGW.ReadLBAppProfileByID(lbAppProfileConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application profile with ID (%s) after update: %s",
			readAppProfile.ID, err)
	}
	return readAppProfile, nil
}

// DeleteLBAppProfile is able to delete the types.LBAppProfile type by Name and/or ID.
// If both - Name and ID are specified it performs a lookup by ID and returns an error if the specified name and found
// name do not match.
func (eGW *EdgeGateway) DeleteLBAppProfile(lbAppProfileConfig *types.LBAppProfile) error {
	err := validateDeleteLBAppProfile(lbAppProfileConfig)
	if err != nil {
		return err
	}

	lbAppProfileConfig.ID, err = eGW.getLBAppProfileIDByNameID(lbAppProfileConfig.Name, lbAppProfileConfig.ID)
	if err != nil {
		return fmt.Errorf("cannot delete load balancer application profile: %s", err)
	}

	httpPath, err := eGW.buildProxiedEdgeEndpointURL(types.LBAppProfilePath + lbAppProfileConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	return eGW.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, "application/xml",
		"unable to delete application profile: %s", nil)
}

// DeleteLBAppProfileByID wraps DeleteLBAppProfile and requires only ID for deletion
func (eGW *EdgeGateway) DeleteLBAppProfileByID(id string) error {
	return eGW.DeleteLBAppProfile(&types.LBAppProfile{ID: id})
}

// DeleteLBAppProfileByName wraps DeleteLBAppProfile and requires only Name for deletion
func (eGW *EdgeGateway) DeleteLBAppProfileByName(name string) error {
	return eGW.DeleteLBAppProfile(&types.LBAppProfile{Name: name})
}

func validateCreateLBAppProfile(lbAppProfileConfig *types.LBAppProfile) error {
	if lbAppProfileConfig.Name == "" {
		return fmt.Errorf("load balancer application profile Name cannot be empty")
	}

	return nil
}

func validateReadLBAppProfile(lbAppProfileConfig *types.LBAppProfile) error {
	if lbAppProfileConfig.ID == "" && lbAppProfileConfig.Name == "" {
		return fmt.Errorf("to read load balancer application profile at least one of `ID`, `Name`" +
			" fields must be specified")
	}

	return nil
}

func validateUpdateLBAppProfile(lbAppProfileConfig *types.LBAppProfile) error {
	// Update and create have the same requirements for now
	return validateCreateLBAppProfile(lbAppProfileConfig)
}

func validateDeleteLBAppProfile(lbAppProfileConfig *types.LBAppProfile) error {
	// Read and delete have the same requirements for now
	return validateReadLBAppProfile(lbAppProfileConfig)
}

// getLBAppProfileIDByNameID checks if at least name or ID is set and returns the ID.
// If the ID is specified - it passes through the ID. If only name was specified
// it will lookup the object by name and return the ID.
func (eGW *EdgeGateway) getLBAppProfileIDByNameID(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or ID must be specific to find load balancer "+
			"application profile got name (%s) ID (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, ID must be found, because only ID can be used in request path
	readlbAppProfile, err := eGW.ReadLBAppProfileByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer application profile by name: %s", err)
	}
	return readlbAppProfile.ID, nil
}
