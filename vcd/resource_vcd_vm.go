package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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

	vmDescription, err := createVMDescription(d, meta)
	if err != nil {
		return fmt.Errorf("Failed to create VMDescription: %#v", err)
	}

	log.Printf("[TRACE] Updating vApp (%s) state", vapp.VApp.Name)
	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("Error refreshing vApp: %#v", err)
	}

	// rand.Seed(time.Now().UnixNano())
	// waitTime := int64(schema.HashString(d.Get("name").(string)))
	// log.Printf("[TRACE] Backoff time: %d", waitTime)
	// duration := time.Duration(waitTime) * time.Nanosecond
	// log.Printf("[TRACE] Sleeping for %s", duration)
	// time.Sleep(duration)

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Submitting new vm request")
		task, err := vapp.AddVMs([]*types.NewVMDescription{vmDescription})
		log.Printf("[TRACE] Submitted new vm request")
		if err != nil {
			switch err.(type) {
			default:
				log.Printf("[TRACE] ERROR: %#v", err)
				return resource.NonRetryableError(fmt.Errorf("Error adding VMs: %#v", err))
			case *types.Error:
				vmError := err.(*types.Error)
				if vmError.MajorErrorCode == 500 &&
					vmError.MinorErrorCode == "INTERNAL_SERVER_ERROR" {
					return resource.RetryableError(err)
				}
				if vmError.MajorErrorCode == 400 &&
					vmError.MinorErrorCode == "BUSY_ENTITY" {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(fmt.Errorf("Error adding VMs: %#v", err))
			}
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	if err != nil {
		return fmt.Errorf("Error completing task: %#v", err)
	}

	vm, err := vapp.GetVmByName(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.Set("href", vm.HREF)
	d.SetId(vm.HREF)

	log.Printf("[DEBUG] (%s) Starting to configure", d.Get("name").(string))
	err = configureVM(d, meta)

	if err != nil {
		return err
	}

	err = readVM(d, meta)

	if err != nil {
		return err
	}

	return nil
}

func resourceVcdVMUpdate(d *schema.ResourceData, meta interface{}) error {
	err := configureVM(d, meta)

	if err != nil {
		return err
	}

	err = readVM(d, meta)

	if err != nil {
		return err
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

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := vapp.RemoveVMs([]*types.VM{vm})
		if err != nil {
			switch err.(type) {
			default:
				log.Printf("[TRACE] ERROR: %#v", err)
				return resource.NonRetryableError(fmt.Errorf("Error adding VMs: %#v", err))
			case *types.Error:
				vmError := err.(*types.Error)
				if vmError.MajorErrorCode == 500 &&
					vmError.MinorErrorCode == "INTERNAL_SERVER_ERROR" {
					return resource.RetryableError(err)
				}
				if vmError.MajorErrorCode == 400 &&
					vmError.MinorErrorCode == "BUSY_ENTITY" {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(fmt.Errorf("Error adding VMs: %#v", err))
			}
		}

		return resource.RetryableError(task.WaitTaskCompletion())
	})

	return nil
}
