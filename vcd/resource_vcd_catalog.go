package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdCatalogCreate,
		Delete: resourceVcdCatalogDelete,
		Read:   resourceVcdCatalogRead,
		Update: resourceVcdCatalogUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceVcdCatalogImport,
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
				ForceNew: true,
			},
			"storage_profile": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional storage profile name",
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
		},
	}
}

func getStorageProfileReferenceByName(adminOrg *govcd.AdminOrg, storageProfileName string) (*types.Reference, error) {
	var storageProfileFound bool
	var storageProfileReference *types.Reference

	allOrgStorageProfiles, err := adminOrg.GetAllStorageProfileReferences(false)
	if err != nil {
		return nil, fmt.Errorf("error looking up storage profiles: %s", err)
	}

	for _, orgStorageProfile := range allOrgStorageProfiles {
		if orgStorageProfile.Name == storageProfileName {
			storageProfileFound = true
			storageProfileReference = orgStorageProfile
			break
		}
	}

	if !storageProfileFound {
		return nil, fmt.Errorf("could not find storage profile with name '%s' in Org '%s'",
			storageProfileName, adminOrg.AdminOrg.Name)
	}

	return storageProfileReference, nil
}

func resourceVcdCatalogCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	var storageProfiles *types.CatalogStorageProfiles
	storageProfileName := d.Get("storage_profile").(string)

	if storageProfileName != "" {
		storageProfileReference, err := getStorageProfileReferenceByName(adminOrg, storageProfileName)
		if err != nil {
			return fmt.Errorf("couuld not proces storage profile '%s': %s", storageProfileName, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	catalog, err := adminOrg.CreateCatalogWithStorageProfile(name, description, storageProfiles)
	if err != nil {
		log.Printf("[TRACE] Error creating Catalog: %#v", err)
		return fmt.Errorf("error creating Catalog: %#v", err)
	}

	d.SetId(catalog.AdminCatalog.ID)
	log.Printf("[TRACE] Catalog created: %#v", catalog)
	return resourceVcdCatalogRead(d, meta)
}

func resourceVcdCatalogRead(d *schema.ResourceData, meta interface{}) error {
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
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return fmt.Errorf("error retrieving storage profile reference: %s", err)
		}
		_ = d.Set("storage_profile", storageProfileReference.Name)
	} else {
		// In case no storage profile are defined in API call
		_ = d.Set("storage_profile", "")
	}

	_ = d.Set("description", adminCatalog.AdminCatalog.Description)
	_ = d.Set("created", adminCatalog.AdminCatalog.DateCreated)
	d.SetId(adminCatalog.AdminCatalog.ID)
	log.Printf("[TRACE] Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdCatalogUpdate does not require actions for  fields "delete_force", "delete_recursive", but does allow to
// change `storage_profile`
func resourceVcdCatalogUpdate(d *schema.ResourceData, meta interface{}) error {
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

	// Perform storage profile updates
	if d.HasChange("storage_profile") {
		storageProfileName := d.Get("storage_profile").(string)

		// Unset storage profile name (use any)
		if storageProfileName == "" {
			// Set empty structure as `nil` would not update it at all
			adminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{}}
		}

		if storageProfileName != "" {
			storageProfileReference, err := getStorageProfileReferenceByName(adminOrg, storageProfileName)
			if err != nil {
				return fmt.Errorf("couuld not proces storage profile '%s': %s", storageProfileName, err)
			}
			adminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
		}
	}

	err = adminCatalog.Update()
	if err != nil {
		return fmt.Errorf("error updating catalog '%s': %s", adminCatalog.AdminCatalog.Name, err)
	}

	return resourceVcdCatalogRead(d, meta)
}

func resourceVcdCatalogDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog delete started")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
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
		return fmt.Errorf("error removing catalog %#v", err)
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
func resourceVcdCatalogImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	d.SetId(catalog.Catalog.ID)

	// Fill in other fields
	err = resourceVcdCatalogRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
