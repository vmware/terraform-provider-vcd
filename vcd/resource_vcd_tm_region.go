package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdTmRegion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmRegionCreate,
		ReadContext:   resourceVcdTmRegionRead,
		UpdateContext: resourceVcdTmRegionUpdate,
		DeleteContext: resourceVcdTmRegionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmRegionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Region name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Region description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Defines whether the Region is enabled or not",
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX Manager ID",
			},

			"cpu_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"cpu_reservation_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"memory_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"memory_reservation_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status",
			},
			"supervisors": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set supervisor IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policies": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set storage policies",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdTmRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getRegionType(d)
	if err != nil {
		return diag.Errorf("error getting Region type: %s", err)
	}

	region, err := vcdClient.CreateRegion(t)
	if err != nil {
		return diag.Errorf("error creating Region: %s", err)
	}

	d.SetId(region.Region.ID)

	return resourceVcdTmRegionRead(ctx, d, meta)
}

func resourceVcdTmRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if tmUpdateFuse {
		return diag.Errorf("UPDATE FUSE ENABLED")
	}

	vcdClient := meta.(*VCDClient)
	region, err := vcdClient.GetRegionById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Region: %s", err)
	}

	t, err := getRegionType(d)
	if err != nil {
		return diag.Errorf("error getting Region type: %s", err)
	}

	region.Update(t)

	return resourceVcdTmRegionRead(ctx, d, meta)
}

func resourceVcdTmRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	region, err := vcdClient.GetRegionById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving Region: %s", err)
	}

	err = setRegionData(d, region.Region)
	if err != nil {
		return diag.Errorf("error storing data: %s", err)
	}

	d.SetId(region.Region.ID)

	return nil
}

func resourceVcdTmRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if tmDeleteFuse {
		return diag.Errorf("DELETE FUSE ENABLED")
	}
	vcdClient := meta.(*VCDClient)
	region, err := vcdClient.GetRegionById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Region: %s", err)
	}

	err = region.Delete()
	if err != nil {
		return diag.Errorf("error deleting Region: %s", err)
	}
	return nil
}

func resourceVcdTmRegionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	region, err := vcdClient.GetRegionByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving Region: %s", err)
	}

	d.SetId(region.Region.ID)

	return []*schema.ResourceData{d}, nil
}

func getRegionType(d *schema.ResourceData) (*types.Region, error) {
	t := &types.Region{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		NsxManager:  &types.OpenApiReference{ID: d.Get("nsx_manager_id").(string)},
	}

	return t, nil
}

func setRegionData(d *schema.ResourceData, region *types.Region) error {
	dSet(d, "name", region.Name)
	dSet(d, "description", region.Description)
	dSet(d, "is_enabled", region.IsEnabled)
	dSet(d, "nsx_manager_id", region.NsxManager.ID)

	dSet(d, "cpu_capacity_mhz", region.CPUCapacityMHz)
	dSet(d, "cpu_reservation_capacity_mhz", region.CPUReservationCapacityMHz)
	dSet(d, "memory_capacity_mhz", region.MemoryCapacityMiB)
	dSet(d, "memory_reservation_capacity_mhz", region.MemoryReservationCapacityMiB)
	dSet(d, "status", region.Status)

	err := d.Set("supervisors", extractIdsFromOpenApiReferences(region.Supervisors))
	if err != nil {
		return fmt.Errorf("error storing 'supervisors': %s", err)
	}

	err = d.Set("storage_policies", region.StoragePolicies)
	if err != nil {
		return fmt.Errorf("error storing 'storage_policies': %s", err)
	}

	return nil
}
