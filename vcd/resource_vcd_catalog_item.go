package vcd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdCatalogItemCreate,
		Delete: resourceVcdCatalogItemDelete,
		Read:   resourceVcdCatalogItemRead,
		Update: resourceVcdCatalogItemUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceVcdCatalogItemImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog name where upload the OVA file",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog item name",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"created": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the item was created",
			},
			"ova_path": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "absolute or relative path to OVA",
			},
			"upload_piece_size": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Default:     1,
				Description: "size of upload file piece size in mega bytes",
			},
			"show_upload_progress": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "shows upload progress in stdout",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for catalog item metadata",
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
		},
	}
}

func resourceVcdCatalogItemCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog item creation initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %#v", err)
		return fmt.Errorf("error finding Catalog: %#v", err)
	}

	uploadPieceSize := d.Get("upload_piece_size").(int)
	itemName := d.Get("name").(string)
	task, err := catalog.UploadOvf(d.Get("ova_path").(string), itemName, d.Get("description").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes
	if err != nil {
		log.Printf("[DEBUG] Error uploading new catalog item: %#v", err)
		return fmt.Errorf("error uploading new catalog item: %#v", err)
	}

	terraformStdout := getTerraformStdout()

	if d.Get("show_upload_progress").(bool) {
		for {
			if err := getError(task); err != nil {
				return err
			}
			_, _ = fmt.Fprint(terraformStdout, "vcd_catalog_item."+itemName+": Upload progress "+task.GetUploadProgress()+"%\n")
			if task.GetUploadProgress() == "100.00" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	if d.Get("show_upload_progress").(bool) {
		for {
			progress, err := task.GetTaskProgress()
			if err != nil {
				log.Printf("vCD Error importing new catalog item: %#v", err)
				return fmt.Errorf("vCD Error importing new catalog item: %#v", err)
			}
			_, _ = fmt.Fprint(terraformStdout, "vcd_catalog_item."+itemName+": vCD import catalog item progress "+progress+"%\n")
			if progress == "100" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task to complete: %+v", err)
	}

	item, err := catalog.GetCatalogItemByName(itemName, true)
	if err != nil {
		return fmt.Errorf("error retrieving catalog item %s: %s", itemName, err)
	}
	d.SetId(item.CatalogItem.ID)

	log.Printf("[TRACE] Catalog item created: %#v", itemName)

	err = createOrUpdateCatalogItemMetadata(d, meta)
	if err != nil {
		return fmt.Errorf("error adding catalog item metadata: %s", err)
	}

	return resourceVcdCatalogItemRead(d, meta)
}

func resourceVcdCatalogItemRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdCatalogItemRead(d, meta, "resource")
}

func genericVcdCatalogItemRead(d *schema.ResourceData, meta interface{}, origin string) error {
	catalogItem, err := findCatalogItem(d, meta.(*VCDClient), origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return err
	}
	if catalogItem == nil {
		log.Printf("[DEBUG] Unable to find media item: %s. Removing from tfstate", err)
		return err
	}

	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return err
	}

	metadata, err := vAppTemplate.GetMetadata()
	if err != nil {
		return err
	}
	_ = d.Set("name", catalogItem.CatalogItem.Name)
	_ = d.Set("created", vAppTemplate.VAppTemplate.DateCreated)
	_ = d.Set("description", catalogItem.CatalogItem.Description)
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))

	return err
}

func resourceVcdCatalogItemDelete(d *schema.ResourceData, meta interface{}) error {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

// currently updates only metadata
func resourceVcdCatalogItemUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createOrUpdateCatalogItemMetadata(d, meta)
	if err != nil {
		return fmt.Errorf("error updating catalog item metadata: %s", err)
	}
	return nil
}

func createOrUpdateCatalogItemMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata for catalog item")

	catalogItem, err := findCatalogItem(d, meta.(*VCDClient), "resource")
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return err
	}

	// We have to add metadata to template to see in UI
	// catalog item is another abstraction and has own metadata which we don't see in UI
	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return err
	}

	if d.HasChange("metadata") {
		oldRaw, newRaw := d.GetChange("metadata")
		oldMetadata := oldRaw.(map[string]interface{})
		newMetadata := newRaw.(map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		for k := range oldMetadata {
			if _, ok := newMetadata[k]; !ok {
				toBeRemovedMetadata = append(toBeRemovedMetadata, k)
			}
		}
		for _, k := range toBeRemovedMetadata {
			err := vAppTemplate.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			_, err := vAppTemplate.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}

// Imports a CatalogItem into Terraform state
// This function task is to get the data from vCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name.catalog_item_name
//
// Example import path (id): org_name.catalog_name.catalog_item_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogItemImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog.catalog_item")
	}
	orgName, catalogName, catalogItemName := resourceURI[0], resourceURI[1], resourceURI[2]

	if orgName == "" {
		return nil, fmt.Errorf("import: empty org name provided")
	}
	if catalogName == "" {
		return nil, fmt.Errorf("import: empty catalog name provided")
	}
	if catalogItemName == "" {
		return nil, fmt.Errorf("import: empty catalog item name provided")
	}

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	catalogItem, err := catalog.GetCatalogItemByName(catalogItemName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	_ = d.Set("org", orgName)
	_ = d.Set("catalog", catalogName)
	_ = d.Set("name", catalogItemName)
	_ = d.Set("description", catalogItem.CatalogItem.Description)
	d.SetId(catalogItem.CatalogItem.ID)

	return []*schema.ResourceData{d}, nil
}
