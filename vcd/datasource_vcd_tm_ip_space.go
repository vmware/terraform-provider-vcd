package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

var dsTmIpSpaceInternalScopeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of internal scope within %s", labelTmIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("Name of internal scope within %s", labelTmIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP block within %s", labelTmIpSpace),
		},
	},
}

func datasourceVcdTmIpSpace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmIpSpaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmIpSpace),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelTmIpSpace),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelTmIpSpace),
			},
			"external_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External scope in CIDR format",
			},
			"default_quota_max_subnet_size": {
				Type:        schema.TypeString, // Values are 'ints', but
				Computed:    true,
				Description: fmt.Sprintf("Maximum subnet size represented as a prefix length (e.g. 24, 28) in %s", labelTmIpSpace),
			},
			"default_quota_max_cidr_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of subnets that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelTmIpSpace),
			},
			"default_quota_max_ip_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of single floating IP addresses that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelTmIpSpace),
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Internal scope of %s", labelTmIpSpace),
				Elem:        dsTmIpSpaceInternalScopeSchema,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelTmIpSpace),
			},
		},
	}
}

func datasourceVcdTmIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	getTmIpSpaceByName := func(name string) (*govcd.TmIpSpace, error) {
		return vcdClient.GetTmIpSpaceByNameAndRegionId(name, d.Get("region_id").(string))
	}

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelTmIpSpace,
		getEntityFunc:  getTmIpSpaceByName,
		stateStoreFunc: setTmIpSpaceData,
	}
	return readDatasource(ctx, d, meta, c)
}
