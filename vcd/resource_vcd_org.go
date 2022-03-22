// /*****************************************************************
// * terraform-provider-vcloud-director
// * Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// * SPDX-License-Identifier: BSD-2-Clause
// ******************************************************************/

package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		CreateContext: resourceOrgCreate,
		ReadContext:   resourceOrgRead,
		UpdateContext: resourceOrgUpdate,
		DeleteContext: resourceOrgDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOrgImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"full_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "True if this organization is enabled (allows login and all other operations).",
			},
			"deployed_vm_quota": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. (0 = unlimited)",
			},
			"stored_vm_quota": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. (0 = unlimited)",
			},
			"can_publish_catalogs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this organization is allowed to share catalogs.",
			},
			"can_publish_external_catalogs":{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if this organization is allowed to publish external catalogs.",
			},
			"can_subscribe_external_catalogs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if this organization is allowed to subscribe to external catalogs.",
			},
			"vapp_lease":{
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Defines lease parameters for vApps created in this organization",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_runtime_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long vApps can run before they are automatically stopped (in seconds). 0 means never expires",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
						"power_off_on_runtime_lease_expiration": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "When true, vApps are powered off when the runtime lease expires. " +
								"When false, vApps are suspended when the runtime lease expires",
						},
						"maximum_storage_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long stopped vApps are available before being automatically cleaned up (in seconds). 0 means never expires",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
						"delete_on_storage_lease_expiration": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "If true, storage for a vApp is deleted when the vApp's lease expires. " +
								"If false, the storage is flagged for deletion, but not deleted.",
						},
					},
				},
			},
			"vapp_template_lease": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Defines lease parameters for vApp templates created in this organization",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_storage_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long vApp templates are available before being automatically cleaned up (in seconds). 0 means never expires",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
						"delete_on_storage_lease_expiration": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "If true, storage for a vAppTemplate is deleted when the vAppTemplate lease expires. " +
								"If false, the storage is flagged for deletion, but not deleted",
						},
					},
				},
			},
			"delay_after_power_on_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Specifies this organization's default for virtual machine boot delay after power on.",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_force=True with delete_recursive=True to remove an org and any objects it contains, regardless of their state.",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_recursive=True to remove the org and any objects it contains that are in a state that normally allows removal.",
			},
		},
	}
}

// creates an organization based on defined resource
func resourceOrgCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vcdClient := m.(*VCDClient)

	orgName, fullName, err := getOrgNames(d)
	if err != nil {
		return diag.FromErr(err)
	}
	isEnabled := d.Get("is_enabled").(bool)
	description := d.Get("description").(string)

	settings := getSettings(d)

	log.Printf("[TRACE] Creating Org: %s", orgName)
	task, err := govcd.CreateOrg(vcdClient.VCDClient, orgName, fullName, description, settings, isEnabled)

	if err != nil {
		log.Printf("[DEBUG] Error creating Org: %s", err)
		return diag.Errorf("[org creation] error creating Org %s: %s", orgName, err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		log.Printf("[DEBUG] Error running Org creation task: %s", err)
		return diag.Errorf("[org creation] error running Org (%s) creation task: %s", orgName, err)
	}

	org, err := vcdClient.GetAdminOrgByName(orgName)
	if err != nil {
		return diag.Errorf("[org creation] error retrieving Org %s after creation: %s", orgName, err)
	}
	log.Printf("[TRACE] Org %s created with id: %s", orgName, org.AdminOrg.ID)

	d.SetId(org.AdminOrg.ID)

	err = createOrUpdateAdminOrgMetadata(d, org)
	if err != nil {
		return diag.Errorf("error adding metadata to Org: %s", err)
	}

	return resourceOrgRead(ctx, d, m)
}

func getSettings(d *schema.ResourceData) *types.OrgSettings {
	settings := &types.OrgSettings{}
	vappLeaseSettings := &types.VAppLeaseSettings{}
	vappTemplateLeaseSettings := &types.VAppTemplateLeaseSettings{}

	vappLeaseInputProvided := false
	vappTemplateLeaseInputProvided := false
	item, ok := d.GetOk("vapp_lease")
	if ok {
		vappLeaseInputProvided = true
		itemSlice := item.([]interface{})
		itemMap := itemSlice[0].(map[string]interface{})
		maxRuntimeLease, isSet := itemMap["maximum_runtime_lease_in_sec"]
		if isSet {
			tmpInt := maxRuntimeLease.(int)
			vappLeaseSettings.DeploymentLeaseSeconds = &tmpInt
		}
		powerOffOnLeaseExpiration, isSet := itemMap["power_off_on_runtime_lease_expiration"]
		if isSet {
			tmpBool := powerOffOnLeaseExpiration.(bool)
			vappLeaseSettings.PowerOffOnRuntimeLeaseExpiration = &tmpBool
		}
		maxStorageLease, isSet := itemMap["maximum_storage_lease_in_sec"]
		if isSet {
			tmpInt := maxStorageLease.(int)
			vappLeaseSettings.StorageLeaseSeconds = &tmpInt
		}
		deleteOnLeaseExpiration, isSet := itemMap["delete_on_storage_lease_expiration"]
		if isSet {
			tmpBool := deleteOnLeaseExpiration.(bool)
			vappLeaseSettings.DeleteOnStorageLeaseExpiration = &tmpBool
		}
	}
	item, ok = d.GetOk("vapp_template_lease")
	if ok {
		vappTemplateLeaseInputProvided = true
		itemSlice := item.([]interface{})
		itemMap := itemSlice[0].(map[string]interface{})
		maxStorageLease, isSet := itemMap["maximum_storage_lease_in_sec"]
		if isSet {
			tmpInt := maxStorageLease.(int)
			vappTemplateLeaseSettings.StorageLeaseSeconds = &tmpInt
		}
		deleteOnLeaseExpiration, isSet := itemMap["delete_on_storage_lease_expiration"]
		if isSet {
			tmpBool := deleteOnLeaseExpiration.(bool)
			vappTemplateLeaseSettings.DeleteOnStorageLeaseExpiration = &tmpBool
		}
	}

	deployedVmQuota := d.Get("deployed_vm_quota").(int)
	storedVmQuota := d.Get("stored_vm_quota").(int)
	delay := d.Get("delay_after_power_on_seconds").(int)
	canPublishCatalogs := d.Get("can_publish_catalogs").(bool)
	canPublishExternalCatalogs := d.Get("can_publish_external_catalogs").(bool)
	canSubscribeExternalCatalogs := d.Get("can_subscribe_external_catalogs").(bool)

	generalSettings := &types.OrgGeneralSettings{
		DeployedVMQuota:          deployedVmQuota,
		StoredVMQuota:            storedVmQuota,
		DelayAfterPowerOnSeconds: delay,
		CanPublishCatalogs:       canPublishCatalogs,
		CanPublishExternally:     canPublishExternalCatalogs,
		CanSubscribe:             canSubscribeExternalCatalogs,
	}

	settings.OrgGeneralSettings = generalSettings
	if vappLeaseInputProvided {
		settings.OrgVAppLeaseSettings = vappLeaseSettings
	}
	if vappTemplateLeaseInputProvided {
		settings.OrgVAppTemplateSettings = vappTemplateLeaseSettings
	}

	return settings
}

// Deletes org
func resourceOrgDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//DELETING
	vcdClient := m.(*VCDClient)
	deleteForce := d.Get("delete_force").(bool)
	deleteRecursive := d.Get("delete_recursive").(bool)

	orgName, _, err := getOrgNames(d)
	if err != nil {
		return diag.FromErr(err)
	}

	identifier := d.Id()
	log.Printf("[TRACE] Reading Org %s", identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
	if govcd.ContainsNotFound(err) && govcd.IsUuid(identifier) {
		adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(orgName)
	}

	if err != nil {
		return diag.Errorf("error fetching Org %s: %s", orgName, err)
	}

	log.Printf("[TRACE] Org %s found", orgName)
	//deletes organization
	log.Printf("[TRACE] Deleting Org %s", orgName)

	err = adminOrg.Delete(deleteForce, deleteRecursive)
	if err != nil {
		log.Printf("[DEBUG] Error deleting org %s: %s", orgName, err)
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] Org %s deleted", orgName)
	return nil
}

// Update the resource
func resourceOrgUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	vcdClient := m.(*VCDClient)

	orgName, fullName, err := getOrgNames(d)
	if err != nil {
		return diag.FromErr(err)
	}

	identifier := d.Id()
	log.Printf("[TRACE] Reading Org %s", identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
	if govcd.ContainsNotFound(err) && govcd.IsUuid(identifier) {
		adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(orgName)
	}

	if err != nil {
		return diag.Errorf("error fetching Org %s: %s", orgName, err)
	}

	settings := getSettings(d)
	adminOrg.AdminOrg.Name = orgName
	adminOrg.AdminOrg.FullName = fullName
	adminOrg.AdminOrg.Description = d.Get("description").(string)
	adminOrg.AdminOrg.IsEnabled = d.Get("is_enabled").(bool)
	adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings = settings.OrgGeneralSettings
	adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings = settings.OrgVAppTemplateSettings
	adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings = settings.OrgVAppLeaseSettings

	log.Printf("[TRACE] Org with id %s found", orgName)
	task, err := adminOrg.Update()

	if err != nil {
		log.Printf("[DEBUG] Error updating Org %s : %s", orgName, err)
		return diag.Errorf("error updating Org %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		log.Printf("[DEBUG] Error completing update of Org %s : %s", orgName, err)
		return diag.Errorf("error completing update of Org %s", err)
	}

	err = createOrUpdateAdminOrgMetadata(d, adminOrg)
	if err != nil {
		return diag.Errorf("error updating metadata from Org: %s", err)
	}

	log.Printf("[TRACE] Org %s updated", orgName)
	return nil
}

// setOrgData sets the data into the resource, taking it from the provided adminOrg
func setOrgData(d *schema.ResourceData, adminOrg *govcd.AdminOrg) error {
	dSet(d, "name", adminOrg.AdminOrg.Name)
	dSet(d, "full_name", adminOrg.AdminOrg.FullName)
	dSet(d, "description", adminOrg.AdminOrg.Description)
	dSet(d, "is_enabled", adminOrg.AdminOrg.IsEnabled)
	dSet(d, "deployed_vm_quota", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota)
	dSet(d, "stored_vm_quota", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.StoredVMQuota)
	dSet(d, "can_publish_catalogs", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs)
	dSet(d, "can_publish_external_catalogs", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally)
	dSet(d, "can_subscribe_external_catalogs", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe)
	dSet(d, "delay_after_power_on_seconds", adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.DelayAfterPowerOnSeconds)
	var err error

	vappLeaseSettings := adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings
	// OrgVAppLeaseSettings should always be filled, as the API silently uses defaults when we don't provide lease values,
	// but let's try to make it future proof and check for initialization
	if vappLeaseSettings != nil {
		var vappLease = make(map[string]interface{})

		if vappLeaseSettings.DeploymentLeaseSeconds != nil {
			vappLease["maximum_runtime_lease_in_sec"] = *vappLeaseSettings.DeploymentLeaseSeconds
		}
		if vappLeaseSettings.StorageLeaseSeconds != nil {
			vappLease["maximum_storage_lease_in_sec"] = *vappLeaseSettings.StorageLeaseSeconds
		}
		if vappLeaseSettings.PowerOffOnRuntimeLeaseExpiration != nil {
			vappLease["power_off_on_runtime_lease_expiration"] = *vappLeaseSettings.PowerOffOnRuntimeLeaseExpiration
		}
		if vappLeaseSettings.DeleteOnStorageLeaseExpiration != nil {
			vappLease["delete_on_storage_lease_expiration"] = *vappLeaseSettings.DeleteOnStorageLeaseExpiration
		}

		vappLeaseSlice := []map[string]interface{}{vappLease}
		err = d.Set("vapp_lease", vappLeaseSlice)
		if err != nil {
			return err
		}
	}

	vappTemplateSettings := adminOrg.AdminOrg.OrgSettings.OrgVAppTemplateSettings
	// OrgVAppTemplateSettings should always be filled, as the API silently uses defaults when we don't provide lease values,
	// but let's try to make it future proof and check for initialization
	if vappTemplateSettings != nil {

		var vappTemplateLease = make(map[string]interface{})

		if vappTemplateSettings.StorageLeaseSeconds != nil {
			vappTemplateLease["maximum_storage_lease_in_sec"] = vappTemplateSettings.StorageLeaseSeconds
		}
		if vappTemplateSettings.DeleteOnStorageLeaseExpiration != nil {
			vappTemplateLease["delete_on_storage_lease_expiration"] = vappTemplateSettings.DeleteOnStorageLeaseExpiration
		}

		vappTemplateLeaseSlice := []map[string]interface{}{vappTemplateLease}
		err = d.Set("vapp_template_lease", vappTemplateLeaseSlice)
		if err != nil {
			return err
		}
	}

	metadata, err := adminOrg.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to get Org metadata")
		return fmt.Errorf("unable to get Org metadata %s", err)
	}
	if err := d.Set("metadata", getMetadataStruct(metadata.MetadataEntry)); err != nil {
		return fmt.Errorf("error setting metadata: %s", err)
	}

	return nil
}

// Retrieves an Org resource from vCD
func resourceOrgRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	vcdClient := m.(*VCDClient)

	orgName, _, err := getOrgNames(d)
	if err != nil {
		return diag.FromErr(err)
	}

	identifier := d.Id()
	log.Printf("[TRACE] Reading Org %s", identifier)
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)

	// The double attempt is a workaround when dealing with
	// organizations created by previous versions, where the ID
	// was not reliable
	if govcd.ContainsNotFound(err) && govcd.IsUuid(identifier) {
		// Identifier was created by previous version and it is not a valid ID
		// If the Org is not found by ID, , the ID is invalid, and we have the name in the resource data,
		// we try to access it using the name.
		identifier = orgName
		if identifier != "" {
			log.Printf("[TRACE] Reading Org %s", identifier)
			adminOrg, err = vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)
		}
	}

	if err != nil {
		log.Printf("[DEBUG] Org %s not found. Setting ID to nothing", identifier)
		d.SetId("")
		return nil
	}
	log.Printf("[TRACE] Org with id %s found", identifier)
	d.SetId(adminOrg.AdminOrg.ID)

	err = setOrgData(d, adminOrg)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// resourceVcdOrgImport is responsible for importing the resource.
// The d.ID() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
// For this resource, the import path is just the org name.
//
// Example import path (id): orgName
func resourceVcdOrgImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	d.SetId(adminOrg.AdminOrg.ID)
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

func createOrUpdateAdminOrgMetadata(d *schema.ResourceData, adminOrg *govcd.AdminOrg) error {
	log.Printf("[TRACE] adding/updating metadata to Org")

	if d.HasChange("metadata") {
		oldRaw, newRaw := d.GetChange("metadata")
		oldMetadata := oldRaw.(map[string]interface{})
		newMetadata := newRaw.(map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		for k := range oldMetadata {
			if _, ok := newMetadata[k]; !ok {
				toBeRemovedMetadata = append(toBeRemovedMetadata, k)
			}
		}
		for _, k := range toBeRemovedMetadata {
			err := adminOrg.DeleteMetadataEntry(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			_, err := adminOrg.AddMetadataEntry(types.MetadataStringValue, k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}
