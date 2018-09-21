/*****************************************************************
* terraform-provider-vcloud-director
* Copyright (c) 2017 VMware, Inc. All Rights Reserved.
* SPDX-License-Identifier: BSD-2-Clause
******************************************************************/

package vcd

import (
	"fmt"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	govcd "github.com/ukcloud/govcloudair"
	"log"
	"strconv"
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
			},
			"stored_vm_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"can_publish_catalogs": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"use_server_boot_sequence": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
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

	settings := map[string]string{}

	canPublishCatalogs, ok := d.GetOk("can_publish_catalogs")
	if ok {
		settings["CanPublishCatalogs"] = strconv.FormatBool(canPublishCatalogs.(bool))
	}

	vmQuota, ok := d.GetOk("deployed_vm_quota")
	if ok {
		settings["DeployedVMQuota"] = strconv.Itoa(vmQuota.(int))
	}

	vmQuota, ok = d.GetOk("stored_vm_quota")
	if ok {
		settings["StoredVMQuota"] = strconv.Itoa(vmQuota.(int))
	}

	delay, ok := d.GetOk("delay_after_power_on_seconds")
	if ok {
		settings["DelayAfterPowerOnSeconds"] = strconv.Itoa(delay.(int))
	}

	serverboot, ok := d.GetOk("use_server_boot_sequence")
	if ok {
		settings["UseServerBootSequence"] = strconv.FormatBool(serverboot.(bool))
	}

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

//Deletes org
func resourceOrgDelete(d *schema.ResourceData, m interface{}) error {

	//DELETING
	vcdClient := m.(*VCDClient)
	force := d.Get("force").(bool)
	recursive := d.Get("recursive").(bool)

	//fetches org
	log.Printf("Reading org with id %s", d.State().ID)
	org, err := govcd.GetAdminOrgById(vcdClient.VCDClient, d.State().ID)
	if err != nil {
		return fmt.Errorf("Error fetching org: %#v", err)
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

	org, err := govcd.GetAdminOrgById(vcdClient.VCDClient, d.State().ID)

	if err != nil {
		return fmt.Errorf("Error fetching org: %#v", err)
	}

	org.AdminOrg.Name = orgName
	canPublishCatalogs, ok := d.GetOk("can_publish_catalogs")
	if ok {
		org.AdminOrg.OrgSettings.General.CanPublishCatalogs = canPublishCatalogs.(bool)
	}

	vmQuota, ok := d.GetOk("deployed_vm_quota")
	if ok {
		org.AdminOrg.OrgSettings.General.DeployedVMQuota = vmQuota.(int)
	}

	vmQuota, ok = d.GetOk("stored_vm_quota")
	if ok {
		org.AdminOrg.OrgSettings.General.StoredVMQuota = vmQuota.(int)
	}

	delay, ok := d.GetOk("delay_after_power_on_seconds")
	if ok {
		org.AdminOrg.OrgSettings.General.DelayAfterPowerOnSeconds = delay.(int)
	}

	serverboot, ok := d.GetOk("use_server_boot_sequence")
	if ok {
		org.AdminOrg.OrgSettings.General.UseServerBootSequence = serverboot.(bool)
	}

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
	_, err := govcd.GetAdminOrgById(vcdClient.VCDClient, d.State().ID)

	if err != nil {
		log.Printf("Org with id %s not found. Setting ID to nothing", d.State().ID)
		d.SetId("")
		return nil
	}
	log.Printf("Org with id %s found", d.State().ID)
	return nil

}
