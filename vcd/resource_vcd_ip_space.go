package vcd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

var ipSpaceIpRangeRange = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of IP Range",
		},
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
			Description: "One or more prefixes",
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
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of IP Prefix",
		},
		"first_ip": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "First IP",
		},
		"prefix_length": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Prefix length",
			ValidateFunc: IsIntAndAtLeast(0),
		},
		"prefix_count": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Number of prefixes to define",
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
				Description: "Org ID for 'SHARED' IP spaces",
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
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of IP space",
				ValidateFunc: validation.StringInSlice([]string{"PUBLIC", "SHARED_SERVICES", "PRIVATE"}, false)},
			"internal_scope": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A set of internal scope IPs in CIDR format",
				Elem: &schema.Schema{
					MinItems: 1,
					Type:     schema.TypeString,
				},
			},
			"ip_range_quota": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "IP ranges quota. '-1' - unlimited, '0' - no quota",
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"ip_range": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "One or more IP ranges for floating IP allocation",
				Elem:        ipSpaceIpRangeRange,
			},
			"ip_prefix": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "One or more IP prefixes within internal scope",
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
			"default_firewall_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag exposing whether default firewall rule creation should be enabled (VCD 10.5.0+)",
			},
			"default_no_snat_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag whether NO SNAT rule creation should be enabled (VCD 10.5.0+)",
			},
			"default_snat_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag whether SNAT rule creation should be enabled (VCD 10.5.0+)",
			},
		},
	}
}

func resourceVcdIpSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space creation initiated")

	ipSpaceConfig, err := getIpSpaceType(vcdClient, d, "create")
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
	log.Printf("[TRACE] IP Space update initiated")

	ipSpaceConfig, err := getIpSpaceType(vcdClient, d, "update")
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
		// Duplicate key case can happen when splitting a single IP range
		// start_address = "11.11.11.100"
		// end_address   = "11.11.11.111"
		// Into two ranges, where one matches `start_address` and another matches `end_address`
		// start_address = "11.11.11.100" <- matches the start_address of single range I had
		// end_address   = "11.11.11.108"
		// start_address = "11.11.11.110"
		// end_address   = "11.11.11.111" <- matches the end_address of single range I had
		//
		// It does not cause any issues on VCD or Terraform state except the split is not done even
		// after multiple apply attempts.
		// The point here is to hint user that what he is doing could be done in two steps
		if strings.Contains(err.Error(), "Duplicate key") {
			return diag.Errorf("error updating IP Space: splitting a single IP Range into two ranges "+
				"with matching 'start_address' and 'end_address' is not supported in a single update "+
				"operation. Please perform it in two steps: %s", err)
		}

		return diag.Errorf("error updating IP Space: %s", err)
	}

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcdClient.Client.WaitForRouteAdvertisementTasks()
	if err != nil {
		return diag.FromErr(err)
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

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcdClient.Client.WaitForRouteAdvertisementTasks()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceVcdIpSpaceImport has two cases:
// * Import global (provider level) IP Space - just a name is required
// * Import Private IP space for an Organization - org-name.ip-space-name is required
func resourceVcdIpSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] IP Space import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 1 && len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as ip-space-name or org-name.ip-space-name")
	}
	vcdClient := meta.(*VCDClient)

	var ipSpace *govcd.IpSpace

	switch {
	case len(resourceURI) == 2: // Resource path is supplied as `org-name.ip-space-name`
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
	case len(resourceURI) == 1: // Resource path is supplied as `ip-space-name`
		var err error
		ipSpaceName := resourceURI[0]
		ipSpace, err = vcdClient.GetIpSpaceByName(ipSpaceName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving IP Space '%s': %s", ipSpaceName, err)
		}
	default:
		return nil, fmt.Errorf("unrecognized path for IP Space import '%s'", d.Id())
	}

	d.SetId(ipSpace.IpSpace.ID)

	return []*schema.ResourceData{d}, nil
}

func getIpSpaceDefaultGatewayServiceConfig(vcdClient *VCDClient, d *schema.ResourceData) (*types.IpSpaceDefaultGatewayServiceConfig, error) {
	gwSvcConfig := &types.IpSpaceDefaultGatewayServiceConfig{
		EnableDefaultFirewallRuleCreation: d.Get("default_firewall_rule_creation_enabled").(bool),
		EnableDefaultNoSnatRuleCreation:   d.Get("default_no_snat_rule_creation_enabled").(bool),
		EnableDefaultSnatRuleCreation:     d.Get("default_snat_rule_creation_enabled").(bool),
	}

	// Firewall, SNAT and NO SNAT rule creation is only supported in VCD 10.5.0+
	// Performing runtime validation as supplying any of these values will cause an ugly error due to these fields
	// not being present on VCDs before 10.5.0
	if !vcdClient.Client.APIVCDMaxVersionIs(">= 38.0") {
		// If any of these values are `true`, but VCD does not support it - return error
		if gwSvcConfig.EnableDefaultFirewallRuleCreation || gwSvcConfig.EnableDefaultNoSnatRuleCreation || gwSvcConfig.EnableDefaultSnatRuleCreation {
			return nil, fmt.Errorf("default Firewall and NAT rule creation is only supported in VCD 10.5.0+")
		}
		// none of these values are `true` therefore simply returning `nil` so that no values are sent
		return nil, nil
	}

	return gwSvcConfig, nil
}

func getIpSpaceType(vcdClient *VCDClient, d *schema.ResourceData, operation string) (*types.IpSpace, error) {
	ipSpace := &types.IpSpace{
		Name:                      d.Get("name").(string),
		Description:               d.Get("description").(string),
		Type:                      d.Get("type").(string),
		IPSpaceInternalScope:      convertSchemaSetToSliceOfStrings(d.Get("internal_scope").(*schema.Set)),
		IPSpaceExternalScope:      d.Get("external_scope").(string),
		RouteAdvertisementEnabled: d.Get("route_advertisement_enabled").(bool),
	}

	// Get default gateway service configuration (firewall, snat, no snat rule creation)
	// It is available in VCD 10.5.0+, therefore it also does validation if any of the fields are
	// set to
	defaultGatewayServiceConfig, err := getIpSpaceDefaultGatewayServiceConfig(vcdClient, d)
	if err != nil {
		return nil, fmt.Errorf("error preparing IP Space configuration: %s", err)
	}
	ipSpace.DefaultGatewayServiceConfig = defaultGatewayServiceConfig

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

			// This is important for update - an ID of IP range must be supplied to prevent
			// recreating an IP Space
			if operation == "update" {
				foundIdInState := getIpRangeIdFromFromPreviousState(d, ipRangeStrings["start_address"], ipRangeStrings["end_address"])
				ipSpace.IPSpaceRanges.IPRanges[ipRangeIndex].ID = foundIdInState
			}

			ipSpace.IPSpaceRanges.IPRanges[ipRangeIndex].StartIPAddress = ipRangeStrings["start_address"]
			ipSpace.IPSpaceRanges.IPRanges[ipRangeIndex].EndIPAddress = ipRangeStrings["end_address"]
		}
	}
	// EOF IP Ranges

	ipPrefixes := d.Get("ip_prefix").(*schema.Set)
	ipPrefixesSlice := ipPrefixes.List()

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

		// Nested prefix definitions within IP Prefixes structure (`ip_prefix.X.prefix` blocks)
		nestedPrefixSet := ipPrefixMap["prefix"].(*schema.Set)
		nestedPrefixSlice := nestedPrefixSet.List()
		if len(nestedPrefixSlice) > 0 {
			ipSpacePrefixType.IPPrefixSequence = []types.IPPrefixSequence{}
		}

		for nestedPrefixSliceIndex := range nestedPrefixSlice {
			ipPrefixMap := convertToStringMap(nestedPrefixSlice[nestedPrefixSliceIndex].(map[string]interface{}))
			prefixLengthInt, _ := strconv.Atoi(ipPrefixMap["prefix_length"])
			prefixLengthCountInt, _ := strconv.Atoi(ipPrefixMap["prefix_count"])

			singlePrefix := types.IPPrefixSequence{
				StartingPrefixIPAddress: ipPrefixMap["first_ip"],
				PrefixLength:            prefixLengthInt,
				TotalPrefixCount:        prefixLengthCountInt,
			}

			// Update operation requires the ID of prefix, otherwise it recreates the IP sequence
			if operation == "update" {
				foundId := getIpPrefixSequenceIdFromFromPreviousState(d, ipPrefixMap["first_ip"], ipPrefixMap["prefix_length"], ipPrefixMap["prefix_count"])
				singlePrefix.ID = foundId
			}

			ipSpacePrefixType.IPPrefixSequence = append(ipSpacePrefixType.IPPrefixSequence, singlePrefix)
		}
		// EOF Nested prefix definitions within IP Prefixes structure (`ip_prefix.X.prefix` blocks)

		// Add each IP Prefix to the list
		ipSpace.IPSpacePrefixes = append(ipSpace.IPSpacePrefixes, ipSpacePrefixType)

	}
	// EOF IP Prefixes (`ip_prefix` blocks)

	// only when `org_id` is set (IP Space is Private)
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

	if ipSpace.DefaultGatewayServiceConfig != nil {
		dSet(d, "default_firewall_rule_creation_enabled", ipSpace.DefaultGatewayServiceConfig.EnableDefaultFirewallRuleCreation)
		dSet(d, "default_no_snat_rule_creation_enabled", ipSpace.DefaultGatewayServiceConfig.EnableDefaultNoSnatRuleCreation)
		dSet(d, "default_snat_rule_creation_enabled", ipSpace.DefaultGatewayServiceConfig.EnableDefaultSnatRuleCreation)
	} else { // default is always `false`
		dSet(d, "default_firewall_rule_creation_enabled", false)
		dSet(d, "default_no_snat_rule_creation_enabled", false)
		dSet(d, "default_snat_rule_creation_enabled", false)
	}

	ipRangeQuotaStr := strconv.Itoa(ipSpace.IPSpaceRanges.DefaultFloatingIPQuota)
	dSet(d, "ip_range_quota", ipRangeQuotaStr)

	// ip_prefix
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

			// Storing ID of this particular prefix which is required during update to prevent
			// recreating IP Prefix
			singlePrefixSequence["id"] = seqVal.ID

			prefSequence[ii] = singlePrefixSequence
		}

		singlePrefix["prefix"] = prefSequence
		prefixesInterface[i] = singlePrefix
	}

	err := d.Set("ip_prefix", prefixesInterface)
	if err != nil {
		return fmt.Errorf("error storing 'ip_prefix': %s", err)
	}
	// EOF ip_prefix

	// IP ranges
	ipRangesInterface := make([]interface{}, len(ipSpace.IPSpaceRanges.IPRanges))
	for i, val := range ipSpace.IPSpaceRanges.IPRanges {
		singleRange := make(map[string]interface{})

		singleRange["start_address"] = val.StartIPAddress
		singleRange["end_address"] = val.EndIPAddress
		// Storing ID of this particular range which is required during update to prevent recreating
		// IP range
		singleRange["id"] = val.ID

		ipRangesInterface[i] = singleRange
	}
	err = d.Set("ip_range", ipRangesInterface)
	if err != nil {
		return fmt.Errorf("error storing 'ip_range': %s", err)
	}
	// EOF IP ranges

	// Internal scope
	setOfIps := convertStringsToTypeSet(ipSpace.IPSpaceInternalScope)
	err = d.Set("internal_scope", setOfIps)
	if err != nil {
		return fmt.Errorf("error storing 'internal_scope': %s", err)
	}

	return nil
}

// getIpRangeIdFromFromPreviousState helps to find ip_range ID from previous state (because the current does not have it) and match it for current configuration
func getIpRangeIdFromFromPreviousState(d *schema.ResourceData, startAddress, endAddress string) string {
	ipRangesOld, _ := d.GetChange("ip_range")
	ipRangesOldSchema := ipRangesOld.(*schema.Set)
	ipRangesSlice := ipRangesOldSchema.List()

	util.Logger.Printf("[TRACE] Looking for ID of 'ip_range' with start_address '%s' and end_address '%s'\n", startAddress, endAddress)

	// Looping over ip_range definitions from state which contained all values and also stored ID. It looks for this ID with 2 priority levels:
	// 1. An exact match with the same start and end IP addresses is found - return it immediately
	// as this particular block has not changed at all for update operation
	// 2. A partial match - if at least the start_address or end_address remained the same - cache
	// this ID as it is the best bet to capture it if nothing was found
	//
	// The problem remains that one might be changing both - start and end IP addresses. In such
	// case this is a completely different configuration from Terraform TypeSet and there is no way
	// to match previous and current values

	var foundPartialId string
	for ipRangeIndex := range ipRangesSlice {
		ipRangeStrings := convertToStringMap(ipRangesSlice[ipRangeIndex].(map[string]interface{}))

		// If both - start and end IP addresses remained the same - we have found the ID and can
		// return it immediatelly
		if ipRangeStrings["start_address"] == startAddress && ipRangeStrings["end_address"] == endAddress {
			util.Logger.Printf("[TRACE] Found exact match for 'ip_range' with start_address '%s' and end_address '%s' - ID is '%s'\n",
				startAddress, endAddress, ipRangeStrings["id"])
			return ipRangeStrings["id"]
		}

		// Search for a partial match where either start_address or end_address matches
		if ipRangeStrings["start_address"] == startAddress || ipRangeStrings["end_address"] == endAddress {
			util.Logger.Printf("[TRACE] Found a partial match for 'ip_range' with start_address '%s' (looked) '%s' (found) and end_address '%s' (looked) '%s' (found) - ID is '%s'. "+
				"Storing until search finalizes.",
				startAddress, ipRangeStrings["start_address"], endAddress, ipRangeStrings["end_address"], ipRangeStrings["id"])
			foundPartialId = ipRangeStrings["id"]
		}
	}

	if foundPartialId != "" {
		util.Logger.Printf("[TRACE] Returning partial match for 'ip_range' with start_address '%s' and end_address '%s'  - ID is '%s'.",
			startAddress, endAddress, foundPartialId)
		return foundPartialId
	}

	util.Logger.Printf("[TRACE] No matches found for 'ip_range' with start_address '%s' and end_address '%s'  - ID is '%s'.",
		startAddress, endAddress, foundPartialId)

	return ""
}

// getIpPrefixSequenceIdFromFromPreviousState helps to find ip_prefix ID from previous state
// (because the current does not have it) and match it for current configuration
func getIpPrefixSequenceIdFromFromPreviousState(d *schema.ResourceData, firstIp, prefixLength, prefixCount string) string {
	ipPrefixOld, _ := d.GetChange("ip_prefix")
	ipPrefixOldSchema := ipPrefixOld.(*schema.Set)
	ipPrefixesSlice := ipPrefixOldSchema.List()

	util.Logger.Printf("[TRACE] Looking for ID of 'ip_prefix' with first_ip '%s', prefix_length '%s' and prefix_count '%s'\n", firstIp, prefixLength, prefixCount)
	var foundPartialId string
	for ipPrefixIndex := range ipPrefixesSlice {
		singleIpPrefix := ipPrefixesSlice[ipPrefixIndex]
		ipPrefixMap := singleIpPrefix.(map[string]interface{})

		// Nested prefix definitions within IP Prefixes structure (`ip_prefix.X.prefix` blocks)
		nestedPrefixSet := ipPrefixMap["prefix"].(*schema.Set)
		nestedPrefixSlice := nestedPrefixSet.List()

		for nestedPrefixSliceIndex := range nestedPrefixSlice {
			ipPrefixMap := convertToStringMap(nestedPrefixSlice[nestedPrefixSliceIndex].(map[string]interface{}))

			// Exact match
			if ipPrefixMap["first_ip"] == firstIp && ipPrefixMap["prefix_length"] == prefixLength && ipPrefixMap["prefix_count"] == prefixCount {
				util.Logger.Printf("[TRACE] Found exact match for ID '%s' of 'ip_prefix' with first_ip '%s', prefix_length '%s' and prefix_count '%s'\n", ipPrefixMap["id"], firstIp, prefixLength, prefixCount)
				return ipPrefixMap["id"]
			}

			if ipPrefixMap["first_ip"] == firstIp {
				util.Logger.Printf("[TRACE] Found partial match for ID '%s' of 'ip_prefix' with first_ip '%s'. 'prefix_length' and 'prefix_count' are ignored'\n", ipPrefixMap["id"], firstIp)
				foundPartialId = ipPrefixMap["id"]
			}
		}
	}

	if foundPartialId != "" {
		util.Logger.Printf("[TRACE] Returning partial match for ID '%s' of 'ip_prefix' with first_ip '%s'. 'prefix_length' and 'prefix_count' are ignored'\n", foundPartialId, firstIp)
		return foundPartialId
	}

	util.Logger.Printf("[TRACE] Not found 'ip_prefix' ID \n")
	// No ID was found at all
	return ""
}
