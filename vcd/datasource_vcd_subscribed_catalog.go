package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdSubscribedCatalog() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSubscribedCatalogRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional storage profile ID",
			},
			"subscription_url": {
				Type:     schema.TypeString,
				Optional: true,
				//ExactlyOneOf: []string{"subscription_url", "subscription_catalog_href"},
				Description: "The URL to subscribe to the external catalog. Required when 'subscription_catalog_href' is not provided",
			},
			"subscription_catalog_href": {
				Type:     schema.TypeString,
				Optional: true,
				//ExactlyOneOf: []string{"subscription_url", "subscription_catalog_href"},
				Description: "The HREF of the external catalog we want to subscribe to. Required when 'subscription_url' is not given",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An optional password to access the catalog.",
			},
			"make_local_copy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Start immediately importing subscribed items into local storage",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog metadata.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog HREF",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},
			"catalog_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Catalog version number.",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name from the catalog.",
			},
			"number_of_vapp_templates": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of vApps templates this catalog contains.",
			},
			"number_of_media": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of Media items this catalog contains.",
			},
			"vapp_template_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of catalog items in this catalog",
			},
			"media_item_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Media items in this catalog",
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is shared.",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is shared to all organizations.",
			},
			// The following properties are not used in this data source. Left here for compatibility with vcd_catalog
			"publish_subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "[UNUSED] PUBLISHED if published externally, SUBSCRIBED if subscribed to an external catalog, UNPUBLISHED otherwise.",
			},
			"publish_subscription_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "[UNUSED] URL to which other catalogs can subscribe",
			},
			"publish_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "[UNUSED] True allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation.",
			},
			"cache_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "[UNUSED] True enables early catalog export to optimize synchronization",
			},
			"preserve_identity_information": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "[UNUSED] Include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary.",
			},
		},
	}
}

func datasourceVcdSubscribedCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Subscribed Catalog read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		return diag.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	if adminCatalog.AdminCatalog.CatalogStorageProfiles != nil && len(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile) > 0 {
		storageProfileId := adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile[0].ID
		dSet(d, "storage_profile_id", storageProfileId)
	} else {
		dSet(d, "storage_profile_id", "")
	}

	dSet(d, "description", adminCatalog.AdminCatalog.Description)
	dSet(d, "created", adminCatalog.AdminCatalog.DateCreated)

	metadata, err := adminCatalog.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog metadata: %s", err)
		return diag.Errorf("%v", err)
	}

	if len(metadata.MetadataEntry) > 0 {
		err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
		if err != nil {
			return diag.Errorf("%v", err)

		}
	}

	if adminCatalog.AdminCatalog.ExternalCatalogSubscription != nil {
		dSet(d, "subscription_url", adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location)
		dSet(d, "make_local_copy", adminCatalog.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	}
	err = setCatalogData(d, adminOrg, adminCatalog)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	dSet(d, "href", adminCatalog.AdminCatalog.HREF)
	d.SetId(adminCatalog.AdminCatalog.ID)
	log.Printf("[TRACE] Subscribed Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}
