package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdIndependentDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdIndependentDiskCreate,
		Read:   resourceVcdIndependentDiskRead,
		Delete: resourceVcdIndependentDiskDelete,

		Schema: map[string]*schema.Schema{
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
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Type:        schema.TypeFloat,
				Required:    true,
				ForceNew:    true,
				Description: "size in GB",
			},
			"bus_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateBusType,
			},
			"bus_sub_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateBusSubType,
			},
		},
	}
}

var busTypes = map[string]string{
	"IDE":  "5",
	"SCSI": "6",
	"SATA": "20",
}

var busSubTypes = map[string]string{
	"ide":         "IDE",
	"buslogic":    "buslogic",
	"lsilogic":    "lsilogic",
	"lsilogicsas": "lsilogicsas",
	"virtualscsi": "VirtualSCSI",
	"ahci":        "vmware.sata.ahci",
}

func resourceVcdIndependentDiskCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	diskName := d.Get("name").(string)
	diskRecord, _ := vdc.QueryDisk(diskName)

	if diskRecord != (govcd.DiskRecord{}) {
		return fmt.Errorf("disk with such name already exist : %s", diskName)
	}

	diskCreateParams := &types.DiskCreateParams{Disk: &types.Disk{
		Name: diskName,
		Size: int(d.Get("size").(float64) * 1024 * 1024 * 1024),
	}}

	var storageReference types.Reference
	storageProfileValue := d.Get("storage_profile").(string)

	if storageProfileValue != "" {
		storageReference, err = vdc.FindStorageProfileReference(storageProfileValue)
		if err != nil {
			return fmt.Errorf("error finding storage profile %s", storageProfileValue)
		}
		diskCreateParams.Disk.StorageProfile = &types.Reference{HREF: storageReference.HREF}
	}

	busTypeValue := d.Get("bus_type").(string)
	if busTypeValue != "" {
		diskCreateParams.Disk.BusType = busTypes[strings.ToUpper(busTypeValue)]
	}

	busSubTypeValue := d.Get("bus_sub_type").(string)
	if busSubTypeValue != "" {
		diskCreateParams.Disk.BusSubType = busSubTypes[strings.ToLower(busSubTypeValue)]
	}

	task, err := vdc.CreateDisk(diskCreateParams)
	if err != nil {
		return fmt.Errorf("error creating independent disk: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting to finish creation of independent disk: %s", err)
	}

	d.SetId(diskName)

	return resourceVcdIndependentDiskRead(d, meta)
}

func resourceVcdIndependentDiskRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	_, err = vdc.QueryDisk(d.Get("name").(string))
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error finding disk or no such disk found: %#v", err)
	}

	log.Printf("[TRACE] Disk read completed.")
	return nil
}

func resourceVcdIndependentDiskDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	diskRecord, err := vdc.QueryDisk(d.Get("name").(string))
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error finding disk : %#v", err)
	}

	if diskRecord.Disk.IsAttached {
		return fmt.Errorf("can not remove disk as it is attached to vm")
	}

	disk, err := vdc.FindDiskByHREF(diskRecord.Disk.HREF)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error getting disk : %#v", err)
	}

	task, err := disk.Delete()
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error deleting disk : %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error waiting for deleting disk : %#v", err)
	}

	return nil
}

func validateBusType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)
	if "" == busTypes[strings.ToUpper(value)] {
		errors = append(errors, fmt.Errorf(
			"%q (%q) value isn't valid", k, value))
	}
	return
}

func validateBusSubType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)
	if "" == busSubTypes[strings.ToLower(value)] {
		errors = append(errors, fmt.Errorf(
			"%q (%q) value isn't valid", k, value))
	}
	return
}
