package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdIndependentDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdIndependentDiskCreate,
		Read:   resourceVcdIndependentDiskRead,
		Delete: resourceVcdIndependentDiskDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdIndependentDiskImport,
		},
		Schema: map[string]*schema.Schema{
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
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Type:          schema.TypeFloat,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"size_in_bytes"},
				Deprecated:    "In favor of size_in_bytes",
				Description:   "size in MB",
			},
			"size_in_bytes": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"size"},
				Description:   "size in bytes",
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
			"iops": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IOPS request for the created disk",
			},
			"owner_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owner name of the disk",
			},
			"datastore_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Datastore name",
			},
			"is_attached": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the disk is already attached",
			},
		},
	}
}

var busTypes = map[string]string{
	"IDE":  "5",
	"SCSI": "6",
	"SATA": "20",
}
var busTypesFromValues = map[string]string{
	"5":  "IDE",
	"6":  "SCSI",
	"20": "SATA",
}

var busSubTypes = map[string]string{
	"ide":         "IDE",
	"buslogic":    "buslogic",
	"lsilogic":    "lsilogic",
	"lsilogicsas": "lsilogicsas",
	"virtualscsi": "VirtualSCSI",
	"ahci":        "vmware.sata.ahci",
}

var busSubTypesFromValues = map[string]string{
	"IDE":              "ide",
	"buslogic":         "buslogic",
	"lsilogic":         "lsilogic",
	"lsilogicsas":      "lsilogicsas",
	"VirtualSCSI":      "virtualscsi",
	"vmware.sata.ahci": "ahci",
}

func resourceVcdIndependentDiskCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	size, sizeProvided := d.GetOk("size")
	sizeInBytes, sizeInBytesProvided := d.GetOk("size_in_bytes")

	if !sizeProvided && !sizeInBytesProvided {
		return fmt.Errorf("size_in_bytes isn't provided")
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	diskName := d.Get("name").(string)
	diskRecord, _ := vdc.QueryDisk(diskName)

	if diskRecord != (govcd.DiskRecord{}) {
		return fmt.Errorf("disk with such name already exist : %s", diskName)
	}

	var diskCreateParams *types.DiskCreateParams
	if sizeProvided {
		diskCreateParams = &types.DiskCreateParams{Disk: &types.Disk{
			Name: diskName,
			Size: int64(size.(float64) * 1024 * 1024),
		}}
	}
	if sizeInBytesProvided {
		diskCreateParams = &types.DiskCreateParams{Disk: &types.Disk{
			Name: diskName,
			Size: int64(sizeInBytes.(int)),
		}}
	}

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

	diskCreateParams.Disk.Description = d.Get("description").(string)

	task, err := vdc.CreateDisk(diskCreateParams)
	if err != nil {
		return fmt.Errorf("error creating independent disk: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting to finish creation of independent disk: %s", err)
	}

	diskHref := task.Task.Owner.HREF
	disk, err := vdc.GetDiskByHref(diskHref)
	if govcd.IsNotFound(err) {
		log.Printf("unable to find disk with href %s: %s. Removing from state", diskHref, err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to find disk with href %s: %s", diskHref, err)
	}

	d.SetId(disk.Disk.Id)

	return resourceVcdIndependentDiskRead(d, meta)
}

func resourceVcdIndependentDiskRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc("", d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	identifier := d.Id()
	var disk *govcd.Disk
	if identifier != "" {
		disk, err = vdc.GetDiskById(identifier, true)
		if govcd.IsNotFound(err) {
			log.Printf("unable to find disk with ID %s: %s. Removing from state", identifier, err)
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("unable to find disk with ID %s: %s", identifier, err)
		}
	} else {
		identifier = d.Get("name").(string)
		disks, err := vdc.GetDisksByName(identifier, true)
		if govcd.IsNotFound(err) {
			log.Printf("unable to find disk with ID %s: %s. Removing from state", identifier, err)
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("unable to find disk with ID %s: %s", identifier, err)
		}
		if len(*disks) > 1 {
			return fmt.Errorf("found more than one disk with ID %s: %s", identifier, err)
		}
		disk = &(*disks)[0]
	}

	diskRecord, err := vdc.QueryDisk(disk.Disk.Name)
	if err != nil {
		return fmt.Errorf("unable to query disk with ID %s: %s", identifier, err)
	}

	d.SetId(disk.Disk.Id)
	_ = d.Set("name", disk.Disk.Name)
	_ = d.Set("description", disk.Disk.Description)
	_ = d.Set("storage_profile", disk.Disk.StorageProfile.Name)
	_ = d.Set("size_in_bytes", disk.Disk.Size)
	_ = d.Set("bus_type", busTypesFromValues[disk.Disk.BusType])
	_ = d.Set("bus_sub_type", busSubTypesFromValues[disk.Disk.BusSubType])
	_ = d.Set("iops", disk.Disk.Iops)
	_ = d.Set("owner_name", disk.Disk.Owner.User.Name)
	_ = d.Set("datastore_name", diskRecord.Disk.DataStoreName)
	_ = d.Set("is_attached", diskRecord.Disk.IsAttached)

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

	disk, err := vdc.GetDiskByHref(diskRecord.Disk.HREF)
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

// resourceVcdIndependentDiskImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_independent_disk.my-disk
// Example import path (_the_id_string_): org-name.vdc-name.my-independent-disk-name
func resourceVcdIndependentDiskImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[independent disk import] resource name must be specified as org-name.vdc-name.my-independent-disk-name")
	}
	orgName, vdcName, diskName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[independent disk import] unable to find VDC %s: %s ", vdcName, err)
	}

	disks, err := vdc.GetDisksByName(diskName, false)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find independent disk with name %s: %s",
			d.Id(), err)
	}
	if len(*disks) > 1 {
		return []*schema.ResourceData{}, fmt.Errorf("found more than one independent disk with name %s: %s",
			d.Id(), err)
	}

	d.SetId((*disks)[0].Disk.Id)
	return []*schema.ResourceData{d}, nil
}
