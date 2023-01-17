package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRde() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Runtime Defined Entity",
				Required:    true,
			},
			"rde_type_id": {
				Type:        schema.TypeString,
				Description: "The type ID of the Runtime Defined Entity",
				Required:    true,
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "An external entity's ID that this Runtime Defined Entity may have a relation to",
				Computed:    true,
			},
			"entity": {
				Type:        schema.TypeString,
				Description: "A JSON representation of the Runtime Defined Entity. The JSON will be validated against the schema of the RDE type that the entity is an instance of",
				Computed:    true,
			},
			"owner_id": {
				Type:        schema.TypeString,
				Description: "The owner of the Runtime Defined Entity",
				Computed:    true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Description: "The organization of the Runtime Defined Entity",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Every Runtime Defined Entity is created in the \"PRE_CREATED\" state. Once an entity is ready to be validated against its schema, it will transition in another state - RESOLVED, if the entity is valid according to the schema, or RESOLUTION_ERROR otherwise. If an entity in an \"RESOLUTION_ERROR\" state is updated, it will transition to the inital \"PRE_CREATED\" state without performing any validation. If its in the \"RESOLVED\" state, then it will be validated against the entity type schema and throw an exception if its invalid",
				Computed:    true,
			},
			"metadata_entry": getOpenApiMetadataEntrySchema("Runtime Defined Entity", true),
		},
	}
}

func datasourceVcdRdeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeRead(ctx, d, meta, "datasource")
}
