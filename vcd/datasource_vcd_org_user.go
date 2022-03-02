package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdOrgUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgUserRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "user_id"},
				Description:  `User's name. Required if "user_id" is not set`,
			},
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "user_id"},
				Description:  `User's id. Required if "name" is not set`,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role within the organization",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's description",
			},
			"provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identity provider type for this this user. One of: 'INTEGRATED', 'SAML', 'OAUTH'. ",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's full name",
			},
			"email_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's email address",
			},
			"telephone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's telephone",
			},
			"instant_messaging": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's telephone",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the user is enabled and can log in.",
			},
			"is_group_role": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this user has a group role.",
			},
			"is_locked": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "If the user account has been locked due to too many invalid login attempts, the value " +
					"will change to true (only the system can lock the user). ",
			},
			"is_external": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If the user account was imported from an external resource, like an LDAP",
			},
			"deployed_vm_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Quota of vApps that this user can deploy. A value of 0 specifies an unlimited quota.",
			},
			"stored_vm_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Quota of vApps that this user can store. A value of 0 specifies an unlimited quota.",
			},
			"groups_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of group names that this user belongs to",
			},
		},
	}
}

func datasourceVcdOrgUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var identifier string
	name := d.Get("name").(string)
	id := d.Get("user_id").(string)

	if name != "" {
		identifier = name
	} else {
		identifier = id
	}

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[datasourceVcdOrgUserRead] error retrieving org : %s", err)
	}

	user, err := adminOrg.GetUserByNameOrId(identifier, false)
	if err != nil {
		return diag.Errorf("error retrieving user %s : %s", identifier, err)
	}

	dSet(d, "user_id", user.User.ID)
	err = setOrgUserData(d, user, adminOrg)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
