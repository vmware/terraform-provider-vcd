package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdResourcePool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceResourcePoolRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T manager.",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the vCenter containing the resource pool",
			},
			"hardware_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default hardware version for this resource pool",
			},
		},
	}
}

func datasourceResourcePoolRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	resourcePoolName := d.Get("name").(string)
	vCenterId := d.Get("vcenter_id").(string)

	vCenter, err := vcdClient.GetVcenterById(vCenterId)
	if err != nil {
		return diag.FromErr(err)
	}
	resourcePool, err := vCenter.GetResourcePoolByName(resourcePoolName)
	if err != nil {
		return diag.Errorf("could not find  resource pool by name '%s': %s", resourcePoolName, err)
	}

	hardwareVersion, err := resourcePool.GetDefaultHardwareVersion()
	if err != nil {
		return diag.Errorf("error retrieving default hardware version for resource pool %s: %s", resourcePoolName, err)
	}
	dSet(d, "name", resourcePool.ResourcePool.Name)
	dSet(d, "hardware_version", hardwareVersion)
	d.SetId(resourcePool.ResourcePool.Moref)

	return nil
}
