package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRdeType() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeTypeRead,
		Schema: map[string]*schema.Schema{
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The vendor name for the Runtime Defined Entity type",
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique namespace associated with the Runtime Defined Entity type",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version of the Runtime Defined Entity type. The version string must follow semantic versioning rules",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Runtime Defined Entity type",
			},
			"interface_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of Defined Interface URNs that this Runtime Defined Entity type is referenced by",
			},
			"schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The JSON-Schema valid definition of the Runtime Defined Entity type",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Runtime Defined Entity type",
			},
			"external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An external entity's id that this definition may apply to",
			},
			"inherited_version": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "To be used when creating a new version of a Runtime Defined Entity type. Specifies the version of the type that will be the template for the authorization configuration of the new version." +
					"The Type ACLs and the access requirements of the Type Behaviors of the new version will be copied from those of the inherited version." +
					"If not set, then the new type version will not inherit another version and will have the default authorization settings, just like the first version of a new type",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Runtime Defined Entity type cannot be modified",
			},
		},
	}
}

func datasourceVcdRdeTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeTypeRead(ctx, d, meta, "datasource")
}
