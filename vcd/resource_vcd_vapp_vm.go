package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	govcd "github.com/vmware/go-vcloud-director/govcd"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"log"
)

func resourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppVmCreate,
		Update: resourceVcdVAppVmUpdate,
		Read:   resourceVcdVAppVmRead,
		Delete: resourceVcdVAppVmDelete,

		Schema: map[string]*schema.Schema{
			"vapp_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
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
			"template_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"catalog_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cpus": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"initscript": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"href": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"accept_all_eulas": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"power_on": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"network_href": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVcdVAppVmCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}
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

	accept_eulas := d.Get("accept_all_eulas").(bool)

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("Error finding Vapp: %#v", err)
	}

	netname := "blank"
	net, err := vdc.FindVDCNetwork(d.Get("network_name").(string))

	if err == nil {
		netname = net.OrgVDCNetwork.Name
	}

	nets := []*types.OrgVDCNetwork{net.OrgVDCNetwork}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()

	vAppNetworkName := "blank"
	if vAppNetworkConfig.NetworkConfig != nil {
		vAppNetworkName = vAppNetworkConfig.NetworkConfig[0].NetworkName
		if netname == "blank" {
			net, err = vdc.FindVDCNetwork(vAppNetworkName)
			if err != nil {
				return fmt.Errorf("Error finding vApp network: %#v", err)
			}

			netname = net.OrgVDCNetwork.Name
		}

	} else {

		if netname == "blank" {
			return fmt.Errorf("'network_name' must be valid when adding VM to raw vapp")
		}

		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.AddRAWNetworkConfig(nets)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error assigning network to vApp: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})

		if err != nil {
			return fmt.Errorf("Error2 assigning network to vApp:: %#v", err)
		} else {
			vAppNetworkName = netname
		}

	}

	if vAppNetworkName != netname {
		return fmt.Errorf("The VDC network '%s' must be assigned to the vApp. Currently the vApp network date is %s", netname, vAppNetworkName)
	}

	log.Printf("[TRACE] Network name found: %s", netname)

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Creating VM: %s", d.Get("name").(string))
		task, err := vapp.AddVM(nets, vapptemplate, d.Get("name").(string), accept_eulas)

		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error adding VM: %#v", err))
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	if err != nil {
		return fmt.Errorf("Error completing tasks: %#v", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM1 : %#v", err)
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		networks := []map[string]interface{}{map[string]interface{}{
			"ip":         d.Get("ip").(string),
			"is_primary": true,
			"orgnetwork": netname,
		}}
		task, err := vm.ChangeNetworkConfig(networks, d.Get("ip").(string))
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Error with Networking change: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return fmt.Errorf("Error changing network: %#v", err)
	}

	initscript, ok := d.GetOk("initscript")

	if ok {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vm.RunCustomizationScript(d.Get("name").(string), initscript.(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Error with setting init script: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("Error completing tasks: %#v", err)
		}
	}
	d.SetId(d.Get("name").(string))

	return resourceVcdVAppVmUpdate(d, meta)
}

func resourceVcdVAppVmUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM2: %#v", err)
	}

	status, err := vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VM status: %#v", err)
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

	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error getting VM3 : %#v", err)
	}

	d.Set("name", vm.VM.Name)
	d.Set("ip", vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress)
	d.Set("href", vm.VM.HREF)

	return nil
}

func resourceVcdVAppVmDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	org, err := govcd.GetOrgByName(vcdClient.VCDClient, d.Get("org").(string))
	if err != nil {
		return fmt.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVdcByName(d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf("Could not find vdc: %v", err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vapp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

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
