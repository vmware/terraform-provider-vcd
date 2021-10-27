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

func resourceVcdRightsBundle() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRightsBundleCreate,
		ReadContext:   resourceRightsBundleRead,
		UpdateContext: resourceRightsBundleUpdate,
		DeleteContext: resourceRightsBundleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRightsBundleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of rights bundle.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Rights bundle description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this rights bundle is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "list of rights assigned to this rights bundle",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_tenants": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Description: "When true, publishes the rights bundle to all tenants",
			},
			"tenants": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "list of tenants to which this rights bundle is published",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourceRightsBundleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightsBundleName := d.Get("name").(string)
	publishToAllTenants := d.Get("publish_to_all_tenants").(bool)

	inputRights, err := getRights(vcdClient, nil, "rights bundle create", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}
	rightsBundle, err := vcdClient.Client.CreateRightsBundle(&types.RightsBundle{
		Name:        rightsBundleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
		PublishAll:  takeBoolPointer(publishToAllTenants),
	})
	if err != nil {
		return diag.Errorf("[rights bundle create] error creating role %s: %s", rightsBundleName, err)
	}
	if len(inputRights) > 0 {
		err = rightsBundle.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[rights bundle create] error adding rights to rights bundle %s: %s", rightsBundleName, err)
		}
	}

	inputTenants, err := getTenants(vcdClient, "rights bundle create", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}
	if publishToAllTenants {
		err = rightsBundle.PublishAllTenants()
		if err != nil {
			return diag.Errorf("[rights bundle create] error publishing to all tenants - rights bundle %s: %s", rightsBundleName, err)
		}
	}
	if len(inputTenants) > 0 {
		err = rightsBundle.PublishTenants(inputTenants)
		if err != nil {
			return diag.Errorf("[rights bundle create] error publishing to tenants - rights bundle %s: %s", rightsBundleName, err)
		}
	}
	d.SetId(rightsBundle.RightsBundle.Id)
	return genericRightsBundleRead(ctx, d, meta, "resource", "create")
}

func resourceRightsBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericRightsBundleRead(ctx, d, meta, "resource", "read")
}

func genericRightsBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	var rightsBundle *govcd.RightsBundle
	var err error
	rightsBundleName := d.Get("name").(string)
	identifier := d.Id()

	if identifier == "" {
		rightsBundle, err = vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	} else {
		rightsBundle, err = vcdClient.Client.GetRightsBundleById(identifier)
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[rights bundle read-%s] error retrieving rights bundle %s: %s", operation, rightsBundleName, err)
	}

	d.SetId(rightsBundle.RightsBundle.Id)
	dSet(d, "description", rightsBundle.RightsBundle.Description)
	dSet(d, "bundle_key", rightsBundle.RightsBundle.BundleKey)
	dSet(d, "read_only", rightsBundle.RightsBundle.ReadOnly)

	rights, err := rightsBundle.GetRights(nil)
	if err != nil {
		return diag.Errorf("[rights bundle read-%s] error while querying rights bundle rights: %s", operation, err)
	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[rights bundle read-%s] error setting rights for rights bundle %s: %s", operation, rightsBundleName, err)
		}
	}

	tenants, err := rightsBundle.GetTenants(nil)
	if err != nil {
		return diag.Errorf("[rights bundle read-%s] error while querying rights bundle tenants: %s", operation, err)
	}
	var registeredTenants []interface{}

	publishAll := false
	if rightsBundle.RightsBundle.PublishAll != nil {
		publishAll = *rightsBundle.RightsBundle.PublishAll
	}
	dSet(d, "publish_to_all_tenants", publishAll)
	for _, tenant := range tenants {
		registeredTenants = append(registeredTenants, tenant.Name)
	}
	if !publishAll {
		if len(registeredTenants) > 0 {
			err = d.Set("tenants", registeredTenants)
			if err != nil {
				return diag.Errorf("[rights bundle read-%s] error setting tenants for rights bundle %s: %s", operation, rightsBundleName, err)
			}
		}
	}

	return nil
}

func resourceRightsBundleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightsBundleName := d.Get("name").(string)

	publishToAllTenants := d.Get("publish_to_all_tenants").(bool)

	rightsBundle, err := vcdClient.Client.GetRightsBundleById(d.Id())
	if err != nil {
		return diag.Errorf("[rights bundle update] error retrieving rights bundle %s: %s", rightsBundleName, err)
	}

	var inputRights []types.OpenApiReference
	var inputTenants []types.OpenApiReference
	var changedRights = d.HasChange("rights")
	var changedTenants = d.HasChange("tenants") || d.HasChange("publish_to_all_tenants")
	if changedRights {
		inputRights, err = getRights(vcdClient, nil, "rights bundle update", d)
		if err != nil {
			return diag.Errorf("%s", err)
		}
	}

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("publish_to_all_tenants") {
		rightsBundle.RightsBundle.Name = rightsBundleName
		rightsBundle.RightsBundle.Description = d.Get("description").(string)
		rightsBundle.RightsBundle.PublishAll = takeBoolPointer(publishToAllTenants)
		_, err = rightsBundle.Update()
		if err != nil {
			return diag.Errorf("[rights bundle update] error updating rights bundle %s: %s", rightsBundleName, err)
		}
	}

	if changedRights {
		if len(inputRights) > 0 {
			err = rightsBundle.UpdateRights(inputRights)
			if err != nil {
				return diag.Errorf("[rights bundle update] error updating rights bundle %s rights: %s", rightsBundleName, err)
			}
		} else {
			currentRights, err := rightsBundle.GetRights(nil)
			if err != nil {
				return diag.Errorf("[rights bundle update] error retrieving rights bundle %s rights: %s", rightsBundleName, err)
			}
			if len(currentRights) > 0 {
				err = rightsBundle.RemoveAllRights()
				if err != nil {
					return diag.Errorf("[rights bundle update] error removing rights bundle %s rights: %s", rightsBundleName, err)
				}
			}
		}
	}
	if changedTenants {
		inputTenants, err = getTenants(vcdClient, "rights bundle create", d)
		if err != nil {
			return diag.Errorf("%s", err)
		}
		if publishToAllTenants {
			err = rightsBundle.PublishAllTenants()
			if err != nil {
				return diag.Errorf("[rights bundle update] error publishing to all tenants - rights bundle %s: %s", rightsBundleName, err)
			}
		} else {
			if len(inputTenants) > 0 {
				err = rightsBundle.ReplacePublishedTenants(inputTenants)
				if err != nil {
					return diag.Errorf("[rights bundle update] error publishing to tenants - rights bundle %s: %s", rightsBundleName, err)
				}
			} else {
				err = rightsBundle.UnpublishAllTenants()
				if err != nil {
					return diag.Errorf("[rights bundle update] error unpublishing from all tenants - rights bundle %s: %s", rightsBundleName, err)
				}
			}
		}
	}

	return genericRightsBundleRead(ctx, d, meta, "resource", "update")
}

func resourceRightsBundleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightsBundleName := d.Get("name").(string)

	var rightsBundle *govcd.RightsBundle
	var err error
	identifier := d.Id()
	if identifier == "" {
		rightsBundle, err = vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	} else {
		rightsBundle, err = vcdClient.Client.GetRightsBundleById(identifier)
	}

	if err != nil {
		return diag.Errorf("[rights bundle delete] error retrieving rights bundle %s: %s", rightsBundleName, err)
	}

	err = rightsBundle.Delete()
	if err != nil {
		return diag.Errorf("[rights bundle delete] error deleting rights bundle %s: %s", rightsBundleName, err)
	}
	return nil
}

func resourceVcdRightsBundleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 1 {
		return nil, fmt.Errorf("resource name must be specified as rightsBundle-name")
	}
	rightsBundleName := resourceURI[0]

	vcdClient := meta.(*VCDClient)

	rightsBundle, err := vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		return nil, fmt.Errorf("[rights bundle import] error retrieving rights bundle %s: %s", rightsBundleName, err)
	}
	dSet(d, "name", rightsBundleName)
	dSet(d, "description", rightsBundle.RightsBundle.Description)
	dSet(d, "bundle_key", rightsBundle.RightsBundle.BundleKey)

	publishAll := false
	if rightsBundle.RightsBundle.PublishAll != nil {
		publishAll = *rightsBundle.RightsBundle.PublishAll
	}
	dSet(d, "publish_to_all_tenants", publishAll)
	d.SetId(rightsBundle.RightsBundle.Id)
	return []*schema.ResourceData{d}, nil
}
