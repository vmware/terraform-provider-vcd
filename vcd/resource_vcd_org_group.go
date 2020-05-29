package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdOrgGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdOrgGroupCreate,
		Read:   resourceVcdOrgGroupRead,
		Update: resourceVcdOrgGroupUpdate,
		Delete: resourceVcdOrgGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdOrgGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // vCD does not allow to change group name
				Description: "SAML group name",
			},
			"provider_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // vCD does not allow to change provider type
				Description:  "SAML group name",
				ValidateFunc: validation.StringInSlice([]string{"SAML", "INTEGRATED"}, false),
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true, // vCD does not set the description
				Description: "Description",
			},
			"role": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name to assign",
			},
		},
	}
}

func resourceVcdOrgGroupCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	roleName := d.Get("role").(string)
	role, err := adminOrg.GetRoleReference(roleName)
	if err != nil {
		return fmt.Errorf("unable to find role %s: %s", roleName, err)
	}

	newGroup := govcd.NewGroup(&vcdClient.Client, adminOrg)
	groupDefinition := types.Group{
		Name:         d.Get("name").(string),
		Role:         role,
		ProviderType: d.Get("provider_type").(string),
	}
	newGroup.Group = &groupDefinition

	createdGroup, err := adminOrg.CreateGroup(newGroup.Group)
	if err != nil {
		return fmt.Errorf("error creating group %s: %s", groupDefinition.Name, err)
	}

	d.SetId(createdGroup.Group.ID)

	return resourceVcdOrgGroupRead(d, meta)
}

func resourceVcdOrgGroupRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	group, err := adminOrg.GetGroupById(d.Id(), false)
	if err != nil {
		return fmt.Errorf("error finding group for deletion %s: %s", group.Group.Name, err)
	}

	d.Set("name", group.Group.Name)
	d.Set("description", group.Group.Description)
	d.Set("role", group.Group.Role.Name)
	d.Set("provider_type", group.Group.ProviderType)

	return nil
}

func resourceVcdOrgGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	// The only possible change for now is 'role'
	if d.HasChange("role") {
		group, err := adminOrg.GetGroupById(d.Id(), false)
		if err != nil {
			return fmt.Errorf("error finding group for update %s: %s", group.Group.Name, err)
		}

		roleName := d.Get("role").(string)
		role, err := adminOrg.GetRoleReference(roleName)
		if err != nil {
			return fmt.Errorf("unable to find role %s: %s", roleName, err)
		}

		group.Group.Role = role
		err = group.Update()

		if err != nil {
			return fmt.Errorf("error updating group %s: %s", group.Group.Name, err)
		}
	}

	return resourceVcdOrgGroupRead(d, meta)
}

func resourceVcdOrgGroupDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	group, err := adminOrg.GetGroupById(d.Id(), false)
	if err != nil {
		return fmt.Errorf("error finding group for deletion %s: %s", group.Group.Name, err)
	}

	err = group.Delete()
	if err != nil {
		return fmt.Errorf("could not delete group %s: %s", group.Group.Name, err)
	}

	return nil
}

func resourceVcdOrgGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.org_group")
	}
	orgName, groupName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	group, err := adminOrg.GetGroupByName(groupName, false)
	if err != nil {
		return nil, fmt.Errorf("[group import] error retrieving group %s: %s", groupName, err)
	}

	d.SetId(group.Group.ID)
	return []*schema.ResourceData{d}, nil
}
