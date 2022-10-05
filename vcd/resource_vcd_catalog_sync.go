package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdCatalogSync() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogSyncCreateUpdate,
		ReadContext:   resourceVcdCatalogSyncRead,
		UpdateContext: resourceVcdCatalogSyncCreateUpdate,
		//DeleteContext: resourceVcdCatalogSyncDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogSyncImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of Catalog to use",
			},
		},
	}
}

func resourceVcdCatalogSyncRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	return diag.Errorf("IMPLEMENT THIS METHOD")
}

func resourceVcdCatalogSyncCreateUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	return diag.Errorf("IMPLEMENT THIS METHOD")
}

func resourceVcdCatalogSyncImport(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	return nil, fmt.Errorf("IMPLEMENT THIS METHOD")
}
