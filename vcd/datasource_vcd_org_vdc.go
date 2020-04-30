package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdOrgVdc() *schema.Resource {
	capacityWithUsage := schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allocated": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Capacity that is committed to be available. Value in MB or MHz. Used with AllocationPool (Allocation pool) and ReservationPool (Reservation pool).",
				},
				"limit": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Capacity limit relative to the value specified for Allocation. It must not be less than that value. If it is greater than that value, it implies over provisioning. A value of 0 specifies unlimited units. Value in MB or MHz. Used with AllocationVApp (Pay as you go).",
				},
				"reserved": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"used": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}

	return &schema.Resource{
		Read: datasourceVcdOrgVdcRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization to create the VDC in",
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"allocation_model": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The allocation model used by this VDC; must be one of {AllocationVApp, AllocationPool, ReservationPool}",
			},
			"compute_capacity": &schema.Schema{
				Computed: true,
				MinItems: 1,
				MaxItems: 1,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu":    &capacityWithUsage,
						"memory": &capacityWithUsage,
					},
				},
				Description: "The compute capacity allocated to this VDC.",
			},
			"nic_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of virtual NICs allowed in this VDC. Defaults to 0, which specifies an unlimited number.",
			},
			"network_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of network objects that can be deployed in this VDC. Defaults to 0, which means no networks can be deployed.",
			},
			"vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum number of VMs that can be created in this VDC. Includes deployed and undeployed VMs in vApps and vApp templates. Defaults to 0, which specifies an unlimited number.",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this VDC is enabled for use by the organization VDCs. Default is true.",
			},
			"storage_profile": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of Provider VDC storage profile.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if this storage profile is enabled for use in the VDC.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of MB allocated for this storage profile. A value of 0 specifies unlimited MB.",
						},
						"default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified.",
						},
					},
				},
				Description: "Storage profiles supported by this VDC.",
			},
			"memory_guaranteed": &schema.Schema{
				Type:     schema.TypeFloat,
				Computed: true,
				Description: "Percentage of allocated memory resources guaranteed to vApps deployed in this VDC. " +
					"For example, if this value is 0.75, then 75% of allocated resources are guaranteed. " +
					"Required when AllocationModel is AllocationVApp or AllocationPool. When Allocation model is AllocationPool minimum value is 0.2. If the element is empty, vCD sets a value.",
			},
			"cpu_guaranteed": &schema.Schema{
				Type:     schema.TypeFloat,
				Computed: true,
				Description: "Percentage of allocated CPU resources guaranteed to vApps deployed in this VDC. " +
					"For example, if this value is 0.75, then 75% of allocated resources are guaranteed. " +
					"Required when AllocationModel is AllocationVApp or AllocationPool. If the element is empty, vCD sets a value",
			},
			"cpu_speed": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Specifies the clock frequency, in Megahertz, for any virtual CPU that is allocated to a VM. A VM with 2 vCPUs will consume twice as much of this value. Ignored for ReservationPool. Required when AllocationModel is AllocationVApp or AllocationPool, and may not be less than 256 MHz. Defaults to 1000 MHz if the element is empty or missing.",
			},
			"enable_thin_provisioning": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean to request thin provisioning. Request will be honored only if the underlying datastore supports it. Thin provisioning saves storage space by committing it on demand. This allows over-allocation of storage.",
			},
			"network_pool_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of a network pool in the Provider VDC. Required if this VDC will contain routed or isolated networks.",
			},
			"provider_vdc_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A reference to the Provider VDC from which this organization VDC is provisioned.",
			},
			"enable_fast_provisioning": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Request for fast provisioning. Request will be honored only if the underlying datas tore supports it. Fast provisioning can reduce the time it takes to create virtual machines by using vSphere linked clones. If you disable fast provisioning, all provisioning operations will result in full clones.",
			},
			//  Always null in the response to a GET request. On update, set to false to disallow the update if the AllocationModel is AllocationPool or ReservationPool
			//  and the ComputeCapacity you specified is greater than what the backing Provider VDC can supply. Defaults to true if empty or missing.
			"allow_over_commit": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Set to false to disallow creation of the VDC if the AllocationModel is AllocationPool or ReservationPool and the ComputeCapacity you specified is greater than what the backing Provider VDC can supply. Default is true.",
			},
			"enable_vm_discovery": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if discovery of vCenter VMs is enabled for resource pools backing this VDC. If left unspecified, the actual behaviour depends on enablement at the organization level and at the system level.",
			},
			"elasticity": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Flex VDC is elastic.",
			},
			"include_vm_memory_overhead": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Flex VDC includes memory overhead into its accounting for admission control.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for Org VDC metadata",
			},
		},
	}
}

func datasourceVcdOrgVdcRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	adminVdc, err := adminOrg.GetAdminVDCByName(d.Get("name").(string), false)
	if err != nil {
		log.Printf("[DEBUG] Unable to find VDC")
		return fmt.Errorf("unable to find VDC %s", err)
	}

	d.SetId(adminVdc.AdminVdc.ID)

	return setOrgVdcData(d, vcdClient, adminOrg, adminVdc)
}
