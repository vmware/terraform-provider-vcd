package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdRdeInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeInterfaceCreate,
		ReadContext:   resourceVcdRdeInterfaceRead,
		UpdateContext: resourceVcdRdeInterfaceUpdate,
		DeleteContext: resourceVcdRdeInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeInterfaceImport,
		},
		Schema: map[string]*schema.Schema{
			"nss": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // Can't update nss
				ValidateFunc: validateAlphanumericWithUnderscoresAndHyphens(),
				Description:  "A unique namespace associated with the Runtime Defined Entity Interface. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't update version
				Description: "The Runtime Defined Entity Interface's version. The version must follow semantic versioning rules. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"vendor": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // Can't update vendor
				ValidateFunc: validateAlphanumericWithUnderscoresAndHyphens(),
				Description:  "The vendor name. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity Interface",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Runtime Defined Entity Interface cannot be modified",
			},
			"behavior": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Each block defines a Behavior on the Runtime Defined Entity Interface",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the Behavior",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the Behavior",
						},
						"execution": {
							Type:        schema.TypeMap,
							Required:    true,
							Description: "Execution map of the Behavior",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The RDE Interface Behavior ID",
						},
						"ref": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Behavior invocation reference to be used for polymorphic behavior invocations",
						},
					},
				},
			},
		},
	}
}

func resourceVcdRdeInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vendor := d.Get("vendor").(string)
	nss := d.Get("nss").(string)
	version := d.Get("version").(string)
	name := d.Get("name").(string)

	di, err := vcdClient.VCDClient.CreateDefinedInterface(&types.DefinedInterface{
		Name:    name,
		Nss:     nss,
		Version: version,
		Vendor:  vendor,
	})
	if err != nil {
		return diag.Errorf("could not create the Runtime Defined Entity Interface with name %s, vendor %s, nss %s and version %s: %s", name, vendor, nss, version, err)
	}

	// Only System Administrators can create Behaviors
	if meta.(*VCDClient).Client.IsSysAdmin {
		behaviorsAttr, isSet := d.GetOk("behavior")
		if isSet {
			for _, behaviorRaw := range behaviorsAttr.(*schema.Set).List() {
				behaviorAttr := behaviorRaw.(map[string]interface{})
				behavior := types.Behavior{
					Description: behaviorAttr["description"].(string),
					Execution:   behaviorAttr["execution"].(map[string]interface{}),
					Name:        behaviorAttr["name"].(string),
				}
				_, err = di.AddBehavior(behavior)
				if err != nil {
					return diag.Errorf("could not add a Behavior to the Runtime Defined Entity Interface with name %s, vendor %s, nss %s and version %s: %s", name, vendor, nss, version, err)
				}
			}
		}
	}

	return resourceVcdRdeInterfaceRead(ctx, d, meta)
}

func resourceVcdRdeInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceRead(ctx, d, meta, "resource")
}

// genericVcdRdeInterfaceRead reads a Defined Interface from VCD and sets the Terraform state accordingly.
// If origin == "datasource", if the referenced Interface doesn't exist, it errors.
// If origin == "resource", if the referenced Interface doesn't exist, it removes it from tfstate and exits normally.
func genericVcdRdeInterfaceRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity Interface no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Behaviors are only available for System Administrators
	if meta.(*VCDClient).Client.IsSysAdmin {
		behaviors, err := di.GetAllBehaviors(nil)
		if err != nil {
			return diag.Errorf("could not retrieve Behaviors for the Runtime Defined Entity Interface %s: %s", di.DefinedInterface.ID, err)
		}
		var behaviorsAttr = make([]map[string]interface{}, len(behaviors))
		for i, behavior := range behaviors {
			behaviorsAttr[i] = map[string]interface{}{
				"id":          behavior.ID,
				"name":        behavior.Name,
				"description": behavior.Description,
				"ref":         behavior.Ref,
				"execution":   behavior.Execution,
			}
		}
		err = d.Set("behavior", behaviorsAttr)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	dSet(d, "vendor", di.DefinedInterface.Vendor)
	dSet(d, "nss", di.DefinedInterface.Nss)
	dSet(d, "version", di.DefinedInterface.Version)
	dSet(d, "name", di.DefinedInterface.Name)
	dSet(d, "readonly", di.DefinedInterface.IsReadOnly)
	d.SetId(di.DefinedInterface.ID)

	return nil
}

// getDefinedInterface retrieves a Defined Interface from VCD with the required attributes from the Terraform config.
func getDefinedInterface(d *schema.ResourceData, meta interface{}) (*govcd.DefinedInterface, error) {
	vcdClient := meta.(*VCDClient)

	if d.Id() != "" {
		return vcdClient.VCDClient.GetDefinedInterfaceById(d.Id())
	}

	vendor := d.Get("vendor").(string)
	nss := d.Get("nss").(string)
	version := d.Get("version").(string)

	return vcdClient.VCDClient.GetDefinedInterface(vendor, nss, version)
}

func resourceVcdRdeInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("name") {
		err = di.Update(types.DefinedInterface{
			Name: d.Get("name").(string), // Only name can be updated
		})
		if err != nil {
			return diag.Errorf("could not update the Runtime Defined Entity Interface: %s", err)
		}
	}

	// Only System Administrators can update Behaviors
	if meta.(*VCDClient).Client.IsSysAdmin && d.HasChange("behavior") {
		behaviorsRaw := d.Get("behavior").(*schema.Set).List()
		var behaviors = make([]*types.Behavior, len(behaviorsRaw))
		for i, behaviorRaw := range behaviorsRaw {
			b := behaviorRaw.(map[string]interface{})
			behaviors[i] = &types.Behavior{
				ID:          b["id"].(string),
				Description: b["description"].(string),
				Execution:   b["execution"].(map[string]interface{}),
				Ref:         b["ref"].(string),
				Name:        b["name"].(string),
			}
		}
		_, err := di.UpdateBehaviors(behaviors)
		if err != nil {
			return diag.Errorf("could not update Behaviors from Runtime Defined Entity Interface '%s': %s", di.DefinedInterface.ID, err)
		}
	}

	return resourceVcdRdeInterfaceRead(ctx, d, meta)
}

func resourceVcdRdeInterfaceDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	// Note: As Behaviors belong to an Interface, we don't need to remove them: they are deleted once the Interface
	// is gone.
	err = di.Delete()
	if err != nil {
		return diag.Errorf("could not delete the Runtime Defined Entity Interface: %s", err)
	}
	return nil
}

// resourceVcdRdeInterfaceImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde_interface.outer-interface
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeInterfaceImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) < 3 {
		return nil, fmt.Errorf("resource identifier must be specified as vendor.nss.version")
	}
	vendor, nss, version := resourceURI[0], resourceURI[1], strings.Join(resourceURI[2:], ".")

	vcdClient := meta.(*VCDClient)
	di, err := vcdClient.GetDefinedInterface(vendor, nss, version)
	if err != nil {
		return nil, fmt.Errorf("error finding Runtime Defined Entity Interface with vendor %s, nss %s and version %s: %s", vendor, nss, version, err)
	}

	d.SetId(di.DefinedInterface.ID)
	return []*schema.ResourceData{d}, nil
}
