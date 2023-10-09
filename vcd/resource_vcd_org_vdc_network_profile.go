package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtOrgVdcNetworkProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtOrgVdcNetworkProfileCreateUpdate,
		ReadContext:   resourceVcdNsxtOrgVdcNetworkProfileRead,
		UpdateContext: resourceVcdNsxtOrgVdcNetworkProfileCreateUpdate,
		DeleteContext: resourceVcdNsxtOrgVdcNetworkProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVdcAccessControlImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of NSX-T Edge Cluster (provider vApp networking services and DHCP capability for Isolated networks)",
			},
			"vdc_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default NSX-T Segment Profile for Org VDC networks",
			},
			"vapp_networks_default_segment_profile_template_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default NSX-T Segment Profile for vApp networks",
			},
		},
	}
}

func resourceVcdNsxtOrgVdcNetworkProfileCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error when retrieving VDC: %s", err)
	}

	if !vdc.IsNsxt() {
		return diag.Errorf("network profile configuration is only supported on NSX-T VDCs")
	}

	vdcNetworkProfileConfig := &types.VdcNetworkProfile{}

	if d.Get("edge_cluster_id").(string) != "" {
		vdcNetworkProfileConfig.ServicesEdgeCluster = &types.VdcNetworkProfileServicesEdgeCluster{BackingID: d.Get("edge_cluster_id").(string)}
	}

	if d.Get("vdc_networks_default_segment_profile_template_id").(string) != "" {
		vdcNetworkProfileConfig.VdcNetworkSegmentProfileTemplateRef = &types.OpenApiReference{ID: d.Get("vdc_networks_default_segment_profile_template_id").(string)}
	}

	if d.Get("vapp_networks_default_segment_profile_template_id").(string) != "" {
		vdcNetworkProfileConfig.VappNetworkSegmentProfileTemplateRef = &types.OpenApiReference{ID: d.Get("vapp_networks_default_segment_profile_template_id").(string)}
	}

	_, err = vdc.UpdateVdcNetworkProfile(vdcNetworkProfileConfig)
	if err != nil {
		return diag.Errorf("error updating VDC network profile configuration: %s", err)
	}

	d.SetId(vdc.Vdc.ID)
	return resourceVcdNsxtOrgVdcNetworkProfileRead(ctx, d, meta)
}

func resourceVcdNsxtOrgVdcNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx, d, meta, "resource")
}

func dataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx, d, meta, "datasource")
}

func resourceDataSourceVcdNsxtOrgVdcNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error when retrieving VDC: %s", err)
	}

	netProfile, err := vdc.GetVdcNetworkProfile()
	if err != nil {
		return diag.Errorf("error getting VDC Network Profile: %s", err)
	}

	dSet(d, "edge_cluster_id", "")
	if netProfile.ServicesEdgeCluster != nil && netProfile.ServicesEdgeCluster.BackingID != "" {
		dSet(d, "edge_cluster_id", netProfile.ServicesEdgeCluster.BackingID)
	}

	dSet(d, "vapp_networks_default_segment_profile_template_id", "")
	if netProfile.VappNetworkSegmentProfileTemplateRef != nil {
		dSet(d, "vapp_networks_default_segment_profile_template_id", netProfile.VappNetworkSegmentProfileTemplateRef.ID)
	}

	dSet(d, "vdc_networks_default_segment_profile_template_id", "")
	if netProfile.VdcNetworkSegmentProfileTemplateRef != nil {
		dSet(d, "vdc_networks_default_segment_profile_template_id", netProfile.VdcNetworkSegmentProfileTemplateRef.ID)
	}

	d.SetId(vdc.Vdc.ID)
	return nil
}

func resourceVcdNsxtOrgVdcNetworkProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error when retrieving VDC: %s", err)
	}

	_, err = vdc.UpdateVdcNetworkProfile(&types.VdcNetworkProfile{})
	if err != nil {
		return diag.Errorf("error deleting VDC network profile configuration: %s", err)
	}

	return nil
}
