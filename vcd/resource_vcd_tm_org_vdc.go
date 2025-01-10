package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmOrgVdc = "TM Org Vdc"

func resourceTmOrgVdc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTmOrgVdcCreate,
		ReadContext:   resourceTmOrgVdcRead,
		UpdateContext: resourceTmOrgVdcUpdate,
		DeleteContext: resourceTmOrgVdcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTmOrgVdcImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Organization ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelTmOrgVdc),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of the %s", labelTmOrgVdc),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Defines if the %s is enabled", labelTmOrgVdc),
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
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelTmOrgVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        tmOrgVdcZoneResourceAllocation,
				Description: "A set of Region Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelTmOrgVdc),
			},
		},
	}
}

var tmOrgVdcZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s Name", labelTmRegionZone),
		},
		"region_zone_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("%s ID", labelTmRegionZone),
		},
		"memory_limit_mib": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "Memory limit in MiB",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"memory_reservation_mib": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "Memory reservation in MiB",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"cpu_limit_mhz": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "CPU limit in MHz",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"cpu_reservation_mhz": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "CPU reservation in MHz",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
	},
}

func resourceTmOrgVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmOrgVdc,
		getTypeFunc:      getTmVdcType,
		stateStoreFunc:   setTmVdcData,
		createFunc:       vcdClient.CreateTmVdc,
		resourceReadFunc: resourceTmOrgVdcRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceTmOrgVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmOrgVdc,
		getTypeFunc:      getTmVdcType,
		getEntityFunc:    vcdClient.GetTmVdcById,
		resourceReadFunc: resourceTmOrgVdcRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceTmOrgVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelTmOrgVdc,
		getEntityFunc:  vcdClient.GetTmVdcById,
		stateStoreFunc: setTmVdcData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceTmOrgVdcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:   labelTmOrgVdc,
		getEntityFunc: vcdClient.GetTmVdcById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceTmOrgVdcImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	vdc, err := vcdClient.GetTmVdcByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelTmOrgVdc, err)
	}

	d.SetId(vdc.TmVdc.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmVdcType(_ *VCDClient, d *schema.ResourceData) (*types.TmVdc, error) {
	t := &types.TmVdc{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		IsEnabled:   addrOf(d.Get("is_enabled").(bool)),
		Org:         &types.OpenApiReference{ID: d.Get("org_id").(string)},
		Region:      &types.OpenApiReference{ID: d.Get("region_id").(string)},
	}

	supervisorIds := convertSchemaSetToSliceOfStrings(d.Get("supervisor_ids").(*schema.Set))
	t.Supervisors = convertSliceOfStringsToOpenApiReferenceIds(supervisorIds)

	zra := d.Get("zone_resource_allocations").(*schema.Set)
	r := make([]*types.TmVdcZoneResourceAllocation, zra.Len())
	for zoneIndex, singleZone := range zra.List() {
		singleZoneMap := singleZone.(map[string]interface{})
		singleZoneType := &types.TmVdcZoneResourceAllocation{
			Zone: &types.OpenApiReference{
				ID: singleZoneMap["region_zone_id"].(string),
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

func setTmVdcData(_ *VCDClient, d *schema.ResourceData, vdc *govcd.TmVdc) error {
	if vdc == nil {
		return fmt.Errorf("provided VDC is nil")
	}

	d.SetId(vdc.TmVdc.ID)
	dSet(d, "name", vdc.TmVdc.Name)
	dSet(d, "description", vdc.TmVdc.Description)
	dSet(d, "is_enabled", vdc.TmVdc.IsEnabled)
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
		oneZone["region_zone_name"] = zone.Zone.Name
		oneZone["region_zone_id"] = zone.Zone.ID
		oneZone["memory_limit_mib"] = zone.ResourceAllocation.MemoryLimitMiB
		oneZone["memory_reservation_mib"] = zone.ResourceAllocation.MemoryReservationMiB
		oneZone["cpu_limit_mhz"] = zone.ResourceAllocation.CPULimitMHz
		oneZone["cpu_reservation_mhz"] = zone.ResourceAllocation.CPUReservationMHz

		zoneCompute[zoneIndex] = oneZone
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(tmOrgVdcZoneResourceAllocation), zoneCompute)
	err = d.Set("zone_resource_allocations", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'zone_resource_allocations' after read: %s", err)
	}

	return nil
}
