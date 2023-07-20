package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func datasourceVcdSubscribedCatalog() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSubscribedCatalogRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional storage profile ID",
			},
			"subscription_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to which this catalog is subscribed.",
			},
			"make_local_copy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If true, subscription to a catalog creates a local copy of all items. If false, it does not create a local copy of catalogItems unless sync operation is performed.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog HREF",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},
			"catalog_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Catalog version number. Inherited from the publishing catalog.",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name from the catalog.",
			},
			"number_of_vapp_templates": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of vApps templates this catalog contains.",
			},
			"number_of_media": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of Media items this catalog contains.",
			},
			"vapp_template_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of catalog items in this catalog",
			},
			"media_item_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Media items in this catalog",
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is shared.",
			},
			"is_local": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog belongs to the current organization.",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is published.",
			},
			"publish_subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PUBLISHED if published externally, SUBSCRIBED if subscribed to an external catalog, UNPUBLISHED otherwise. (Always SUBSCRIBED for this data source)",
			},
			"running_tasks": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of running synchronization tasks",
			},
			"failed_tasks": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of failed synchronization tasks",
			},
			"filter": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a catalog by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"date":       elementDate,
						"earliest":   elementEarliest,
						"latest":     elementLatest,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func datasourceVcdSubscribedCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Subscribed Catalog data source read initiated\n")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("name").(string)
	identifier := d.Id()
	if identifier == "" {
		identifier = catalogName
	}

	var adminCatalog *govcd.AdminCatalog
	filter, hasFilter := d.GetOk("filter")

	if hasFilter {
		adminCatalog, err = getCatalogByFilter(adminOrg, filter, vcdClient.Client.IsSysAdmin)
	} else {
		adminCatalog, err = adminOrg.GetAdminCatalogByNameOrId(identifier, false)
	}
	if err != nil {
		return diag.Errorf("error retrieving catalog '%s.%s' : %s", adminOrg.AdminOrg.Name, identifier, err)
	}

	// We try to determine early if this catalog was created by subscription, and fail otherwise
	if adminCatalog.AdminCatalog.ExternalCatalogSubscription == nil ||
		adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		return diag.Errorf("catalog '%s' is not a subscribed catalog - please use 'vcd_catalog' instead", adminCatalog.AdminCatalog.Name)
	}

	if adminCatalog.AdminCatalog.CatalogStorageProfiles != nil && len(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile) > 0 {
		storageProfileId := adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile[0].ID
		dSet(d, "storage_profile_id", storageProfileId)
	} else {
		dSet(d, "storage_profile_id", "")
	}

	dSet(d, "description", adminCatalog.AdminCatalog.Description)
	dSet(d, "created", adminCatalog.AdminCatalog.DateCreated)
	dSet(d, "publish_subscription_type", "SUBSCRIBED")
	dSet(d, "is_published", false)

	dSet(d, "subscription_url", adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location)
	dSet(d, "make_local_copy", adminCatalog.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	err = setCatalogData(d, vcdClient, adminOrg.AdminOrg.Name, adminOrg.AdminOrg.ID, adminCatalog)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	log.Printf("[TRACE] Catalog sync read initiated")

	taskIdCollection, err := readTaskIdCollection(vcdClient, adminCatalog.AdminCatalog.ID, d)
	if err != nil {
		return diag.Errorf("error retrieving task list for catalog %s: %s", adminCatalog.AdminCatalog.ID, err)
	}
	newTaskIdCollection, err := skimTaskCollection(vcdClient, taskIdCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("running_tasks", newTaskIdCollection.Running)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("failed_tasks", newTaskIdCollection.Failed)
	if err != nil {
		return diag.FromErr(err)
	}
	dSet(d, "href", adminCatalog.AdminCatalog.HREF)
	d.SetId(adminCatalog.AdminCatalog.ID)
	util.Logger.Printf("[TRACE] Subscribed Catalog data source read completed: %#v\n", adminCatalog.AdminCatalog)
	return nil
}
