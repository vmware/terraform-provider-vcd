package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
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
			"template_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"catalog_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"networks": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"orgnetwork": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"is_primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
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
		},
	}
}

func resourceVcdVAppCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	if _, ok := d.GetOk("template_name"); ok {
		if _, ok := d.GetOk("catalog_name"); ok {

			catalog, err := vcdClient.Org.FindCatalog(d.Get("catalog_name").(string))
			if err != nil {
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

			networks := []*types.OrgVDCNetwork{}

			if nets := d.Get("networks").(*schema.Set).List(); nets != nil {
				for _, network := range nets {
					n := network.(map[string]interface{})
					net, err := vcdClient.OrgVdc.FindVDCNetwork(n["orgnetwork"].(string))
					networks = append(networks, net.OrgVDCNetwork)
					if err != nil {
						return fmt.Errorf("error finding OrgVCD Network: %#v", err)
					}
				}
			}

			storage_profile_reference := types.Reference{}

			// Override default_storage_profile if we find the given storage profile
			if d.Get("storage_profile").(string) != "" {
				storage_profile_reference, err = vcdClient.OrgVdc.FindStorageProfileReference(d.Get("storage_profile").(string))
				if err != nil {
					return fmt.Errorf("error finding storage profile %s", d.Get("storage_profile").(string))
				}
			}

			log.Printf("storage_profile %s", storage_profile_reference)

			vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("name").(string))

			if err != nil {
				vapp = vcdClient.NewVApp(&vcdClient.Client)

				err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
					task, err := vapp.ComposeVApp(networks, vapptemplate, storage_profile_reference, d.Get("name").(string), d.Get("description").(string))
					if err != nil {
						return resource.RetryableError(fmt.Errorf("error creating vapp: %#v", err))
					}

					return resource.RetryableError(task.WaitTaskCompletion())
				})

				if err != nil {
					return fmt.Errorf("error creating vapp: %#v", err)
				}
			}

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeVMName(d.Get("name").(string))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error with vm name change: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error changing vmname: %#v", err)
			}

			n := []map[string]interface{}{}

			nets := d.Get("networks").(*schema.Set).List()
			for _, network := range nets {
				n = append(n, network.(map[string]interface{}))
			}
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.ChangeNetworkConfig(n, d.Get("ip").(string))
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

			if d.Get("power_on").(bool) == true {
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

			initscript := d.Get("initscript").(string)

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				log.Printf("running customisation script")
				task, err := vapp.RunCustomizationScript(d.Get("name").(string), initscript)
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error with setting init script: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error completing tasks: %#v", err)
			}

		}
	} else {
		err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			e := vcdClient.OrgVdc.ComposeRawVApp(d.Get("name").(string))

			if e != nil {
				return resource.RetryableError(fmt.Errorf("error: %#v", e))
			}

			e = vcdClient.OrgVdc.Refresh()
			if e != nil {
				return resource.RetryableError(fmt.Errorf("error: %#v", e))
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
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("error getting VApp status: %#v", err)
	}

	if d.HasChange("metadata") {
		oraw, nraw := d.GetChange("metadata")
		metadata := oraw.(map[string]interface{})
		for k := range metadata {
			task, err := vapp.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("error completing tasks: %#v", err)
			}
		}
		metadata = nraw.(map[string]interface{})
		for k, v := range metadata {
			task, err := vapp.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("error completing tasks: %#v", err)
			}
		}

	}

	if d.HasChange("networks") {
		n := []map[string]interface{}{}

		nets := d.Get("networks").(*schema.Set).List()
		for _, network := range nets {
			n = append(n, network.(map[string]interface{}))
		}
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.ChangeNetworkConfig(n, d.Get("ip").(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error with Networking change: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("error changing network: %#v", err)
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
					return fmt.Errorf("error completing tasks: %#v", err)
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
				task, err := vapp.ChangeCPUcount(d.Get("cpus").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing cpu count: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error completing task: %#v", err)
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
					return resource.RetryableError(fmt.Errorf("error set ovf: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf("error completing tasks: %#v", err)
			}
		}

	}

	return resourceVcdVAppRead(d, meta)
}

func resourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vdc: %#v", err)
	}

	_, err = vcdClient.OrgVdc.FindVAppByName(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find vapp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	if _, ok := d.GetOk("ip"); ok {
		var ip string

		oldIp, newIp := d.GetChange("ip")

		log.Printf("[DEBUG] IP has changes, old: %s - new: %s", oldIp, newIp)

		if newIp != "allocated" {
			log.Printf("[DEBUG] IP is assigned. Lets get it (%s)", d.Get("ip"))
			ip, err = getVAppIPAddress(d, meta)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] IP is 'allocated'")
		}

		d.Set("ip", ip)
	}

	if _, ok := d.GetOk("networks"); ok {
		networks := []map[string]interface{}{}

		oldNetworks, newNetworks := d.GetChange("networks")

		log.Printf("[DEBUG] Networks has changes, old: %s - new: %s", oldNetworks, newNetworks)

		if newNetworks != "allocated" {
			log.Printf("[DEBUG] IP is assigned. Lets get it (%s)", d.Get("ip"))
			networks, err = getVAppNetwork(d, meta)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] IP is 'allocated'")
		}

		d.Set("networks", networks)
	}

	return nil
}

func getVAppIPAddress(d *schema.ResourceData, meta interface{}) (string, error) {
	vcdClient := meta.(*VCDClient)
	var ip string

	err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		err := vcdClient.OrgVdc.Refresh()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error refreshing vdc: %#v", err))
		}
		vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())
		if err != nil {
			return resource.RetryableError(fmt.Errorf("unable to find vapp"))
		}

		// getting the IP of the specific Vm, rather than index zero.
		// Required as once we add more VM's, index zero doesn't guarantee the
		// 'first' one, and tests will fail sometimes (annoying huh?)
		vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

		ip = vm.VM.NetworkConnectionSection.NetworkConnection[vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex].IPAddress
		if ip == "" {
			return resource.RetryableError(fmt.Errorf("timeout: VM did not acquire IP address"))
		}
		return nil
	})

	return ip, err
}

func getVAppNetwork(d *schema.ResourceData, meta interface{}) ([]map[string]interface{}, error) {
	vcdClient := meta.(*VCDClient)
	networks := []map[string]interface{}{}

	err := retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		err := vcdClient.OrgVdc.Refresh()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error refreshing vdc: %#v", err))
		}
		vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())
		if err != nil {
			return resource.RetryableError(fmt.Errorf("unable to find vapp"))
		}

		// getting the IP of the specific Vm, rather than index zero.
		// Required as once we add more VM's, index zero doesn't guarantee the
		// 'first' one, and tests will fail sometimes (annoying huh?)
		vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

		for _, v := range vm.VM.NetworkConnectionSection.NetworkConnection {
			if v.IPAddress != "" {
				n := make(map[string]interface{})
				n["orgnetwork"] = v.Network
				n["ip"] = v.IPAddress
				networks = append(networks, n)
			}
		}
		if networks == nil {
			return resource.RetryableError(fmt.Errorf("timeout: VM did not acquire IP address"))
		}
		return nil
	})

	return networks, err
}

func resourceVcdVAppDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	if err != nil {
		return fmt.Errorf("error getting VApp status: %#v", err)
	}

	_ = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.Undeploy()
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
