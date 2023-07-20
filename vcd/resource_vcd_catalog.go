package vcd

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// moreUpdateCatalogFunc is a typed func used to pass actions to the catalog update
type moreUpdateCatalogFunc func(d *schema.ResourceData, vcdClient *VCDClient, catalog *govcd.AdminCatalog, operation string) error

func resourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogCreate,
		DeleteContext: resourceVcdCatalogDelete,
		ReadContext:   resourceVcdCatalogRead,
		UpdateContext: resourceVcdCatalogUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogImport,
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
				// Not ForceNew, to allow the resource name to be updated
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
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove a catalog and any objects it contains, regardless of their state.",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_recursive=True to remove the catalog and any objects it contains that are in a state that normally allows removal.",
			},
			"publish_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation.",
			},
			"cache_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True enables early catalog export to optimize synchronization",
			},
			"preserve_identity_information": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "An optional password to access the catalog. Only ASCII characters are allowed in a valid password.",
			},
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
				Description:   "Key and value pairs for catalog metadata.",
			},
			"metadata_entry": metadataEntryResourceSchema("Catalog"),
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog HREF",
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
				Description: "Number of Medias this catalog contains.",
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
			"is_local": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog belongs to the current organization.",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this catalog is published.",
			},
			"publish_subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PUBLISHED if published externally, SUBSCRIBED if subscribed to an external catalog, UNPUBLISHED otherwise.",
			},
			"publish_subscription_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL to which other catalogs can subscribe",
			},
		},
	}
}

func resourceVcdCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog creation initiated")

	vcdClient := meta.(*VCDClient)

	// catalog creation is accessible only in administrator API part
	// (only administrator, organization administrator and Catalog author are allowed)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var storageProfiles *types.CatalogStorageProfiles
	storageProfileId := d.Get("storage_profile_id").(string)
	if storageProfileId != "" {
		storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
		if err != nil {
			return diag.Errorf("error looking up Storage Profile '%s' reference: %s", storageProfileId, err)
		}
		storageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	catalog, err := adminOrg.CreateCatalogWithStorageProfile(name, description, storageProfiles)
	if err != nil {
		log.Printf("[TRACE] Error creating Catalog: %#v", err)
		return diag.Errorf("error creating Catalog: %#v", err)
	}

	d.SetId(catalog.AdminCatalog.ID)

	publishEnabled := d.Get("publish_enabled").(bool)
	if publishEnabled {
		err = updatePublishToExternalOrgSettings(d, catalog)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = waitForMetadataReadiness(catalog)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] adding metadata for catalog")
	err = createOrUpdateMetadata(d, catalog, "metadata")
	if err != nil {
		return diag.Errorf("error adding catalog metadata: %s", err)
	}

	log.Printf("[TRACE] Catalog created: %#v", catalog)
	return resourceVcdCatalogRead(ctx, d, meta)
}

// waitForMetadataReadiness waits for the Catalog to have links to add metadata, so it can be added without errors.
// It will wait for 30 seconds maximum, or less if the links are ready.
func waitForMetadataReadiness(catalog *govcd.AdminCatalog) error {
	timeout := time.Second * 30
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("error waiting for the Catalog '%s' to be ready", catalog.AdminCatalog.ID)
		}
		link := catalog.AdminCatalog.Link.ForType("application/vnd.vmware.vcloud.metadata+xml", "add")
		if link != nil {
			util.Logger.Printf("catalog '%s' - metadata link found after %s\n", catalog.AdminCatalog.Name, time.Since(startTime))
			break
		}
		err := catalog.Refresh()
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func updatePublishToExternalOrgSettings(d *schema.ResourceData, adminCatalog *govcd.AdminCatalog) error {
	err := adminCatalog.PublishToExternalOrganizations(types.PublishExternalCatalogParams{
		IsPublishedExternally:    addrOf(d.Get("publish_enabled").(bool)),
		IsCachedEnabled:          addrOf(d.Get("cache_enabled").(bool)),
		PreserveIdentityInfoFlag: addrOf(d.Get("preserve_identity_information").(bool)),
		Password:                 d.Get("password").(string),
	})
	if err != nil {
		return fmt.Errorf("[updatePublishToExternalOrgSettings] error: %s", err)
	}
	return nil
}

func resourceVcdCatalogRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog read initiated")

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

	// Check if storage profile is set. Although storage profile structure accepts a list, in UI only one can be picked
	if adminCatalog.AdminCatalog.CatalogStorageProfiles != nil && len(adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile) > 0 {
		// By default, API does not return Storage Profile Name in response. It has ID and HREF, but not Name so name
		// must be looked up
		storageProfileId := adminCatalog.AdminCatalog.CatalogStorageProfiles.VdcStorageProfile[0].ID
		dSet(d, "storage_profile_id", storageProfileId)
	} else {
		// In case no storage profile are defined in API call
		dSet(d, "storage_profile_id", "")
	}

	dSet(d, "description", adminCatalog.AdminCatalog.Description)
	dSet(d, "created", adminCatalog.AdminCatalog.DateCreated)
	if adminCatalog.AdminCatalog.PublishExternalCatalogParams != nil {
		dSet(d, "publish_enabled", adminCatalog.AdminCatalog.PublishExternalCatalogParams.IsPublishedExternally)
		dSet(d, "cache_enabled", adminCatalog.AdminCatalog.PublishExternalCatalogParams.IsCachedEnabled)
		dSet(d, "preserve_identity_information", adminCatalog.AdminCatalog.PublishExternalCatalogParams.PreserveIdentityInfoFlag)
		subscriptionUrl, err := adminCatalog.FullSubscriptionUrl()
		if err != nil {
			return diag.Errorf("error retrieving subscription URL from catalog %s: %s", adminCatalog.AdminCatalog.Name, err)
		}
		dSet(d, "publish_subscription_url", subscriptionUrl)
	} else {
		dSet(d, "publish_enabled", false)
		dSet(d, "cache_enabled", false)
		dSet(d, "preserve_identity_information", false)
		dSet(d, "password", "")
	}

	err = setCatalogData(d, vcdClient, adminOrg.AdminOrg.Name, adminOrg.AdminOrg.ID, adminCatalog)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "href", adminCatalog.AdminCatalog.HREF)
	d.SetId(adminCatalog.AdminCatalog.ID)

	diagErr := updateMetadataInState(d, vcdClient, "vcd_catalog", adminCatalog)
	if diagErr != nil {
		log.Printf("[DEBUG] Unable to update catalog metadata: %s", err)
		return diagErr
	}
	log.Printf("[TRACE] Catalog read completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdCatalogUpdate does not require actions for  fields "delete_force", "delete_recursive",
// but does allow changing `storage_profile`
func resourceVcdCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericResourceVcdCatalogUpdate(ctx, d, meta, nil, resourceVcdCatalogRead)
}

// genericResourceVcdCatalogUpdate can handle update for both vcd_catalog and vcd_subscribed_catalog
// The mucf parameter is a slice of updating functions which –if provided– will be processed sequentially
// The readFunc parameter is the Read function to be used at the end of update.
func genericResourceVcdCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, mucf []moreUpdateCatalogFunc, readFunc schema.ReadContextFunc) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return diag.Errorf("error retrieving catalog %s : %s", d.Id(), err)
	}

	// A subscribed catalog has some restrictions on update.
	isSubscribed := adminCatalog.AdminCatalog.ExternalCatalogSubscription != nil &&
		adminCatalog.AdminCatalog.ExternalCatalogSubscription.Location != ""

	// Create a copy of adminCatalog to only set and change things which are related to this update section and skip the
	// other fields. This is important as this provider does not cover all settings available in API and they should not be
	// overwritten.
	newAdminCatalog := govcd.NewAdminCatalogWithParent(&vcdClient.VCDClient.Client, adminOrg)
	newAdminCatalog.AdminCatalog.ID = adminCatalog.AdminCatalog.ID
	newAdminCatalog.AdminCatalog.HREF = adminCatalog.AdminCatalog.HREF
	newAdminCatalog.AdminCatalog.Name = adminCatalog.AdminCatalog.Name

	// Perform storage profile updates
	if d.HasChange("storage_profile_id") {
		storageProfileId := d.Get("storage_profile_id").(string)

		// Unset storage profile (use any available in Org)
		if storageProfileId == "" {
			// Set empty structure as `nil` would not update it at all
			newAdminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{}}
		}

		if storageProfileId != "" {
			storageProfileReference, err := adminOrg.GetStorageProfileReferenceById(storageProfileId, false)
			if err != nil {
				return diag.Errorf("could not process Storage Profile '%s': %s", storageProfileId, err)
			}
			newAdminCatalog.AdminCatalog.CatalogStorageProfiles = &types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{storageProfileReference}}
		}
	}

	if !isSubscribed && d.HasChange("description") {
		// Subscribed catalogs get their description from the publishing catalog.
		// Attempting to change it will fail silently
		newAdminCatalog.AdminCatalog.Description = d.Get("description").(string)
	}

	if d.HasChanges("name", "description", "storage_profile_id") {
		newAdminCatalog.AdminCatalog.Name = d.Get("name").(string)
		err = newAdminCatalog.Update()
		if err != nil {
			return diag.Errorf("error updating catalog '%s': %s", adminCatalog.AdminCatalog.Name, err)
		}
	}

	// Subscribed catalogs cannot add or change publishing parameters or metadata
	if !isSubscribed {
		if d.HasChanges("publish_enabled", "cache_enabled", "preserve_identity_information", "password") {
			err = updatePublishToExternalOrgSettings(d, newAdminCatalog)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		log.Printf("[TRACE] updating metadata for catalog")
		err = createOrUpdateMetadata(d, adminCatalog, "metadata")
		if err != nil {
			return diag.Errorf("error updating catalog metadata: %s", err)
		}
	}

	// If there are custom catalog update functions, we run them one at the time
	for _, f := range mucf {
		if f != nil {
			err = f(d, vcdClient, newAdminCatalog, "update")
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return readFunc(ctx, d, meta)
}

func resourceVcdCatalogDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Catalog delete started")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	adminCatalog, err := adminOrg.GetAdminCatalogByNameOrId(d.Id(), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find catalog. Removing from tfstate")
		d.SetId("")
		return nil
	}

	err = adminCatalog.Delete(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		log.Printf("[DEBUG] Error removing catalog %#v", err)
		return diag.Errorf("error removing catalog %#v", err)
	}

	log.Printf("[TRACE] Catalog delete completed: %#v", adminCatalog.AdminCatalog)
	return nil
}

// resourceVcdCatalogImport imports a Catalog into Terraform state
// This function task is to get the data from vCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of org_name.catalog_name
//
// Example import path (id): org_name.catalog_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	catalog, err := adminOrg.GetCatalogByName(catalogName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "name", catalogName)
	dSet(d, "description", catalog.Catalog.Description)
	d.SetId(catalog.Catalog.ID)

	// Fill in other fields
	diagErr := resourceVcdCatalogRead(ctx, d, meta)
	if diagErr != nil {
		return nil, fmt.Errorf("error during read after import: %v", diagErr)
	}

	return []*schema.ResourceData{d}, nil
}

func setCatalogData(d *schema.ResourceData, vcdClient *VCDClient, orgName, orgId string, adminCatalog *govcd.AdminCatalog) error {
	// Catalog record is retrieved to get the owner name, number of vApp templates and medias, and if the catalog is shared and published
	catalogRecords, err := vcdClient.VCDClient.Client.QueryCatalogRecords(adminCatalog.AdminCatalog.Name, govcd.TenantContext{OrgName: orgName, OrgId: orgId})
	if err != nil {
		log.Printf("[DEBUG] [setCatalogData] Unable to retrieve catalog records: %s", err)
		return fmt.Errorf("[setCatalogData] unable to retrieve catalog records - %s", err)
	}
	var catalogRecord *types.CatalogRecord

	for _, cr := range catalogRecords {
		if cr.OrgName == orgName {
			catalogRecord = cr
			break
		}
	}
	if catalogRecord == nil {
		return fmt.Errorf("[setCatalogData] error retrieving catalog record for catalog '%s' from org '%s'", adminCatalog.AdminCatalog.Name, orgName)
	}

	util.Logger.Printf("[setCatalogData] catalogRecord %# v\n", pretty.Formatter(catalogRecord))
	dSet(d, "catalog_version", catalogRecord.Version)
	dSet(d, "owner_name", catalogRecord.OwnerName)
	dSet(d, "is_published", catalogRecord.IsPublished)
	dSet(d, "is_shared", catalogRecord.IsShared)
	dSet(d, "is_local", catalogRecord.IsLocal)
	dSet(d, "publish_subscription_type", catalogRecord.PublishSubscriptionType)

	var rawMediaItemsList []interface{}
	var rawVappTemplatesList []interface{}

	var mediaItemList []string
	var vappTemplateList []string

	mediaItems, err := adminCatalog.QueryMediaList()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	for _, media := range mediaItems {
		mediaItemList = append(mediaItemList, media.Name)
	}

	vappTemplates, err := adminCatalog.QueryVappTemplateList()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	for _, vappTemplate := range vappTemplates {
		vappTemplateList = append(vappTemplateList, vappTemplate.Name)
	}

	// Sort the lists, so that they will always match in state
	sort.Strings(mediaItemList)
	sort.Strings(vappTemplateList)
	for _, mediaName := range mediaItemList {
		rawMediaItemsList = append(rawMediaItemsList, mediaName)
	}
	for _, vappTemplateName := range vappTemplateList {
		rawVappTemplatesList = append(rawVappTemplatesList, vappTemplateName)
	}
	err = d.Set("media_item_list", rawMediaItemsList)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	err = d.Set("vapp_template_list", rawVappTemplatesList)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	dSet(d, "number_of_vapp_templates", len(vappTemplates))
	dSet(d, "number_of_media", len(mediaItems))

	return nil
}
