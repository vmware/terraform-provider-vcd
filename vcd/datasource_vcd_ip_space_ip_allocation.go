package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdIpAllocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdIpAllocationRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Required if 'type' is IP_PREFIX",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "",
			},
			"prefix_length": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Required if 'type' is IP_PREFIX",
			},
			"usage_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "",
			},
			"used_by_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"allocation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		},
	}
}

func datasourceVcdIpAllocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space IP Allocation DS read initiated")

	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)
	ipAddress := d.Get("ip_address").(string)
	ipAllocationType := d.Get("type").(string)

	org, err := vcdClient.GetOrgById(orgId)
	if err != nil {
		return diag.Errorf("error getting Org by id: %s", err)
	}

	// org.GetIpSpaceAllocationByValue(ipSpaceId, ipAllocationValue, nil)

	ipAllocation, err := org.GetIpSpaceAllocationByTypeAndValue(ipSpaceId, ipAllocationType, ipAddress, nil)
	if err != nil {
		return diag.Errorf("error getting IP Allocation: %s", err)
	}

	dSet(d, "description", ipAllocation.IpSpaceIpAllocation.Description)
	dSet(d, "type", ipAllocation.IpSpaceIpAllocation.Type)
	if ipAllocation.IpSpaceIpAllocation.OrgRef != nil {
		dSet(d, "org_id", ipAllocation.IpSpaceIpAllocation.OrgRef.ID)
	}
	if ipAllocation.IpSpaceIpAllocation.UsedByRef != nil {
		// used_by_id
		dSet(d, "used_by_id", ipAllocation.IpSpaceIpAllocation.UsedByRef.ID)
	}
	dSet(d, "allocation_date", ipAllocation.IpSpaceIpAllocation.AllocationDate)
	dSet(d, "usage_state", ipAllocation.IpSpaceIpAllocation.UsageState)

	d.SetId(ipAllocation.IpSpaceIpAllocation.ID)

	return nil
}
