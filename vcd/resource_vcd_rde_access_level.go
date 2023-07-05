package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdRdeBehaviorAccessLevel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeBehaviorAccessLevelCreate,
		ReadContext:   resourceVcdRdeBehaviorAccessLevelRead,
		UpdateContext: resourceVcdRdeBehaviorAccessLevelUpdate,
		DeleteContext: resourceVcdRdeBehaviorAccessLevelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeBehaviorAccessLevelImport,
		},
		Schema: map[string]*schema.Schema{
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the RDE Type",
			},
			"behavior_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of either a RDE Interface Behavior or RDE Type Behavior",
			},
			"access_level_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Set of access level IDs to associate to this Behavior",
			},
		},
	}
}

func resourceVcdRdeBehaviorAccessLevelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdRdeBehaviorAccessLevelCreateOrUpdate(ctx, d, meta, "create")
}

func resourceVcdRdeBehaviorAccessLevelCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level %s] could not retrieve the RDE Type with ID '%s': %s", operation, rdeTypeId, err)
	}

	behaviorId := d.Get("behavior_id").(string)
	behavior, err := rdeType.GetBehaviorById(behaviorId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level %s] could not retrieve the Behavior with ID '%s': %s", operation, behaviorId, err)
	}

	definedAcls := d.Get("access_level_ids").(*schema.Set).List()
	var payload []*types.BehaviorAccess

	if operation == "update" {
		// We get "old" ACLs as there can be more ACLs from other Behaviors that would be deleted otherwise.
		allAcls, err := rdeType.GetAllBehaviorsAccessControls(nil)
		if err != nil {
			return diag.Errorf("[RDE Behavior Access Level %s] could not get the Behavior '%s' Access Levels: %s", operation, behavior.ID, err)
		}

		for _, acl := range allAcls {
			for _, definedAcl := range definedAcls {
				// We just ignore the entries that belong to the same Behavior but have been changed (because they need to be updated)
				if acl.BehaviorId == behaviorId && acl.AccessLevelId != definedAcl {
					continue
				}
				payload = append(payload, acl)
			}
		}
	}

	var newAcls = make([]*types.BehaviorAccess, len(definedAcls))
	for i, acl := range definedAcls {
		newAcls[i] = &types.BehaviorAccess{
			AccessLevelId: acl.(string),
			BehaviorId:    behavior.ID,
		}
	}

	payload = append(payload, newAcls...)
	err = rdeType.SetBehaviorAccessControls(payload)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level %s] could not set the Behavior '%s' Access Levels: %s", operation, behavior.ID, err)
	}

	return genericVcdRdeBehaviorAccessLevelRead(ctx, d, meta)
}

func resourceVcdRdeBehaviorAccessLevelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeBehaviorAccessLevelRead(ctx, d, meta)
}

func genericVcdRdeBehaviorAccessLevelRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level read] could not retrieve the RDE Type with ID '%s': %s", rdeTypeId, err)
	}

	behaviorId := d.Get("behavior_id").(string)
	// This is not really needed, but this way we assure the Behavior exists
	behavior, err := rdeType.GetBehaviorById(behaviorId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level read] could not retrieve the Behavior with ID '%s': %s", behaviorId, err)
	}

	acls, err := rdeType.GetAllBehaviorsAccessControls(nil)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level read] could not read the Behavior Access Levels of RDE Type with ID '%s': %s", rdeTypeId, err)
	}
	var aclsAttr []string
	for _, acl := range acls {
		// The RDE Type can have ACLs for more Behaviors, we only fetch the ones from this specific behavior.
		if acl.BehaviorId == behavior.ID {
			aclsAttr = append(aclsAttr, acl.AccessLevelId)
		}
	}
	err = d.Set("access_level_ids", aclsAttr)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(behavior.ID)

	return nil
}

func resourceVcdRdeBehaviorAccessLevelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdRdeBehaviorAccessLevelCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcdRdeBehaviorAccessLevelDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level delete] could not retrieve the RDE Type with ID '%s': %s", rdeTypeId, err)
	}
	err = rdeType.SetBehaviorAccessControls([]*types.BehaviorAccess{})
	if err != nil {
		return diag.Errorf("[RDE Behavior Access Level delete] could not delete the Access Levels of RDE Type with ID '%s': %s", rdeTypeId, err)
	}
	return nil
}

// resourceVcdRdeBehaviorAccessLevelImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde_behavior_acl.behavior_acl1
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0.myBehavior
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeBehaviorAccessLevelImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	log.Printf("[DEBUG] importing vcd_rde_type_behavior resource with provided id %s", d.Id())
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	var rdeType *govcd.DefinedEntityType
	var behaviorName string
	var err error
	switch len(resourceURI) {
	case 4: // ie: VCD_IMPORT_SEPARATOR="_" vendor_nss_1.2.3_name
		rdeType, err = vcdClient.GetRdeType(resourceURI[0], resourceURI[1], resourceURI[2])
		behaviorName = resourceURI[3]
	case 6: // ie: vendor.nss.1.2.3.name
		rdeType, err = vcdClient.GetRdeType(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]))
		behaviorName = resourceURI[5]
	default:
		return nil, fmt.Errorf("the import ID should be specified like 'rdeTypeVendor.rdeTypeNss.rdeTypeVersion.behaviorName'")
	}
	if err != nil {
		return nil, fmt.Errorf("could not find any RDE Type with the provided ID '%s': %s", d.Id(), err)
	}

	behavior, err := rdeType.GetBehaviorByName(behaviorName)
	if err != nil {
		return nil, fmt.Errorf("could not get the Behavior '%s' of the RDE Type with ID '%s': %s", behaviorName, rdeType.DefinedEntityType.ID, err)
	}

	d.SetId(behavior.ID)
	dSet(d, "rde_type_id", rdeType.DefinedEntityType.ID)
	dSet(d, "behavior_id", behavior.ID)
	return []*schema.ResourceData{d}, nil
}
