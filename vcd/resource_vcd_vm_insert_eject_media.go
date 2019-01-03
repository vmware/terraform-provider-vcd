package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
	"log"
)

func resourceVcdInsertEjectMedia() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdMediaInsert,
		Delete: resourceVcdMediaEject,
		Read:   resourceVcdVmInsertedMediaRead,

		Schema: map[string]*schema.Schema{
			"vdc": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"catalog": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog name where to find media file",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "media name to use",
			},
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vapp to use",
			},
			"vm_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vm in vapp in which media will be inserted or ejected",
			},
		},
	}
}

func resourceVcdMediaInsert(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] VM media insert initiated")

	vm, org, err := getVM(d, meta)
	if err != nil || org == (govcd.Org{}) {
		return fmt.Errorf("error: %#v", err)
	}

	task, err := vm.HandleInsertMedia(&org, d.Get("catalog").(string), d.Get("name").(string))

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	d.SetId(d.Get("vapp_name").(string) + "_" + d.Get("vm_name").(string) + "_" + d.Get("name").(string))
	return resourceVcdVmInsertedMediaRead(d, meta)
}

func resourceVcdVmInsertedMediaRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] VM insert media read initiated")

	vm, _, err := getVM(d, meta)
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	isIsoMounted := false
	for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
		if hardwareItem.ResourceSubType == "vmware.cdrom.iso" {
			isIsoMounted = true
			break
		}
	}

	if !isIsoMounted {
		log.Printf("[DEBUG] Didn't find mounted iso in VM. Removing from tfstate")
		d.SetId("")
	}

	log.Printf("[TRACE] VM insert media read completed.")
	return nil
}

func resourceVcdMediaEject(d *schema.ResourceData, meta interface{}) error {
	vm, org, err := getVM(d, meta)
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	task, err := vm.HandleEjectMedia(&org, d.Get("catalog").(string), d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	return nil
}

func getVM(d *schema.ResourceData, meta interface{}) (govcd.VM, govcd.Org, error) {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil || org == (govcd.Org{}) || vdc == (govcd.Vdc{}) {
		return govcd.VM{}, govcd.Org{}, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vmRecord, err := vdc.QueryVM(d.Get("vapp_name").(string), d.Get("vm_name").(string))
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM. Removing from tfstate")
		d.SetId("")
		return govcd.VM{}, govcd.Org{}, fmt.Errorf("unableto find VM. Removing from tfstate. Err: #%v", err)
	}

	vm, err := vcdClient.Client.FindVMByHREF(vmRecord.VM.HREF)
	if err != nil || vm == (govcd.VM{}) {
		log.Printf("[DEBUG] Unable to get vm data")
		return govcd.VM{}, govcd.Org{}, fmt.Errorf("error getting VM data: %s", err)
	}
	return vm, org, nil
}
