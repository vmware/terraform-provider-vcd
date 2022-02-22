package vcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdIndependentDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIndependentDiskCreate,
		ReadContext:   resourceVcdIndependentDiskRead,
		DeleteContext: resourceVcdIndependentDiskDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdIndependentDiskImport,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"size_in_mb": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "size in MB",
			},
			"bus_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateBusType,
			},
			"bus_sub_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateBusSubType,
			},
			"iops": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IOPS request for the created disk",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owner name of the disk",
			},
			"datastore_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Datastore name",
			},
			"is_attached": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the disk is already attached",
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
				Description: "Key value map of metadata to assign to this disk. Key and value can be any string.",
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
	"ide":              "IDE",
	"buslogic":         "buslogic",
	"lsilogic":         "lsilogic",
	"lsilogicsas":      "lsilogicsas",
	"VirtualSCSI":      "VirtualSCSI",
	"vmware.sata.ahci": "ahci",
}

func resourceVcdIndependentDiskCreate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	size, sizeProvided := d.GetOk("size_in_mb")

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	diskName := d.Get("name").(string)
	diskRecord, err := vdc.QueryDisk(diskName)
	if diskRecord != (govcd.DiskRecord{}) || err == nil {
		return diag.Errorf("disk with such name already exist : %s", diskName)
	}

	var diskCreateParams = &types.DiskCreateParams{
		Disk: &types.Disk{
			Name: diskName,
		},
	}

	if sizeProvided {
		diskCreateParams.Disk.SizeMb = int64(size.(int))
	}

	var storageReference types.Reference
	storageProfileValue := d.Get("storage_profile").(string)

	if storageProfileValue != "" {
		storageReference, err = vdc.FindStorageProfileReference(storageProfileValue)
		if err != nil {
			return diag.Errorf("error finding storage profile %s", storageProfileValue)
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
		return diag.Errorf("error creating independent disk: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error waiting to finish creation of independent disk: %s", err)
	}

	diskHref := task.Task.Owner.HREF
	disk, err := vdc.GetDiskByHref(diskHref)
	if err != nil {
		return diag.Errorf("unable to find disk with href %s: %s", diskHref, err)
	}

	d.SetId(disk.Disk.Id)

	err = createOrUpdateDiskMetadata(d, disk)
	if err != nil {
		return diag.Errorf("error adding metadata to independent disk: %s", err)
	}

	return resourceVcdIndependentDiskRead(c, d, meta)
}

func resourceVcdIndependentDiskRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
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
			return diag.Errorf("unable to find disk with ID %s: %s", identifier, err)
		}
	} else {
		identifier = d.Get("name").(string)
		disks, err := vdc.GetDisksByName(identifier, true)
		if govcd.IsNotFound(err) {
			log.Printf("unable to find disk with name %s: %s. Removing from state", identifier, err)
			d.SetId("")
			return nil
		}
		if err != nil {
			return diag.Errorf("unable to find disk with name %s: %s", identifier, err)
		}
		if len(*disks) > 1 {
			return diag.Errorf("found more than one disk with name %s: %s", identifier, err)
		}
		disk = &(*disks)[0]
	}

	diskRecords, err := vdc.QueryDisks(disk.Disk.Name)
	if err != nil {
		return diag.Errorf("unable to query disk with name %s: %s", identifier, err)
	}

	var diskRecord *types.DiskRecordType
	for _, entity := range *diskRecords {
		if entity.HREF == disk.Disk.HREF {
			diskRecord = entity
		}
	}

	if diskRecord == nil {
		return diag.Errorf("unable to find queried disk with name %s: and href: %s, %s", identifier, disk.Disk.HREF, err)
	}

	err = setMainData(d, disk)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "datastore_name", diskRecord.DataStoreName)
	dSet(d, "is_attached", diskRecord.IsAttached)

	log.Printf("[TRACE] Disk read completed.")
	return nil
}

func setMainData(d *schema.ResourceData, disk *govcd.Disk) error {
	d.SetId(disk.Disk.Id)
	dSet(d, "name", disk.Disk.Name)
	dSet(d, "description", disk.Disk.Description)
	dSet(d, "storage_profile", disk.Disk.StorageProfile.Name)
	dSet(d, "size_in_mb", disk.Disk.SizeMb)
	dSet(d, "bus_type", busTypesFromValues[disk.Disk.BusType])
	dSet(d, "bus_sub_type", busSubTypesFromValues[disk.Disk.BusSubType])
	dSet(d, "iops", disk.Disk.Iops)
	dSet(d, "owner_name", disk.Disk.Owner.User.Name)

	metadata, err := disk.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Unable to get Disk metadata")
		return fmt.Errorf("unable to get Disk metadata %s", err)
	}
	if err := d.Set("metadata", getMetadataStruct(metadata.MetadataEntry)); err != nil {
		return fmt.Errorf("error setting metadata: %s", err)
	}
	return nil
}

func resourceVcdIndependentDiskDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	diskRecord, err := vdc.QueryDisk(d.Get("name").(string))
	if err != nil {
		d.SetId("")
		return diag.Errorf("error finding disk : %#v", err)
	}

	if diskRecord.Disk.IsAttached {
		return diag.Errorf("can not remove disk as it is attached to vm")
	}

	disk, err := vdc.GetDiskByHref(diskRecord.Disk.HREF)
	if err != nil {
		d.SetId("")
		return diag.Errorf("error getting disk : %#v", err)
	}

	task, err := disk.Delete()
	if err != nil {
		d.SetId("")
		return diag.Errorf("error deleting disk : %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		d.SetId("")
		return diag.Errorf("error waiting for deleting disk : %#v", err)
	}

	return nil
}

var errHelpDiskImport = fmt.Errorf(`resource id must be specified in one of these formats:
'org-name.vdc-name.my-independent-disk-id' to import by rule id
'list@org-name.vdc-name.my-independent-disk-name' to get a list of disks with their IDs`)

// resourceVcdIndependentDiskImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2a. If the `_the_id_string_` contains a dot formatted path to resource as in the example below
// it will try to import it. If it is found - the ID is set
// 2b. If the `_the_id_string_` starts with `list@` and contains path to disk name similar to
// `list@org-name.vdc-name.my-independent-disk-name` then the function lists all independent disks and their IDs in that vdc
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_independent_disk.my-disk
// Example import path (_the_id_string_): org-name.vdc-name.my-independent-disk-id
// Example list path (_the_id_string_): list@org-name.vdc-name.my-independent-disk-name
func resourceVcdIndependentDiskImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	var commandOrgName, orgName, vdcName, diskName, diskId string

	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing vcd_independent_disk resource with provided id %s", d.Id())

	if len(resourceURI) != 3 {
		return nil, errHelpDiskImport
	}

	if strings.Contains(d.Id(), "list@") {
		commandOrgName, vdcName, diskName = resourceURI[0], resourceURI[1], resourceURI[2]
		commandOrgNameSplit := strings.Split(commandOrgName, "@")
		if len(commandOrgNameSplit) != 2 {
			return nil, errHelpDiskImport
		}
		orgName = commandOrgNameSplit[1]
		return listDisksForImport(meta, orgName, vdcName, diskName)
	} else {
		orgName, vdcName, diskId = resourceURI[0], resourceURI[1], resourceURI[2]
		return getDiskForImport(d, meta, orgName, vdcName, diskId)
	}
}

func getDiskForImport(d *schema.ResourceData, meta interface{}, orgName, vdcName, diskId string) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[independent disk import] unable to find VDC %s: %s ", vdcName, err)
	}

	disk, err := vdc.GetDiskById(diskId, false)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find independent disk with id %s: %s",
			d.Id(), err)
	}

	d.SetId(disk.Disk.Id)
	dSet(d, "name", disk.Disk.Name)
	return []*schema.ResourceData{d}, nil
}

func listDisksForImport(meta interface{}, orgName, vdcName, diskName string) ([]*schema.ResourceData, error) {

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[independent disk import] unable to find VDC %s: %s ", vdcName, err)
	}

	buf := new(bytes.Buffer)
	_, err = fmt.Fprintln(buf, "Retrieving all disks by name")
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
	}
	disks, err := vdc.GetDisksByName(diskName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve disks by name: %s", err)
	}

	writer := tabwriter.NewWriter(buf, 0, 8, 1, '\t', tabwriter.AlignRight)

	_, err = fmt.Fprintf(writer, "No\tID\tName\tDescription\tSizeMb\n")
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
	}
	_, err = fmt.Fprintf(writer, "--\t--\t----\t------\t----\n")
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
	}
	for index, disk := range *disks {
		_, err = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%d\n", index+1, disk.Disk.Id, disk.Disk.Name, disk.Disk.Description, disk.Disk.SizeMb)
		if err != nil {
			logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
		}
	}
	err = writer.Flush()
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error flushing buffer: %s", err))
	}

	return nil, fmt.Errorf("resource was not imported! %s\n%s", errHelpDiskImport, buf.String())
}

func createOrUpdateDiskMetadata(d *schema.ResourceData, disk *govcd.Disk) error {
	log.Printf("[TRACE] adding/updating metadata to Org")

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
			err := disk.DeleteMetadata(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		// Add new metadata
		for k, v := range newMetadata {
			_, err := disk.AddMetadata(k, v.(string))
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}
