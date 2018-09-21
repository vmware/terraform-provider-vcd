package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	govcd "github.com/vmware/go-vcloud-director/govcd"
	types "github.com/vmware/go-vcloud-director/types/v56"
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
				Required: true,
				ForceNew: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Required: true,
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
	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}

	if _, ok := d.GetOk("template_name"); ok {
		if _, ok := d.GetOk("catalog_name"); ok {

			catalog, err := org.FindCatalog(d.Get("catalog_name").(string))
			if err != nil {
				return fmt.Errorf("Error finding catalog: %#v", err)
			}

			catalogitem, err := catalog.FindCatalogItem(d.Get("template_name").(string))
			if err != nil {
				return fmt.Errorf("Error finding catalog item: %#v", err)
			}

			vapptemplate, err := catalogitem.GetVAppTemplate()
			if err != nil {
				return fmt.Errorf("Error finding VAppTemplate: %#v", err)
			}

			log.Printf("[DEBUG] VAppTemplate: %#v", vapptemplate)
			net, err := vdc.FindVDCNetwork(d.Get("network_name").(string))
			if err != nil {
				return fmt.Errorf("Error finding OrgVCD Network: %#v", err)
			}
			nets := []*types.OrgVDCNetwork{net.OrgVDCNetwork}

			storage_profile_reference := types.Reference{}

			// Override default_storage_profile if we find the given storage profile
			if d.Get("storage_profile").(string) != "" {
				storage_profile_reference, err = vdc.FindStorageProfileReference(d.Get("storage_profile").(string))
				if err != nil {
					return fmt.Errorf("Error finding storage profile %s", d.Get("storage_profile").(string))
				}
			}

			log.Printf("storage_profile %s", storage_profile_reference)

			vapp, err := vdc.FindVAppByName(d.Get("name").(string))

			if err != nil {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vdc.ComposeVApp(nets, vapptemplate, storage_profile_reference, d.Get("name").(string), d.Get("description").(string), d.Get("accept_all_eulas").(bool))

					if err != nil {
						return resource.RetryableError(fmt.Errorf("Error creating vapp: %#v", err))
					}

					return resource.RetryableError(task.WaitTaskCompletion())
				})

				if err != nil {
					return fmt.Errorf("Error creating vapp: %#v", err)
				}
				vapp, err = vdc.FindVAppByName(d.Get("name").(string))
				if err != nil {
					return fmt.Errorf("Error creating vapp: %#v", err)
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
				return fmt.Errorf("Error changing vmname: %#v", err)
			}

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				networks := []map[string]interface{}{map[string]interface{}{
					"ip":         d.Get("ip").(string),
					"is_primary": true,
					"orgnetwork": d.Get("network_name").(string),
				}}
				task, err := vapp.ChangeNetworkConfig(networks, d.Get("ip").(string))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error with Networking change: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("Error changing network: %#v", err)
			}

			if ovf, ok := d.GetOk("ovf"); ok {
				err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vapp.SetOvf(convertToStringMap(ovf.(map[string]interface{})))

					if err != nil {
						return resource.RetryableError(fmt.Errorf("Error set ovf: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})
				if err != nil {
					return fmt.Errorf("Error completing tasks: %#v", err)
				}
			}

			if d.Get("power_on").(bool) == true {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vapp.PowerOn()
					if err != nil {
						return resource.RetryableError(fmt.Errorf("Error powerOn machine: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})

				if err != nil {
					return fmt.Errorf("Error completing powerOn tasks: %#v", err)
				}
			}

			initscript, ok := d.GetOk("initscript")
			if ok {
				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					log.Printf("running customisation script")
					task, err := vapp.RunCustomizationScript(d.Get("name").(string), initscript.(string))
					if err != nil {
						return resource.RetryableError(fmt.Errorf("Error with setting init script: %#v", err))
					}
					return resource.RetryableError(task.WaitTaskCompletion())
				})
				if err != nil {
					return fmt.Errorf("Error completing tasks: %#v", err)
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
	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}
	vapp, err := vdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v", err)
	}

	if d.HasChange("metadata") {
		oraw, nraw := d.GetChange("metadata")
		metadata := oraw.(map[string]interface{})
		for k := range metadata {
			task, err := vapp.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("Error deleting metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}
		metadata = nraw.(map[string]interface{})
		for k, v := range metadata {
			task, err := vapp.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("Error adding metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}

	}

	if d.HasChange("storage_profile") {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.ChangeStorageProfile(d.Get("storage_profile").(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error changing storage_profile: %#v", err))
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
					return fmt.Errorf("Error Powering Off: %#v", err)
				}
			}

			if task.Task != nil {
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("Error completing tasks: %#v", err)
				}
			}
		}

		if d.HasChange("memory") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeMemorySize(d.Get("memory").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error changing memory size: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return err
			}
		}

		if d.HasChange("cpus") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeCPUcount(d.Get("cpus").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error changing cpu count: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("Error completing task: %#v", err)
			}
		}

		if d.Get("power_on").(bool) {
			task, err := vapp.PowerOn()
			if err != nil {
				return fmt.Errorf("Error Powering Up: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
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
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}

	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}
	err = vdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
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
		err := vdc.Refresh()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error refreshing vdc: %#v", err))
		}
		vapp, err := vdc.FindVAppByName(d.Id())
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Unable to find vapp."))
		}

		// getting the IP of the specific Vm, rather than index zero.
		// Required as once we add more VM's, index zero doesn't guarantee the
		// 'first' one, and tests will fail sometimes (annoying huh?)
		vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

		ip = vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
		if ip == "" {
			return resource.RetryableError(fmt.Errorf("Timeout: VM did not acquire IP address"))
		}
		return nil
	})

	return ip, err
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}
	vapp, err := vdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v", err)
	}

	_ = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Undeploy()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error undeploying: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Delete()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error deleting: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	return err
}
