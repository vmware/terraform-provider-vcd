package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"regexp"
	"strings"
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
					ok, err := regexp.MatchString(`^.+\.[z|Z][i|I][p|P]$`, value.(string))
					if err != nil {
						return diag.Errorf("could not validate %s", value.(string))
					}
					if !ok {
						return diag.Errorf("the UI Plugin should be a ZIP bundle, but it is %s", value.(string))
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
				Type:        schema.TypeBool,
				Required:    true,
				Description: "When `true`, publishes the UI Plugin to all tenants",
			},
			"published_tenant_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true,
				Description: "Set of organization IDs to which this UI Plugin must be published",
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
	publishToAllTenants := d.Get("publish_to_all_tenants").(bool)
	publishedOrgIds, isPublishedOrgIdsSet := d.GetOk("published_tenant_ids")

	if publishToAllTenants && isPublishedOrgIdsSet {
		return fmt.Errorf("`publish_to_all_tenants` can't be true if `published_tenant_ids` is also set")
	}

	if d.HasChange("publish_to_all_tenants") {
		if d.Get("publish_to_all_tenants").(bool) {
			err := uiPlugin.PublishAll()
			if err != nil {
				return fmt.Errorf("could not publish the UI Plugin %s to all tenants: %s", uiPlugin.UIPluginMetadata.ID, err)
			}
		} else {
			err := uiPlugin.UnpublishAll()
			if err != nil {
				return fmt.Errorf("could not unpublish the UI Plugin %s from all tenants: %s", uiPlugin.UIPluginMetadata.ID, err)
			}
		}
		return nil
	}

	if d.HasChange("published_tenant_ids") {
		orgIds := publishedOrgIds.(*schema.Set).List()
		if len(orgIds) == 0 {
			return nil
		}
		orgList, err := vcdClient.GetOrgList()
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to tenants '%v': %s", uiPlugin.UIPluginMetadata.ID, orgIds, err)
		}
		var orgRefs types.OpenApiReferences
		for _, org := range orgList.Org {
			for _, orgId := range orgIds {
				// We do this as org.ID is empty, so we need to re-build the URN with the HREF
				uuid := extractUuid(orgId.(string))
				if strings.Contains(org.HREF, uuid) {
					orgRefs = append(orgRefs, types.OpenApiReference{ID: "urn:cloud:org:" + uuid, Name: org.Name})
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
	if len(orgRefs) > 0 {
		var orgIds = make([]string, len(orgRefs))
		for i, orgRef := range orgRefs {
			orgIds[i] = orgRef.ID
		}
		err = d.Set("published_tenant_ids", convertStringsToTypeSet(orgIds))
		if err != nil {
			return diag.FromErr(err)
		}
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

	err = uiPlugin.Delete()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
