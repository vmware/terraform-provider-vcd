package vcd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var ipSpaceIpRangeRange = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Start address of the IP range",
			ValidateFunc: validation.IsIPAddress,
		},
		"end_address": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "End address of the IP range",
			ValidateFunc: validation.IsIPAddress,
		},
	},
}

var ipPrefixes = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"prefix": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "IP ranges (should match internal scope)",
			Elem:        ipSpacePrefix,
		},
		"default_quota": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Floating IP quota",
			ValidateFunc: IsIntAndAtLeast(-1), // -1 - unlimited, 0 - no quota
		},
	},
}

var ipSpacePrefix = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"first_ip": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "First IP in CIDR format",
			// ValidateFunc: validation.IsIPAddress,
		},
		"prefix_length": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "First IP in CIDR format",
			ValidateFunc: IsIntAndAtLeast(0),
		},
		"prefix_count": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Prefix count",
			ValidateFunc: IsIntAndAtLeast(1),
		},
	},
}

func resourceVcdIpSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIpSpaceCreate,
		ReadContext:   resourceVcdIpSpaceRead,
		UpdateContext: resourceVcdIpSpaceUpdate,
		DeleteContext: resourceVcdIpSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdIpSpaceImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For 'SHARED' (Org bound) IP spaces - Org ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of IP space",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of IP space",
			},

			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of IP space",
				// PUBLIC, SHARED_SERVICES, PRIVATE
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A set of up internal scope IPs in CIDR format",
				Elem: &schema.Schema{
					MinItems: 1,
					Type:     schema.TypeString,
				},
			},
			"ip_range_quota": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "IP ranges (should match internal scope)",
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"ip_range": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP ranges (should match internal scope)",
				Elem:        ipSpaceIpRangeRange,
			},
			"ip_prefix": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP prefixes (should match internal scope)",
				Elem:        ipPrefixes,
			},
			"external_scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "External scope in CIDR format",
			},
			"route_advertisement_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag whether route advertisement should be enabled",
			},
		},
	}
}

func resourceVcdIpSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space creation initiated")

	ipSpaceConfig, err := getIpSpaceType(d)
	if err != nil {
		return diag.Errorf("could not get IP Space type: %s", err)
	}

	createdIpSpace, err := vcdClient.CreateIpSpace(ipSpaceConfig)
	if err != nil {
		return diag.Errorf("error creating IP Space: %s", err)
	}

	d.SetId(createdIpSpace.IpSpace.ID)

	return resourceVcdIpSpaceRead(ctx, d, meta)
}

func resourceVcdIpSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space creation initiated")

	ipSpaceConfig, err := getIpSpaceType(d)
	if err != nil {
		return diag.Errorf("could not get IP Space type: %s", err)
	}

	ipSpace, err := vcdClient.GetIpSpaceById(d.Id())
	if err != nil {
		return diag.Errorf("error finding IP Space by ID '%s': %s", d.Id(), err)
	}

	ipSpaceConfig.ID = d.Id()

	_, err = ipSpace.Update(ipSpaceConfig)
	if err != nil {
		return diag.Errorf("error updating IP Space: %s", err)
	}

	return resourceVcdIpSpaceRead(ctx, d, meta)
}

func resourceVcdIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space read initiated")

	ipSpace, err := vcdClient.GetIpSpaceById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error finding IP Space by ID '%s': %s", d.Id(), err)
	}

	err = setIpSpaceData(d, ipSpace.IpSpace)
	if err != nil {
		return diag.Errorf("error storing IP Space state: %s", err)
	}

	return nil
}

func resourceVcdIpSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space deletion initiated")

	ipSpace, err := vcdClient.GetIpSpaceById(d.Id())
	if err != nil {
		return diag.Errorf("error finding IP Space by ID '%s': %s", d.Id(), err)
	}

	err = ipSpace.Delete()
	if err != nil {
		return diag.Errorf("error deleting IP space by ID '%s': %s", d.Id(), err)
	}

	return nil
}

func resourceVcdIpSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] IP Space import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 1 && len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as ip-space-name or org-name.ip-space-name")
	}
	vcdClient := meta.(*VCDClient)

	var ipSpace *govcd.IpSpace

	if len(resourceURI) == 2 {
		ipSpaceName := resourceURI[1]
		orgName := resourceURI[0]

		org, err := vcdClient.GetOrgByName(orgName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Org '%s': %s", orgName, err)
		}

		ipSpace, err = vcdClient.GetIpSpaceByNameAndOrgId(ipSpaceName, org.Org.ID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving IP Space '%s' in Org '%s': %s", ipSpaceName, orgName, err)
		}
		dSet(d, "org_id", org.Org.ID)

	} else {
		var err error
		ipSpaceName := resourceURI[0]
		ipSpace, err = vcdClient.GetIpSpaceByName(ipSpaceName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving IP Space '%s': %s", ipSpaceName, err)
		}
	}

	d.SetId(ipSpace.IpSpace.ID)

	return []*schema.ResourceData{d}, nil
}

func getIpSpaceType(d *schema.ResourceData) (*types.IpSpace, error) {

	ipSpace := &types.IpSpace{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),

		// Utilization: types.Utilization{
		// 	FloatingIPs: types.FloatingIPs{
		// 		TotalCount:          "",
		// 		AllocatedCount:      "",
		// 		UsedCount:           "",
		// 		UnusedCount:         "",
		// 		AllocatedPercentage: 0,
		// 		UsedPercentage:      0,
		// 	},
		// 	IPPrefixes: types.IPPrefixes{
		// 		TotalCount:               "",
		// 		AllocatedCount:           "",
		// 		UsedCount:                "",
		// 		UnusedCount:              "",
		// 		AllocatedPercentage:      0,
		// 		UsedPercentage:           0,
		// 		PrefixLengthUtilizations: []types.PrefixLengthUtilizations{},
		// 	},
		// },
		// IPSpaceRanges: types.IPSpaceRanges{
		// 	IPRanges: []types.IpSpaceRangeValues{
		// 		{
		// 			ID:             "",
		// 			StartIPAddress: "",
		// 			EndIPAddress:   "",
		// 			// TotalIPCount:          "",
		// 			// AllocatedIPCount:      "",
		// 			// AllocatedIPPercentage: 0,
		// 		},
		// 	},
		// 	DefaultFloatingIPQuota: 0,
		// },
		// IPSpacePrefixes: []types.IPSpacePrefixes{
		// 	{
		// 		IPPrefixSequence: []types.IPPrefixSequence{
		// 			{
		// 				// ID:                        "",
		// 				StartingPrefixIPAddress: "",
		// 				PrefixLength:            0,
		// 				TotalPrefixCount:        0,

		// 				// AllocatedPrefixCount:      0,
		// 				// AllocatedPrefixPercentage: 0,
		// 			},
		// 		},
		// 		DefaultQuotaForPrefixLength: 0,
		// 	},
		// },

		IPSpaceInternalScope:      convertSchemaSetToSliceOfStrings(d.Get("internal_scope").(*schema.Set)),
		IPSpaceExternalScope:      d.Get("external_scope").(string),
		RouteAdvertisementEnabled: d.Get("route_advertisement_enabled").(bool),
	}

	// IP Ranges
	ipRangeQuota := d.Get("ip_range_quota").(string)
	ipRangeQuotaInt, _ := strconv.Atoi(ipRangeQuota) // error is ignored because validation is enforced in schema

	ipSpace.IPSpaceRanges = types.IPSpaceRanges{
		DefaultFloatingIPQuota: ipRangeQuotaInt,
	}

	ipRanges := d.Get("ip_range").(*schema.Set)
	ipRangesSlice := ipRanges.List()
	if len(ipRangesSlice) > 0 {
		ipSpace.IPSpaceRanges.IPRanges = make([]types.IpSpaceRangeValues, len(ipRangesSlice))
		for ipRangeIndex := range ipRangesSlice {
			ipRangeStrings := convertToStringMap(ipRangesSlice[ipRangeIndex].(map[string]interface{}))

			ipSpace.IPSpaceRanges.IPRanges[ipRangeIndex].StartIPAddress = ipRangeStrings["start_address"]
			ipSpace.IPSpaceRanges.IPRanges[ipRangeIndex].EndIPAddress = ipRangeStrings["end_address"]
		}
	}

	// EOF IP Ranges

	// IP Prefixes
	ipPrefixes := d.Get("ip_prefix").(*schema.Set)
	ipPrefixesSlice := ipPrefixes.List()

	// Initialize structure
	if len(ipPrefixesSlice) > 0 {
		ipSpace.IPSpacePrefixes = []types.IPSpacePrefixes{}
	}

	for ipPrefixIndex := range ipPrefixesSlice {

		singleIpPrefix := ipPrefixesSlice[ipPrefixIndex]
		ipPrefixMap := singleIpPrefix.(map[string]interface{})
		ipPrefixQuota := ipPrefixMap["default_quota"].(string)
		ipPrefixQuotaInt, _ := strconv.Atoi(ipPrefixQuota) // ignoring error as validation is enforce in schema

		ipSpacePrefixType := types.IPSpacePrefixes{
			DefaultQuotaForPrefixLength: ipPrefixQuotaInt,
		}

		// Extract IP prefixess

		// 'prefix'
		ipPrefixPrefix := ipPrefixMap["prefix"].(*schema.Set)
		ipPrefixPrefixSlice := ipPrefixPrefix.List()
		if len(ipPrefixPrefixSlice) > 0 {
			ipSpacePrefixType.IPPrefixSequence = []types.IPPrefixSequence{}
		}

		for ipPrefixPrefixIndex := range ipPrefixPrefixSlice {

			ipPrefixMap := convertToStringMap(ipPrefixPrefixSlice[ipPrefixPrefixIndex].(map[string]interface{}))
			prefixLengthInt, _ := strconv.Atoi(ipPrefixMap["prefix_length"])
			prefixLengthCountInt, _ := strconv.Atoi(ipPrefixMap["prefix_count"])

			ipSpacePrefixType.IPPrefixSequence = append(ipSpacePrefixType.IPPrefixSequence, types.IPPrefixSequence{
				StartingPrefixIPAddress: ipPrefixMap["first_ip"],
				PrefixLength:            prefixLengthInt,
				TotalPrefixCount:        prefixLengthCountInt,
			})
		}

		// EOF // Extract IP prefixess

		// Add to the list
		ipSpace.IPSpacePrefixes = append(ipSpace.IPSpacePrefixes, ipSpacePrefixType)

	}

	// EOF IP Prefixes

	// only with
	orgId := d.Get("org_id").(string)
	if orgId != "" {
		ipSpace.OrgRef = &types.OpenApiReference{ID: orgId}
	}

	return ipSpace, nil
}

func setIpSpaceData(d *schema.ResourceData, ipSpace *types.IpSpace) error {

	dSet(d, "name", ipSpace.Name)
	dSet(d, "description", ipSpace.Description)
	dSet(d, "type", ipSpace.Type)
	dSet(d, "route_advertisement_enabled", ipSpace.RouteAdvertisementEnabled)
	dSet(d, "external_scope", ipSpace.IPSpaceExternalScope)

	if ipSpace.OrgRef != nil && ipSpace.OrgRef.ID != "" {
		dSet(d, "org_id", ipSpace.OrgRef.ID)
	}

	ipRangeQuotaStr := strconv.Itoa(ipSpace.IPSpaceRanges.DefaultFloatingIPQuota)
	dSet(d, "ip_range_quota", ipRangeQuotaStr)

	prefixesInterface := make([]interface{}, len(ipSpace.IPSpacePrefixes))
	for i, val := range ipSpace.IPSpacePrefixes {
		singlePrefix := make(map[string]interface{})

		strQuotaPrefixLength := strconv.Itoa(val.DefaultQuotaForPrefixLength)
		singlePrefix["default_quota"] = strQuotaPrefixLength

		prefSequence := make([]interface{}, len(val.IPPrefixSequence))
		for ii, seqVal := range val.IPPrefixSequence {

			singlePrefixSequence := make(map[string]interface{})

			prefixLengthStr := strconv.Itoa(seqVal.PrefixLength)
			prefixCountStr := strconv.Itoa(seqVal.TotalPrefixCount)

			singlePrefixSequence["first_ip"] = seqVal.StartingPrefixIPAddress
			singlePrefixSequence["prefix_length"] = prefixLengthStr
			singlePrefixSequence["prefix_count"] = prefixCountStr

			prefSequence[ii] = singlePrefixSequence
		}

		singlePrefix["prefix"] = prefSequence
		prefixesInterface[i] = singlePrefix
	}

	err := d.Set("ip_prefix", prefixesInterface)
	if err != nil {
		return fmt.Errorf("error storing 'ip_prefix': %s", err)
	}

	// IP ranges
	ipRangesInterface := make([]interface{}, len(ipSpace.IPSpaceRanges.IPRanges))
	for i, val := range ipSpace.IPSpaceRanges.IPRanges {
		singleRange := make(map[string]interface{})

		singleRange["start_address"] = val.StartIPAddress
		singleRange["end_address"] = val.EndIPAddress

		ipRangesInterface[i] = singleRange
	}
	err = d.Set("ip_range", ipRangesInterface)
	if err != nil {
		return fmt.Errorf("error storing 'ip_range': %s", err)
	}
	// IP ranges

	// Internal scope
	setOfStrs := convertStringsToTypeSet(ipSpace.IPSpaceInternalScope)
	err = d.Set("internal_scope", setOfStrs)
	if err != nil {
		return fmt.Errorf("error storing 'internal_scope': %s", err)
	}

	return nil
}
