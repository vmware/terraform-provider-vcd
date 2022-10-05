package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdSubscribedCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSubscribedCatalogCreate,
		ReadContext:   resourceVcdSubscribedCatalogRead,
		UpdateContext: resourceVcdSubscribedCatalogUpdate,
		DeleteContext: resourceVcdSubscribedCatalogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSUbscribedCatalogImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional storage profile ID",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state.",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal.",
			},
			"subscription_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subscription_url", "subscription_catalog_href"},
				Description:  "The URL to subscribe to the external catalog. Required when 'subscription_catalog_href' is not provided",
			},
			"subscription_catalog_href": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subscription_url", "subscription_catalog_href"},
				Description:  "The HREF of the external catalog we want to subscribe to. Required when 'subscription_url' is not given",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "An optional password to access the catalog. Only ASCII characters are allowed in a valid password.",
			},
			"make_local_copy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Start immediately importing subscribed items into local storage",
			},
			"synchronize": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Get subscribed contents (only used during updates)",
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      15,
				RequiredWith: []string{"make_local_copy"},
				Description:  "Timeout (in minutes) for import completion. Required when 'make_local_copy' is true. Default 15 minutes. (0 = no timeout)",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
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
			// The following properties are not used in this resource. Left here for compatibility with vcd_catalog
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

func resourceVcdSubscribedCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Subscribed Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var adminCatalog *govcd.AdminCatalog
	var storageProfiles *types.CatalogStorageProfiles

	catalogName := d.Get("name").(string)
	catalogDescription := d.Get("description").(string)
	password := d.Get("password").(string)
	storageProfileId := d.Get("storage_profile_id").(string)
	subscriptionUrl := d.Get("subscription_url").(string)
	makeLocalCopy := d.Get("make_local_copy").(bool)
	timeout := d.Get("timeout").(int)
	if storageProfileId != "" {
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return diag.Errorf("error looking up Storage Profile '%s' reference: %s", storageProfileId, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}
	subscriptionCatalogHref := d.Get("subscription_catalog_href").(string)
	if subscriptionCatalogHref != "" {
		if subscriptionUrl != "" {
			return diag.Errorf("only one of 'subscription_catalog_href' or 'subscription_url' can be filled")
		}
		fromCatalog, err := vcdClient.Client.GetAdminCatalogByHref(subscriptionCatalogHref)
		if err != nil {
			return diag.Errorf("error fetching admin catalog %s: %s", subscriptionCatalogHref, err)
		}
		if makeLocalCopy {
			adminCatalog, err = adminOrg.ImportFromCatalog(fromCatalog, storageProfiles,
				catalogName, catalogDescription, password, makeLocalCopy, time.Duration(timeout)*time.Minute)
		} else {
			adminCatalog, err = adminOrg.ImportFromCatalogAsync(fromCatalog, storageProfiles, catalogName, catalogDescription, password, makeLocalCopy)
		}
		if err != nil {
			return diag.Errorf("error importing from catalog %s: %s", fromCatalog.AdminCatalog.Name, err)
		}
	} else {
		if makeLocalCopy {
			adminCatalog, err = adminOrg.CreateCatalogFromSubscription(types.ExternalCatalogSubscription{
				SubscribeToExternalFeeds: true,
				Location:                 subscriptionUrl,
				Password:                 password,
				LocalCopy:                makeLocalCopy,
			}, storageProfiles, catalogName, catalogDescription, password, makeLocalCopy, time.Duration(timeout)*time.Minute)
		} else {
			adminCatalog, err = adminOrg.CreateCatalogFromSubscriptionAsync(types.ExternalCatalogSubscription{
				SubscribeToExternalFeeds: true,
				Location:                 subscriptionUrl,
				Password:                 password,
				LocalCopy:                makeLocalCopy,
			}, storageProfiles, catalogName, catalogDescription, password, makeLocalCopy)
		}
		if err != nil {
			return diag.Errorf("error creating catalog %s from subscription: %s", catalogName, err)
		}
	}
	d.SetId(adminCatalog.AdminCatalog.ID)

	log.Printf("[TRACE] creating metadata for catalog")
	err = createOrUpdateMetadata(d, adminCatalog, "metadata")
	if err != nil {
		return diag.Errorf("%v", err)
	}
	log.Printf("[TRACE] Subscribed Catalog created: %#v", adminCatalog)
	return resourceVcdSubscribedCatalogRead(ctx, d, meta)
}

func resourceVcdSubscribedCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Subscribed Catalog read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
			d.SetId("")
			return nil
		}

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

	err = setCatalogData(d, adminOrg, adminCatalog)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	dSet(d, "href", adminCatalog.AdminCatalog.HREF)
	d.SetId(adminCatalog.AdminCatalog.ID)
	log.Printf("[TRACE] Subscribed Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

func resourceVcdSubscribedCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// The update in this resource differs only in the "syncronize", "make_local_copy", and "password" fields
	// Thus, we use the regular catalog update first, and add the specifics for subscribed next
	err := resourceVcdCatalogUpdate(ctx, d, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	// TODO: add a synchronise method
	// TODO: add an update subscription settings method

	return resourceVcdSubscribedCatalogRead(ctx, d, meta)
}

func resourceVcdSubscribedCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := resourceVcdCatalogDelete(ctx, d, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	return nil
}

func resourceVcdSUbscribedCatalogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog")
	}
	orgName, catalogName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	catalog, err := adminOrg.GetAdminCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "name", catalogName)
	dSet(d, "description", catalog.AdminCatalog.Description)
	d.SetId(catalog.AdminCatalog.ID)

	return []*schema.ResourceData{d}, nil
}
