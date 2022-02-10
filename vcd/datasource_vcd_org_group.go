package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
)

func datasourceVcdOrgGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgGroupRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "Name of the Organization group",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "Organization group ID",
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

	// get by ID when it's available
	var orgGroup *govcd.OrgGroup
	identifier := d.Get("id").(string)
	if identifier != "" {
		orgGroup, err = adminOrg.GetGroupById(identifier, false)
	} else if d.Get("name").(string) != "" {
		identifier = d.Get("name").(string)
		orgGroup, err = adminOrg.GetGroupByName(identifier, false)
	} else {
		return diag.Errorf("Id or Name value is missing %s", err)
	}

	if err != nil {
		return diag.Errorf("org group %s not found: %s", identifier, err)
	}

	log.Printf("Org group with name %s found", identifier)
	d.SetId(orgGroup.Group.ID)
	dSet(d, "name", orgGroup.Group.Name)
	dSet(d, "provider_type", orgGroup.Group.ProviderType)
	dSet(d, "description", orgGroup.Group.Description)
	dSet(d, "role", orgGroup.Group.Role.Name)
	return nil
}
