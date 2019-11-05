/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Definition of an OrgUser
type OrgUser struct {
	User     *types.User
	client   *Client
	AdminOrg *AdminOrg // needed to be able to update, as the list of roles is found in the Org
}

// Simplified structure to insert or modify an organization user
type OrgUserConfiguration struct {
	Name            string // Mandatory
	Password        string // Mandatory
	RoleName        string // Mandatory
	ProviderType    string // Optional: defaults to "INTEGRATED"
	IsEnabled       bool   // Optional: defaults to false
	IsLocked        bool   // Only used for updates
	DeployedVmQuota int    // Optional: 0 means "unlimited"
	StoredVmQuota   int    // Optional: 0 means "unlimited"
	FullName        string // Optional
	Description     string // Optional
	EmailAddress    string // Optional
	Telephone       string // Optional
	IM              string // Optional
}

const (
	// Common role names and provider types are kept here to reduce hard-coded text and prevent mistakes
	// Roles that are added to the organization need to be entered as free text

	OrgUserRoleOrganizationAdministrator = "Organization Administrator"
	OrgUserRoleCatalogAuthor             = "Catalog Author"
	OrgUserRoleVappAuthor                = "vApp Author"
	OrgUserRoleVappUser                  = "vApp User"
	OrgUserRoleConsoleAccessOnly         = "Console Access Only"
	OrgUserRoleDeferToIdentityProvider   = "Defer to Identity Provider"

	// Allowed values for provider types
	OrgUserProviderIntegrated = "INTEGRATED" // The user is created locally or imported from LDAP
	OrgUserProviderSAML       = "SAML"       // The user is imported from a SAML identity provider.
	OrgUserProviderOAUTH      = "OAUTH"      // The user is imported from an OAUTH identity provider
)

// Used to check the validity of provider type on creation
var OrgUserProviderTypes = []string{
	OrgUserProviderIntegrated,
	OrgUserProviderSAML,
	OrgUserProviderOAUTH,
}

// NewUser creates an empty user
func NewUser(cli *Client, org *AdminOrg) *OrgUser {
	return &OrgUser{
		User:     new(types.User),
		client:   cli,
		AdminOrg: org,
	}
}

// FetchUserByHref returns a user by its HREF
// Deprecated: use GetUserByHref instead
func (adminOrg *AdminOrg) FetchUserByHref(href string) (*OrgUser, error) {
	return adminOrg.GetUserByHref(href)
}

// FetchUserByName returns a user by its Name
// Deprecated: use GetUserByName instead
func (adminOrg *AdminOrg) FetchUserByName(name string, refresh bool) (*OrgUser, error) {
	return adminOrg.GetUserByName(name, refresh)
}

// FetchUserById returns a user by its ID
// Deprecated: use GetUserById instead
func (adminOrg *AdminOrg) FetchUserById(id string, refresh bool) (*OrgUser, error) {
	return adminOrg.GetUserById(id, refresh)
}

// FetchUserById returns a user by its Name or ID
// Deprecated: use GetUserByNameOrId instead
func (adminOrg *AdminOrg) FetchUserByNameOrId(identifier string, refresh bool) (*OrgUser, error) {
	return adminOrg.GetUserByNameOrId(identifier, refresh)
}

// GetUserByHref returns a user by its HREF, without need for
// searching in the adminOrg user list
func (adminOrg *AdminOrg) GetUserByHref(href string) (*OrgUser, error) {
	orgUser := NewUser(adminOrg.client, adminOrg)

	_, err := adminOrg.client.ExecuteRequest(href, http.MethodGet,
		types.MimeAdminUser, "error getting user: %s", nil, orgUser.User)

	if err != nil {
		return nil, err
	}
	return orgUser, nil
}

// GetUserByName retrieves a user within an admin organization by name
// Returns a valid user if it exists. If it doesn't, returns nil and ErrorEntityNotFound
// If argument refresh is true, the AdminOrg will be refreshed before searching.
// This is usually done after creating, modifying, or deleting users.
// If it is false, it will search within the data already in memory (useful when
// looping through the users and we know that no changes have occurred in the meantime)
func (adminOrg *AdminOrg) GetUserByName(name string, refresh bool) (*OrgUser, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, user := range adminOrg.AdminOrg.Users.User {
		if user.Name == name {
			return adminOrg.GetUserByHref(user.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetUserById retrieves a user within an admin organization by ID
// Returns a valid user if it exists. If it doesn't, returns nil and ErrorEntityNotFound
// If argument refresh is true, the AdminOrg will be refreshed before searching.
// This is usually done after creating, modifying, or deleting users.
// If it is false, it will search within the data already in memory (useful when
// looping through the users and we know that no changes have occurred in the meantime)
func (adminOrg *AdminOrg) GetUserById(id string, refresh bool) (*OrgUser, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, user := range adminOrg.AdminOrg.Users.User {
		if user.ID == id {
			return adminOrg.GetUserByHref(user.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetUserByNameOrId retrieves a user within an admin organization
// by either name or ID
// Returns a valid user if it exists. If it doesn't, returns nil and ErrorEntityNotFound
// If argument refresh is true, the AdminOrg will be refreshed before searching.
// This is usually done after creating, modifying, or deleting users.
// If it is false, it will search within the data already in memory (useful when
// looping through the users and we know that no changes have occurred in the meantime)
func (adminOrg *AdminOrg) GetUserByNameOrId(identifier string, refresh bool) (*OrgUser, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetUserByName(name, refresh) }
	getById := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetUserById(name, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*OrgUser), err
}

// GetRole finds a role within the organization
// Deprecated: use GetRoleReference
func (adminOrg *AdminOrg) GetRole(roleName string) (*types.Reference, error) {
	return adminOrg.GetRoleReference(roleName)
}

// GetRoleReference finds a role within the organization
func (adminOrg *AdminOrg) GetRoleReference(roleName string) (*types.Reference, error) {

	// There is no need to refresh the AdminOrg, until we implement CRUD for roles
	for _, role := range adminOrg.AdminOrg.RoleReferences.RoleReference {
		if role.Name == roleName {
			return role, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// Retrieves a user within the boundaries of MaxRetryTimeout
func retrieveUserWithTimeout(adminOrg *AdminOrg, userName string) (*OrgUser, error) {

	// Attempting to retrieve the user
	delayPerAttempt := 200 * time.Millisecond
	maxOperationTimeout := time.Duration(adminOrg.client.MaxRetryTimeout) * time.Second

	// We make sure that the timeout is never less than 2 seconds
	if maxOperationTimeout < 2*time.Second {
		maxOperationTimeout = 2 * time.Second
	}

	// If maxRetryTimeout is set to a higher limit, we lower it to match the
	// expectations for this operation. If the user is not created within 10 seconds,
	// there is no need to wait for more. Usually, the operation lasts between 200ms and 900ms
	if maxOperationTimeout > 10*time.Second {
		maxOperationTimeout = 10 * time.Second
	}

	startTime := time.Now()
	elapsed := time.Since(startTime)
	var newUser *OrgUser
	var err error
	for elapsed < maxOperationTimeout {
		newUser, err = adminOrg.GetUserByName(userName, true)
		if err == nil {
			break
		}
		time.Sleep(delayPerAttempt)
		elapsed = time.Since(startTime)
	}

	elapsed = time.Since(startTime)

	// If the user was not retrieved within the allocated time, we inform the user about the failure
	// and the time it occurred to get to this point, so that they may try with a longer time
	if err != nil {
		return nil, fmt.Errorf("failure to retrieve a new user after %s : %s", elapsed, err)
	}

	return newUser, nil
}

// CreateUser creates an OrgUser from a full configuration structure
// The timeOut variable is the maximum time we wait for the user to be ready
// (This operation does not return a task)
// This function returns as soon as the user has been created, which could be as
// little as 200ms or as much as Client.MaxRetryTimeout
// Mandatory fields are: Name, Role, Password.
// https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/POST-CreateUser.html
func (adminOrg *AdminOrg) CreateUser(userConfiguration *types.User) (*OrgUser, error) {
	err := validateUserForCreation(userConfiguration)
	if err != nil {
		return nil, err
	}

	userCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parsing admin org url: %s", err)
	}
	userCreateHREF.Path += "/users"

	user := NewUser(adminOrg.client, adminOrg)

	_, err = adminOrg.client.ExecuteRequest(userCreateHREF.String(), http.MethodPost,
		types.MimeAdminUser, "error creating user: %s", userConfiguration, user.User)
	if err != nil {
		return nil, err
	}

	// If there is a valid task, we try to follow through
	// A valid task exists if the Task object in the user structure
	// is not nil and contains at least a task
	if user.User.Tasks != nil && len(user.User.Tasks.Task) > 0 {
		task := NewTask(adminOrg.client)
		task.Task = user.User.Tasks.Task[0]
		err = task.WaitTaskCompletion()

		if err != nil {
			return nil, err
		}
	}

	return retrieveUserWithTimeout(adminOrg, userConfiguration.Name)
}

// CreateUserSimple creates an org user from a simplified structure
func (adminOrg *AdminOrg) CreateUserSimple(userData OrgUserConfiguration) (*OrgUser, error) {

	if userData.Name == "" {
		return nil, fmt.Errorf("name is mandatory to create a user")
	}
	if userData.Password == "" {
		return nil, fmt.Errorf("password is mandatory to create a user")
	}
	if userData.RoleName == "" {
		return nil, fmt.Errorf("role is mandatory to create a user")
	}
	role, err := adminOrg.GetRoleReference(userData.RoleName)
	if err != nil {
		return nil, fmt.Errorf("error finding a role named %s", userData.RoleName)
	}

	var userConfiguration = types.User{
		Xmlns:           types.XMLNamespaceVCloud,
		Type:            types.MimeAdminUser,
		ProviderType:    userData.ProviderType,
		Name:            userData.Name,
		IsEnabled:       userData.IsEnabled,
		Password:        userData.Password,
		DeployedVmQuota: userData.DeployedVmQuota,
		StoredVmQuota:   userData.StoredVmQuota,
		FullName:        userData.FullName,
		EmailAddress:    userData.EmailAddress,
		Description:     userData.Description,
		Telephone:       userData.Telephone,
		IM:              userData.IM,
		Role:            &types.Reference{HREF: role.HREF},
	}

	// ShowUser(userConfiguration)
	return adminOrg.CreateUser(&userConfiguration)
}

// GetRoleName retrieves the name of the role currently assigned to the user
func (user *OrgUser) GetRoleName() string {
	if user.User.Role == nil {
		return ""
	}
	return user.User.Role.Name
}

// Delete removes the user, returning an error if the call fails.
// if requested, it will attempt to take ownership before the removal.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/DELETE-User.html
// Note: in the GUI we need to disable the user before deleting.
// There is no such constraint with the API.
//
// Expected behaviour:
// with takeOwnership = true, all entities owned by the user being deleted will be transferred to the caller.
// with takeOwnership = false, if the user own catalogs, networks, or running VMs/vApps, the call will fail.
//                             If the user owns only powered-off VMs/vApps, the call will succeeds and the
//                             VMs/vApps will be removed.
func (user *OrgUser) Delete(takeOwnership bool) error {
	util.Logger.Printf("[TRACE] Deleting user: %#v (take ownership: %v)", user.User.Name, takeOwnership)

	if takeOwnership {
		err := user.TakeOwnership()
		if err != nil {
			return err
		}
	}

	userHREF, err := url.ParseRequestURI(user.User.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %s", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for deleting user : %#v and name: %s", userHREF, user.User.Name)

	return user.client.ExecuteRequestWithoutResponse(userHREF.String(), http.MethodDelete,
		types.MimeAdminUser, "error deleting user : %s", nil)
}

// UpdateSimple updates the user, using ALL the fields in userData structure
// returning an error if the call fails.
// Careful: DeployedVmQuota and StoredVmQuota use a `0` value to mean "unlimited"
func (user *OrgUser) UpdateSimple(userData OrgUserConfiguration) error {
	util.Logger.Printf("[TRACE] Updating user: %#v", user.User.Name)

	if userData.Name != "" {
		user.User.Name = userData.Name
	}
	if userData.ProviderType != "" {
		user.User.ProviderType = userData.ProviderType
	}
	if userData.Description != "" {
		user.User.Description = userData.Description
	}
	if userData.FullName != "" {
		user.User.FullName = userData.FullName
	}
	if userData.EmailAddress != "" {
		user.User.EmailAddress = userData.EmailAddress
	}
	if userData.Telephone != "" {
		user.User.Telephone = userData.Telephone
	}
	if userData.Password != "" {
		user.User.Password = userData.Password
	}
	user.User.StoredVmQuota = userData.StoredVmQuota
	user.User.DeployedVmQuota = userData.DeployedVmQuota
	user.User.IsEnabled = userData.IsEnabled
	user.User.IsLocked = userData.IsLocked

	if userData.RoleName != "" && user.User.Role != nil && user.User.Role.Name != userData.RoleName {
		newRole, err := user.AdminOrg.GetRoleReference(userData.RoleName)
		if err != nil {
			return err
		}
		user.User.Role = newRole
	}
	return user.Update()
}

// Update updates the user, using its own configuration data
// returning an error if the call fails.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/PUT-User.html
func (user *OrgUser) Update() error {
	util.Logger.Printf("[TRACE] Updating user: %s", user.User.Name)

	// Makes sure that GroupReferences is either properly filled or nil,
	// because otherwise vCD will complain that the payload is not well formatted when
	// the configuration contains a non-empty password.
	if user.User.GroupReferences != nil {
		if len(user.User.GroupReferences.GroupReference) == 0 {
			user.User.GroupReferences = nil
		}
	}

	userHREF, err := url.ParseRequestURI(user.User.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %s", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for updating user : %#v and name: %s", userHREF, user.User.Name)

	_, err = user.client.ExecuteRequest(userHREF.String(), http.MethodPut,
		types.MimeAdminUser, "error updating user : %s", user.User, nil)
	return err
}

// Disable disables a user, if it is enabled. Fails otherwise.
func (user *OrgUser) Disable() error {
	util.Logger.Printf("[TRACE] Disabling user: %s", user.User.Name)

	if !user.User.IsEnabled {
		return fmt.Errorf("user %s is already disabled", user.User.Name)
	}
	user.User.IsEnabled = false

	return user.Update()
}

// ChangePassword changes user's password
// Constraints: the password must be non-empty, with a minimum of 6 characters
func (user *OrgUser) ChangePassword(newPass string) error {
	util.Logger.Printf("[TRACE] Changing user's password user: %s", user.User.Name)

	user.User.Password = newPass

	return user.Update()
}

// Enable enables a user if it was disabled. Fails otherwise.
func (user *OrgUser) Enable() error {
	util.Logger.Printf("[TRACE] Enabling user: %s", user.User.Name)

	if user.User.IsEnabled {
		return fmt.Errorf("user %s is already enabled", user.User.Name)
	}
	user.User.IsEnabled = true

	return user.Update()
}

// Unlock unlocks a user that was locked out by the system.
// Note that there is no procedure to LOCK a user: it is locked by the system when it exceeds the number of
// unauthorized access attempts
func (user *OrgUser) Unlock() error {
	util.Logger.Printf("[TRACE] Unlocking user: %s", user.User.Name)

	if !user.User.IsLocked {
		return fmt.Errorf("user %s is not locked", user.User.Name)
	}
	user.User.IsLocked = false

	return user.Update()
}

// ChangeRole changes a user's role
// Fails is we try to set the same role as the current one.
// Also fails if the provided role name is not found.
func (user *OrgUser) ChangeRole(roleName string) error {
	util.Logger.Printf("[TRACE] Changing user's role: %s", user.User.Name)

	if roleName == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if user.User.Role != nil && user.User.Role.Name == roleName {
		return fmt.Errorf("new role is the same as current role")
	}

	newRole, err := user.AdminOrg.GetRoleReference(roleName)
	if err != nil {
		return err
	}
	user.User.Role = newRole

	return user.Update()
}

// TakeOwnership takes ownership of the user's objects.
// Ownership is transferred to the caller.
// This is a call to make before deleting. Calling user.DeleteTakeOwnership() will
// run TakeOwnership before the actual user removal.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/POST-TakeOwnership.html
func (user *OrgUser) TakeOwnership() error {
	util.Logger.Printf("[TRACE] Taking ownership from user: %s", user.User.Name)

	userHREF, err := url.ParseRequestURI(user.User.Href + "/action/takeOwnership")
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %s", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for taking ownership from user : %#v and name: %s", userHREF, user.User.Name)

	return user.client.ExecuteRequestWithoutResponse(userHREF.String(), http.MethodPost,
		types.MimeAdminUser, "error taking ownership from user : %s", nil)
}

// validateUserForInput makes sure that the minimum data
// needed for creating an org user has been included in the configuration
func validateUserForCreation(user *types.User) error {
	var missingField = "missing field %s"
	if user.Xmlns == "" {
		user.Xmlns = types.XMLNamespaceVCloud
	}
	if user.Type == "" {
		user.Type = types.MimeAdminUser
	}
	if user.Name == "" {
		return fmt.Errorf(missingField, "Name")
	}
	if user.Password == "" {
		return fmt.Errorf(missingField, "Password")
	}
	if user.ProviderType != "" {
		validProviderType := false
		for _, pt := range OrgUserProviderTypes {
			if user.ProviderType == pt {
				validProviderType = true
			}
		}
		if !validProviderType {
			return fmt.Errorf("'%s' is not a valid provider type", user.ProviderType)
		}
	}
	if user.Role.HREF == "" {
		return fmt.Errorf(missingField, "Role.HREF")
	}
	return nil
}
