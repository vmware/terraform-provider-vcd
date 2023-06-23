package vcd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// resourceVcdIpSpaceCustomQuota has a non standard behavior due to how VCD API works. UI works this
// way as well.
//
// IP Space Custom Quota (IP Space Org Assignment name in API) is implicitly created when and Edge
// Gateway (vcd_nsxt_edgegateway) backed by Provider Gateway (vcd_nsxt_external_network_v2). To set
// custom quota, one does not need to create new Org Assignment entity, but rather find an existing
// one based on IP Space ID and Org ID and update `CustomQuotas` parameter.
// Due to this reason, delete operation also does not destroy the Assignment itself, but rather just
// removes `CustomQuotas` parameter.
func resourceVcdIpSpaceCustomQuota() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIpSpaceCustomQuotaCreate,
		ReadContext:   resourceVcdIpSpaceCustomQuotaRead,
		UpdateContext: resourceVcdIpSpaceCustomQuotaUpdate,
		DeleteContext: resourceVcdIpSpaceCustomQuotaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdIpSpaceCustomQuotaImport,
		},

		Schema: map[string]*schema.Schema{
			"ip_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of IP Space",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID for setting custom quota",
			},
			"ip_range_quota": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IP range quota. '-1' - unlimited, '0' - no quota",
				ValidateFunc: IsIntAndAtLeast(-1),
			},
			"ip_prefix_quota": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "One or more IP prefixes within internal scope",
				Elem:        ipPrefixeQuota,
			},
		},
	}
}

var ipPrefixeQuota = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"prefix_length": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Prefix length",
			ValidateFunc: IsIntAndAtLeast(0),
		},
		"quota": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "IP Prefix Quota",
			ValidateFunc: IsIntAndAtLeast(-1), // -1 - unlimited, 0 - no quota
		},
	},
}

func resourceVcdIpSpaceCustomQuotaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdIpSpaceCustomQuotaCreateUpdate(ctx, d, meta, "create")
}

func resourceVcdIpSpaceCustomQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdIpSpaceCustomQuotaCreateUpdate(ctx, d, meta, "update")
}

func resourceVcdIpSpaceCustomQuotaCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	log.Printf("[TRACE] IP Space Custom Quota %s initiated", operation)

	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)
	ipSpace, err := vcdClient.GetIpSpaceById(ipSpaceId)
	if err != nil {
		return diag.Errorf("error getting IP Space by ID '%s': %s", ipSpaceId, err)
	}

	// The Custom Quota mechanism (or as it is called in IP Org Assignment) is being implicitly
	// created when Edge Gateway (T1) backed by Provider Gateway (T0) with assigned uplinks.
	// Instead of creating, one must find it and update Custom Quota values.
	orgAssignment, err := ipSpace.GetOrgAssignmentByOrgId(orgId)
	if err != nil {
		return diag.Errorf("error finding Org Assignment during %s: %s", operation, err)
	}

	newOrgAssignmentConfig := getIpSpaceOrgAssignmentType(d, orgAssignment.IpSpaceOrgAssignment.ID, orgAssignment.IpSpaceOrgAssignment.IPSpaceType)
	_, err = orgAssignment.Update(newOrgAssignmentConfig)
	if err != nil {
		return diag.Errorf("error updating custom quotas during %s: %s", operation, err)
	}
	d.SetId(orgAssignment.IpSpaceOrgAssignment.ID)

	return resourceVcdIpSpaceCustomQuotaRead(ctx, d, meta)
}

func resourceVcdIpSpaceCustomQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space Custom quota read initiated")

	vcdClient := meta.(*VCDClient)

	ipSpaceId := d.Get("ip_space_id").(string)
	ipSpace, err := vcdClient.GetIpSpaceById(ipSpaceId)
	if err != nil {
		return diag.Errorf("error getting IP Space by ID '%s': %s", ipSpaceId, err)
	}

	orgAssignment, err := ipSpace.GetOrgAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error getting IP Space Org Assignment by ID: %s", err)
	}

	err = setIpSpaceOrgAssignmentData(d, orgAssignment.IpSpaceOrgAssignment)
	if err != nil {
		return diag.Errorf("error storing data to state: %s", err)
	}

	return nil
}

func resourceVcdIpSpaceCustomQuotaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space Custom quota deletion initiated")

	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)
	ipSpace, err := vcdClient.GetIpSpaceById(ipSpaceId)
	if err != nil {
		return diag.Errorf("error getting IP Space by ID '%s': %s", ipSpaceId, err)
	}

	orgAssignment, err := ipSpace.GetOrgAssignmentByOrgId(orgId)
	if err != nil {
		return diag.Errorf("error finding Org Assignment: %s", err)
	}

	// Reset custom quota definitions
	orgAssignment.IpSpaceOrgAssignment.CustomQuotas = nil
	_, err = orgAssignment.Update(orgAssignment.IpSpaceOrgAssignment)
	if err != nil {
		return diag.Errorf("error removing custom quotas: %s", err)
	}

	return nil
}

func resourceVcdIpSpaceCustomQuotaImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] IP Space Custom Quota import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as ip-space-name.org-name")
	}
	vcdClient := meta.(*VCDClient)

	var ipSpace *govcd.IpSpace

	ipSpaceName := resourceURI[0]
	orgName := resourceURI[1]

	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org '%s': %s", orgName, err)
	}

	ipSpace, err = vcdClient.GetIpSpaceByName(ipSpaceName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Space '%s: %s", orgName, err)
	}

	dSet(d, "org_id", org.Org.ID)
	dSet(d, "ip_space_id", ipSpace.IpSpace.ID)

	orgAssignment, err := ipSpace.GetOrgAssignmentByOrgId(org.Org.ID)
	if err != nil {
		return nil, fmt.Errorf("error finding Org Assignment: %s", err)
	}

	d.SetId(orgAssignment.IpSpaceOrgAssignment.ID)

	return []*schema.ResourceData{d}, nil
}

func getIpSpaceOrgAssignmentType(d *schema.ResourceData, assignmentId, ipSpaceType string) *types.IpSpaceOrgAssignment {
	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)

	newOrgAssignmentConfig := types.IpSpaceOrgAssignment{
		ID:          assignmentId,
		IPSpaceType: ipSpaceType,
		IPSpaceRef:  &types.OpenApiReference{ID: ipSpaceId},
		OrgRef:      &types.OpenApiReference{ID: orgId},
	}

	// Extract IP Range (floating IP) quotas
	if d.Get("ip_range_quota").(string) != "" {
		floatingIps, _ := strconv.Atoi(d.Get("ip_range_quota").(string))

		// We get the Org Assignment and expect to update some of its values
		newOrgAssignmentConfig.CustomQuotas = &types.IpSpaceOrgAssignmentQuotas{
			FloatingIPQuota: &floatingIps,
		}
	} else {
		newOrgAssignmentConfig.CustomQuotas = &types.IpSpaceOrgAssignmentQuotas{
			FloatingIPQuota: nil,
		}
	}

	ipPrefixQuotas := d.Get("ip_prefix_quota").(*schema.Set)
	ipPrefixeQuotaSlice := ipPrefixQuotas.List()

	if len(ipPrefixeQuotaSlice) > 0 {

		prefixQuotas := make([]types.IpSpaceOrgAssignmentIPPrefixQuotas, len(ipPrefixeQuotaSlice))

		for ipPrefixQuotaIndex := range ipPrefixeQuotaSlice {
			singleIpPrefixQuota := ipPrefixeQuotaSlice[ipPrefixQuotaIndex]
			ipPrefixMap := singleIpPrefixQuota.(map[string]interface{})
			ipPrefixQuota := ipPrefixMap["quota"].(string)
			ipPrefixQuotaInt, _ := strconv.Atoi(ipPrefixQuota) // ignoring error as validation is enforce in schema

			ipPrefixLength := ipPrefixMap["prefix_length"].(string)
			ipPrefixLengthInt, _ := strconv.Atoi(ipPrefixLength)

			prefixQuotas[ipPrefixQuotaIndex] = types.IpSpaceOrgAssignmentIPPrefixQuotas{
				PrefixLength: &ipPrefixLengthInt,
				Quota:        &ipPrefixQuotaInt,
			}
		}

		newOrgAssignmentConfig.CustomQuotas.IPPrefixQuotas = prefixQuotas
	} else {
		newOrgAssignmentConfig.CustomQuotas.IPPrefixQuotas = nil
	}

	return &newOrgAssignmentConfig
}

func setIpSpaceOrgAssignmentData(d *schema.ResourceData, orgAssignment *types.IpSpaceOrgAssignment) error {
	// Store `nil` values by default
	dSet(d, "ip_range_quota", nil)
	err := d.Set("ip_prefix_quota", nil)
	if err != nil {
		return fmt.Errorf("error storing nil 'ip_prefix_quota': %s", err)
	}

	// Override values if they are present
	if orgAssignment.CustomQuotas != nil {

		// ip_range_quota
		if orgAssignment.CustomQuotas.FloatingIPQuota != nil {
			stringValue := strconv.Itoa(*orgAssignment.CustomQuotas.FloatingIPQuota)
			dSet(d, "ip_range_quota", stringValue)
		} else {
			dSet(d, "ip_range_quota", nil)
		}

		// ip_prefix_quota
		prefixQuotaInterface := make([]interface{}, len(orgAssignment.CustomQuotas.IPPrefixQuotas))
		if len(orgAssignment.CustomQuotas.IPPrefixQuotas) > 0 {

			for i, val := range orgAssignment.CustomQuotas.IPPrefixQuotas {
				singlePrefixQuota := make(map[string]interface{})

				strQuotaPrefixLength := strconv.Itoa(*val.PrefixLength)
				singlePrefixQuota["prefix_length"] = strQuotaPrefixLength

				strQuota := strconv.Itoa(*val.Quota)
				singlePrefixQuota["quota"] = strQuota

				prefixQuotaInterface[i] = singlePrefixQuota
			}

			err := d.Set("ip_prefix_quota", prefixQuotaInterface)
			if err != nil {
				return fmt.Errorf("error storing 'ip_prefix_quota': %s", err)
			}
		}

	}

	return nil
}
