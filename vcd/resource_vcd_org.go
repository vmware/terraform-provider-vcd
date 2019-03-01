// /*****************************************************************
// * terraform-provider-vcloud-director
// * Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// * SPDX-License-Identifier: BSD-2-Clause
// ******************************************************************/

package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
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
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization.",
			},
			"stored_vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization.",
			},
			"can_publish_catalogs": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this organization is allowed to share catalogs.",
			},
			// "use_server_boot_sequence" was removed, as the docs definition is "This value is ignored."
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
// need to add vdc and network support
func resourceOrgCreate(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	orgName := d.Get("name").(string)
	fullName := d.Get("full_name").(string)
	isEnabled := d.Get("is_enabled").(bool)
	description := d.Get("description").(string)

	settings := getSettings(d)

	log.Printf("Creating Org: %s", orgName)
	task, err := govcd.CreateOrg(vcdClient.VCDClient, orgName, fullName, description, settings, isEnabled)

	if err != nil {
		log.Printf("Error creating Org: %#v", err)
		return fmt.Errorf("error creating Org: %#v", err)
	}

	log.Printf("Org %s Created with id: %s", orgName, task.Task.ID[15:])
	d.SetId(task.Task.ID[15:])
	return nil
}

func getSettings(d *schema.ResourceData) *types.OrgSettings {
	settings := &types.OrgSettings{}
	General := &types.OrgGeneralSettings{}
	if d.Get("deployed_vm_quota").(int) != -1 {
		General.DeployedVMQuota = d.Get("deployed_vm_quota").(int)
	}
	if d.Get("stored_vm_quota").(int) != -1 {
		General.StoredVMQuota = d.Get("stored_vm_quota").(int)
	}

	delay, ok := d.GetOk("delay_after_power_on_seconds")
	if ok {
		General.DelayAfterPowerOnSeconds = delay.(int)
	}

	General.CanPublishCatalogs = d.Get("can_publish_catalogs").(bool)

	settings.OrgGeneralSettings = General
	return settings
}

//Deletes org
func resourceOrgDelete(d *schema.ResourceData, m interface{}) error {

	//DELETING
	vcdClient := m.(*VCDClient)
	deleteForce := d.Get("delete_force").(bool)
	deleteRecursive := d.Get("delete_recursive").(bool)

	//fetches org
	log.Printf("Reading Org with id %s", d.State().ID)
	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))
	if err != nil || org == (govcd.AdminOrg{}) {
		return fmt.Errorf("error fetching Org: %s", d.Get("name").(string))
	}

	log.Printf("Org with id %s found", d.State().ID)
	//deletes organization
	log.Printf("Deleting Org with id %s", d.State().ID)

	err = org.Delete(deleteForce, deleteRecursive)
	if err != nil {
		log.Printf("Error Deleting Org with id %s and error : %#v", d.State().ID, err)
		return err
	}
	log.Printf("Org with id %s deleted", d.State().ID)
	return nil
}

//updated the resource
func resourceOrgUpdate(d *schema.ResourceData, m interface{}) error {

	vcdClient := m.(*VCDClient)

	orgName := d.Get("name").(string)
	oldOrgFullNameRaw, newOrgFullNameRaw := d.GetChange("full_name")
	oldOrgFullName := oldOrgFullNameRaw.(string)
	newOrgFullName := newOrgFullNameRaw.(string)

	if !strings.EqualFold(oldOrgFullName, newOrgFullName) {
		return fmt.Errorf("__ERROR__ Not Updating org_full_name , API NOT IMPLEMENTED !!!!")
	}

	log.Printf("Reading Org WIth id %s", d.State().ID)

	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))

	if err != nil || org == (govcd.AdminOrg{}) {
		return fmt.Errorf("error fetching Org: %s", d.Get("name").(string))
	}

	settings := getSettings(d)
	org.AdminOrg.Name = orgName
	org.AdminOrg.OrgSettings.OrgGeneralSettings = settings.OrgGeneralSettings

	log.Printf("Org with id %s found", d.State().ID)
	_, err = org.Update()

	if err != nil {
		log.Printf("Error updating Org with id %s : %#v", d.State().ID, err)
		return fmt.Errorf("error updating Org %#v", err)
	}

	log.Printf("Org with id %s updated", d.State().ID)
	return nil
}

func resourceOrgRead(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	log.Printf("Reading Org with id %s", d.State().ID)
	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))

	if err != nil || org == (govcd.AdminOrg{}) {
		log.Printf("Org with id %s not found. Setting ID to nothing", d.State().ID)
		d.SetId("")
		return nil
	}
	log.Printf("Org with id %s found", d.State().ID)
	return nil

}
