package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
)

func resourceVcdOrgVdcAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVdcAccessControlCreateUpdate,
		ReadContext:   resourceVcdVdcAccessControlRead,
		UpdateContext: resourceVcdVdcAccessControlCreateUpdate,
		DeleteContext: resourceVcdVdcAccessControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVdcAccessControlImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"shared_with_everyone": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the VDC is shared with everyone",
			},
			"everyone_access_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly}, true),
				Description:  "Access level when the VDC is shared with everyone (only ReadOnly is available). Required when shared_with_everyone is set",
			},
			"shared_with": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the user to which we are sharing. Required if group_id is not set",
						},
						"group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the group to which we are sharing. Required if user_id is not set",
						},
						"subject_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subject (group or user) with which we are sharing",
						},
						"access_level": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly}, true),
							Description:  "The access level for the user or group to which we are sharing. (Only ReadOnly is available)",
						},
					},
				},
			},
		},
	}
}

func resourceVcdVdcAccessControlCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	isSharedWithEveryone := d.Get("shared_with_everyone").(bool)
	everyoneAccessLevel, everyoneAccessLevelSet := d.GetOk("everyone_access_level")
	sharedList := d.Get("shared_with").(*schema.Set).List()

	// Do some checks before proceeding to contact the API
	err := checkParamsVdcAccessControl(isSharedWithEveryone, everyoneAccessLevelSet, sharedList)
	if err != nil {
		return diag.Errorf("error when checking schema - %s", err)
	}

	var accessSettings []*types.AccessSetting

	if !isSharedWithEveryone {
		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return diag.Errorf("error when retrieving AdminOrg - %s", err)
		}

		accessSettings, err = sharedSetToAccessControl(adminOrg, sharedList)
		if err != nil {
			return diag.Errorf("error when reading shared_with from schema - %s", err)
		}
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error when retrieving VDC - %s", err)
	}

	_, err = vdc.SetControlAccess(isSharedWithEveryone, everyoneAccessLevel.(string), accessSettings, true)
	if err != nil {
		return diag.Errorf("error when setting VDC control access parameters - %s", err)
	}

	d.SetId(vdc.Vdc.ID)
	return resourceVcdVdcAccessControlRead(ctx, d, meta)
}

func resourceVcdVdcAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error while reading Org - %s", err)
	}

	vdc, err := org.GetVDCById(d.Id(), false)
	if err != nil {
		if govcd.IsNotFound(err) {
			d.SetId("")
		} else {
			return diag.Errorf("error while reading VDC - %s", err)
		}
	}

	controlAccessParams, err := vdc.GetControlAccess(true)
	if err != nil {
		return diag.Errorf("error getting control access parameters - %s", err)
	}

	dSet(d, "shared_with_everyone", controlAccessParams.IsSharedToEveryone)
	if controlAccessParams.EveryoneAccessLevel != nil {
		dSet(d, "everyone_access_level", *controlAccessParams.EveryoneAccessLevel)
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

func resourceVcdVdcAccessControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// When deleting VDC access control, VDC won't be share with anyone, neither everyone not any single user/group
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		diag.Errorf("error when retrieving VDC - %s", err)
	}

	_, err = vdc.DeleteControlAccess(true)
	if err != nil {
		return diag.Errorf("error when deleting VDC access control - %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdVdcAccessControlImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog")
	}

	orgName, vdcName := resourceURI[0], resourceURI[1]
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	accessControlParams, err := vdc.GetControlAccess(true)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve access control parameters - %s", err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "shared_with_everyone", accessControlParams.IsSharedToEveryone)
	if accessControlParams.EveryoneAccessLevel != nil {
		dSet(d, "everyone_access_level", *accessControlParams.EveryoneAccessLevel)
	} else {
		dSet(d, "everyone_access_level", "")
	}
	if accessControlParams.AccessSettings != nil {
		accessControlListSet, err := accessControlListToSharedSet(accessControlParams.AccessSettings.AccessSetting)
		if err != nil {
			return nil, fmt.Errorf("error converting slice AccessSetting into set - %s", err)
		}

		err = d.Set("shared_with", accessControlListSet)
		if err != nil {
			return nil, fmt.Errorf("error setting shared_with attribute - %s", err)
		}
	}

	d.SetId(vdc.Vdc.ID)

	return []*schema.ResourceData{d}, nil
}

func checkParamsVdcAccessControl(isSharedWithEveryone bool, everyoneAccessLevelSet bool, sharedList []interface{}) error {
	if isSharedWithEveryone && len(sharedList) > 0 {
		return fmt.Errorf("if shared_with_everyone is set to true, shared_with must be an empty set")
	}

	if !isSharedWithEveryone && len(sharedList) == 0 {
		return fmt.Errorf("if shared_with_everyone is set to false, shared_with must contain at least one user/group")
	}

	if isSharedWithEveryone && !everyoneAccessLevelSet {
		return fmt.Errorf("if shared_with_everyone is set to true, everyone_access_level needs to be set")
	}

	return nil
}
