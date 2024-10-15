package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TODO: TM: validate compatibility with old data source

func datasourceVcdNsxtManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtManagerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Manager",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of NSX-T Manager",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NSX-T Manager",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username for authenticating to NSX-T Manager",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of NSX-T Manager",
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Provider Scope for NSX-T Manager",
			},
			"status": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of NSX-T Manager",
			},
		},
	}
}

func datasourceVcdNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelNsxtManager,
		getEntityFunc:  vcdClient.GetNsxtManagerOpenApiByName,
		stateStoreFunc: setTmNsxtManagerData,
	}
	return readDatasource(ctx, d, meta, c)
}
