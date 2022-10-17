package vcd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type TaskIdCollection struct {
	Running []string
	Failed  []string
}

/*
// TODO: remove this resource, once it has been proven that vcd_subscribed_catalog can handle all scenarios

	func resourceVcdCatalogSync() *schema.Resource {
		return &schema.Resource{
			CreateContext: resourceVcdCatalogSyncCreate,
			ReadContext:   resourceVcdCatalogSyncRead,
			UpdateContext: resourceVcdCatalogSyncUpdate,
			DeleteContext: resourceVcdCatalogSyncDelete,
			Importer: &schema.ResourceImporter{
				StateContext: resourceVcdCatalogSyncImport,
			},
			Schema: map[string]*schema.Schema{
				"catalog_id": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "The ID of Catalog to use",
				},
				"catalog_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of Catalog to use",
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

	func resourceVcdCatalogSyncCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		return resourceVcdCatalogSyncCreateUpdate(ctx, d, meta, "create")
	}

	func resourceVcdCatalogSyncUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		return resourceVcdCatalogSyncRead(ctx, d, meta)
	}

	func resourceVcdCatalogSyncCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
		log.Printf("[TRACE] Catalog sync %s initiated", operation)

		vcdClient := meta.(*VCDClient)

		catalogId := d.Get("catalog_id").(string)

		adminCatalog, err := vcdClient.VCDClient.Client.GetAdminCatalogById(catalogId)
		if err != nil {
			return diag.Errorf("error retrieving catalog %s: %s", catalogId, err)
		}
		err = resourceVcdSubscribedCatalogSync(d, vcdClient, adminCatalog, "update")
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(adminCatalog.AdminCatalog.ID)
		dSet(d, "catalog_name", adminCatalog.AdminCatalog.Name)
		return resourceVcdCatalogSyncRead(ctx, d, meta)
	}

	func resourceVcdCatalogSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		log.Printf("[TRACE] Catalog sync read initiated")

		vcdClient := meta.(*VCDClient)

		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return diag.Errorf(errorRetrievingOrg, err)
		}
		catalogId := d.Get("catalog_id").(string)

		adminCatalog, err := adminOrg.GetAdminCatalogById(catalogId, false)
		if err != nil {
			return diag.Errorf("error retrieving catalog %s: %s", catalogId, err)
		}
		d.SetId(adminCatalog.AdminCatalog.ID)
		dSet(d, "catalog_name", adminCatalog.AdminCatalog.Name)

		taskIdCollection, err := readTaskIdCollection(vcdClient, catalogId, d)
		if err != nil {
			return diag.Errorf("error retrieving task list for catalog %s: %s", catalogId, err)
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
		err = storeTaskIdCollection(catalogId, newTaskIdCollection, d)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	func resourceVcdCatalogSyncDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		d.SetId("")
		return nil
	}

// resourceVcdCatalogSyncImport imports a vcd_catalog_sync into state
// It can be identified in three ways:
// * terraform import vcd_catalog_sync.name catalog-ID
// * terraform import vcd_catalog_sync.name org-name.catalog-name
// * terraform import vcd_catalog_sync.name org-name.catalog-ID
func resourceVcdCatalogSyncImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

		vcdClient := meta.(*VCDClient)
		resourceURI := strings.Split(d.Id(), ImportSeparator)
		var catalog *govcd.Catalog
		var catalogIdentifier string
		var err error

		if len(resourceURI) == 1 {
			catalogIdentifier = resourceURI[0]
			catalog, err = vcdClient.VCDClient.Client.GetCatalogById(catalogIdentifier)
		} else {
			var orgName string
			if len(resourceURI) != 2 {
				return nil, fmt.Errorf("resource name must be specified as org.catalogID or org.catalogName")
			}

			orgName, catalogIdentifier = resourceURI[0], resourceURI[1]

			org, err := vcdClient.GetOrg(orgName)
			if err != nil {
				return nil, fmt.Errorf(errorRetrievingOrg, err)
			}
			catalog, err = org.GetCatalogByNameOrId(catalogIdentifier, false)
		}
		if err != nil {
			return nil, fmt.Errorf("[catalog sync import] error retrieving catalog '%s'", catalogIdentifier)
		}
		dSet(d, "catalog_id", catalog.Catalog.ID)
		d.SetId(catalog.Catalog.ID)

		return []*schema.ResourceData{d}, nil
	}
*/

// TODO: if we remove vcd_catalog_sync, move the functions below to resource_vcd_subscribed_catalog.go
func getTaskListFileName(catalogId string, d *schema.ResourceData) (string, error) {
	fileName, err := filepath.Abs(strings.Replace(taskFileName, "{ID}", extractUuid(catalogId), 1))
	if err != nil {
		return "", err
	}
	dSet(d, "tasks_file_name", fileName)
	return fileName, nil
}

func storeTaskIdCollection(catalogId string, collection TaskIdCollection, d *schema.ResourceData) error {
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
	err = os.WriteFile(fileName, encoded, 0666)
	if err != nil {
		return err
	}

	util.Logger.Printf(" storing %# v\n", pretty.Formatter(collection))
	return nil
}

func readTaskIdCollection(vcdClient *VCDClient, catalogId string, d *schema.ResourceData) (TaskIdCollection, error) {
	var collection TaskIdCollection
	fileName, err := getTaskListFileName(catalogId, d)
	if err != nil {
		return TaskIdCollection{}, fmt.Errorf("error setting file name for task collection: %s", err)
	}
	contents, err := os.ReadFile(fileName)
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

func skimTaskCollection(vcdClient *VCDClient, collection TaskIdCollection) (TaskIdCollection, error) {

	running, failed, err := vcdClient.Client.SkimTasksList(collection.Running)
	if err != nil {
		return TaskIdCollection{}, err
	}
	return TaskIdCollection{
		Running: running,
		Failed:  failed,
	}, nil
}

func resourceVcdSubscribedCatalogSync(d *schema.ResourceData, vcdClient *VCDClient, adminCatalog *govcd.AdminCatalog, operation string) error {
	log.Printf("[TRACE] Catalog '%s' sync initiated [%s]", adminCatalog.AdminCatalog.Name, operation)

	catalogId := adminCatalog.AdminCatalog.ID
	syncAll := d.Get("sync_all").(bool)
	syncCatalog := d.Get("sync_catalog").(bool)
	syncAllVappTemplates := d.Get("sync_all_vapp_templates").(bool)
	syncAllMediaItems := d.Get("sync_all_media_items").(bool)
	rawSyncMediaItems := d.Get("sync_media_items").([]interface{})
	rawSyncVappTemplates := d.Get("sync_vapp_templates").([]interface{})

	var collection TaskIdCollection
	var taskList []string
	var err error
	if operation == "update" || operation == "refresh" {
		collection, err = readTaskIdCollection(vcdClient, catalogId, d)
		if err != nil {
			return fmt.Errorf("error reading catalog task IDs")
		}
		taskList = collection.Running
	}

	if syncCatalog || syncAll {
		//task, err := adminCatalog.LaunchSync()
		err = adminCatalog.Sync()
		if err != nil {
			return fmt.Errorf("error synchronising catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		//taskList = append(taskList, task.Task.ID)
	}
	if syncAllVappTemplates || syncAll {
		tasks, err := adminCatalog.LaunchSynchronisationAllVappTemplates()
		if err != nil {
			return fmt.Errorf("error synchronising all vApp templates for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			taskList = append(taskList, task.Task.ID)
		}
	}
	if syncAllMediaItems || syncAll {
		tasks, err := adminCatalog.LaunchSynchronisationAllMediaItems()
		if err != nil {
			return fmt.Errorf("error synchronising all media items for catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		for _, task := range tasks {
			taskList = append(taskList, task.Task.ID)
		}
	}
	if len(rawSyncVappTemplates) > 0 {
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
