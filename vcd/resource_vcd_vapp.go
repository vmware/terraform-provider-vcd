package vcd

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppCreate,
		Update: resourceVcdVAppUpdate,
		Read:   resourceVcdVAppRead,
		Delete: resourceVcdVAppDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"catalog_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cpus": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"initscript": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
			"ovf": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"href": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"power_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"accept_all_eulas": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

	if _, ok := d.GetOk("template_name"); ok {
		if _, ok := d.GetOk("catalog_name"); ok {

			catalog, err := org.FindCatalog(d.Get("catalog_name").(string))
			if err != nil || catalog == (govcd.Catalog{}) {
				return fmt.Errorf("error finding catalog: %#v", err)
			}

			catalogitem, err := catalog.FindCatalogItem(d.Get("template_name").(string))
			if err != nil {
				return fmt.Errorf("error finding catalog item: %#v", err)
			}

			vapptemplate, err := catalogitem.GetVAppTemplate()
			if err != nil {
				return fmt.Errorf("error finding VAppTemplate: %#v", err)
			}

			log.Printf("[DEBUG] VAppTemplate: %#v", vapptemplate)
			net, err := vdc.FindVDCNetwork(d.Get("network_name").(string))
			if err != nil {
				return fmt.Errorf("error finding OrgVCD Network: %#v", err)
			}
			nets := []*types.OrgVDCNetwork{net.OrgVDCNetwork}

			storage_profile_reference := types.Reference{}

			// Override default_storage_profile if we find the given storage profile
			if d.Get("storage_profile").(string) != "" {
				storage_profile_reference, err = vdc.FindStorageProfileReference(d.Get("storage_profile").(string))
				if err != nil {
					return fmt.Errorf("error finding storage profile %s", d.Get("storage_profile").(string))
				}
			}

			log.Printf("storage_profile %s", storage_profile_reference)

			vapp, err := vdc.FindVAppByName(d.Get("name").(string))

			if err != nil {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vdc.ComposeVApp(nets, vapptemplate, storage_profile_reference, d.Get("name").(string), d.Get("description").(string), d.Get("accept_all_eulas").(bool))

					if err != nil {
						return resource.RetryableError(fmt.Errorf("error creating vapp: %#v", err))
					}

					return resource.RetryableError(task.WaitTaskCompletion())
				})

				if err != nil {
					return fmt.Errorf("error creating vapp: %#v", err)
				}
				vapp, err = vdc.FindVAppByName(d.Get("name").(string))
				if err != nil {
					return fmt.Errorf("error creating vapp: %#v", err)
				}
			}

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeVMName(d.Get("name").(string))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error with vm name change: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error changing vmname: %#v", err)
			}

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				networks := []map[string]interface{}{map[string]interface{}{
					"ip":         d.Get("ip").(string),
					"is_primary": true,
					"orgnetwork": d.Get("network_name").(string),
				}}
				task, err := vapp.ChangeNetworkConfig(networks, d.Get("ip").(string))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error with Networking change: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error changing network: %#v", err)
			}

			if ovf, ok := d.GetOk("ovf"); ok {
				err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

					if err != nil {
						return resource.RetryableError(fmt.Errorf("error set ovf: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})
				if err != nil {
					return fmt.Errorf("error completing tasks: %#v", err)
				}
			}

			initscript, ok := d.GetOk("initscript")
			if ok {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					log.Printf("running customisation script")
					task, err := vapp.RunCustomizationScript(d.Get("name").(string), initscript.(string))
					if err != nil {
						return resource.RetryableError(fmt.Errorf("error with setting init script: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})
				if err != nil {
					return fmt.Errorf(errorCompletingTask, err)
				}
			}

			if d.Get("power_on").(bool) {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vapp.PowerOn()
					if err != nil {
						return resource.RetryableError(fmt.Errorf("error powerOn machine: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})

				if err != nil {
					return fmt.Errorf("error completing powerOn tasks: %#v", err)
				}
			}

		}
	} else {
		err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			e := vdc.ComposeRawVApp(d.Get("name").(string))

			if e != nil {
				return resource.RetryableError(fmt.Errorf("Error: %#v", e))
			}

			e = vdc.Refresh()
			if e != nil {
				return resource.RetryableError(fmt.Errorf("Error: %#v", e))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	d.SetId(d.Get("name").(string))

	return resourceVcdVAppUpdate(d, meta)
}

func resourceVcdVAppUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("error getting VApp status: %#v", err)
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
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.ChangeStorageProfile(d.Get("storage_profile").(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error changing storage_profile: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
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
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeMemorySize(d.Get("memory").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing memory size: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return err
			}
		}

		if d.HasChange("cpus") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeCPUCount(d.Get("cpus").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing cpu count: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
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
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error set ovf: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	_, err = vdc.FindVAppByName(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find vapp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	if _, ok := d.GetOk("ip"); ok {
		ip := "allocated"

		oldIp, newIp := d.GetChange("ip")

		log.Printf("[DEBUG] IP has changes, old: %s - new: %s", oldIp, newIp)

		if newIp != "allocated" {
			log.Printf("[DEBUG] IP is assigned. Lets get it (%s)", d.Get("ip"))
			ip, err = getVAppIPAddress(d, meta, vdc, org)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] IP is 'allocated'")
		}

		d.Set("ip", ip)
	} else {
		d.Set("ip", "allocated")
	}

	return nil
}

func getVAppIPAddress(d *schema.ResourceData, meta interface{}, vdc govcd.Vdc, org govcd.Org) (string, error) {
	vcdClient := meta.(*VCDClient)
	var ip string
	err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		vapp, err := vdc.FindVAppByName(d.Id())
		if err != nil {
			return resource.RetryableError(fmt.Errorf("unable to find vapp"))
		}

		// getting the IP of the specific Vm, rather than index zero.
		// Required as once we add more VM's, index zero doesn't guarantee the
		// 'first' one, and tests will fail sometimes (annoying huh?)
		vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))
		if err != nil {
			return resource.RetryableError(fmt.Errorf("unable to find vm: %s", err))
		}

		ip = vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
		if ip == "" {
			return resource.RetryableError(fmt.Errorf("timeout: VM did not acquire IP address"))
		}
		return nil
	})

	return ip, err
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Id())
	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	_ = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Undeploy()
		// Very often the vApp is powered off at this point and Undeploy() waits as long as
		// vcdClient.MaxRetryTimeout before just going to Delete phase. vCD documentation makes it hard to comprehend
		// when an Undeploy() is actually needed, but the below error string is a 100% match:
		// "The requested operation could not be executed since vApp __vApp_name__ is not running"
		// If the error matches we just quit retrying and fast forward to vapp.Delete()
		var reErr = regexp.MustCompile(`.*The requested operation could not be executed since vApp.*is not running.*`)
		if err != nil && reErr.MatchString(err.Error()) {
			return &resource.RetryError{Err: err, Retryable: false}
		}

		if err != nil {
			return resource.RetryableError(fmt.Errorf("error undeploying: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Delete()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error deleting: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	return err
}
