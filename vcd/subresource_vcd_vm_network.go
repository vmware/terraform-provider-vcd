package vcd

import (
	"net"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func VirtualMachineNetworkSubresourceSchema() map[string]*schema.Schema {

	s := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ip": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: suppressIPDifferences,
			ValidateFunc:     ValidateIPv4(),
		},
		"ip_allocation_mode": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				types.IPAllocationModeDHCP,
				types.IPAllocationModeManual,
				// types.IPAllocationModeNone, This is not accepted by vCloud for some reason
				types.IPAllocationModePool,
			}, false),
		},
		"is_primary": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"is_connected": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"adapter_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "VMXNET3",
			ValidateFunc: validation.StringInSlice([]string{
				"VMXNET3",
				"E1000",
				"E1000E",
			}, false),
		},
	}
	return s
}

type VirtualMachineNetworkSubresource struct {
	*Subresource
}

func NewVirtualMachineNetworkSubresource(d, old map[string]interface{}) *VirtualMachineNetworkSubresource {
	sr := &VirtualMachineNetworkSubresource{
		Subresource: &Subresource{
			schema:  VirtualMachineNetworkSubresourceSchema(),
			data:    d,
			olddata: old,
			// rdd:     rdd,
		},
	}
	return sr
}

// Suppress Diff on equal ip
func suppressIPDifferences(k, old, new string, d *schema.ResourceData) bool {
	o := net.ParseIP(old)
	n := net.ParseIP(new)

	if o != nil && n != nil {
		return o.Equal(n)
	}
	return false
}
