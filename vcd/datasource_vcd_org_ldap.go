package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceLdapUserAttributes defines the elements of types.OrgLdapUserAttributes
// The field names are the ones used in the GUI, with a comment to indicate which structure field each one corresponds to
var datasourceLdapUserAttributes = &schema.Schema{
	Type:        schema.TypeList,
	Computed:    true,
	Description: "Custom settings when `ldap_mode` is CUSTOM",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"object_class": { // ObjectClass
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP objectClass of which imported users are members. For example, user or person",
			},
			"unique_identifier": { // ObjectIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use as the unique identifier for a user. For example, objectGuid",
			},
			"user_name": { // Username
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use when looking up a user name to import. For example, userPrincipalName or samAccountName",
			},
			"email": { // Email
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the user's email address. For example, mail",
			},
			"full_name": { // FullName
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the user's full name. For example, displayName",
			},
			"given_name": { // GivenName
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the user's given name. For example, givenName",
			},
			"surname": { // Surname
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the user's surname. For example, sn",
			},
			"telephone": { // Telephone
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the user's telephone number. For example, telephoneNumber",
			},
			"group_membership_identifier": { // GroupMembershipIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute that identifies a user as a member of a group. For example, dn",
			},
			"group_back_link_identifier": { // GroupBackLinkIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute that returns the identifiers of all the groups of which the user is a member",
			},
		},
	},
}

// datasourceLdapGroupAttributes defines the elements of types.OrgLdapGroupAttributes
// The field names are the ones used in the GUI, with a comment to indicate which structure field each one corresponds to
var datasourceLdapGroupAttributes = &schema.Schema{
	Type:        schema.TypeList,
	Computed:    true,
	Description: "Custom settings when `ldap_mode` is CUSTOM",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"object_class": { // ObjectClass
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP objectClass of which imported groups are members. For example, group",
			},
			"unique_identifier": { // ObjectIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use as the unique identifier for a group. For example, objectGuid",
			},
			"name": { // GroupName
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use for the group name. For example, cn",
			},
			"membership": { // Membership
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute to use when getting the members of a group. For example, member",
			},
			"group_membership_identifier": { // MembershipIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP attribute that identifies a group as a member of another group. For example, dn",
			},
			"group_back_link_identifier": { // BackLinkIdentifier
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP group attribute used to identify a group member",
			},
		},
	},
}

func datasourceVcdOrgLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgLdapRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization name",
			},
			"ldap_mode": { // OrgLdapMode
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of LDAP settings (one of NONE, SYSTEM, CUSTOM)",
			},
			"custom_settings": { // CustomOrgLdapSettings
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Custom settings when `ldap_mode` is CUSTOM",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": { // Hostname
							Type:        schema.TypeString,
							Computed:    true,
							Description: "host name or IP of the LDAP server",
						},
						"port": { // Port
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port number for LDAP service",
						},
						"authentication_method": { // AuthenticationMechanism
							Type:        schema.TypeString,
							Computed:    true,
							Description: "authentication method: one of SIMPLE, MD5DIGEST, NTLM",
						},
						"connector_type": { // ConnectorType
							Type:        schema.TypeString,
							Computed:    true,
							Description: "type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
						},
						"base_distinguished_name": { //SearchBase
							Type:        schema.TypeString,
							Computed:    true,
							Description: "LDAP search base",
						},
						"is_ssl": { // IsSsl
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the LDAP service requires an SSL connection",
						},
						"username": { // Username
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value pairs (for example: cn="ldap-admin", c="example", dc="com")`,
						},
						"password": { // Password
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Password for the user identified by UserName. This value is never returned by GET. It is inspected on create and modify. On modify, the absence of this element indicates that the password should not be changed`,
						},
						"user_attributes":  datasourceLdapUserAttributes,
						"group_attributes": datasourceLdapGroupAttributes,
					},
				},
			},
		},
	}
}

func datasourceVcdOrgLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgLdapRead(ctx, d, meta, "datasource")
}
