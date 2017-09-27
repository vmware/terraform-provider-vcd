package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ukcloud/govcloudair/types/v56"
	"log"
)

func resourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppVmCreate,
		Update: resourceVcdVAppVmUpdate,
		Read:   resourceVcdVAppVmRead,
		Delete: resourceVcdVAppVmDelete,

		Schema: map[string]*schema.Schema{
			"vapp_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"catalog_name": {
				Type:     schema.TypeString,
				Required: true,
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
			"initscript": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
		},
	}
}

func resourceVcdVAppVmCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	catalog, err := vcdClient.Org.FindCatalog(d.Get("catalog_name").(string))
	if err != nil {
		return fmt.Errorf("Error finding catalog: %#v", err)
	}

	catalogitem, err := catalog.FindCatalogItem(d.Get("template_name").(string))
	if err != nil {
		return fmt.Errorf("Error finding catelog item: %#v", err)
	}

	vapptemplate, err := catalogitem.GetVAppTemplate()
	if err != nil {
		return fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}

	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("Error finding Vapp: %#v", err)
	}

	var netnames []string
	var vAppNetworkName []string

	networks := []*types.OrgVDCNetwork{}

	if nets := d.Get("networks").(*schema.Set).List(); nets != nil {
		for _, network := range nets {
			n := network.(map[string]interface{})
			net, err := vcdClient.OrgVdc.FindVDCNetwork(n["orgnetwork"].(string))
			networks = append(networks, net.OrgVDCNetwork)
			netnames = append(netnames, net.OrgVDCNetwork.Name)

			if err != nil {
				return fmt.Errorf("Error finding vApp network: %#v", err)
			}
		}

		vAppNetworkConfig, err := vapp.GetNetworkConfig()
		log.Printf("Networkconfig: %s", vAppNetworkConfig)

		if vAppNetworkConfig.NetworkConfig != nil {
			for _, v := range vAppNetworkConfig.NetworkConfig {
				if netnames == nil {
					net, err := vcdClient.OrgVdc.FindVDCNetwork(v.NetworkName)
					if err != nil {
						return fmt.Errorf("Error finding vApp network: %#v", err)
					}

					netnames = append(netnames, net.OrgVDCNetwork.Name)
				}

			}

		} else {

			if netnames == nil {
				return fmt.Errorf("'networks' must be valid when adding VM to raw vapp")
			}

			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vapp.AddRAWNetworkConfig(networks)
				if err != nil {
					return resource.RetryableError(fmt.Errorf("Error assigning network to vApp: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})

			if err != nil {
				return fmt.Errorf("Error2 assigning network to vApp:: %#v", err)
			} else {
				vAppNetworkName = netnames
			}
		}
	}

	if len(vAppNetworkName) != len(netnames) {
		return fmt.Errorf("The VDC network '%s' must be assigned to the vApp. Currently the vApp network date is %s", netnames, vAppNetworkName)
	}

	log.Printf("[TRACE] Network names found: %s", netnames)

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Creating VM: %s", d.Get("name").(string))
		task, err := vapp.AddVM(networks, vapptemplate, d.Get("name").(string))

		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error adding VM: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	if err != nil {
		return fmt.Errorf("Error completing tasks: %#v", err)
	}

	vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM1 : %#v", err)
	}

	n := []map[string]interface{}{}

	nets := d.Get("networks").(*schema.Set).List()
	for _, network := range nets {
		n = append(n, network.(map[string]interface{}))
	}
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.ChangeNetworkConfig(n, d.Get("ip").(string))
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error with Networking change: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error changing network: %#v", err)
	}

	initscript := d.Get("initscript").(string)

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vm.RunCustomizationScript(d.Get("name").(string), initscript)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error with setting init script: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error completing tasks: %#v", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceVcdVAppVmUpdate(d, meta)
}

func resourceVcdVAppVmUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM2: %#v", err)
	}

	status, err := vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VM status: %#v", err)
	}

	if d.HasChange("networks") {
		n := []map[string]interface{}{}

		nets := d.Get("networks").(*schema.Set).List()
		for _, network := range nets {
			n = append(n, network.(map[string]interface{}))
		}
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vm.ChangeNetworkConfig(n, d.Get("ip").(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error with Networking change: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("Error changing network: %#v", err)
		}
	}

	if d.HasChange("memory") || d.HasChange("cpus") || d.HasChange("power_on") {
		if status != "POWERED_OFF" {
			task, err := vm.PowerOff()
			if err != nil {
				return fmt.Errorf("Error Powering Off: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}

		if d.HasChange("memory") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vm.ChangeMemorySize(d.Get("memory").(int))
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
				task, err := vm.ChangeCPUcount(d.Get("cpus").(int))
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
			task, err := vm.PowerOn()
			if err != nil {
				return fmt.Errorf("Error Powering Up: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf("Error completing tasks: %#v", err)
			}
		}

	}

	return resourceVcdVAppVmRead(d, meta)
}

func resourceVcdVAppVmRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM3 : %#v", err)
	}

	d.Set("name", vm.VM.Name)

	networks := []map[string]interface{}{}

	for _, v := range vm.VM.NetworkConnectionSection.NetworkConnection {
		n := make(map[string]interface{})
		n["orgnetwork"] = v.Network
		if v.IPAddress != "" {
			n["ip"] = v.IPAddress
		} else {
			n["ip"] = "allocated"
		}
		if vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex == v.NetworkConnectionIndex {
			n["is_primary"] = true
		}
		networks = append(networks, n)
	}

	d.Set("networks", networks)
	d.Set("ip", vm.VM.NetworkConnectionSection.NetworkConnection[vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex].IPAddress)
	d.Set("href", vm.VM.HREF)

	return nil
}

func resourceVcdVAppVmDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vcdClient.OrgVdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		return fmt.Errorf("Error getting VM4 : %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting vApp status: %#v", err)
	}

	log.Printf("[TRACE] Vapp Status:: %s", status)
	if status != "POWERED_OFF" {
		log.Printf("[TRACE] Undeploying vApp: %s", vapp.VApp.Name)
		task, err := vapp.Undeploy()
		if err != nil {
			return fmt.Errorf("Error Undeploying vApp: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error completing tasks: %#v", err)
		}
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Removing VM: %s", vm.VM.Name)
		err := vapp.RemoveVM(vm)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error deleting: %#v", err))
		}

		return nil
	})

	if status != "POWERED_OFF" {
		log.Printf("[TRACE] Redeploying vApp: %s", vapp.VApp.Name)
		task, err := vapp.Deploy()
		if err != nil {
			return fmt.Errorf("Error Deploying vApp: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error completing tasks: %#v", err)
		}

		log.Printf("[TRACE] Powering on vApp: %s", vapp.VApp.Name)
		task, err = vapp.PowerOn()
		if err != nil {
			return fmt.Errorf("Error Powering on vApp: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error completing tasks: %#v", err)
		}
	}

	return err
}
