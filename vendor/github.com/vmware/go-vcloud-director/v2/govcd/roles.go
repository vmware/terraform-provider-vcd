package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type OpenApiRole struct {
	Role   *types.Role
	client *Client
}

// GetOpenApiRoleById retrieves role by given ID
func (adminOrg *AdminOrg) GetOpenApiRoleById(id string) (*OpenApiRole, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty role id")
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	role := &OpenApiRole{
		Role:   &types.Role{},
		client: adminOrg.client,
	}

	err = adminOrg.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, role.Role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetAllOpenApiRoles retrieves all roles using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func (adminOrg *AdminOrg) GetAllOpenApiRoles(queryParameters url.Values) ([]*types.Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.Role{{}}

	err = adminOrg.client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

// Create creates a new role using OpenAPI endpoint
func (role *OpenApiRole) Create(newRole *types.Role) (*OpenApiRole, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnRole := &OpenApiRole{
		Role:   &types.Role{},
		client: role.client,
	}

	err = role.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newRole, returnRole.Role)
	if err != nil {
		return nil, fmt.Errorf("error creating role: %s", err)
	}

	return returnRole, nil
}

// Update updates existing OpenAPI role
func (role *OpenApiRole) Update() (*OpenApiRole, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if role.Role.ID == "" {
		return nil, fmt.Errorf("cannot update role without id")
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint, role.Role.ID)
	if err != nil {
		return nil, err
	}

	returnRole := &OpenApiRole{
		Role:   &types.Role{},
		client: role.client,
	}

	err = role.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, role.Role, returnRole.Role)
	if err != nil {
		return nil, fmt.Errorf("error updating role: %s", err)
	}

	return returnRole, nil
}

// Delete deletes OpenAPI role
func (role *OpenApiRole) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if role.Role.ID == "" {
		return fmt.Errorf("cannot delete role without id")
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint, role.Role.ID)
	if err != nil {
		return err
	}

	err = role.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting role: %s", err)
	}

	return nil
}
