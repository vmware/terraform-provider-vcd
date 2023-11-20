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
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization that owns this Runtime Defined Entity, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity",
			},
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Runtime Defined Entity Type ID",
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "An external entity's ID that this Runtime Defined Entity may have a relation to",
				Computed:    true,
			},
			"entity": {
				Type:        schema.TypeString,
				Description: "A JSON representation of the Runtime Defined Entity",
				Computed:    true,
			},
			"owner_user_id": {
				Type:        schema.TypeString,
				Description: "The ID of the user that owns the Runtime Defined Entity",
				Computed:    true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Description: "The organization of the Runtime Defined Entity",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Specifies whether the entity is correctly resolved or not. One of PRE_CREATED, RESOLVED or RESOLUTION_ERROR",
				Computed:    true,
			},
			"metadata_entry": openApiMetadataEntryDatasourceSchema("Runtime Defined Entity"),
		},
	}
}

func datasourceVcdRdeRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient, "datasource")
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", rde.DefinedEntity.Name)
	dSet(d, "external_id", rde.DefinedEntity.ExternalId)
	dSet(d, "state", rde.DefinedEntity.State)

	jsonEntity, err := jsonToCompactString(rde.DefinedEntity.Entity)
	if err != nil {
		return diag.Errorf("could not save the Runtime Defined Entity JSON into state: %s", err)
	}
	err = d.Set("entity", jsonEntity)
	if err != nil {
		return diag.FromErr(err)
	}

	if rde.DefinedEntity.Org != nil {
		dSet(d, "org_id", rde.DefinedEntity.Org.ID)
	}
	if rde.DefinedEntity.Owner != nil {
		dSet(d, "owner_user_id", rde.DefinedEntity.Owner.ID)
	}

	d.SetId(rde.DefinedEntity.ID)

	return nil
}
