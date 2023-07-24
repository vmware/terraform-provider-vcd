package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdCatalog() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCatalogRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Name of the catalog. (Optional if 'filter' is used)",
				ExactlyOneOf: []string{"name", "filter"},
			},
			"storage_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile ID",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the catalog was created",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"publish_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation. Default is `false`.",
			},
			"cache_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True enables early catalog export to optimize synchronization",
			},
			"preserve_identity_information": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog HREF",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog metadata",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("Catalog"),
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
				Description: "Number of vApps this catalog contains.",
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
				Description: "True if this catalog is shared to all organizations.",
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
			"filter": {
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

func getCatalogFromResource(catalogName string, d *schema.ResourceData, meta interface{}) (*govcd.AdminCatalog, error) {
	vcdClient := meta.(*VCDClient)

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("'org' property not supplied in the resource or in provider: %s", err)
	}

	tenantContext := govcd.TenantContext{}
	if vcdClient.Client.IsSysAdmin {
		org, err := vcdClient.GetAdminOrgByName(orgName)
		if err != nil {
			return nil, fmt.Errorf("[getCatalogFromResource] error retrieving org %s: %s", orgName, err)
		}
		tenantContext.OrgId = org.AdminOrg.ID
		tenantContext.OrgName = orgName
	}
	catalogRecords, err := vcdClient.VCDClient.Client.QueryCatalogRecords(catalogName, tenantContext)
	if err != nil {
		return nil, fmt.Errorf("[getCatalogFromResource] error retrieving catalog records for catalog %s: %s", catalogName, err)
	}
	var catalogRecord *types.CatalogRecord
	var orgNames []string
	for _, cr := range catalogRecords {
		orgNames = append(orgNames, cr.OrgName)
		if cr.OrgName == orgName {
			catalogRecord = cr
			break
		}
	}
	if catalogRecord == nil {
		message := fmt.Sprintf("no records found for catalog '%s' in org '%s'", catalogName, orgName)
		if len(orgNames) > 0 {
			message = fmt.Sprintf("%s\nThere are catalogs with the same name in other orgs: %v", message, orgNames)
		}
		return nil, fmt.Errorf(message)
	}
	return vcdClient.Client.GetAdminCatalogByHref(catalogRecord.HREF)
}

func datasourceVcdCatalogRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		vcdClient             = meta.(*VCDClient)
		catalog               *govcd.AdminCatalog
		err                   error
		catalogOrgIsAvailable bool = true
	)

	if !nameOrFilterIsSet(d) {
		return diag.Errorf(noNameOrFilterError, "vcd_catalog")
	}

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return diag.Errorf("'org' property not supplied in the resource or in provider: %s", err)
	}

	adminOrg, orgErr := vcdClient.GetAdminOrgFromResource(d)
	orgFound := ""
	if orgErr != nil {
		util.Logger.Printf("[TRACE] error retrieving org: %s", orgErr)
		catalogOrgIsAvailable = false
		orgFound = "NOT"
	}
	util.Logger.Printf("[TRACE] Org '%s' %s found", orgName, orgFound)

	identifier := d.Get("name").(string)

	filter, hasFilter := d.GetOk("filter")

	if hasFilter {
		if !catalogOrgIsAvailable {
			return diag.Errorf("cannot search by filter when org is not reachable by the current user. "+
				"If the catalog is shared, it will be retrieved only by name: %s", orgErr)
		}
		catalog, err = getCatalogByFilter(adminOrg, filter, vcdClient.Client.IsSysAdmin)
	} else {
		catalog, err = getCatalogFromResource(identifier, d, meta)
	}
	if err != nil {
		return diag.Errorf("[catalog read DS] error retrieving catalog %s: %s - %s", identifier, govcd.ErrorEntityNotFound, err)
	}

	dSet(d, "description", catalog.AdminCatalog.Description)
	dSet(d, "created", catalog.AdminCatalog.DateCreated)
	dSet(d, "name", catalog.AdminCatalog.Name)

	dSet(d, "href", catalog.AdminCatalog.HREF)
	d.SetId(catalog.AdminCatalog.ID)
	if catalog.AdminCatalog.PublishExternalCatalogParams != nil {
		dSet(d, "publish_enabled", catalog.AdminCatalog.PublishExternalCatalogParams.IsPublishedExternally)
		dSet(d, "cache_enabled", catalog.AdminCatalog.PublishExternalCatalogParams.IsCachedEnabled)
		dSet(d, "preserve_identity_information", catalog.AdminCatalog.PublishExternalCatalogParams.PreserveIdentityInfoFlag)
		subscriptionUrl, err := catalog.FullSubscriptionUrl()
		if err != nil {
			return diag.FromErr(err)
		}
		dSet(d, "publish_subscription_url", subscriptionUrl)
	}

	orgId := ""
	if adminOrg != nil {
		orgId = adminOrg.AdminOrg.ID
	}
	err = setCatalogData(d, vcdClient, orgName, orgId, catalog)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr := updateMetadataInState(d, vcdClient, "vcd_catalog", catalog)
	if diagErr != nil {
		return diagErr
	}

	return nil
}
