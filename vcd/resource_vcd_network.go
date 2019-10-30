package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DEPRECATED: use vcd_network_routed instead
func resourceVcdNetwork() *schema.Resource {
	newRes := resourceVcdNetworkRouted()
	newRes.DeprecationMessage = "Deprecated. Use vcd_network_routed instead"
	return newRes
}
