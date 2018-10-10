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

		//resource "vcd_catalog" "OperatingSystems" {
		//org = "Solpan" # Optional, if defined at provider level
		//vdc = "SolpanVDC" # Optional, if defined at provider level
		//
		//name = "OperatingSystems"
		//description = "Fresh OS templates"
		//
		//force      = "true"
		//recursive  = "true"
		//}
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"vdc": {
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
				Required: true,
				ForceNew: true,
			},
			"force": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"recursive": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
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

	catalog, err := adminOrg.CreateCatalog(d.Get("name").(string), d.Get("description").(string), false)
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

	err = adminCatalog.Delete(d.Get("force").(bool), d.Get("recursive").(bool))
	if err != nil {
		log.Printf("Error removing catalog %#v", err)
		return fmt.Errorf("error removing catalog %#v", err)
	}

	log.Printf("[TRACE] Catalog delete completed: %#v", adminCatalog.AdminCatalog)
	return nil
}
