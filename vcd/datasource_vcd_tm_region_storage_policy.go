package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const labelTmRegionStoragePolicy = "Region Storage Policy"

func datasourceVcdTmRegionStoragePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmRegionStoragePolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s name", labelTmRegionStoragePolicy),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The Region that this %s belongs to", labelTmRegionStoragePolicy),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelTmRegionStoragePolicy),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The creation status of the %s. Can be [NOT_READY, READY]", labelTmRegionStoragePolicy),
			},
			"storage_capacity_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Storage capacity in megabytes for this %s", labelTmRegionStoragePolicy),
			},
			"storage_consumed_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Consumed storage in megabytes for this %s", labelTmRegionStoragePolicy),
			},
		},
	}
}

func datasourceVcdTmRegionStoragePolicyRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	regionId := d.Get("region_id").(string)
	region, err := vcdClient.GetRegionById(regionId)
	if err != nil {
		return diag.Errorf("error retrieving Region with ID '%s': %s", regionId, err)
	}

	rspName := d.Get("name").(string)
	rsp, err := region.GetStoragePolicyByName(rspName)
	if err != nil {
		return diag.Errorf("error retrieving Region Storage Policy '%s': %s", rspName, err)
	}

	err = setRegionStoragePolicyData(d, rsp.RegionStoragePolicy)
	if err != nil {
		return diag.Errorf("error saving Region Storage Policy data into state: %s", err)
	}

	d.SetId(rsp.RegionStoragePolicy.ID)
	return nil
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
