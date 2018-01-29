package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ukcloud/govcloudair"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func resourceVcdVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVMCreate,
		Update: resourceVcdVMUpdate,
		Read:   resourceVcdVMRead,
		Delete: resourceVcdVMDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"catalog_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cpus": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"network": {
				Type:     schema.TypeList,
				Required: true,

				Elem: &schema.Resource{
					Schema: VirtualMachineNetworkSubresourceSchema(),
				},
			},
			"initscript": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"href": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vapp_href": {
				Type:     schema.TypeString,
				Required: true,
			},
			"power_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"nested_hypervisor_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_password_auto": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"admin_password": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVcdVMCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] Updating state from VCD")
	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_href").(string))
	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting VApp status: %#v, %s", err, status)
	}

	sourceItem, err := composeSourceItem(d, meta)
	if err != nil {
		return fmt.Errorf("Failed to create VMDescription: %#v", err)
	}

	log.Printf("[TRACE] Updating vApp (%s) state", vapp.VApp.Name)
	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vApp: %#v", err)
	}

	err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
		return vapp.AddVMs([]*types.SourcedCompositionItemParam{sourceItem})
	})

	if err != nil {
		return fmt.Errorf("Error completing task: %#v", err)
	}

	vm, err := vapp.GetVmByName(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.Set("href", vm.VM.HREF)
	d.SetId(vm.VM.HREF)

	log.Printf("[DEBUG] (%s) Starting to configure", d.Get("name").(string))

	err = configureVM(d, vm)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] (%s) Sending reconfiguration event to VCD", vm.VM.Name)
	err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
		return vm.Reconfigure()
	})
	if err != nil {
		return err
	}

	err = readVM(d, meta)

	if err != nil {
		return err
	}

	status, err = vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting vm status: %#v, %s", err, status)
	}

	if d.Get("power_on").(bool) && status != types.VAppStatuses[4] {
		log.Printf("[DEBUG] (%s) Powering on VM", vm.VM.Name)
		err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
			return vm.PowerOn()
		})
		if err != nil {
			return err
		}
	} else if !d.Get("power_on").(bool) && status != types.VAppStatuses[8] {
		log.Printf("[DEBUG] (%s) Powering off VM", vm.VM.Name)
		err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
			return vm.PowerOff()
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceVcdVMUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Get VM object from VCD
	vm, err := vcdClient.FindVMByHREF(d.Get("href").(string))

	if err != nil {
		return fmt.Errorf("Could not find VM (%s)(%s) in VCD", d.Get("name").(string), d.Get("href").(string))
	}

	err = configureVM(d, &vm)

	if err != nil {
		return err
	}

	status, err := vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting vm status: %#v, %s", err, status)
	}

	if status != types.VAppStatuses[8] {
		log.Printf("[DEBUG] (%s) Powering off VM", vm.VM.Name)
		err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
			return vm.PowerOff()
		})
		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] (%s) Sending reconfiguration event to VCD", vm.VM.Name)
	err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
		return vm.Reconfigure()
	})
	if err != nil {
		return err
	}

	err = readVM(d, meta)

	if err != nil {
		return err
	}

	status, err = vm.GetStatus()
	if err != nil {
		return fmt.Errorf("Error getting vm status: %#v, %s", err, status)
	}

	if d.Get("power_on").(bool) && status != types.VAppStatuses[4] {
		log.Printf("[DEBUG] (%s) Powering on VM", vm.VM.Name)
		err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
			return vm.PowerOn()
		})
		if err != nil {
			return err
		}
	} else if !d.Get("power_on").(bool) && status != types.VAppStatuses[8] {
		log.Printf("[DEBUG] (%s) Powering off VM", vm.VM.Name)
		err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
			return vm.PowerOff()
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceVcdVMRead(d *schema.ResourceData, meta interface{}) error {
	err := readVM(d, meta)

	if err != nil {
		return err
	}

	return nil
}

func resourceVcdVMDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] Updating state from VCD")
	err := vcdClient.OrgVdc.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vdc: %#v", err)
	}

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Get("vapp_href").(string))
	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	log.Printf("[TRACE] Updating vApp (%s) state", vapp.VApp.Name)
	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vApp: %#v", err)
	}

	vm, err := vapp.GetVmByHREF(d.Id())

	if err != nil {
		return err
	}

	err = retryCallWithVcloudErrorHandling(vcdClient.MaxRetryTimeout, func() (govcloudair.Task, error) {
		return vapp.RemoveVMs([]*types.VM{vm.VM})
	})
	if err != nil {
		return err
	}

	return nil
}
