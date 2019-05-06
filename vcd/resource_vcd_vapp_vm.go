package vcd

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"net"
	"sort"
	"strconv"
)

func resourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVAppVmCreate,
		Update: resourceVcdVAppVmUpdate,
		Read:   resourceVcdVAppVmRead,
		Delete: resourceVcdVAppVmDelete,

		Schema: map[string]*schema.Schema{
			"vapp_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"template_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"catalog_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cpus": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cpu_cores": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip": &schema.Schema{
				Computed:         true,
				ConflictsWith:    []string{"networks"},
				Deprecated:       "In favor of networks",
				DiffSuppressFunc: suppressIfIPIsOneOf(),
				ForceNew:         true,
				Optional:         true,
				Type:             schema.TypeString,
			},
			"mac": {
				Computed:      true,
				ConflictsWith: []string{"networks"},
				Optional:      true,
				Type:          schema.TypeString,
			},
			"initscript": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
			"href": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"accept_all_eulas": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"power_on": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"network_href": &schema.Schema{
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "In favor of networks",
			},
			"networks": {
				ConflictsWith: []string{"ip", "network_name", "vapp_network_name"},
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_type": {
							ForceNew:     true,
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"vapp", "org", "none"}, false), // none?
							Description:  "Type of network to use 'vapp' or 'org'. 'vapp' uses vApp level network while 'org' attaches Org VDC network.",
						},
						"ip_allocation_mode": {
							ForceNew:     true,
							Optional:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"POOL", "DHCP", "MANUAL", "NONE"}, false),
						},
						"network_name": {
							ForceNew: false,
							Optional: true, // In case of ip_allocation_mode = NONE it is not required
							Type:     schema.TypeString,
						},
						"ip": {
							Computed: true,
							ForceNew: true,
							Optional: true,
							Type:     schema.TypeString,
							// Must accept empty string because of looping over vars purposes
							ValidateFunc: checkEmptyOrSingleIP(),
						},
						"is_primary": {
							Default:  false,
							ForceNew: true,
							Optional: true,
							// schema change to false would not do anything useful unless
							// other adapter is set to true therefore this change is suppressed.
							DiffSuppressFunc: falseBoolSuppress(),
							Type:             schema.TypeBool,
						},
						//  Cannot be used right now because changing adapter_type would need
						//  a bigger rework of AddVM() function in go-vcloud-director library
						//  to allow to set adapter_type while creation of NetworkConnection.
						//
						//  "adapter_type": {
						//  	Type:     schema.TypeString,
						//  	ForceNew: true,
						//  	Optional: true,
						//  },
						"mac": {
							Computed: true,
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"network_name": &schema.Schema{
				ConflictsWith: []string{"networks"},
				Deprecated:    "In favor of networks",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeString,
			},
			"vapp_network_name": &schema.Schema{
				Deprecated: "In favor of networks",
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
			},
			"disk": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"bus_number": {
						Type:     schema.TypeString,
						Required: true,
					},
					"unit_number": {
						Type:     schema.TypeString,
						Required: true,
					},
				}},
				Optional: true,
				Set:      resourceVcdVmIndependentDiskHash,
			},
			"expose_hardware_virtualization": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Expose hardware-assisted CPU virtualization to guest OS.",
			},
		},
	}
}

func checkEmptyOrSingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if net.ParseIP(v) == nil && v != "" {
			es = append(es, fmt.Errorf(
				"expected %s to be empty or contain a valid IP, got: %s", k, v))
		}
		return
	}
}

// TODO 3.0 remove once `ip` and `network_name` attributes are removed
func suppressIfIPIsOneOf() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		switch {
		case new == "dhcp" && (old == "na" || net.ParseIP(old) != nil):
			return true
		case new == "allocated" && net.ParseIP(old) != nil:
			return true
		case new == "" && net.ParseIP(old) != nil:
			return true
		default:
			return false
		}
	}
}

// falseBoolSuppress suppresses change if value is set to false
func falseBoolSuppress() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		_, isTrue := d.GetOk(k)
		return !isTrue
	}
}

func resourceVcdVAppVmCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	catalog, err := org.FindCatalog(d.Get("catalog_name").(string))
	if err != nil || catalog == (govcd.Catalog{}) {
		return fmt.Errorf("error finding catalog: %s", d.Get("catalog_name").(string))
	}

	catalogItem, err := catalog.FindCatalogItem(d.Get("template_name").(string))
	if err != nil || catalogItem == (govcd.CatalogItem{}) {
		return fmt.Errorf("error finding catalog item: %#v", err)
	}

	vappTemplate, err := catalogItem.GetVAppTemplate()
	if err != nil {
		return fmt.Errorf("error finding VAppTemplate: %#v", err)
	}

	acceptEulas := d.Get("accept_all_eulas").(bool)

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))
	if err != nil {
		return fmt.Errorf("error finding vApp: %#v", err)
	}

	// Networking part

	// Step 1 - gather vars
	legacyNetworkName := d.Get("network_name").(string)
	legacyVappNetworkName := d.Get("vapp_network_name").(string)

	allNetworks := d.Get("networks").([]interface{})

	var vAppNets []*types.OrgVDCNetwork
	var vdcNets []*types.OrgVDCNetwork

	// Step 2 - process legacy "network_name" and "vapp_network_name"
	// TODO 3.0 remove when "network_name" and "vapp_network_name" is deprecated
	if legacyNetworkName != "" {
		legacyNetwork, err := addVdcNetwork(d.Get("network_name").(string), vdc, vapp, vcdClient)
		if err != nil {
			return err
		}
		vdcNets = append(vdcNets, legacyNetwork)
	}
	if legacyVappNetworkName != "" {
		isVappNetwork, err := isItVappNetwork(legacyVappNetworkName, vapp)
		if err != nil {
			return err
		}
		if !isVappNetwork {
			return fmt.Errorf("vapp_network_name: %s is not found", legacyVappNetworkName)
		}
	}
	// TODO 3.0 EO remove when "network_name" and "vapp_network_name" is deprecated

	// Step 3 - process new networks of all types
	for _, network := range allNetworks {
		net := network.(map[string]interface{})
		netType := net["network_type"].(string)
		netName := net["network_name"].(string)

		switch netType {
		case "vapp":
			isVappNetwork, err := isItVappNetwork(netName, vapp)
			if err != nil {
				return fmt.Errorf("unable to find vApp network %s: %s", netName, err)
			}
			if !isVappNetwork {
				return fmt.Errorf("vApp network : %s is not found", netName)
			}
			vAppNets = append(vAppNets, &types.OrgVDCNetwork{Name: netName})
		case "org":
			vdcNet, err := vdc.FindVDCNetwork(netName)
			if err != nil {
				return fmt.Errorf("error finding org vdc network: %#v", err)
			}
			vdcNets = append(vdcNets, vdcNet.OrgVDCNetwork)
		case "none":
			// Special reference type 'none' network  does not exist anywhere, but allows to add unattached NIC.
			// Store empty object with hardcoded type.
			vdcNets = append(vdcNets, &types.OrgVDCNetwork{Name: types.NoneNetwork})
		}

	}

	// Step 4 - ensure all specified vdc networks are attached to vApp
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		fmt.Errorf("unable to retrieve vApp network config: %s", err)
	}
	vAppNetworkNames := vAppNetworkConfig.NetworkNames()
	// Check if a vdc network is attached to vApp. If not - attach it. Skip types.NoneNetwork.
	for _, vdcNet := range vdcNets {
		if !stringInSlice(vdcNet.Name, vAppNetworkNames) && vdcNet.Name != types.NoneNetwork {
			// TODO optimize addVdcNetwork as it does additional unnecessary calls
			_, err = addVdcNetwork(vdcNet.Name, vdc, vapp, vcdClient)
			if err != nil {
				return fmt.Errorf("unable to assign vdc network %s to vApp %s", vdcNet.Name, vAppNetworkNames)
			}
		}
	}

	// After attaching org vdc networks it is safe to add vApp network names to the list for being added
	vdcNets = append(vdcNets, vAppNets...)

	// Adds all network cards with default settings
	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Creating VM: %s", d.Get("name").(string))
		// TODO refactor vapp.AddVM to create NICs with proper configuration
		// instead of default NIC parameters. Currently it depends on vm.ChangeNetworkConfig(n) in the Update()
		// to set proper network allocations
		task, err := vapp.AddVM(vdcNets, legacyVappNetworkName, vappTemplate, d.Get("name").(string), acceptEulas)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error adding VM: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})

	if err != nil {
		return fmt.Errorf(errorCompletingTask, err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("error getting VM1 : %#v", err)
	}

	// TODO 3.0 remove when "network_name" and "vapp_network_name" is deprecated
	if legacyNetworkName != "" || legacyVappNetworkName != "" {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			var networksChanges []map[string]interface{}
			if legacyVappNetworkName != "" {
				networksChanges = append(networksChanges, map[string]interface{}{
					"ip":           d.Get("ip").(string),
					"network_name": legacyVappNetworkName,
				})
			}
			if legacyNetworkName != "" {
				networksChanges = append(networksChanges, map[string]interface{}{
					"ip":           d.Get("ip").(string),
					"network_name": legacyNetworkName,
				})
			}

			task, err := vm.ChangeNetworkConfig(networksChanges)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error with Networking change: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
	}
	// TODO 3.0 EO remove when "network_name" and "vapp_network_name" is deprecated

	if err != nil {
		return fmt.Errorf("error changing network: %#v", err)
	}
	// The below operation assumes VM is powered off and does not check for it because VM is being
	// powered on in the last stage of create/update cycle
	if d.Get("expose_hardware_virtualization").(bool) {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vm.ToggleHardwareVirtualization(true)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error enabling hardware assisted virtualization: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}
	}

	if initScript, ok := d.GetOk("initscript"); ok {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vm.RunCustomizationScript(d.Get("name").(string), initScript.(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error with setting init script: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}
	}

	d.SetId(d.Get("name").(string))

	err = resourceVcdVAppVmUpdate(d, meta)
	if err != nil {
		errAttachedDisk := updateStateOfAttachedDisks(d, vm, vdc)
		if errAttachedDisk != nil {
			d.Set("disk", nil)
			return fmt.Errorf("error reading attached disks : %#v and internal error : %#v", errAttachedDisk, err)
		}
		return err
	}
	return nil
}

// Adds existing org VDC network to VM network configuration
// Returns configured OrgVDCNetwork for Vm, networkName, error if any occur
func addVdcNetwork(networkNameToAdd string, vdc govcd.Vdc, vapp govcd.VApp, vcdClient *VCDClient) (*types.OrgVDCNetwork, error) {
	if networkNameToAdd == "" {
		return &types.OrgVDCNetwork{}, fmt.Errorf("'network_name' must be valid when adding VM to raw vApp")
	}

	net, err := vdc.FindVDCNetwork(networkNameToAdd)
	if err != nil {
		return &types.OrgVDCNetwork{}, fmt.Errorf("network %s wasn't found as VDC network", networkNameToAdd)
	}
	vdcNetwork := net.OrgVDCNetwork

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return &types.OrgVDCNetwork{}, fmt.Errorf("could not get network config: %s", err)
	}

	isAlreadyVappNetwork := false
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkNameToAdd {
			log.Printf("[TRACE] VDC network found as vApp network: %s", networkNameToAdd)
			isAlreadyVappNetwork = true
		}
	}

	if !isAlreadyVappNetwork {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{vdcNetwork})
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error assigning network to vApp: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})

		if err != nil {
			return &types.OrgVDCNetwork{}, fmt.Errorf("error assigning network to vApp:: %#v", err)
		}
	}

	return vdcNetwork, nil
}

// Checks if vapp network available for using
func isItVappNetwork(vAppNetworkName string, vapp govcd.VApp) (bool, error) {
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return false, fmt.Errorf("error getting vApp networks: %#v", err)
	}

	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vAppNetworkName {
			log.Printf("[TRACE] vApp network found: %s", vAppNetworkName)
			return true, nil
		}
	}

	return false, fmt.Errorf("configured vApp network isn't found: %#v", err)
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
				return nil, fmt.Errorf("value `%s` bus_number is not number. err: %#v", busNumber, err)
			}
			addParams.busNumber = &convertedBusNumber
		}

		unitNumber := original["unit_number"].(string)
		if unitNumber != "" {
			convertedUnitNumber, err := strconv.Atoi(unitNumber)
			if err != nil {
				return nil, fmt.Errorf("value `%s` unit_number is not number. err: %#v", unitNumber, err)
			}
			addParams.unitNumber = &convertedUnitNumber
		}

		diskParamsArray = append(diskParamsArray, addParams)
	}
	return diskParamsArray, nil
}

func getVmIndependentDisks(vm govcd.VM) []string {

	var disks []string
	for _, item := range vm.VM.VirtualHardwareSection.Item {
		// disk resource type is 17
		if item.ResourceType == 17 && "" != item.HostResource[0].Disk {
			disks = append(disks, item.HostResource[0].Disk)
		}
	}
	return disks
}

func resourceVcdVAppVmUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vApp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("error getting VM2: %#v", err)
	}

	status, err := vm.GetStatus()
	if err != nil {
		return fmt.Errorf("error getting VM status: %#v", err)
	}

	// VM does not have to be in POWERED_OFF state for metadata operations
	if d.HasChange("metadata") {
		oldRaw, newRaw := d.GetChange("metadata")
		oldMetadata := oldRaw.(map[string]interface{})
		newMetadata := newRaw.(map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		for k := range oldMetadata {
			if _, ok := newMetadata[k]; !ok {
				toBeRemovedMetadata = append(toBeRemovedMetadata, k)
			}
		}
		for _, k := range toBeRemovedMetadata {
			task, err := vm.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}
		for k, v := range newMetadata {
			task, err := vm.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}
	}

	// When there is more then one VM in a vApp Terraform will try to parallelise their creation.
	// However, vApp throws errors when simultaneous requests are executed.
	// To avoid them, below block is using retryCall in multiple places as a workaround,
	// so that the VMs are created regardless of parallelisation.
	if d.HasChange("memory") || d.HasChange("cpus") || d.HasChange("cpu_cores") || d.HasChange("networks") ||
		d.HasChange("disk") || d.HasChange("power_on") || d.HasChange("expose_hardware_virtualization") {
		if status != "POWERED_OFF" {
			task, err := vm.PowerOff()
			if err != nil {
				return fmt.Errorf("error Powering Off: %#v", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

		if d.HasChange("memory") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vm.ChangeMemorySize(d.Get("memory").(int))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing memory size: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return err
			}
		}

		if d.HasChange("cpus") || d.HasChange("cpu_cores") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				var task govcd.Task
				var err error
				if d.Get("cpu_cores") != nil {
					coreCounts := d.Get("cpu_cores").(int)
					task, err = vm.ChangeCPUCountWithCore(d.Get("cpus").(int), &coreCounts)
				} else {
					task, err = vm.ChangeCPUCount(d.Get("cpus").(int))
				}
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing cpu count: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

		if d.HasChange("expose_hardware_virtualization") {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vm.ToggleHardwareVirtualization(d.Get("expose_hardware_virtualization").(bool))
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing hardware assisted virtualization: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return err
			}
		}

		if d.HasChange("networks") {
			n := []map[string]interface{}{}

			nets := d.Get("networks").([]interface{})
			for _, network := range nets {
				n = append(n, network.(map[string]interface{}))
			}
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vm.ChangeNetworkConfig(n)
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error changing network: %#v", err))
				}
				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

		// detaching independent disks - only possible when VM power off
		if d.HasChange("disk") {
			err = attachDetachDisks(d, vm, vdc)
			if err != nil {
				errAttachedDisk := updateStateOfAttachedDisks(d, vm, vdc)
				if errAttachedDisk != nil {
					d.Set("disk", nil)
					return fmt.Errorf("error reading attached disks : %#v and internal error : %#v", errAttachedDisk, err)
				}
				return fmt.Errorf("error attaching-detaching  disks when updating resource : %#v", err)
			}
		}

		if d.Get("power_on").(bool) {
			err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
				task, err := vm.PowerOn()
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error Powering Up: %#v", err))
				}

				return resource.RetryableError(task.WaitTaskCompletion())
			})
			if err != nil {
				return fmt.Errorf(errorCompletingTask, err)
			}
		}

	}

	return resourceVcdVAppVmRead(d, meta)
}

// updates attached disks to latest state. Removed not needed and add new ones
func attachDetachDisks(d *schema.ResourceData, vm govcd.VM, vdc govcd.Vdc) error {
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
			return fmt.Errorf("did not find disk `%s`: %#v", diskData.name, err)
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
			return fmt.Errorf("error detaching disk `%s` to vm %#v", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete detaching disk `%s` to vm %#v", diskData.name, err)
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
			return fmt.Errorf("did not find disk `%s`: %#v", diskData.name, err)
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
			return fmt.Errorf("error attaching disk `%s` to vm %#v", diskData.name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting for task to complete attaching disk `%s` to vm %#v", diskData.name, err)
		}
	}
	return nil
}

func resourceVcdVAppVmRead(d *schema.ResourceData, meta interface{}) error {
	log.Println("[TRACE] resourceVcdVAppVmRead")
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vApp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		d.SetId("")
		return fmt.Errorf("error getting VM : %#v", err)
	}

	d.Set("name", vm.VM.Name)

	switch {
	// network_name is not set. networks is set in config
	// TODO remove this case block when we cleanup deprecated 'ip' and 'network_name' attributes
	case d.Get("network_name").(string) != "":
		ip := vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddress
		// If allocation mode is DHCP and we're not getting the IP - we set this to na (not available)
		if vm.VM.NetworkConnectionSection.NetworkConnection[0].IPAddressAllocationMode == types.IPAllocationModeDHCP && ip == "" {
			ip = "na"
		}

		d.Set("ip", ip)
		d.Set("mac", vm.VM.NetworkConnectionSection.NetworkConnection[0].MACAddress)
		// TODO EO remove this case block when we cleanup deprecated 'ip' and 'network_name' attributes
	// We are using networks block and rebuilding statefile
	case len(d.Get("networks").([]interface{})) > 0:
		// Determine network_type for all networks in vApp
		vAppNetworkConfig, err := vapp.GetNetworkConfig()
		if err != nil {
			return fmt.Errorf("error getting vApp networks: %#v", err)
		}
		// If vApp network is "isolated" and has no ParentNetwork - it is a vApp network.
		// https://code.vmware.com/apis/72/vcloud/doc/doc/types/NetworkConfigurationType.html
		vAppNetworkTypes := make(map[string]string)
		for _, netConfig := range vAppNetworkConfig.NetworkConfig {
			if netConfig.Configuration.ParentNetwork == nil && netConfig.Configuration.FenceMode == "isolated" {
				vAppNetworkTypes[netConfig.NetworkName] = "vapp"
			} else {
				vAppNetworkTypes[netConfig.NetworkName] = "org"
			}
		}

		var nets []map[string]interface{}
		// Sort NIC cards by their virtual slot numbers as the API returns then in random order
		sort.SliceStable(vm.VM.NetworkConnectionSection.NetworkConnection, func(i, j int) bool {
			return vm.VM.NetworkConnectionSection.NetworkConnection[i].NetworkConnectionIndex <
				vm.VM.NetworkConnectionSection.NetworkConnection[j].NetworkConnectionIndex
		})

		for i, vmNet := range vm.VM.NetworkConnectionSection.NetworkConnection {
			singleNIC := make(map[string]interface{})
			singleNIC["ip_allocation_mode"] = vmNet.IPAddressAllocationMode
			singleNIC["ip"] = vmNet.IPAddress
			singleNIC["mac"] = vmNet.MACAddress
			if vmNet.Network != types.NoneNetwork {
				singleNIC["network_name"] = vmNet.Network
			}

			singleNIC["is_primary"] = false
			if i == vm.VM.NetworkConnectionSection.PrimaryNetworkConnectionIndex {
				singleNIC["is_primary"] = true
			}

			if vmNet.Network == types.NoneNetwork {
				singleNIC["network_type"] = types.NoneNetwork
			} else {
				var ok bool
				if singleNIC["network_type"], ok = vAppNetworkTypes[vmNet.Network]; !ok {
					return fmt.Errorf("unable to determine vApp network type for: %s", vmNet.Network)
				}
			}

			nets = append(nets, singleNIC)
		}

		d.Set("networks", nets)
	}

	d.Set("href", vm.VM.HREF)
	d.Set("expose_hardware_virtualization", vm.VM.NestedHypervisorEnabled)

	err = updateStateOfAttachedDisks(d, vm, vdc)
	if err != nil {
		d.Set("disk", nil)
		return fmt.Errorf("error reading attached disks : %#v", err)
	}

	return nil
}

func updateStateOfAttachedDisks(d *schema.ResourceData, vm govcd.VM, vdc govcd.Vdc) error {
	// Check VM independent disks state
	diskProperties, err := expandDisksProperties(d.Get("disk"))
	if err != nil {
		return err
	}

	existingDisks := getVmIndependentDisks(vm)
	transformed := schema.NewSet(resourceVcdVmIndependentDiskHash, []interface{}{})

	for _, existingDiskHref := range existingDisks {
		disk, err := vdc.FindDiskByHREF(existingDiskHref)
		if err != nil {
			return fmt.Errorf("did not find disk `%s`: %#v", existingDiskHref, err)
		}

		// where isn't way to find bus_number and unit_number, so need copied from old values to not lose them
		var oldValues diskParams
		for _, oldDiskData := range diskProperties {
			if oldDiskData.name == disk.Disk.Name {
				oldValues = diskParams{name: oldDiskData.name, busNumber: oldDiskData.busNumber, unitNumber: oldDiskData.unitNumber}
			}
		}

		newValues := map[string]interface{}{
			"name": disk.Disk.Name,
		}

		if (diskParams{}) != oldValues {
			if nil != oldValues.busNumber {
				newValues["bus_number"] = strconv.Itoa(*oldValues.busNumber)
			}
			if nil != oldValues.unitNumber {
				newValues["unit_number"] = strconv.Itoa(*oldValues.unitNumber)
			}
		}

		transformed.Add(newValues)
	}

	d.Set("disk", transformed)
	return nil
}

func resourceVcdVAppVmDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.FindVAppByName(d.Get("vapp_name").(string))

	if err != nil {
		return fmt.Errorf("error finding vApp: %s", err)
	}

	vm, err := vdc.FindVMByName(vapp, d.Get("name").(string))

	if err != nil {
		return fmt.Errorf("error getting VM4 : %#v", err)
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("error getting vApp status: %#v", err)
	}

	log.Printf("[TRACE] vApp Status:: %s", status)
	if status != "POWERED_OFF" {
		log.Printf("[TRACE] Undeploying vApp: %s", vapp.VApp.Name)
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := vapp.Undeploy()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error Undeploying: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("error Undeploying vApp: %#v", err)
		}
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		log.Printf("[TRACE] Removing VM: %s", vm.VM.Name)
		err := vapp.RemoveVM(vm)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error deleting: %#v", err))
		}

		return nil
	})

	if status != "POWERED_OFF" {
		log.Printf("[TRACE] Redeploying vApp: %s", vapp.VApp.Name)
		task, err := vapp.Deploy()
		if err != nil {
			return fmt.Errorf("error Deploying vApp: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}

		log.Printf("[TRACE] Powering on vApp: %s", vapp.VApp.Name)
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err = vapp.PowerOn()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error Powering Up vApp: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf("error Powering Up vApp: %#v", err)
		}
	}

	return err
}

func resourceVcdVmIndependentDiskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-",
		m["name"].(string)))
	if nil != m["bus_number"] {
		buf.WriteString(fmt.Sprintf("%s-",
			m["bus_number"].(string)))
	}
	if nil != m["unit_number"] {
		buf.WriteString(fmt.Sprintf("%s-",
			m["unit_number"].(string)))
	}
	return hashcode.String(buf.String())
}

func stringInSlice(search string, list []string) bool {
	for _, str := range list {
		if str == search {
			return true
		}
	}
	return false
}
