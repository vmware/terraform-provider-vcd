package vcd

import (
	"flag"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
	"log"
	"os"
	"time"
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
				Required: false,
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
				Required: false,
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
				Required:    false,
				Optional:    true,
				ForceNew:    false,
				Default:     1,
				Description: "size of upload file piece size in mega bytes",
			},
			"show_upload_progress": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				ForceNew:    false,
				Description: "shows upload progress in stdout",
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
		return fmt.Errorf("error waiting from task to complete: %+v", err)
	}

	d.SetId(catalogName + ":" + itemName)

	log.Printf("[TRACE] Catalog item created: %#v", itemName)
	return resourceVcdCatalogItemRead(d, meta)
}

func resourceVcdCatalogItemRead(d *schema.ResourceData, meta interface{}) error {
	return findCatalogItem(d, meta.(*VCDClient))
}

func resourceVcdCatalogItemDelete(d *schema.ResourceData, meta interface{}) error {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

//update function for "show_upload_progress" and "upload_piece_size"
func resourceVcdCatalogItemUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}
