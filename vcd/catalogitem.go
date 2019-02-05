package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
)

// Deletes catalog item which can be vapp template OVA or media ISO file
func deleteCatalogItem(d *schema.ResourceData, vcdClient *VCDClient) error {
	log.Printf("[TRACE] Catalog item delete started")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.FindCatalog(d.Get("catalog").(string))
	if err != nil || catalog == (govcd.Catalog{}) {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

	catalogItem, err := catalog.FindCatalogItem(d.Get("name").(string))
	if err != nil || catalogItem == (govcd.CatalogItem{}) {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		d.SetId("")
		return nil
	}

	err = catalogItem.Delete()
	if err != nil {
		log.Printf("Error removing catalog item %#v", err)
		return fmt.Errorf("error removing catalog item %#v", err)
	}

	log.Printf("[TRACE] Catalog item delete completed: %#v", catalogItem.CatalogItem)

	return nil
}

// Finds catalog item which can be vapp template OVA or media ISO file
func findCatalogItem(d *schema.ResourceData, vcdClient *VCDClient) error {
	log.Printf("[TRACE] Catalog item read initiated")

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.FindCatalog(d.Get("catalog").(string))
	if err != nil || catalog == (govcd.Catalog{}) {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

	catalogItem, err := catalog.FindCatalogItem(d.Get("name").(string))
	if err != nil || catalogItem == (govcd.CatalogItem{}) {
		log.Printf("[DEBUG] Unable to find catalog item. Removing from tfstate")
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] Catalog item read completed: %#v", catalogItem.CatalogItem)
	return nil
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
