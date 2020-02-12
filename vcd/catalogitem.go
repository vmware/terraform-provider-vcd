package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// Deletes catalog item which can be vapp template OVA or media ISO file
func deleteCatalogItem(d *schema.ResourceData, vcdClient *VCDClient) error {
	log.Printf("[TRACE] Catalog item delete started")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		return fmt.Errorf("unable to find catalog")
	}

	catalogItemName := d.Get("name").(string)
	catalogItem, err := catalog.GetCatalogItemByName(catalogItemName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		return fmt.Errorf("unable to find catalog item %s", catalogItemName)
	}

	err = catalogItem.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing catalog item %s", err)
		return fmt.Errorf("error removing catalog item %s", err)
	}

	_, err = catalog.GetCatalogItemByName(catalogItemName, true)
	if err == nil {
		return fmt.Errorf("catalog item %s still found after deletion", catalogItemName)
	}
	log.Printf("[TRACE] Catalog item delete completed: %s", catalogItemName)

	return nil
}

// Finds catalog item which can be vapp template OVA or media ISO file
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

	catalogItem, err := catalog.GetCatalogItemByNameOrId(identifier, false)
	if err != nil {
		errMessage := fmt.Errorf("unable to find catalog item with ID %s: %s", d.Id(), err)
		if govcd.IsNotFound(err) {
			log.Printf("[INFO] %s", errMessage)
			d.SetId("")
			return nil, errMessage
		}
		return nil, errMessage
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find catalog item %s: %s", identifier, err)
	}

	d.SetId(catalogItem.CatalogItem.ID)
	log.Printf("[TRACE] Catalog item read completed: %#v", catalogItem.CatalogItem)
	return catalogItem, nil
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
