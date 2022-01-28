package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogCreate,
		DeleteContext: resourceVcdCatalogDelete,
		ReadContext:   resourceVcdCatalogRead,
		UpdateContext: resourceVcdCatalogUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional storage profile ID",
			},
			"created": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},
			"delete_force": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state.",
			},
			"delete_recursive": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal.",
			},
			"publish_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation.",
			},
			"cache_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True enables early catalog export to optimize synchronization",
			},
			"preserve_identity_information": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "An optional password to access the catalog. Only ASCII characters are allowed in a valid password.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for catalog metadata",
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
		},
	}
}

func resourceVcdCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var storageProfiles *types.CatalogStorageProfiles
	storageProfileId := d.Get("storage_profile_id").(string)
	if storageProfileId != "" {
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return diag.Errorf("error looking up Storage Profile '%s' reference: %s", storageProfileId, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	catalog, err := adminOrg.CreateCatalogWithStorageProfile(name, description, storageProfiles)
	if err != nil {
		log.Printf("[TRACE] Error creating Catalog: %#v", err)
		return diag.Errorf("error creating Catalog: %#v", err)
	}

	d.SetId(catalog.AdminCatalog.ID)

	if d.Get("publish_enabled").(bool) {
		err = updatePublishToExternalOrgSettings(d, catalog)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = createOrUpdateAdminCatalogMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error adding media item metadata: %s", err)
	}

	log.Printf("[TRACE] Catalog created: %#v", catalog)
	return resourceVcdCatalogRead(ctx, d, meta)
}

func updatePublishToExternalOrgSettings(d *schema.ResourceData, adminCatalog *govcd.AdminCatalog) error {
	err := adminCatalog.PublishToExternalOrganizations(types.PublishExternalCatalogParams{
		IsPublishedExternally:    takeBoolPointer(d.Get("publish_enabled").(bool)),
		IsCachedEnabled:          takeBoolPointer(d.Get("cache_enabled").(bool)),
		PreserveIdentityInfoFlag: takeBoolPointer(d.Get("preserve_identity_information").(bool)),
		Password:                 d.Get("password").(string),
	})
	if err != nil {
		return fmt.Errorf("[updatePublishToExternalOrgSettings] error: %s", err)
	}
	return nil
}

func resourceVcdCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := genericResourceVcdCatalogRead(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func genericResourceVcdCatalogRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return fmt.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	// Check if storage profile is set. Although storage profile structure accepts a list, in UI only one can be picked
	if adminCatalog.AdminCatalog.CatalogStorageProfiles != nil && len(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile) > 0 {
		// By default API does not return Storage Profile Name in response. It has ID and HREF, but not Name so name
		// must be looked up
		storageProfileId := adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile[0].ID
		dSet(d, "storage_profile_id", storageProfileId)
	} else {
		// In case no storage profile are defined in API call
		dSet(d, "storage_profile_id", "")
	}

	dSet(d, "description", adminCatalog.AdminCatalog.Description)
	dSet(d, "created", adminCatalog.AdminCatalog.DateCreated)
	if adminCatalog.AdminCatalog.PublishExternalCatalogParams != nil {
		dSet(d, "publish_enabled", adminCatalog.AdminCatalog.PublishExternalCatalogParams.IsPublishedExternally)
		dSet(d, "cache_enabled", adminCatalog.AdminCatalog.PublishExternalCatalogParams.IsCachedEnabled)
		dSet(d, "preserve_identity_information", adminCatalog.AdminCatalog.PublishExternalCatalogParams.PreserveIdentityInfoFlag)
	} else {
		dSet(d, "publish_enabled", false)
		dSet(d, "cache_enabled", false)
		dSet(d, "preserve_identity_information", false)
		dSet(d, "password", "")
	}

	metadata, err := adminCatalog.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find media item metadata: %s", err)
		return err
	}

	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return err
	}

	d.SetId(adminCatalog.AdminCatalog.ID)
	log.Printf("[TRACE] Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdCatalogUpdate does not require actions for  fields "delete_force", "delete_recursive",
// but does allow changing `storage_profile`
func resourceVcdCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return diag.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	// Create a copy of adminCatalog to only set and change things which are related to this update section and skip the
	// other fields. This is important as this provider does not cover all settings available in API and they should not be
	// overwritten.
	newAdminCatalog := govcd.NewAdminCatalogWithParent(&vcdClient.VCDClient.Client, adminOrg)
	newAdminCatalog.AdminCatalog.ID = adminCatalog.AdminCatalog.ID
	newAdminCatalog.AdminCatalog.HREF = adminCatalog.AdminCatalog.HREF
	newAdminCatalog.AdminCatalog.Name = adminCatalog.AdminCatalog.Name

	// Perform storage profile updates
	if d.HasChange("storage_profile_id") {
		storageProfileId := d.Get("storage_profile_id").(string)

		// Unset storage profile (use any available in Org)
		if storageProfileId == "" {
			// Set empty structure as `nil` would not update it at all
			newAdminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{}}
		}

		if storageProfileId != "" {
			storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
			if err != nil {
				return diag.Errorf("could not process Storage Profile '%s': %s", storageProfileId, err)
			}
			newAdminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
		}
	}

	if d.HasChange("description") {
		newAdminCatalog.AdminCatalog.Description = d.Get("description").(string)
	}

	err = newAdminCatalog.Update()
	if err != nil {
		return diag.Errorf("error updating catalog '%s': %s", adminCatalog.AdminCatalog.Name, err)
	}

	if d.HasChanges("publish_enabled", "cache_enabled", "preserve_identity_information", "password") {
		err = updatePublishToExternalOrgSettings(d, newAdminCatalog)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = createOrUpdateAdminCatalogMetadata(d, meta)
	if err != nil {
		return diag.Errorf("error updating media item metadata: %s", err)
	}

	return resourceVcdCatalogRead(ctx, d, meta)
}

func resourceVcdCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog delete started")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

	err = adminCatalog.Delete(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		log.Printf("[DEBUG] Error removing catalog %#v", err)
		return diag.Errorf("error removing catalog %#v", err)
	}

	log.Printf("[TRACE] Catalog delete completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdCatalogImport imports a Catalog into Terraform state
// This function task is to get the data from vCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name
//
// Example import path (id): org_name.catalog_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog")
	}
	orgName, catalogName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "name", catalogName)
	dSet(d, "description", catalog.Catalog.Description)
	d.SetId(catalog.Catalog.ID)

	// Fill in other fields
	err = genericResourceVcdCatalogRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func createOrUpdateAdminCatalogMetadata(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[TRACE] adding/updating metadata for media item")

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := org.GetAdminCatalogByName(d.Get("name").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return fmt.Errorf("unable to find catalog: %s", err)
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
			err := catalog.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			_, err = catalog.AddMetadata(types.MetadataStringValue, k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}
