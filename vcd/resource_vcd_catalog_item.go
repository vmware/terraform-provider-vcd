package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"

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
			"metadata_entry": metadataEntryResourceSchema("Catalog Item"),
			"catalog_item_metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Description:   "Key and value pairs for catalog item metadata",
				Deprecated:    "Use metadata_entry instead",
				Computed:      true, // To be compatible with `metadata_entry`
				ConflictsWith: []string{"metadata_entry"},
			},
		},
	}
}

func resourceVcdCatalogItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog item creation initiated")

	vcdClient := meta.(*VCDClient)

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return diag.Errorf("error getting org name for vcd_catalog_item: %s", err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := vcdClient.Client.GetCatalogByName(orgName, catalogName)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	var diagError diag.Diagnostics
	itemName := d.Get("name").(string)
	if d.Get("ova_path").(string) != "" {
		diagError = uploadOvaFromFilePath(d, catalog, itemName, "vcd_catalog_item")
	} else if d.Get("ovf_url").(string) != "" {
		diagError = uploadFromUrl(d, catalog, itemName, "vcd_catalog_item")
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

func resourceVcdCatalogItemRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogItemRead(d, meta, "resource")
}

func genericVcdCatalogItemRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	catalogItem, err := findCatalogItem(d, vcdClient, origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog item: %s", err)
		return diag.Errorf("Unable to find catalog item: %s", err)
	}
	if catalogItem == nil {
		log.Printf("[DEBUG] Unable to find catalog item: %s. Removing from tfstate", err)
		return diag.Errorf("Unable to find catalog item")
	}

	vAppTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return diag.Errorf("Unable to find Vapp template: %s", err)
	}

	dSet(d, "name", catalogItem.CatalogItem.Name)
	dSet(d, "created", vAppTemplate.VAppTemplate.DateCreated)
	dSet(d, "description", catalogItem.CatalogItem.Description)

	// Catalog item metadata:
	// We can't use updateMetadataInState(d, catalogItem) because the attribute name is different.

	// We temporarily remove the ignored metadata filter to retrieve the deprecated metadata and vApp Template metadata contents,
	// which should not be affected by it.
	// This closure makes the unlocking more optimal, as it unlocks when the closure returns.
	getAllMetadata := func() (*types.Metadata, *types.Metadata, error) {
		vcdMutexKV.kvLock("metadata") // The lock is needed as we're modifying shared client internals
		defer vcdMutexKV.kvUnlock("metadata")
		ignoredMetadata := vcdClient.VCDClient.SetMetadataToIgnore(nil)
		deprecatedCatalogItemMetadata, err1 := catalogItem.GetMetadata()
		vAppTemplateMetadata, err2 := vAppTemplate.GetMetadata()
		vcdClient.VCDClient.SetMetadataToIgnore(ignoredMetadata)
		if err1 != nil {
			return nil, nil, fmt.Errorf("error getting metadata to save in state: %s", err)
		}
		if err2 != nil {
			return nil, nil, fmt.Errorf("unable to find catalog item's associated vApp template metadata: %s", err)
		}
		return deprecatedCatalogItemMetadata, vAppTemplateMetadata, nil
	}
	deprecatedCatalogItemMetadata, vAppTemplateMetadata, err := getAllMetadata()
	if err != nil {
		return diag.FromErr(err)
	}

	// Set deprecated metadata attribute of catalog item, just for compatibility reasons
	err = d.Set("catalog_item_metadata", getMetadataStruct(deprecatedCatalogItemMetadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set catalog item's metadata: %s", err)
	}
	// Set vApp Template metadata
	err = d.Set("metadata", getMetadataStruct(vAppTemplateMetadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set metadata for the catalog item's associated vApp template: %s", err)
	}

	// We finally enter the `metadata_entry` metadata with filtering enabled
	diagErr := checkIgnoredMetadataConflicts(d, vcdClient, "vcd_catalog_item")
	if diagErr != nil {
		return diagErr
	}

	metadata, err := catalogItem.GetMetadata()
	if err != nil {
		return diag.Errorf("Unable to find catalog item's metadata: %s", err)
	}

	err = setMetadataEntryInState(d, metadata.MetadataEntry)
	if err != nil {
		return diag.Errorf("Unable to set catalog item's metadata entries: %s", err)
	}

	return nil
}

func resourceVcdCatalogItemDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteCatalogItem(d, meta.(*VCDClient))
}

func resourceVcdCatalogItemUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	err = createOrUpdateMetadata(d, catalogItem, "catalog_item_metadata")
	if err != nil {
		return err
	}

	// vApp Template metadata
	if d.HasChange("metadata") && !d.HasChange("metadata_entry") {
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
			err := vAppTemplate.DeleteMetadataEntry(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		if len(newMetadata) > 0 {
			err := vAppTemplate.MergeMetadata(types.MetadataStringValue, newMetadata)
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}

	return nil
}

// Imports a CatalogItem into Terraform state
// This function task is to get the data from VCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name.catalog_item_name
//
// Example import path (id): org_name.catalog_name.catalog_item_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogItemImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	catalog, err := vcdClient.Client.GetCatalogByName(orgName, catalogName)
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
