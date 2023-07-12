package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdClonedVApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdClonedVAppCreate,
		ReadContext:   resourceVcdClonedVAppRead,
		UpdateContext: resourceVcdClonedVAppUpdate,
		DeleteContext: resourceVcdClonedVAppDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A name for the vApp, unique withing the VDC",
			},
			"org": {
				Type:     schema.TypeString,
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the vApp",
			},
			"power_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "A boolean value stating if this vApp should be powered on",
			},
			"delete_source": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If true, it will delete the source (vApp or template) after creating the new vApp",
			},
			"source_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of the source to use for the creation of this vApp (one of 'vapp' or 'template')",
				ValidateFunc: validation.StringInSlice([]string{"vapp", "template"}, true),
			},
			"source_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the source to use for the creation of this vApp",
			},
			"vm_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of VMs contained in this vApp (in alphabetic order)",
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
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vApp Hyper Reference",
			},
		},
	}
}

func resourceVcdClonedVAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving Org and VDC: %s", err)
	}

	vappName := d.Get("name").(string)
	vappDescription := d.Get("description").(string)
	powerOn := d.Get("power_on").(bool)

	sourceType := strings.ToLower(d.Get("source_type").(string))
	sourceId := d.Get("source_id").(string)
	deleteSource := d.Get("delete_source").(bool)

	var vapp *govcd.VApp
	if sourceType == "vapp" {
		sourceVapp, err := vdc.GetVAppByNameOrId(sourceId, true)
		if err != nil {
			return diag.Errorf("error retrieving source vApp with ID '%s': %s", sourceId, err)
		}
		params := &types.CloneVAppParams{
			Name:        vappName,
			Deploy:      true,
			PowerOn:     powerOn,
			Description: vappDescription,
			Source: &types.Reference{
				HREF: sourceVapp.VApp.HREF,
				ID:   sourceVapp.VApp.ID,
				Type: sourceVapp.VApp.Type,
				Name: sourceVapp.VApp.Name,
			},
			IsSourceDelete: addrOf(deleteSource),
		}
		sourceStatus, err := sourceVapp.GetStatus()
		if err != nil {
			return diag.Errorf("error getting the status of source vApp %s: %s", sourceVapp.VApp.Name, err)
		}
		if sourceStatus != "POWERED_OFF" {
			if deleteSource {
				return diag.Errorf("'delete_source' was requested, but the vApp '%s' is not powered off."+
					" Either set 'delete_source' to false or power off the vApp", sourceVapp.VApp.Name)
			}
		}
		vapp, err = vdc.CloneVapp(params)
		if err != nil {
			return diag.Errorf("error cloning vApp %s from another vApp: %s", vappName, err)
		}
	} else {
		sourceTemplate, err := vcdClient.GetVAppTemplateById(sourceId)
		if err != nil {
			return diag.Errorf("error retrieving vApp template with ID '%s': %s", sourceId, err)
		}
		params := &types.InstantiateVAppTemplateParams{
			Name:        vappName,
			Deploy:      true,
			PowerOn:     powerOn,
			Description: vappDescription,
			Source: &types.Reference{
				HREF: sourceTemplate.VAppTemplate.HREF,
				ID:   sourceTemplate.VAppTemplate.ID,
				Type: sourceTemplate.VAppTemplate.Type,
				Name: sourceTemplate.VAppTemplate.Name,
			},
			IsSourceDelete:   deleteSource,
			AllEULAsAccepted: true,
		}
		vapp, err = vdc.CreateVappFromTemplate(params)
		if err != nil {
			return diag.Errorf("error creating vApp %s from template: %s", vappName, err)
		}
	}

	d.SetId(vapp.VApp.ID)

	vappStatus, err := vapp.GetStatus()
	if err != nil {
		return diag.Errorf("error retrieving vApp '%s' status: %s", vappName, err)
	}
	// If the cloned vApp is suspended and power ON was requested, we adjust the status accordingly
	if vappStatus == "SUSPENDED" && powerOn {
		err = vapp.DiscardSuspendedState()
		if err != nil {
			return diag.Errorf("error requesting suspended state change on vApp '%s': %s", vappName, err)
		}
		task, err := vapp.PowerOn()
		if err != nil {
			return diag.Errorf("error requesting power change on vApp '%s': %s", vappName, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error while powering on vApp '%s': %s", vappName, err)
		}
	}

	return resourceVcdClonedVAppRead(ctx, d, meta)
}

func resourceVcdClonedVAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}
	identifier := d.Id()

	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return diag.Errorf("[vapp read] no identifier provided")
	}
	vapp, err := vdc.GetVAppByNameOrId(identifier, false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp. Removing from tfstate")
		d.SetId("")
		return nil
	}

	statusText, err := vapp.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	dSet(d, "status", vapp.VApp.Status)
	dSet(d, "status_text", statusText)
	dSet(d, "href", vapp.VApp.HREF)
	dSet(d, "description", vapp.VApp.Description)

	var vmList []string
	if vapp.VApp.Children != nil {
		for _, vm := range vapp.VApp.Children.VM {
			vmList = append(vmList, vm.Name)
		}
		sort.Strings(vmList)
	}

	err = d.Set("vm_list", vmList)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vapp.VApp.ID)
	return nil
}

func resourceVcdClonedVAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdVAppDelete(ctx, d, meta)
}

func resourceVcdClonedVAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("this resource doesn't support updates")
}
