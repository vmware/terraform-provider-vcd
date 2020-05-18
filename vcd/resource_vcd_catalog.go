package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func resourceVcdCatalogCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.CreateCatalog(d.Get("name").(string), d.Get("description").(string))
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

	catalog, err := adminOrg.GetCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return fmt.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	_ = d.Set("description", catalog.Catalog.Description)
	_ = d.Set("created", catalog.Catalog.DateCreated)
	d.SetId(catalog.Catalog.ID)
	log.Printf("[TRACE] Catalog read completed: %#v", catalog.Catalog)
	return nil
}

//update function for "delete_force", "delete_recursive" no actions needed
func resourceVcdCatalogUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
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

// Imports a Catalog into Terraform state
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

	_ = d.Set("org", orgName)
	_ = d.Set("name", catalogName)
	_ = d.Set("description", catalog.Catalog.Description)
	d.SetId(catalog.Catalog.ID)

	return []*schema.ResourceData{d}, nil
}
