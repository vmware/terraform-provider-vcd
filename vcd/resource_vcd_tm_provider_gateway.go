package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmProviderGateway = "TM Provider Gateway"
const labelTmProviderGatewayIpSpaceAssociations = "TM IP Space Associations"

// TM Provider Gateway has an asymmetric API - it requires are least one IP Space reference when
// creating a Provider Gateway, but it will not return Associated IP Spaces afterwards. Instead,
// to update associated IP Spaces one must use separate API endpoint and structure
// (`TmIpSpaceAssociation`) to manage associations after initial Provider Gateway creation

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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelTmProviderGateway),
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelTmProviderGateway),
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

	// Update IP Space associations using separate endpoint (more details at the top of file)
	if d.HasChange("ip_space_ids") {
		previous, new := d.GetChange("ip_space_ids")
		previousSet := previous.(*schema.Set)
		newSet := new.(*schema.Set)

		toRemoveSet := previousSet.Difference(newSet)
		toAddSet := newSet.Difference(previousSet)

		// Adding new ones first, because it can happen that all previous IP Spaces are removed and
		// new ones added, however API prohibits removal of all IP Space associations for Provider
		// Gateway (at least one IP Space must always be associated)
		err := addIpSpaceAssociations(vcdClient, d.Id(), convertSchemaSetToSliceOfStrings(toAddSet))
		if err != nil {
			return diag.FromErr(err)
		}

		// Remove associations that are no more in configuration
		err = removeIpSpaceAssociations(vcdClient, d.Id(), convertSchemaSetToSliceOfStrings(toRemoveSet))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// This is the default entity update path - other fields can be updated, by updating IP Space itself
	if d.HasChangeExcept("ip_space_ids") {
		c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
			entityLabel:      labelTmProviderGateway,
			getTypeFunc:      getTmProviderGatewayType,
			getEntityFunc:    vcdClient.GetTmProviderGatewayById,
			resourceReadFunc: resourceVcdTmProviderGatewayRead,
		}

		return updateResource(ctx, d, meta, c)
	}

	return nil
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
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.provider-gateway-name")
	}
	regionName, providerGatewayName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	region, err := vcdClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelTmRegion, regionName, err)
	}

	providerGateway, err := vcdClient.GetTmProviderGatewayByNameAndRegionId(providerGatewayName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Provider Gateway: %s", err)
	}

	d.SetId(providerGateway.TmProviderGateway.ID)
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

	// Update operation fails if the ID is not set for update
	if d.Id() != "" {
		t.ID = d.Id()
	}

	// IP Spaces associations are populated on create only. Updates are done using separate endpoint
	// (more details at the top of file)
	if d.Id() != "" {
		t.IPSpaceRefs = []types.OpenApiReference{}
	}

	return t, nil
}

func setTmProviderGatewayData(vcdClient *VCDClient, d *schema.ResourceData, p *govcd.TmProviderGateway) error {
	if p == nil || p.TmProviderGateway == nil {
		return fmt.Errorf("nil entity received")
	}

	d.SetId(p.TmProviderGateway.ID)
	dSet(d, "name", p.TmProviderGateway.Name)
	dSet(d, "description", p.TmProviderGateway.Description)
	dSet(d, "region_id", p.TmProviderGateway.RegionRef.ID)
	dSet(d, "nsxt_tier0_gateway_id", p.TmProviderGateway.BackingRef.ID)
	dSet(d, "status", p.TmProviderGateway.Status)

	// IP Space Associations have to be read separatelly after creation (more details at the top of file)
	associations, err := vcdClient.GetAllTmIpSpaceAssociationsByProviderGatewayId(p.TmProviderGateway.ID)
	if err != nil {
		return fmt.Errorf("error retrieving %s for %s", labelTmProviderGatewayIpSpaceAssociations, labelTmProviderGateway)
	}
	associationIds := make([]string, len(associations))
	for index, singleAssociation := range associations {
		associationIds[index] = singleAssociation.TmIpSpaceAssociation.IPSpaceRef.ID
	}

	err = d.Set("ip_space_ids", associationIds)
	if err != nil {
		return fmt.Errorf("error storing 'ip_space_ids': %s", err)
	}

	return nil
}

func addIpSpaceAssociations(vcdClient *VCDClient, providerGatewayId string, addIpSpaceIds []string) error {
	for _, addIpSpaceId := range addIpSpaceIds {
		at := &types.TmIpSpaceAssociation{
			IPSpaceRef:         &types.OpenApiReference{ID: addIpSpaceId},
			ProviderGatewayRef: &types.OpenApiReference{ID: providerGatewayId},
		}
		_, err := vcdClient.CreateTmIpSpaceAssociation(at)
		if err != nil {
			return fmt.Errorf("error adding new %s for %s with ID '%s': %s",
				labelTmProviderGatewayIpSpaceAssociations, labelTmIpSpace, addIpSpaceId, err)
		}
	}

	return nil
}

func removeIpSpaceAssociations(vcdClient *VCDClient, providerGatewayId string, removeIpSpaceIds []string) error {
	existingIpSpaceAssociations, err := vcdClient.GetAllTmIpSpaceAssociationsByProviderGatewayId(providerGatewayId)
	if err != nil {
		return fmt.Errorf("error reading %s for update: %s", labelTmProviderGatewayIpSpaceAssociations, err)
	}

	for _, singleIpSpaceId := range removeIpSpaceIds {
		for _, singleAssociation := range existingIpSpaceAssociations {
			if singleAssociation.TmIpSpaceAssociation.IPSpaceRef.ID == singleIpSpaceId {
				err = singleAssociation.Delete()
				if err != nil {
					return fmt.Errorf("error removing %s '%s' for %s '%s': %s",
						labelTmProviderGatewayIpSpaceAssociations, singleAssociation.TmIpSpaceAssociation.ID, labelTmIpSpace, singleIpSpaceId, err)
				}
			}
		}
	}

	return nil
}
