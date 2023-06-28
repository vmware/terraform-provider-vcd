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

func resourceVcdRdeInterfaceBehavior() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeInterfaceBehaviorCreate,
		ReadContext:   resourceVcdRdeInterfaceBehaviorRead,
		UpdateContext: resourceVcdRdeInterfaceBehaviorUpdate,
		DeleteContext: resourceVcdRdeInterfaceBehaviorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeInterfaceBehaviorImport,
		},
		Schema: map[string]*schema.Schema{
			"interface_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the RDE Interface that owns the Behavior",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Behavior",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description specifying the contract of the Behavior",
			},
			"execution": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Execution map of the Behavior",
			},
			"ref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Behavior invocation reference to be used for polymorphic behavior invocations",
			},
		},
	}
}

func resourceVcdRdeInterfaceBehaviorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	interfaceId := d.Get("interface_id").(string)
	rdeInterface, err := vcdClient.VCDClient.GetDefinedInterfaceById(interfaceId)
	if err != nil {
		return diag.Errorf("[RDE Behavior create] could not read the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}
	_, err = rdeInterface.AddBehavior(types.Behavior{
		Description: d.Get("description").(string),
		Execution:   d.Get("execution").(map[string]interface{}),
		Name:        d.Get("name").(string),
	})
	if err != nil {
		return diag.Errorf("[RDE Behavior create] could not create the Behavior in the RDE Interface with ID '%s': %s", interfaceId, err)
	}
	return genericVcdRdeInterfaceBehaviorRead(ctx, d, meta, "resource")
}

func resourceVcdRdeInterfaceBehaviorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceBehaviorRead(ctx, d, meta, "resource")
}

func genericVcdRdeInterfaceBehaviorRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	interfaceId := d.Get("interface_id").(string)
	rdeInterface, err := vcdClient.VCDClient.GetDefinedInterfaceById(interfaceId)
	if err != nil {
		return diag.Errorf("[RDE Behavior read] could not read the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}

	var behavior *types.Behavior
	if d.Id() != "" {
		behavior, err = rdeInterface.GetBehaviorById(d.Id())
	} else {
		behavior, err = rdeInterface.GetBehaviorByName(d.Get("name").(string))
	}
	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Behavior no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[RDE Behavior read] could not read the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}

	dSet(d, "name", behavior.Name)
	dSet(d, "ref", behavior.Ref)
	dSet(d, "description", behavior.Description)
	err = d.Set("execution", behavior.Execution)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(behavior.ID)

	return nil
}

func resourceVcdRdeInterfaceBehaviorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	interfaceId := d.Get("interface_id").(string)
	rdeInterface, err := vcdClient.VCDClient.GetDefinedInterfaceById(interfaceId)
	if err != nil {
		return diag.Errorf("[RDE Behavior update] could not read the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}
	_, err = rdeInterface.UpdateBehavior(types.Behavior{
		ID:          d.Id(),
		Description: d.Get("description").(string),
		Execution:   d.Get("execution").(map[string]interface{}),
	})
	if err != nil {
		return diag.Errorf("[RDE Behavior update] could not update the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}
	return genericVcdRdeInterfaceBehaviorRead(ctx, d, meta, "resource")
}

func resourceVcdRdeInterfaceBehaviorDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	interfaceId := d.Get("interface_id").(string)
	rdeInterface, err := vcdClient.VCDClient.GetDefinedInterfaceById(interfaceId)
	if err != nil {
		return diag.Errorf("[RDE Behavior delete] could not read the Behavior of RDE Interface with ID '%s': %s", interfaceId, err)
	}
	err = rdeInterface.DeleteBehavior(d.Id())
	if err != nil {
		return diag.Errorf("[RDE Behavior delete] could not delete the Behavior '%s' of RDE Interface with ID '%s': %s", d.Id(), interfaceId, err)
	}
	return nil
}

// resourceVcdRdeInterfaceBehaviorImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde_interface_behavior.behavior1
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0.myBehavior
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeInterfaceBehaviorImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	log.Printf("[DEBUG] importing vcd_rde_interface_behavior resource with provided id %s", d.Id())
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	var rdeInterface *govcd.DefinedInterface
	var behaviorName string
	var err error
	switch len(resourceURI) {
	case 4: // ie: VCD_IMPORT_SEPARATOR="_" vendor_nss_1.2.3_name
		rdeInterface, err = vcdClient.GetDefinedInterface(resourceURI[0], resourceURI[1], resourceURI[2])
		behaviorName = resourceURI[3]
	case 6: // ie: vendor.nss.1.2.3.name
		rdeInterface, err = vcdClient.GetDefinedInterface(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]))
		behaviorName = resourceURI[5]
	default:
		return nil, fmt.Errorf("the import ID should be specified like 'interfaceVendor.interfaceNss.interfaceVersion.behaviorName'")
	}
	if err != nil {
		return nil, fmt.Errorf("could not find any RDE Interface with the provided ID '%s': %s", d.Id(), err)
	}

	behavior, err := rdeInterface.GetBehaviorByName(behaviorName)
	if err != nil {
		return nil, fmt.Errorf("could not find any Behavior with the name '%s' from the given ID '%s': %s", behaviorName, d.Id(), err)
	}

	d.SetId(behavior.ID)
	dSet(d, "interface_id", rdeInterface.DefinedInterface.ID)
	dSet(d, "name", behavior.Name)
	return []*schema.ResourceData{d}, nil
}
