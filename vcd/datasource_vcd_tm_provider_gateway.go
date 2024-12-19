package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmProviderGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmProviderGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmProviderGateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelTmRegion, labelTmProviderGateway),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelTmProviderGateway),
			},
			"nsxt_tier0_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelTmTier0Gateway, labelTmProviderGateway),
			},
			"ip_space_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelTmRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelTmProviderGateway),
			},
		},
	}
}

func datasourceVcdTmProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	getProviderGateway := func(name string) (*govcd.TmProviderGateway, error) {
		return vcdClient.GetTmProviderGatewayByNameAndRegionId(name, d.Get("region_id").(string))
	}
	c := dsReadConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelTmProviderGateway,
		getEntityFunc:  getProviderGateway,
		stateStoreFunc: setTmProviderGatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}
