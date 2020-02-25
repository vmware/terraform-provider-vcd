package vcd

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// vcdResourceDataInterface defines an interface for Terraform's 'schema.ResourceData' type.
type vcdResourceDataInterface interface {
	Id() string
	Get(key string) interface{}
	SetId(v string)

	// Unused in testing context, but still useful to have for a better auto-completion
	UnsafeSetFieldRaw(key string, value string)
	GetChange(key string) (interface{}, interface{})
	GetOk(key string) (interface{}, bool)
	GetOkExists(key string) (interface{}, bool)
	HasChanges(keys ...string) bool
	HasChange(key string) bool
	Partial(on bool)
	Set(key string, value interface{}) error
	SetPartial(k string)
	MarkNewResource()
	IsNewResource() bool
	ConnInfo() map[string]string
	SetConnInfo(v map[string]string)
	SetType(t string)
	State() *terraform.InstanceState
	Timeout(key string) time.Duration
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
