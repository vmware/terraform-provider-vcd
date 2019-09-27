package vcd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const implicitVmInVappDeprecation = "Creation of vApp and an implicit single VM in the same structure is deprecated"
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
				Description: "A unique name for the vApp",
			},
			"org": {
				Type:     schema.TypeString,
				Required: false,
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
			"template_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the vApp Template to use",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"catalog_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The catalog name in which to find the given vApp Template",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"network_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Name of the network this vApp should join",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The amount of RAM (in MB) to allocate to the vApp",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"cpus": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of virtual CPUs to allocate to the vApp",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The IP to assign to this vApp",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"storage_profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Storage profile to be used by the vApp",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the vApp",
			},
			"initscript": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A script to be run only on initial boot",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
				Description: "Key value map of metadata to assign to this vApp. Key and value can be any string.",
			},
			"ovf": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key value map of ovf parameters to assign to VM product section",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"href": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"power_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "A boolean value stating if this vApp should be powered on",
			},
			"accept_all_eulas": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Automatically accept EULA if OVA has it",
				Deprecated:  implicitVmInVappDeprecation,
			},
			"guest_properties": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key/value settings for guest properties",
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
		},
	}
}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf("error retrieving Org and VDC: %s", err)
	}

	vappName := d.Get("name").(string)
	vcdClient.lockVapp(d)
	defer vcdClient.unLockVapp(d)

	if _, ok := d.GetOk("template_name"); ok {
		if _, ok := d.GetOk("catalog_name"); ok {

			catalog, err := org.GetCatalogByName(d.Get("catalog_name").(string), false)
			if err != nil {
				return fmt.Errorf("error finding catalog: %#v", err)
			}

			catalogitem, err := catalog.GetCatalogItemByName(d.Get("template_name").(string), false)
			if err != nil {
				return fmt.Errorf("error finding catalog item: %#v", err)
			}

			vappTemplate, err := catalogitem.GetVAppTemplate()
			if err != nil {
				return fmt.Errorf("error finding VAppTemplate: %#v", err)
			}

			log.Printf("[DEBUG] VAppTemplate: %#v", vappTemplate)
			net, err := vdc.GetOrgVdcNetworkByName(d.Get("network_name").(string), false)
			if err != nil {
				return fmt.Errorf("error finding OrgVCD Network: %s", err)
			}
			nets := []*types.OrgVDCNetwork{net.OrgVDCNetwork}

			storageProfileReference := types.Reference{}

			// Override default_storage_profile if we find the given storage profile
			if d.Get("storage_profile").(string) != "" {
				storageProfileReference, err = vdc.FindStorageProfileReference(d.Get("storage_profile").(string))
				if err != nil {
					return fmt.Errorf("error finding storage profile %s", d.Get("storage_profile").(string))
				}
			}

			log.Printf("storage_profile %s", storageProfileReference)

			vapp, err := vdc.GetVAppByName(vappName, false)
			if err != nil {
				task, err := vdc.ComposeVApp(nets, vappTemplate, storageProfileReference, vappName, d.Get("description").(string), d.Get("accept_all_eulas").(bool))
				if err != nil {
					return fmt.Errorf("error creating vApp: %#v", err)
				}

				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("error creating vApp: %#v", err)
				}
				vapp, err = vdc.GetVAppByName(vappName, true)
				if err != nil {
					return fmt.Errorf("error creating vApp: %#v", err)
				}
			}

			err = vapp.BlockWhileStatus("UNRESOLVED", vcdClient.MaxRetryTimeout)
			if err != nil {
				return fmt.Errorf("error composing vApp: %s", err)
			}

			task, err := vapp.ChangeVMName(vappName)
			if err != nil {
				return fmt.Errorf("error with VM name change: %#v", err)
			}

			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("error changing vmname: %#v", err)
			}

			networks := []map[string]interface{}{map[string]interface{}{
				"ip":         d.Get("ip").(string),
				"is_primary": true,
				"orgnetwork": d.Get("network_name").(string),
			}}
			task, err = vapp.ChangeNetworkConfig(networks, d.Get("ip").(string))
			if err != nil {
				return fmt.Errorf("error with networking change: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("error changing network: %#v", err)
			}

			if ovf, ok := d.GetOk("ovf"); ok {
				task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

				if err != nil {
					return fmt.Errorf("error setting the ovf: %#v", err)
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("error completing tasks: %#v", err)
				}
			}

			initscript, ok := d.GetOk("initscript")
			if ok {
				log.Printf("running customisation script")
				task, err := vapp.RunCustomizationScript(d.Get("name").(string), initscript.(string))
				if err != nil {
					return fmt.Errorf("error with init script setting: %#v", err)
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf(errorCompletingTask, err)
				}
			}

			if d.Get("power_on").(bool) {

				task, err := vapp.PowerOn()
				if err != nil {
					return fmt.Errorf("error powering on the machine: %#v", err)
				}
				err = task.WaitTaskCompletion()

				if err != nil {
					return fmt.Errorf("error completing powerOn tasks: %#v", err)
				}
			}
		}
	} else {

		e := vdc.ComposeRawVApp(d.Get("name").(string))

		if e != nil {
			return fmt.Errorf("error: %#v", e)
		}

		e = vdc.Refresh()
		if e != nil {
			return fmt.Errorf("error: %#v", e)
		}
	}

	vapp, err := vdc.GetVAppByName(vappName, true)
	if err != nil {
		return fmt.Errorf("unable to find vApp by name %s: %s", vappName, err)
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

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByNameOrId(d.Id(), false)

	if err != nil {
		return fmt.Errorf("error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("error getting VApp status: %#v", err)
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
			task, err := vapp.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}
		for k, v := range newMetadata {
			task, err := vapp.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}
	}

	if d.HasChange("storage_profile") {
		task, err := vapp.ChangeStorageProfile(d.Get("storage_profile").(string))
		if err != nil {
			return fmt.Errorf("error changing storage_profile: %#v", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
	}

	if d.HasChange("memory") || d.HasChange("cpus") || d.HasChange("power_on") || d.HasChange("ovf") {

		if status != "POWERED_OFF" {

			task, err := vapp.PowerOff()
			if err != nil {
				// can't *always* power off an empty vApp so not necesarrily an error
				if _, ok := d.GetOk("template_name"); ok {
					return fmt.Errorf("error Powering Off: %#v", err)
				}
			}

			if task.Task != nil {
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf(errorCompletingTask, err)
				}
			}
		}

		if d.HasChange("memory") {

			task, err := vapp.ChangeMemorySize(d.Get("memory").(int))
			if err != nil {
				return fmt.Errorf("error changing memory size: %#v", err)
			}

			err = task.WaitTaskCompletion()
			if err != nil {
				return err
			}
		}

		if d.HasChange("cpus") {
			task, err := vapp.ChangeCPUCount(d.Get("cpus").(int))
			if err != nil {
				return fmt.Errorf("error changing cpu count: %#v", err)
			}

			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

		if d.Get("power_on").(bool) {
			task, err := vapp.PowerOn()
			if err != nil {
				return fmt.Errorf("error Powering Up: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("error completing tasks: %#v", err)
			}
		}

		if ovf, ok := d.GetOk("ovf"); ok {
			task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

			if err != nil {
				return fmt.Errorf("error setting the ovf: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVAppRead(d, meta, "resource")
}

func genericVcdVAppRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
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

	if _, ok := d.GetOk("ip"); ok {
		ip := "allocated"

		oldIp, newIp := d.GetChange("ip")

		log.Printf("[DEBUG] IP has changes, old: %s - new: %s", oldIp, newIp)

		if newIp != "allocated" {
			log.Printf("[DEBUG] IP is assigned. Lets get it (%s)", d.Get("ip"))
			ip, err = getVAppIPAddress(d, meta, *vdc, *org)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] IP is 'allocated'")
		}

		_ = d.Set("ip", ip)
	} else {
		_ = d.Set("ip", "allocated")
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
	// _ = d.Set("power_on", vapp.VApp.Status == 4)
	_ = d.Set("href", vapp.VApp.HREF)
	_ = d.Set("description", vapp.VApp.Description)
	statusText, err := vapp.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	_ = d.Set("status", vapp.VApp.Status)
	_ = d.Set("status_text", statusText)
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

func getVAppIPAddress(d *schema.ResourceData, meta interface{}, vdc govcd.Vdc, org govcd.Org) (string, error) {
	var ip string

	vapp, err := vdc.GetVAppByNameOrId(d.Id(), true)
	if err != nil {
		return "", fmt.Errorf("unable to find vApp")
	}

	// getting the IP of the specific Vm, rather than index zero.
	// Required as once we add more VM's, index zero doesn't guarantee the
	// 'first' one, and tests will fail sometimes (annoying huh?)
	vm, err := vapp.GetVMByName(d.Get("name").(string), false)
	if err != nil {
		return "", fmt.Errorf("unable to find VM: %s", err)
	}

	ip = vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
	if ip == "" {
		return "", fmt.Errorf("timeout: VM did not acquire IP address")
	}

	return ip, err
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
// Example resource name (_resource_name_): vcd_vapp.my-vapp
// Example import path (_the_id_string_): org.vdc.my-vapp
func resourceVcdVappImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[network direct import] resource name must be specified as org-name.vdc-name.vapp-name")
	}
	orgName, vdcName, vappName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[network direct import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vapp import] error retrieving vapp %s: %s", vappName, err)
	}
	_ = d.Set("name", vappName)
	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	d.SetId(vapp.VApp.ID)
	return []*schema.ResourceData{d}, nil
}
