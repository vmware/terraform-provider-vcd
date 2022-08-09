package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceVcdCatalogVappTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogVappTemplateCreate,
		ReadContext:   resourceVcdCatalogVappTemplateRead,
		UpdateContext: resourceVcdCatalogVappTemplateUpdate,
		DeleteContext: resourceVcdCatalogVappTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogVappTemplateImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog name where upload the OVA file",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "catalog item name",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the item was created",
			},
			"ova_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"ova_path", "ovf_url"},
				Description:  "Absolute or relative path to OVA",
			},
			"ovf_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"ova_path", "ovf_url"},
				Description:  "URL of OVF file",
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "size of upload file piece size in mega bytes",
			},
			"show_upload_progress": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "shows upload progress in stdout",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for the metadata of the vApp template associated to this catalog item",
			},
			"catalog_item_metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key and value pairs for catalog item metadata",
			},
		},
	}
}

func resourceVcdCatalogVappTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog vApp Template creation initiated")

	vcdClient := meta.(*VCDClient)

	// TODO: Why admin?
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogName := d.Get("catalog").(string)
	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	var diagError diag.Diagnostics
	vappTemplateName := d.Get("name").(string)
	if d.Get("ova_path").(string) != "" {
		diagError = uploadFile(d, catalog, vappTemplateName, "vcd_catalog_vapp_template")
	} else if d.Get("ovf_url").(string) != "" {
		diagError = uploadFromUrl(d, catalog, vappTemplateName)
	} else {
		return diag.Errorf("`ova_path` or `ovf_url` value is missing %s", err)
	}
	if diagError != nil {
		return diagError
	}

	item, err := catalog.GetCatalogItemByName(vappTemplateName, true)
	if err != nil {
		return diag.Errorf("error retrieving vApp Template %s: %s", vappTemplateName, err)
	}
	vAppTemplate, err := item.GetVAppTemplate()
	if err != nil {
		return diag.Errorf("error retrieving vApp Template %s: %s", vappTemplateName, err)
	}
	d.SetId(vAppTemplate.VAppTemplate.ID)

	log.Printf("[TRACE] Catalog vApp Template created: %s", vappTemplateName)

	// TODO
	err = createOrUpdateCatalogItemMetadata(d, meta)
	if diagError != nil {
		return diag.FromErr(err)
	}

	return resourceVcdCatalogVappTemplateRead(ctx, d, meta)
}

func resourceVcdCatalogVappTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

}

func resourceVcdCatalogVappTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

}

func resourceVcdCatalogVappTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

}

func resourceVcdCatalogVappTemplateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

}
