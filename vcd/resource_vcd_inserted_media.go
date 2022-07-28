package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdInsertedMedia() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdMediaInsert,
		DeleteContext: resourceVcdMediaEject,
		ReadContext:   resourceVcdVmInsertedMediaRead,
		UpdateContext: resourceVcdMediaEjectUpdate,

		Schema: map[string]*schema.Schema{
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "catalog name where to find media file",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "media name to use",
			},
			"vapp_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp to use",
			},
			"vm_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "VM in vApp in which media will be inserted or ejected",
			},
			"eject_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "When ejecting answers automatically to question yes",
			},
		},
	}
}

func resourceVcdMediaInsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] VM media insert initiated")

	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentVm(d)
	defer vcdClient.unLockParentVm(d)

	vm, org, err := getVM(d, meta)
	if err != nil || org == nil {
		return diag.Errorf("error: %#v", err)
	}

	task, err := vm.HandleInsertMedia(org, d.Get("catalog").(string), d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error: %#v", err)
	}

	d.SetId(d.Get("vapp_name").(string) + "_" + d.Get("vm_name").(string) + "_" + d.Get("name").(string))
	return resourceVcdVmInsertedMediaRead(ctx, d, meta)
}

func resourceVcdVmInsertedMediaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] VM insert media read initiated")

	vm, _, err := getVM(d, meta)
	if err != nil {
		// error logged and d.SetId("") is done in getVM function
		return nil
	}

	isIsoMounted := false
	for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
		if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
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

func resourceVcdMediaEject(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentVm(d)
	defer vcdClient.unLockParentVm(d)

	vm, org, err := getVM(d, meta)
	if err != nil {
		return diag.Errorf("error: %#v", err)
	}

	task, err := vm.HandleEjectMedia(org, d.Get("catalog").(string), d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error: %#v", err)
	}

	err = task.WaitTaskCompletion(d.Get("eject_force").(bool))
	if err != nil {
		return diag.Errorf("error: %#v", err)
	}

	return nil
}

func getVM(d *schema.ResourceData, meta interface{}) (*govcd.VM, *govcd.Org, error) {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil || org == nil || vdc == nil {
		return nil, nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vmRecord, err := vdc.QueryVM(d.Get("vapp_name").(string), d.Get("vm_name").(string))
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM. Removing from tfstate")
		d.SetId("")
		return nil, nil, fmt.Errorf("unable to find VM. Removing from tfstate. Err: #%v", err)
	}

	vm, err := vcdClient.Client.GetVMByHref(vmRecord.VM.HREF)
	if err != nil {
		log.Printf("[DEBUG] Unable to get VM data")
		return nil, nil, fmt.Errorf("error getting VM data: %s", err)
	}
	return vm, org, nil
}

//update function for "eject_force"
func resourceVcdMediaEjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	dSet(d, "eject_force", d.Get("eject_force"))
	return nil
}
