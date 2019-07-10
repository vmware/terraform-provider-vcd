package vcd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdOrgUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdOrgUserCreate,
		Read:   resourceVcdOrgUserRead,
		Delete: resourceVcdOrgUserDelete,
		Update: resourceVcdOrgUserUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceVcdOrgUserImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCase("lower"),
				Description:  "User's name. Only lowercase letters allowed. Cannot be changed after creation",
			},
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization this user belongs to",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Role within the organization",
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,

				ConflictsWith: []string{"password_file"},
				Description: "The user's password. This value is never returned on read. " +
					`Either "password" or "password_file" must be included on creation.`,
			},
			"password_file": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"password"},
				Description: "Name of a file containing the user's password. " +
					`Either "password_file" or "password" must be included on creation.`,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The Org User's description",
			},
			"provider_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     govcd.OrgUserProviderIntegrated,
				ForceNew:    false,
				Description: "Identity provider type for this this user. One of: 'INTEGRATED', 'SAML', 'OAUTH'. " +
					"When empty, the default value 'INTEGRATED' is used.",
			},
			"full_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The Org User's full name",
			},
			"email_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The Org User's email address",
			},
			"telephone": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The Org User's telephone",
			},
			"instant_messaging": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The Org User's telephone",
			},
			"is_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    false,
				Description: "True if the user is enabled and can log in.",
			},
			"is_group_role": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    false,
				Description: "True if this user has a group role.",
			},
			"is_locked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Description: "True if the user account has been locked due to too many invalid login attempts. " +
					"A locked user account can be re-enabled by updating the user with this flag set to false. " +
					"Only the system can set the value to true.",
			},
			"take_ownership": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    false,
				Description: "Take ownership of user's objects on deletion.",
			},
			"deployed_vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				ForceNew:    false,
				Description: "Quota of vApps that this user can deploy. A value of 0 specifies an unlimited quota.",
			},
			"stored_vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				ForceNew:    false,
				Description: "Quota of vApps that this user can store. A value of 0 specifies an unlimited quota.",
			},
		},
	}
}

func resourceToUserData(d *schema.ResourceData, meta interface{}) (*govcd.OrgUserConfiguration, *govcd.AdminOrg, error) {
	vcdClient := meta.(*VCDClient)
	orgName := d.Get("org").(string)
	if orgName == "" {
		orgName = vcdClient.Org
	}
	if orgName == "" {
		return nil, nil, fmt.Errorf("missing org name")
	}
	adminOrg, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, orgName)
	if err != nil {
		return nil, nil, err
	}
	if adminOrg.AdminOrg == nil || adminOrg.AdminOrg.HREF == "" {
		return nil, nil, fmt.Errorf("error retrieving org %s", orgName)
	}

	var userData govcd.OrgUserConfiguration
	userData.RoleName = d.Get("role").(string)

	userData.Name = d.Get("name").(string)
	userData.Description = d.Get("description").(string)
	userData.FullName = d.Get("full_name").(string)
	userData.EmailAddress = d.Get("email_address").(string)
	userData.Telephone = d.Get("telephone").(string)
	userData.ProviderType = d.Get("provider_type").(string)
	userData.IsEnabled = d.Get("is_enabled").(bool)
	userData.IsLocked = d.Get("is_locked").(bool)
	userData.DeployedVmQuota = d.Get("deployed_vm_quota").(int)
	userData.StoredVmQuota = d.Get("stored_vm_quota").(int)
	userData.IM = d.Get("instant_messaging").(string)

	password := d.Get("password").(string)
	if password != "" {
		userData.Password = password
	}
	passwordFile := d.Get("password_file").(string)
	if passwordFile != "" {
		passwordBytes, err := ioutil.ReadFile(passwordFile)
		if err != nil {
			return nil, nil, err
		}
		passwordStr := strings.TrimSpace(string(passwordBytes))
		if passwordStr != "" {
			userData.Password = passwordStr
		}
	}

	if userData.Password == "" {
		return nil, nil, fmt.Errorf(`no password provided in either "password"" or "password_file"`)
	}
	return &userData, &adminOrg, nil
}

func resourceToOrgUser(d *schema.ResourceData, meta interface{}) (*govcd.OrgUser, *govcd.AdminOrg, error) {

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, nil, err
	}
	if adminOrg.AdminOrg == nil || adminOrg.AdminOrg.HREF == "" {
		return nil, nil, fmt.Errorf("error retrieving org %s", d.Get("org").(string))
	}
	userName := d.Get("name").(string)
	orgUser, err := adminOrg.FetchUserByName(userName, false)
	if err != nil {
		return nil, nil, err
	}

	return orgUser, &adminOrg, nil
}

func setOrgUserData(d *schema.ResourceData, orgUser *govcd.OrgUser, adminOrg *govcd.AdminOrg) error {

	d.SetId(orgUser.User.ID)
	err := d.Set("org", adminOrg.AdminOrg.Name)
	if err != nil {
		return err
	}
	err = d.Set("name", orgUser.User.Name)
	if err != nil {
		return err
	}
	err = d.Set("provider_type", orgUser.User.ProviderType)
	if err != nil {
		return err
	}
	err = d.Set("is_group_role", orgUser.User.IsGroupRole)
	if err != nil {
		return err
	}
	err = d.Set("description", orgUser.User.Description)
	if err != nil {
		return err
	}
	err = d.Set("full_name", orgUser.User.FullName)
	if err != nil {
		return err
	}
	err = d.Set("email_address", orgUser.User.EmailAddress)
	if err != nil {
		return err
	}
	err = d.Set("telephone", orgUser.User.Telephone)
	if err != nil {
		return err
	}
	err = d.Set("instant_messaging", orgUser.User.IM)
	if err != nil {
		return err
	}
	err = d.Set("is_enabled", orgUser.User.IsEnabled)
	if err != nil {
		return err
	}
	err = d.Set("is_locked", orgUser.User.IsLocked)
	if err != nil {
		return err
	}
	err = d.Set("deployed_vm_quota", orgUser.User.DeployedVmQuota)
	if err != nil {
		return err
	}
	err = d.Set("stored_vm_quota", orgUser.User.StoredVmQuota)
	if err != nil {
		return err
	}
	err = d.Set("role", orgUser.User.Role.Name)
	if err != nil {
		return err
	}
	return nil
}

func resourceVcdOrgUserCreate(d *schema.ResourceData, meta interface{}) error {

	userData, adminOrg, err := resourceToUserData(d, meta)
	if err != nil {
		return err
	}
	_, err = adminOrg.CreateUserSimple(*userData)
	if err != nil {
		return err
	}
	return resourceVcdOrgUserRead(d, meta)
}

func resourceVcdOrgUserDelete(d *schema.ResourceData, meta interface{}) error {

	takeOwnership := d.Get("take_ownership").(bool)
	orgUser, _, err := resourceToOrgUser(d, meta)
	if err != nil {
		return err
	}
	return orgUser.Delete(takeOwnership)
}

func resourceVcdOrgUserRead(d *schema.ResourceData, meta interface{}) error {

	orgUser, adminOrg, err := resourceToOrgUser(d, meta)
	if err != nil {
		return err
	}
	return setOrgUserData(d, orgUser, adminOrg)
}

func resourceVcdOrgUserUpdate(d *schema.ResourceData, meta interface{}) error {

	orgUser, _, err := resourceToOrgUser(d, meta)
	if err != nil {
		return err
	}
	userData, _, err := resourceToUserData(d, meta)
	if err != nil {
		return err
	}
	err = orgUser.UpdateSimple(*userData)
	if err != nil {
		return err
	}
	return resourceVcdOrgUserRead(d, meta)
}

func resourceVcdOrgUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.org_user")
	}
	orgName, userName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil || adminOrg == (govcd.AdminOrg{}) {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	user, err := adminOrg.FetchUserByName(userName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	err = setOrgUserData(d, user, &adminOrg)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
