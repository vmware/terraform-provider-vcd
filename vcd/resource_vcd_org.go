// /*****************************************************************
// * terraform-provider-vcloud-director
// * Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// * SPDX-License-Identifier: BSD-2-Clause
// ******************************************************************/

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	govcd "github.com/vmware/go-vcloud-director/govcd"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"log"
	"strings"
)

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

			"is_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"deployed_vm_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"stored_vm_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"can_publish_catalogs": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"use_server_boot_sequence": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"delay_after_power_on_seconds": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"force": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"recursive": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
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

	settings := getSettings(d)

	log.Printf("CREATING ORG: %s", orgName)
	task, err := govcd.CreateOrg(vcdClient.VCDClient, orgName, fullName, isEnabled, settings)

	if err != nil {
		log.Printf("Error creating organization: %#v", err)
		return fmt.Errorf("Error creating organization: %#v", err)
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
	General.UseServerBootSequence = d.Get("use_server_boot_sequence").(bool)

	settings.OrgGeneralSettings = General
	return settings
}

//Deletes org
func resourceOrgDelete(d *schema.ResourceData, m interface{}) error {

	//DELETING
	vcdClient := m.(*VCDClient)
	force := d.Get("force").(bool)
	recursive := d.Get("recursive").(bool)

	//fetches org
	log.Printf("Reading org with id %s", d.State().ID)
	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))
	if err != nil || org == (govcd.AdminOrg{}) {
		return fmt.Errorf("Error fetching org: %s", d.Get("name").(string))
	}

	log.Printf("org with id %s found", d.State().ID)
	//deletes organization
	log.Printf("Deleting Org with id %s", d.State().ID)

	err = org.Delete(force, recursive)
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

	log.Printf("Reading org with id %s", d.State().ID)

	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))

	if err != nil || org == (govcd.AdminOrg{}) {
		return fmt.Errorf("Error fetching org: %s", d.Get("name").(string))
	}

	settings := getSettings(d)
	org.AdminOrg.Name = orgName
	org.AdminOrg.OrgSettings.OrgGeneralSettings = settings.OrgGeneralSettings

	log.Printf("org with id %s found", d.State().ID)
	_, err = org.Update()

	if err != nil {
		log.Printf("Error updating org with id %s : %#v", d.State().ID, err)
		return fmt.Errorf("Error updating org %#v", err)
	}

	log.Printf("Org with id %s updated", d.State().ID)
	return nil
}

func resourceOrgRead(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	log.Printf("Reading org with id %s", d.State().ID)
	org, err := govcd.GetAdminOrgByName(vcdClient.VCDClient, d.Get("name").(string))

	if err != nil || org == (govcd.AdminOrg{}) {
		log.Printf("Org with id %s not found. Setting ID to nothing", d.State().ID)
		d.SetId("")
		return nil
	}
	log.Printf("Org with id %s found", d.State().ID)
	return nil

}
