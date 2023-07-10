package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type TaskIdCollection struct {
	Running []string
	Failed  []string
	skip    []string
}

type contextString string

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
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of organization to use, optional if defined at provider level.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the catalog",
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
				Required:    true,
				ForceNew:    true,
				Description: "The URL to subscribe to the external catalog.",
			},
			"subscription_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				// Those unusual password rules are dictated by the API
				// https://developer.vmware.com/apis/1260/vmware-cloud-director/doc/doc//types/ExternalCatalogSubscriptionParamsType.html
				Description: "An optional password to access the catalog. " +
					"Only ASCII characters are allowed in a valid password. " +
					"Passing in six asterisks '******' indicates to keep current password. " +
					"Passing in null or empty string indicates to remove password.",
			},
			"make_local_copy": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "If true, subscription to a catalog creates a local copy of all items. " +
					"Defaults to false, which does not create a local copy of catalog items unless a sync operation is performed. ",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog metadata. Inherited from publishing catalog",
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
				Description: "Catalog version number. Inherited from publishing catalog and updated on sync.",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name from the catalog.",
			},
			"number_of_vapp_templates": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of vApp templates this catalog contains.",
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
				Description: "True if this catalog is published. (Always false)",
			},
			"publish_subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PUBLISHED if published externally, SUBSCRIBED if subscribed to an external catalog, UNPUBLISHED otherwise. (Always SUBSCRIBED)",
			},
			"sync_on_refresh": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value that shows if sync should be performed on every refresh.",
			},
			"cancel_failed_tasks": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, the subscribed catalog will attempt canceling failed tasks",
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
				Description: "If true, synchronise this catalog. " +
					"This operation fetches the list of items. " +
					"If `make_local_copy` is set, it also fetches the items data.",
			},
			"sync_all_vapp_templates": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"sync_all", "sync_vapp_templates"},
				Description:   "If true, synchronises all vApp templates",
			},
			"sync_all_media_items": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"sync_all", "sync_media_items"},
				Description:   "If true, synchronises all media items",
			},
			"sync_vapp_templates": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"sync_all", "sync_all_vapp_templates"},
				Description:   "Synchronises vApp templates from this list of names.",
			},
			"sync_media_items": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"sync_all", "sync_all_media_items"},
				Description:   "Synchronises media items from this list of names.",
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
				Default:     false,
				Description: "If true, saves list of tasks to file for later update",
			},
			"tasks_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Where the running tasks IDs have been stored. Only if `store_tasks` is set",
			},
		},
	}
}

// resourceVcdSubscribedCatalogCreate creates a subscribed catalog
func resourceVcdSubscribedCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Println("[TRACE] entering resourceVcdSubscribedCatalogCreate")

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
	if storageProfileId != "" {
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return diag.Errorf("error looking up Storage Profile '%s' reference: %s", storageProfileId, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}
	adminCatalog, err = adminOrg.CreateCatalogFromSubscriptionAsync(types.ExternalCatalogSubscription{
		SubscribeToExternalFeeds: true,
		Location:                 subscriptionUrl,
		Password:                 password,
		LocalCopy:                makeLocalCopy,
	}, storageProfiles, catalogName, password, makeLocalCopy)
	if err != nil {
		return diag.Errorf("error creating catalog %s from subscription: %s", catalogName, err)
	}
	d.SetId(adminCatalog.AdminCatalog.ID)

	// Creation will start the initial synchronisation. A new one should not be run when `sync_on_refresh` is set
	ctx = context.WithValue(ctx, contextString("operation"), contextString("create"))
	util.Logger.Printf("[TRACE] Subscribed Catalog created: %#v\n", adminCatalog)
	return resourceVcdSubscribedCatalogRead(ctx, d, meta)
}

// resourceVcdSubscribedCatalogRead reads or refreshes the subscribed catalog
// if `sync_on_refresh` was set, it also performs the catalog synchronisation
func resourceVcdSubscribedCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Println("[TRACE] entering resourceVcdSubscribedCatalogRead")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			util.Logger.Printf("[DEBUG] Unable to find catalog. Removing from tfstate\n")
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

	var metadataStruct StringMap
	metadata, err := adminCatalog.GetMetadata()
	if err != nil {
		util.Logger.Printf("[DEBUG] Unable to find catalog metadata: %s\n", err)
		return diag.Errorf("%v", err)
	}

	if len(metadata.MetadataEntry) > 0 {
		metadataStruct = getMetadataStruct(metadata.MetadataEntry)
	}
	err = d.Set("metadata", metadataStruct)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if adminCatalog.AdminCatalog.ExternalCatalogSubscription == nil ||
		adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		// If this error occurs, we have probably a catalog data mismatch or replacement
		// i.e. we are trying to read a regular catalog as a subscribed one
		return diag.Errorf("catalog '%s' doesn't have a subscription - please use 'vcd_catalog' instead", adminCatalog.AdminCatalog.Name)
	}
	dSet(d, "subscription_url", adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location)
	dSet(d, "make_local_copy", adminCatalog.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	err = setCatalogData(d, vcdClient, adminOrg.AdminOrg.Name, adminOrg.AdminOrg.ID, adminCatalog)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	syncOnRefresh := d.Get("sync_on_refresh").(bool)
	// If the operation value was set, it indicates that a synchronisation was already started by update or create,
	// and a new one is not needed
	if ctx.Value("operation") != nil {
		value := ctx.Value("operation")
		if value == "update" || value == "create" {
			syncOnRefresh = false
		}
	}
	if syncOnRefresh {
		err = runSubscribedCatalogSyncOperations(d, vcdClient, adminCatalog, "refresh")
		if err != nil {
			return diag.Errorf("error running synchronisation for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
	}
	taskIdCollection, err := readTaskIdCollection(vcdClient, adminCatalog.AdminCatalog.ID, d)
	if err != nil {
		return diag.Errorf("error retrieving task list for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
	}
	newTaskIdCollection, err := skimTaskCollection(vcdClient, taskIdCollection)
	if err != nil && !syncOnRefresh {
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
	util.Logger.Printf("[TRACE] Subscribed Catalog read completed: %#v\n", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdSubscribedCatalogUpdate updates a subscribed catalog
func resourceVcdSubscribedCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Println("[TRACE] entering resourceVcdSubscribedCatalogUpdate")
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
	ctx = context.WithValue(ctx, contextString("operation"), contextString("update"))
	return genericResourceVcdCatalogUpdate(ctx, d, meta,
		[]moreUpdateCatalogFunc{
			updateSubscriptionFunc,
			runSubscribedCatalogSyncOperations,
		},
		resourceVcdSubscribedCatalogRead)
}

// resourceVcdSubscribedCatalogDelete deletes a subscribed catalog
// It defers the main operation to the regular catalog resource, and then removes the tasks file, if it exists
func resourceVcdSubscribedCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Println("[TRACE] entering resourceVcdSubscribedCatalogDelete")
	diagErr := resourceVcdCatalogDelete(ctx, d, meta)
	fileName, err := getTaskListFileName(d.Id(), nil)
	if err == nil && fileName != "" {
		err = os.Remove(fileName)
		if err != nil {
			util.Logger.Printf("[INFO] no task file found for catalog %s\n", d.Get("name").(string))
		}
	}
	return diagErr
}

// resourceVcdSubscribedCatalogImport imports a subscribed catalog
// The catalog can be retrieved by name or ID. e.g.:
// terraform import vcd_subscribed_catalog.catalog-name org-name.catalog-name
// or
// terraform import vcd_subscribed_catalog.catalog-name  org-name.catalog-id
func resourceVcdSubscribedCatalogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	util.Logger.Println("[TRACE] entering resourceVcdSubscribedCatalogImport")
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

	if catalog.AdminCatalog.ExternalCatalogSubscription == nil ||
		catalog.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		// If this error occurs, we have probably a catalog data mismatch or replacement
		// i.e. we are trying to read a regular catalog as a subscribed one
		return nil, fmt.Errorf("catalog '%s' doesn't have a subscription- Please use 'vcd_catalog' instead", catalog.AdminCatalog.Name)
	}
	dSet(d, "org", orgName)
	dSet(d, "name", catalogIdentifier)
	dSet(d, "description", catalog.AdminCatalog.Description)
	d.SetId(catalog.AdminCatalog.ID)

	return []*schema.ResourceData{d}, nil
}

// getTaskListFileName retrieves the task list file name.
// The name will never change for this catalog
func getTaskListFileName(catalogId string, d *schema.ResourceData) (string, error) {
	util.Logger.Println("[TRACE] entering getTaskListFileName")
	fileName, err := filepath.Abs(strings.Replace(taskFileName, "{ID}", extractUuid(catalogId), 1))
	if err != nil {
		return "", err
	}
	if d != nil {
		dSet(d, "tasks_file_name", fileName)
	}
	return fileName, nil
}

// storeTaskIdCollection will store the task ID collections into the tasks file if "store_tasks" was set
func storeTaskIdCollection(catalogId string, collection TaskIdCollection, d *schema.ResourceData) error {
	util.Logger.Println("[TRACE] entering storeTaskIdCollection")
	if !d.Get("store_tasks").(bool) {
		return nil
	}
	fileName, err := getTaskListFileName(catalogId, d)
	if err != nil {
		return fmt.Errorf("error setting file name for task collection: %s", err)
	}
	encoded, err := json.MarshalIndent(collection, " ", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Clean(fileName), encoded, 0600)
	if err != nil {
		return err
	}

	util.Logger.Printf(" storing %# v\n", pretty.Formatter(collection))
	return nil
}

// readTaskIdCollection reads the collection of tasks from file, if "store_tasks" was set
func readTaskIdCollection(vcdClient *VCDClient, catalogId string, d *schema.ResourceData) (TaskIdCollection, error) {
	util.Logger.Println("[TRACE] entering readTaskIdCollection")
	var collection TaskIdCollection
	rawRunningTasks := d.Get("running_tasks")
	if rawRunningTasks != nil {
		for _, item := range rawRunningTasks.([]interface{}) {
			collection.Running = append(collection.Running, item.(string))
		}
	}
	rawFailedTasks := d.Get("failed_tasks")
	if rawFailedTasks != nil {
		for _, item := range rawFailedTasks.([]interface{}) {
			collection.Failed = append(collection.Failed, item.(string))
		}
	}
	rawStoreTasks := d.Get("store_tasks")
	storeTasks := rawStoreTasks != nil && rawStoreTasks.(bool)
	if !storeTasks {
		return collection, nil
	}
	fileName, err := getTaskListFileName(catalogId, d)
	if err != nil {
		return TaskIdCollection{}, fmt.Errorf("error setting file name for task collection: %s", err)
	}
	contents, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return TaskIdCollection{}, nil
		}
		return TaskIdCollection{}, err
	}
	err = json.Unmarshal(contents, &collection)
	if err != nil {
		return TaskIdCollection{}, err
	}
	util.Logger.Printf("reading %# v\n", pretty.Formatter(collection))
	// before setting the list of tasks, we remove the ones that are already completed
	collection, err = skimTaskCollection(vcdClient, collection)
	if err != nil {
		return TaskIdCollection{}, err
	}
	err = d.Set("running_tasks", collection.Running)
	if err != nil {
		return TaskIdCollection{}, err
	}
	err = d.Set("failed_tasks", collection.Failed)
	if err != nil {
		return TaskIdCollection{}, err
	}
	return collection, nil
}

func processTasks(vcdClient *VCDClient, collection TaskIdCollection, callNo int) (TaskIdCollection, error, bool) {
	needsRepeating := false
	running, failed, err := vcdClient.Client.SkimTasksList(collection.Running)
	if err != nil {
		return TaskIdCollection{}, err, false
	}
	var newFailedTasks []string
	if len(failed) > 0 {
		errorMessage := ""
		taskTypes := make(map[string]bool)
		for _, taskId := range failed {
			task, err := vcdClient.VCDClient.Client.GetTaskById(taskId)
			if err != nil {
				return TaskIdCollection{}, err, false
			}
			taskTypes[task.Task.Name] = true
			if task.Task.Error == nil {
				continue
			}
			if errorMessage != "" {
				errorMessage += "\n"
			}
			// Catch a possible race condition
			// If this message is in the task error message, we don't return the task ID to the caller,
			// but keep it for the next run
			if strings.Contains(task.Task.Error.Error(), "updated or deleted by another transaction") {
				util.Logger.Printf("[TRACE] [SKIP-FAILED-TASK] task %s skipped. Error: %s\n", taskId, task.Task.Error)
				needsRepeating = true
				collection.skip = append(collection.skip, taskId)
			} else {
				errorMessage += task.Task.Error.Error()
				newFailedTasks = append(newFailedTasks, taskId)
			}
		}
		taskTypeText := ""
		for k := range taskTypes {
			taskTypeText += k + " "
		}
		err = fmt.Errorf("%d tasks have failed - task types [%s]: %s", len(failed), taskTypeText, errorMessage)

	}
	// On the second run, if we still have task failures, we add the failed tasks from the previous run
	if callNo > 1 && len(newFailedTasks) > 0 {
		newFailedTasks = append(newFailedTasks, collection.skip...)
	}
	return TaskIdCollection{
		Running: running,
		Failed:  newFailedTasks,
		skip:    collection.skip,
	}, err, needsRepeating
}

// skimTaskCollection will remove from the task list all the tasks that are already complete
func skimTaskCollection(vcdClient *VCDClient, collection TaskIdCollection) (TaskIdCollection, error) {
	util.Logger.Println("[TRACE] entering skimTaskCollection")
	var needsRepeating bool
	var err error
	collection, err, needsRepeating = processTasks(vcdClient, collection, 1)
	if err != nil {
		// If a possible race was detected, repeat the skim a second time
		if needsRepeating {
			time.Sleep(10 * time.Second)
			collection, err, _ = processTasks(vcdClient, collection, 2)
		}
	}
	if err != nil {
		return TaskIdCollection{}, err
	}
	return collection, nil
}

// runSubscribedCatalogSyncOperations runs all the requested synchronisation operations
func runSubscribedCatalogSyncOperations(d *schema.ResourceData, vcdClient *VCDClient, adminCatalog *govcd.AdminCatalog, operation string) error {
	util.Logger.Printf("[TRACE] Catalog '%s' sync initiated [%s]\n", adminCatalog.AdminCatalog.Name, operation)

	catalogId := adminCatalog.AdminCatalog.ID
	makeLocalCopy := d.Get("make_local_copy").(bool)
	syncAll := d.Get("sync_all").(bool)
	syncCatalog := d.Get("sync_catalog").(bool)
	syncAllVappTemplates := d.Get("sync_all_vapp_templates").(bool)
	syncAllMediaItems := d.Get("sync_all_media_items").(bool)
	rawSyncMediaItems := d.Get("sync_media_items").([]interface{})
	rawSyncVappTemplates := d.Get("sync_vapp_templates").([]interface{})

	var collection TaskIdCollection
	var taskList []string
	var err error
	collection, err = readTaskIdCollection(vcdClient, catalogId, d)
	if err != nil {
		return fmt.Errorf("error reading catalog task IDs: %s", err)
	}
	taskList = collection.Running

	cancelFailedTasks := d.Get("cancel_failed_tasks").(bool)
	if len(collection.Failed) > 0 {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - collecting failed tasks\n", adminCatalog.AdminCatalog.Name)
		for _, taskId := range collection.Failed {
			task, err := vcdClient.Client.GetTaskById(taskId)
			if err == nil {
				util.Logger.Printf("[runSubscribedCatalogSyncOperations] %s (%s) %s\n",
					task.Task.OperationName, task.Task.Status, task.Task.Operation)
				if cancelFailedTasks {
					err = task.CancelTask()
					if err != nil {
						util.Logger.Printf("[runSubscribedCatalogSyncOperations] error canceling task %s\n", taskId)
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
		collection, err = skimTaskCollection(vcdClient, collection)
		if err != nil {
			return fmt.Errorf("error running skimTaskCollection: %s", err)
		}
	}

	if len(collection.Failed) > 0 {
		if operation == "refresh" {
			return nil
		}
		return fmt.Errorf("%d tasks failed. See logs for details", len(collection.Failed))
	}

	if syncCatalog || syncAll {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - sync_catalog [make_local_copy=%v]\n", adminCatalog.AdminCatalog.Name, makeLocalCopy)
		var task *govcd.Task
		if makeLocalCopy {
			task, err = adminCatalog.LaunchSync()
		} else {
			err = adminCatalog.Sync()
		}
		if err != nil {
			return fmt.Errorf("error synchronising catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		if makeLocalCopy && task != nil && task.Task != nil {
			taskList = append(taskList, task.Task.ID)
		}
	}
	// If the `make_local_copy` property was set, we don't need to synchronise anything more than the catalog
	if makeLocalCopy {
		collection.Running = taskList
		return storeTaskIdCollection(catalogId, collection, d)
	}
	if syncAllVappTemplates || syncAll {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - sync_all_vapp_templates [make_local_copy=%v]\n", adminCatalog.AdminCatalog.Name, makeLocalCopy)
		tasks, err := adminCatalog.LaunchSynchronisationAllVappTemplates()
		if err != nil {
			return fmt.Errorf("error synchronising all vApp templates for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			taskList = append(taskList, task.Task.ID)
		}
	}
	if syncAllMediaItems || syncAll {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - sync_all_media_items [make_local_copy=%v]\n", adminCatalog.AdminCatalog.Name, makeLocalCopy)
		tasks, err := adminCatalog.LaunchSynchronisationAllMediaItems()
		if err != nil {
			return fmt.Errorf("error synchronising all media items for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			if task != nil {
				taskList = append(taskList, task.Task.ID)
			}
		}
	}
	if len(rawSyncVappTemplates) > 0 {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - sync_vapp_templates [make_local_copy=%v]\n", adminCatalog.AdminCatalog.Name, makeLocalCopy)
		var syncVappTemplates []string
		for _, item := range rawSyncVappTemplates {
			syncVappTemplates = append(syncVappTemplates, item.(string))
		}
		tasks, err := adminCatalog.LaunchSynchronisationVappTemplates(syncVappTemplates)
		if err != nil {
			return fmt.Errorf("error synchronising vApp templates for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			taskList = append(taskList, task.Task.ID)
		}
	}
	if len(rawSyncMediaItems) > 0 {
		util.Logger.Printf("[TRACE] Catalog '%s' sync - sync_all_media_items [make_local_copy=%v]\n", adminCatalog.AdminCatalog.Name, makeLocalCopy)
		var syncMediaItems []string
		for _, item := range rawSyncMediaItems {
			syncMediaItems = append(syncMediaItems, item.(string))
		}
		tasks, err := adminCatalog.LaunchSynchronisationMediaItems(syncMediaItems)
		if err != nil {
			return fmt.Errorf("error synchronising media items for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			taskList = append(taskList, task.Task.ID)
		}
	}

	collection.Running = taskList
	return storeTaskIdCollection(catalogId, collection, d)
}
