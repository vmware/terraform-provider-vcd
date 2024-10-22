package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

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

	d.SetId(rsp.RegionStoragePolicy.Id)
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
