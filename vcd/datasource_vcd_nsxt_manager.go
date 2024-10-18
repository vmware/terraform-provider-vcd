package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TODO: TM: validate compatibility with old data source

func datasourceVcdNsxtManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtManagerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelNsxtManager),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelNsxtManager),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username for authenticating to %s", labelNsxtManager),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL of %s", labelNsxtManager),
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Network Provider Scope for %s", labelNsxtManager),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelNsxtManager),
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("HREF of %s", labelNsxtManager),
			},
		},
	}
}

func datasourceVcdNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelNsxtManager,
		getEntityFunc:  vcdClient.GetNsxtManagerOpenApiByName,
		stateStoreFunc: setNsxtManagerData,
	}
	return readDatasource(ctx, d, meta, c)
}
