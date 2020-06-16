package vcd

import (
	"fmt"
	"log"
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
				Description: "Group name",
			},
			"provider_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // vCD does not allow to change provider type
				Description:  "Identity provider type - 'SAML' or 'INTEGRATED' for LDAP",
				ValidateFunc: validation.StringInSlice([]string{"SAML", "INTEGRATED"}, false),
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description",
			},
			"role": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Existing role name to assign",
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
		Description:  d.Get("description").(string),
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
	if govcd.IsNotFound(err) {
		log.Printf("error finding group %s: %s. Removing from state", d.Id(), err)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error finding group %s: %s", d.Id(), err)
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

	group, err := adminOrg.GetGroupById(d.Id(), false)
	if err != nil {
		return fmt.Errorf("error finding group for update %s: %s", d.Id(), err)
	}

	// Role change
	if d.HasChange("role") {
		roleName := d.Get("role").(string)
		role, err := adminOrg.GetRoleReference(roleName)
		if err != nil {
			return fmt.Errorf("unable to find role %s: %s", roleName, err)
		}
		group.Group.Role = role
	}

	// vCD API and UI at the moment do not update description when provider_type=SAML.
	if d.HasChange("description") {
		group.Group.Description = d.Get("description").(string)
	}

	err = group.Update()
	if err != nil {
		return fmt.Errorf("error updating group %s: %s", group.Group.Name, err)
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
		return fmt.Errorf("error finding group for deletion %s: %s", d.Id(), err)
	}

	err = group.Delete()
	if err != nil {
		return fmt.Errorf("could not delete group %s: %s", group.Group.Name, err)
	}

	return nil
}

// resourceVcdOrgGroupImport imports an org group into Terraform state
// This function task is to get the data from vCD and fill the resource data container
// Expects the d.ID() to be a path to the resource made of Org name + dot + OrgGroup name
//
// Example import path (id): my-org.my-group
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
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
