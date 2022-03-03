package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdOrgGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgGroupRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of organization to use, optional if defined at provider level",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the group to lookup",
			},
			"provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"users_list": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of user names that belong to the group",
			},
		},
	}
}

func datasourceVcdOrgGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	groupName := d.Get("name").(string)
	orgGroup, err := adminOrg.GetGroupByName(groupName, false)
	if err != nil {
		return diag.Errorf("error finding group with name %s: %s", groupName, err)
	}

	d.SetId(orgGroup.Group.ID)
	dSet(d, "provider_type", orgGroup.Group.ProviderType)
	dSet(d, "description", orgGroup.Group.Description)
	dSet(d, "role", orgGroup.Group.Role.Name)
	var users []string
	for _, userRef := range orgGroup.Group.UsersList.UserReference {
		users = append(users, userRef.Name)
	}
	err = d.Set("users_list", convertStringsTotTypeSet(users))
	if err != nil {
		return diag.Errorf("could not set users_list field: %s", err)
	}

	return nil
}
