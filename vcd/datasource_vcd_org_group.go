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
				Required:    true,
				Description: "Organization name for lookup",
			},
			"name": {
				Type:     schema.TypeString,
				Required:    true,
				Description: "Name of the group for lookup",
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
	return nil
}
