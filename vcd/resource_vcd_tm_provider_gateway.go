package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmProviderGateway = "TM Provider Gateway"

func resourceVcdTmProviderGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmProviderGatewayCreate,
		ReadContext:   resourceVcdTmProviderGatewayRead,
		UpdateContext: resourceVcdTmProviderGatewayUpdate,
		DeleteContext: resourceVcdTmProviderGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmProviderGatewayImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmProviderGateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelTmRegion, labelTmProviderGateway),
			},
			"nsxt_tier0_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelTmTier0Gateway, labelTmProviderGateway),
			},
			"ip_space_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelTmRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelTmProviderGateway),
			},
		},
	}
}

func resourceVcdTmProviderGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:      labelTmProviderGateway,
		getTypeFunc:      getTmProviderGatewayType,
		stateStoreFunc:   setTmProviderGatewayData,
		createFunc:       vcdClient.CreateTmProviderGateway,
		resourceReadFunc: resourceVcdTmProviderGatewayRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmProviderGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:      labelTmProviderGateway,
		getTypeFunc:      getTmProviderGatewayType,
		getEntityFunc:    vcdClient.GetTmProviderGatewayById,
		resourceReadFunc: resourceVcdTmProviderGatewayRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelTmProviderGateway,
		getEntityFunc:  vcdClient.GetTmProviderGatewayById,
		stateStoreFunc: setTmProviderGatewayData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmProviderGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:   labelTmProviderGateway,
		getEntityFunc: vcdClient.GetTmProviderGatewayById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmProviderGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	d.SetId(vcdClient.Org)
	return []*schema.ResourceData{d}, nil
}

func getTmProviderGatewayType(vcdClient *VCDClient, d *schema.ResourceData) (*types.TmProviderGateway, error) {
	t := &types.TmProviderGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		RegionRef:   types.OpenApiReference{ID: d.Get("region_id").(string)},
		BackingRef:  types.OpenApiReference{ID: d.Get("nsxt_tier0_gateway_id").(string)},
	}

	ipSpaceIds := convertSchemaSetToSliceOfStrings(d.Get("ip_space_ids").(*schema.Set))
	t.IPSpaceRefs = convertSliceOfStringsToOpenApiReferenceIds(ipSpaceIds)

	return t, nil
}

func setTmProviderGatewayData(d *schema.ResourceData, p *govcd.TmProviderGateway) error {
	if p == nil || p.TmProviderGateway == nil {
		return fmt.Errorf("nil entity received")
	}

	d.SetId(p.TmProviderGateway.ID)
	dSet(d, "name", p.TmProviderGateway.Name)
	dSet(d, "description", p.TmProviderGateway.Description)
	dSet(d, "region_id", p.TmProviderGateway.RegionRef.ID)
	dSet(d, "nsxt_tier0_gateway_id", p.TmProviderGateway.BackingRef.ID)

	err := d.Set("ip_space_ids", extractIdsFromOpenApiReferences(p.TmProviderGateway.IPSpaceRefs))
	if err != nil {
		return fmt.Errorf("error storing 'ip_space_ids': %s", err)
	}

	return nil
}
