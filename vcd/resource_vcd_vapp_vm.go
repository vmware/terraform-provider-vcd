package vcd

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type typeOfVm string

const (
	standaloneVmType typeOfVm = "vcd_vm"
	vappVmType       typeOfVm = "vcd_vapp_vm"
)

// Maintenance guide for VM code
//
// VM codebase grew to be quite complicated because of a few reasons:
// * It is a resource with many API quirks
// * There are 4 different go-vcloud-director SDK types for creating VM
//   * `types.InstantiateVmTemplateParams` (Standalone VM from template)
//   * `types.ReComposeVAppParams` (vApp VM from template)
//   * `types.RecomposeVAppParamsForEmptyVm` (Empty vApp VM)
//   * `types.CreateVmParams` (Empty Standalone VM)
// They also use different functions. All VM types are directly populated in resource code instead of
// pre-creating functions in go-vcloud-director SDK. This is done because we had to constantly change
// parent SDK functions and add a new field just because there is a new feature - be it storage
// profile, compute policy, or something else.
//
// The best chance to avoid breaking feature parity between all 4 types is to create them with minimal
// configuration and then perform additional updates in a code that is shared between all 4 types.
// Some features though must go into VM definition during creation (for example storage profile,
// CPU/RAM/sizing policy configuration)
//
//
// CPU/Memory management. CPU and Memory settings can come from 3 different places:
// * `memory` and `cpu` fields in resource definition
// * from specified sizing policy
// * inherited from template (only for template based VMs)
//
// There are quite a few `vm.Refresh` operations being run in creation code. This is done because
// some explicit endpoint API calls still change the main VM structure. Calling multiple different
// functions in a row poses a risk of re-applying older VM structure with subsequent calls.
// Time cost of `vm.Refresh` was measured to be from ~0.5s up to ~1.2s per call depending on client
// latency.
//
// Important notes.
// * Whenever calling VM update functions, be sure that VM is refreshed after applying them as the
// next function call may reset the value to old one as VM does not have flexible structure and
// often changing the name requires "reconfigure" operation.

func resourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVAppVmCreate,
		UpdateContext: resourceVcdVAppVmUpdate,
		ReadContext:   resourceVcdVAppVmRead,
		DeleteContext: resourceVcdVAppVmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVappVmImport,
		},
		Schema: vmSchemaFunc(vappVmType),
	}
}

// VM Schema is defined as global so that it can be directly accessible in other places
func vmSchemaFunc(vmType typeOfVm) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"vapp_name": {
			Type:        schema.TypeString,
			Required:    vmType == vappVmType,
			Optional:    vmType == standaloneVmType,
			Computed:    vmType == standaloneVmType,
			ForceNew:    vmType == vappVmType,
			Description: "The vApp this VM belongs to - Required, unless it is a standalone VM",
		},
		"vm_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("Type of VM: either '%s' or '%s'", vappVmType, standaloneVmType),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "A name for the VM, unique within the vApp",
		},
		"computer_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Computer name to assign to this virtual machine",
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
		"template_name": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Deprecated:    "Use `vapp_template_id` instead",
			Description:   "The name of the vApp Template to use",
			ConflictsWith: []string{"vapp_template_id"},
		},
		"vapp_template_id": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Description:   "The URN of the vApp Template to use",
			ConflictsWith: []string{"template_name", "catalog_name"},
		},
		"vm_name_in_template": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The name of the VM in vApp Template to use. In cases when vApp template has more than one VM",
		},
		"catalog_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Deprecated:    "You should use `vapp_template_id` or `boot_image_id` without the need of a catalog name",
			Description:   "The catalog name in which to find the given vApp Template or media for boot_image",
			ConflictsWith: []string{"vapp_template_id", "boot_image_id"},
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The VM description",
		},
		"memory": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			Description:  "The amount of RAM (in MB) to allocate to the VM",
			ValidateFunc: validateMultipleOf4(),
		},
		"memory_reservation": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The amount of RAM (in MB) reservation on the underlying virtualization infrastructure",
		},
		"memory_priority": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload",
			ValidateFunc: validation.StringInSlice([]string{"LOW", "NORMAL", "HIGH", "CUSTOM"}, false),
		},
		"memory_shares": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "Custom priority for the resource. This is a read-only, unless the `memory_priority` is CUSTOM",
		},
		"memory_limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The limit for how much of memory can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited.",
		},
		"cpus": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The number of virtual CPUs to allocate to the VM",
		},
		"cpu_cores": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The number of cores per socket",
		},
		"cpu_reservation": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The amount of MHz reservation on the underlying virtualization infrastructure",
		},
		"cpu_priority": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload",
			ValidateFunc: validation.StringInSlice([]string{"LOW", "NORMAL", "HIGH", "CUSTOM"}, false),
		},
		"cpu_shares": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "Custom priority for the resource. This is a read-only, unless the `cpu_priority` is CUSTOM",
		},
		"cpu_limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The limit for how much of CPU can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited.",
		},
		"metadata": {
			Type:          schema.TypeMap,
			Optional:      true,
			Computed:      true, // To be compatible with `metadata_entry`
			Description:   "Key value map of metadata to assign to this VM",
			Deprecated:    "Use metadata_entry instead",
			ConflictsWith: []string{"metadata_entry"},
		},
		"metadata_entry": metadataEntryResourceSchema("VM"),
		"href": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "VM Hyper Reference",
		},
		"accept_all_eulas": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Automatically accept EULA if OVA has it",
		},
		"power_on": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "A boolean value stating if this VM should be powered on",
		},
		"storage_profile": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Storage profile to override the default one",
		},
		"os_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Operating System type. Possible values can be found in documentation.",
		},
		"hardware_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Virtual Hardware Version (e.g.`vmx-14`, `vmx-13`, `vmx-12`, etc.)",
		},
		"boot_image": {
			Type:          schema.TypeString,
			Optional:      true,
			Deprecated:    "Use `boot_image_id` instead",
			Description:   "Media name to add as boot image.",
			ConflictsWith: []string{"boot_image_id"},
		},
		"boot_image_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "The URN of the media to use as boot image.",
			ConflictsWith: []string{"boot_image", "catalog_name"},
		},
		"network_dhcp_wait_seconds": {
			Optional:     true,
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Description: "Optional number of seconds to try and wait for DHCP IP (valid for " +
				"'network' block only)",
		},
		"network": {
			Optional:    true,
			Type:        schema.TypeList,
			Description: " A block to define network interface. Multiple can be used.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Required:     true,
						Type:         schema.TypeString,
						ValidateFunc: vmNetworkTypeValidator(vmType),
						Description:  "Network type to use: 'vapp', 'org' or 'none'. Use 'vapp' for vApp network, 'org' to attach Org VDC network. 'none' for empty NIC.",
					},
					"ip_allocation_mode": {
						Optional:     true,
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"POOL", "DHCP", "MANUAL", "NONE"}, false),
						Description:  "IP address allocation mode. One of POOL, DHCP, MANUAL, NONE",
					},
					"name": {
						ForceNew:    false,
						Optional:    true, // In case of type = none it is not required
						Type:        schema.TypeString,
						Description: "Name of the network this VM should connect to. Always required except for `type` `NONE`",
					},
					"ip": {
						Computed:     true,
						Optional:     true,
						Type:         schema.TypeString,
						ValidateFunc: checkEmptyOrSingleIP(), // Must accept empty string to ease using HCL interpolation
						Description:  "IP of the VM. Settings depend on `ip_allocation_mode`. Omitted or empty for DHCP, POOL, NONE. Required for MANUAL",
					},
					"is_primary": {
						Optional: true,
						Computed: true,
						// By default, if the value is omitted it will report schema change
						// on every terraform operation. The below function
						// suppresses such cases "" => "false" when applying.
						DiffSuppressFunc: falseBoolSuppress(),
						Type:             schema.TypeBool,
						Description:      "Set to true if network interface should be primary. First network card in the list will be primary by default",
					},
					"mac": {
						Computed:    true,
						Optional:    true,
						Type:        schema.TypeString,
						Description: "Mac address of network interface",
					},
					"adapter_type": {
						Type:             schema.TypeString,
						Computed:         true,
						Optional:         true,
						DiffSuppressFunc: suppressCase,
						Description:      "Network card adapter type. (e.g. 'E1000', 'E1000E', 'SRIOVETHERNETCARD', 'VMXNET3', 'PCNet32')",
					},
					"connected": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "It defines if NIC is connected or not.",
					},
				},
			},
		},
		"disk": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Independent disk name",
				},
				"bus_number": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Bus number on which to place the disk controller",
				},
				"unit_number": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unit number (slot) on the bus specified by BusNumber",
				},
				"size_in_mb": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The size of the disk in MB.",
				},
			}},
			Optional: true,
			Set:      resourceVcdVmIndependentDiskHash,
		},
		"override_template_disk": {
			Type:        schema.TypeSet,
			Optional:    true,
			ForceNew:    true,
			Description: "A block to match internal_disk interface in template. Multiple can be used. Disk will be matched by bus_type, bus_number and unit_number.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"bus_type": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice([]string{"ide", "parallel", "sas", "paravirtual", "sata", "nvme"}, false),
					Description:  "The type of disk controller. Possible values: ide, parallel( LSI Logic Parallel SCSI), sas(LSI Logic SAS (SCSI)), paravirtual(Paravirtual (SCSI)), sata, nvme",
				},
				"size_in_mb": {
					Type:        schema.TypeInt,
					ForceNew:    true,
					Required:    true,
					Description: "The size of the disk in MB.",
				},
				"bus_number": {
					Type:        schema.TypeInt,
					ForceNew:    true,
					Required:    true,
					Description: "The number of the SCSI or IDE controller itself.",
				},
				"unit_number": {
					Type:        schema.TypeInt,
					ForceNew:    true,
					Required:    true,
					Description: "The device number on the SCSI or IDE controller of the disk.",
				},
				"iops": {
					Type:        schema.TypeInt,
					ForceNew:    true,
					Optional:    true,
					Description: "Specifies the IOPS for the disk. Default is 0.",
				},
				"storage_profile": {
					Type:        schema.TypeString,
					ForceNew:    true,
					Optional:    true,
					Description: "Storage profile to override the VM default one",
				},
			}},
		},
		"internal_disk": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A block will show internal disk details",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"disk_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The disk ID.",
				},
				"bus_type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of disk controller. Possible values: ide, parallel( LSI Logic Parallel SCSI), sas(LSI Logic SAS (SCSI)), paravirtual(Paravirtual (SCSI)), sata, nvme",
				},
				"size_in_mb": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The size of the disk in MB.",
				},
				"bus_number": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The number of the SCSI or IDE controller itself.",
				},
				"unit_number": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The device number on the SCSI or IDE controller of the disk.",
				},
				"thin_provisioned": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Specifies whether the disk storage is pre-allocated or allocated on demand.",
				},
				"iops": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Specifies the IOPS for the disk. Default is 0.",
				},
				"storage_profile": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Storage profile to override the VM default one",
				},
			}},
		},
		"expose_hardware_virtualization": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Expose hardware-assisted CPU virtualization to guest OS.",
		},
		"guest_properties": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Key/value settings for guest properties",
		},
		"customization": {
			Optional:    true,
			Computed:    true,
			MinItems:    1,
			MaxItems:    1,
			Type:        schema.TypeList,
			Description: "Guest customization block",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"force": {
						ValidateFunc: noopValueWarningValidator(true,
							"Using 'true' value for field 'vcd_vapp_vm.customization.force' will reboot VM on every 'terraform apply' operation"),
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
						// This settings is used as a 'flag' and it does not matter what is set in the
						// state. If it is 'true' - then it means that 'update' procedure must set the
						// VM for customization at next boot and reboot it.
						DiffSuppressFunc: suppressFalse(),
						Description:      "'true' value will cause the VM to reboot on every 'apply' operation",
					},
					"enabled": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "'true' value will enable guest customization. It may occur on first boot or when 'force' is used",
					},
					"change_sid": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "'true' value will change SID. Applicable only for Windows VMs",
					},
					"allow_local_admin_password": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Allow local administrator password",
					},
					"must_change_password_on_first_login": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Require Administrator to change password on first login",
					},
					"auto_generate_password": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Auto generate password",
					},
					"admin_password": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Description: "Manually specify admin password",
					},
					"number_of_auto_logons": {
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						Description:  "Number of times to log on automatically. '0' - disabled.",
						ValidateFunc: validation.IntAtLeast(0),
					},
					"join_domain": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Enable this VM to join a domain",
					},
					"join_org_domain": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Use organization's domain for joining",
					},
					"join_domain_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Custom domain name for join",
					},
					"join_domain_user": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Username for custom domain name join",
					},
					"join_domain_password": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
						Description: "Password for custom domain name join",
					},
					"join_domain_account_ou": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Account organizational unit for domain name join",
					},
					"initscript": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Script to run on initial boot or with customization.force=true set",
					},
				},
			},
		},
		"cpu_hot_add_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "True if the virtual machine supports addition of virtual CPUs while powered on.",
		},
		"memory_hot_add_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "True if the virtual machine supports addition of memory while powered on.",
		},
		"prevent_update_power_off": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "True if the update of resource should fail when virtual machine power off needed.",
		},
		"sizing_policy_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true, // As it can get populated automatically by VDC default policy
			Description: "VM sizing policy ID. Has to be assigned to Org VDC.",
		},
		"placement_policy_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true, // As it can get populated automatically by VDC default policy
			Description: "VM placement policy ID. Has to be assigned to Org VDC.",
		},
		"security_tags": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Security tags to assign to this VM",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"status": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Shows the status code of the VM",
		},
		"status_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Shows the status of the VM",
		},
	}
}

// resourceVcdVAppVmCreate is an entry function for VM within vApp creation. It locks parent vApp and cascades down the
// other functions that need to be run
func resourceVcdVAppVmCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startTime := time.Now()
	util.Logger.Printf("[DEBUG] [VM create] started VM creation in vApp")

	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		return diag.Errorf("vApp name is mandatory for vApp VM (resource `vcd_vapp_vm`)")
	}

	// vApp lock must be acquired for VMs that are vApp members
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	err := genericResourceVmCreate(d, meta, vappVmType)
	if err != nil {
		return err
	}

	timeElapsed := time.Since(startTime)
	util.Logger.Printf("[DEBUG] [VM create] finished VM creation in vApp [took %f seconds]", timeElapsed.Seconds())

	return genericVcdVmRead(d, meta, "resource")
}

// genericResourceVmCreate does the following:
// * Executes VM create functions based on the type of VM (standalone or vApp member)
// * Runs additional customization functions which are common for all 4 types of VMs
func genericResourceVmCreate(d *schema.ResourceData, meta interface{}, vmType typeOfVm) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Deprecated: If at least Catalog Name and Template name are set - a VM from vApp template is being created
	isVmFromTemplateDeprecated := d.Get("catalog_name").(string) != "" && d.Get("template_name").(string) != ""

	isVmFromTemplate := d.Get("vapp_template_id").(string) != ""
	isEmptyVm := !isVmFromTemplate && !isVmFromTemplateDeprecated

	////////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code conditionally calls functions for VM creation from template and empty VMs
	// Each of these functions internally can handle both standalone VMs and VMs in vApp
	////////////////////////////////////////////////////////////////////////////////////////////////

	var err error
	var vm *govcd.VM
	switch {
	case isVmFromTemplateDeprecated || isVmFromTemplate:
		util.Logger.Printf("[DEBUG] [VM create] creating VM from template")
		vm, err = createVmFromTemplate(d, meta, vmType)
		if err != nil {
			return diag.Errorf("error creating VM from template: %s", err)
		}
	case isEmptyVm:
		util.Logger.Printf("[DEBUG] [VM create] creating empty VM")
		vm, err = createVmEmpty(d, meta, vmType)
		if err != nil {
			return diag.Errorf("error creating empty VM: %s", err)
		}
	default:
		return diag.Errorf("unknown VM type")
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code performs any additional operations that should be applied to all 4 VM types
	// and could not be applied during initial create VM API call.
	// Note. The final call should be VM power management.
	//
	// IMPORTANT. If any of the functions change VM structure, be sure to refresh `vm` structure so
	// that the next function does not accidentally apply old values.
	////////////////////////////////////////////////////////////////////////////////////////////////

	if err = vm.Refresh(); err != nil {
		return diag.Errorf("error refreshing VM: %s", err)
	}

	// Handle Metadata
	// Such schema fields are processed:
	// * metadata
	err = createOrUpdateMetadata(d, vm, "metadata")
	if err != nil {
		return diag.Errorf("error setting metadata: %s", err)
	}

	// Handle Hardware Virtualization setting (used for hypervisor nesting)
	// Such schema fields are processed:
	// * expose_hardware_virtualization
	err = handleExposeHardwareVirtualization(d, vm)
	if err != nil {
		return diag.Errorf("error updating hardware virtualization setting: %s", err)
	}

	// Handle Guest Properties
	// Such schema fields are processed:
	// * guest_properties
	err = addRemoveGuestProperties(d, vm)
	if err != nil {
		return diag.Errorf("error setting guest properties: %s", err)
	}

	// vm.VM structure contains ProductSection so it needs to be refreshed after
	// `addRemoveGuestProperties`
	if err = vm.Refresh(); err != nil {
		return diag.Errorf("error refreshing VM: %s", err)
	}

	// Handle Guest Customization Section
	// Such schema fields are processed:
	// * customization
	// * computer_name
	// * name
	err = updateGuestCustomizationSetting(d, vm)
	if err != nil {
		return diag.Errorf("error setting guest customization during creation: %s", err)
	}

	// vm.VM structure contains GuestCustomizationSection so it needs to be refreshed after
	// `updateGuestCustomizationSetting`
	if err = vm.Refresh(); err != nil {
		return diag.Errorf("error refreshing VM: %s", err)
	}

	// Explicitly setting CPU and Memory Hot Add settings
	// Note. VM Creation bodies allow specifying these values, but they are ignored therefore using
	// an explicit "/vmCapabilities" API endpoint
	// Such schema fields are processed:
	// * cpu_hot_add_enabled
	// * memory_hot_add_enabled
	_, err = vm.UpdateVmCpuAndMemoryHotAdd(d.Get("cpu_hot_add_enabled").(bool), d.Get("memory_hot_add_enabled").(bool))
	if err != nil {
		return diag.Errorf("error setting VM CPU/Memory HotAdd capabilities: %s", err)
	}

	// Independent disk handling
	// Such schema fields are processed:
	// * disk
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}
	err = attachDetachIndependentDisks(d, *vm, vdc)
	if err != nil {
		return diag.Errorf("error attaching-detaching independent disks when creating VM : %s", err)
	}

	// Handle Advanced compute settings CPU and Memory shares, limits and reservation
	// Such schema fields are processed:
	// * memory_priority
	// * memory_limit
	// * memory_shares
	// * memory_reservation
	// * cpu_priority
	// * cpu_limit
	// * cpu_shares
	// * cpu_reservation
	// Note. vm.Refresh happens inside `updateAdvancedComputeSettings`
	err = updateAdvancedComputeSettings(d, vm)
	if err != nil {
		return diag.Errorf("error applying advanced compute settings for VM %s : %s", vm.VM.Name, err)
	}

	// Handle VM Security Tags settings
	// Such schema fields are processed:
	// * security_tags
	if _, isSet := d.GetOk("security_tags"); isSet {
		err = createOrUpdateVmSecurityTags(d, vm)
		if err != nil {
			return diag.Errorf("[VM create] error creating security tags for VM %s : %s", vm.VM.Name, err)
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// VM power on handling is the last step, no other VM adjustment operations should be performed
	// after this
	////////////////////////////////////////////////////////////////////////////////////////////////

	// By default, the VM is created in POWERED_OFF state
	if d.Get("power_on").(bool) {
		// When customization is requested VM must be un-deployed before starting it
		customizationNeeded := isForcedCustomization(d.Get("customization"))
		if customizationNeeded {
			log.Printf("[TRACE] Powering on VM %s with forced customization", vm.VM.Name)
			err := vm.PowerOnAndForceCustomization()
			if err != nil {
				return diag.Errorf("failed powering on with customization: %s", err)
			}
		} else {
			task, err := vm.PowerOn()
			if err != nil {
				return diag.Errorf("error powering on: %s", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.Errorf(errorCompletingTask, err)
			}
		}

	}
	////////////////////////////////////////////////////////////////////////////////////////////////
	// VM power on handling was the last step, no other VM adjustment operations should be performed
	////////////////////////////////////////////////////////////////////////////////////////////////

	// Read function is called in wrapper functions `resourceVcdVAppVmCreate` and
	// `resourceVcdStandaloneVmCreate`
	return nil
}

// createVmFromTemplate is responsible for create VMs from template of two types:
// * Standalone VMs
// * VMs inside vApp (vApp VMs)
//
// Code flow has 3 layers:
// 1. Lookup common information, required for both types of VMs (Standalone and vApp child). Things such as
//   - Template to be used
//   - Network adapter configuration
//   - Storage profile configuration
//   - VM compute policy configuration
//
// 2. Perform VM creation operation based on type in separate switch/case
//   - standaloneVmType
//   - vAppVmType
//
// # This part includes defining initial structures for VM and also any explicitly required operations for that type of VM
//
// 3. Perform additional operations which are common for both types of VMs
//
// Note. VM Power ON (if it wasn't disabled in HCL configuration) occurs as last step after all configuration is done.
func createVmFromTemplate(d *schema.ResourceData, meta interface{}, vmType typeOfVm) (*govcd.VM, error) {
	vcdClient := meta.(*VCDClient)

	// Step 1 - lookup common information
	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	// Look up VM template inside vApp template - either specified by `vm_name_in_template` or the
	// first one in vApp
	vmTemplate, err := lookupvAppTemplateforVm(d, vcdClient, org, vdc)
	if err != nil {
		return nil, fmt.Errorf("error finding vApp template: %s", err)
	}

	// Look up vApp before setting up network configuration. Having a vApp set, will enable
	// additional network availability in vApp validations in `networksToConfig` function.
	// It is only possible for vApp VMs, as empty VMs will get their hidden vApps created after the
	// VM is created.
	var vapp *govcd.VApp
	if vmType == vappVmType {
		vappName := d.Get("vapp_name").(string)
		vapp, err = vdc.GetVAppByName(vappName, false)
		if err != nil {
			return nil, fmt.Errorf("[VM create] error finding vApp %s: %s", vappName, err)
		}
	}

	// Build up network configuration
	networkConnectionSection, err := networksToConfig(d, vapp)
	if err != nil {
		return nil, fmt.Errorf("unable to process network configuration: %s", err)
	}
	util.Logger.Printf("[VM create] networkConnectionSection %# v", pretty.Formatter(networkConnectionSection))

	// Lookup storage profile reference if it was specified
	storageProfilePtr, err := lookupStorageProfile(d, vdc)
	if err != nil {
		return nil, fmt.Errorf("error finding storage profile: %s", err)
	}

	// Look up compute policies
	sizingPolicy, err := lookupComputePolicy(d, vcdClient, "sizing_policy_id")
	if err != nil {
		return nil, fmt.Errorf("error finding sizing policy: %s", err)
	}
	placementPolicy, err := lookupComputePolicy(d, vcdClient, "placement_policy_id")
	if err != nil {
		return nil, fmt.Errorf("error finding placement policy: %s", err)
	}
	var vmComputePolicy *types.ComputePolicy
	if sizingPolicy != nil || placementPolicy != nil {
		vmComputePolicy = &types.ComputePolicy{}
		if sizingPolicy != nil {
			vmComputePolicy.VmSizingPolicy = &types.Reference{HREF: sizingPolicy.Href}
		}
		if placementPolicy != nil {
			vmComputePolicy.VmPlacementPolicy = &types.Reference{HREF: placementPolicy.Href}
		}
	}

	var vm *govcd.VM
	vmName := d.Get("name").(string)

	// Step 2 - perform VM creation operation based on type
	// VM creation uses different structure depending on if it is a standaloneVmType or vappVmType
	// These structures differ and one might accept all required parameters, while other
	switch vmType {
	case standaloneVmType:
		standaloneVmParams := types.InstantiateVmTemplateParams{
			Xmlns:            types.XMLNamespaceVCloud,
			Name:             vmName, // VM name post creation
			PowerOn:          false,  // VM will be powered on after all configuration is done
			AllEULAsAccepted: d.Get("accept_all_eulas").(bool),
			ComputePolicy:    vmComputePolicy,
			SourcedVmTemplateItem: &types.SourcedVmTemplateParams{
				Source: &types.Reference{
					HREF: vmTemplate.VAppTemplate.HREF,
					ID:   vmTemplate.VAppTemplate.ID,
					Type: vmTemplate.VAppTemplate.Type,
					Name: vmTemplate.VAppTemplate.Name,
				},
				VmGeneralParams: &types.VMGeneralParams{
					Description: d.Get("description").(string),
				},
				VmTemplateInstantiationParams: &types.InstantiationParams{
					// If a MAC address is specified for NIC - it does not get set with this call,
					// therefore an additional `vm.UpdateNetworkConnectionSection` is required.
					NetworkConnectionSection: &networkConnectionSection,
				},
				StorageProfile: storageProfilePtr,
			},
		}

		util.Logger.Printf("%# v", pretty.Formatter(standaloneVmParams))
		vm, err = vdc.CreateStandaloneVMFromTemplate(&standaloneVmParams)
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("[VM creation] error creating standalone VM from template %s : %s", vmName, err)
		}

		d.SetId(vm.VM.ID)

		util.Logger.Printf("[VM create] VM from template after creation %# v", pretty.Formatter(vm.VM))
		vapp, err = vm.GetParentVApp()
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("[VM creation] error retrieving vApp from standalone VM %s : %s", vmName, err)
		}
		util.Logger.Printf("[VM create] vApp after creation %# v", pretty.Formatter(vapp.VApp))
		dSet(d, "vapp_name", vapp.VApp.Name)
		dSet(d, "vm_type", string(standaloneVmType))

	////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Explicitly__ template based Standalone VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////

	case vappVmType:
		vappName := d.Get("vapp_name").(string)
		vapp, err = vdc.GetVAppByName(vappName, false)
		if err != nil {
			return nil, fmt.Errorf("[VM create] error finding vApp %s: %s", vappName, err)
		}

		vappVmParams := &types.ReComposeVAppParams{
			Ovf:              types.XMLNamespaceOVF,
			Xsi:              types.XMLNamespaceXSI,
			Xmlns:            types.XMLNamespaceVCloud,
			AllEULAsAccepted: d.Get("accept_all_eulas").(bool),
			Name:             vapp.VApp.Name,
			PowerOn:          false, // VM will be powered on after all configuration is done
			SourcedItem: &types.SourcedCompositionItemParam{
				Source: &types.Reference{
					HREF: vmTemplate.VAppTemplate.HREF,
					Name: vmName, // This VM name defines the VM name after creation
				},
				VMGeneralParams: &types.VMGeneralParams{
					Description: d.Get("description").(string),
				},
				InstantiationParams: &types.InstantiationParams{
					// If a MAC address is specified for NIC - it does not get set with this call,
					// therefore an additional `vm.UpdateNetworkConnectionSection` is required.
					NetworkConnectionSection: &networkConnectionSection,
				},
				ComputePolicy:  vmComputePolicy,
				StorageProfile: storageProfilePtr,
			},
		}

		vm, err = vapp.AddRawVM(vappVmParams)
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("[VM creation] error getting VM %s : %s", vmName, err)
		}

		d.SetId(vm.VM.ID)
		dSet(d, "vm_type", string(vappVmType))

	////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Explicitly__ template based vApp VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////

	default:
		return nil, fmt.Errorf("unknown VM type %s", vmType)
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Only__ template based VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////////

	// If a MAC address is specified for NIC - it does not get set with initial create call therefore
	// running additional update call to make sure it is set correctly

	err = vm.UpdateNetworkConnectionSection(&networkConnectionSection)
	if err != nil {
		return nil, fmt.Errorf("unable to setup network configuration for empty VM %s", err)
	}

	// Refresh VM to have the latest structure
	if err := vm.Refresh(); err != nil {
		return nil, fmt.Errorf("error refreshing VM %s : %s", vmName, err)
	}

	// update existing internal disks in template (it is only applicable to VMs created
	// Such fields are processed:
	// * override_template_disk
	err = updateTemplateInternalDisks(d, meta, *vm)
	if err != nil {
		dSet(d, "override_template_disk", nil)
		return nil, fmt.Errorf("error managing internal disks : %s", err)
	}

	if err := vm.Refresh(); err != nil {
		return nil, fmt.Errorf("error refreshing VM %s : %s", vmName, err)
	}

	// OS Type and Hardware version should only be changed if specified. (Only applying to VMs from
	// templates as empty VMs require this by default)
	// Such fields are processed:
	// * os_type
	// * hardware_version
	err = updateHardwareVersionAndOsType(d, vm)
	if err != nil {
		return nil, fmt.Errorf("error updating hardware version and OS type : %s", err)
	}

	if err := vm.Refresh(); err != nil {
		return nil, fmt.Errorf("error refreshing VM %s : %s", vmName, err)
	}

	// Template VMs require CPU/Memory setting
	// Lookup CPU values either from schema or from sizing policy. If nothing is set - it will be
	// inherited from template
	var cpuCores, cpuCoresPerSocket *int
	var memory *int64
	if sizingPolicy != nil {
		cpuCores, cpuCoresPerSocket, memory, err = getCpuMemoryValues(d, sizingPolicy.VdcComputePolicyV2)
	} else {
		cpuCores, cpuCoresPerSocket, memory, err = getCpuMemoryValues(d, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting CPU/Memory compute values: %s", err)
	}

	if cpuCores != nil || cpuCoresPerSocket != nil {
		err = vm.ChangeCPUAndCoreCount(cpuCores, cpuCoresPerSocket)
		if err != nil {
			return nil, fmt.Errorf("error changing CPU settings: %s", err)
		}

		if err := vm.Refresh(); err != nil {
			return nil, fmt.Errorf("error refreshing VM %s : %s", vmName, err)
		}
	}

	if memory != nil {
		err = vm.ChangeMemory(*memory)
		if err != nil {
			return nil, fmt.Errorf("error setting memory size from schema for VM from template: %s", err)
		}

		if err := vm.Refresh(); err != nil {
			return nil, fmt.Errorf("error refreshing VM %s : %s", vmName, err)
		}
	}

	return vm, nil
}

// createVmEmpty is responsible for creating empty VMs of two types:
// * Standalone VMs
// * VMs inside vApp (vApp VMs)
//
// Code flow has 3 layers:
// 1. Lookup common information, required for both types of VMs (Standalone and vApp child). Things such as
//   - OS Type
//   - Hardware version
//   - Storage profile configuration
//   - VM compute policy configuration
//   - Boot image
//
// 2. Perform VM creation operation based on type in separate switch/case
//   - standaloneVmType
//   - vAppVmType
//
// # This part includes defining initial structures for VM and also any explicitly required operations for that type of VM
//
// 3. Perform additional operations which are common for both types of VMs
//
// Note. VM Power ON (if it wasn't disabled in HCL configuration) occurs as last step after all configuration is done.
func createVmEmpty(d *schema.ResourceData, meta interface{}, vmType typeOfVm) (*govcd.VM, error) {
	util.Logger.Printf("[TRACE] Creating empty VM: %s", d.Get("name").(string))

	vcdClient := meta.(*VCDClient)
	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	var vapp *govcd.VApp

	if vmType == vappVmType {
		vappName := d.Get("vapp_name").(string)
		vapp, err = vdc.GetVAppByName(vappName, false)
		if err != nil {
			return nil, fmt.Errorf("[VM create] error finding vApp for empty VM %s: %s", vappName, err)
		}
	}

	var ok bool
	var osType interface{}

	if osType, ok = d.GetOk("os_type"); !ok {
		return nil, fmt.Errorf("`os_type` is required when creating empty VM")
	}

	var hardWareVersion interface{}
	if hardWareVersion, ok = d.GetOk("hardware_version"); !ok {
		return nil, fmt.Errorf("`hardware_version` is required when creating empty VM")
	}

	var computerName interface{}
	if computerName, ok = d.GetOk("computer_name"); !ok {
		return nil, fmt.Errorf("`computer_name` is required when creating empty VM")
	}

	_, bootImageIdSet := d.GetOk("boot_image_id")
	_, bootImageName := d.GetOk("boot_image")
	var bootImage *types.Media
	if bootImageIdSet || bootImageName {
		var bootMediaIdentifier string
		var mediaRecord *govcd.MediaRecord
		var err error
		if bootImageIdSet {
			bootMediaIdentifier = d.Get("boot_image_id").(string)
			mediaRecord, err = vcdClient.QueryMediaById(bootMediaIdentifier)
		} else {
			// Deprecated way of using media item
			bootMediaIdentifier = d.Get("boot_image").(string)
			var catalogName interface{}
			if catalogName, ok = d.GetOk("catalog_name"); !ok {
				return nil, fmt.Errorf("`catalog_name` is required when creating empty VM with boot_image")
			}
			var catalog *govcd.Catalog
			catalog, err = org.GetCatalogByName(catalogName.(string), false)
			if err != nil {
				return nil, fmt.Errorf("error finding catalog %s: %s", catalogName, err)
			}
			mediaRecord, err = catalog.QueryMedia(bootMediaIdentifier)
		}
		if err != nil {
			return nil, fmt.Errorf("[VM creation] error getting boot image %s: %s", bootMediaIdentifier, err)
		}

		// This workaround is to check that the Media file is synchronized in catalog, even if it isn't an iso
		// file. It's not officially documented that IsIso==true means that, but it's the only way we have at the moment.
		if !mediaRecord.MediaRecord.IsIso {
			return nil, fmt.Errorf("[VM creation] error getting boot image %s: Media is not synchronized in the catalog", bootMediaIdentifier)
		}

		bootImage = &types.Media{HREF: mediaRecord.MediaRecord.HREF}
	}

	storageProfilePtr, err := lookupStorageProfile(d, vdc)
	if err != nil {
		return nil, fmt.Errorf("error finding storage profile: %s", err)
	}

	customizationSection := &types.GuestCustomizationSection{}
	customizationSection.ComputerName = computerName.(string)

	// Process parameters from 'customization' block
	updateCustomizationSection(d.Get("customization"), d, customizationSection)

	isVirtualCpuType64 := strings.Contains(d.Get("os_type").(string), "64")
	virtualCpuType := "VM32"
	if isVirtualCpuType64 {
		virtualCpuType = "VM64"
	}

	// Look up compute policies
	sizingPolicy, err := lookupComputePolicy(d, vcdClient, "sizing_policy_id")
	if err != nil {
		return nil, fmt.Errorf("error finding sizing policy: %s", err)
	}
	placementPolicy, err := lookupComputePolicy(d, vcdClient, "placement_policy_id")
	if err != nil {
		return nil, fmt.Errorf("error finding placement policy: %s", err)
	}
	var vmComputePolicy *types.ComputePolicy
	if sizingPolicy != nil || placementPolicy != nil {
		vmComputePolicy = &types.ComputePolicy{}
		if sizingPolicy != nil {
			vmComputePolicy.VmSizingPolicy = &types.Reference{HREF: sizingPolicy.Href}
		}
		if placementPolicy != nil {
			vmComputePolicy.VmPlacementPolicy = &types.Reference{HREF: placementPolicy.Href}
		}
	}

	// Lookup CPU/Memory parameters
	var cpuCores, cpuCoresPerSocket *int
	var memory *int64
	if sizingPolicy != nil {
		cpuCores, cpuCoresPerSocket, memory, err = getCpuMemoryValues(d, sizingPolicy.VdcComputePolicyV2)
	} else {
		cpuCores, cpuCoresPerSocket, memory, err = getCpuMemoryValues(d, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting CPU/Memory compute values: %s", err)
	}
	// Wrap memory definition into a suitable construct if it is set
	var memoryResourceMb *types.MemoryResourceMb
	if memory != nil {
		memoryResourceMb = &types.MemoryResourceMb{Configured: *memory}
	}

	vmName := d.Get("name").(string)
	var newVm *govcd.VM

	switch vmType {
	case standaloneVmType:
		var mediaReference *types.Reference
		if bootImage != nil {
			mediaReference = &types.Reference{
				HREF: bootImage.HREF,
				ID:   bootImage.ID,
				Type: bootImage.Type,
				Name: bootImage.Name,
			}
		}
		params := types.CreateVmParams{
			Xmlns:       types.XMLNamespaceVCloud,
			Name:        vmName,
			PowerOn:     false, // Power on is handled at the end of VM creation process
			Description: d.Get("description").(string),
			CreateVm: &types.Vm{
				Name:          vmName,
				ComputePolicy: vmComputePolicy,
				// BUG in VCD, do not allow empty NetworkConnectionSection, so we pass simplest
				// network configuration and after VM created update with real config
				NetworkConnectionSection: &types.NetworkConnectionSection{
					PrimaryNetworkConnectionIndex: 0,
					NetworkConnection: []*types.NetworkConnection{
						{Network: "none", NetworkConnectionIndex: 0, IPAddress: "any", IsConnected: false, IPAddressAllocationMode: "NONE"}},
				},
				VmSpecSection: &types.VmSpecSection{
					Modified:          addrOf(true),
					Info:              "Virtual Machine specification",
					OsType:            osType.(string),
					CpuResourceMhz:    &types.CpuResourceMhz{Configured: 0},
					NumCpus:           cpuCores,
					NumCoresPerSocket: cpuCoresPerSocket,
					MemoryResourceMb:  memoryResourceMb,

					// can be created with resource internal_disk
					DiskSection:     &types.DiskSection{DiskSettings: []*types.DiskSettings{}},
					HardwareVersion: &types.HardwareVersion{Value: hardWareVersion.(string)}, // need support older version vCD
					VirtualCpuType:  virtualCpuType,
				},
				GuestCustomizationSection: customizationSection,
				StorageProfile:            storageProfilePtr,
			},
			Media: mediaReference,
		}

		newVm, err = vdc.CreateStandaloneVm(&params)
		if err != nil {
			return nil, err
		}
		// VM created - store its ID
		d.SetId(newVm.VM.ID)
		dSet(d, "vm_type", string(standaloneVmType))

	////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Explicitly__ empty  Standalone VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////

	case vappVmType:
		recomposeVAppParamsForEmptyVm := &types.RecomposeVAppParamsForEmptyVm{
			XmlnsVcloud: types.XMLNamespaceVCloud,
			XmlnsOvf:    types.XMLNamespaceOVF,
			PowerOn:     false, // Power on is handled at the end of VM creation process
			CreateItem: &types.CreateItem{
				Name: vmName,
				// BUG in VCD, do not allow empty NetworkConnectionSection, so we pass simplest
				// network configuration and after VM created update with real config
				NetworkConnectionSection: &types.NetworkConnectionSection{
					PrimaryNetworkConnectionIndex: 0,
					NetworkConnection: []*types.NetworkConnection{
						{Network: "none", NetworkConnectionIndex: 0, IPAddress: "any", IsConnected: false, IPAddressAllocationMode: "NONE"}},
				},
				StorageProfile:            storageProfilePtr,
				ComputePolicy:             vmComputePolicy,
				Description:               d.Get("description").(string),
				GuestCustomizationSection: customizationSection,
				VmSpecSection: &types.VmSpecSection{
					Modified:          addrOf(true),
					Info:              "Virtual Machine specification",
					OsType:            osType.(string),
					NumCpus:           cpuCores,
					NumCoresPerSocket: cpuCoresPerSocket,
					MemoryResourceMb:  memoryResourceMb,
					// can be created with resource internal_disk
					DiskSection:     &types.DiskSection{DiskSettings: []*types.DiskSettings{}},
					HardwareVersion: &types.HardwareVersion{Value: hardWareVersion.(string)}, // need support older version vCD
					VirtualCpuType:  virtualCpuType,
				},
				BootImage: bootImage,
			},
		}

		util.Logger.Printf("[VM create - add empty VM] recomposeVAppParamsForEmptyVm %# v", pretty.Formatter(recomposeVAppParamsForEmptyVm))
		newVm, err = vapp.AddEmptyVm(recomposeVAppParamsForEmptyVm)
		if err != nil {
			return nil, fmt.Errorf("[VM creation] error creating VM %s : %s", vmName, err)
		}
		// VM created - store its ID
		d.SetId(newVm.VM.ID)
		dSet(d, "vm_type", string(vappVmType))

	////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Explicitly__ empty vApp VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////

	default:
		return nil, fmt.Errorf("unknown VM type %s", vmType)
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// This part of code handles additional VM create operations, which can not be set during
	// initial VM creation.
	// __Only__ empty VMs are addressed here.
	////////////////////////////////////////////////////////////////////////////////////////////////

	util.Logger.Printf("[VM create] VM after creation %# v", pretty.Formatter(newVm.VM))
	vapp, err = newVm.GetParentVApp()
	if err != nil {
		return nil, fmt.Errorf("[VM creation] error retrieving vApp from standalone VM %s : %s", vmName, err)
	}
	util.Logger.Printf("[VM create] vApp after creation %# v", pretty.Formatter(vapp.VApp))
	dSet(d, "vapp_name", vapp.VApp.Name)

	// Due to the Bug in VCD, VM creation works only with Org VDC networks, not vApp networks - we
	// setup network configuration with update.

	// firstly cleanup dummy network as network adapter type can't be changed
	err = newVm.UpdateNetworkConnectionSection(&types.NetworkConnectionSection{})
	if err != nil {
		return nil, fmt.Errorf("unable to setup network configuration for empty VM %s", err)
	}

	networkConnectionSection, err := networksToConfig(d, vapp)
	if err != nil {
		return nil, fmt.Errorf("unable to setup network configuration for empty VM: %s", err)
	}

	// add real network configuration
	err = newVm.UpdateNetworkConnectionSection(&networkConnectionSection)
	if err != nil {
		return nil, fmt.Errorf("unable to setup network configuration for empty VM %s", err)
	}

	return newVm, nil
}

func resourceVcdVAppVmUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericResourceVcdVmUpdate(d, meta, vappVmType)
}

func genericResourceVcdVmUpdate(d *schema.ResourceData, meta interface{}, vmType typeOfVm) diag.Diagnostics {
	log.Printf("[DEBUG] [VM update] started with lock")
	vcdClient := meta.(*VCDClient)

	// When there is more then one VM in a vApp Terraform will try to parallelise their creation.
	// However, vApp throws errors when simultaneous requests are executed.
	// To avoid them, below block is using mutex as a workaround,
	// so that the one vApp VMs are created not in parallelisation.

	if vmType == vappVmType {
		vcdClient.lockParentVapp(d)
		defer vcdClient.unLockParentVapp(d)
	}

	// Exit early only if "network_dhcp_wait_seconds" is changed because this field only supports
	// update so that its value can be written into statefile and be accessible in read function
	if onlyHasChange("network_dhcp_wait_seconds", vmSchemaFunc(vmType), d) {
		log.Printf("[DEBUG] [VM update] exiting early because only 'network_dhcp_wait_seconds' has change")
		return genericVcdVmRead(d, meta, "resource")
	}

	err := resourceVmHotUpdate(d, meta, vmType)
	if err != nil {
		return err
	}

	return resourceVcdVAppVmUpdateExecute(d, meta, "update", vmType, nil)
}

func resourceVmHotUpdate(d *schema.ResourceData, meta interface{}, vmType typeOfVm) diag.Diagnostics {
	_, _, vdc, vapp, _, vm, err := getVmFromResource(d, meta, vmType)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Get("memory_hot_add_enabled").(bool) && d.HasChange("memory") {
		err = vm.ChangeMemory(int64(d.Get("memory").(int)))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.Get("cpu_hot_add_enabled").(bool) && d.HasChange("cpus") {
		err = vm.ChangeCPU(d.Get("cpus").(int), d.Get("cpu_cores").(int))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// * Primary NIC cannot be removed on a powered on VM
	if d.HasChange("network") && !isPrimaryNicRemoved(d) {
		networkConnectionSection, err := networksToConfig(d, vapp)
		if err != nil {
			return diag.Errorf("unable to setup network configuration for update: %s", err)
		}
		err = vm.UpdateNetworkConnectionSection(&networkConnectionSection)
		if err != nil {
			return diag.Errorf("unable to update network configuration: %s", err)
		}
	}

	err = createOrUpdateMetadata(d, vm, "metadata")
	if err != nil {
		return diag.FromErr(err)
	}

	err = addRemoveGuestProperties(d, vm)
	if err != nil {
		return diag.FromErr(err)
	}

	sizingId, newSizingId := d.GetChange("sizing_policy_id")
	placementId, newPlacementId := d.GetChange("placement_policy_id")

	sizingPolicyChanged := d.HasChange("sizing_policy_id")
	placementPolicyChanged := d.HasChange("placement_policy_id")
	if !sizingPolicyChanged {
		// As sizing_policy_id is Computed+Optional, the only way to unset it should be to write `sizing_policy_id = ""`
		// in the HCL. However, when this is done, Terraform SDK (terraform-plugin-sdk v2.24.0) doesn't detect this change:
		// d.HasChange() returns false, d.GetChange returns both old values and d.Get returns old value,
		// hence `sizingPolicyChanged` will be always false.
		// We need to inspect the raw HCL to get the correct value.
		hclMap := d.GetRawConfig().AsValueMap()
		if hclValue, ok := hclMap["sizing_policy_id"]; ok && !hclValue.IsNull() && strings.TrimSpace(hclValue.AsString()) == "" {
			sizingId = ""
		}
	}
	if !placementPolicyChanged {
		// Same as above
		hclMap := d.GetRawConfig().AsValueMap()
		if hclValue, ok := hclMap["placement_policy_id"]; ok && !hclValue.IsNull() && strings.TrimSpace(hclValue.AsString()) == "" {
			placementId = ""
		}
	}

	if sizingPolicyChanged || placementPolicyChanged {
		// This is done because we need to update both policies at the same time, as not populating one of them will make
		// that policy to be unassigned from the VM.
		// Therefore, we need to use the old value if the policy didn't change to preserve it, or update to the new if it changed.
		if placementPolicyChanged {
			placementId = newPlacementId
		}
		if sizingPolicyChanged {
			sizingId = newSizingId
		}
		_, err = vm.UpdateComputePolicyV2(sizingId.(string), placementId.(string), "")
		if err != nil {
			return diag.Errorf("error updating compute policy: %s", err)
		}
	}

	storageProfileName := d.Get("storage_profile").(string)
	if d.HasChange("storage_profile") && storageProfileName != "" {
		storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
		if err != nil {
			return diag.Errorf("[vm update] error retrieving storage profile %s : %s", storageProfileName, err)
		}
		_, err = vm.UpdateStorageProfile(storageProfile.HREF)
		if err != nil {
			return diag.Errorf("error updating changing storage profile to %s: %s", storageProfileName, err)
		}
	}

	return nil
}

func resourceVcdVAppVmUpdateExecute(d *schema.ResourceData, meta interface{}, executionType string, vmType typeOfVm, computePolicy *types.VdcComputePolicy) diag.Diagnostics {
	log.Printf("[DEBUG] [VM update] started without lock")

	_, org, vdc, vapp, identifier, vm, err := getVmFromResource(d, meta, vmType)
	if err != nil {
		return diag.FromErr(err)
	}

	vmStatusBeforeUpdate, err := vm.GetStatus()
	if err != nil {
		return diag.Errorf("[VM update] error getting VM (%s) status before update: %s", identifier, err)
	}

	// Check if the user requested for forced customization of VM
	customizationNeeded := isForcedCustomization(d.Get("customization"))

	// Update guest customization if any of the customization related fields have changed
	if d.HasChanges("customization", "computer_name", "name") {
		log.Printf("[TRACE] VM %s customization has changes: customization(%t), computer_name(%t), name(%t)",
			vm.VM.Name, d.HasChange("customization"), d.HasChange("computer_name"), d.HasChange("name"))
		err = updateGuestCustomizationSetting(d, vm)
		if err != nil {
			return diag.Errorf("errors updating guest customization: %s", err)
		}

	}

	if d.HasChanges("memory_reservation", "memory_priority", "memory_shares", "memory_limit",
		"cpu_reservation", "cpu_priority", "cpu_limit", "cpu_shares") {
		err = updateAdvancedComputeSettings(d, vm)
		if err != nil {
			return diag.Errorf("[VM update] error advanced compute settings for standalone VM %s : %s", vm.VM.Name, err)
		}
	}

	memoryNeedsColdChange := false
	cpusNeedsColdChange := false
	networksNeedsColdChange := false
	if executionType == "update" {
		if !d.Get("memory_hot_add_enabled").(bool) && d.HasChange("memory") {
			memoryNeedsColdChange = true
		}
		if !d.Get("cpu_hot_add_enabled").(bool) && d.HasChange("cpus") {
			cpusNeedsColdChange = true
		}
		if d.HasChange("network") && isPrimaryNicRemoved(d) {
			networksNeedsColdChange = true
		}
	}
	if executionType == "create" && len(d.Get("network").([]interface{})) > 0 {
		networksNeedsColdChange = true
	}
	log.Printf("[TRACE] VM %s requires cold changes: memory(%t), cpu(%t), network(%t)", vm.VM.Name, memoryNeedsColdChange, cpusNeedsColdChange, networksNeedsColdChange)

	// this represents fields which have to be changed in cold (with VM power off)
	if d.HasChanges("cpu_cores", "power_on", "disk", "expose_hardware_virtualization", "boot_image",
		"hardware_version", "os_type", "description", "cpu_hot_add_enabled",
		"memory_hot_add_enabled") || memoryNeedsColdChange || cpusNeedsColdChange || networksNeedsColdChange {

		log.Printf("[TRACE] VM %s has changes: memory(%t), cpus(%t), cpu_cores(%t), power_on(%t), disk(%t), expose_hardware_virtualization(%t),"+
			" boot_image(%t), hardware_version(%t), os_type(%t), description(%t), cpu_hot_add_enabled(%t), memory_hot_add_enabled(%t), network(%t)",
			vm.VM.Name, d.HasChange("memory"), d.HasChange("cpus"), d.HasChange("cpu_cores"), d.HasChange("power_on"), d.HasChange("disk"),
			d.HasChange("expose_hardware_virtualization"), d.HasChange("boot_image"), d.HasChange("hardware_version"),
			d.HasChange("os_type"), d.HasChange("description"), d.HasChange("cpu_hot_add_enabled"), d.HasChange("memory_hot_add_enabled"), d.HasChange("network"))

		if vmStatusBeforeUpdate != "POWERED_OFF" {
			if d.Get("prevent_update_power_off").(bool) && executionType == "update" {
				return diag.Errorf("update stopped: VM needs to power off to change properties, but `prevent_update_power_off` is `true`")
			}
			log.Printf("[DEBUG] Un-deploying VM %s for offline update. Previous state %s",
				vm.VM.Name, vmStatusBeforeUpdate)
			task, err := vm.Undeploy()
			if err != nil {
				return diag.Errorf("error triggering undeploy for VM %s: %s", vm.VM.Name, err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.Errorf("error waiting for undeploy task for VM %s: %s", vm.VM.Name, err)
			}
		}

		// detaching independent disks - only possible when VM power off
		if d.HasChange("disk") {
			err = attachDetachIndependentDisks(d, *vm, vdc)
			if err != nil {
				errAttachedDisk := updateStateOfAttachedIndependentDisks(d, *vm)
				if errAttachedDisk != nil {
					dSet(d, "disk", nil)
					return diag.Errorf("error reading attached disks : %s and internal error : %s", errAttachedDisk, err)
				}
				return diag.Errorf("error attaching-detaching  disks when updating resource : %s", err)
			}
		}

		if memoryNeedsColdChange || executionType == "create" {
			memory, isMemorySet := d.GetOk("memory")
			isMemoryComingFromSizingPolicy := computePolicy != nil && (computePolicy.Memory != nil && !isMemorySet)
			if isMemoryComingFromSizingPolicy && isMemorySet {
				logForScreen("vcd_vapp_vm", fmt.Sprintf("WARNING: sizing policy is specifying a memory of %d that won't be overriden by `memory` attribute", *computePolicy.Memory))
			}

			if !isMemoryComingFromSizingPolicy {
				err = vm.ChangeMemory(int64(memory.(int)))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if d.HasChange("cpu_cores") {
			err = vm.ChangeCPU(d.Get("cpus").(int), d.Get("cpu_cores").(int))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if cpusNeedsColdChange || (executionType == "create") {
			cpus, isCpusSet := d.GetOk("cpus")
			cpuCores, isCpuCoresSet := d.GetOk("cpu_cores")
			isCpuComingFromSizingPolicy := computePolicy != nil && ((computePolicy.CPUCount != nil && !isCpusSet) || (computePolicy.CoresPerSocket != nil && !isCpuCoresSet))
			if isCpuComingFromSizingPolicy && isCpusSet {
				logForScreen("vcd_vapp_vm", fmt.Sprintf("WARNING: sizing policy is specifying CPU count of %d that won't be overriden by `cpus` attribute", *computePolicy.CPUCount))
			}
			if isCpuComingFromSizingPolicy && isCpuCoresSet {
				logForScreen("vcd_vapp_vm", fmt.Sprintf("WARNING: sizing policy is specifying %d CPU cores that won't be overriden by `cpu_cores` attribute", *computePolicy.CoresPerSocket))
			}

			if !isCpuComingFromSizingPolicy {
				err = vm.ChangeCPU(cpus.(int), cpuCores.(int))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if networksNeedsColdChange {
			networkConnectionSection, err := networksToConfig(d, vapp)
			if err != nil {
				return diag.Errorf("unable to setup network configuration for update: %s", err)
			}
			err = vm.UpdateNetworkConnectionSection(&networkConnectionSection)
			if err != nil {
				return diag.Errorf("unable to update network configuration: %s", err)
			}
		}

		if d.HasChange("expose_hardware_virtualization") {

			task, err := vm.ToggleHardwareVirtualization(d.Get("expose_hardware_virtualization").(bool))
			if err != nil {
				return diag.Errorf("error changing hardware assisted virtualization: %s", err)
			}

			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// updating fields of VM spec section
		if d.HasChange("hardware_version") || d.HasChange("os_type") || d.HasChange("description") {
			vmSpecSection := vm.VM.VmSpecSection
			description := vm.VM.Description
			if d.HasChange("hardware_version") {
				vmSpecSection.HardwareVersion = &types.HardwareVersion{Value: d.Get("hardware_version").(string)}
			}
			if d.HasChange("os_type") {
				vmSpecSection.OsType = d.Get("os_type").(string)
			}

			if d.HasChange("description") {
				description = d.Get("description").(string)
			}

			_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
			if err != nil {
				return diag.Errorf("error changing VM spec section: %s", err)
			}
		}

		if d.HasChange("cpu_hot_add_enabled") || d.HasChange("memory_hot_add_enabled") {
			_, err := vm.UpdateVmCpuAndMemoryHotAdd(d.Get("cpu_hot_add_enabled").(bool), d.Get("memory_hot_add_enabled").(bool))
			if err != nil {
				return diag.Errorf("error changing VM capabilities: %s", err)
			}
		}

		// we detach boot image if it's value change to empty.
		bootImage := d.Get("boot_image")
		if d.HasChange("boot_image") && bootImage.(string) == "" {
			previousBootImageValue, _ := d.GetChange("boot_image")
			previousCatalogName, _ := d.GetChange("catalog_name")
			catalog, err := org.GetCatalogByName(previousCatalogName.(string), false)
			if err != nil {
				return diag.Errorf("[VM Update] error finding catalog %s: %s", previousCatalogName, err)
			}
			result, err := catalog.GetMediaByName(previousBootImageValue.(string), false)
			if err != nil {
				return diag.Errorf("[VM Update] error getting boot image %s : %s", previousBootImageValue, err)
			}

			task, err := vm.HandleEjectMedia(org, previousCatalogName.(string), result.Media.Name)
			if err != nil {
				return diag.Errorf("error: %#v", err)
			}

			err = task.WaitTaskCompletion(true)
			if err != nil {
				return diag.Errorf("error: %#v", err)
			}
		}
	}

	// Update the security tags
	if d.HasChange("security_tags") {
		err = createOrUpdateVmSecurityTags(d, vm)
		if err != nil {
			return diag.Errorf("[VM Update] error updating security tags for VM %s : %s", vm.VM.Name, err)
		}
	}

	// If the VM was powered off during update but it has to be powered on
	if d.Get("power_on").(bool) {
		vmStatus, err := vm.GetStatus()
		if err != nil {
			return diag.Errorf("error getting VM status before ensuring it is powered on: %s", err)
		}

		// Simply power on if customization is not requested
		if !customizationNeeded && vmStatus != "POWERED_ON" {
			log.Printf("[DEBUG] Powering on VM %s after update. Previous state %s", vm.VM.Name, vmStatus)
			task, err := vm.PowerOn()
			if err != nil {
				return diag.Errorf("error powering on: %s", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.Errorf(errorCompletingTask, err)
			}
		}

		// When customization is requested VM must be un-deployed before starting it
		if customizationNeeded {
			log.Printf("[TRACE] forced customization for VM %s was requested. Current state %s",
				vm.VM.Name, vmStatus)

			if vmStatus != "POWERED_OFF" {
				log.Printf("[TRACE] VM %s is in state %s. Un-deploying", vm.VM.Name, vmStatus)
				task, err := vm.Undeploy()
				if err != nil {
					return diag.Errorf("error triggering undeploy for VM %s: %s", vm.VM.Name, err)
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return diag.Errorf("error waiting for undeploy task for VM %s: %s", vm.VM.Name, err)
				}
			}

			log.Printf("[TRACE] Powering on VM %s with forced customization", vm.VM.Name)
			err = vm.PowerOnAndForceCustomization()
			if err != nil {
				return diag.Errorf("failed powering on with customization: %s", err)
			}
		}
	}

	log.Printf("[DEBUG] [VM update] finished")
	return genericVcdVmRead(d, meta, "resource")
}

func resourceVcdVAppVmRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmRead(d, meta, "resource")
}

func genericVcdVmRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	log.Printf("[DEBUG] [VM read] started with origin %s", origin)
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[VM read ]"+errorRetrievingOrgAndVdc, err)
	}

	isStandalone := false
	setVmType, ok := d.GetOk("vm_type")
	if ok {
		isStandalone = setVmType.(string) == string(standaloneVmType)
	}
	var vapp *govcd.VApp
	var vm *govcd.VM
	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return diag.Errorf("[VM read] neither name or ID were set for this VM")
	}
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		vm, err = vdc.QueryVmById(identifier)
		if govcd.IsNotFound(err) {
			vmByName, listStr, errByName := getVmByName(vcdClient, vdc, identifier)
			if errByName != nil && listStr != "" {
				return diag.Errorf("[VM read] error retrieving VM %s by name: %s\n%s\n%s", identifier, errByName, listStr, err)
			}
			vm = vmByName
			err = errByName
		}
	} else {
		vapp, err = vdc.GetVAppByName(vappName, false)
		if err != nil {
			additionalMessage := ""
			if isStandalone {
				additionalMessage = fmt.Sprintf("\nAdding a vApp name to a standalone VM is not allowed." +
					"Please use 'vcd_vapp_vm' resource to specify vApp")
				dSet(d, "vapp_name", "")
			}
			if govcd.IsNotFound(err) {
				log.Printf("[VM read] error finding vApp '%s': %s%s. Removing it from state.", vappName, err, additionalMessage)
				d.SetId("")
				return nil
			}
			return diag.Errorf("[VM read] error finding vApp '%s': %s%s", vappName, err, additionalMessage)
		}
		vm, err = vapp.GetVMByNameOrId(identifier, false)
	}
	if err != nil {
		if origin == "resource" {
			log.Printf("[DEBUG] Unable to find VM. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return diag.Errorf("[VM read] error getting VM %s : %s", identifier, err)
	}
	if vapp == nil {
		vapp, err = vm.GetParentVApp()
		if err != nil {
			return diag.Errorf("[VM read] error retrieving parent vApp for VM %s: %s", vm.VM.Name, err)
		}
	}

	var computedVmType string
	if vapp.VApp.IsAutoNature {
		computedVmType = string(standaloneVmType)
	} else {
		computedVmType = string(vappVmType)
	}
	// org, vdc, and vapp_name are already implicitly set
	dSet(d, "name", vm.VM.Name)
	dSet(d, "vapp_name", vapp.VApp.Name)
	dSet(d, "description", vm.VM.Description)
	d.SetId(vm.VM.ID)
	dSet(d, "vm_type", computedVmType)

	networks, err := readNetworks(d, *vm, *vapp, vdc)
	if err != nil {
		return diag.Errorf("[VM read] failed reading network details: %s", err)
	}

	err = d.Set("network", networks)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "href", vm.VM.HREF)
	dSet(d, "expose_hardware_virtualization", vm.VM.NestedHypervisorEnabled)
	dSet(d, "cpu_hot_add_enabled", vm.VM.VMCapabilities.CPUHotAddEnabled)
	dSet(d, "memory_hot_add_enabled", vm.VM.VMCapabilities.MemoryHotAddEnabled)

	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.MemoryResourceMb != nil {
		dSet(d, "memory", vm.VM.VmSpecSection.MemoryResourceMb.Configured)
		dSet(d, "memory_priority", vm.VM.VmSpecSection.MemoryResourceMb.SharesLevel)
		if vm.VM.VmSpecSection.MemoryResourceMb.Reservation != nil {
			dSet(d, "memory_reservation", vm.VM.VmSpecSection.MemoryResourceMb.Reservation)
		}
		if vm.VM.VmSpecSection.MemoryResourceMb.Limit != nil {
			dSet(d, "memory_limit", vm.VM.VmSpecSection.MemoryResourceMb.Limit)
		}
		if vm.VM.VmSpecSection.MemoryResourceMb.Shares != nil {
			dSet(d, "memory_shares", vm.VM.VmSpecSection.MemoryResourceMb.Shares)
		}
	}
	dSet(d, "cpus", vm.VM.VmSpecSection.NumCpus)
	dSet(d, "cpu_cores", vm.VM.VmSpecSection.NumCoresPerSocket)
	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.CpuResourceMhz != nil {
		if vm.VM.VmSpecSection.CpuResourceMhz.Reservation != nil {
			dSet(d, "cpu_reservation", vm.VM.VmSpecSection.CpuResourceMhz.Reservation)
		}
		if vm.VM.VmSpecSection.CpuResourceMhz.Limit != nil {
			dSet(d, "cpu_limit", vm.VM.VmSpecSection.CpuResourceMhz.Limit)
		}
		if vm.VM.VmSpecSection.CpuResourceMhz.Shares != nil {
			dSet(d, "cpu_shares", vm.VM.VmSpecSection.CpuResourceMhz.Shares)
		}
		dSet(d, "cpu_priority", vm.VM.VmSpecSection.CpuResourceMhz.SharesLevel)
	}

	if vm.VM.StorageProfile != nil {
		dSet(d, "storage_profile", vm.VM.StorageProfile.Name)
	}

	// update guest properties
	guestProperties, err := vm.GetProductSectionList()
	if err != nil {
		return diag.Errorf("[VM read] unable to read guest properties: %s", err)
	}

	err = setGuestProperties(d, guestProperties)
	if err != nil {
		return diag.Errorf("[VM read] unable to set guest properties in state: %s", err)
	}

	err = updateStateOfInternalDisks(d, *vm)
	if err != nil {
		dSet(d, "internal_disk", nil)
		return diag.Errorf("[VM read] error reading internal disks : %s", err)
	}

	err = updateStateOfAttachedIndependentDisks(d, *vm)
	if err != nil {
		dSet(d, "disk", nil)
		return diag.Errorf("[VM read] error reading attached disks : %s", err)
	}

	if err := setGuestCustomizationData(d, vm); err != nil {
		return diag.Errorf("error storing customization block: %s", err)
	}

	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.HardwareVersion != nil && vm.VM.VmSpecSection.HardwareVersion.Value != "" {
		dSet(d, "hardware_version", vm.VM.VmSpecSection.HardwareVersion.Value)
	}
	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.OsType != "" {
		dSet(d, "os_type", vm.VM.VmSpecSection.OsType)
	}

	if vm.VM.ComputePolicy != nil {
		dSet(d, "sizing_policy_id", "")
		if vm.VM.ComputePolicy.VmSizingPolicy != nil {
			dSet(d, "sizing_policy_id", vm.VM.ComputePolicy.VmSizingPolicy.ID)
		}
		dSet(d, "placement_policy_id", "")
		if vm.VM.ComputePolicy.VmPlacementPolicy != nil {
			dSet(d, "placement_policy_id", vm.VM.ComputePolicy.VmPlacementPolicy.ID)
		}
	}

	entitySecurityTags, err := vm.GetVMSecurityTags()
	if err != nil {
		return diag.Errorf("[VM read] unable to read VM security tags: %s", err)
	}
	dSet(d, "security_tags", convertStringsToTypeSet(entitySecurityTags.Tags))

	statusText, err := vm.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	dSet(d, "status", vm.VM.Status)
	dSet(d, "status_text", statusText)

	diagErr := updateMetadataInState(d, vcdClient, "vcd_vapp_vm", vm)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] [VM read] finished with origin %s", origin)
	return nil
}

func resourceVcdVAppVmDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] [VM delete] started")

	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)

	if err != nil {
		return diag.Errorf("[VM delete] error finding vApp '%s': %s", vappName, err)
	}

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return diag.Errorf("[VM delete] neither ID or name provided")
	}
	vm, err := vapp.GetVMByNameOrId(identifier, false)

	if err != nil {
		return diag.Errorf("[VM delete] error getting VM %s : %s", identifier, err)
	}

	// If it is a standalone VM, we remove it in one go
	if vapp.VApp.IsAutoNature {
		err = vm.Delete()
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	util.Logger.Printf("[VM delete] vApp before deletion %# v", pretty.Formatter(vapp.VApp))
	util.Logger.Printf("[VM delete] VM before deletion %# v", pretty.Formatter(vm.VM))
	deployed, err := vm.IsDeployed()
	if err != nil {
		return diag.Errorf("error getting VM deploy status: %s", err)
	}

	log.Printf("[TRACE] VM deploy Status: %t", deployed)
	if deployed {
		log.Printf("[TRACE] Undeploying VM: %s", vm.VM.Name)
		task, err := vm.Undeploy()
		if err != nil {
			return diag.Errorf("error Undeploying: %s", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error Undeploying VM: %s", err)
		}
	}

	// to avoid race condition for independent disks is attached or not - detach before removing vm
	existingDisks := getVmIndependentDisks(*vm)

	for _, existingDiskHref := range existingDisks {
		disk, err := vdc.GetDiskByHref(existingDiskHref)
		if err != nil {
			return diag.Errorf("did not find disk `%s`: %s", existingDiskHref, err)
		}

		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}
		task, err := vm.DetachDisk(attachParams)
		if err != nil {
			return diag.Errorf("error detaching disk `%s`: %s", existingDiskHref, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error waiting detaching disk task to finish`%s`: %s", existingDiskHref, err)
		}
	}

	log.Printf("[TRACE] Removing VM: %s", vm.VM.Name)
	err = vapp.RemoveVM(*vm)
	if err != nil {
		return diag.Errorf("error deleting: %s", err)
	}
	log.Printf("[DEBUG] [VM delete] finished")
	return nil
}

// resourceVcdVappVmImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_vm.VM_name
// Example import path for VM within vApp (_the_id_string_): org-name.vdc-name.vapp-name.vm-name
// Example import path for standalone VM (_the_id_string_): org-name.vdc-name.vm-name
// or
// Example import path for standalone VM (_the_id_string_): org-name.vdc-name.vm-ID
//
// The VM identifier can be either the VM name or its ID
// If we are dealing with standalone VMs, the name can retrieve duplicates. When that happens, the import fails
// and a list of VM information (ID, guest OS, network, IP) is returned
func resourceVcdVappVmImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	vcdClient := meta.(*VCDClient)

	var vapp *govcd.VApp
	var vm *govcd.VM
	var orgName string
	var vdcName string
	var vappName string
	var vmIdentifier string
	standaloneVm := false

	// With three arguments, we expect a standalone VM
	if len(resourceURI) == 3 {
		// standalone VM
		orgName, vdcName, vmIdentifier = resourceURI[0], resourceURI[1], resourceURI[2]
		standaloneVm = true
	} else {
		// With 4 arguments, it's a VM within a vApp
		if len(resourceURI) != 4 {
			return nil, fmt.Errorf("[VM import] resource name must be specified as org-name.vdc-name.vapp-name.vm-name-or-ID")
		}
		orgName, vdcName, vappName, vmIdentifier = resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]
	}

	isUuid := extractUuid(vmIdentifier) != ""
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[VM import] unable to find VDC %s: %s ", vdcName, err)
	}

	if standaloneVm {
		if isUuid {
			vm, err = vdc.QueryVmById(vmIdentifier)
			if err != nil {
				return nil, fmt.Errorf("[VM import] error retrieving VM %s by ID: %s", vmIdentifier, err)
			}
		} else {
			vmByName, listStr, err := getVmByName(vcdClient, vdc, vmIdentifier)
			if err != nil {
				return nil, fmt.Errorf("[VM import] error retrieving VM %s by name: %s\n%s", vmIdentifier, err, listStr)
			}
			vm = vmByName
		}

	} else {
		vapp, err = vdc.GetVAppByName(vappName, false)
		if err != nil {
			return nil, fmt.Errorf("[VM import] error retrieving vApp %s: %s", vappName, err)
		}
		vm, err = vapp.GetVMByNameOrId(vmIdentifier, false)
		if err != nil {
			return nil, fmt.Errorf("[VM import] error retrieving VM %s: %s", vmIdentifier, err)
		}
	}

	dSet(d, "name", vmIdentifier)
	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "vapp_name", vappName)
	d.SetId(vm.VM.ID)
	return []*schema.ResourceData{d}, nil
}
