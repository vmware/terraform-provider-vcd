package vcd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

const vAppUnknownStatus = "-unknown-status-"

func resourceVcdVApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppCreate,
		Update: resourceVcdVAppUpdate,
		Read:   resourceVcdVAppRead,
		Delete: resourceVcdVAppDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVappImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A name for the vApp, unique withing the VDC",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the vApp",
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
				Description: "Key value map of metadata to assign to this vApp. Key and value can be any string.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vApp Hyper Reference",
			},
			"power_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "A boolean value stating if this vApp should be powered on",
			},
			"guest_properties": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key/value settings for guest properties. Will be picked up by new VMs when created.",
			},
			"status": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Shows the status code of the vApp",
			},
			"status_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Shows the status of the vApp",
			},
			"lease": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Defines lease parameters for this vApp",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"runtime_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long any of the VMs in the vApp can run before the vApp is automatically powered off or suspended. 0 means never expires",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
						"storage_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long the vApp is available before being automatically deleted or marked as expired. 0 means never expires",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
					},
				},
			},
		},
	}
}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf("error retrieving Org and VDC: %s", err)
	}

	vappName := d.Get("name").(string)
	vappDescription := d.Get("description").(string)
	vcdClient.lockVapp(d)
	defer vcdClient.unLockVapp(d)

	vapp, err := vdc.CreateRawVApp(vappName, vappDescription)
	if err != nil {
		return fmt.Errorf("error creating vApp %s: %s", vappName, err)
	}

	if _, ok := d.GetOk("guest_properties"); ok {

		// Even though vApp has a task and waits for its completion it happens that it is not ready
		// for operation just after provisioning therefore we wait for it to exit UNRESOLVED state
		err = vapp.BlockWhileStatus("UNRESOLVED", vcdClient.MaxRetryTimeout)
		if err != nil {
			return fmt.Errorf("timed out waiting for vApp to exit UNRESOLVED state: %s", err)
		}

		guestProperties, err := getGuestProperties(d)
		if err != nil {
			return fmt.Errorf("unable to convert guest properties to data structure")
		}

		log.Printf("[TRACE] Setting vApp guest properties")
		_, err = vapp.SetProductSectionList(guestProperties)
		if err != nil {
			return fmt.Errorf("error setting guest properties: %s", err)
		}
	}

	d.SetId(vapp.VApp.ID)

	return resourceVcdVAppUpdate(d, meta)
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByNameOrId(d.Id(), false)

	if err != nil {
		return fmt.Errorf("error finding VApp: %s", err)
	}

	var runtimeLease = vapp.VApp.LeaseSettingsSection.DeploymentLeaseInSeconds
	var storageLease = vapp.VApp.LeaseSettingsSection.StorageLeaseInSeconds
	rawLeaseSection1, ok := d.GetOk("lease")
	if ok {
		// We have a lease block
		rawLeaseSection2 := rawLeaseSection1.([]interface{})
		leaseSection := rawLeaseSection2[0].(map[string]interface{})
		runtimeLease = leaseSection["runtime_lease_in_sec"].(int)
		storageLease = leaseSection["storage_lease_in_sec"].(int)
	} else {
		// No lease block: we read the lease defaults from the Org
		adminOrg, err := vcdClient.GetAdminOrgById(org.Org.ID)
		if err != nil {
			return fmt.Errorf("error retrieving admin Org from parent Org in vApp %s: %s", vapp.VApp.Name, err)
		}
		if adminOrg.AdminOrg.OrgSettings == nil || adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings == nil {
			return fmt.Errorf("error retrieving Org lease settings")
		}
		runtimeLease = *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.DeploymentLeaseSeconds
		storageLease = *adminOrg.AdminOrg.OrgSettings.OrgVAppLeaseSettings.StorageLeaseSeconds
	}

	if runtimeLease != vapp.VApp.LeaseSettingsSection.DeploymentLeaseInSeconds ||
		storageLease != vapp.VApp.LeaseSettingsSection.StorageLeaseInSeconds {
		err = vapp.RenewLease(runtimeLease, storageLease)
		if err != nil {
			return fmt.Errorf("error updating VApp lease terms: %s", err)
		}
	}
	if d.HasChange("description") {
		err = vapp.UpdateNameDescription(d.Get("name").(string), d.Get("description").(string))
		if err != nil {
			return fmt.Errorf("error updating VApp: %s", err)
		}
	}
	if d.HasChange("guest_properties") {
		vappProperties, err := getGuestProperties(d)
		if err != nil {
			return fmt.Errorf("unable to convert guest properties to data structure")
		}

		log.Printf("[TRACE] Updating vApp guest properties")
		_, err = vapp.SetProductSectionList(vappProperties)
		if err != nil {
			return fmt.Errorf("error setting guest properties: %s", err)
		}
	}

	err = createOrUpdateMetadata(d, vapp, "metadata")
	if err != nil {
		return err
	}

	if d.HasChange("power_on") && d.Get("power_on").(bool) {
		task, err := vapp.PowerOn()
		if err != nil {
			return fmt.Errorf("error Powering Up: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error completing tasks: %#v", err)
		}
	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVAppRead(d, meta, "resource")
}

func genericVcdVAppRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}
	identifier := d.Id()

	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return fmt.Errorf("[vapp read] no identifier provided")
	}
	vapp, err := vdc.GetVAppByNameOrId(identifier, false)
	if err != nil {
		if origin == "resource" {
			log.Printf("[DEBUG] Unable to find vApp. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[vapp read] error retrieving vApp %s: %s", identifier, err)
	}

	// update guest properties
	guestProperties, err := vapp.GetProductSectionList()
	if err != nil {
		return fmt.Errorf("unable to read guest properties: %s", err)
	}

	err = setGuestProperties(d, guestProperties)
	if err != nil {
		return fmt.Errorf("unable to set guest properties in state: %s", err)
	}

	leaseInfo, err := vapp.GetLease()
	if err != nil {
		return fmt.Errorf("unable to get lease information: %s", err)
	}
	leaseData := []map[string]interface{}{
		{
			"runtime_lease_in_sec": leaseInfo.DeploymentLeaseInSeconds,
			"storage_lease_in_sec": leaseInfo.StorageLeaseInSeconds,
		},
	}
	err = d.Set("lease", leaseData)
	if err != nil {
		return fmt.Errorf("unable to set lease information in state: %s", err)
	}

	statusText, err := vapp.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	dSet(d, "status", vapp.VApp.Status)
	dSet(d, "status_text", statusText)
	dSet(d, "href", vapp.VApp.HREF)
	dSet(d, "description", vapp.VApp.Description)
	metadata, err := vapp.GetMetadata()
	if err != nil {
		return fmt.Errorf("[vapp read] error retrieving metadata: %s", err)
	}
	metadataStruct := getMetadataStruct(metadata.MetadataEntry)
	err = d.Set("metadata", metadataStruct)
	if err != nil {
		return fmt.Errorf("[vapp read] error setting metadata: %s", err)
	}

	d.SetId(vapp.VApp.ID)

	return nil
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockVapp(d)
	defer vcdClient.unLockVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByNameOrId(d.Id(), false)
	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	// to avoid network destroy issues - detach networks from vApp
	task, err := vapp.RemoveAllNetworks()
	if err != nil {
		return fmt.Errorf("error with networking change: %#v", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error changing network: %#v", err)
	}

	err = tryUndeploy(*vapp)
	if err != nil {
		return err
	}

	task, err = vapp.Delete()
	if err != nil {
		return fmt.Errorf("error deleting: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error with deleting vApp task: %#v", err)
	}

	return nil
}

// Try to undeploy a vApp, but do not throw an error if the vApp is powered off.
// Very often the vApp is powered off at this point and Undeploy() would fail with error:
// "The requested operation could not be executed since vApp vApp_name is not running"
// So, if the error matches we just ignore it and the caller may fast forward to vapp.Delete()
func tryUndeploy(vapp govcd.VApp) error {
	task, err := vapp.Undeploy()
	var reErr = regexp.MustCompile(`.*The requested operation could not be executed since vApp.*is not running.*`)
	if err != nil && reErr.MatchString(err.Error()) {
		// ignore - can't be undeployed
		return nil
	} else if err != nil {
		return fmt.Errorf("error undeploying vApp: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error undeploying vApp: %#v", err)
	}
	return nil
}

// resourceVcdVappImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp.vapp_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name
func resourceVcdVappImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[vapp import] resource name must be specified as org-name.vdc-name.vapp-name")
	}
	orgName, vdcName, vappName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vapp import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vapp import] error retrieving vapp %s: %s", vappName, err)
	}
	dSet(d, "name", vappName)
	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	d.SetId(vapp.VApp.ID)
	return []*schema.ResourceData{d}, nil
}
