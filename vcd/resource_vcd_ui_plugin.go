package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdUIPlugin() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdUIPluginCreate,
		ReadContext:   resourceVcdUIPluginRead,
		UpdateContext: resourceVcdUIPluginUpdate,
		DeleteContext: resourceVcdUIPluginDelete,
		Schema: map[string]*schema.Schema{
			"plugin_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateDiagFunc: func(value interface{}, _ cty.Path) diag.Diagnostics {
					valueString, ok := value.(string)
					if !ok {
						return diag.Errorf("expected type of %v to be string", value)
					}
					if !strings.HasSuffix(valueString, "zip") && !strings.HasSuffix(valueString, "ZIP") {
						return diag.Errorf("the UI Plugin should be a ZIP bundle, but it is %s", valueString)
					}
					return nil
				},
				Description: "Absolute or relative path to the ZIP file containing the UI Plugin",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "true to make the UI Plugin enabled. 'false' to make it disabled",
			},
			"provider_scoped": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "This value is calculated automatically on create by reading the UI Plugin ZIP file contents. You can update" +
					"it to `true` to make it provider scoped or `false` otherwise",
			},
			"tenant_scoped": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "This value is calculated automatically on create by reading the UI Plugin ZIP file contents. You can update" +
					"it to `true` to make it tenant scoped or `false` otherwise",
			},
			"publish_to_all_tenants": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"published_tenant_ids"},
				Description:   "When `true`, publishes the UI Plugin to all tenants",
			},
			"published_tenant_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"publish_to_all_tenants"},
				Description:   "Set of organization IDs to which this UI Plugin must be published",
			},
			"vendor": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UI Plugin vendor name",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UI Plugin name",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the UI Plugin",
			},
			"license": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The license of the UI Plugin",
			},
			"link": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The website of the UI Plugin",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the UI Plugin",
			},
		},
	}
}

func resourceVcdUIPluginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	uiPlugin, err := vcdClient.AddUIPlugin(d.Get("plugin_path").(string), d.Get("enabled").(bool))
	if err != nil {
		return diag.Errorf("could not create the UI Plugin: %s", err)
	}

	err = publishUIPluginToTenants(vcdClient, uiPlugin, d, "create")
	if err != nil {
		return diag.FromErr(err)
	}

	// We set the ID early so the read function can locate the plugin in VCD, as there's no identifying argument on Create.
	// All identifying elements such as vendor, plugin name and version are inside the uploaded ZIP file and populated
	// in Terraform state after a Read.
	d.SetId(uiPlugin.UIPluginMetadata.ID)
	return resourceVcdUIPluginRead(ctx, d, meta)
}

// publishUIPluginToTenants performs a publish/unpublish operation for the given UI plugin.
func publishUIPluginToTenants(vcdClient *VCDClient, uiPlugin *govcd.UIPlugin, d *schema.ResourceData, operation string) error {
	if d.HasChange("publish_to_all_tenants") && d.Get("publish_to_all_tenants").(bool) {
		err := uiPlugin.PublishAll()
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to all tenants: %s", uiPlugin.UIPluginMetadata.ID, err)
		}
		return nil // This return is needed despite this field conflicts with `published_tenant_ids`, as the latter is also computed.
	}

	orgIdsRaw, isSet := d.GetOk("published_tenant_ids")
	if isSet {
		orgIds := orgIdsRaw.(*schema.Set).List()
		orgList, err := vcdClient.GetOrgList()
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to tenants '%v': %s", uiPlugin.UIPluginMetadata.ID, orgIds, err)
		}
		var orgRefs types.OpenApiReferences
		for _, org := range orgList.Org {
			for _, orgId := range orgIds {
				if orgId == org.ID {
					orgRefs = append(orgRefs, types.OpenApiReference{ID: org.ID, Name: org.Name})
				}
			}
		}
		if operation == "update" {
			err = uiPlugin.UnpublishAll() // We need to clean up the already-published Orgs to put the new ones during an Update.
			if err != nil {
				return fmt.Errorf("could not publish the UI Plugin %s to tenants '%v': %s", uiPlugin.UIPluginMetadata.ID, orgIds, err)
			}
		}
		err = uiPlugin.Publish(orgRefs)
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to tenants '%v': %s", uiPlugin.UIPluginMetadata.ID, orgIds, err)
		}
	}
	return nil
}

func resourceVcdUIPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdUIPluginRead(ctx, d, meta, "resource")
}

// getUIPlugin retrieves the UI Plugin from VCD using the resource/data source information.
// Returns a nil govcd.UIPlugin if it doesn't exist in VCD and origin is a "resource".
func getUIPlugin(vcdClient *VCDClient, d *schema.ResourceData, origin string) (*govcd.UIPlugin, error) {
	var uiPlugin *govcd.UIPlugin
	var err error
	if d.Id() != "" {
		uiPlugin, err = vcdClient.GetUIPluginById(d.Id())
	} else {
		uiPlugin, err = vcdClient.GetUIPlugin(d.Get("vendor").(string), d.Get("name").(string), d.Get("version").(string))
	}

	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] UI Plugin no longer exists. Removing from tfstate")
		d.SetId("")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return uiPlugin, nil
}

func genericVcdUIPluginRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	uiPlugin, err := getUIPlugin(vcdClient, d, origin)
	if err != nil {
		return diag.FromErr(err)
	}
	if uiPlugin == nil {
		return nil
	}

	dSet(d, "name", uiPlugin.UIPluginMetadata.PluginName)
	dSet(d, "vendor", uiPlugin.UIPluginMetadata.Vendor)
	dSet(d, "version", uiPlugin.UIPluginMetadata.Version)
	dSet(d, "license", uiPlugin.UIPluginMetadata.License)
	dSet(d, "link", uiPlugin.UIPluginMetadata.Link)
	dSet(d, "tenant_scoped", uiPlugin.UIPluginMetadata.TenantScoped)
	dSet(d, "provider_scoped", uiPlugin.UIPluginMetadata.ProviderScoped)
	dSet(d, "enabled", uiPlugin.UIPluginMetadata.Enabled)
	dSet(d, "description", uiPlugin.UIPluginMetadata.Description)

	orgRefs, err := uiPlugin.GetPublishedTenants()
	if err != nil {
		return diag.Errorf("error retrieving the organizations where the plugin with ID '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	var orgIds = make([]string, len(orgRefs))
	for i, orgRef := range orgRefs {
		orgIds[i] = orgRef.ID
	}
	err = d.Set("published_tenant_ids", orgIds)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uiPlugin.UIPluginMetadata.ID)
	return nil
}

func resourceVcdUIPluginUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	uiPlugin, err := getUIPlugin(vcdClient, d, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	if uiPlugin == nil {
		return nil
	}

	err = uiPlugin.Update(d.Get("enabled").(bool), d.Get("provider_scoped").(bool), d.Get("tenant_scoped").(bool))
	if err != nil {
		return diag.Errorf("could not update the UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	err = publishUIPluginToTenants(vcdClient, uiPlugin, d, "update")
	if err != nil {
		return diag.Errorf("could not update the published tenants of the UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	return nil
}

func resourceVcdUIPluginDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	uiPlugin, err := getUIPlugin(vcdClient, d, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	if uiPlugin == nil {
		return nil
	}

	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] UI Plugin no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = uiPlugin.Delete()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
