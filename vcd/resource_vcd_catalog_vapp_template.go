package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
	"strings"
	"time"
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
			"catalog_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the catalog where to upload the OVA file",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vApp Template name",
			},
			"description": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true, // Due to a bug in VCD when using `ovf_url`, `description` is overridden by the target OVA's description.
				Description:   "Description of the vApp Template. Not to be used with `ovf_url` when target OVA has a description",
				ConflictsWith: []string{"ovf_url"}, // This is to avoid the bug mentioned above.
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of when the vApp Template was created",
			},
			"ova_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"ova_path", "ovf_url"},
				Description:  "Absolute or relative path to OVA",
			},
			"ovf_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"ova_path", "ovf_url"},
				ConflictsWith: []string{"description"}, // This is to avoid the bug mentioned above.
				Description:   "URL of OVF file",
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Size of upload file piece size in megabytes",
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

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogId := d.Get("catalog_id").(string)
	catalog, err := org.GetCatalogById(catalogId, false)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	var diagError diag.Diagnostics
	vappTemplateName := d.Get("name").(string)
	if d.Get("ova_path").(string) != "" {
		diagError = uploadOvaFromResource(d, catalog, vappTemplateName, "vcd_catalog_vapp_template")
	} else if d.Get("ovf_url").(string) != "" {
		diagError = uploadFromUrl(d, catalog, vappTemplateName, "vcd_catalog_vapp_template")
	} else {
		return diag.Errorf("`ova_path` or `ovf_url` value is missing %s", err)
	}
	if diagError != nil {
		return diagError
	}

	vAppTemplate, err := catalog.GetVAppTemplateByName(vappTemplateName)
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
	vAppTemplate, err := findVAppTemplate(d, meta.(*VCDClient), origin)
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
	d.SetId(vAppTemplate.VAppTemplate.ID)
	return nil
}

func resourceVcdCatalogVappTemplateUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vAppTemplate, err := findVAppTemplate(d, meta.(*VCDClient), "resource")

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

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalog, err := org.GetCatalogById(d.Get("catalog_id").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		return diag.Errorf("unable to find catalog")
	}

	vAppTemplateName := d.Get("name").(string)
	vAppTemplate, err := catalog.GetVAppTemplateByName(vAppTemplateName)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template. Removing from tfstate")
		return diag.Errorf("unable to find vApp Template %s", vAppTemplateName)
	}

	err = vAppTemplate.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing vApp Template %s", err)
		return diag.Errorf("error removing vApp Template %s", err)
	}

	_, err = catalog.GetVAppTemplateByName(vAppTemplateName)
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
	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	vAppTemplate, err := catalog.GetVAppTemplateByName(vAppTemplateName)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "catalog_id", catalog.Catalog.ID)
	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	dSet(d, "description", vAppTemplate.VAppTemplate.Description)
	d.SetId(vAppTemplate.VAppTemplate.ID)

	return []*schema.ResourceData{d}, nil
}

// uploadOvaFromResource uploads an OVA file specified in the resource to the given catalog
func uploadOvaFromResource(d *schema.ResourceData, catalog *govcd.Catalog, vappTemplate, resourceName string) diag.Diagnostics {
	uploadPieceSize := d.Get("upload_piece_size").(int)
	task, err := catalog.UploadOvf(d.Get("ova_path").(string), vappTemplate, d.Get("description").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes
	if err != nil {
		log.Printf("[DEBUG] Error uploading file: %s", err)
		return diag.Errorf("error uploading file: %s", err)
	}

	return finishHandlingTask(d, *task.Task, vappTemplate, resourceName)
}

func uploadFromUrl(d *schema.ResourceData, catalog *govcd.Catalog, itemName, resourceName string) diag.Diagnostics {
	task, err := catalog.UploadOvfByLink(d.Get("ovf_url").(string), itemName, d.Get("description").(string))
	if err != nil {
		log.Printf("[DEBUG] Error uploading OVF from URL: %s", err)
		return diag.Errorf("error uploading OVF from URL: %s", err)
	}

	return finishHandlingTask(d, task, itemName, resourceName)
}

func finishHandlingTask(d *schema.ResourceData, task govcd.Task, itemName string, resourceName string) diag.Diagnostics {
	// This is a deprecated feature from vcd_catalog_item, to be removed with vcd_catalog_item
	if resourceName == "vcd_catalog_item" && d.Get("show_upload_progress").(bool) {
		for {
			progress, err := task.GetTaskProgress()
			if err != nil {
				log.Printf("VCD Error importing new catalog item: %s", err)
				return diag.Errorf("VCD Error importing new catalog item: %s", err)
			}
			logForScreen("vcd_catalog_item", fmt.Sprintf("vcd_catalog_item."+itemName+": VCD import catalog item progress "+progress+"%%\n"))
			if progress == "100" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	err := task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error waiting for task to complete: %+v", err)
	}
	return nil
}
