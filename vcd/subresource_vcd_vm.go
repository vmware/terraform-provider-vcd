package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func VirtualMachineSubresourceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"catalog_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"template_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"memory": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"cpus": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"network": {
			Type:     schema.TypeList,
			Required: true,

			Elem: &schema.Resource{
				Schema: VirtualMachineNetworkSubresourceSchema(),
			},
		},
		"initscript": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"power_on": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"nested_hypervisor_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
	return s
}

type VirtualMachineSubresource struct {
	*Subresource
}

func NewVirtualMachineSubresource(client *VCDClient, d, old map[string]interface{}, idx int) *VirtualMachineSubresource {
	sr := &VirtualMachineSubresource{
		Subresource: &Subresource{
			schema:  VirtualMachineSubresourceSchema(),
			data:    d,
			olddata: old,
			// rdd:     rdd,
		},
	}
	sr.Index = idx
	return sr
}
