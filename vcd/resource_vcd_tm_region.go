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
		// UpdateContext: resourceVcdTmRegionUpdate, // TODO: TM: Update is not yet supported
		DeleteContext: resourceVcdTmRegionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmRegionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // TODO: TM: remove once update works
				Description: fmt.Sprintf("%s name", labelTmRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true, // TODO: TM: remove once update works
				Description: fmt.Sprintf("%s description", labelTmRegion),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true, // TODO: TM: remove once update works
				Default:     true,
				Description: fmt.Sprintf("Defines whether the %s is enabled or not", labelTmRegion),
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX Manager ID",
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true, // TODO: TM: remove once update works
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelTmRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policy_names": { // TODO: TM: check if the API accepts IDs and if it should use
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true, // TODO: TM: remove once update works
				Description: "A set of storage policy names",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
				Description: fmt.Sprintf("Status of the %s", labelTmRegion),
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

// TODO: TM: Update is not yet supported
// func resourceVcdTmRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	vcdClient := meta.(*VCDClient)
// 	c := crudConfig[*govcd.Region, types.Region]{
// 		entityLabel:      labelTmRegion,
// 		getTypeFunc:      getRegionType,
// 		getEntityFunc:    vcdClient.GetRegionById,
// 		resourceReadFunc: resourceVcdTmRegionRead,
// 	}
// 	return updateResource(ctx, d, meta, c)
// }

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

func getRegionType(vcdClient *VCDClient, d *schema.ResourceData) (*types.Region, error) {
	t := &types.Region{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		NsxManager:  &types.OpenApiReference{ID: d.Get("nsx_manager_id").(string)},
		IsEnabled:   d.Get("is_enabled").(bool),
	}

	// API requires Names to be sent with IDs, but Terraform native approach is to use IDs only
	// therefore names need to be looked up for IDs
	supervisorIds := convertSchemaSetToSliceOfStrings(d.Get("supervisor_ids").(*schema.Set))
	superVisorReferences := make([]types.OpenApiReference, 0)
	for _, singleSupervisorId := range supervisorIds {
		supervisor, err := vcdClient.GetSupervisorById(singleSupervisorId)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Supervisor with ID %s: %s", singleSupervisorId, err)
		}

		superVisorReferences = append(superVisorReferences, types.OpenApiReference{
			ID:   supervisor.Supervisor.SupervisorID,
			Name: supervisor.Supervisor.Name,
		})
	}
	t.Supervisors = superVisorReferences

	storagePolicyNames := convertSchemaSetToSliceOfStrings(d.Get("storage_policy_names").(*schema.Set))
	t.StoragePolicies = storagePolicyNames

	return t, nil
}

func setRegionData(_ *VCDClient, d *schema.ResourceData, r *govcd.Region) error {
	if r == nil || r.Region == nil {
		return fmt.Errorf("nil Region entity")
	}

	d.SetId(r.Region.ID)
	dSet(d, "name", r.Region.Name)
	dSet(d, "description", r.Region.Description)
	// dSet(d, "is_enabled", r.Region.IsEnabled) // TODO: TM: region is reported as false even when sending true
	dSet(d, "nsx_manager_id", r.Region.NsxManager.ID)

	dSet(d, "cpu_capacity_mhz", r.Region.CPUCapacityMHz)
	dSet(d, "cpu_reservation_capacity_mhz", r.Region.CPUReservationCapacityMHz)
	dSet(d, "memory_capacity_mib", r.Region.MemoryCapacityMiB)
	dSet(d, "memory_reservation_capacity_mib", r.Region.MemoryReservationCapacityMiB)
	dSet(d, "status", r.Region.Status)

	err := d.Set("supervisor_ids", extractIdsFromOpenApiReferences(r.Region.Supervisors))
	if err != nil {
		return fmt.Errorf("error storing 'supervisors': %s", err)
	}

	err = d.Set("storage_policy_names", r.Region.StoragePolicies)
	if err != nil {
		return fmt.Errorf("error storing 'storage_policy_names': %s", err)
	}

	return nil
}
