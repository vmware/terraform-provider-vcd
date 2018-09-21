/*****************************************************************
* terraform-provider-vcloud-director
* Copyright (c) 2017 VMware, Inc. All Rights Reserved.
* SPDX-License-Identifier: BSD-2-Clause
******************************************************************/

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
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
			"network_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vm_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"can_publish_catalogs": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
	_, id, err := vcdClient.CreateOrg(orgName, fullName, *settings, isEnabled)

	if err != nil {
		log.Printf("Error creating organization: %#v", err)
		return fmt.Errorf("Error creating organization: %#v", err)
	}

	log.Printf("Org %s Created with id: %s", orgName, id[15:])
	d.SetId(id[15:])
	return nil
}

//Delete org //only works if the org is empty //TODO: gonna add stuff to delete everything in the org
func resourceOrgDelete(d *schema.ResourceData, m interface{}) error {

	//DELETING
	vcdClient := m.(*VCDClient)
	log.Printf("Deleting Org with id %s", d.State().ID)
	force := d.Get("force").(bool)
	recursive := d.Get("recursive").(bool)

	if force && recursive {
		err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vcdClient.RemoveAllVDCs(d.State().ID)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error changing memory size: %#v", err))
			}
			if *task.Task == (types.Task{}) {
				return nil
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return err
		}
	}
	_, err := vcdClient.DeleteOrg(d.State().ID)
	log.Printf("Org with id %s deleted", d.State().ID)
	return err
}

func getSettings(d *schema.ResourceData) *types.OrgSettings {
	var settings *types.OrgSettings
	if d.Get("vm_quota").(int) != -1 {
		settings = &types.OrgSettings{
			General: &types.OrgGeneralSettings{
				CanPublishCatalogs: d.Get("can_publish_catalogs").(bool),
				DeployedVMQuota:    d.Get("vm_quota").(int),
				StoredVMQuota:      d.Get("vm_quota").(int),
			},
		}
	} else {
		settings = &types.OrgSettings{
			General: &types.OrgGeneralSettings{
				CanPublishCatalogs: d.Get("can_publish_catalogs").(bool),
			},
		}
	}
	return settings
}

//updated the resource
func resourceOrgUpdate(d *schema.ResourceData, m interface{}) error {

	vcdClient := m.(*VCDClient)

	orgName := d.Get("name").(string)
	oldOrgFullNameRaw, newOrgFullNameRaw := d.GetChange("full_name")
	oldOrgFullName := oldOrgFullNameRaw.(string)
	newOrgFullName := newOrgFullNameRaw.(string)
	isEnabled := d.Get("is_enabled").(bool)

	settings := getSettings(d)

	vcomp := &types.OrgParams{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        orgName,
		IsEnabled:   isEnabled,
		FullName:    oldOrgFullName,
		OrgSettings: settings,
	}

	if !strings.EqualFold(oldOrgFullName, newOrgFullName) {
		return fmt.Errorf("__ERROR__ Not Updating org_full_name , API NOT IMPLEMENTED !!!!")
	}

	_, err := vcdClient.UpdateOrg(vcomp, d.State().ID)

	if err != nil {
		fmt.Errorf("Error updating org %#v", err)
	}
	return nil
}

func resourceOrgRead(d *schema.ResourceData, m interface{}) error {
	vcdClient := m.(*VCDClient)

	log.Printf("Reading org with id %s", d.State().ID)
	_, _, err := vcdClient.GetOrg(d.State().ID)

	log.Printf("Org with id %s found", d.State().ID)
	if err != nil {
		log.Printf("Org with id %s not found. Setting ID to nothing", d.State().ID)
		d.SetId("")
		return nil
	}
	return nil

}
