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

func resourceVcdCatalogMedia() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdMediaCreate,
		Delete: resourceVcdMediaDelete,
		Read:   resourceVcdMediaRead,
		Update: resourceVcdMediaUpdate,

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
				Description: "catalog name where upload the Media file",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "media name",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"media_path": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "absolute or relative path to Media file",
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

func resourceVcdMediaCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog media creation initiated")

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
	mediaName := d.Get("name").(string)
	task, err := catalog.UploadMediaImage(mediaName, d.Get("description").(string), d.Get("media_path").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes)
	if err != nil {
		log.Printf("Error uploading new catalog media: %#v", err)
		return fmt.Errorf("error uploading new catalog media: %#v", err)
	}

	terraformStdout := GetTerraformStdout()

	if d.Get("show_upload_progress").(bool) {
		for {
			if err := getError(task); err != nil {
				return err
			}

			_, _ = fmt.Fprint(terraformStdout, "vcd_catalog_media."+mediaName+": Upload progress "+task.GetUploadProgress()+"%\n")
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
			_, _ = fmt.Fprint(terraformStdout, "vcd_catalog_media."+mediaName+": vCD import catalog item progress "+progress+"%\n")
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

	d.SetId(catalogName + ":" + mediaName)

	log.Printf("[TRACE] Catalog media created: %#v", mediaName)

	err = createOrUpdateMediaItemMetadata(d, meta)
	if err != nil {
		return fmt.Errorf("error adding media item metadata: %s", err)
	}

	return resourceVcdMediaRead(d, meta)
}

func GetTerraformStdout() *os.File {
	var terraformStdout *os.File
	if v := flag.Lookup("test.v"); v == nil || v.Value.String() != "true" {
		terraformStdout = os.NewFile(uintptr(4), "stdout")
	} else {
		terraformStdout = os.Stdout
	}
	return terraformStdout
}

func resourceVcdMediaRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc("", "")
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	mediaItem, err := vdc.FindMediaImage(d.Get("name").(string))
	if err != nil || mediaItem == (govcd.MediaItem{}) {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return err
	}

	metadata, err := mediaItem.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item metadata: %s", err)
		return err
	}

	d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	return err
}

func resourceVcdMediaDelete(d *schema.ResourceData, meta interface{}) error {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

// currently updates only metadata
func resourceVcdMediaUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createOrUpdateMediaItemMetadata(d, meta)
	if err != nil {
		return fmt.Errorf("error updating media item metadata: %s", err)
	}
	return resourceVcdMediaRead(d, meta)
}

func createOrUpdateMediaItemMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata for media item")

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc("", "")
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	mediaItem, err := vdc.FindMediaImage(d.Get("name").(string))
	if err != nil || mediaItem == (govcd.MediaItem{}) {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return fmt.Errorf("unable to find media item: %s", err)
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
			err := mediaItem.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			_, err = mediaItem.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}
