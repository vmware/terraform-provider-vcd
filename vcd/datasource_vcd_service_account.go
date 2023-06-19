package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of service account",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"software_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of software, e.g: 12345678-1234-5678-90ab-1234567890ab",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role ID of service account",
			},
			"software_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version of software using the service account",
			},
			"uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI of the client using the service account",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of the service account.",
			},
		},
	}
}

func datasourceVcdServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdServiceAccountRead(ctx, d, meta, "datasource")
}
