package vcd

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbCloud() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbCloudCreate,
		ReadContext:   resourceVcdAlbCloudRead,
		// Update of NSX-T ALB Cloud configuration is not supported in VCD <= 10.3
		// UpdateContext: resourceVcdAlbCloudUpdate,
		DeleteContext: resourceVcdAlbCloudDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbCloudImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Cloud name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Cloud description",
			},
			"controller_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Controller ID",
			},
			"importable_cloud_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Importable Cloud ID",
			},
			"network_pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Network pool ID for NSX-T ALB Importable Cloud",
			},
			"network_pool_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool name of NSX-T ALB Cloud",
			},
			"health_status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Cloud health status",
			},
			"health_message": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Cloud detailed health message",
			},
		},
	}
}

func resourceVcdAlbCloudCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albCloudConfig := getNsxtAlbCloudType(d)
	createdAlbCloud, err := vcdClient.CreateAlbCloud(albCloudConfig)
	if err != nil {
		return diag.Errorf("error creating NSX-T ALB Cloud '%s': %s", albCloudConfig.Name, err)
	}

	d.SetId(createdAlbCloud.NsxtAlbCloud.ID)

	return resourceVcdAlbCloudRead(ctx, d, meta)
}

func resourceVcdAlbCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albCloud, err := vcdClient.GetAlbCloudById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to find NSX-T ALB Cloud: %s", err)
	}

	setNsxtAlbCloudData(d, albCloud.NsxtAlbCloud)

	return nil
}

func resourceVcdAlbCloudDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albCloud, err := vcdClient.GetAlbCloudById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Cloud: %s", err)
	}

	err = albCloud.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T ALB Cloud: %s", err)
	}

	return nil
}

func resourceVcdAlbCloudImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("this resource is only supported for Providers")
	}

	resourceURI := d.Id()
	albCloud, err := vcdClient.GetAlbCloudByName(resourceURI)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Cloud with Name '%s': %s", d.Id(), err)
	}

	d.SetId(albCloud.NsxtAlbCloud.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxtAlbCloudType(d *schema.ResourceData) *types.NsxtAlbCloud {
	albCloudType := &types.NsxtAlbCloud{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		LoadBalancerCloudBacking: types.NsxtAlbCloudBacking{
			BackingId: d.Get("importable_cloud_id").(string),
			LoadBalancerControllerRef: types.OpenApiReference{
				ID: d.Get("controller_id").(string),
			},
		},
		NetworkPoolRef: &types.OpenApiReference{ID: d.Get("network_pool_id").(string)},
	}

	return albCloudType
}

func setNsxtAlbCloudData(d *schema.ResourceData, albCloud *types.NsxtAlbCloud) {
	_ = d.Set("name", albCloud.Name)
	_ = d.Set("description", albCloud.Description)
	_ = d.Set("health_status", albCloud.HealthStatus)
	_ = d.Set("health_message", albCloud.DetailedHealthMessage)
	_ = d.Set("importable_cloud_id", albCloud.LoadBalancerCloudBacking.BackingId)
	_ = d.Set("controller_id", albCloud.LoadBalancerCloudBacking.LoadBalancerControllerRef.ID)

	if albCloud.NetworkPoolRef != nil {
		_ = d.Set("network_pool_name", albCloud.NetworkPoolRef.Name)
		_ = d.Set("network_pool_id", albCloud.NetworkPoolRef.ID)
	}
}
