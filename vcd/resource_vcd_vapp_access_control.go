package vcd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdAccessControlVapp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessControlVappCreate,
		ReadContext:   resourceAccessControlVappRead,
		UpdateContext: resourceAccessControlVappUpdate,
		DeleteContext: resourceAccessControlVappDelete,
		Importer: &schema.ResourceImporter{
			StateContext: accessControlVappImport,
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
			"vapp_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp identifier",
			},
			"shared_with_everyone": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the vApp is shared with everyone",
			},
			"everyone_access_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly, types.ControlAccessReadWrite, types.ControlAccessFullControl}, true),
				Description:  "Access level when the vApp is shared with everyone (one of ReadOnly, Change, FullControl). Required when shared_with_everyone is set",
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
							ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly, types.ControlAccessReadWrite, types.ControlAccessFullControl}, true),
							Description:  "The access level for the user or group to which we are sharing. (One of ReadOnly, Change, FullControl)",
						},
					},
				},
			},
		},
	}
}

// tenantContext defines whether we run access control operations in the context of the original caller.
// By default it is ON (= run as tenant). We can turn it off by setting the environment variable VCD_ORIGINAL_CONTEXT.
var tenantContext = os.Getenv("VCD_ORIGINAL_CONTEXT") == ""

func resourceAccessControlVappCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceAccessControlVappUpdate(ctx, d, meta)
}

func resourceAccessControlVappUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vcdClient := meta.(*VCDClient)

	var accessControl types.ControlAccessParams

	isSharedWithEveryone := d.Get("shared_with_everyone").(bool)
	everyoneAccessLevel := d.Get("everyone_access_level").(string)
	sharedList := d.Get("shared_with").(*schema.Set).List()

	// Early checks, so that we can fail as soon as possible
	if isSharedWithEveryone {
		if everyoneAccessLevel == "" {
			return diag.Errorf("[resourceAccessControlVappUpdate] 'shared_with_everyone' was set, but 'everyone_access_level' was not")
		}
		accessControl.IsSharedToEveryone = true
		accessControl.EveryoneAccessLevel = &everyoneAccessLevel
		if len(sharedList) > 0 {
			return diag.Errorf("[resourceAccessControlVappUpdate] when 'shared_with_everyone' is true, 'shared_with' must not be filled")
		}
	} else {
		if everyoneAccessLevel != "" {
			return diag.Errorf("[resourceAccessControlVappUpdate] if 'shared_with_everyone' is false, we can't set 'everyone_access_level'")
		}
	}

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	adminOrg, err := vcdClient.GetAdminOrgById(org.Org.ID)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}
	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return diag.Errorf("[resourceAccessControlVappUpdate] error finding vApp %s. %s", vappId, err)
	}
	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	if !isSharedWithEveryone {
		accessControlList, err := sharedSetToAccessControl(vcdClient, adminOrg, sharedList, []string{"user_id", "group_id"})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(accessControlList) > 0 {
			accessControl.AccessSettings = &types.AccessSettingList{
				AccessSetting: accessControlList,
			}
		}
	}

	err = vapp.SetAccessControl(&accessControl, tenantContext)

	if err != nil {
		return diag.Errorf("[resourceAccessControlVappUpdate] error setting access control for vApp %s: %s", vapp.VApp.Name, err)
	}

	return resourceAccessControlVappRead(ctx, d, meta)
}

func resourceAccessControlVappRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			// If vApp is not found - it means that Access Control must be recreated as well
			log.Printf("parent vApp not found. Removing access control from state file: %s", err)
			d.SetId("")
			return nil
		}
		return diag.Errorf("[resourceAccessControlVappRead] error retrieving vApp %s. %s", vappId, err)
	}

	accessControl, err := vapp.GetAccessControl(tenantContext)
	if err != nil {
		return diag.Errorf("[resourceAccessControlVappRead] error retrieving access control for vApp %s : %s", vapp.VApp.Name, err)
	}

	if accessControl.AccessSettings != nil {
		sharedList, err := accessControlListToSharedSet(accessControl.AccessSettings.AccessSetting)
		if err != nil {
			return diag.Errorf("[resourceAccessControlVappRead] error converting access control list %s", err)
		}
		err = d.Set("shared_with", sharedList)
		if err != nil {
			return diag.Errorf("[resourceAccessControlVappRead] error setting access control list %s", err)
		}
	}
	dSet(d, "vapp_id", vapp.VApp.ID)
	dSet(d, "shared_with_everyone", accessControl.IsSharedToEveryone)
	if accessControl.IsSharedToEveryone {
		dSet(d, "everyone_access_level", accessControl.EveryoneAccessLevel)
	}
	d.SetId(vapp.VApp.ID)

	return nil
}

func resourceAccessControlVappDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}
	err = vapp.RemoveAccessControl(tenantContext)
	if err != nil {
		return diag.Errorf("error removing access control for vApp %s: %s", vapp.VApp.Name, err)
	}
	d.SetId("")
	return nil
}

func accessControlVappImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[vApp access control import] resource identifier must be specified as org.vdc.my-vapp")
	}
	listRequested := false
	orgName, vdcName, vappIdentifier := resourceURI[0], resourceURI[1], resourceURI[2]
	if strings.HasPrefix(orgName, "list@") {
		listRequested = true
		orgNameList := strings.Split(orgName, "@")
		if len(orgNameList) < 2 {
			return nil, fmt.Errorf("[vApp access control import] empty Org name provided with list@ request")
		}
		orgName = orgNameList[1]
	}
	if orgName == "" {
		return nil, fmt.Errorf("[vApp access control import] empty org name provided")
	}
	if vdcName == "" {
		return nil, fmt.Errorf("[vApp access control import] empty VDC name provided")
	}
	if vappIdentifier == "" {
		return nil, fmt.Errorf("[vApp access control import] empty vApp access control identifier provided")
	}

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	vdc, err := adminOrg.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("[vApp access control import] error retrieving VDC %s, %s", vdcName, err)
	}

	if listRequested {
		var vappList = vdc.GetVappList()
		return nil, fmt.Errorf("[vApp access control import] list of all vApps:\n%s", formatVappList(vappList))
	}

	var vapp *govcd.VApp
	vapp, err = vdc.GetVAppByNameOrId(vappIdentifier, false)

	if err != nil {
		return nil, fmt.Errorf("[vApp access control import] error retrieving vapp %s: %s",
			vappIdentifier, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "vapp_id", vapp.VApp.ID)
	d.SetId(vapp.VApp.ID)

	return []*schema.ResourceData{d}, nil
}

func formatVappList(list []*types.ResourceReference) string {
	result := ""
	for i, ref := range list {
		result += fmt.Sprintf("%3d %-30s %s\n", i, ref.Name, ref.ID)
	}
	return result
}

func accessControlListToSharedSet(input []*types.AccessSetting) ([]map[string]interface{}, error) {
	var output []map[string]interface{}

	for _, item := range input {
		var setting = make(map[string]interface{})

		switch item.Subject.Type {
		case types.MimeAdminUser:
			setting["user_id"] = "urn:vcloud:user:" + extractUuid(item.Subject.HREF)
		case types.MimeAdminGroup:
			setting["group_id"] = extractUuid(item.Subject.HREF)
		case types.MimeOrg, types.MimeAdminOrg:
			setting["org_id"] = "urn:vcloud:org:" + extractUuid(item.Subject.HREF)
		default:
			return nil, fmt.Errorf("unhandled type '%s' for item %s", item.Subject.Type, item.Subject.Name)
		}
		setting["access_level"] = item.AccessLevel
		setting["subject_name"] = item.Subject.Name

		output = append(output, setting)
	}
	return output, nil
}

func sharedSetToAccessControl(client *VCDClient, org *govcd.AdminOrg, input []interface{}, validIds []string) ([]*types.AccessSetting, error) {
	var output []*types.AccessSetting

	for _, item := range input {
		usedUp := false
		setting, ok := item.(map[string]interface{})
		if !ok {
			return output, fmt.Errorf("item is not a string map %#v", item)
		}
		var subjectHref string
		var subjectType string
		var subjectName string
		var orgId string

		for _, id := range validIds {
			switch id {
			case "user_id":
				userId, ok := setting[id].(string)
				if ok && userId != "" {
					if usedUp {
						return nil, fmt.Errorf("only one of %v IDs can be used", validIds)
					}
					user, err := org.GetUserById(userId, false)
					if err != nil {
						return nil, fmt.Errorf("error retrieving user %s: %s", userId, err)
					}
					usedUp = true
					subjectHref = user.User.Href
					subjectType = user.User.Type
					subjectName = user.User.Name
				}
			case "group_id":
				groupId, ok := setting["group_id"].(string)
				if ok && groupId != "" {
					if usedUp {
						return nil, fmt.Errorf("only one of %v IDs can be used", validIds)
					}
					group, err := org.GetGroupById(groupId, false)
					if err != nil {
						return nil, fmt.Errorf("error retrieving group %s: %s", groupId, err)
					}
					usedUp = true
					subjectHref = group.Group.Href
					subjectType = group.Group.Type
					subjectName = group.Group.Name
				}
			case "org_id":
				orgId, ok = setting[id].(string)
				if ok && orgId != "" {
					if usedUp {
						return nil, fmt.Errorf("only one of %v IDs can be used", validIds)
					}
					org, err := client.GetOrgById(orgId)
					if err != nil {
						return nil, fmt.Errorf("error retrieving Org %s: %s", orgId, err)
					}
					usedUp = true
					subjectHref = org.Org.HREF
					subjectType = org.Org.Type
					subjectName = org.Org.Name
				}
			default:
				return nil, fmt.Errorf("[sharedFullSetToAccessControl] invalid ID %s", id)
			}
		}
		if !usedUp {
			return nil, fmt.Errorf("[sharedFullSetToAccessControl] no filled ID found among %v for entry %#v", validIds, item)
		}
		if subjectHref == "" {
			return nil, fmt.Errorf("no org, group, or user found for entry %#v", item)
		}
		accessLevel := setting["access_level"].(string)
		if orgId != "" && accessLevel != types.ControlAccessReadOnly {
			return nil, fmt.Errorf("access level for an Organization can only be %s", types.ControlAccessReadOnly)
		}

		output = append(output, &types.AccessSetting{
			Subject: &types.LocalSubject{
				HREF: subjectHref,
				Name: subjectName,
				Type: subjectType,
			},
			ExternalSubject: nil,
			AccessLevel:     accessLevel,
		})
	}
	return output, nil
}
