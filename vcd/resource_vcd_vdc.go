package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
	types "github.com/vmware/go-vcloud-director/types/v56"
)

func resourceVcdVdc() *schema.Resource {
	capacityWithUsage := schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		ForceNew: false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"units": {
					Type:     schema.TypeString,
					Required: true,
				},
				"allocated": {
					Type:     schema.TypeInt,
					Required: false,
					Optional: true,
				},
				"limit": {
					Type:     schema.TypeInt,
					Required: false,
					Optional: true,
				},
				"reserved": {
					Type:     schema.TypeInt,
					Required: false,
					Optional: true,
				},
				"used": {
					Type:     schema.TypeInt,
					Required: false,
					Optional: true,
				},
				"overhead": {
					Type:     schema.TypeInt,
					Required: false,
					Optional: true,
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourceVcdVdcCreate,
		Delete: resourceVcdVdcDelete,
		Read:   resourceVcdVdcRead,
		Update: resourceVcdVdcUpdate,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"allocation_model": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					switch v {
					case
						"AllocationVApp",
						"AllocationPool",
						"ReservationPool":
						return
					default:
						errs = append(errs, fmt.Errorf("%q must be one of {AllocationVApp, AllocationPool, ReservationPool}, got: %s", key, v))
					}
					return
				},
			},
			"compute_capacity": &schema.Schema{
				Required: true,
				ForceNew: false,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu":    &capacityWithUsage,
						"memory": &capacityWithUsage,
					},
				},
			},
			"nic_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"network_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"vm_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"is_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"storage_profile": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: false,
							Optional: true,
						},
						"units": {
							Type:     schema.TypeString,
							Required: true,
						},
						"limit": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"default": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"provider": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"resource_guaranteed_memory": &schema.Schema{
				Type:     schema.TypeFloat,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"resource_guaranteed_cpu": &schema.Schema{
				Type:     schema.TypeFloat,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"v_cpu_in_mhz": &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"is_thin_provision": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"network_pool": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"provider_vdc": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"uses_fast_provisioning": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"over_commit_allowed": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
			"vm_discovery_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},

			"delete_force": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_force=True to remove a vdc and any objects it contains, regardless of their state.",
			},
			"delete_recursive": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    false,
				Description: "When destroying use delete_recursive=True to remove the vdc and any objects it contains that are in a state that normally allows removal.",
			},
		},
	}
}

func resourceVcdVdcCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] vdc creation initiated")

	vcdClient := meta.(*VCDClient)

	// vdc creation is accessible only in administrator API part
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	params, err := getVcdVdcInput(d, vcdClient)
	if err != nil {
		return err
	}

	task, err := adminOrg.CreateVdc(params)
	if err != nil {
		log.Printf("[DEBUG] Error creating vdc: %#v", err)
		return fmt.Errorf("error creating vdc: %#v", err)
	}

	d.SetId(d.Get("name").(string))
	log.Printf("[TRACE] vdc created: %#v", task)
	return resourceVcdVdcRead(d, meta)
}

func resourceVcdVdcRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] vdc read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	vdc, err := adminOrg.GetVdcByName(d.Id())
	if err != nil || vdc == (govcd.Vdc{}) {
		log.Printf("[DEBUG] Unable to find vdc. Removing from tfstate")
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] vdc read completed: %#v", vdc.Vdc)
	return nil
}

//update function for "delete_force", "delete_recursive" no actions needed
func resourceVcdVdcUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceVcdVdcDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] vdc delete started")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrg, err)
	}

	vdc, err := adminOrg.GetVdcByName(d.Id())
	if err != nil || vdc == (govcd.Vdc{}) {
		log.Printf("[DEBUG] Unable to find vdc. Removing from tfstate")
		d.SetId("")
		return nil
	}

	_, err = vdc.Delete(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		log.Printf("[DEBUG] Error removing vdc %#v", err)
		return fmt.Errorf("error removing vdc %#v", err)
	}

	log.Printf("[TRACE] vdc delete completed: %#v", vdc.Vdc)
	return nil
}

func capacityWithUsage(d map[string]interface{}) *types.CapacityWithUsage {
	v := &types.CapacityWithUsage{
		Units: d["units"].(string),
	}

	if allocated, ok := d["allocated"]; ok {
		v.Allocated = int64(allocated.(int))
	}

	if limit, ok := d["limit"]; ok {
		v.Limit = int64(limit.(int))
	}

	if reserved, ok := d["reserved"]; ok {
		v.Reserved = int64(reserved.(int))
	}

	if used, ok := d["used"]; ok {
		v.Used = int64(used.(int))
	}

	if overhead, ok := d["overhead"]; ok {
		v.Overhead = int64(overhead.(int))
	}

	return v
}

func getVcdVdcInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcConfiguration, error) {
	computeCapacity := d.Get("compute_capacity").(*schema.Set).List()[0].(map[string]interface{})
	storageProfileMap := d.Get("storage_profile").(*schema.Set).List()[0].(map[string]interface{})

	params := &types.VdcConfiguration{
		Name:            d.Get("name").(string),
		Xmlns:           "http://www.vmware.com/vcloud/v1.5",
		AllocationModel: d.Get("allocation_model").(string),
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU:    capacityWithUsage(computeCapacity["cpu"].(*schema.Set).List()[0].(map[string]interface{})),
				Memory: capacityWithUsage(computeCapacity["memory"].(*schema.Set).List()[0].(map[string]interface{})),
			},
		},
		VdcStorageProfile: &types.VdcStorageProfile{
			Units:   storageProfileMap["units"].(string),
			Limit:   int64(storageProfileMap["limit"].(int)),
			Default: storageProfileMap["default"].(bool),
			ProviderVdcStorageProfile: &types.Reference{
				HREF: fmt.Sprintf("%s/admin/pvdcStorageProfile/%s", vcdClient.Client.VCDHREF.String(), storageProfileMap["provider"].(string)),
			},
		},
		ProviderVdcReference: &types.Reference{
			HREF: fmt.Sprintf("%s/admin/providervdc/%s", vcdClient.Client.VCDHREF.String(), d.Get("provider_vdc").(string)),
		},
	}

	if storageEnabled, ok := storageProfileMap["enabled"]; ok {
		params.VdcStorageProfile.Enabled = storageEnabled.(bool)
	}

	if description, ok := d.GetOk("description"); ok {
		params.Description = description.(string)
	}

	if nicQuota, ok := d.GetOk("nic_quota"); ok {
		params.NicQuota = nicQuota.(int)
	}

	if networkQuota, ok := d.GetOk("network_quota"); ok {
		params.NetworkQuota = networkQuota.(int)
	}

	if vmQuota, ok := d.GetOk("vm_quota"); ok {
		params.VmQuota = vmQuota.(int)
	}

	if isEnabled, ok := d.GetOk("is_enabled"); ok {
		params.IsEnabled = isEnabled.(bool)
	}

	if resourceGuaranteedMemory, ok := d.GetOk("resource_guaranteed_memory"); ok {
		params.ResourceGuaranteedMemory = resourceGuaranteedMemory.(float64)
	}

	if resourceGuaranteedCpu, ok := d.GetOk("resource_guaranteed_cpu"); ok {
		params.ResourceGuaranteedCpu = resourceGuaranteedCpu.(float64)
	}

	if vCpuInMhz, ok := d.GetOk("v_cpu_in_mhz"); ok {
		params.VCpuInMhz = vCpuInMhz.(int64)
	}

	if isThinProvision, ok := d.GetOk("is_thin_provision"); ok {
		params.IsThinProvision = isThinProvision.(bool)
	}

	if networkPool, ok := d.GetOk("network_pool"); ok {
		params.NetworkPoolReference = &types.Reference{
			HREF: fmt.Sprintf("%s/admin/extension/networkPool/%s", vcdClient.Client.VCDHREF.String(), networkPool.(string)),
		}
	}

	if usesFastProvisioning, ok := d.GetOk("uses_fast_provisioning"); ok {
		params.UsesFastProvisioning = usesFastProvisioning.(bool)
	}

	if overCommitAllowed, ok := d.GetOk("over_commit_allowed"); ok {
		params.OverCommitAllowed = overCommitAllowed.(bool)
	}

	if vmDiscoveryEnabled, ok := d.GetOk("vm_discovery_enabled"); ok {
		params.VmDiscoveryEnabled = vmDiscoveryEnabled.(bool)
	}

	return params, nil
}
