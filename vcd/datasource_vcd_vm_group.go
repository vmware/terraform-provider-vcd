package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

// datasourceVcdVmGroup defines the data source for a VM Group, used to create VM Placement Policies.
func datasourceVcdVmGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVmGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VM Group to fetch",
			},
			"provider_vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Provider VDC to which the VM Group to fetch belongs",
			},
			"cluster_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the vSphere cluster associated to this VM Group",
			},
			"named_vm_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the named VM Group. Used to create Logical VM Groups",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the vCenter server",
			},
			"cluster_moref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Managed object reference of the vSphere cluster associated to this VM Group",
			},
		},
	}
}

func datasourceVcdVmGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	name := d.Get("name").(string)
	providerVdcId := d.Get("provider_vdc_id").(string)

	vmGroup, err := vcdClient.GetVmGroupByNameAndProviderVdcUrn(name, providerVdcId)
	if err != nil {
		log.Printf("[DEBUG] Could not find any VM Group with name %s and pVDC %s: %s", name, providerVdcId, err)
		return diag.Errorf("could not find any VM Group with name %s and pVDC %s: %s", name, providerVdcId, err)
	}

	dSet(d, "cluster_name", vmGroup.VmGroup.ClusterName)
	dSet(d, "named_vm_group_id", vmGroup.VmGroup.NamedVmGroupId)
	dSet(d, "vcenter_id", vmGroup.VmGroup.VcenterId)
	dSet(d, "cluster_moref", vmGroup.VmGroup.ClusterMoref)

	d.SetId(vmGroup.VmGroup.ID)
	return nil
}
