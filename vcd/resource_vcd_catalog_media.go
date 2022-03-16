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

func resourceVcdCatalogMedia() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdMediaCreate,
		DeleteContext: resourceVcdMediaDelete,
		ReadContext:   resourceVcdMediaRead,
		UpdateContext: resourceVcdMediaUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogMediaImport,
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
			"is_iso": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is ISO",
			},
			"owner_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name",
			},
			"is_published": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is in a published catalog",
			},
			"creation_date": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date",
			},
			"size": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Media storage in Bytes",
			},
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media status",
			},
			"storage_profile_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile name",
			},
		},
	}
}

func resourceVcdMediaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog media creation initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		log.Printf("Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	uploadPieceSize := d.Get("upload_piece_size").(int)
	mediaName := d.Get("name").(string)
	task, err := catalog.UploadMediaImage(mediaName, d.Get("description").(string), d.Get("media_path").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes)
	if err != nil {
		log.Printf("Error uploading new catalog media: %s", err)
		return diag.Errorf("error uploading new catalog media: %s", err)
	}

	if d.Get("show_upload_progress").(bool) {
		for {
			if err := getError(task); err != nil {
				return diag.FromErr(err)
			}

			logForScreen("vcd_catalog_media", fmt.Sprintf("vcd_catalog_media."+mediaName+": Upload progress "+task.GetUploadProgress()+"%%\n"))
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
				log.Printf("vCD Error importing new catalog item: %s", err)
				return diag.Errorf("vCD Error importing new catalog item: %s", err)
			}
			logForScreen("vcd_catalog_media", fmt.Sprintf("vcd_catalog_media."+mediaName+": vCD import catalog item progress "+progress+"%%\n"))
			if progress == "100" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error waiting for task to complete: %+v", err)
	}

	log.Printf("[TRACE] Catalog media created: %#v", mediaName)

	err = createOrUpdateMediaItemMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error adding media item metadata: %s", err)
	}

	return resourceVcdMediaRead(ctx, d, meta)
}

func resourceVcdMediaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdMediaRead(d, meta, "resource")
}

func genericVcdMediaRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := org.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return diag.Errorf("unable to find catalog: %s", err)
	}

	var media *govcd.Media

	if origin == "datasource" {
		if !nameOrFilterIsSet(d) {
			return diag.Errorf(noNameOrFilterError, "vcd_catalog_media")
		}
		filter, hasFilter := d.GetOk("filter")
		if hasFilter {

			media, err = getMediaByFilter(catalog, filter, vcdClient.Client.IsSysAdmin)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	identifier := d.Id()
	if media == nil {
		if identifier == "" {
			identifier = d.Get("name").(string)
		}
		if identifier == "" {
			return diag.Errorf("media identifier is empty")
		}
		media, err = catalog.GetMediaByNameOrId(identifier, false)
	}
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find media with ID %s: %s. Removing from state", identifier, err)
		d.SetId("")
		return nil
	}
	if err != nil {
		log.Printf("[DEBUG] Unable to find media: %s", err)
		return diag.FromErr(err)
	}

	d.SetId(media.Media.ID)

	mediaRecord, err := catalog.QueryMedia(media.Media.Name)
	if err != nil {
		log.Printf("[DEBUG] Unable to query media: %s", err)
		return diag.FromErr(err)
	}

	dSet(d, "name", media.Media.Name)
	dSet(d, "description", media.Media.Description)
	dSet(d, "is_iso", mediaRecord.MediaRecord.IsIso)
	dSet(d, "owner_name", mediaRecord.MediaRecord.OwnerName)
	dSet(d, "is_published", mediaRecord.MediaRecord.IsPublished)
	dSet(d, "creation_date", mediaRecord.MediaRecord.CreationDate)
	dSet(d, "size", mediaRecord.MediaRecord.StorageB)
	dSet(d, "status", mediaRecord.MediaRecord.Status)
	dSet(d, "storage_profile_name", mediaRecord.MediaRecord.StorageProfileName)

	metadata, err := media.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item metadata: %s", err)
		return diag.FromErr(err)
	}

	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdMediaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

// currently updates only metadata
func resourceVcdMediaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := createOrUpdateMediaItemMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error updating media item metadata: %s", err)
	}
	return resourceVcdMediaRead(ctx, d, meta)
}

func createOrUpdateMediaItemMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata for media item")

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := org.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return fmt.Errorf("unable to find catalog: %s", err)
	}

	media, err := catalog.GetMediaByName(d.Get("name").(string), false)
	if err != nil {
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
			err := media.DeleteMetadataEntry(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			err = media.AddMetadataEntry(types.MetadataStringValue, k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}

// resourceVcdCatalogMediaImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_catalog_media.my-media
// Example import path (_the_id_string_): org.catalog.my-media-name
func resourceVcdCatalogMediaImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog.my-media-name")
	}
	orgName, catalogName, mediaName := resourceURI[0], resourceURI[1], resourceURI[2]

	if orgName == "" {
		return nil, fmt.Errorf("import: empty org name provided")
	}
	if catalogName == "" {
		return nil, fmt.Errorf("import: empty catalog name provided")
	}
	if mediaName == "" {
		return nil, fmt.Errorf("import: empty media item name provided")
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

	media, err := catalog.GetMediaByName(mediaName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "catalog", catalogName)
	dSet(d, "name", mediaName)
	dSet(d, "description", media.Media.Description)
	d.SetId(media.Media.ID)

	return []*schema.ResourceData{d}, nil
}
