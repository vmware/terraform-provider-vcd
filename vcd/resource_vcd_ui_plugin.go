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
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdUIPluginImport,
		},
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
			"tenant_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Set of organization IDs to which this UI Plugin must be published",
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the UI Plugin",
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
	if d.HasChange("tenant_ids") {
		orgsToPublish := d.Get("tenant_ids").(*schema.Set).List()
		existingOrgs, err := vcdClient.GetOrgList()
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to Organizations '%v': %s", uiPlugin.UIPluginMetadata.ID, orgsToPublish, err)
		}
		var orgsToPubRefs types.OpenApiReferences
		for _, org := range existingOrgs.Org {
			for _, orgId := range orgsToPublish {
				// We do this as org.ID is empty, so we need to reconstruct the URN with the HREF
				uuid := extractUuid(orgId.(string))
				if strings.Contains(org.HREF, uuid) {
					orgsToPubRefs = append(orgsToPubRefs, types.OpenApiReference{ID: "urn:cloud:org:" + uuid, Name: org.Name})
				}
			}
		}
		if operation == "update" {
			err = uiPlugin.UnpublishAll() // We need to clean up the already-published Orgs to put the new ones during an Update.
			if err != nil {
				return fmt.Errorf("could not publish the UI Plugin %s to Organizations '%v': %s", uiPlugin.UIPluginMetadata.ID, orgsToPublish, err)
			}
		}
		err = uiPlugin.Publish(orgsToPubRefs)
		if err != nil {
			return fmt.Errorf("could not publish the UI Plugin %s to Organizations '%v': %s", uiPlugin.UIPluginMetadata.ID, orgsToPublish, err)
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
	dSet(d, "status", uiPlugin.UIPluginMetadata.PluginStatus)
	err = setUIPluginTenantIds(uiPlugin, d, origin)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uiPlugin.UIPluginMetadata.ID)
	return nil
}

// setUIPluginTenantIds reads the published tenants for a given UI Plugin.
func setUIPluginTenantIds(uiPlugin *govcd.UIPlugin, d *schema.ResourceData, origin string) error {
	// tenant_ids is only Computed in a data source or during an import
	if origin != "datasource" && origin != "import" {
		return nil
	}
	orgRefs, err := uiPlugin.GetPublishedTenants()
	if err != nil {
		return fmt.Errorf("could not update the published Organizations of the UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	var orgIds = make([]string, len(orgRefs))
	for i, orgRef := range orgRefs {
		orgIds[i] = orgRef.ID
	}
	return d.Set("tenant_ids", convertStringsToTypeSet(orgIds))
}

func resourceVcdUIPluginUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	uiPlugin, err := getUIPlugin(vcdClient, d, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	if uiPlugin == nil {
		return nil
	}
	if d.HasChange("enabled") || d.HasChange("provider_scoped") || d.HasChange("tenant_scoped") {
		err = uiPlugin.Update(d.Get("enabled").(bool), d.Get("provider_scoped").(bool), d.Get("tenant_scoped").(bool))
		if err != nil {
			return diag.Errorf("could not update the UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
		}
	}

	err = publishUIPluginToTenants(vcdClient, uiPlugin, d, "update")
	if err != nil {
		return diag.Errorf("could not update the published Organizations of the UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	return resourceVcdUIPluginRead(ctx, d, meta)
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

// resourceVcdUIPluginImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_ui_plugin.existing_ui_plugin
// Example import path (_the_id_string_): VMware."Customize Portal".3.1.4
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdUIPluginImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) < 3 {
		return nil, fmt.Errorf("resource identifier must be specified as vendor.pluginName.version")
	}
	vendor, name, version := resourceURI[0], resourceURI[1], strings.Join(resourceURI[2:], ".")

	vcdClient := meta.(*VCDClient)
	uiPlugin, err := vcdClient.GetUIPlugin(vendor, name, version)
	if err != nil {
		return nil, fmt.Errorf("error finding UI Plugin with vendor %s, nss %s and version %s: %s", vendor, name, version, err)
	}

	err = setUIPluginTenantIds(uiPlugin, d, "import")
	if err != nil {
		return nil, err
	}
	d.SetId(uiPlugin.UIPluginMetadata.ID)
	return []*schema.ResourceData{d}, nil
}
