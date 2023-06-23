package vcd

import (
	"context"
	"log"
	"strings"

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
				Description: "IP Address or Prefix of the allocation",
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
				Description: "IP Allocation Description",
			},
			"prefix_length": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Exposed only if 'type' is 'IP_PREFIX'",
			},
			"usage_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "One of 'UNUSED', 'USED', 'USED_MANUAL",
			},
			"used_by_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An ID of entity this IP Allocation is assigned to",
			},
			"allocation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Allocation date in ISO 8601 format (e.g. 2023-06-07T09:57:58.721Z)",
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Allocated IP address",
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

	// When IP Prefix is allocated, the returned value is in CIDR format (e.g. 192.168.1.0/24), and
	// although it can be split using Terraform native functions, we're adding a convenience layer for
	// users by splitting this address into IP and prefix length
	if ipAllocation.IpSpaceIpAllocation.Type == "IP_PREFIX" {
		splitCidr := strings.Split(ipAllocation.IpSpaceIpAllocation.Value, "/")
		if len(splitCidr) == 2 {
			dSet(d, "ip", splitCidr[0])
			dSet(d, "prefix_length", splitCidr[1])
		}
	}

	if ipAllocation.IpSpaceIpAllocation.Type == "FLOATING_IP" {
		dSet(d, "ip", ipAllocation.IpSpaceIpAllocation.Value)
	}

	d.SetId(ipAllocation.IpSpaceIpAllocation.ID)

	return nil
}
