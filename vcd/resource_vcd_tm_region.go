package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmRegion = "TM Region"

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
				Description: fmt.Sprintf("%s name", labelTmRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("%s description", labelTmRegion),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Defines whether the %s is enabled or not", labelTmRegion),
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX Manager ID",
			},
			"cpu_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU Capacity in MHz",
			},
			"cpu_reservation_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU reservation in MHz",
			},
			"memory_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory capacity in MiB",
			},
			"memory_reservation_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory reservation in MiB",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the region",
			},
			"supervisors": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelTmRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policies": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of storage policies",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdTmRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:      labelTmRegion,
		getTypeFunc:      getRegionType,
		stateStoreFunc:   setRegionData,
		createFunc:       vcdClient.CreateRegion,
		resourceReadFunc: resourceVcdTmRegionRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:      labelTmRegion,
		getTypeFunc:      getRegionType,
		getEntityFunc:    vcdClient.GetRegionById,
		resourceReadFunc: resourceVcdTmRegionRead,
		// preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.Region, *types.Region]{},
	}
	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:    labelTmRegion,
		getEntityFunc:  vcdClient.GetRegionById,
		stateStoreFunc: setRegionData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:   labelTmRegion,
		getEntityFunc: vcdClient.GetRegionById,
		// preDeleteHooks: []outerEntityHook[*govcd.Region]{},
	}

	return deleteResource(ctx, d, meta, c)
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

func setRegionData(d *schema.ResourceData, r *govcd.Region) error {
	if r == nil || r.Region == nil {
		return fmt.Errorf("nil entity")

	}

	d.SetId(r.Region.ID)
	dSet(d, "name", r.Region.Name)
	dSet(d, "description", r.Region.Description)
	dSet(d, "is_enabled", r.Region.IsEnabled)
	dSet(d, "nsx_manager_id", r.Region.NsxManager.ID)

	dSet(d, "cpu_capacity_mhz", r.Region.CPUCapacityMHz)
	dSet(d, "cpu_reservation_capacity_mhz", r.Region.CPUReservationCapacityMHz)
	dSet(d, "memory_capacity_mib", r.Region.MemoryCapacityMiB)
	dSet(d, "memory_reservation_capacity_mib", r.Region.MemoryReservationCapacityMiB)
	dSet(d, "status", r.Region.Status)

	err := d.Set("supervisors", extractIdsFromOpenApiReferences(r.Region.Supervisors))
	if err != nil {
		return fmt.Errorf("error storing 'supervisors': %s", err)
	}

	err = d.Set("storage_policies", r.Region.StoragePolicies)
	if err != nil {
		return fmt.Errorf("error storing 'storage_policies': %s", err)
	}

	return nil
}
