package vcd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdAccessControlVapp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessControlVappCreate,
		Read:   resourceAccessControlVappRead,
		Update: resourceAccessControlVappUpdate,
		Delete: resourceAccessControlVappDelete,
		Importer: &schema.ResourceImporter{
			State: accessControlVappImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"vapp_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp identifier",
			},
			"shared_with_everyone": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the vApp is shared with everyone",
			},
			"everyone_access_level": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{types.ControlAccessReadOnly, types.ControlAccessReadWrite, types.ControlAccessFullControl}, true),
				Description:  "Access level when the vApp is shared with everyone (one of ReadOnly, Change, FullControl). Required when shared_with_everyone is set",
			},
			"shared_with": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the user to which we are sharing. Required if group_id is not set",
						},
						"group_id": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the group to which we are sharing. Required if user_id is not set",
						},
						"subject_name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subject (group or user) with which we are sharing",
						},
						"access_level": &schema.Schema{
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

func resourceAccessControlVappCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAccessControlVappUpdate(d, meta)
}

func resourceAccessControlVappUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	var accessControl types.ControlAccessParams

	isSharedToEveryone := d.Get("shared_with_everyone").(bool)
	everyoneAccessLevel := d.Get("everyone_access_level").(string)
	sharedList := d.Get("shared_with").(*schema.Set).List()

	// Early checks, so that we can fail as soon as possible
	if isSharedToEveryone {
		accessControl.IsSharedToEveryone = true
		accessControl.EveryoneAccessLevel = &everyoneAccessLevel
		if len(sharedList) > 0 {
			return fmt.Errorf("[resourceAccessControlVappUpdate] when 'shared_with_everyone' is true, 'shared_with' must not be filled")
		}
	} else {
		if everyoneAccessLevel != "" {
			return fmt.Errorf("[resourceAccessControlVappUpdate] if 'shared_with_everyone' is false, we can't set 'everyone_access_level'")
		}
	}

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	adminOrg, err := vcdClient.GetAdminOrgById(org.Org.ID)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}
	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return fmt.Errorf("[resourceAccessControlVappUpdate] error finding vApp %s. %s", vappId, err)
	}
	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	if !isSharedToEveryone {
		accessControlList, err := sharedSetToAccessControl(adminOrg, sharedList)
		if err != nil {
			return err
		}
		if len(accessControlList) > 0 {
			accessControl.AccessSettings = &types.AccessSettingList{
				AccessSetting: accessControlList,
			}
		}
	}

	err = vapp.SetAccessControl(&accessControl, tenantContext)

	if err != nil {
		return fmt.Errorf("[resourceAccessControlVappUpdate] error setting access control for vApp %s: %s", vapp.VApp.Name, err)
	}

	return resourceAccessControlVappRead(d, meta)
}

func resourceAccessControlVappRead(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return fmt.Errorf("[resourceAccessControlVappRead] error retrieving vApp %s. %s", vappId, err)
	}

	accessControl, err := vapp.GetAccessControl(tenantContext)
	if err != nil {
		return fmt.Errorf("[resourceAccessControlVappRead] error retrieving access control for vApp %s : %s", vapp.VApp.Name, err)
	}

	if accessControl.AccessSettings != nil {
		sharedList, err := accessControlListToSharedSet(accessControl.AccessSettings.AccessSetting)
		if err != nil {
			return fmt.Errorf("[resourceAccessControlVappRead] error converting access control list %s", err)
		}
		err = d.Set("shared_with", sharedList)
		if err != nil {
			return fmt.Errorf("[resourceAccessControlVappRead] error setting access control list %s", err)
		}
	}
	_ = d.Set("vapp_id", vapp.VApp.ID)
	_ = d.Set("shared_with_everyone", accessControl.IsSharedToEveryone)
	if accessControl.IsSharedToEveryone {
		_ = d.Set("everyone_access_level", accessControl.EveryoneAccessLevel)
	}
	d.SetId(vapp.VApp.ID)

	return nil
}

func resourceAccessControlVappDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}
	err = vapp.RemoveAccessControl(tenantContext)
	if err != nil {
		return fmt.Errorf("error removing access control for vApp %s: %s", vapp.VApp.Name, err)
	}
	d.SetId("")
	return nil
}

func accessControlVappImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
		var vappList []*types.ResourceReference = vdc.GetVappList()
		return nil, fmt.Errorf("[vApp access control import] list of all vApps:\n%s", formatVappList(vappList))
	}

	var vapp *govcd.VApp
	vapp, err = vdc.GetVAppByNameOrId(vappIdentifier, false)

	if err != nil {
		return nil, fmt.Errorf("[vApp access control import] error retrieving vapp %s: %s",
			vappIdentifier, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("vapp_id", vapp.VApp.ID)
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
		case types.MimeOrg:
			setting["org_id"] = extractUuid(item.Subject.HREF)
		default:
			return nil, fmt.Errorf("unhandled type '%s' for item %s", item.Subject.Type, item.Subject.Name)
		}
		setting["access_level"] = item.AccessLevel
		setting["subject_name"] = item.Subject.Name

		output = append(output, setting)
	}
	return output, nil
}

func sharedSetToAccessControl(org *govcd.AdminOrg, input []interface{}) ([]*types.AccessSetting, error) {
	var output []*types.AccessSetting
	for _, item := range input {
		setting, ok := item.(map[string]interface{})
		if !ok {
			return output, fmt.Errorf("item is not a string map %#v", item)
		}
		var subjectHref string
		var subjectType string
		var subjectName string

		userId, ok := setting["user_id"].(string)
		if ok && userId != "" {
			user, err := org.GetUserById(userId, false)
			if err != nil {
				return nil, fmt.Errorf("error retrieving user %s: %s", userId, err)
			}
			subjectHref = user.User.Href
			subjectType = user.User.Type
			subjectName = user.User.Name
		}

		groupId, ok := setting["group_id"].(string)
		if ok && groupId != "" {
			if userId != "" {
				return nil, fmt.Errorf("either user ID or group ID can be set, not both")
			}
			group, err := org.GetGroupById(groupId, false)
			if err != nil {
				return nil, fmt.Errorf("error retrieving group %s: %s", groupId, err)
			}
			subjectHref = group.Group.Href
			subjectType = group.Group.Type
			subjectName = group.Group.Name
		}
		if subjectHref == "" {
			return nil, fmt.Errorf("no group or user found for entry %#v", item)
		}
		accessLevel := setting["access_level"].(string)

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
