package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
	if err != nil || catalog == nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return fmt.Errorf("unable to find catalog")
	}

	catalogItem, err := catalog.GetCatalogItemByName(d.Get("name").(string), false)
	if err != nil || catalogItem == nil {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		d.SetId("")
		return fmt.Errorf("unable to find catalog item")
	}

	err = catalogItem.Delete()
	if err != nil {
		log.Printf("Error removing catalog item %s", err)
		return fmt.Errorf("error removing catalog item %s", err)
	}

	catalogItem, err = catalog.GetCatalogItemByName(d.Get("name").(string), true)
	if catalogItem != nil || err == nil {
		return fmt.Errorf("catalog item %s still found after deletion", d.Get("name").(string))
	}
	log.Printf("[TRACE] Catalog item delete completed: %s", d.Get("name").(string))

	return nil
}

// Finds catalog item which can be vapp template OVA or media ISO file
func findCatalogItem(d *schema.ResourceData, vcdClient *VCDClient) (*govcd.CatalogItem, error) {
	log.Printf("[TRACE] Catalog item read initiated")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil || catalog == nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil, fmt.Errorf("unable to find catalog")
	}

	catalogItem, err := catalog.GetCatalogItemByName(d.Get("name").(string), false)
	if err != nil || catalogItem == nil {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		d.SetId("")
		return nil, fmt.Errorf("unable to find catalog item: %s", err)
	}

	d.SetId(catalogItem.CatalogItem.ID)
	log.Printf("[TRACE] Catalog item read completed: %#v", catalogItem.CatalogItem)
	return catalogItem, nil
}

func getError(task govcd.UploadTask) error {
	if task.GetUploadError() != nil {
		err := task.CancelTask()
		if err != nil {
			log.Printf("error cancelling media upload task: %#v", err)
		}
		return fmt.Errorf("error uploading media: %#v", task.GetUploadError())
	}
	return nil
}
