package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// This resource is quite special, as it represents an imperative call to a function (aka invoking a behavior)
// in the declarative world of Terraform. It is not a data source to highlight the implications of invoking a behavior, which
// can mutate RDEs. For that reason, we add the attribute "invoke_on_every_refresh", so users can decide whether to invoke the behavior on
// every read operation (plan, refresh, apply), or just when the resource is created.
//
// To avoid unnecessary complexity, all arguments have ForceNew=true, this way we avoid the Update operation and the combination
// Read+Update, which could lead to accidental double invocations.
// Also, there's no Delete as there's nothing to delete.
func datasourceVcdRdeBehaviorInvocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeBehaviorInvocationRead,
		Schema: map[string]*schema.Schema{
			"rde_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the RDE for which the Behavior will be invoked",
			},
			"behavior_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of either a RDE Interface Behavior or RDE Type Behavior to be invoked",
			},
			"arguments": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The arguments to be passed to the invoked Behavior",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Metadata to be passed to the invoked Behavior",
			},
			"result": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The raw result of the Behavior invocation",
			},
		},
	}
}

func datasourceVcdRdeBehaviorInvocationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeId := d.Get("rde_id").(string)
	behaviorId := d.Get("behavior_id").(string)

	rde, err := vcdClient.GetRdeById(rdeId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Invocation] could not retrieve the RDE with ID '%s': %s", rdeId, err)
	}
	result, err := rde.InvokeBehavior(behaviorId, types.BehaviorInvocation{
		Arguments: d.Get("arguments").(map[string]interface{}),
		Metadata:  d.Get("metadata").(map[string]interface{}),
	})
	if err != nil {
		return diag.Errorf("[RDE Behavior Invocation] could not invoke the Behavior of the RDE with ID '%s': %s", rdeId, err)
	}
	d.SetId(rdeId + "|" + behaviorId)
	dSet(d, "result", result)
	return nil
}
