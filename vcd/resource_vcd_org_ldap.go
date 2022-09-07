package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// resourceLdapUserAttributes defines the elements of types.OrgLdapUserAttributes
// The field names are the ones used in the GUI, with a comment to indicate which API field each one corresponds to
var resourceLdapUserAttributes = &schema.Schema{
	Type:        schema.TypeList,
	Required:    true,
	MaxItems:    1,
	Description: "Custom user settings when `ldap_mode` is CUSTOM",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"object_class": { // ObjectClass
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP objectClass of which imported users are members. For example, user or person",
			},
			"unique_identifier": { // ObjectIdentifier
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use as the unique identifier for a user. For example, objectGuid",
			},
			"user_name": { // Username
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use when looking up a user name to import. For example, userPrincipalName or samAccountName",
			},
			"email": { // Email
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the user's email address. For example, mail",
			},
			"full_name": { // FullName
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the user's full name. For example, displayName",
			},
			"given_name": { // GivenName
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the user's given name. For example, givenName",
			},
			"surname": { // Surname
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the user's surname. For example, sn",
			},
			"telephone": { // Telephone
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the user's telephone number. For example, telephoneNumber",
			},
			"group_membership_identifier": { // GroupMembershipIdentifier
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute that identifies a user as a member of a group. For example, dn",
			},
			"group_back_link_identifier": { // GroupBackLinkIdentifier
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LDAP attribute that returns the identifiers of all the groups of which the user is a member",
			},
		},
	},
}

// resourceLdapGroupAttributes defines the elements of types.OrgLdapGroupAttributes
// The field names are the ones used in the GUI, with a comment to indicate which API field each one corresponds to
var resourceLdapGroupAttributes = &schema.Schema{
	Type:        schema.TypeList,
	Required:    true,
	MaxItems:    1,
	Description: "Custom group settings when `ldap_mode` is CUSTOM",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"object_class": { // ObjectClass
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP objectClass of which imported groups are members. For example, group",
			},
			"unique_identifier": { // ObjectIdentifier
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use as the unique identifier for a group. For example, objectGuid",
			},
			"name": { // GroupName
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use for the group name. For example, cn",
			},
			"membership": { // Membership
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute to use when getting the members of a group. For example, member",
			},
			"group_membership_identifier": { // MembershipIdentifier
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP attribute that identifies a group as a member of another group. For example, dn",
			},
			"group_back_link_identifier": { // BackLinkIdentifier
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LDAP group attribute used to identify a group member",
			},
		},
	},
}

// resourceVcdOrgLdap defines types.OrgLdapSettingsType
// The field names are the ones used in the GUI, with a comment to indicate which API field each one corresponds to
func resourceVcdOrgLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdOrgLdapRead,
		CreateContext: resourceVcdOrgLdapCreate,
		UpdateContext: resourceVcdOrgLdapUpdate,
		DeleteContext: resourceVcdOrgLdapDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization name",
			},
			"ldap_mode": { // OrgLdapMode
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of LDAP settings (one of NONE, SYSTEM, CUSTOM)",
				ValidateFunc: validation.StringInSlice([]string{types.LdapModeNone, types.LdapModeSystem, types.LdapModeCustom}, false),
			},
			"custom_settings": { // CustomOrgLdapSettings
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Custom settings when `ldap_mode` is CUSTOM",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": { // HostName
							Type:        schema.TypeString,
							Required:    true,
							Description: "host name or IP of the LDAP server",
						},
						"port": { // Port
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port number for LDAP service",
						},
						"authentication_method": { // AuthenticationMechanism
							Type:         schema.TypeString,
							Required:     true,
							Description:  "authentication method: one of SIMPLE, MD5DIGEST, NTLM",
							ValidateFunc: validation.StringInSlice([]string{"SIMPLE", "MD5DIGEST", "NTLM"}, false),
						},
						"connector_type": { // ConnectorType
							Type:         schema.TypeString,
							Required:     true,
							Description:  "type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
							ValidateFunc: validation.StringInSlice([]string{"OPEN_LDAP", "ACTIVE_DIRECTORY"}, false),
						},
						"base_distinguished_name": { //SearchBase
							Type:        schema.TypeString,
							Optional:    true,
							Description: "LDAP search base",
						},
						"is_ssl": { // IsSsl
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "True if the LDAP service requires an SSL connection",
						},
						"username": { // Username
							Type:        schema.TypeString,
							Optional:    true,
							Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value pairs (for example: cn="ldap-admin", c="example", dc="com")`,
						},
						"password": { // Password
							Type:        schema.TypeString,
							Optional:    true,
							Description: `Password for the user identified by UserName. This value is never returned by GET. It is inspected on create and modify. On modify, the absence of this element indicates that the password should not be changed`,
						},
						"user_attributes":  resourceLdapUserAttributes,  // CustomOrgLdapSettings.UserAttributes
						"group_attributes": resourceLdapGroupAttributes, // CustomOrgLdapSettings.GroupAttributes
					},
				},
			},
		},
	}
}

func resourceVcdOrgLdapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("name").(string)

	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return diag.Errorf("[Org LDAP read] error searching for Org %s: %s", orgName, err)
	}

	var settings types.OrgLdapSettingsType
	newSettings, err := adminOrg.LdapConfigure(&settings)
	if err != nil {
		return diag.Errorf("[Org LDAP create] error setting org '%s' LDAP configuration: %s", orgName, err)
	}
	err = validateLdapSettings(&settings, newSettings)
	if err != nil {
		return diag.Errorf("[Org LDAP create] error validating LDAP settings: %s", err)
	}
	return resourceVcdOrgLdapRead(ctx, d, meta)
}

func resourceVcdOrgLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgLdapRead(ctx, d, meta, "resource")
}

func genericVcdOrgLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("name").(string)

	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find Organization %s LDAP settings: %s. Removing from state", orgName, err)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("unable to find organization %s: %s", orgName, err)
	}

	config, err := adminOrg.GetLdapConfiguration()
	if err != nil {
		d.SetId("")
		return diag.Errorf("[Org LDAP read] error getting LDAP settings for Org %s: %s", orgName, err)
	}

	dSet(d, "name", orgName)
	dSet(d, "ldap_mode", config.OrgLdapMode)
	d.SetId(adminOrg.AdminOrg.ID)

	if config.OrgLdapMode == "CUSTOM" {
		customSettings := map[string]interface{}{
			"server":                  config.CustomOrgLdapSettings.HostName,
			"port":                    config.CustomOrgLdapSettings.Port,
			"authentication_method":   config.CustomOrgLdapSettings.AuthenticationMechanism,
			"connector_type":          config.CustomOrgLdapSettings.ConnectorType,
			"base_distinguished_name": config.CustomOrgLdapSettings.SearchBase,
			"is_ssl":                  config.CustomOrgLdapSettings.IsSsl,
			"username":                config.CustomOrgLdapSettings.Username,
			"password":                config.CustomOrgLdapSettings.Password,
			"user_attributes": []map[string]interface{}{
				{
					"object_class":                config.CustomOrgLdapSettings.UserAttributes.ObjectClass,
					"unique_identifier":           config.CustomOrgLdapSettings.UserAttributes.ObjectIdentifier,
					"user_name":                   config.CustomOrgLdapSettings.UserAttributes.Username,
					"email":                       config.CustomOrgLdapSettings.UserAttributes.Email,
					"full_name":                   config.CustomOrgLdapSettings.UserAttributes.FullName,
					"given_name":                  config.CustomOrgLdapSettings.UserAttributes.GivenName,
					"surname":                     config.CustomOrgLdapSettings.UserAttributes.Surname,
					"telephone":                   config.CustomOrgLdapSettings.UserAttributes.Telephone,
					"group_membership_identifier": config.CustomOrgLdapSettings.UserAttributes.GroupMembershipIdentifier,
					"group_back_link_identifier":  config.CustomOrgLdapSettings.UserAttributes.GroupBackLinkIdentifier,
				},
			},
			"group_attributes": []map[string]interface{}{
				{
					"object_class":                config.CustomOrgLdapSettings.GroupAttributes.ObjectClass,
					"unique_identifier":           config.CustomOrgLdapSettings.GroupAttributes.ObjectIdentifier,
					"name":                        config.CustomOrgLdapSettings.GroupAttributes.GroupName,
					"membership":                  config.CustomOrgLdapSettings.GroupAttributes.Membership,
					"group_membership_identifier": config.CustomOrgLdapSettings.GroupAttributes.MembershipIdentifier,
					"group_back_link_identifier":  config.CustomOrgLdapSettings.GroupAttributes.BackLinkIdentifier,
				},
			},
		}
		err = d.Set("custom_settings", []map[string]interface{}{customSettings})
		if err != nil {
			return diag.Errorf("[Org LDAP read] error setting 'user_attributes' field: %s", err)
		}
	}
	return nil
}

func resourceVcdOrgLdapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("[Org LDAP update] function not yet implemented")
}

func resourceVcdOrgLdapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("name").(string)

	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return diag.Errorf("[Org LDAP delete] error searching for Org %s: %s", orgName, err)
	}
	return diag.FromErr(adminOrg.LdapDisable())
}

func validateLdapSettings(wantedSettings, retrievedSettings *types.OrgLdapSettingsType) error {

	return nil
}
