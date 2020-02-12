package vcd

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

// vcdResourceDataInterface defines a wider interface for Terraform's 'schema.ResourceData' type.
type vcdResourceDataInterface interface {
	Id() string
	Get(key string) interface{}
	SetId(v string)
}

// vcdDeleteFunc type is almost as Terraform's schema.DeleteFunc, but it accepts a locally
// defined interface 'vcdInterfacedFunction' rather than the original 'd *schema.ResourceData'
// This allows us to inject a fake state into functions defined by Terraform to reuse them. For example it is very
// useful to re-use this code for deletion by injecting proper parameters
//
// This function can be converted to the one matching schema.DeleteFunc by using 'toDeleteFunc'
type vcdDeleteFunc func(vcdResourceDataInterface, interface{}) error

// toDeleteFunc takes a `vcdDeleteFunc` (which accepts vcdResourceDataInterface instead of *schema.ResourceData) and
// converts it to schema.DeleteFunc which is required by Terraform's resources.
// This helps to make re-usable DeleteFunc functions for both testing and original Terraform resource deletions
func toDeleteFunc(vcdInterfacedFunction vcdDeleteFunc) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		return vcdInterfacedFunction(d, meta)
	}
}
