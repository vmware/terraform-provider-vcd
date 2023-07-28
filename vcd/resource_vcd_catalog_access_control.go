package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"strings"
	"time"
)

func resourceVcdCatalogAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogAccessControlCreateUpdate,
		ReadContext:   resourceVcdCatalogAccessControlRead,
		UpdateContext: resourceVcdCatalogAccessControlCreateUpdate,
		DeleteContext: resourceVcdCatalogAccessControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogAccessControlImport,
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
			"shared_with_everyone": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the Catalog is shared with everyone",
			},
			"everyone_access_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly}, true),
				Description:  "Access level when the catalog is shared with everyone (only ReadOnly is available). Required when shared_with_everyone is set",
			},
			"read_only_shared_with_other_orgs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "If true, the catalog is shared as read-only with all organizations",
			},
			"shared_with": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"everyone_access_level"},
				MinItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the Org to which we are sharing. Required if user_id or group_id is not set",
						},
						"user_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the user to which we are sharing. Required if group_id or org_id is not set",
						},
						"group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the group to which we are sharing. Required if user_id or org_id is not set",
						},
						"subject_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subject (org, group, or user) with which we are sharing",
						},
						"access_level": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly, types.ControlAccessReadWrite, types.ControlAccessFullControl}, true),
							Description:  "The access level for the org, user, or group to which we are sharing. One of [ReadOnly, Change, FullControl] for users and groups, but just ReadOnly for Organizations",
						},
					},
				},
			},
		},
	}
}
func resourceVcdCatalogAccessControlCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	sessionInfo, err := vcdClient.Client.GetSessionInfo()
	if err != nil {
		return diag.Errorf("error retrieving session info: %s", err)
	}
	sessionText := fmt.Sprintf("[ vcd_catalog_access_control set - org: %s - user: %s]", sessionInfo.Org.Name, sessionInfo.User.Name)
	var accessSettings []*types.AccessSetting
	readOnlySharedwithOtherOrgs := d.Get("read_only_shared_with_other_orgs").(bool)
	isSharedWithEveryone := d.Get("shared_with_everyone").(bool)
	everyoneAccessLevel := getStringAttributeAsPointer(d, "everyone_access_level")
	isEveryoneAccessLevelSet := everyoneAccessLevel != nil
	sharedList := d.Get("shared_with").(*schema.Set).List()

	if !isSharedWithEveryone && isEveryoneAccessLevelSet {
		return diag.Errorf("if shared_with_everyone is set to false, everyone_access_level must not be set")
	}

	if isSharedWithEveryone && len(sharedList) > 0 {
		return diag.Errorf("if shared_with_everyone is set to true, shared_with must not be set")
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error when retrieving AdminOrg - %s", err)
	}

	if !isSharedWithEveryone {
		everyoneAccessLevel = nil

		accessSettings, err = sharedSetToAccessControl(vcdClient, adminOrg, sharedList, []string{"org_id", "user_id", "group_id"})
		if err != nil {
			return diag.Errorf("%s error when reading shared_with from schema - %s", sessionText, err)
		}
	}

	catalogId := d.Get("catalog_id").(string)
	catalog, err := adminOrg.GetAdminCatalogById(catalogId, false)
	if err != nil {
		return diag.Errorf("%s error when retrieving catalog %s - %s", sessionText, catalogId, err)
	}
	sessionText = fmt.Sprintf("[ vcd_catalog_access_control set - org: %s - user: %s - catalog: %s]",
		sessionInfo.Org.Name, sessionInfo.User.Name, catalog.AdminCatalog.Name)
	var accessSettingsList *types.AccessSettingList
	if accessSettings != nil {
		accessSettingsList = &types.AccessSettingList{
			AccessSetting: accessSettings,
		}
	} else {
		accessSettingsList = nil
	}
	var accessControlParams = types.ControlAccessParams{
		IsSharedToEveryone:  isSharedWithEveryone,
		EveryoneAccessLevel: everyoneAccessLevel,
		AccessSettings:      accessSettingsList,
	}
	_, err = retry(sessionText,
		"error when setting Catalog control access parameters",
		time.Second*30,
		nil, //func() error { return catalog.Refresh() },
		func() (any, error) {
			var err error
			if readOnlySharedwithOtherOrgs {
				err = catalog.SetReadOnlyAccessControl(true)
			} else {
				err = catalog.SetAccessControl(&accessControlParams, true)
			}
			return nil, err
		})
	//if readOnlySharedwithOtherOrgs {
	//	err = catalog.SetReadOnlyAccessControl(true)
	//} else {
	//	err = catalog.SetAccessControl(&accessControlParams, true)
	//}
	if err != nil {
		//fmt.Printf("%# v\n", pretty.Formatter(accessControlParams))
		return diag.Errorf("%s error when setting Catalog control access parameters - %s", sessionText, err)
	}

	d.SetId(catalog.AdminCatalog.ID)
	return resourceVcdCatalogAccessControlRead(ctx, d, meta)
}

func resourceVcdCatalogAccessControlRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	sessionInfo, err := vcdClient.Client.GetSessionInfo()
	if err != nil {
		return diag.Errorf("error retrieving session info: %s", err)
	}
	sessionText := fmt.Sprintf("[ vcd_catalog_access_control read - org: %s - user: %s]", sessionInfo.Org.Name, sessionInfo.User.Name)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("%s error while reading Org - %s", sessionText, err)
	}

	catalog, err := org.GetCatalogById(d.Id(), false)
	if err != nil {
		if govcd.IsNotFound(err) {
			d.SetId("")
			return nil
		} else {
			return diag.Errorf("%s error while reading Catalog - %s", sessionText, err)
		}
	}

	sessionText = fmt.Sprintf("[ vcd_catalog_access_control read - org: %s - user: %s - catalog: %s]",
		sessionInfo.Org.Name, sessionInfo.User.Name, catalog.Catalog.Name)

	sharedReadOnly, err := retry(sessionText,
		fmt.Sprintf("%s error checking catalog read-only sharing status", sessionText),
		time.Second*30,
		nil, //func() error { return catalog.Refresh() },
		func() (any, error) {
			return catalog.IsSharedReadOnly()
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "read_only_shared_with_other_orgs", sharedReadOnly.(bool))

	result, err := retry(
		fmt.Sprintf("%s getting control access parameters", sessionText),
		fmt.Sprintf("%s error getting control access parameters", sessionText),
		time.Second*30,
		nil,
		func() (any, error) {
			return catalog.GetAccessControl(true)
		},
	)
	//controlAccessParams, err := catalog.GetAccessControl(true)
	if err != nil {
		return diag.Errorf("%s error getting control access parameters - %s", sessionText, err)
	}
	controlAccessParams := result.(*types.ControlAccessParams)

	dSet(d, "shared_with_everyone", controlAccessParams.IsSharedToEveryone)
	if controlAccessParams.EveryoneAccessLevel != nil {
		dSet(d, "everyone_access_level", *controlAccessParams.EveryoneAccessLevel)
	} else {
		dSet(d, "everyone_access_level", "")
	}

	if controlAccessParams.AccessSettings != nil {
		accessControlListSet, err := accessControlListToSharedSet(controlAccessParams.AccessSettings.AccessSetting)
		if err != nil {
			return diag.Errorf("error converting slice AccessSetting into set - %s", err)
		}

		err = d.Set("shared_with", accessControlListSet)
		if err != nil {
			return diag.Errorf("error setting shared_with attribute - %s", err)
		}
	}

	return nil
}

func resourceVcdCatalogAccessControlDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// When deleting Catalog access control, Catalog won't be shared with anyone
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		diag.Errorf("error when retrieving Org - %s", err)
	}

	catalog, err := org.GetAdminCatalogById(d.Id(), false)
	if err != nil {
		diag.Errorf("error when retrieving Catalog - %s", err)
	}

	sharedReadOnly, err := catalog.IsSharedReadOnly()
	if err != nil {
		return diag.Errorf("error checking catalog read-only sharing status: %s", err)
	}

	if sharedReadOnly {
		_, err = retry(fmt.Sprintf("removing catalog '%s' shared read-only ", catalog.AdminCatalog.Name),
			"error removing catalog read-only access control",
			30*time.Second,
			nil,
			func() (any, error) {
				err := catalog.SetReadOnlyAccessControl(false)
				return nil, err
			})
		//err = catalog.SetReadOnlyAccessControl(false)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}
	_, err = retry(fmt.Sprintf("deleting catalog %s access control", catalog.AdminCatalog.Name),
		fmt.Sprintf("error when deleting catalog '%s' access control", catalog.AdminCatalog.Name),
		time.Second*30,
		nil,
		func() (any, error) {
			return nil, catalog.RemoveAccessControl(true)
		})
	//err = catalog.RemoveAccessControl(true)
	if err != nil {
		return diag.Errorf("error when deleting Catalog access control - %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdCatalogAccessControlImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.catalogID or org.catalogName")
	}

	orgName, catalogIdentifier := resourceURI[0], resourceURI[1]
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[catalog access control import] "+errorRetrievingOrg, err)
	}

	catalog, err := org.GetCatalogByNameOrId(catalogIdentifier, false)
	if err != nil {
		return nil, fmt.Errorf("[catalog access control import] error retrieving catalog '%s' from org '%s'", catalogIdentifier, org.Org.Name)
	}
	dSet(d, "org", orgName)
	dSet(d, "catalog_id", catalog.Catalog.ID)
	d.SetId(catalog.Catalog.ID)

	return []*schema.ResourceData{d}, nil
}

func retry(label, message string, timeout time.Duration, refresh func() error, operation func() (any, error)) (any, error) {
	if os.Getenv("VCD_RETRY") == "" {
		return operation()
	}
	if operation == nil {
		return nil, fmt.Errorf("argument 'operation' cannot be null")
	}
	start := time.Now()
	elapsed := time.Since(start)
	attempts := 0
	var err error
	var result any
	for elapsed < timeout {
		if refresh != nil {
			err = refresh()
			if err != nil {
				return nil, err
			}
		}
		result, err = operation()
		if err == nil {
			fmt.Printf("%s attempts: %d - elapsed: %s\n", label, attempts, elapsed)
			return result, nil
		}
		elapsed = time.Since(start)
		attempts++
		if elapsed < timeout {
			continue
		}
	}
	return nil, fmt.Errorf(message+" :%s", err)
}
