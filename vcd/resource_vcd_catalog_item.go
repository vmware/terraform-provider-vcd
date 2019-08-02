package vcd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdCatalogItemCreate,
		Delete: resourceVcdCatalogItemDelete,
		Read:   resourceVcdCatalogItemRead,
		Update: resourceVcdCatalogItemUpdate,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
	if err != nil || adminOrg == (govcd.AdminOrg{}) {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := adminOrg.FindCatalog(catalogName)
	if err != nil || catalog == (govcd.Catalog{}) {
		log.Printf("Error finding Catalog: %#v", err)
		return fmt.Errorf("error finding Catalog: %#v", err)
	}

	uploadPieceSize := d.Get("upload_piece_size").(int)
	itemName := d.Get("name").(string)
	task, err := catalog.UploadOvf(d.Get("ova_path").(string), itemName, d.Get("description").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes
	if err != nil {
		log.Printf("Error uploading new catalog item: %#v", err)
		return fmt.Errorf("error uploading new catalog item: %#v", err)
	}

	var terraformStdout *os.File
	// Needed to avoid errors when uintptr(4) is used
	if v := flag.Lookup("test.v"); v == nil || v.Value.String() != "true" {
		terraformStdout = os.NewFile(uintptr(4), "stdout")
	} else {
		terraformStdout = os.Stdout
	}

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

	d.SetId(catalogName + ":" + itemName)

	log.Printf("[TRACE] Catalog item created: %#v", itemName)

	err = createOrUpdateCatalogItemMetadata(d, meta)
	if err != nil {
		return fmt.Errorf("error adding catalog item metadata: %s", err)
	}

	return resourceVcdCatalogItemRead(d, meta)
}

func resourceVcdCatalogItemRead(d *schema.ResourceData, meta interface{}) error {
	catalogItem, err := findCatalogItem(d, meta.(*VCDClient))
	if err != nil {
		return err
	}

	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return err
	}

	metadata, err := vAppTemplate.GetMetadata()
	d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
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

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.FindCatalog(d.Get("catalog").(string))
	if err != nil || catalog == (govcd.Catalog{}) {
		log.Printf("[DEBUG] Unable to find catalog: %s", err)
		return nil
	}

	catalogItem, err := catalog.FindCatalogItem(d.Get("name").(string))
	if err != nil || catalogItem == (govcd.CatalogItem{}) {
		log.Printf("[DEBUG] Unable to find catalog item: %s", err)
		return nil
	}

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
