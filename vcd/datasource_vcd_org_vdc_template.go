package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdOrgVdcTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVdcTemplateRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template as seen by the System administrator",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the VDC Template as seen by the System administrator",
			},
			"tenant_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the VDC Template as seen by the tenants (organizations)",
			},
			"tenant_description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the VDC Template as seen by the tenants (organizations)",
			},
			"provider_vdc": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A Provider VDC that the VDCs instantiated from this template use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of Provider VDC",
						},
						"external_network_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the External network that the VDCs instantiated from this template use",
						},
						"gateway_edge_cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the Edge Cluster that the VDCs instantiated from this template use with the NSX-T Gateway",
						},
						"services_edge_cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the Edge Cluster that the VDCs instantiated from this template use for services",
						},
					},
				},
			},
			"allocation_model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Allocation model that the VDCs instantiated from this template use",
			},
			"compute_configuration": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The compute configuration for the VDCs instantiated from this template",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_allocated": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum amount of CPU, in MHz, available to the VMs running within the VDC that is instantiated from this template",
						},
						"cpu_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The limit amount of CPU, in MHz, of the VDC that is instantiated from this template. 0 means unlimited",
						},
						"cpu_guaranteed": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The percentage of the CPU guaranteed to be available to VMs running within the VDC instantiated from this template",
						},
						"cpu_speed": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Specifies the clock frequency, in MHz, for any virtual CPU that is allocated to a VM",
						},
						"memory_allocated": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum amount of Memory, in MB, available to the VMs running within the VDC that is instantiated from this template",
						},
						"memory_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The limit amount of Memory, in MB, of the VDC that is instantiated from this template. 0 means unlimited",
						},
						"memory_guaranteed": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The percentage of the Memory guaranteed to be available to VMs running within the VDC instantiated from this template",
						},
						"elasticity": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if compute capacity can grow or shrink based on demand",
						},
						"include_vm_memory_overhead": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the instantiated VDC includes memory overhead into its accounting for admission control",
						},
					},
				},
			},
			"storage_profile": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Storage profiles that the VDCs instantiated from this template use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of Provider VDC storage profile to use for the VDCs instantiated from this template",
						},
						"default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if this is default storage profile for the VDCs instantiated from this template",
						},
						"limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Storage limit of the VDCs instantiated from this template, in Megabytes. 0 means unlimited",
						},
					},
				},
			},
			"enable_fast_provisioning": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If 'true', the VDCs instantiated from this template have Fast provisioning enabled",
			},
			"enable_thin_provisioning": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If 'true', the VDCs instantiated from this template have Thin provisioning enabled",
			},
			"edge_gateway": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "VDCs instantiated from this template create a new Edge Gateway with the provided setup",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Edge Gateway",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the Edge Gateway",
						},
						"ip_allocation_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Storage used in MB",
						},
						"routed_network_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the routed network to create with the Edge Gateway",
						},
						"routed_network_description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the routed network to create with the Edge Gateway",
						},
						"routed_network_gateway_cidr": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "CIDR of the Edge Gateway for the routed network created with the Edge Gateway",
						},
						"static_ip_pool": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "IP ranges used for the network created with the Edge Gateway. Only required if the 'edge_gateway' block is used",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Start address of the IP range",
									},
									"end_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "End address of the IP range",
									},
								},
							},
						},
					},
				},
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Network pool of the instantiated VDCs. If empty, means that it is automatically chosen",
			},
			"nic_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Quota of the NICs of the instantiated VDCs. 0 means unlimited",
			},
			"vm_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Quota for the VMs of the instantiated VDCs. 0 means unlimited",
			},
			"provisioned_network_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Quota for the provisioned networks of the instantiated VDCs. 0 means unlimited",
			},
			"readable_by_org_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IDs of the Organizations that will be able to view and instantiate this VDC template. This attribute is not available for tenants",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta, "datasource")
}
