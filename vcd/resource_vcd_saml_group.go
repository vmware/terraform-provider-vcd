package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdOrgSamlGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdOrgSamlGroupCreate,
		Read:   resourceVcdOrgSamlGroupRead,
		Update: resourceVcdOrgSamlGroupUpdate,
		Delete: resourceVcdOrgSamlGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdOrgSamlGroupImport,
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

func resourceVcdOrgSamlGroupCreate(d *schema.ResourceData, meta interface{}) error {
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
		ProviderType: govcd.OrgUserProviderSAML, // 'SAML' is the only accepted. Others get HTTP 403
	}
	newGroup.Group = &groupDefinition

	createdGroup, err := adminOrg.CreateGroup(newGroup.Group)
	if err != nil {
		return fmt.Errorf("error creating group %s: %s", groupDefinition.Name, err)
	}

	d.SetId(createdGroup.Group.ID)

	return resourceVcdOrgSamlGroupRead(d, meta)
}

// resourceVcdOrgSamlGroupUpdate
func resourceVcdOrgSamlGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdOrgSamlGroupRead(d, meta)
}

// resourceVcdOrgSamlGroupRead
func resourceVcdOrgSamlGroupRead(d *schema.ResourceData, meta interface{}) error {
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

	return nil
}

// resourceVcdOrgSamlGroupDelete
func resourceVcdOrgSamlGroupDelete(d *schema.ResourceData, meta interface{}) error {
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

// resourceVcdOrgSamlGroupImport
func resourceVcdOrgSamlGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
