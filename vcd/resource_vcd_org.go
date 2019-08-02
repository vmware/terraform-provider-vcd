// /*****************************************************************
// * terraform-provider-vcloud-director
// * Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// * SPDX-License-Identifier: BSD-2-Clause
// ******************************************************************/

package vcd

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Organization resource definition
// See details at
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/OrgType.html
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/ReferenceType.html
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/DELETE-Organization.html
func resourceOrg() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrgCreate,
		Read:   resourceOrgRead,
		Update: resourceOrgUpdate,
		Delete: resourceOrgDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdOrgImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"full_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"is_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "True if this organization is enabled (allows login and all other operations).",
			},
			"deployed_vm_quota": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. (0 = unlimited)",
			},
			"stored_vm_quota": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. (0 = unlimited)",
			},
			"can_publish_catalogs": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this organization is allowed to share catalogs.",
			},
			"delay_after_power_on_seconds": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Specifies this organization's default for virtual machine boot delay after power on.",
			},
			"delete_force": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove an org and any objects it contains, regardless of their state.",
			},
			"delete_recursive": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_recursive=True to remove the org and any objects it contains that are in a state that normally allows removal.",
			},
		},
	}
}

// creates an organization based on defined resource
func resourceOrgCreate(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	orgName, fullName, err := getOrgNames(d)
	if err != nil {
		return err
	}
	isEnabled := d.Get("is_enabled").(bool)
	description := d.Get("description").(string)

	settings := getSettings(d)

	log.Printf("Creating Org: %s", orgName)
	task, err := govcd.CreateOrg(vcdClient.VCDClient, orgName, fullName, description, settings, isEnabled)

	if err != nil {
		log.Printf("Error creating Org: %#v", err)
		return fmt.Errorf("error creating Org: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		log.Printf("Error running Org creation task: %s", err)
		return fmt.Errorf("error running Org creation task: %s", err)
	}

	org, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return fmt.Errorf("error retrieving Org %s after creation: %s", orgName, err)
	}
	log.Printf("Org %s created with id: %s", orgName, org.AdminOrg.ID)

	d.SetId(org.AdminOrg.ID)
	return nil
}

func getSettings(d *schema.ResourceData) *types.OrgSettings {
	settings := &types.OrgSettings{}
	General := &types.OrgGeneralSettings{}

	General.DeployedVMQuota = d.Get("deployed_vm_quota").(int)
	General.StoredVMQuota = d.Get("stored_vm_quota").(int)

	delay, ok := d.GetOk("delay_after_power_on_seconds")
	if ok {
		General.DelayAfterPowerOnSeconds = delay.(int)
	}

	General.CanPublishCatalogs = d.Get("can_publish_catalogs").(bool)

	settings.OrgGeneralSettings = General
	return settings
}

// Deletes org
func resourceOrgDelete(d *schema.ResourceData, m interface{}) error {

	//DELETING
	vcdClient := m.(*VCDClient)
	deleteForce := d.Get("delete_force").(bool)
	deleteRecursive := d.Get("delete_recursive").(bool)

	orgName, _, err := getOrgNames(d)
	if err != nil {
		return err
	}

	identifier := d.Id()
	log.Printf("Reading Org %s", identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
	if govcd.ContainsNotFound(err) && isUuid(identifier) {
		adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(orgName)
	}

	if err != nil {
		return fmt.Errorf("error fetching Org %s: %s", orgName, err)
	}

	log.Printf("Org %s found", orgName)
	//deletes organization
	log.Printf("Deleting Org %s", orgName)

	err = adminOrg.Delete(deleteForce, deleteRecursive)
	if err != nil {
		log.Printf("Error deleting org %s: %s", orgName, err)
		return err
	}
	log.Printf("Org %s deleted", orgName)
	return nil
}

// Update the resource
func resourceOrgUpdate(d *schema.ResourceData, m interface{}) error {

	vcdClient := m.(*VCDClient)

	orgName, fullName, err := getOrgNames(d)
	if err != nil {
		return err
	}

	identifier := d.Id()
	log.Printf("Reading Org %s", identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
	if govcd.ContainsNotFound(err) && isUuid(identifier) {
		adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(orgName)
	}

	if err != nil {
		return fmt.Errorf("error fetching Org %s: %s", orgName, err)
	}

	settings := getSettings(d)
	adminOrg.AdminOrg.Name = orgName
	adminOrg.AdminOrg.FullName = fullName
	adminOrg.AdminOrg.Description = d.Get("description").(string)
	adminOrg.AdminOrg.IsEnabled = d.Get("is_enabled").(bool)
	adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings = settings.OrgGeneralSettings

	log.Printf("Org with id %s found", orgName)
	task, err := adminOrg.Update()

	if err != nil {
		log.Printf("Error updating Org %s : %s", orgName, err)
		return fmt.Errorf("error updating Org %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		log.Printf("Error completing update of Org %s : %s", orgName, err)
		return fmt.Errorf("error completing update of Org %s", err)
	}

	log.Printf("Org %s updated", orgName)
	return nil
}

func setOrgData(d *schema.ResourceData, adminOrg *govcd.AdminOrg) error {
	err := d.Set("name", adminOrg.AdminOrg.Name)
	if err != nil {
		return err
	}
	err = d.Set("full_name", adminOrg.AdminOrg.FullName)
	if err != nil {
		return err
	}
	err = d.Set("description", adminOrg.AdminOrg.Description)
	if err != nil {
		return err
	}
	err = d.Set("is_enabled", adminOrg.AdminOrg.IsEnabled)
	if err != nil {
		return err
	}
	err = d.Set("deployed_vm_quota", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota)
	if err != nil {
		return err
	}
	err = d.Set("stored_vm_quota", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.StoredVMQuota)
	if err != nil {
		return err
	}
	err = d.Set("can_publish_catalogs", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs)

	return err
}

// Retrieves an Org resource from vCD
func resourceOrgRead(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	orgName, _, err := getOrgNames(d)
	if err != nil {
		return err
	}

	identifier := d.Id()
	log.Printf("Reading Org %s", identifier)
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	if govcd.ContainsNotFound(err) && isUuid(identifier) {
		// Identifier was created by previous version and it is not a valid ID
		// If the Org is not found by ID, , the ID is invalid, and we have the name in the resource data,
		// we try to access it using the name.
		identifier = orgName
		if identifier != "" {
			log.Printf("Reading Org %s", identifier)
			adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
		}
	}

	if err != nil {
		log.Printf("Org %s not found. Setting ID to nothing", identifier)
		d.SetId("")
		return fmt.Errorf("org %s not found", identifier)
	}
	log.Printf("Org with id %s found", identifier)
	d.SetId(adminOrg.AdminOrg.ID)
	return setOrgData(d, adminOrg)
}

// resourceVcdOrgImport is responsible for importing the resource.
// The d.ID() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
// For this resource, the import path is just the org name.
//
// Example import path (id): orgName
func resourceVcdOrgImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	orgName := d.Id()

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	err = setOrgData(d, adminOrg)

	if err != nil {
		return []*schema.ResourceData{}, err
	}
	return []*schema.ResourceData{d}, nil
}

// Returns name and full_name for an organization, making sure that they are not empty
func getOrgNames(d *schema.ResourceData) (orgName string, fullName string, err error) {
	orgName = d.Get("name").(string)
	fullName = d.Get("full_name").(string)

	if orgName == "" {
		return "", "", fmt.Errorf(`the value for "name" cannot be empty`)
	}
	if fullName == "" {
		return "", "", fmt.Errorf(`the value for "full_name" cannot be empty`)
	}
	return orgName, fullName, nil
}

// Returns true if the identifier is a bare UUID
func isUuid(identifier string) bool {
	reUuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	return reUuid.MatchString(identifier)
}
