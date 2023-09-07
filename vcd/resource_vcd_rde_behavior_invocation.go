package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// This resource is quite special, as it represents an imperative call to a function (aka invoking a behavior)
// in the declarative world of Terraform.
// For that reason, we add the attribute "invoke_on_every_refresh", so users can decide whether to invoke the behavior on
// every read operation (plan, refresh, apply), or just when the resource is created.
//
// To avoid unnecessary complexity, Update does nothing (invocation happens on Read or Create), this way we avoid invoking
// Behaviors twice in a row by mistake (one during Read if invoke_on_every_refresh=true and another in Update). Delete also does
// nothing as there's nothing to delete.
func resourceVcdRdeBehaviorInvocation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeBehaviorInvocationCreate,
		ReadContext:   resourceVcdRdeBehaviorInvocationRead,
		UpdateContext: resourceVcdRdeBehaviorInvocationUpdate,
		DeleteContext: resourceVcdRdeBehaviorInvocationDelete,
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
			"invoke_on_every_refresh": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "The raw result of the Behavior invocation",
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

func resourceVcdRdeBehaviorInvocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdRdeBehaviorInvocationExecute(ctx, d, meta, "create")
}

func resourceVcdRdeBehaviorInvocationExecute(_ context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeId := d.Get("rde_id").(string)
	behaviorId := d.Get("behavior_id").(string)

	rde, err := vcdClient.GetRdeById(rdeId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Invocation %s] could not retrieve the RDE with ID '%s': %s", operation, rdeId, err)
	}
	result, err := rde.InvokeBehavior(behaviorId, types.BehaviorInvocation{
		Arguments: d.Get("arguments").(map[string]interface{}),
		Metadata:  d.Get("metadata").(map[string]interface{}),
	})
	if err != nil {
		return diag.Errorf("[RDE Behavior Invocation %s] could not invoke the Behavior of the RDE with ID '%s': %s", operation, rdeId, err)
	}
	d.SetId(rdeId + "|" + behaviorId)
	dSet(d, "result", result)
	return nil
}

func resourceVcdRdeBehaviorInvocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// If the operation is Read and user doesn't want to invoke the behavior on each Read, we return early.
	invokeOnRefresh := d.Get("invoke_on_every_refresh").(bool)
	if !invokeOnRefresh {
		return nil
	}
	return resourceVcdRdeBehaviorInvocationExecute(ctx, d, meta, "read")
}

func resourceVcdRdeBehaviorInvocationUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// There's nothing to do here.
	return nil
}

func resourceVcdRdeBehaviorInvocationDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// There's nothing to do here.
	return nil
}
