package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdVmSizingPolicy() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdVmSizingPolicyRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Unneeded property, which was included by mistake",
				Description: "The name of organization to use - Deprecated and unneeded: will be ignored if used ",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"speed_in_mhz": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the vCPU speed of a core in MHz.",
						},
						"count": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the number of vCPUs configured for a VM. This is a VM hardware configuration. When a tenant assigns the VM sizing policy to a VM, this count becomes the configured number of vCPUs for the VM.",
						},
						"cores_per_socket": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The number of cores per socket for a VM. This is a VM hardware configuration. The number of vCPUs that is defined in the VM sizing policy must be divisible by the number of cores per socket. If the number of vCPUs is not divisible by the number of cores per socket, the number of cores per socket becomes invalid.",
						},
						"reservation_guarantee": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines how much of the CPU resources of a VM are reserved. The allocated CPU for a VM equals the number of vCPUs times the vCPU speed in MHz. The value of the attribute ranges between 0 and one. Value of 0 CPU reservation guarantee defines no CPU reservation. Value of 1 defines 100% of CPU reserved.",
						},
						"limit_in_mhz": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the CPU limit in MHz for a VM. If not defined in the VDC compute policy, CPU limit is equal to the vCPU speed multiplied by the number of vCPUs.",
						},
						"shares": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the number of CPU shares for a VM. Shares specify the relative importance of a VM within a virtual data center. If a VM has twice as many shares of CPU as another VM, it is entitled to consume twice as much CPU when these two virtual machines are competing for resources. If not defined in the VDC compute policy, normal shares are applied to the VM.",
						},
					},
				},
			},
			"memory": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_mb": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the memory configured for a VM in MB. This is a VM hardware configuration. When a tenant assigns the VM sizing policy to a VM, the VM receives the amount of memory defined by this attribute.",
						},
						"reservation_guarantee": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the reserved amount of memory that is configured for a VM. The value of the attribute ranges between 0 and one. Value of 0 memory reservation guarantee defines no memory reservation. Value of 1 defines 100% of memory reserved.",
						},
						"limit_in_mb": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the memory limit in MB for a VM. If not defined in the VM sizing policy, memory limit is equal to the allocated memory for the VM.",
						},
						"shares": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines the number of memory shares for a VM. Shares specify the relative importance of a VM within a virtual data center. If a VM has twice as many shares of memory as another VM, it is entitled to consume twice as much memory when these two virtual machines are competing for resources. If not defined in the VDC compute policy, normal shares are applied to the VM.",
						},
					},
				},
			},
		},
	}
}

// datasourceVcdVmSizingPolicyRead reads a data source VM sizing policy
func datasourceVcdVmSizingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmSizingPolicyRead(ctx, d, meta)
}
