package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmVdc = "TM Vdc"

func resourceTmVdc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTmVdcCreate,
		ReadContext:   resourceTmVdcRead,
		UpdateContext: resourceTmVdcUpdate,
		DeleteContext: resourceTmVdcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTmVdcImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelTmVdc),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of the %s", labelTmVdc),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Defines if the %s is enabled", labelTmVdc),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Organization ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Region ID",
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelTmVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        tmVdcZoneResourceAllocation,
				Description: "A set of Supervisor Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelTmVdc),
			},
		},
	}
}

var tmVdcZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Supervisor Zone Name",
		},
		"zone_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Supervisor Zone ID",
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory limit in MiB",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory reservation in MiB",
		},
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "CPU limit in MHz",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "CPU reservation in MHz",
		},
	},
}

func resourceTmVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmVdc,
		getTypeFunc:      getTmVdcType,
		stateStoreFunc:   setTmVdcData,
		createFunc:       vcdClient.CreateTmVdc,
		resourceReadFunc: resourceTmVdcRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceTmVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmVdc,
		getTypeFunc:      getTmVdcType,
		getEntityFunc:    vcdClient.GetTmVdcById,
		resourceReadFunc: resourceTmVdcRead,
		// preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmVdc, *types.TmVdc]{},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceTmVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelTmVdc,
		getEntityFunc:  vcdClient.GetTmVdcById,
		stateStoreFunc: setTmVdcData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceTmVdcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:   labelTmVdc,
		getEntityFunc: vcdClient.GetTmVdcById,
		// preDeleteHooks: []outerEntityHook[*govcd.TmVdc]{},
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceTmVdcImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	vdc, err := vcdClient.GetTmVdcByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelTmVdc, err)
	}

	d.SetId(vdc.TmVdc.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmVdcType(_ *VCDClient, d *schema.ResourceData) (*types.TmVdc, error) {
	t := &types.TmVdc{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		// IsEnabled:   d.Get("is_enabled").(bool),
		Org:    &types.OpenApiReference{ID: d.Get("org_id").(string)},
		Region: &types.OpenApiReference{ID: d.Get("region_id").(string)},
	}

	supervisorIds := convertSchemaSetToSliceOfStrings(d.Get("supervisor_ids").(*schema.Set))
	t.Supervisors = convertSliceOfStringsToOpenApiReferenceIds(supervisorIds)

	zra := d.Get("zone_resource_allocations").(*schema.Set)
	r := make([]*types.TmVdcZoneResourceAllocation, zra.Len())
	for zoneIndex, singleZone := range zra.List() {
		singleZoneMap := singleZone.(map[string]interface{})
		singleZoneType := &types.TmVdcZoneResourceAllocation{
			Zone: &types.OpenApiReference{
				ID: singleZoneMap["zone_id"].(string),
			},
			ResourceAllocation: types.TmVdcResourceAllocation{
				CPULimitMHz:          singleZoneMap["cpu_limit_mhz"].(int),
				CPUReservationMHz:    singleZoneMap["cpu_reservation_mhz"].(int),
				MemoryLimitMiB:       singleZoneMap["memory_limit_mib"].(int),
				MemoryReservationMiB: singleZoneMap["memory_reservation_mib"].(int),
			},
		}
		r[zoneIndex] = singleZoneType
	}
	t.ZoneResourceAllocation = r

	return t, nil
}

func setTmVdcData(d *schema.ResourceData, vdc *govcd.TmVdc) error {
	if vdc == nil {
		return fmt.Errorf("provided VDC is nil")
	}

	d.SetId(vdc.TmVdc.ID)
	dSet(d, "name", vdc.TmVdc.Name)
	dSet(d, "description", vdc.TmVdc.Description)
	// dSet(d, "is_enabled", vdc.TmVdc.IsEnabled) TODO: TM: the field is ineffective and always returns false
	dSet(d, "status", vdc.TmVdc.Status)

	orgId := ""
	if vdc.TmVdc.Org != nil {
		orgId = vdc.TmVdc.Org.ID
	}
	dSet(d, "org_id", orgId)

	regionId := ""
	if vdc.TmVdc.Region != nil {
		regionId = vdc.TmVdc.Region.ID
	}
	dSet(d, "region_id", regionId)

	supervisors := extractIdsFromOpenApiReferences(vdc.TmVdc.Supervisors)
	err := d.Set("supervisor_ids", supervisors)
	if err != nil {
		return fmt.Errorf("error storing 'supervisor_ids': %s", err)
	}

	zoneCompute := make([]interface{}, len(vdc.TmVdc.ZoneResourceAllocation))
	for zoneIndex, zone := range vdc.TmVdc.ZoneResourceAllocation {
		oneZone := make(map[string]interface{})

		oneZone["zone_name"] = zone.Zone.Name
		oneZone["zone_id"] = zone.Zone.ID

		oneZone["memory_limit_mib"] = zone.ResourceAllocation.MemoryLimitMiB
		oneZone["memory_reservation_mib"] = zone.ResourceAllocation.MemoryReservationMiB
		oneZone["cpu_limit_mhz"] = zone.ResourceAllocation.CPULimitMHz
		oneZone["cpu_reservation_mhz"] = zone.ResourceAllocation.CPUReservationMHz

		zoneCompute[zoneIndex] = oneZone
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(tmVdcZoneResourceAllocation), zoneCompute)
	err = d.Set("zone_resource_allocations", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'zone_resource_allocations' after read: %s", err)
	}

	return nil
}
