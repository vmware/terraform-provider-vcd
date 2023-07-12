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
				Computed: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "catalog name where to upload the Media file",
				Deprecated:   "Use catalog_id instead",
				ExactlyOneOf: []string{"catalog", "catalog_id"},
			},
			"catalog_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "ID of the catalog where to upload the Media file",
				ExactlyOneOf: []string{"catalog", "catalog_id"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "media name",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"media_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "absolute or relative path to Media file",
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Default:     1,
				Description: "size of upload file piece size in mega bytes",
			},
			"show_upload_progress": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "shows upload progress in stdout",
			},
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Description:   "Key and value pairs for catalog item metadata",
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
			},
			"metadata_entry": metadataEntryResourceSchema("Catalog Media"),
			"is_iso": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is ISO",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is in a published catalog",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Media storage in Bytes",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media status",
			},
			"storage_profile_name": {
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

	var catalog *govcd.Catalog
	var err error
	catalogName := d.Get("catalog").(string)
	catalogId := d.Get("catalog_id").(string)
	if catalogId == "" {
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return diag.Errorf(errorRetrievingOrg, err)
		}
		catalog, err = adminOrg.GetCatalogByName(catalogName, false)
	} else {
		catalog, err = vcdClient.Client.GetCatalogById(catalogId)
	}
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

	err = createOrUpdateMediaItemMetadata(d, meta, "create")
	if err != nil {
		return diag.Errorf("error adding media item metadata: %s", err)
	}

	return genericVcdMediaRead(d, meta, "resource", "create")
}

func resourceVcdMediaRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdMediaRead(d, meta, "resource", "read")
}

func genericVcdMediaRead(d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	var catalog *govcd.Catalog
	var err error
	var orgName string
	catalogName := d.Get("catalog").(string)
	catalogId := d.Get("catalog_id").(string)

	if catalogId == "" {
		orgName, err = vcdClient.GetOrgNameFromResource(d)
		if err != nil {
			return diag.Errorf("error getting the Org name for vcd_catalog_media: %s", err)
		}
		catalog, err = vcdClient.Client.GetCatalogByName(orgName, catalogName)
	} else {
		catalog, err = vcdClient.Client.GetCatalogById(catalogId)
	}
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

	dSet(d, "catalog", catalog.Catalog.Name)
	dSet(d, "catalog_id", catalog.Catalog.ID)
	dSet(d, "name", media.Media.Name)
	dSet(d, "description", media.Media.Description)
	dSet(d, "is_iso", mediaRecord.MediaRecord.IsIso)
	dSet(d, "owner_name", mediaRecord.MediaRecord.OwnerName)
	dSet(d, "is_published", mediaRecord.MediaRecord.IsPublished)
	dSet(d, "creation_date", mediaRecord.MediaRecord.CreationDate)
	dSet(d, "size", mediaRecord.MediaRecord.StorageB)
	dSet(d, "status", mediaRecord.MediaRecord.Status)
	dSet(d, "storage_profile_name", mediaRecord.MediaRecord.StorageProfileName)

	diagErr := updateMetadataInState(d, vcdClient, "vcd_catalog_media", media)
	if diagErr != nil {
		log.Printf("[DEBUG] Unable to update media item metadata: %s", err)
		return diagErr
	}
	return nil
}

func resourceVcdMediaDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

// currently updates only metadata
func resourceVcdMediaUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := createOrUpdateMediaItemMetadata(d, meta, "update")
	if err != nil {
		return diag.Errorf("error updating media item metadata: %s", err)
	}
	return genericVcdMediaRead(d, meta, "resource", "update")
}

func createOrUpdateMediaItemMetadata(d *schema.ResourceData, meta interface{}, operation string) error {

	log.Printf("[TRACE] %s metadata for media item", operation)

	vcdClient := meta.(*VCDClient)

	var catalog *govcd.Catalog
	var err error
	var orgName string
	catalogName := d.Get("catalog").(string)
	catalogId := d.Get("catalog_id").(string)
	if catalogId == "" {
		orgName, err = vcdClient.GetOrgNameFromResource(d)
		if err != nil {
			return fmt.Errorf("error retrieving Org name for vcd_catalog_media: %s", err)
		}
		catalog, err = vcdClient.Client.GetCatalogByName(orgName, catalogName)
	} else {
		catalog, err = vcdClient.Client.GetCatalogById(catalogId)
	}
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return fmt.Errorf("unable to find catalog: %s", err)
	}

	media, err := catalog.GetMediaByName(d.Get("name").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return fmt.Errorf("unable to find media item: %s", err)
	}

	return createOrUpdateMetadata(d, vcdClient, media, "metadata", operation)
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
func resourceVcdCatalogMediaImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	catalog, err := vcdClient.Client.GetCatalogByName(orgName, catalogName)
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
