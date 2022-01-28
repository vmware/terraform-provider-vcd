package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCatalogRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Name of the catalog. (Optional if 'filter' is used)",
				ExactlyOneOf: []string{"name", "filter"},
			},
			"storage_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile ID",
			},
			"created": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"publish_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation. Default is `false`.",
			},
			"cache_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True enables early catalog export to optimize synchronization",
			},
			"preserve_identity_information": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog metadata",
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
			"filter": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a catalog by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"date":       elementDate,
						"earliest":   elementEarliest,
						"latest":     elementLatest,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func datasourceVcdCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		vcdClient = meta.(*VCDClient)
		err       error
		adminOrg  *govcd.AdminOrg
		catalog   *govcd.Catalog
	)

	if !nameOrFilterIsSet(d) {
		return diag.Errorf(noNameOrFilterError, "vcd_catalog")
	}
	orgName := d.Get("org").(string)
	identifier := d.Get("name").(string)
	log.Printf("[TRACE] Reading Org %s", orgName)
	adminOrg, err = vcdClient.VCDClient.GetAdminOrgByName(orgName)

	if err != nil {
		log.Printf("[DEBUG] Org %s not found. Setting ID to nothing", orgName)
		d.SetId("")
		return nil
	}
	log.Printf("[TRACE] Org %s found", orgName)

	filter, hasFilter := d.GetOk("filter")

	if hasFilter {
		catalog, err = getCatalogByFilter(adminOrg, filter, vcdClient.Client.IsSysAdmin)
	} else {
		catalog, err = adminOrg.GetCatalogByNameOrId(identifier, false)
	}
	if err != nil {
		log.Printf("[DEBUG] Catalog %s not found. Setting ID to nothing", identifier)
		d.SetId("")
		return diag.Errorf("error retrieving catalog %s: %s", identifier, err)
	}

	metadata, err := catalog.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog metadata: %s", err)
		return diag.Errorf("There was an issue when retrieving metadata - %s", err)
	}

	dSet(d, "description", catalog.Catalog.Description)
	dSet(d, "created", catalog.Catalog.DateCreated)
	dSet(d, "name", catalog.Catalog.Name)
	d.SetId(catalog.Catalog.ID)
	if catalog.Catalog.PublishExternalCatalogParams != nil {
		dSet(d, "publish_enabled", catalog.Catalog.PublishExternalCatalogParams.IsPublishedExternally)
		dSet(d, "cache_enabled", catalog.Catalog.PublishExternalCatalogParams.IsCachedEnabled)
		dSet(d, "preserve_identity_information", catalog.Catalog.PublishExternalCatalogParams.PreserveIdentityInfoFlag)
	}

	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
	}

	return nil
}
