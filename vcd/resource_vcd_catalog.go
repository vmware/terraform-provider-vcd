package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
	"log"
)

func resourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdCatalogCreate,
		Delete: resourceVcdCatalogDelete,
		Read:   resourceVcdCatalogRead,
		Update: resourceVcdCatalogUpdate,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
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
		log.Printf("Error creating Catalog: %#v", err)
		return fmt.Errorf("error creating Catalog: %#v", err)
	}

	d.SetId(d.Get("name").(string))
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

	catalog, err := adminOrg.FindCatalog(d.Id())
	if err != nil || catalog == (govcd.Catalog{}) {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

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

	adminCatalog, err := adminOrg.FindAdminCatalog(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

	err = adminCatalog.Delete(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		log.Printf("Error removing catalog %#v", err)
		return fmt.Errorf("error removing catalog %#v", err)
	}

	log.Printf("[TRACE] Catalog delete completed: %#v", adminCatalog.AdminCatalog)
	return nil
}
