package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// Deletes catalog item which can be vApp template OVA or media ISO file
func deleteCatalogItem(d *schema.ResourceData, vcdClient *VCDClient) diag.Diagnostics {
	log.Printf("[TRACE] Catalog item delete started")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		return diag.Errorf("unable to find catalog")
	}

	catalogItemName := d.Get("name").(string)
	catalogItem, err := catalog.GetCatalogItemByName(catalogItemName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		return diag.Errorf("unable to find catalog item %s", catalogItemName)
	}

	err = catalogItem.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing catalog item %s", err)
		return diag.Errorf("error removing catalog item %s", err)
	}

	_, err = catalog.GetCatalogItemByName(catalogItemName, true)
	if err == nil {
		return diag.Errorf("catalog item %s still found after deletion", catalogItemName)
	}
	log.Printf("[TRACE] Catalog item delete completed: %s", catalogItemName)

	return nil
}

// Finds catalog item which can be vApp template OVA or media ISO file
// Deprecated: This function is deprecated as described in Issue #502. Use findVappTemplate instead.
func findCatalogItem(d *schema.ResourceData, vcdClient *VCDClient, origin string) (*govcd.CatalogItem, error) {
	log.Printf("[TRACE] Catalog item read initiated")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return nil, fmt.Errorf("unable to find catalog: %s", err)
	}

	identifier := d.Id()

	// Check if identifier is still in deprecated style `catalogName:mediaName`
	// Required for backwards compatibility as identifier has been changed to vCD ID in 2.5.0
	if identifier == "" || strings.Count(identifier, ":") <= 1 {
		identifier = d.Get("name").(string)
	}

	var catalogItem *govcd.CatalogItem
	if origin == "datasource" {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_catalog_item")
		}
		filter, hasFilter := d.GetOk("filter")
		if hasFilter {

			catalogItem, err = getCatalogItemByFilter(catalog, filter, vcdClient.Client.IsSysAdmin)
			if err != nil {
				return nil, err
			}

			d.SetId(catalogItem.CatalogItem.ID)
			return catalogItem, nil
		}
	}
	// No filter: we continue with single item  GET

	catalogItem, err = catalog.GetCatalogItemByNameOrId(identifier, false)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] Unable to find catalog item %s. Removing from tfstate", identifier)
		d.SetId("")
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find catalog item %s: %s", identifier, err)
	}

	d.SetId(catalogItem.CatalogItem.ID)
	log.Printf("[TRACE] Catalog item read completed: %#v", catalogItem.CatalogItem)
	return catalogItem, nil
}

// Finds a vApp Template with the information given in the resource data. If it's a data source it uses a filtering
// mechanism, if it's a resource it just gets the information.
func findVappTemplate(d *schema.ResourceData, vcdClient *VCDClient, origin string) (*govcd.VAppTemplate, error) {
	log.Printf("[TRACE] vApp template search initiated")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog.")
		return nil, fmt.Errorf("unable to find catalog: %s", err)
	}

	identifier := d.Id()

	// Check if identifier is still in deprecated style `catalogName:mediaName`
	// Required for backwards compatibility as identifier has been changed to vCD ID in 2.5.0
	if identifier == "" || strings.Count(identifier, ":") <= 1 {
		identifier = d.Get("name").(string)
	}

	var vAppTemplate *govcd.VAppTemplate
	if origin == "datasource" {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_catalog_vapp_template")
		}
		filter, hasFilter := d.GetOk("filter")
		if hasFilter {

			vAppTemplate, err = getVappTemplateByFilter(catalog, filter, vcdClient.Client.IsSysAdmin)
			if err != nil {
				return nil, err
			}

			d.SetId(vAppTemplate.VAppTemplate.ID)
			return vAppTemplate, nil
		}
	}
	// No filter: we continue with single item  GET

	vAppTemplate, err = catalog.GetVAppTemplateByNameOrId(identifier, false)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] Unable to find vApp Template %s. Removing from tfstate", identifier)
		d.SetId("")
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find vApp Template %s: %s", identifier, err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)
	log.Printf("[TRACE] Catalog item read completed: %#v", vAppTemplate.VAppTemplate)
	return vAppTemplate, nil
}

func getError(task govcd.UploadTask) error {
	if task.GetUploadError() != nil {
		err := task.CancelTask()
		if err != nil {
			log.Printf("[DEBUG] error cancelling media upload task: %#v", err)
		}
		return fmt.Errorf("error uploading media: %#v", task.GetUploadError())
	}
	return nil
}
