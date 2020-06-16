/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// LdapConfigure allows to configure LDAP mode in use by the Org
func (adminOrg *AdminOrg) LdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	util.Logger.Printf("[DEBUG] Configuring LDAP mode for Org name %s", adminOrg.AdminOrg.Name)

	// Xmlns field is not mandatory when `types.OrgLdapSettingsType` is set as part of whole
	// `AdminOrg` structure but it must be set when directly updating LDAP. For that reason
	// `types.OrgLdapSettingsType` Xmlns struct tag has 'omitempty' set
	settings.Xmlns = types.XMLNamespaceVCloud

	href := adminOrg.AdminOrg.HREF + "/settings/ldap"
	_, err := adminOrg.client.ExecuteRequest(href, http.MethodPut, types.MimeOrgLdapSettings,
		"error updating LDAP settings: %s", settings, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating LDAP mode for Org name '%s': %s", adminOrg.AdminOrg.Name, err)
	}

	ldapSettings, err := adminOrg.GetLdapConfiguration()
	if err != nil {
		return nil, fmt.Errorf("error retrieving LDAP configuration:  %s", err)
	}

	return ldapSettings, nil
}

// LdapDisable wraps LdapConfigure to disable LDAP configuration for org
func (adminOrg *AdminOrg) LdapDisable() error {
	_, err := adminOrg.LdapConfigure(&types.OrgLdapSettingsType{OrgLdapMode: types.LdapModeNone})
	return err
}

// GetLdapConfiguration retrieves LDAP configuration structure
func (adminOrg *AdminOrg) GetLdapConfiguration() (*types.OrgLdapSettingsType, error) {
	util.Logger.Printf("[DEBUG] Reading LDAP configuration for Org name %s", adminOrg.AdminOrg.Name)

	ldapSettings := &types.OrgLdapSettingsType{}

	href := adminOrg.AdminOrg.HREF + "/settings/ldap"

	_, err := adminOrg.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgLdapSettings,
		"error getting LDAP settings: %s", nil, ldapSettings)

	if err != nil {
		return nil, err
	}

	return ldapSettings, nil
}
