package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func slcChildComponent(title string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: fmt.Sprintf("ID of %s", title),
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("Name of %s", title),
				},
				"is_default": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: fmt.Sprintf("Boolean value that marks if this %s should be default", title),
				},
				"capabilities": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: fmt.Sprintf("Set of capabilities for %s", title),
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func resourceVcdSolutionLandingZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionLandingZoneCreate,
		ReadContext:   resourceVcdSolutionLandingZoneRead,
		UpdateContext: resourceVcdSolutionLandingZoneUpdate,
		DeleteContext: resourceVcdSolutionLandingZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionLandingZoneImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// Description: "The name of organization to use, optional if defined at provider " +
				// 	"level. Useful when connected as sysadmin working across different organizations",
			},

			"state": {
				Type:        schema.TypeString,
				Description: "State reports RDE state",
				Computed:    true,
			},
			"catalog": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "IP Address of pool member",
				// 	// Warning: This catalog stores all executable .ISO files for your solution add-ons.
				// 	//
				// 	// Selecting another catalog to use in the Solution Add-On Landing Zone does not affect the solution add-ons that you already installed, but prevents you from running day-2 operations on them. То ensure that you can run day-2 operations on the add-ons that are already installed, reupload their original add-on .ISO files.
				// 	// Capabilities???
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Shows is the member is enabled or not",
						},
						"capabilities": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"vdc": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Shows is the member is enabled or not",
						},
						"capabilities": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"org_vdc_network": slcChildComponent("Org VDC Network"),
						"storage_policy":  slcChildComponent("Storage Policy"),
						"compute_policy":  slcChildComponent("Compute Policy"),
					},
				},
			},
		},
	}
}

func resourceVcdSolutionLandingZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slzCfg, err := getSlzType(vcdClient, d)
	if err != nil {
		return diag.Errorf("error getting Solution Landing Zone configuration: %s", err)
	}

	slz, err := vcdClient.CreateSolutionLandingZone(slzCfg)
	if err != nil {
		return diag.Errorf("error creating Solution Landing Zone: %s", err)
	}

	// The real ID of Solution Landing Zone is RDE ID
	d.SetId(slz.DefinedEntity.DefinedEntity.ID)

	return resourceVcdSolutionLandingZoneRead(ctx, d, meta)
}

func resourceVcdSolutionLandingZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionLandingZoneById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving ID: %s", err)
	}

	slzCfg, err := getSlzType(vcdClient, d)
	if err != nil {
		return diag.Errorf("error getting Solution Landing Zone configuration: %s", err)
	}

	_, err = slz.Update(slzCfg)
	if err != nil {
		return diag.Errorf("error updating Solution Landing Zone: %s", err)
	}

	return resourceVcdSolutionLandingZoneRead(ctx, d, meta)
}

func resourceVcdSolutionLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionLandingZoneById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Landing Zone: %s", err)
	}

	err = setSlzData(d, slz)
	if err != nil {
		return diag.Errorf("error storing data to schema: %s", err)
	}

	return nil
}

func resourceVcdSolutionLandingZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionLandingZoneById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving ID: %s", err)
	}

	err = slz.Delete()
	if err != nil {
		return diag.Errorf("error deleting Solution Landing Zone RDE: %s", err)
	}

	return nil
}

func resourceVcdSolutionLandingZoneImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// TODO TODO
	// Can there be more than 1 SLZ?
	return []*schema.ResourceData{d}, nil
}

func getSlzType(vcdClient *VCDClient, d *schema.ResourceData) (*types.SolutionLandingZoneType, error) {
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org: %s", err)
	}
	slzCfg := &types.SolutionLandingZoneType{
		Name: org.Org.Name,
		ID:   org.Org.ID,
		Vdcs: make([]types.SolutionLandingZoneVdc, 1),
	}

	vdcs := d.Get("vdc").(*schema.Set)
	vdcsList := vdcs.List()

	// Construct all VDCs
	for vdcIndex, vdc := range vdcsList {
		vdcMap := vdc.(map[string]interface{})

		vdcId := vdcMap["id"].(string)
		vdc, err := org.GetVDCById(vdcId, false)
		if err != nil {
			return nil, fmt.Errorf("error retrieving VDC by ID: %s", err)
		}

		slzCfg.Vdcs[vdcIndex] = types.SolutionLandingZoneVdc{
			ID:           vdcId,
			Name:         vdc.Vdc.Name,
			Capabilities: convertSchemaSetToSliceOfStrings(vdcMap["capabilities"].(*schema.Set)),
			IsDefault:    vdcMap["is_default"].(bool),
		}

		// Org VDC Networks
		orgNetworkNameRetriever := func(id string) (string, error) {
			orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkById(id)
			if err != nil {
				return "", fmt.Errorf("error retrieving Org VDC Network by name: %s", err)
			}
			return orgNetwork.OpenApiOrgVdcNetwork.Name, nil
		}
		networks, err := getSlzChildType(vdcMap["org_vdc_network"].(*schema.Set), orgNetworkNameRetriever)
		if err != nil {
			return nil, fmt.Errorf("error getting child entity type for Org VDC Networks: %s", err)
		}
		slzCfg.Vdcs[vdcIndex].Networks = networks

		// Storage Policies
		storageProfileNameRetriever := func(id string) (string, error) {
			storageProfile, err := vcdClient.GetStorageProfileById(id)
			if err != nil {
				return "", fmt.Errorf("error retrieving storage profile by ID: %s", err)
			}
			return storageProfile.Name, nil
		}
		storagePolicies, err := getSlzChildType(vdcMap["storage_policy"].(*schema.Set), storageProfileNameRetriever)
		if err != nil {
			return nil, fmt.Errorf("error getting child entity type for Storage Policies: %s", err)
		}
		slzCfg.Vdcs[vdcIndex].StoragePolicies = storagePolicies

		// Compute Policies
		computePolicyNameRetriever := func(id string) (string, error) {
			computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(id)
			if err != nil {
				return "", fmt.Errorf("error retrieving compute policy by ID: %S", err)
			}
			return computePolicy.VdcComputePolicyV2.Name, nil
		}
		computePolicies, err := getSlzChildType(vdcMap["compute_policy"].(*schema.Set), computePolicyNameRetriever)
		if err != nil {
			return nil, fmt.Errorf("error getting child entity type for Compute Policies: %s", err)
		}
		slzCfg.Vdcs[vdcIndex].ComputePolicies = computePolicies
	}

	// Construct Catalog list
	catalogSet := d.Get("catalog").(*schema.Set)
	catalogList := catalogSet.List()

	slzCfg.Catalogs = make([]types.SolutionLandingZoneCatalog, len(catalogList))
	for catalogIndex, catalog := range catalogList {
		catalogMap := catalog.(map[string]interface{})

		catalogId := catalogMap["id"].(string)
		catalog, err := org.GetCatalogById(catalogId, false)
		if err != nil {
			return nil, fmt.Errorf("error retrieving catalog by ID: %s", err)
		}

		slzCfg.Catalogs[catalogIndex] = types.SolutionLandingZoneCatalog{
			ID:           catalogId,
			Name:         catalog.Catalog.Name,
			Capabilities: convertSchemaSetToSliceOfStrings(catalogMap["capabilities"].(*schema.Set)),
		}

	}

	return slzCfg, nil
}

func getSlzChildType(entrySet *schema.Set, entryNameLookupFunc func(string) (string, error)) ([]types.SolutionLandingZoneVdcChild, error) {
	entityList := entrySet.List()

	results := make([]types.SolutionLandingZoneVdcChild, len(entityList))
	for entityIndex, entity := range entityList {
		entityMap := entity.(map[string]interface{})

		childEntityId := entityMap["id"].(string)
		childEntityName, err := entryNameLookupFunc(childEntityId) // API requires name to be present
		if err != nil {
			return nil, fmt.Errorf("error retrieving child entity '%s' name: %s", childEntityId, err)
		}

		results[entityIndex] = types.SolutionLandingZoneVdcChild{
			ID:           childEntityId,
			Name:         childEntityName,
			Capabilities: convertSchemaSetToSliceOfStrings(entityMap["capabilities"].(*schema.Set)),
			IsDefault:    entityMap["is_default"].(bool),
		}
	}
	return results, nil
}

func setSlzData(d *schema.ResourceData, slz *govcd.SolutionLandingZone) error {
	dSet(d, "state", slz.DefinedEntity.DefinedEntity.State)

	catalogSchema := make([]interface{}, len(slz.SolutionLandingZoneType.Catalogs))
	for catalogIndex, singleCatalog := range slz.SolutionLandingZoneType.Catalogs {
		catalogEntry := make(map[string]interface{})

		catalogEntry["id"] = singleCatalog.ID
		catalogEntry["capabilities"] = convertStringsToTypeSet(singleCatalog.Capabilities)

		catalogSchema[catalogIndex] = catalogEntry
	}

	err := d.Set("catalog", catalogSchema)
	if err != nil {
		return fmt.Errorf("error storing 'catalog' to schema: %s", err)
	}

	vdcSchema := make([]interface{}, len(slz.SolutionLandingZoneType.Vdcs))
	for vdcIndex, singleVdc := range slz.SolutionLandingZoneType.Vdcs {
		vdcEntry := make(map[string]interface{})

		vdcEntry["id"] = singleVdc.ID
		vdcEntry["is_default"] = singleVdc.IsDefault
		vdcEntry["capabilities"] = convertStringsToTypeSet(singleVdc.Capabilities)

		vdcEntry["org_vdc_network"] = setChildData(slz.SolutionLandingZoneType.Vdcs[vdcIndex].Networks)
		vdcEntry["storage_policy"] = setChildData(slz.SolutionLandingZoneType.Vdcs[vdcIndex].StoragePolicies)
		vdcEntry["compute_policy"] = setChildData(slz.SolutionLandingZoneType.Vdcs[vdcIndex].ComputePolicies)

		vdcSchema[vdcIndex] = vdcEntry
	}

	err = d.Set("vdc", vdcSchema)
	if err != nil {
		return fmt.Errorf("error storing 'vdc' to schema: %s", err)
	}

	return nil
}

func setChildData(children []types.SolutionLandingZoneVdcChild) []interface{} {
	allEntries := make([]interface{}, len(children))
	for entityIndex, singleEntity := range children {
		singleEntry := make(map[string]interface{})

		singleEntry["id"] = singleEntity.ID
		singleEntry["name"] = singleEntity.Name
		singleEntry["is_default"] = singleEntity.IsDefault
		singleEntry["capabilities"] = convertStringsToTypeSet(singleEntity.Capabilities)

		allEntries[entityIndex] = singleEntry
	}

	return allEntries
}
