package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdDsePublish() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdDsePublishRead,

		Schema: map[string]*schema.Schema{
			"data_solution_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of Data Solution",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A tenant ID that will get the Data Solution published",
			},
			"confluent_license_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Only available 'Confluent Platform'. One of '%s' or '%s'", confluentLicenseTypeNoLicense, confluentLicenseTypeWithLicense),
			},
			"dso_acl_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ACL ID for Data Solutions Operator",
			},
			"template_acl_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Solution Template ACL IDs provisioned to the tenant",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ds_org_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data Solution Org Configuration ID",
			},
		},
	}
}

func datasourceVcdDsePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdDsePublishRead(ctx, d, meta, "datasource")
}
