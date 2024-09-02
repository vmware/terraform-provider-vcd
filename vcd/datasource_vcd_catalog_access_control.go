package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdCatalogAccessControl() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCatalogAccessControlRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of Catalog to read",
			},
			"shared_with_everyone": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the Catalog is shared with everyone",
			},
			"everyone_access_level": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Access level when the catalog is shared with everyone",
			},
			"read_only_shared_with_all_orgs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If true, the catalog is shared as read-only with all organizations",
			},
			"shared_with": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the Org to which we are sharing",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the user to which we are sharing",
						},
						"group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the group to which we are sharing",
						},
						"subject_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subject (org, group, or user) with which we are sharing",
						},
						"access_level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The access level for the org, user, or group to which we are sharing. One of [ReadOnly, Change, FullControl] for users and groups, but just ReadOnly for Organizations",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdCatalogAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogAccessControlRead(ctx, d, meta, "datasource")
}
