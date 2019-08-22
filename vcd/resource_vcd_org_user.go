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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Sensitive:     true,
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
				Description: "The user's description",
			},
			"provider_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  govcd.OrgUserProviderIntegrated,
				ForceNew: false,
				Description: "Identity provider type for this this user. One of: 'INTEGRATED', 'SAML', 'OAUTH'. " +
					"When empty, the default value 'INTEGRATED' is used.",
			},
			"full_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The user's full name",
			},
			"email_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The user's email address",
			},
			"telephone": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The user's telephone",
			},
			"instant_messaging": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The user's telephone",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
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
				Description: "If the user account has been locked due to too many invalid login attempts, the value " +
					"will change to true (only the system can lock the user). " +
					"To unlock the user re-set this flag to false.",
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

// Converts resource data into a OrgUserConfiguration structure
// used for both creation and update.
func resourceToUserData(d *schema.ResourceData, meta interface{}) (*govcd.OrgUserConfiguration, *govcd.AdminOrg, error) {
	vcdClient := meta.(*VCDClient)
	orgName := d.Get("org").(string)
	if orgName == "" {
		orgName = vcdClient.Org
	}
	if orgName == "" {
		return nil, nil, fmt.Errorf("missing org name")
	}
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByName(orgName)
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
	userData.IsEnabled = d.Get("enabled").(bool)
	userData.IsLocked = d.Get("is_locked").(bool)
	userData.DeployedVmQuota = d.Get("deployed_vm_quota").(int)
	userData.StoredVmQuota = d.Get("stored_vm_quota").(int)
	userData.IM = d.Get("instant_messaging").(string)

	password := d.Get("password").(string)
	if password != "" {
		userData.Password = password
	}
	passwordFile := d.Get("password_file").(string)

	if password != "" && passwordFile != "" {
		return nil, nil, fmt.Errorf(`either "password" or "password_file" should be given, but not both`)
	}

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

	return &userData, adminOrg, nil
}

// Retrieve an OrgUser and an AdminOrg from the data in the resource.
// Used wherever we need to read the object from vCD with input provided in the resource fields
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
	orgUser, err := adminOrg.GetUserByName(userName, false)
	if err != nil {
		return nil, nil, err
	}

	return orgUser, adminOrg, nil
}

// Fills a ResourceData container with data retrieved from an OrgUser and an AdminOrg
// Used after retrieving the user (read, import), to fill the Terraform container appropriately
func setOrgUserData(d *schema.ResourceData, orgUser *govcd.OrgUser, adminOrg *govcd.AdminOrg) error {
	d.SetId(orgUser.User.ID)
	_ = d.Set("org", adminOrg.AdminOrg.Name)
	_ = d.Set("name", orgUser.User.Name)
	_ = d.Set("provider_type", orgUser.User.ProviderType)
	_ = d.Set("is_group_role", orgUser.User.IsGroupRole)
	_ = d.Set("description", orgUser.User.Description)
	_ = d.Set("full_name", orgUser.User.FullName)
	_ = d.Set("email_address", orgUser.User.EmailAddress)
	_ = d.Set("telephone", orgUser.User.Telephone)
	_ = d.Set("instant_messaging", orgUser.User.IM)
	_ = d.Set("enabled", orgUser.User.IsEnabled)
	_ = d.Set("is_locked", orgUser.User.IsLocked)
	_ = d.Set("deployed_vm_quota", orgUser.User.DeployedVmQuota)
	_ = d.Set("stored_vm_quota", orgUser.User.StoredVmQuota)
	_ = d.Set("role", orgUser.User.Role.Name)
	return nil
}

// Creates an OrgUser from data provided in the resource
func resourceVcdOrgUserCreate(d *schema.ResourceData, meta interface{}) error {

	userData, adminOrg, err := resourceToUserData(d, meta)
	if err != nil {
		return err
	}
	if userData.Password == "" {
		return fmt.Errorf(`no password provided with either "password"" or "password_file" properties`)
	}
	_, err = adminOrg.CreateUserSimple(*userData)
	if err != nil {
		return err
	}
	return resourceVcdOrgUserRead(d, meta)
}

// Deletes an OrgUser
func resourceVcdOrgUserDelete(d *schema.ResourceData, meta interface{}) error {

	takeOwnership := d.Get("take_ownership").(bool)
	orgUser, _, err := resourceToOrgUser(d, meta)
	if err != nil {
		return err
	}
	return orgUser.Delete(takeOwnership)
}

// Reads the OrgUser from vCD and fills the resource container appropriately
func resourceVcdOrgUserRead(d *schema.ResourceData, meta interface{}) error {

	orgUser, adminOrg, err := resourceToOrgUser(d, meta)
	if err != nil {
		return err
	}
	return setOrgUserData(d, orgUser, adminOrg)
}

// Updates an OrgUser with the data passed through the resource
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

// Imports an OrgUser into Terraform state
// This function task is to get the data from vCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of Org name + dot + OrgUser name
//
// Example import path (id): my-org.my-user-admin
func resourceVcdOrgUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.org_user")
	}
	orgName, userName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	user, err := adminOrg.GetUserByName(userName, false)
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	err = setOrgUserData(d, user, adminOrg)
	if err != nil {
		return nil, err
	}

	d.SetId(user.User.ID)
	return []*schema.ResourceData{d}, nil
}
