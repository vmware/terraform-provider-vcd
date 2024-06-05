package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdVdcTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVdcTemplateCreate,
		UpdateContext: resourceVcdVdcTemplateUpdate,
		ReadContext:   resourceVcdVdcTemplateRead,
		DeleteContext: resourceVcdVdcTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVdcTemplateImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template",
			},
			"network_provider_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Type of network provider. One of: 'NSX_V' or 'NSX_T'",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"NSX_V", "NSX_T"}, false)),
			},
			"provider_vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Provider VDC that the VDCs instantiated from this template will use",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the External network that the VDCs instantiated from this template will use",
			},
			"nsxv_primary_edge_cluster_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "NSX-V only: ID of the Edge Cluster that the VDCs instantiated from this template will use as primary",
				ConflictsWith: []string{"nsxt_gateway_edge_cluster_id", "nsxt_services_edge_cluster_id"},
			},
			"nsxv_secondary_edge_cluster_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "NSX-V only: ID of the Edge Cluster that the VDCs instantiated from this template will use as secondary",
				ConflictsWith: []string{"nsxt_gateway_edge_cluster_id", "nsxt_services_edge_cluster_id"},
			},
			"nsxt_gateway_edge_cluster_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "NSX-T only: ID of the Edge Cluster that the VDCs instantiated from this template will use with the NSX-T Gateway",
				ConflictsWith: []string{"nsxv_primary_edge_cluster_id", "nsxv_secondary_edge_cluster_id"},
			},
			"nsxt_services_edge_cluster_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "NSX-T only: ID of the Edge Cluster that the VDCs instantiated from this template will use for services",
				ConflictsWith: []string{"nsxv_primary_edge_cluster_id", "nsxv_secondary_edge_cluster_id"},
			},
			"allocation_model": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Allocation model that the VDCs instantiated from this template will use. Must be one of: 'AllocationVApp', 'AllocationPool', 'ReservationPool' or 'Flex'}",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}, false)),
			},
			// TODO: Missing CPU, memory and so on
			"storage_profile": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "Storage profiles that the VDCs instantiated from this template will use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of VDC storage profile",
						},
						"default": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified.",
						},
						"storage_used_in_mb": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Storage used in MB",
						},
					},
				},
			},
			"enable_fast_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If 'true', the VDCs instantiated from this template will have Fast provisioning enabled",
			},
			"thin_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If 'true', the VDCs instantiated from this template will have Thin provisioning enabled",
			},
			"edge_gateway": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "VDCs instantiated from this template will create a new Edge Gateway with the provided setup",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the Edge Gateway",
						},
						"description": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Description of the Edge Gateway",
						},
						"ip_allocation_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          0,
							Description:      "Storage used in MB",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
						},
						"network_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the network to create with the Edge Gateway",
						},
						"network_description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the network to create with the Edge Gateway",
						},
						"gateway_cidr": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "CIDR of the Edge Gateway",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
						},
					},
				},
			},
			"edge_gateway_static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP ranges used for the network created with the Edge Gateway. Only required if the 'edge_gateway' block is used",
				Elem:        networkV2IpRange,
			},
			"network_pool_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "If set, specifies the Network pool for the instantiated VDCs",
			},
			"nics_quota": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				Description:      "Quota for the NICs of the instantiated VDCs. 0 means unlimited",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"provisioned_networks_quota": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				Description:      "Quota for the provisioned networks of the instantiated VDCs. 0 means unlimited",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"view_and_instantiate_org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of the Organizations that will be able to view and instantiate this VDC template",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vdc_template_system_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template as seen by the System administrator",
			},
			"vdc_template_tenant_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template as seen by the tenants (organizations)",
			},
			"vdc_template_system_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the VDC Template as seen by the System administrator",
			},
			"vdc_template_tenant_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the VDC Template as seen by the tenants (organizations)",
			},
		},
	}
}

func resourceVcdVdcTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Checks:
	// NSX-V edge clusters are used when type=NSX_V
	// NSX-T edge clusters are used when type=NSX_V
	// edge_gateway_static_ip_pool is present if and only if the edge_gateway block is present
	return resourceVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	err = vdcTemplate.Delete()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdVdcTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vdcTemplate, err := meta.(*VCDClient).GetVdcTemplateByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("could not import VDC Template with name %s: %s", d.Id(), err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	d.SetId(vdcTemplate.VdcTemplate.ID)
	return []*schema.ResourceData{d}, nil
}

func genericVcdVdcTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	d.SetId(vdcTemplate.VdcTemplate.ID)

	return nil
}

// getVdcTemplate retrieves a VDC Template with the available information in the configuration.
func getVdcTemplate(d *schema.ResourceData, vcdClient *VCDClient) (*govcd.VdcTemplate, error) {
	var vdcTemplate *govcd.VdcTemplate
	var err error
	if d.Id() == "" {
		name := d.Get("name").(string)
		vdcTemplate, err = vcdClient.GetVdcTemplateByName(name)
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with name %s: %s", name, err)
		}
	} else {
		vdcTemplate, err = vcdClient.GetVdcTemplateById(d.Id())
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with ID %s: %s", d.Id(), err)
		}
	}
	return vdcTemplate, nil
}
