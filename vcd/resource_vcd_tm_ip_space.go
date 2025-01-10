package vcd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelTmIpSpace = "TM IP Space"

var tmIpSpaceInternalScopeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of internal scope within %s", labelTmIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("Name of internal scope within %s", labelTmIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP block within %s", labelTmIpSpace),
		},
	},
}

func resourceVcdTmIpSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmIpSpaceCreate,
		ReadContext:   resourceVcdTmIpSpaceRead,
		UpdateContext: resourceVcdTmIpSpaceUpdate,
		DeleteContext: resourceVcdTmIpSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmIpSpaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmIpSpace),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelTmIpSpace),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelTmIpSpace),
			},
			"external_scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "External scope in CIDR format",
			},
			"default_quota_max_subnet_size": {
				Type:         schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:     true,
				Description:  fmt.Sprintf("Maximum subnet size represented as a prefix length (e.g. 24, 28) in %s", labelTmIpSpace),
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"default_quota_max_cidr_count": {
				Type:         schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:     true,
				Description:  fmt.Sprintf("Maximum number of subnets that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelTmIpSpace),
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"default_quota_max_ip_count": {
				Type:         schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:     true,
				Description:  fmt.Sprintf("Maximum number of single floating IP addresses that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelTmIpSpace),
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("Internal scope of %s", labelTmIpSpace),
				Elem:        tmIpSpaceInternalScopeSchema,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelTmIpSpace),
			},
		},
	}
}

func resourceVcdTmIpSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelTmIpSpace,
		getTypeFunc:      getTmIpSpaceType,
		stateStoreFunc:   setTmIpSpaceData,
		createFunc:       vcdClient.CreateTmIpSpace,
		resourceReadFunc: resourceVcdTmIpSpaceRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelTmIpSpace,
		getTypeFunc:      getTmIpSpaceType,
		getEntityFunc:    vcdClient.GetTmIpSpaceById,
		resourceReadFunc: resourceVcdTmIpSpaceRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelTmIpSpace,
		getEntityFunc:  vcdClient.GetTmIpSpaceById,
		stateStoreFunc: setTmIpSpaceData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:   labelTmIpSpace,
		getEntityFunc: vcdClient.GetTmIpSpaceById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.ip-space-name")
	}
	regionName, ipSpaceName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	region, err := vcdClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelTmRegion, regionName, err)
	}

	ipSpace, err := vcdClient.GetTmIpSpaceByNameAndRegionId(ipSpaceName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by given name '%s': %s", labelTmIpSpace, d.Id(), err)
	}

	dSet(d, "region_id", region.Region.ID)
	d.SetId(ipSpace.TmIpSpace.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmIpSpaceType(vcdClient *VCDClient, d *schema.ResourceData) (*types.TmIpSpace, error) {
	t := &types.TmIpSpace{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		RegionRef:         types.OpenApiReference{ID: d.Get("region_id").(string)},
		ExternalScopeCidr: d.Get("external_scope").(string),
	}

	// error is ignored because validation is enforced in schema fields
	maxCidrCountInt, _ := strconv.Atoi(d.Get("default_quota_max_cidr_count").(string))
	maxIPCountInt, _ := strconv.Atoi(d.Get("default_quota_max_ip_count").(string))
	maxSubnetSizeInt, _ := strconv.Atoi(d.Get("default_quota_max_subnet_size").(string))
	t.DefaultQuota = types.TmIpSpaceDefaultQuota{
		MaxCidrCount:  maxCidrCountInt,
		MaxIPCount:    maxIPCountInt,
		MaxSubnetSize: maxSubnetSizeInt,
	}

	// internal_scope
	internalScope := d.Get("internal_scope").(*schema.Set)
	internalScopeSlice := internalScope.List()
	if len(internalScopeSlice) > 0 {
		isSlice := make([]types.TmIpSpaceInternalScopeCidrBlocks, len(internalScopeSlice))
		for internalScopeIndex := range internalScopeSlice {
			internalScopeBlockStrings := convertToStringMap(internalScopeSlice[internalScopeIndex].(map[string]interface{}))

			isSlice[internalScopeIndex].Name = internalScopeBlockStrings["name"]
			isSlice[internalScopeIndex].Cidr = internalScopeBlockStrings["cidr"]

			// ID of internal_scope is important for updates
			// Terraform TypeSet cannot natively identify the ID between previous and new states
			// To work around this, an attempt to retrieve ID from state and correlate it with new payload is done
			// An important fact is that `cidr` field is not updatable, therefore one can be sure
			// that ID from state can be looked up based on CIDR.
			// If there was no such cidr in previous state - it means that this is a new 'internal_scope' block
			// and it doesn't need an ID
			isSlice[internalScopeIndex].ID = getInternalScopeIdFromFromPreviousState(d, internalScopeBlockStrings["name"], internalScopeBlockStrings["cidr"])

		}
		t.InternalScopeCidrBlocks = isSlice
	}

	return t, nil
}

func setTmIpSpaceData(_ *VCDClient, d *schema.ResourceData, i *govcd.TmIpSpace) error {
	if i == nil || i.TmIpSpace == nil {
		return fmt.Errorf("nil %s received", labelTmIpSpace)
	}

	d.SetId(i.TmIpSpace.ID)
	dSet(d, "name", i.TmIpSpace.Name)
	dSet(d, "description", i.TmIpSpace.Description)
	dSet(d, "region_id", i.TmIpSpace.RegionRef.ID)
	dSet(d, "external_scope", i.TmIpSpace.ExternalScopeCidr)
	dSet(d, "status", i.TmIpSpace.Status)

	// maxSubnetSizeInt, _ := strconv.Atoi(i.TmIpSpace.DefaultQuota.MaxSubnetSize) // error is ignored because validation is enforced in schema

	dSet(d, "default_quota_max_subnet_size", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxSubnetSize))
	dSet(d, "default_quota_max_cidr_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxCidrCount))
	dSet(d, "default_quota_max_ip_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxIPCount))

	// internal_scope
	internalScopeInterface := make([]interface{}, len(i.TmIpSpace.InternalScopeCidrBlocks))
	for i, val := range i.TmIpSpace.InternalScopeCidrBlocks {
		singleScope := make(map[string]interface{})

		singleScope["id"] = val.ID
		singleScope["name"] = val.Name
		singleScope["cidr"] = val.Cidr

		internalScopeInterface[i] = singleScope
	}
	err := d.Set("internal_scope", internalScopeInterface)
	if err != nil {
		return fmt.Errorf("error storing 'internal_scope': %s", err)
	}

	return nil
}

func getInternalScopeIdFromFromPreviousState(d *schema.ResourceData, desiredName, desiredCidr string) string {
	internalScopeOld, _ := d.GetChange("internal_scope")
	internalScopeOldSchema := internalScopeOld.(*schema.Set)
	internalScopeOldSlice := internalScopeOldSchema.List()

	util.Logger.Printf("[TRACE] Looking for ID of 'internal_scope' with name '%s', cidr '%s'\n", desiredName, desiredCidr)
	var foundPartialId string
	for internalScopeIndex := range internalScopeOldSlice {
		singleScopeOld := internalScopeOldSlice[internalScopeIndex]
		singleScopeOldMap := convertToStringMap(singleScopeOld.(map[string]interface{}))

		// exact match
		if singleScopeOldMap["cidr"] == desiredCidr && singleScopeOldMap["name"] == desiredName {
			util.Logger.Printf("[TRACE] Found exact match for ID '%s' of 'internal_scope' with name '%s', cidr '%s' \n", singleScopeOldMap["id"], desiredName, desiredCidr)
			return singleScopeOldMap["id"]
		}

		// partial match based on cidr
		if singleScopeOldMap["cidr"] == desiredCidr {
			util.Logger.Printf("[TRACE] Found partial match for ID '%s' of 'internal_scope' with cidr '%s'. 'name' is ignored'\n", singleScopeOldMap["id"], desiredCidr)
			foundPartialId = singleScopeOldMap["id"]
		}
	}

	if foundPartialId != "" {
		util.Logger.Printf("[TRACE] Returning partial match for ID '%s' of 'internal_scope' with cidr '%s'. 'name' are ignored'\n", desiredCidr, desiredName)
		return foundPartialId
	}

	util.Logger.Printf("[TRACE] Not found 'internal_scope' ID with name '%s', cidr '%s'\n", desiredName, desiredCidr)
	// No ID was found at all
	return ""
}
