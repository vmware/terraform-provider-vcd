/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// OrgGroup defines group structure
type OrgGroup struct {
	Group    *types.Group
	client   *Client
	AdminOrg *AdminOrg // needed to be able to update, as the list of roles is found in the Org
}

// NewGroup creates a new group structure which still needs to have Group attribute populated
func NewGroup(cli *Client, org *AdminOrg) *OrgGroup {
	return &OrgGroup{
		Group:    new(types.Group),
		client:   cli,
		AdminOrg: org,
	}
}

// CreateGroup creates a group in Org. Supported provider types are `OrgUserProviderIntegrated` and
// `OrgUserProviderSAML`.
//
// Note. This request will return HTTP 403 if Org is not configured for SAML or LDAP usage.
func (adminOrg *AdminOrg) CreateGroup(group *types.Group) (*OrgGroup, error) {
	if err := validateCreateUpdateGroup(group); err != nil {
		return nil, err
	}

	groupCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parsing admin org url: %s", err)
	}
	groupCreateHREF.Path += "/groups"

	grpgroup := NewGroup(adminOrg.client, adminOrg)
	// Add default XML types
	group.Xmlns = types.XMLNamespaceVCloud
	group.Type = types.MimeAdminGroup

	_, err = adminOrg.client.ExecuteRequest(groupCreateHREF.String(), http.MethodPost,
		types.MimeAdminGroup, "error creating group: %s", group, grpgroup.Group)
	if err != nil {
		return nil, err
	}

	return grpgroup, nil
}

// GetGroupByHref retrieves group by HREF
func (adminOrg *AdminOrg) GetGroupByHref(href string) (*OrgGroup, error) {
	orgGroup := NewGroup(adminOrg.client, adminOrg)

	_, err := adminOrg.client.ExecuteRequest(href, http.MethodGet,
		types.MimeAdminUser, "error getting group: %s", nil, orgGroup.Group)

	if err != nil {
		return nil, err
	}
	return orgGroup, nil
}

// GetGroupByName retrieves group by Name
func (adminOrg *AdminOrg) GetGroupByName(name string, refresh bool) (*OrgGroup, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, group := range adminOrg.AdminOrg.Groups.Group {
		if group.Name == name {
			return adminOrg.GetGroupByHref(group.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetGroupById retrieves group by Id
func (adminOrg *AdminOrg) GetGroupById(id string, refresh bool) (*OrgGroup, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, group := range adminOrg.AdminOrg.Groups.Group {
		if group.ID == id {
			return adminOrg.GetGroupByHref(group.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetGroupByNameOrId retrieves group by Name or Id. Id is prioritized for search
func (adminOrg *AdminOrg) GetGroupByNameOrId(identifier string, refresh bool) (*OrgGroup, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetGroupByName(name, refresh) }
	getById := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetGroupById(name, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*OrgGroup), err
}

// Update allows to update group. vCD API allows to update only role
func (group *OrgGroup) Update() error {
	util.Logger.Printf("[TRACE] Updating group: %s", group.Group.Name)

	if err := validateCreateUpdateGroup(group.Group); err != nil {
		return err
	}

	groupHREF, err := url.ParseRequestURI(group.Group.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for group %s : %s", group.Group.Href, err)
	}
	util.Logger.Printf("[TRACE] Url for updating group : %s and name: %s", groupHREF.String(), group.Group.Name)

	_, err = group.client.ExecuteRequest(groupHREF.String(), http.MethodPut,
		types.MimeAdminGroup, "error updating group : %s", group.Group, nil)
	return err
}

// Delete removes a group
func (group *OrgGroup) Delete() error {
	if err := validateDeleteGroup(group.Group); err != nil {
		return err
	}

	groupHREF, err := url.ParseRequestURI(group.Group.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for group %s : %s", group.Group.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for deleting group : %s and name: %s", groupHREF, group.Group.Name)

	return group.client.ExecuteRequestWithoutResponse(groupHREF.String(), http.MethodDelete,
		types.MimeAdminGroup, "error deleting group : %s", nil)
}

// validateCreateGroup checks if mandatory fields are set for group creation and update
func validateCreateUpdateGroup(group *types.Group) error {
	if group == nil {
		return fmt.Errorf("group cannot be nil")
	}

	if group.Name == "" {
		return fmt.Errorf("group must have a name")
	}

	if group.ProviderType == "" {
		return fmt.Errorf("group must have provider type set")
	}

	if group.Role.HREF == "" {
		return fmt.Errorf("group role must have HREF set")
	}
	return nil
}

// validateDeleteGroup checks if mandatory fields are set for delete
func validateDeleteGroup(group *types.Group) error {
	if group == nil {
		return fmt.Errorf("group cannot be nil")
	}

	if group.Href == "" {
		return fmt.Errorf("HREF must be set to delete group")
	}

	return nil
}
