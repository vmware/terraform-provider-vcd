package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdTmRegionStoragePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmRegionStoragePolicyCreate,
		ReadContext:   resourceVcdTmRegionStoragePolicyRead,
		UpdateContext: resourceVcdTmRegionStoragePolicyUpdate,
		DeleteContext: resourceVcdTmRegionStoragePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmRegionStoragePolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Region Storage Policy name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Region Storage Policy",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Region that this Region Storage Policy belongs to",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation status of the Region Storage Policy. Can be [NOT_READY, READY]",
			},
			"storage_capacity_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Storage capacity in megabytes for this Region Storage Policy",
			},
			"storage_consumed_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Consumed storage in megabytes for this Region Storage Policy",
			},
		},
	}
}

func resourceVcdTmRegionStoragePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getRegionStoragePolicyType(d)
	if err != nil {
		return diag.Errorf("error getting Region Storage Policy type: %s", err)
	}

	region, err := vcdClient.CreateRegionStoragePolicy(t)
	if err != nil {
		return diag.Errorf("error creating Region Storage Policy: %s", err)
	}

	d.SetId(region.RegionStoragePolicy.ID)

	return resourceVcdTmRegionStoragePolicyRead(ctx, d, meta)
}

func resourceVcdTmRegionStoragePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rsp, err := vcdClient.GetRegionStoragePolicyById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Region Storage Policy: %s", err)
	}

	t, err := getRegionStoragePolicyType(d)
	if err != nil {
		return diag.Errorf("error getting Region Storage Policy type: %s", err)
	}

	_, err = rsp.Update(t)
	if err != nil {
		return diag.Errorf("error updating Region Storage Policy Type: %s", err)
	}

	return resourceVcdTmRegionStoragePolicyRead(ctx, d, meta)
}

func resourceVcdTmRegionStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdTmRegionStoragePolicyRead(ctx, d, meta, "resource")
}
func genericVcdTmRegionStoragePolicyRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	var rsp *govcd.RegionStoragePolicy
	var err error
	if d.Id() != "" {
		rsp, err = vcdClient.GetRegionStoragePolicyById(d.Id())
	} else {
		rsp, err = vcdClient.GetRegionStoragePolicyByName(d.Get("name").(string))
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving Region Storage Policy: %s", err)
	}

	err = setRegionStoragePolicyData(d, rsp.RegionStoragePolicy)
	if err != nil {
		return diag.Errorf("error saving Region Storage Policy data into state: %s", err)
	}

	d.SetId(rsp.RegionStoragePolicy.ID)
	return nil
}

func resourceVcdTmRegionStoragePolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	region, err := vcdClient.GetRegionById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Region Storage Policy: %s", err)
	}

	err = region.Delete()
	if err != nil {
		return diag.Errorf("error deleting Region Storage Policy: %s", err)
	}

	return nil
}

func resourceVcdTmRegionStoragePolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	rsp, err := vcdClient.GetRegionStoragePolicyByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving Region Storage Policy with name '%s': %s", d.Id(), err)
	}

	d.SetId(rsp.RegionStoragePolicy.ID)
	dSet(d, "name", rsp.RegionStoragePolicy.Name)
	return []*schema.ResourceData{d}, nil
}

func getRegionStoragePolicyType(d *schema.ResourceData) (*types.RegionStoragePolicy, error) {
	t := &types.RegionStoragePolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	return t, nil
}

func setRegionStoragePolicyData(d *schema.ResourceData, rsp *types.RegionStoragePolicy) error {
	dSet(d, "name", rsp.Name)
	dSet(d, "description", rsp.Description)
	regionId := ""
	if rsp.Region != nil {
		regionId = rsp.Region.ID
	}
	dSet(d, "region_id", regionId)
	dSet(d, "storage_capacity_mb", rsp.StorageCapacityMB)
	dSet(d, "storage_consumed_mb", rsp.StorageConsumedMB)
	dSet(d, "status", rsp.Status)

	return nil
}
