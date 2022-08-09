package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
	"strings"
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
				Description: "vApp Template name",
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
				Description: "Key and value pairs for the metadata of this vApp Template",
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

	vAppTemplate, err := catalog.GetVappTemplateByName(vappTemplateName, true)
	if err != nil {
		return diag.Errorf("error retrieving vApp Template %s: %s", vappTemplateName, err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)

	log.Printf("[TRACE] Catalog vApp Template created: %s", vappTemplateName)

	err = createOrUpdateMetadata(d, vAppTemplate, "metadata")
	if diagError != nil {
		return diag.FromErr(err)
	}

	return resourceVcdCatalogVappTemplateRead(ctx, d, meta)
}

func resourceVcdCatalogVappTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogVappTemplateRead(ctx, d, meta, "resource")
}

func genericVcdCatalogVappTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vAppTemplate, err := findVappTemplate(d, meta.(*VCDClient), origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template: %s", err)
		return diag.Errorf("Unable to find vApp Template: %s", err)
	}

	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	dSet(d, "created", vAppTemplate.VAppTemplate.DateCreated)
	dSet(d, "description", vAppTemplate.VAppTemplate.Description)

	metadata, err := vAppTemplate.GetMetadata()
	if err != nil {
		return diag.Errorf("Unable to find vApp template metadata: %s", err)
	}
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set metadata for the vApp Template: %s", err)
	}
	return nil
}

func resourceVcdCatalogVappTemplateUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vAppTemplate, err := findVappTemplate(d, meta.(*VCDClient), "resource")

	if d.HasChange("description") || d.HasChange("name") {
		if err != nil {
			return diag.Errorf("Unable to find vApp Template: %s", err)
		}

		vAppTemplate.VAppTemplate.Description = d.Get("description").(string)
		vAppTemplate.VAppTemplate.Name = d.Get("name").(string)
		_, err = vAppTemplate.Update()
		if err != nil {
			return diag.Errorf("error updating vApp Template: %s", err)
		}
	}

	err = createOrUpdateMetadata(d, vAppTemplate, "metadata")
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdCatalogVappTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] vApp Template delete started")
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := adminOrg.GetCatalogByName(d.Get("catalog").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		return diag.Errorf("unable to find catalog")
	}

	vAppTemplateName := d.Get("name").(string)
	vAppTemplate, err := catalog.GetVappTemplateByName(vAppTemplateName, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template. Removing from tfstate")
		return diag.Errorf("unable to find vApp Template %s", vAppTemplateName)
	}

	err = vAppTemplate.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing vApp Template %s", err)
		return diag.Errorf("error removing vApp Template %s", err)
	}

	_, err = catalog.GetVappTemplateByName(vAppTemplateName, true)
	if err == nil {
		return diag.Errorf("vApp Template %s still found after deletion", vAppTemplateName)
	}
	log.Printf("[TRACE] vApp Template delete completed: %s", vAppTemplateName)

	return nil
}

// Imports a vApp Template into Terraform state
// This function task is to get the data from VCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name.vapp_template_name
//
// Example import path (id): myOrg1.myCatalog2.myvAppTemplate3
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogVappTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog_name.vapp_template_name")
	}
	orgName, catalogName, vAppTemplateName := resourceURI[0], resourceURI[1], resourceURI[2]

	if orgName == "" {
		return nil, fmt.Errorf("import: empty Org name provided")
	}
	if catalogName == "" {
		return nil, fmt.Errorf("import: empty Catalog name provided")
	}
	if vAppTemplateName == "" {
		return nil, fmt.Errorf("import: empty vApp Template name provided")
	}

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	vAppTemplate, err := catalog.GetVappTemplateByName(vAppTemplateName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "catalog", catalogName)
	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	dSet(d, "description", vAppTemplate.VAppTemplate.Description)
	d.SetId(vAppTemplate.VAppTemplate.ID)

	return []*schema.ResourceData{d}, nil
}
