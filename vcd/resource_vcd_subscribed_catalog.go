package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const taskFileName = "vcd-catalog-sync-tasks-{ID}.json"

func resourceVcdSubscribedCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSubscribedCatalogCreate,
		ReadContext:   resourceVcdSubscribedCatalogRead,
		UpdateContext: resourceVcdSubscribedCatalogUpdate,
		DeleteContext: resourceVcdSubscribedCatalogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSubscribedCatalogImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A subscribed catalog description is inherited from the publisher catalog and cannot be changed. It is updated on sync",
			},
			"storage_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional storage profile ID",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state.",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal.",
			},
			"subscription_url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The URL to subscribe to the external catalog. Required when 'subscription_catalog_href' is not provided",
			},
			"subscription_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				Description: "An optional password to access the catalog. Only ASCII characters are allowed in a valid password." +
					"Passing in six asterisks '******' indicates to keep current password. Passing in null or empty string indicates to remove password.",
			},
			"make_local_copy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If true, subscription to a catalog creates a local copy of all items. Defaults to false, which does not create a local copy of catalogItems unless sync operation is performed. ",
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				//Default:      15,
				RequiredWith: []string{"make_local_copy"},
				Description:  "Timeout (in minutes) for import completion. Required when 'make_local_copy' is true. Default 15 minutes. (0 = no timeout)",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for catalog metadata.",
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
				Description: "Catalog version number.",
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
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is shared to all organizations.",
			},
			"sync_on_refresh": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value that shows if sync should be performed on every refresh",
			},
			"sync_all": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{"sync_catalog", "sync_all_vapp_templates", "sync_vapp_templates",
					"sync_all_media_items", "sync_media_items", "sync_vapp_templates"},
				Description: "If true, synchronise this catalog and all items",
			},
			"sync_catalog": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"sync_all"},
				Description:   "If true, synchronise this catalog",
			},
			"sync_all_vapp_templates": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"sync_all", "sync_vapp_templates"},
				Description:   "if true, synchronises all vApp templates",
			},
			"sync_all_media_items": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"sync_all", "sync_media_items"},
				Description:   "if true, synchronises all media items",
			},
			"sync_vapp_templates": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"sync_all", "sync_all_vapp_templates"},
				Description:   "Synchronises vApp templates from this list of names",
			},
			"sync_media_items": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"sync_all", "sync_all_media_items"},
				Description:   "Synchronises media items from this list of names",
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
			"store_tasks": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, saves list of tasks to file for later update",
			},
			"tasks_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "where the running tasks IDs have been stored",
			},
		},
	}
}

func resourceVcdSubscribedCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Subscribed Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var adminCatalog *govcd.AdminCatalog
	var storageProfiles *types.CatalogStorageProfiles

	catalogName := d.Get("name").(string)
	password := d.Get("subscription_password").(string)
	storageProfileId := d.Get("storage_profile_id").(string)
	subscriptionUrl := d.Get("subscription_url").(string)
	makeLocalCopy := d.Get("make_local_copy").(bool)
	rawTimeout, okTimeout := d.GetOk("timeout")
	var timeout int
	if okTimeout {
		timeout = rawTimeout.(int)
	}
	if storageProfileId != "" {
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return diag.Errorf("error looking up Storage Profile '%s' reference: %s", storageProfileId, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}
	if makeLocalCopy && okTimeout {
		adminCatalog, err = adminOrg.CreateCatalogFromSubscription(types.ExternalCatalogSubscription{
			SubscribeToExternalFeeds: true,
			Location:                 subscriptionUrl,
			Password:                 password,
			LocalCopy:                makeLocalCopy,
		}, storageProfiles, catalogName, password, makeLocalCopy, time.Duration(timeout)*time.Minute)
	} else {
		adminCatalog, err = adminOrg.CreateCatalogFromSubscriptionAsync(types.ExternalCatalogSubscription{
			SubscribeToExternalFeeds: true,
			Location:                 subscriptionUrl,
			Password:                 password,
			LocalCopy:                makeLocalCopy,
		}, storageProfiles, catalogName, password, makeLocalCopy)
	}
	if err != nil {
		return diag.Errorf("error creating catalog %s from subscription: %s", catalogName, err)
	}
	d.SetId(adminCatalog.AdminCatalog.ID)

	log.Printf("[TRACE] creating metadata for catalog")
	err = createOrUpdateMetadata(d, adminCatalog, "metadata")
	if err != nil {
		return diag.Errorf("%v", err)
	}
	log.Printf("[TRACE] Subscribed Catalog created: %#v", adminCatalog)
	return resourceVcdSubscribedCatalogRead(ctx, d, meta)
}

func resourceVcdSubscribedCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Subscribed Catalog read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
			d.SetId("")
			return nil
		}

		return diag.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	if adminCatalog.AdminCatalog.CatalogStorageProfiles != nil && len(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile) > 0 {
		storageProfileId := adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile[0].ID
		dSet(d, "storage_profile_id", storageProfileId)
	} else {
		dSet(d, "storage_profile_id", "")
	}

	dSet(d, "description", adminCatalog.AdminCatalog.Description)
	dSet(d, "created", adminCatalog.AdminCatalog.DateCreated)

	metadata, err := adminCatalog.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog metadata: %s", err)
		return diag.Errorf("%v", err)
	}

	if len(metadata.MetadataEntry) > 0 {
		err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
		if err != nil {
			return diag.Errorf("%v", err)

		}
	}
	if adminCatalog.AdminCatalog.ExternalCatalogSubscription != nil {
		dSet(d, "subscription_url", adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location)
		dSet(d, "make_local_copy", adminCatalog.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	}
	err = setCatalogData(d, adminOrg, adminCatalog, "vcd_subscribed_catalog")
	if err != nil {
		return diag.Errorf("%v", err)
	}

	syncOnRefresh := d.Get("sync_on_refresh").(bool)
	if syncOnRefresh {
		err = resourceVcdSubscribedCatalogSync(d, vcdClient, adminCatalog, "refresh")
	}
	taskIdCollection, err := readTaskIdCollection(vcdClient, adminCatalog.AdminCatalog.ID, d)
	if err != nil {
		return diag.Errorf("error retrieving task list for catalog %s: %s", adminCatalog.AdminCatalog.ID, err)
	}
	newTaskIdCollection, err := skimTaskCollection(vcdClient, taskIdCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	// add internal tasks
	if adminCatalog.AdminCatalog.Tasks != nil {
		var seenTasks = make(map[string]bool)
		for _, existingTask := range taskIdCollection.Running {
			seenTasks[existingTask] = true
		}
		tasks := adminCatalog.AdminCatalog.Tasks.Task
		for _, task := range tasks {
			_, seen := seenTasks[task.ID]
			if !seen {
				taskIdCollection.Running = append(taskIdCollection.Running, task.ID)
				seenTasks[task.ID] = true
			}
			// add the subtasks, if any
			if task.Tasks != nil {
				for _, subTask := range task.Tasks.Task {
					_, seen = seenTasks[subTask.ID]
					if !seen {
						taskIdCollection.Running = append(taskIdCollection.Running, subTask.ID)
					}
					seenTasks[subTask.ID] = true
				}
			}
		}
	}
	// give it a chance to remove finished tasks
	if syncOnRefresh {
		time.Sleep(3 * time.Second)
		newTaskIdCollection, err = skimTaskCollection(vcdClient, taskIdCollection)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = d.Set("running_tasks", newTaskIdCollection.Running)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("failed_tasks", newTaskIdCollection.Failed)
	if err != nil {
		return diag.FromErr(err)
	}
	err = storeTaskIdCollection(adminCatalog.AdminCatalog.ID, newTaskIdCollection, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dSet(d, "href", adminCatalog.AdminCatalog.HREF)
	d.SetId(adminCatalog.AdminCatalog.ID)
	log.Printf("[TRACE] Subscribed Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

func resourceVcdSubscribedCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// The update in this resource differs from vcd_catalog only in the subscription and synchronising data
	// Thus, we use the vcd_catalog update with a custom function to update the subscription and sync parameters

	var updateSubscriptionFunc moreUpdateCatalogFunc
	if d.HasChanges("subscription_url", "make_local_copy", "subscription_password") {
		params := types.ExternalCatalogSubscription{
			SubscribeToExternalFeeds: true,
			Location:                 d.Get("subscription_url").(string),
			Password:                 d.Get("subscription_password").(string),
			LocalCopy:                d.Get("make_local_copy").(bool),
		}
		updateSubscriptionFunc = func(_ *schema.ResourceData, _ *VCDClient, c *govcd.AdminCatalog, _ string) error {
			return c.UpdateSubscriptionParams(params)
		}
	}
	return genericResourceVcdCatalogUpdate(ctx, d, meta,
		[]moreUpdateCatalogFunc{
			updateSubscriptionFunc,
			resourceVcdSubscribedCatalogSync,
		},
		resourceVcdSubscribedCatalogRead)
}

func resourceVcdSubscribedCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdCatalogDelete(ctx, d, meta)
}

func resourceVcdSubscribedCatalogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.catalog-name or org-name.catalog-ID")
	}
	orgName, catalogIdentifier := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)

	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := adminOrg.GetAdminCatalogByNameOrId(catalogIdentifier, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "name", catalogIdentifier)
	dSet(d, "description", catalog.AdminCatalog.Description)
	d.SetId(catalog.AdminCatalog.ID)

	return []*schema.ResourceData{d}, nil
}
