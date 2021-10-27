package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdGlobalRole() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceGlobalRoleRead,
		CreateContext: resourceGlobalRoleCreate,
		UpdateContext: resourceGlobalRoleUpdate,
		DeleteContext: resourceGlobalRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdGlobalRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of global role.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Global role description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this global role is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "list of rights assigned to this global role",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_tenants": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Description: "When true, publishes the global role to all tenants",
			},
			"tenants": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "list of tenants to which this global role is published ",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceGlobalRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	globalRoleName := d.Get("name").(string)
	publishToAllTenants := d.Get("publish_to_all_tenants").(bool)

	inputRights, err := getRights(vcdClient, nil, "global role create", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}
	globalRole, err := vcdClient.Client.CreateGlobalRole(&types.GlobalRole{
		Name:        globalRoleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
		PublishAll:  takeBoolPointer(publishToAllTenants),
	})
	if err != nil {
		return diag.Errorf("[global role create] error creating role %s: %s", globalRoleName, err)
	}
	if len(inputRights) > 0 {
		err = globalRole.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[global role create] error adding rights to global role %s: %s", globalRoleName, err)
		}
	}

	inputTenants, err := getTenants(vcdClient, "global role create", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}
	if publishToAllTenants {
		err = globalRole.PublishAllTenants()
		if err != nil {
			return diag.Errorf("[global role create] error publishing to all tenants - global role %s: %s", globalRoleName, err)
		}
	}
	if len(inputTenants) > 0 {
		err = globalRole.PublishTenants(inputTenants)
		if err != nil {
			return diag.Errorf("[global role create] error publishing to tenants - global role %s: %s", globalRoleName, err)
		}
	}
	d.SetId(globalRole.GlobalRole.Id)
	return genericGlobalRoleRead(ctx, d, meta, "resource", "create")
}

func resourceGlobalRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericGlobalRoleRead(ctx, d, meta, "resource", "read")
}

func genericGlobalRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	var globalRole *govcd.GlobalRole
	var err error
	globalRoleName := d.Get("name").(string)
	identifier := d.Id()
	if identifier == "" {
		globalRole, err = vcdClient.Client.GetGlobalRoleByName(globalRoleName)
	} else {
		globalRole, err = vcdClient.Client.GetGlobalRoleById(identifier)
	}

	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[global role read-%s] error retrieving global role %s: %s", operation, globalRoleName, err)
	}

	publishAll := false
	if globalRole.GlobalRole.PublishAll != nil {
		publishAll = *globalRole.GlobalRole.PublishAll
	}
	d.SetId(globalRole.GlobalRole.Id)
	dSet(d, "description", globalRole.GlobalRole.Description)
	dSet(d, "bundle_key", globalRole.GlobalRole.BundleKey)
	dSet(d, "read_only", globalRole.GlobalRole.ReadOnly)
	err = d.Set("publish_to_all_tenants", publishAll)
	if err != nil {
		return diag.Errorf("[global role read-%s] error setting publish_to_all_tenants: %s", operation, err)
	}

	rights, err := globalRole.GetRights(nil)
	if err != nil {
		return diag.Errorf("[global role read-%s] error while querying global role rights: %s", operation, err)
	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[global role read-%s] error setting rights for global role %s: %s", operation, globalRoleName, err)
		}
	}

	tenants, err := globalRole.GetTenants(nil)
	if err != nil {
		return diag.Errorf("[global role read-%s] error while querying global role tenants: %s", operation, err)
	}
	var registeredTenants []interface{}

	for _, tenant := range tenants {
		registeredTenants = append(registeredTenants, tenant.Name)
	}
	if len(registeredTenants) > 0 {
		err = d.Set("tenants", registeredTenants)
		if err != nil {
			return diag.Errorf("[global role read-%s] error setting tenants for global role %s: %s", operation, globalRoleName, err)
		}
	}

	return nil
}

func resourceGlobalRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	globalRoleName := d.Get("name").(string)

	publishToAllTenants := d.Get("publish_to_all_tenants").(bool)

	globalRole, err := vcdClient.Client.GetGlobalRoleById(d.Id())
	if err != nil {
		return diag.Errorf("[global role update] error retrieving global role %s: %s", globalRoleName, err)
	}

	var inputRights []types.OpenApiReference
	var inputTenants []types.OpenApiReference
	var changedRights = d.HasChange("rights")
	var changedTenants = d.HasChange("tenants") || d.HasChange("publish_to_all_tenants")

	if changedRights {
		inputRights, err = getRights(vcdClient, nil, "global role update", d)
		if err != nil {
			return diag.Errorf("%s", err)
		}
	}

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("publish_to_all_tenants") {
		globalRole.GlobalRole.Name = globalRoleName
		globalRole.GlobalRole.Description = d.Get("description").(string)
		globalRole.GlobalRole.PublishAll = takeBoolPointer(publishToAllTenants)
		_, err = globalRole.Update()
		if err != nil {
			return diag.Errorf("[global role update] error updating global role %s: %s", globalRoleName, err)
		}
	}

	if changedRights {
		if len(inputRights) > 0 {
			err = globalRole.UpdateRights(inputRights)
			if err != nil {
				return diag.Errorf("[global role update] error updating global role %s rights: %s", globalRoleName, err)
			}
		} else {
			currentRights, err := globalRole.GetRights(nil)
			if err != nil {
				return diag.Errorf("[global role update] error retrieving global role %s rights: %s", globalRoleName, err)
			}
			if len(currentRights) > 0 {
				err = globalRole.RemoveAllRights()
				if err != nil {
					return diag.Errorf("[global role update] error removing global role %s rights: %s", globalRoleName, err)
				}
			}
		}
	}
	if changedTenants {
		inputTenants, err = getTenants(vcdClient, "global role create", d)
		if err != nil {
			return diag.Errorf("%s", err)
		}
		if publishToAllTenants {
			err = globalRole.PublishAllTenants()
			if err != nil {
				return diag.Errorf("[global role update] error publishing to all tenants - global role %s: %s", globalRoleName, err)
			}
		} else {
			if len(inputTenants) > 0 {
				err = globalRole.ReplacePublishedTenants(inputTenants)
				if err != nil {
					return diag.Errorf("[global role update] error publishing to tenants - global role %s: %s", globalRoleName, err)
				}
			} else {
				if !publishToAllTenants {
					err = globalRole.UnpublishAllTenants()
					if err != nil {
						return diag.Errorf("[global role update] error unpublishing from all tenants - global role %s: %s", globalRoleName, err)
					}
				}
			}
		}
	}

	return genericGlobalRoleRead(ctx, d, meta, "resource", "update")
}

func resourceGlobalRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	globalRoleName := d.Get("name").(string)

	var globalRole *govcd.GlobalRole
	var err error
	identifier := d.Id()
	if identifier == "" {
		globalRole, err = vcdClient.Client.GetGlobalRoleByName(globalRoleName)
	} else {
		globalRole, err = vcdClient.Client.GetGlobalRoleById(identifier)
	}

	if err != nil {
		return diag.Errorf("[global role delete] error retrieving global role %s: %s", globalRoleName, err)
	}

	if err != nil {
		return diag.Errorf("[global role delete] error retrieving global role %s: %s", globalRoleName, err)
	}
	err = globalRole.Delete()
	if err != nil {
		return diag.Errorf("[global role delete] error deleting global role %s: %s", globalRoleName, err)
	}
	return nil
}

// getTenants returns a list of tenants for provider level rights containers (global role, rights bundle)
func getTenants(client *VCDClient, label string, d *schema.ResourceData) ([]types.OpenApiReference, error) {
	var inputTenants []types.OpenApiReference

	tenants := d.Get("tenants").(*schema.Set).List()

	for _, r := range tenants {
		tenantName := r.(string)
		org, err := client.GetAdminOrgByName(tenantName)
		if err != nil {
			return nil, fmt.Errorf("[%s] error retrieving tenant %s: %s", label, tenantName, err)
		}
		inputTenants = append(inputTenants, types.OpenApiReference{Name: tenantName, ID: org.AdminOrg.ID})
	}
	return inputTenants, nil
}

func resourceVcdGlobalRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 1 {
		return nil, fmt.Errorf("resource name must be specified as globalrole-name")
	}
	globalRoleName := resourceURI[0]

	vcdClient := meta.(*VCDClient)

	globalRole, err := vcdClient.Client.GetGlobalRoleByName(globalRoleName)
	if err != nil {
		return nil, fmt.Errorf("[global role import] error retrieving global role %s: %s", globalRoleName, err)
	}
	dSet(d, "name", globalRoleName)
	dSet(d, "description", globalRole.GlobalRole.Description)
	dSet(d, "bundle_key", globalRole.GlobalRole.BundleKey)
	publishAll := false
	if globalRole.GlobalRole.PublishAll != nil {
		publishAll = *globalRole.GlobalRole.PublishAll
	}
	dSet(d, "publish_to_all_tenants", publishAll)
	d.SetId(globalRole.GlobalRole.Id)
	return []*schema.ResourceData{d}, nil
}
