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

func resourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogItemCreate,
		DeleteContext: resourceVcdCatalogItemDelete,
		ReadContext:   resourceVcdCatalogItemRead,
		UpdateContext: resourceVcdCatalogItemUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogItemImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog name where upload the OVA file",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "catalog item name",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the item was created",
			},
			"ova_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"ova_path", "ovf_url"},
				Description:  "Absolute or relative path to OVA",
			},
			"ovf_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"ova_path", "ovf_url"},
				Description:  "URL of OVF file",
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "size of upload file piece size in mega bytes",
			},
			"show_upload_progress": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "shows upload progress in stdout",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for the metadata of the vApp template associated to this catalog item",
			},
			"catalog_item_metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for catalog item metadata",
			},
		},
	}
}

func resourceVcdCatalogItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog item creation initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	var diagError diag.Diagnostics
	itemName := d.Get("name").(string)
	if d.Get("ova_path").(string) != "" {
		diagError = uploadFile(d, catalog, itemName)
	} else if d.Get("ovf_url").(string) != "" {
		diagError = uploadFromUrl(d, catalog, itemName)
	} else {
		return diag.Errorf("`ova_path` or `ovf_url` value is missing %s", err)
	}
	if diagError != nil {
		return diagError
	}

	item, err := catalog.GetCatalogItemByName(itemName, true)
	if err != nil {
		return diag.Errorf("error retrieving catalog item %s: %s", itemName, err)
	}
	d.SetId(item.CatalogItem.ID)

	log.Printf("[TRACE] Catalog item created: %s", itemName)

	err = createOrUpdateCatalogItemMetadata(d, meta)
	if diagError != nil {
		return diag.FromErr(err)
	}

	return resourceVcdCatalogItemRead(ctx, d, meta)
}

func uploadFile(d *schema.ResourceData, catalog *govcd.Catalog, itemName string) diag.Diagnostics {
	uploadPieceSize := d.Get("upload_piece_size").(int)
	task, err := catalog.UploadOvf(d.Get("ova_path").(string), itemName, d.Get("description").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes
	if err != nil {
		log.Printf("[DEBUG] Error uploading new catalog item: %s", err)
		return diag.Errorf("error uploading new catalog item: %s", err)
	}

	if d.Get("show_upload_progress").(bool) {
		for {
			if err := getError(task); err != nil {
				return diag.FromErr(err)
			}
			logForScreen("vcd_catalog_item", fmt.Sprintf("vcd_catalog_item."+itemName+": Upload progress "+task.GetUploadProgress()+"%%\n"))
			if task.GetUploadProgress() == "100.00" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	return finishHandlingTask(d, *task.Task, itemName)
}

func uploadFromUrl(d *schema.ResourceData, catalog *govcd.Catalog, itemName string) diag.Diagnostics {
	task, err := catalog.UploadOvfByLink(d.Get("ovf_url").(string), itemName, d.Get("description").(string))
	if err != nil {
		log.Printf("[DEBUG] Error uploading new catalog item from URL: %s", err)
		return diag.Errorf("error uploading new catalog item from URL: %s", err)
	}

	return finishHandlingTask(d, task, itemName)
}

func finishHandlingTask(d *schema.ResourceData, task govcd.Task, itemName string) diag.Diagnostics {
	if d.Get("show_upload_progress").(bool) {
		for {
			progress, err := task.GetTaskProgress()
			if err != nil {
				log.Printf("VCD Error importing new catalog item: %s", err)
				return diag.Errorf("VCD Error importing new catalog item: %s", err)
			}
			logForScreen("vcd_catalog_item", fmt.Sprintf("vcd_catalog_item."+itemName+": VCD import catalog item progress "+progress+"%%\n"))
			if progress == "100" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	err := task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error waiting for task to complete: %+v", err)
	}
	return nil
}

func resourceVcdCatalogItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogItemRead(d, meta, "resource")
}

func genericVcdCatalogItemRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	catalogItem, err := findCatalogItem(d, meta.(*VCDClient), origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return diag.Errorf("Unable to find media item: %s", err)
	}
	if catalogItem == nil {
		log.Printf("[DEBUG] Unable to find media item: %s. Removing from tfstate", err)
		return diag.Errorf("Unable to find media item")
	}

	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return diag.Errorf("Unable to find Vapp template: %s", err)
	}

	vAppTemplateMetadata, err := vAppTemplate.GetMetadata()
	if err != nil {
		return diag.Errorf("Unable to find catalog item's associated vApp template metadata: %s", err)
	}
	catalogItemMetadata, err := catalogItem.GetMetadata()
	if err != nil {
		return diag.Errorf("Unable to find metadata for the catalog item: %s", err)
	}

	dSet(d, "name", catalogItem.CatalogItem.Name)
	dSet(d, "created", vAppTemplate.VAppTemplate.DateCreated)
	dSet(d, "description", catalogItem.CatalogItem.Description)
	err = d.Set("metadata", getMetadataStruct(vAppTemplateMetadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set metadata for the catalog item's associated vApp template: %s", err)
	}
	err = d.Set("catalog_item_metadata", getMetadataStruct(catalogItemMetadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set metadata for the catalog item: %s", err)
	}

	return nil
}

func resourceVcdCatalogItemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

func resourceVcdCatalogItemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("description") || d.HasChange("name") {
		catalogItem, err := findCatalogItem(d, meta.(*VCDClient), "resource")
		if err != nil {
			log.Printf("[DEBUG] Unable to find media item: %s", err)
			return diag.Errorf("Unable to find media item: %s", err)
		}
		if catalogItem == nil {
			log.Printf("[DEBUG] Unable to find media item: %s. Removing from tfstate", err)
			return diag.Errorf("Unable to find media item")
		}

		vAppTemplate, err := catalogItem.GetVAppTemplate()
		if err != nil {
			return diag.Errorf("Unable to find Vapp template: %s", err)
		}

		vAppTemplate.VAppTemplate.Description = d.Get("description").(string)
		vAppTemplate.VAppTemplate.Name = d.Get("name").(string)
		_, err = vAppTemplate.Update()
		if err != nil {
			return diag.Errorf("error updating catalog item: %s", err)
		}
	}

	err := createOrUpdateCatalogItemMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error updating catalog item metadata: %s", err)
	}
	return nil
}

func createOrUpdateCatalogItemMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata for catalog item")

	catalogItem, err := findCatalogItem(d, meta.(*VCDClient), "resource")
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item: %s", err)
		return fmt.Errorf("%s", err)
	}

	// We have to add metadata to template to see in UI
	// catalog item is another abstraction and has own metadata which we don't see in UI
	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return err
	}

	err = createOrUpdateMetadata(d, &vAppTemplate, "metadata")
	if err != nil {
		return err
	}

	return createOrUpdateMetadata(d, catalogItem, "catalog_item_metadata")
}

// Imports a CatalogItem into Terraform state
// This function task is to get the data from VCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name.catalog_item_name
//
// Example import path (id): org_name.catalog_name.catalog_item_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogItemImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	dSet(d, "org", orgName)
	dSet(d, "catalog", catalogName)
	dSet(d, "name", catalogItemName)
	dSet(d, "description", catalogItem.CatalogItem.Description)
	d.SetId(catalogItem.CatalogItem.ID)

	return []*schema.ResourceData{d}, nil
}
