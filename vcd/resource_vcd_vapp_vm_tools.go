package vcd

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func lookupvAppTemplateforVm(d *schema.ResourceData, org *govcd.Org, vdc *govcd.Vdc) (govcd.VAppTemplate, error) {
	catalogName := d.Get("catalog_name").(string)
	templateName := d.Get("template_name").(string)

	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog %s: %s", catalogName, err)
	}

	var vappTemplate govcd.VAppTemplate
	if vmNameInTemplate, ok := d.GetOk("vm_name_in_template"); ok {
		vmInTemplateRecord, err := vdc.QueryVappVmTemplate(catalogName, templateName, vmNameInTemplate.(string))
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error quering VM template %s: %s", vmNameInTemplate, err)
		}
		util.Logger.Printf("[VM create] vmInTemplateRecord %# v", pretty.Formatter(vmInTemplateRecord))
		returnedVappTemplate, err := catalog.GetVappTemplateByHref(vmInTemplateRecord.HREF)
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error quering VM template %s: %s", vmNameInTemplate, err)
		}
		util.Logger.Printf("[VM create] returned VappTemplate %#v", pretty.Formatter(returnedVappTemplate))
		vappTemplate = *returnedVappTemplate
	} else {
		catalogItem, err := catalog.GetCatalogItemByName(templateName, false)
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog item %s: %s", templateName, err)
		}
		vappTemplate, err = catalogItem.GetVAppTemplate()
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("[VM create] error finding VAppTemplate %s: %s", templateName, err)
		}

	}

	return vappTemplate, nil
}

func lookupStorageProfile(d *schema.ResourceData, vdc *govcd.Vdc) (*types.Reference, error) {
	// If no storage profile lookup was requested - bail out early and return nil reference
	storageProfileName := d.Get("storage_profile").(string)
	if storageProfileName == "" {
		return nil, nil
	}

	storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
	if err != nil {
		return nil, fmt.Errorf("[vm creation] error retrieving storage profile %s : %s", storageProfileName, err)
	}

	return &storageProfile, nil

}

func lookupComputePolicy(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcComputePolicy, *types.ComputePolicy, error) {
	var sizingPolicy *types.VdcComputePolicy
	var vmComputePolicy *types.ComputePolicy
	if value, ok := d.GetOk("sizing_policy_id"); ok {
		vdcComputePolicy, err := vcdClient.Client.GetVdcComputePolicyById(value.(string))
		if err != nil {
			return nil, nil, fmt.Errorf("error getting sizing policy %s: %s", value.(string), err)
		}
		sizingPolicy = vdcComputePolicy.VdcComputePolicy
		if vdcComputePolicy.Href == "" {
			return nil, nil, fmt.Errorf("empty sizing policy HREF detected")
		}
		vmComputePolicy = &types.ComputePolicy{
			VmSizingPolicy: &types.Reference{HREF: vdcComputePolicy.Href},
		}
		util.Logger.Printf("[VM create] sizingPolicy (%s) %# v", vdcComputePolicy.Href, pretty.Formatter(sizingPolicy))
	}

	return sizingPolicy, vmComputePolicy, nil
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
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The name of the vApp Template to use",
			//RequiredWith: []string{"template_name", "catalog_name"},
		},
		"vm_name_in_template": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The name of the VM in vApp Template to use. In cases when vApp template has more than one VM",
			//RequiredWith: []string{"vm_name_in_template", "catalog_name", "template_name"},
		},
		"catalog_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The catalog name in which to find the given vApp Template or media for boot_image",
			//RequiredWith: []string{"catalog_name", "template_name", "boot_image"},
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
			Type:     schema.TypeMap,
			Optional: true,
			// For now underlying go-vcloud-director repo only supports
			// a value of type String in this map.
			Description: "Key value map of metadata to assign to this VM",
		},
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
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Media name to add as boot image.",
			//RequiredWith: []string{"boot_image", "catalog_name"},
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
			Computed:    true,
			Description: "VM sizing policy ID. Has to be assigned to Org VDC.",
		},
	}
}

// vmTemplatefromVappTemplate returns a given VM from a vApp template
// If no name is provided, it returns the first VM from the template
func vmTemplatefromVappTemplate(name string, vappTemplate *types.VAppTemplate) *types.VAppTemplate {
	if vappTemplate.Children == nil {
		return nil
	}
	for _, vm := range vappTemplate.Children.VM {
		if name == vm.Name || name == "" {
			return vm
		}
	}
	return nil
}

func updateAdvancedComputeSettings(d *schema.ResourceData, vm *govcd.VM) error {
	vmSpecSection := vm.VM.VmSpecSection
	description := vm.VM.Description
	// update treats same values as changes and fails, with no values provided - no changes are made for that section
	vmSpecSection.DiskSection = nil

	updateNeeded := false

	if memorySharesLevel, ok := d.GetOk("memory_priority"); ok {
		vmSpecSection.MemoryResourceMb.SharesLevel = memorySharesLevel.(string)
		updateNeeded = true
	}

	if memoryLimit, ok := d.GetOk("memory_limit"); ok {
		vmSpecSection.MemoryResourceMb.Limit = takeInt64Pointer(int64(memoryLimit.(int)))
		updateNeeded = true
	}

	if memoryShares, ok := d.GetOk("memory_shares"); ok {
		vmSpecSection.MemoryResourceMb.Shares = takeIntPointer(memoryShares.(int))
		updateNeeded = true
	}

	if memoryReservation, ok := d.GetOk("memory_reservation"); ok {
		vmSpecSection.MemoryResourceMb.Reservation = takeInt64Pointer(int64(memoryReservation.(int)))
		updateNeeded = true
	}

	if memorySharesLevel, ok := d.GetOk("cpu_priority"); ok {
		vmSpecSection.CpuResourceMhz.SharesLevel = memorySharesLevel.(string)
		updateNeeded = true
	}

	if memoryLimit, ok := d.GetOk("cpu_limit"); ok {
		vmSpecSection.CpuResourceMhz.Limit = takeInt64Pointer(int64(memoryLimit.(int)))
		updateNeeded = true
	}

	if memoryShares, ok := d.GetOk("cpu_shares"); ok {
		vmSpecSection.CpuResourceMhz.Shares = takeIntPointer(memoryShares.(int))
		updateNeeded = true
	}

	if memoryReservation, ok := d.GetOk("cpu_reservation"); ok {
		vmSpecSection.CpuResourceMhz.Reservation = takeInt64Pointer(int64(memoryReservation.(int)))
		updateNeeded = true
	}

	if updateNeeded {
		err := updateVmSpecSection(vmSpecSection, vm, description)
		if err != nil {
			return fmt.Errorf("error updating advanced compute settings: %s", err)
		}
	}
	return nil
}

// isItVappNetwork checks if it is an vApp network (not vApp Org Network)
func isItVappNetwork(vAppNetworkName string, vapp govcd.VApp) (bool, error) {
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error getting vApp networks: %s", err)
	}

	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vAppNetworkName &&
			govcd.IsVappNetwork(networkConfig.Configuration) {
			log.Printf("[TRACE] vApp network found: %s", vAppNetworkName)
			return true, nil
		}
	}

	return false, fmt.Errorf("configured vApp network isn't found: %s", vAppNetworkName)
}

type diskParams struct {
	name       string
	busNumber  *int
	unitNumber *int
}

func expandDisksProperties(v interface{}) ([]diskParams, error) {
	v = v.(*schema.Set).List()
	l := v.([]interface{})
	diskParamsArray := make([]diskParams, 0, len(l))

	for _, raw := range l {
		if raw == nil {
			continue
		}
		original := raw.(map[string]interface{})
		addParams := diskParams{name: original["name"].(string)}

		busNumber := original["bus_number"].(string)
		if busNumber != "" {
			convertedBusNumber, err := strconv.Atoi(busNumber)
			if err != nil {
				return nil, fmt.Errorf("value `%s` bus_number is not number. err: %s", busNumber, err)
			}
			addParams.busNumber = &convertedBusNumber
		}

		unitNumber := original["unit_number"].(string)
		if unitNumber != "" {
			convertedUnitNumber, err := strconv.Atoi(unitNumber)
			if err != nil {
				return nil, fmt.Errorf("value `%s` unit_number is not number. err: %s", unitNumber, err)
			}
			addParams.unitNumber = &convertedUnitNumber
		}

		diskParamsArray = append(diskParamsArray, addParams)
	}
	return diskParamsArray, nil
}

func getVmIndependentDisks(vm govcd.VM) []string {

	var disks []string
	// We use VirtualHardwareSection because in time of implementation we didn't have access to VmSpecSection which we used for internal disks.
	for _, item := range vm.VM.VirtualHardwareSection.Item {
		// disk resource type is 17
		if item.ResourceType == 17 && item.HostResource[0].Disk != "" {
			disks = append(disks, item.HostResource[0].Disk)
		}
	}
	return disks
}

func addRemoveGuestProperties(d *schema.ResourceData, vm *govcd.VM) error {
	if d.HasChange("guest_properties") {
		vmProperties, err := getGuestProperties(d)
		if err != nil {
			return fmt.Errorf("unable to convert guest properties to data structure")
		}

		log.Printf("[TRACE] Updating VM guest properties")
		_, err = vm.SetProductSectionList(vmProperties)
		if err != nil {
			return fmt.Errorf("error setting guest properties: %s", err)
		}
	}
	return nil
}

// isPrimaryNicRemoved checks if new schema has a primary NIC at all
func isPrimaryNicRemoved(d *schema.ResourceData) bool {
	_, newNetworkRaw := d.GetChange("network")
	newNetworks := newNetworkRaw.([]interface{})

	var foundPrimaryNic bool
	for _, newNet := range newNetworks {
		netMap := newNet.(map[string]interface{})
		isPrimary := netMap["is_primary"].(bool)
		if isPrimary {
			foundPrimaryNic = true
			break
		}
	}

	return !foundPrimaryNic
}

func updateVmSpecSection(vmSpecSection *types.VmSpecSection, vm *govcd.VM, description string) error {
	// add missing values if not inherited from template, otherwise API throws error if some value is nil
	if vmSpecSection.MemoryResourceMb.Reservation == nil {
		vmSpecSection.MemoryResourceMb.Reservation = takeInt64Pointer(int64(0))
	}
	if vmSpecSection.MemoryResourceMb.Limit == nil {
		vmSpecSection.MemoryResourceMb.Limit = takeInt64Pointer(int64(-1))
	}
	if vmSpecSection.MemoryResourceMb.SharesLevel == "" {
		vmSpecSection.MemoryResourceMb.SharesLevel = "NORMAL"
	}
	if vmSpecSection.CpuResourceMhz.Reservation == nil {
		vmSpecSection.CpuResourceMhz.Reservation = takeInt64Pointer(int64(0))
	}
	if vmSpecSection.CpuResourceMhz.Limit == nil {
		vmSpecSection.CpuResourceMhz.Limit = takeInt64Pointer(int64(-1))
	}
	if vmSpecSection.CpuResourceMhz.SharesLevel == "" {
		vmSpecSection.CpuResourceMhz.SharesLevel = "NORMAL"
	}
	_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
	if err != nil {
		return fmt.Errorf("error updating Vm Spec Section: %s", err)
	}
	return nil
}

func getVmFromResource(d *schema.ResourceData, meta interface{}, vmType typeOfVm) (*VCDClient, *govcd.Org, *govcd.Vdc, *govcd.VApp, string, *govcd.VM, error) {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, nil, nil, nil, "", nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)

	if err != nil {
		additionalMessage := ""
		if vmType == standaloneVmType {
			additionalMessage = fmt.Sprintf("\nAdding a vApp name to a standalone VM is not allowed." +
				"Please use 'vcd_vapp_vm' resource to specify vApp")
			dSet(d, "vapp_name", "")
		}

		return nil, nil, nil, nil, "", nil, fmt.Errorf("[getVmFromResource] error finding vApp '%s': %s%s", vappName, err, additionalMessage)
	}

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, nil, nil, nil, "", nil, fmt.Errorf("[VM update] neither name or ID was set")
	}

	vm, err := vapp.GetVMByNameOrId(identifier, false)

	if err != nil {
		d.SetId("")
		return nil, nil, nil, nil, "", nil, fmt.Errorf("[VM update] error getting VM %s: %s", identifier, err)
	}
	return vcdClient, org, vdc, vapp, identifier, vm, nil
}

// updates attached disks to latest state. Removed not needed and add new ones
func attachDetachIndependentDisks(d *schema.ResourceData, vm govcd.VM, vdc *govcd.Vdc) error {
	oldValues, newValues := d.GetChange("disk")

	attachDisks := newValues.(*schema.Set).Difference(oldValues.(*schema.Set))
	detachDisks := oldValues.(*schema.Set).Difference(newValues.(*schema.Set))

	removeDiskProperties, err := expandDisksProperties(detachDisks)
	if err != nil {
		return err
	}

	for _, diskData := range removeDiskProperties {
		disk, err := vdc.QueryDisk(diskData.name)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", diskData.name, err)
		}

		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}
		if diskData.unitNumber != nil {
			attachParams.UnitNumber = diskData.unitNumber
		}
		if diskData.busNumber != nil {
			attachParams.BusNumber = diskData.busNumber
		}

		task, err := vm.DetachDisk(attachParams)
		if err != nil {
			return fmt.Errorf("error detaching disk `%s` to vm %s", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete detaching disk `%s` to vm %s", diskData.name, err)
		}
	}

	// attach new independent disks
	newDiskProperties, err := expandDisksProperties(attachDisks)
	if err != nil {
		return err
	}

	sort.SliceStable(newDiskProperties, func(i, j int) bool {
		if newDiskProperties[i].busNumber == newDiskProperties[j].busNumber {
			return *newDiskProperties[i].unitNumber > *newDiskProperties[j].unitNumber
		}
		return *newDiskProperties[i].busNumber > *newDiskProperties[j].busNumber
	})

	for _, diskData := range newDiskProperties {
		disk, err := vdc.QueryDisk(diskData.name)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", diskData.name, err)
		}

		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}
		if diskData.unitNumber != nil {
			attachParams.UnitNumber = diskData.unitNumber
		}
		if diskData.busNumber != nil {
			attachParams.BusNumber = diskData.busNumber
		}

		task, err := vm.AttachDisk(attachParams)
		if err != nil {
			return fmt.Errorf("error attaching disk `%s` to vm %s", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete attaching disk `%s` to vm %s", diskData.name, err)
		}
	}
	return nil
}

func updateStateOfAttachedIndependentDisks(d *schema.ResourceData, vm govcd.VM) error {

	existingDisks := getVmIndependentDisks(vm)
	transformed := schema.NewSet(resourceVcdVmIndependentDiskHash, []interface{}{})

	for _, existingDiskHref := range existingDisks {
		diskSettings, err := getIndependentDiskFromVmDisks(vm, existingDiskHref)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %s", existingDiskHref, err)
		}
		newValues := map[string]interface{}{
			"name":        diskSettings.Disk.Name,
			"bus_number":  strconv.Itoa(diskSettings.BusNumber),
			"unit_number": strconv.Itoa(diskSettings.UnitNumber),
			"size_in_mb":  diskSettings.SizeMb,
		}

		transformed.Add(newValues)
	}

	return d.Set("disk", transformed)
}

// getIndependentDiskFromVmDisks finds independent disk in VM disk list.
func getIndependentDiskFromVmDisks(vm govcd.VM, diskHref string) (*types.DiskSettings, error) {
	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return nil, govcd.ErrorEntityNotFound
	}
	for _, disk := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if disk.Disk != nil && disk.Disk.HREF == diskHref {
			return disk, nil
		}
	}
	return nil, govcd.ErrorEntityNotFound
}

func updateStateOfInternalDisks(d *schema.ResourceData, vm govcd.VM) error {
	err := vm.Refresh()
	if err != nil {
		return err
	}

	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return fmt.Errorf("[updateStateOfInternalDisks] VmSpecSection part is missing")
	}
	existingInternalDisks := vm.VM.VmSpecSection.DiskSection.DiskSettings
	var internalDiskList []map[string]interface{}
	for _, internalDisk := range existingInternalDisks {
		// API shows internal disk and independent disks in one list. If disk.Disk != nil then it's independent disk
		// We use VmSpecSection as it is newer type than VirtualHardwareSection. It is used by HTML5 vCD client, has easy understandable structure.
		// VirtualHardwareSection has undocumented relationships between elements and very hard to use without issues for internal disks.
		if internalDisk.Disk == nil {
			newValue := map[string]interface{}{
				"disk_id":         internalDisk.DiskId,
				"bus_type":        internalDiskBusTypesFromValues[internalDisk.AdapterType],
				"size_in_mb":      int(internalDisk.SizeMb),
				"bus_number":      internalDisk.BusNumber,
				"unit_number":     internalDisk.UnitNumber,
				"storage_profile": internalDisk.StorageProfile.Name,
			}

			// There have been real cases where these values were `nil` and caused panic of plugin.
			if internalDisk.Iops != nil {
				newValue["iops"] = int(*internalDisk.Iops)
			}
			if internalDisk.ThinProvisioned != nil {
				newValue["thin_provisioned"] = *internalDisk.ThinProvisioned
			}

			internalDiskList = append(internalDiskList, newValue)
		}
	}

	return d.Set("internal_disk", internalDiskList)
}

func updateTemplateInternalDisks(d *schema.ResourceData, meta interface{}, vm govcd.VM) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	if vm.VM.VmSpecSection == nil || vm.VM.VmSpecSection.DiskSection == nil {
		return fmt.Errorf("[updateTemplateInternalDisks] VmSpecSection part is missing")
	}

	diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings

	var storageProfilePrt *types.Reference
	var overrideVmDefault bool

	internalDisksList := d.Get("override_template_disk").(*schema.Set).List()

	if len(internalDisksList) == 0 {
		return nil
	}

	for _, internalDisk := range internalDisksList {
		internalDiskProvidedConfig := internalDisk.(map[string]interface{})
		diskCreatedByTemplate := getMatchedDisk(internalDiskProvidedConfig, diskSettings)

		storageProfileName := internalDiskProvidedConfig["storage_profile"].(string)
		if storageProfileName != "" {
			storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
			if err != nil {
				return fmt.Errorf("[vm creation] error retrieving storage profile %s : %s", storageProfileName, err)
			}
			storageProfilePrt = &storageProfile
			overrideVmDefault = true
		} else {
			storageProfilePrt = vm.VM.StorageProfile
			overrideVmDefault = false
		}

		if diskCreatedByTemplate == nil {
			return fmt.Errorf("[vm creation] disk with bus type %s, bus number %d and unit number %d not found",
				internalDiskProvidedConfig["bus_type"].(string), internalDiskProvidedConfig["bus_number"].(int), internalDiskProvidedConfig["unit_number"].(int))
		}

		// Update details of internal disk for disk existing in template
		if value, ok := internalDiskProvidedConfig["iops"]; ok {
			iops := int64(value.(int))
			diskCreatedByTemplate.Iops = &iops
		}

		// value is required but not treated.
		isThinProvisioned := true
		diskCreatedByTemplate.ThinProvisioned = &isThinProvisioned

		diskCreatedByTemplate.SizeMb = int64(internalDiskProvidedConfig["size_in_mb"].(int))
		diskCreatedByTemplate.StorageProfile = storageProfilePrt
		diskCreatedByTemplate.OverrideVmDefault = overrideVmDefault
	}

	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings
	_, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return fmt.Errorf("error updating VM disks: %s", err)
	}

	return nil
}

// getMatchedDisk returns matched disk by adapter type, bus number and unit number
func getMatchedDisk(internalDiskProvidedConfig map[string]interface{}, diskSettings []*types.DiskSettings) *types.DiskSettings {
	for _, diskSetting := range diskSettings {
		if diskSetting.AdapterType == internalDiskBusTypes[internalDiskProvidedConfig["bus_type"].(string)] &&
			diskSetting.BusNumber == internalDiskProvidedConfig["bus_number"].(int) &&
			diskSetting.UnitNumber == internalDiskProvidedConfig["unit_number"].(int) {
			return diskSetting
		}
	}
	return nil
}

// getGuestProperties returns a struct for setting guest properties
func getGuestProperties(d *schema.ResourceData) (*types.ProductSectionList, error) {
	guestProperties := d.Get("guest_properties")
	guestProp := convertToStringMap(guestProperties.(map[string]interface{}))
	vmProperties := &types.ProductSectionList{
		ProductSection: &types.ProductSection{
			Info:     "Custom properties",
			Property: []*types.Property{},
		},
	}
	for key, value := range guestProp {
		log.Printf("[TRACE] Adding guest property: key=%s, value=%s to object", key, value)
		oneProp := &types.Property{
			UserConfigurable: true,
			Type:             "string",
			Key:              key,
			Label:            key,
			Value:            &types.Value{Value: value},
		}
		vmProperties.ProductSection.Property = append(vmProperties.ProductSection.Property, oneProp)
	}

	return vmProperties, nil
}

// setGuestProperties sets guest properties into state
func setGuestProperties(d *schema.ResourceData, properties *types.ProductSectionList) error {
	data := make(map[string]string)

	// if properties object does not have actual properties - do not set it at all (leave Terraform 'null')
	log.Printf("[TRACE] Setting empty properties into statefile because no properties were specified")
	if properties == nil || properties.ProductSection == nil || len(properties.ProductSection.Property) == 0 {
		return nil
	}

	for _, prop := range properties.ProductSection.Property {
		// if a value was set - use it
		if prop.Value != nil {
			data[prop.Key] = prop.Value.Value
		}
	}

	log.Printf("[TRACE] Setting properties into statefile")
	return d.Set("guest_properties", data)
}

// isForcedCustomization checks "customization" block in resource and checks if the value of field "force"
// is set to "true". It returns false if the value is not set or is set to false
func isForcedCustomization(customizationBlock interface{}) bool {
	customizationSlice := customizationBlock.([]interface{})

	if len(customizationSlice) != 1 {
		return false
	}

	cust := customizationSlice[0]
	fc := cust.(map[string]interface{})
	forceCust, ok := fc["force"]
	forceCustBool := forceCust.(bool)

	if !ok || !forceCustBool {
		return false
	}

	return true
}

func resourceVcdVmIndependentDiskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	_, err := buf.WriteString(fmt.Sprintf("%s-",
		m["name"].(string)))
	// We use the name and no other identifier to calculate the hash
	// With the VM resource, we assume that disks have a unique name.
	// In the event that this is not true, we return an error
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	return hashcodeString(buf.String())
}

// networksToConfig converts terraform schema for 'network' and converts to types.NetworkConnectionSection
// which is used for creating new VM
func networksToConfig(d *schema.ResourceData, vapp *govcd.VApp) (types.NetworkConnectionSection, error) {
	networks := d.Get("network").([]interface{})

	isStandaloneVm := vapp == nil || (vapp != nil && vapp.VApp.IsAutoNature)
	networkConnectionSection := types.NetworkConnectionSection{}

	// sets existing primary network connection index. Further changes index only if change is found
	for index, singleNetwork := range networks {
		nic := singleNetwork.(map[string]interface{})
		isPrimary := nic["is_primary"].(bool)
		if isPrimary {
			networkConnectionSection.PrimaryNetworkConnectionIndex = index
		}
	}

	for index, singleNetwork := range networks {

		nic := singleNetwork.(map[string]interface{})
		netConn := &types.NetworkConnection{}

		networkName := nic["name"].(string)
		if networkName == "" {
			networkName = "none"
		}
		ipAllocationMode := nic["ip_allocation_mode"].(string)
		ip := nic["ip"].(string)
		macAddress, macIsSet := nic["mac"].(string)

		isPrimary := nic["is_primary"].(bool)
		nicHasPrimaryChange := d.HasChange("network." + strconv.Itoa(index) + ".is_primary")
		if nicHasPrimaryChange && isPrimary {
			networkConnectionSection.PrimaryNetworkConnectionIndex = index
		}

		networkType := nic["type"].(string)
		if networkType == "org" && !isStandaloneVm {
			isVappOrgNetwork, err := isItVappOrgNetwork(networkName, *vapp)
			if err != nil {
				return types.NetworkConnectionSection{}, err
			}
			if !isVappOrgNetwork {
				return types.NetworkConnectionSection{}, fmt.Errorf("vApp Org network : %s is not found", networkName)
			}
		}
		if networkType == "vapp" && !isStandaloneVm {
			isVappNetwork, err := isItVappNetwork(networkName, *vapp)
			if err != nil {
				return types.NetworkConnectionSection{}, fmt.Errorf("unable to find vApp network %s: %s", networkName, err)
			}
			if !isVappNetwork {
				return types.NetworkConnectionSection{}, fmt.Errorf("vApp network : %s is not found", networkName)
			}
		}

		netConn.IsConnected = nic["connected"].(bool)
		netConn.IPAddressAllocationMode = ipAllocationMode
		netConn.NetworkConnectionIndex = index
		netConn.Network = networkName
		if macIsSet {
			netConn.MACAddress = macAddress
		}

		if ipAllocationMode == types.IPAllocationModeNone {
			netConn.Network = types.NoneNetwork
		}

		if net.ParseIP(ip) != nil {
			netConn.IPAddress = ip
		}

		adapterType, isSetAdapterType := nic["adapter_type"]
		if isSetAdapterType {
			netConn.NetworkAdapterType = adapterType.(string)
		}

		networkConnectionSection.NetworkConnection = append(networkConnectionSection.NetworkConnection, netConn)
	}
	return networkConnectionSection, nil
}

// isItVappOrgNetwork checks if it is an vApp Org network (not vApp Network)
func isItVappOrgNetwork(vAppNetworkName string, vapp govcd.VApp) (bool, error) {
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error getting vApp networks: %s", err)
	}

	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vAppNetworkName &&
			!govcd.IsVappNetwork(networkConfig.Configuration) {
			log.Printf("[TRACE] vApp Org network found: %s", vAppNetworkName)
			return true, nil
		}
	}

	return false, fmt.Errorf("configured vApp Org network isn't found: %s", vAppNetworkName)
}

// getVmNicIndexesWithDhcpEnabled loops over VMs NICs and returns list of indexes for the ones using
// DHCP
func getVmNicIndexesWithDhcpEnabled(networkConnectionSection *types.NetworkConnectionSection) []int {

	var nicIndexes []int

	// Sort NIC cards by their virtual slot numbers as the API returns them in random order
	sort.SliceStable(networkConnectionSection.NetworkConnection, func(i, j int) bool {
		return networkConnectionSection.NetworkConnection[i].NetworkConnectionIndex <
			networkConnectionSection.NetworkConnection[j].NetworkConnectionIndex
	})

	for nicIndex, singleNic := range networkConnectionSection.NetworkConnection {

		// validate if the NIC is suitable for DHCP waiting (has DHCP interface)
		if singleNic.IPAddressAllocationMode != types.IPAllocationModeDHCP {
			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] NIC '%d' is not using DHCP in 'ip_allocation_mode'. Skipping IP wait", nicIndex)
			continue
		}
		log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] NIC '%d' is using DHCP in 'ip_allocation_mode'.", nicIndex)
		nicIndexes = append(nicIndexes, singleNic.NetworkConnectionIndex)

	}
	return nicIndexes
}

// getVmByName returns a VM by the given name if found unequivocally
// If there are more than one instance by the wanted name, it also returns a list of
// matching VMs with sample information (ID, guest OS, network name, IP address)
func getVmByName(client *VCDClient, vdc *govcd.Vdc, name string) (*govcd.VM, string, error) {

	vmList, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
	if err != nil {
		return nil, "", err
	}

	var foundList []*types.QueryResultVMRecordType
	for _, vm := range vmList {
		if vm.Name == name {
			foundList = append(foundList, vm)
		}
	}

	if len(foundList) == 0 {
		return nil, "", govcd.ErrorEntityNotFound
	}
	if len(foundList) == 1 {
		vm, err := client.Client.GetVMByHref(foundList[0].HREF)
		if err != nil {
			return nil, "", err
		}
		return vm, "", nil
	}
	// More than one element found for the given name. Returning the list
	listStr := fmt.Sprintf("%-50s %-30s %s\n", "ID", "Guest OS", "Network")
	listStr += fmt.Sprintf("%-50s %-30s %s\n", strings.Repeat("-", 50), strings.Repeat("-", 30), strings.Repeat("-", 20))
	for _, vm := range foundList {
		id := extractUuid(vm.HREF)
		networkInfo := ""
		if vm.NetworkName != "" || vm.IpAddress != "" {
			networkInfo = fmt.Sprintf("(%s - %s)", vm.NetworkName, vm.IpAddress)
		}
		listStr += fmt.Sprintf("urn:vcloud:vm:%s %-30s %s\n", id, vm.GuestOS, networkInfo)
	}
	return nil, listStr, fmt.Errorf("more than one VM found with name %s", name)
}

// updateGuestCustomizationSetting is responsible for setting all the data related to VM customization
func updateGuestCustomizationSetting(d *schema.ResourceData, vm *govcd.VM) error {
	// Retrieve existing customization section to only customize what was throughout this function
	customizationSection, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return fmt.Errorf("error getting existing customization section before changing: %s", err)
	}

	// for back compatibility we allow to set computer name from `name` if computer_name isn't provided
	var computerName string
	if cName, ok := d.GetOk("computer_name"); ok {
		computerName = cName.(string)
	} else {
		computerName = d.Get("name").(string)
	}

	if _, isSetComputerName := d.GetOk("computer_name"); isSetComputerName {
		customizationSection.ComputerName = computerName
	}

	// Process parameters from 'customization' block
	updateCustomizationSection(d.Get("customization"), d, customizationSection)

	// Apply any of the settings we have set
	if _, err = vm.SetGuestCustomizationSection(customizationSection); err != nil {
		return fmt.Errorf("error applying guest customization details: %s", err)
	}

	return nil
}

func updateCustomizationSection(customizationInterface interface{}, d *schema.ResourceData, customizationSection *types.GuestCustomizationSection) {
	customizationSlice := customizationInterface.([]interface{})
	if len(customizationSlice) == 1 {
		cust := customizationSlice[0]
		if cust != nil {

			if enabled, isSetEnabled := d.GetOkExists("customization.0.enabled"); isSetEnabled {
				customizationSection.Enabled = takeBoolPointer(enabled.(bool))
			}
			if initScript, isSetInitScript := d.GetOkExists("customization.0.initscript"); isSetInitScript {
				customizationSection.CustomizationScript = initScript.(string)
			}

			if changeSid, isSetChangeSid := d.GetOkExists("customization.0.change_sid"); isSetChangeSid {
				customizationSection.ChangeSid = takeBoolPointer(changeSid.(bool))
			}

			if allowLocalAdminPasswd, isSetAllowLocalAdminPasswd := d.GetOkExists("customization.0.allow_local_admin_password"); isSetAllowLocalAdminPasswd {
				customizationSection.AdminPasswordEnabled = takeBoolPointer(allowLocalAdminPasswd.(bool))

			}

			if mustChangeOnFirstLogin, isSetMustChangeOnFirstLogin := d.GetOkExists("customization.0.must_change_password_on_first_login"); isSetMustChangeOnFirstLogin {
				customizationSection.ResetPasswordRequired = takeBoolPointer(mustChangeOnFirstLogin.(bool))
			}

			if autoGeneratePasswd, isSetAutoGeneratePasswd := d.GetOkExists("customization.0.auto_generate_password"); isSetAutoGeneratePasswd {
				customizationSection.AdminPasswordAuto = takeBoolPointer(autoGeneratePasswd.(bool))
			}

			if adminPasswd, isSetAdminPasswd := d.GetOkExists("customization.0.admin_password"); isSetAdminPasswd {
				customizationSection.AdminPassword = adminPasswd.(string)
				// customizationSection.AdminPasswordEnabled = takeBoolPointer(true)
			}

			if nrTimesForLogin, isSetNrTimesForLogin := d.GetOkExists("customization.0.number_of_auto_logons"); isSetNrTimesForLogin {
				// The AdminAutoLogonEnabled is "hidden" from direct user input to behave exactly like UI does. UI sets
				// the value of this field behind the scenes based on number_of_auto_logons count.
				// AdminAutoLogonEnabled=false if number_of_auto_logons == 0
				// AdminAutoLogonEnabled=true if number_of_auto_logons > 0
				isMoreThanZero := nrTimesForLogin.(int) > 0
				customizationSection.AdminAutoLogonEnabled = takeBoolPointer(isMoreThanZero)

				customizationSection.AdminAutoLogonCount = nrTimesForLogin.(int)
			}

			if joinDomain, isSetJoinDomain := d.GetOkExists("customization.0.join_domain"); isSetJoinDomain {
				customizationSection.JoinDomainEnabled = takeBoolPointer(joinDomain.(bool))
			}

			if joinOrgDomain, isSetJoinOrgDomain := d.GetOkExists("customization.0.join_org_domain"); isSetJoinOrgDomain {
				customizationSection.UseOrgSettings = takeBoolPointer(joinOrgDomain.(bool))
			}

			if joinDomainName, isSetJoinDomainName := d.GetOkExists("customization.0.join_domain_name"); isSetJoinDomainName {
				customizationSection.DomainName = joinDomainName.(string)
			}

			if joinDomainUser, isSetJoinDomainUser := d.GetOkExists("customization.0.join_domain_user"); isSetJoinDomainUser {
				customizationSection.DomainUserName = joinDomainUser.(string)
			}

			if joinDomainPasswd, isSetJoinDomainPasswd := d.GetOkExists("customization.0.join_domain_password"); isSetJoinDomainPasswd {
				customizationSection.DomainUserPassword = joinDomainPasswd.(string)
			}

			if joinDomainOu, isSetJoinDomainOu := d.GetOkExists("customization.0.join_domain_account_ou"); isSetJoinDomainOu {
				customizationSection.MachineObjectOU = joinDomainOu.(string)
			}

		}
	}
}

// setGuestCustomizationData is responsible for persisting all guest customization details into statefile
func setGuestCustomizationData(d *schema.ResourceData, vm *govcd.VM) error {
	customizationSection, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return fmt.Errorf("unable to get guest customization section: %s", err)
	}

	dSet(d, "computer_name", customizationSection.ComputerName)

	customizationBlock := make([]interface{}, 1)
	customizationBlockAttributes := make(map[string]interface{})

	customizationBlockAttributes["enabled"] = customizationSection.Enabled
	customizationBlockAttributes["change_sid"] = customizationSection.ChangeSid
	customizationBlockAttributes["allow_local_admin_password"] = customizationSection.AdminPasswordEnabled
	customizationBlockAttributes["must_change_password_on_first_login"] = customizationSection.ResetPasswordRequired
	customizationBlockAttributes["auto_generate_password"] = customizationSection.AdminPasswordAuto
	customizationBlockAttributes["admin_password"] = customizationSection.AdminPassword
	customizationBlockAttributes["number_of_auto_logons"] = customizationSection.AdminAutoLogonCount
	customizationBlockAttributes["join_domain"] = customizationSection.JoinDomainEnabled
	customizationBlockAttributes["join_org_domain"] = customizationSection.UseOrgSettings
	customizationBlockAttributes["join_domain_name"] = customizationSection.DomainName
	customizationBlockAttributes["join_domain_user"] = customizationSection.DomainUserName
	customizationBlockAttributes["join_domain_password"] = customizationSection.DomainUserPassword
	customizationBlockAttributes["join_domain_account_ou"] = customizationSection.MachineObjectOU
	customizationBlockAttributes["initscript"] = customizationSection.CustomizationScript

	customizationBlock[0] = customizationBlockAttributes

	err = d.Set("customization", customizationBlock)
	if err != nil {
		return fmt.Errorf("")
	}

	return nil
}

func addSizingPolicy(d *schema.ResourceData, vcdClient *VCDClient, recomposeVAppParamsForEmptyVm *types.RecomposeVAppParamsForEmptyVm) error {
	vcdComputePolicyHref, err := vcdClient.Client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointVdcComputePolicies)
	if err != nil {
		return fmt.Errorf("error constructing HREF for compute policy")
	}

	if value, ok := d.GetOk("sizing_policy_id"); ok {
		recomposeVAppParamsForEmptyVm.CreateItem.ComputePolicy = &types.ComputePolicy{VmSizingPolicy: &types.Reference{HREF: vcdComputePolicyHref.String() + value.(string)}}
		sizingPolicy, err := vcdClient.Client.GetVdcComputePolicyById(value.(string))
		if err != nil {
			return fmt.Errorf("error getting sizing policy %s: %s", value.(string), err)
		}
		if _, ok = d.GetOk("cpus"); !ok && sizingPolicy.VdcComputePolicy.CPUCount == nil {
			return fmt.Errorf("`cpus` has to be defined as provided sizing policy `sizing_policy_id` cpu count isn't configured")
		}
		if sizingPolicy.VdcComputePolicy.CPUCount != nil {
			recomposeVAppParamsForEmptyVm.CreateItem.VmSpecSection.NumCpus = sizingPolicy.VdcComputePolicy.CPUCount
		}
		if _, ok = d.GetOk("cpu_cores"); !ok && sizingPolicy.VdcComputePolicy.CoresPerSocket == nil {
			return fmt.Errorf("`cpu_cores` has to be defined as provided sizing policy `sizing_policy_id` cpu cores per socket isn't configured")
		}
		if sizingPolicy.VdcComputePolicy.CoresPerSocket != nil {
			recomposeVAppParamsForEmptyVm.CreateItem.VmSpecSection.NumCoresPerSocket = sizingPolicy.VdcComputePolicy.CoresPerSocket
		}
		if _, ok = d.GetOk("memory"); !ok && sizingPolicy.VdcComputePolicy.Memory == nil {
			return fmt.Errorf("`memory` has to be defined as provided sizing policy `sizing_policy_id` memory isn't configured")
		}
		if sizingPolicy.VdcComputePolicy.Memory != nil {
			recomposeVAppParamsForEmptyVm.CreateItem.VmSpecSection.MemoryResourceMb.Configured = int64(*sizingPolicy.VdcComputePolicy.Memory)
		}
	}
	return nil
}

// handleExposeHardwareVirtualization toggles hardware virtualization according `expose_hardware_virtualization` field value.
func handleExposeHardwareVirtualization(d *schema.ResourceData, newVm *govcd.VM) error {
	// The below operation assumes VM is powered off and does not check for it because VM is being
	// powered on in the last stage of create/update cycle
	if d.Get("expose_hardware_virtualization").(bool) {
		task, err := newVm.ToggleHardwareVirtualization(true)
		if err != nil {
			return fmt.Errorf("error enabling hardware assisted virtualization: %s", err)
		}
		err = task.WaitTaskCompletion()

		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}
	}
	return nil
}

// readNetworks returns network configuration for saving into statefile
func readNetworks(d *schema.ResourceData, vm govcd.VM, vapp govcd.VApp, vdc *govcd.Vdc) ([]map[string]interface{}, error) {
	// Determine type for all networks in vApp
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return []map[string]interface{}{}, fmt.Errorf("error getting vApp networks: %s", err)
	}
	// If vApp network is "isolated" and has no ParentNetwork - it is a vApp network.
	// https://code.vmware.com/apis/72/vcloud/doc/doc/types/NetworkConfigurationType.html
	vAppNetworkTypes := make(map[string]string)
	for _, netConfig := range vAppNetworkConfig.NetworkConfig {
		switch {
		case netConfig.NetworkName == types.NoneNetwork:
			vAppNetworkTypes[netConfig.NetworkName] = types.NoneNetwork
		case govcd.IsVappNetwork(netConfig.Configuration):
			vAppNetworkTypes[netConfig.NetworkName] = "vapp"
		default:
			vAppNetworkTypes[netConfig.NetworkName] = "org"
		}
	}

	var nets []map[string]interface{}
	// Sort NIC cards by their virtual slot numbers as the API returns them in random order
	sort.SliceStable(vm.VM.NetworkConnectionSection.NetworkConnection, func(i, j int) bool {
		return vm.VM.NetworkConnectionSection.NetworkConnection[i].NetworkConnectionIndex <
			vm.VM.NetworkConnectionSection.NetworkConnection[j].NetworkConnectionIndex
	})

	for _, vmNet := range vm.VM.NetworkConnectionSection.NetworkConnection {
		singleNIC := make(map[string]interface{})
		singleNIC["ip_allocation_mode"] = vmNet.IPAddressAllocationMode
		singleNIC["ip"] = vmNet.IPAddress
		singleNIC["mac"] = vmNet.MACAddress
		singleNIC["adapter_type"] = vmNet.NetworkAdapterType
		singleNIC["connected"] = vmNet.IsConnected
		if vmNet.Network != types.NoneNetwork {
			singleNIC["name"] = vmNet.Network
		}

		singleNIC["is_primary"] = false
		if vmNet.NetworkConnectionIndex == vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex {
			singleNIC["is_primary"] = true
		}

		var ok bool
		if singleNIC["type"], ok = vAppNetworkTypes[vmNet.Network]; !ok {
			// Prior vCD 10.1 used to return a placeholder for none networks. It allowed to identify
			// NIC type for types.NoneNetwork. This was removed in 10.1 therefore when vApp network
			// type has no details - the NIC network type is types.NoneNetwork
			singleNIC["type"] = types.NoneNetwork
		}

		nets = append(nets, singleNIC)
	}

	vmStatus, err := vm.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("unablet to check if VM is powered on: %s", err)
	}

	// If at least one`network_dhcp_wait_seconds` was defined
	if maxDhcpWaitSeconds, ok := d.GetOk("network_dhcp_wait_seconds"); ok && vmStatus == "POWERED_ON" {
		maxDhcpWaitSecondsInt := maxDhcpWaitSeconds.(int)

		// lookup NIC indexes which have DHCP enabled
		dhcpNicIndexes := getVmNicIndexesWithDhcpEnabled(vm.VM.NetworkConnectionSection)
		log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] '%s' DHCP is used on NICs %v with wait time '%d seconds'",
			vm.VM.Name, dhcpNicIndexes, maxDhcpWaitSecondsInt)
		if len(dhcpNicIndexes) == 0 {
			logForScreen("vcd_vapp_vm", "INFO: Using 'network_dhcp_wait_seconds' only "+
				"makes sense if at least one NIC is using 'ip_allocation_mode=DHCP'\n")
		}

		if len(dhcpNicIndexes) > 0 { // at least one NIC uses DHCP for IP allocation mode
			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] '%s' waiting for DHCP IPs up to '%d' seconds on NICs %v",
				vm.VM.Name, maxDhcpWaitSeconds, dhcpNicIndexes)

			start := time.Now()

			// Only use DHCP lease check if it is NSX-V as NSX-T Edge Gateway does not expose it and errors on such query
			useNsxvDhcpLeaseCheck := vdc.IsNsxv()
			nicIps, timeout, err := vm.WaitForDhcpIpByNicIndexes(dhcpNicIndexes, maxDhcpWaitSecondsInt, useNsxvDhcpLeaseCheck)
			if err != nil {
				return nil, fmt.Errorf("unable to to lookup DHCP IPs for VM NICs '%v': %s", dhcpNicIndexes, err)
			}

			if timeout {
				log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM %s timed out waiting %d seconds "+
					"to report DHCP IPs. You may want to increase 'network_dhcp_wait_seconds' or ensure "+
					"your DHCP settings are correct.\n", vm.VM.Name, maxDhcpWaitSeconds)
				logForScreen("vcd_vapp_vm", fmt.Sprintf("WARNING: VM %s timed out waiting %d seconds "+
					"to report DHCP IPs. You may want to increase 'network_dhcp_wait_seconds' or ensure "+
					"your DHCP settings are correct.", vm.VM.Name, maxDhcpWaitSeconds))
			}

			log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM '%s' waiting for DHCP IPs took '%s' (of '%ds')",
				vm.VM.Name, time.Since(start), maxDhcpWaitSeconds)

			for sliceIndex, nicIndex := range dhcpNicIndexes {
				log.Printf("[DEBUG] [VM read] [DHCP IP Lookup] VM '%s' NIC %d reported IP %s",
					vm.VM.Name, nicIndex, nicIps[sliceIndex])
				nets[nicIndex]["ip"] = nicIps[sliceIndex]
			}
		}
	}

	return nets, nil
}
