package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v3/govcd"

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
			"cluster_moref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The managed object reference of the Cluster in which the resource pool exists.",
			},
		},
	}
}

func datasourceResourcePoolRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	resourcePoolName := d.Get("name").(string)
	vCenterId := d.Get("vcenter_id").(string)

	vCenter, err := vcdClient.GetVCenterById(vCenterId)
	if err != nil {
		return diag.FromErr(err)
	}
	var resourcePool *govcd.ResourcePool

	// If the resource pool has a duplicate name within the same vCenter, we can use the ID instead
	resourcePool, err = vCenter.GetResourcePoolByName(resourcePoolName)
	if err != nil {
		firstErr := err
		resourcePool, err = vCenter.GetResourcePoolById(resourcePoolName)
		if err != nil {
			return diag.Errorf("could not find resource pool by name '%s': %s", resourcePoolName, firstErr)
		}
	}

	hardwareVersion, err := resourcePool.GetDefaultHardwareVersion()
	if err != nil {
		return diag.Errorf("error retrieving default hardware version for resource pool %s: %s", resourcePoolName, err)
	}
	dSet(d, "name", resourcePool.ResourcePool.Name)
	dSet(d, "hardware_version", hardwareVersion)
	dSet(d, "cluster_moref", resourcePool.ResourcePool.ClusterMoref)
	d.SetId(resourcePool.ResourcePool.Moref)

	return nil
}
