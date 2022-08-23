package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceVcdVmPlacementPolicy() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceVmPlacementPolicyCreate,
		ReadContext:   resourceVmPlacementPolicyRead,
		UpdateContext: resourceVmPlacementPolicyUpdate,
		DeleteContext: resourceVmPlacementPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVmPlacementPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vm_groups": {
				Type:     schema.TypeSet,
				Required: true,
				Description: "Collection of VMs with similar host requirements",
				ConflictsWith: []string { "logical_vm_groups"},
			},
			"logical_vm_groups": {
				Type:     schema.TypeSet,
				Required: true,
				Description: "One or more Logical VM Groups to create the VM Placement policy. There would be an AND relationship among all the entries specified in this attribute",
				ConflictsWith: []string { "vm_groups"},
			},
		},
	}
}

func resourceVmPlacementPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceVcdVmAffinityRuleRead reads a resource VM affinity rule
func resourceVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmPlacementPolicyRead(ctx, d, meta)
}

// Fetches information about an existing VM sizing policy for a data definition
func genericVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy read initiated: %s", policyName)

	return nil
}

//resourceVmPlacementPolicyUpdate function updates resource with found configurations changes
func resourceVmPlacementPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVmPlacementPolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceVmSizingPolicyImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vm_sizing_policy.my_existing_policy_name
// Example import path (_the_id_string_): my_existing_vm_sizing_policy_id
// Example list path (_the_id_string_): list@
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVmPlacementPolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}