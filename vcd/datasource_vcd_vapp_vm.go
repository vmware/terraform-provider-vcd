package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func datasourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdVAppVmRead,

		Schema: map[string]*schema.Schema{
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The vApp this VM belongs to",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the VM, unique within the vApp",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"computer_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Computer name assigned to this virtual machine",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VM description",
				// Currently, this field has the description of the OVA used to create the VM
			},
			"memory": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The amount of RAM (in MB) to allocate to the VM",
			},
			"cpus": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of virtual CPUs to allocate to the VM",
			},
			"cpu_cores": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of cores per socket",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata to assign to this VM",
			},
			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "VM Hyper Reference",
			},
			"storage_profile": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile used with the VM",
			},
			"network_dhcp_wait_seconds": {
				Optional:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Optional number of seconds to try and wait for DHCP IP (valid for " +
					"'network' block only)",
			},
			"network": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Network type",
						},
						"ip_allocation_mode": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "IP address allocation mode.",
						},
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of the network this VM should connect to.",
						},
						"ip": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "IP of the VM. Settings depend on `ip_allocation_mode`",
						},
						"is_primary": {
							Computed:    true,
							Type:        schema.TypeBool,
							Description: "Set to true if network interface should be primary. First network card in the list will be primary by default",
						},
						"mac": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Mac address of network interface",
						},
					},
				},
			},
			"disk": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Independent disk name",
					},
					"bus_number": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Bus number on which to place the disk controller",
					},
					"unit_number": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Unit number (slot) on the bus specified by BusNumber",
					},
					"size_in_mb": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The size of the disk in MB.",
					},
				}},
				Computed: true,
				Set:      resourceVcdVmIndependentDiskHash,
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
						Description: "The type of disk controller. Possible values: ide, parallel( LSI Logic Parallel SCSI), sas(LSI Logic SAS (SCSI)), paravirtual(Paravirtual (SCSI)), sata",
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
						Description: "Specifies the IOPS for the disk. Default - 0.",
					},
					"storage_profile": &schema.Schema{
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Storage profile to override the VM default one",
					},
				}},
			},
			"expose_hardware_virtualization": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Expose hardware-assisted CPU virtualization to guest OS.",
			},
			"guest_properties": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key/value settings for guest properties",
			},
		},
	}
}

func datasourceVcdVAppVmRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVAppVmRead(d, meta, "datasource")
}
