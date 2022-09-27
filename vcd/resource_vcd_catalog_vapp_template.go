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
				Description: "ID of the Catalog where to upload the OVA file",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the VDC to which the vApp Template belongs",
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
			"vm_names": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of VM names within the vApp template",
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
		diagError = diag.Errorf("`ova_path` or `ovf_url` value is missing %s", err)
	}
	if diagError != nil {
		return diagError
	}

	vAppTemplate, err := catalog.GetVAppTemplateByName(vappTemplateName)
	if err != nil {
		return diag.Errorf("error retrieving vApp Template %s: %s", vappTemplateName, err)
	}

	err = createOrUpdateMetadata(d, vAppTemplate, "metadata")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)
	log.Printf("[TRACE] Catalog vApp Template created: %s", vappTemplateName)

	return resourceVcdCatalogVappTemplateRead(ctx, d, meta)
}

func resourceVcdCatalogVappTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogVappTemplateRead(ctx, d, meta, "resource")
}

// genericVcdCatalogVappTemplateRead performs a Read operation for the vApp Template resource (origin="resource")
// and data source (origin="datasource").
func genericVcdCatalogVappTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vAppTemplate, err := findVAppTemplate(d, vcdClient, origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template: %s", err)
		return diag.Errorf("Unable to find vApp Template: %s", err)
	}

	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	dSet(d, "created", vAppTemplate.VAppTemplate.DateCreated)
	dSet(d, "description", vAppTemplate.VAppTemplate.Description)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	_, isCatalogIdSet := d.GetOk("catalog_id")
	if !isCatalogIdSet { // This can only happen in the data source.
		catalogName, err := vAppTemplate.GetCatalogName()
		if err != nil {
			return diag.Errorf("error retrieving the Catalog name to which the vApp Template '%s' belongs: %s", vAppTemplate.VAppTemplate.Name, err)
		}
		catalog, err := adminOrg.GetCatalogByName(catalogName, false)
		if err != nil {
			return diag.Errorf("error retrieving Catalog from vApp Template with name %s: %s", vAppTemplate.VAppTemplate.Name, err)
		}
		dSet(d, "catalog_id", catalog.Catalog.ID)
	} else {
		vdcName, err := vAppTemplate.GetVdcName()
		if err != nil {
			return diag.Errorf("error retrieving the VDC name to which the vApp Template '%s' belongs: %s", vAppTemplate.VAppTemplate.Name, err)
		}
		vdc, err := adminOrg.GetVDCByName(vdcName, false)
		if err != nil {
			return diag.Errorf("error retrieving the VDC to which the vApp Template '%s' belongs: %s", vAppTemplate.VAppTemplate.Name, err)
		}
		dSet(d, "vdc_id", vdc.Vdc.ID)
	}

	metadata, err := vAppTemplate.GetMetadata()
	if err != nil {
		return diag.Errorf("Unable to find vApp template metadata: %s", err)
	}
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("Unable to set metadata for the vApp Template: %s", err)
	}

	var vmNames []string
	if vAppTemplate.VAppTemplate.Children != nil {
		for _, vm := range vAppTemplate.VAppTemplate.Children.VM {
			vmNames = append(vmNames, vm.Name)
		}
	}
	err = d.Set("vm_names", vmNames)
	if err != nil {
		diag.Errorf("Unable to set attribute 'vm_names' for the vApp Template: %s", err)
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
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	catalogId := d.Get("catalog_id").(string)
	catalog, err := org.GetCatalogById(catalogId, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find Catalog with ID %s", catalogId)
		return diag.Errorf("unable to find Catalog with ID %s", catalogId)
	}

	vAppTemplateName := d.Get("name").(string)
	vAppTemplate, err := catalog.GetVAppTemplateByName(vAppTemplateName)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template with name %s", vAppTemplateName)
		return diag.Errorf("unable to find vApp Template with name %s", vAppTemplateName)
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

// Finds a vApp Template with the information given in the resource data. If it's a data source it uses a filtering
// mechanism, if it's a resource it just gets the information.
func findVAppTemplate(d *schema.ResourceData, vcdClient *VCDClient, origin string) (*govcd.VAppTemplate, error) {
	log.Printf("[TRACE] vApp template search initiated")

	identifier := d.Id()
	// Check if identifier is still in deprecated style `catalogName:mediaName`
	// Required for backwards compatibility as identifier has been changed to vCD ID in 2.5.0
	if identifier == "" || strings.Count(identifier, ":") <= 1 {
		identifier = d.Get("name").(string)
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	// Get the catalog only if its ID is set, as in data source we can search with VDC ID instead.
	var catalog *govcd.Catalog
	var vdc *govcd.Vdc
	catalogId, isSearchedByCatalog := d.GetOk("catalog_id")
	if isSearchedByCatalog {
		catalog, err = adminOrg.GetCatalogById(catalogId.(string), false)
		if err != nil {
			log.Printf("[DEBUG] Unable to find Catalog.")
			return nil, fmt.Errorf("unable to find Catalog: %s", err)
		}
	} else {
		vdc, err = adminOrg.GetVDCById(d.Get("vdc_id").(string), false)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VDC.")
			return nil, fmt.Errorf("unable to find VDC: %s", err)
		}
	}

	var vAppTemplate *govcd.VAppTemplate
	if origin == "datasource" {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_catalog_vapp_template")
		}

		filter, hasFilter := d.GetOk("filter")

		if hasFilter {
			if isSearchedByCatalog {
				vAppTemplate, err = getVappTemplateByCatalogAndFilter(catalog, filter, vcdClient.Client.IsSysAdmin)
			} else {
				vAppTemplate, err = getVappTemplateByVdcAndFilter(vdc, filter, vcdClient.Client.IsSysAdmin)
			}
			if err != nil {
				return nil, err
			}
			d.SetId(vAppTemplate.VAppTemplate.ID)
			return vAppTemplate, nil
		}
	}
	// No filter: we continue with single item  GET

	if isSearchedByCatalog {
		// In a resource, this is the only possibility
		vAppTemplate, err = catalog.GetVAppTemplateByNameOrId(identifier, false)
	} else {
		vAppTemplate, err = vdc.GetVAppTemplateByNameOrId(identifier, false)
	}

	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] Unable to find vApp Template %s. Removing from tfstate", identifier)
		d.SetId("")
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find vApp Template %s: %s", identifier, err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)
	log.Printf("[TRACE] vApp Template read completed: %#v", vAppTemplate.VAppTemplate)
	return vAppTemplate, nil
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
