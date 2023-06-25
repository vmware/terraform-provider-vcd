package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdIpSpaceCustomQuota() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdIpSpaceCustomQuotaRead,

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
				Description: "Organization ID for which custom quota is set",
			},
			"ip_range_quota": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP range quota. '-1' - unlimited, '0' - no quota",
			},
			"ip_prefix_quota": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "One or more IP prefixes within internal scope",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix_length": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Prefix length",
						},
						"quota": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP Prefix Quota",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdIpSpaceCustomQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space Custom Quota DS read initiated")

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

	err = setIpSpaceOrgAssignmentData(d, orgAssignment.IpSpaceOrgAssignment)
	if err != nil {
		return diag.Errorf("error storing Org Assignment: %s", err)
	}
	d.SetId(orgAssignment.IpSpaceOrgAssignment.ID)

	return nil
}
