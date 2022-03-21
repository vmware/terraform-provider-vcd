package vcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Key for independent disk resource lock when disk type is shared and attached to more than one VM
const globalIndependentDiskLockKey = "globalIndependentDiskLockKey"

func resourceVcdIndependentDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIndependentDiskCreate,
		ReadContext:   resourceVcdIndependentDiskRead,
		UpdateContext: resourceVcdIndependentDiskUpdate,
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
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"size_in_mb": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "size in MB",
			},
			"bus_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateBusType,
			},
			"bus_sub_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateBusSubType,
			},
			"encrypted": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if disk is encrypted",
			},
			"sharing_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"DiskSharing", "ControllerSharing"}, false),
				Description:  "This is the sharing type. This attribute can only have values defined one of: `DiskSharing`,`ControllerSharing`",
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of this named disk's device backing",
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
			"attached_vm_ids": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of VM IDs which are using the disk",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

var busTypes = map[string]string{
	"IDE":  "5",
	"SCSI": "6",
	"SATA": "20",
	"NVME": "20",
}

var busTypesFromValues = map[string]string{
	"5":  "IDE",
	"6":  "SCSI",
	"20": "SATA",
	"21": "NVME", // in API NVME is 20, the same as SATA. In state file we save 21 to know if it's NVME or SATA
}

var busSubTypes = map[string]string{
	"ide":            "IDE",
	"buslogic":       "buslogic",
	"lsilogic":       "lsilogic",
	"lsilogicsas":    "lsilogicsas",
	"virtualscsi":    "VirtualSCSI",
	"ahci":           "vmware.sata.ahci",
	"nvmecontroller": "vmware.nvme.controller",
}

var busSubTypesFromValues = map[string]string{
	"ide":                    "IDE",
	"buslogic":               "buslogic",
	"lsilogic":               "lsilogic",
	"lsilogicsas":            "lsilogicsas",
	"VirtualSCSI":            "VirtualSCSI",
	"vmware.sata.ahci":       "ahci",
	"vmware.nvme.controller": "nvmecontroller",
}

func resourceVcdIndependentDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if value, ok := d.GetOk("sharing_type"); ok {
		if vcdClient.Client.APIVCDMaxVersionIs("< 36.0") {
			return diag.Errorf("`sharing_type` is supported from VCD 10.3+ version")
		}
		diskCreateParams.Disk.SharingType = value.(string)
	}

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

	return resourceVcdIndependentDiskRead(ctx, d, meta)
}

func resourceVcdIndependentDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if d.HasChanges("size_in_mb", "storage_profile", "description") {
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf(errorRetrievingOrgAndVdc, err)
		}

		storageProfileValue := d.Get("storage_profile").(string)
		var storageProfileRef *types.Reference

		if storageProfileValue != "" {
			storageReference, err := vdc.FindStorageProfileReference(storageProfileValue)
			if err != nil {
				return diag.Errorf("error finding storage profile %s", storageProfileValue)
			}
			storageProfileRef = &types.Reference{HREF: storageReference.HREF}
		}

		disk, err := vdc.GetDiskById(d.Id(), true)
		if err != nil {
			return diag.Errorf("error fetching independent disk: %s", err)
		}

		diskAttachedVmsHrefs, err := disk.GetAttachedVmsHrefs()
		if err != nil {
			return diag.Errorf("error resourceVcdIndependentDiskUpdate faced issue fetching attached VMs")
		}

		// Lock on resource if independent disk is shared between more than one VM. This allows to avoid deadlock below
		// when different independent shared disks try to acquire locks on same VMs
		// Avoids situation like this:
		// DiskA attached to VM2 and VM1 -> locked VM2, but can't lock VM1 because
		// DiskB attached to VM1 and VM2 -> locked VM1, but can't lock VM2 (because it was locked by DiskA already)
		// This is "blind" lock as it locks resource even independent disks are attached to different VMs.
		if len(diskAttachedVmsHrefs) > 1 {
			lockIndependentDiskOpsGlobally()
			defer unlockIndependentDiskOpsGlobally()
		}
		// Lock on attached VMs as another independent disk resource attached to the same VMs may be already doing an update
		// DiskA attached to VM1 -> locks VM1
		// DiskB attached to VM1 -> locks VM1 when DiskA releases lock for VM1
		lockVmsForIndependentDisks(diskAttachedVmsHrefs)
		defer unlockVmsForIndependentDisks(diskAttachedVmsHrefs)

		diskDetailsForReAttach, diagErr := detachVms(vcdClient, disk, diskAttachedVmsHrefs)
		if diagErr != nil {
			return diagErr
		}

		err = disk.Refresh()
		if err != nil {
			return diag.Errorf("error resourceVcdIndependentDiskUpdate error refreshing independent disk: %s", err)
		}

		disk.Disk.SizeMb = int64(d.Get("size_in_mb").(int))
		disk.Disk.Description = d.Get("description").(string)
		if storageProfileRef != nil {
			disk.Disk.StorageProfile = storageProfileRef
		}

		task, err := disk.Update(disk.Disk)
		if err != nil {
			return diag.Errorf("error updating independent disk: %s", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error waiting to finish updating of independent disk: %s", err)
		}

		diagErr = attachBackVms(vcdClient, disk, diskDetailsForReAttach, diskAttachedVmsHrefs)
		if diagErr != nil {
			return diagErr
		}

	}
	return resourceVcdIndependentDiskRead(ctx, d, meta)
}

// lockIndependentDiskOpsGlobally acquire lock for independent disk resource using key `globalIndependentDiskLockKey`
func lockIndependentDiskOpsGlobally() {
	vcdMutexKV.kvLock(globalIndependentDiskLockKey)
}

// unlockIndependentDiskOpsGlobally release lock for independent disk resource using key `globalIndependentDiskLockKey`
func unlockIndependentDiskOpsGlobally() {
	vcdMutexKV.kvUnlock(globalIndependentDiskLockKey)
}

// lockVmsForIndependentDisks acquire locks to VMs which independent disk is attached
func lockVmsForIndependentDisks(sliceOfVmsHrefs []string) {
	for _, vmHref := range sliceOfVmsHrefs {
		key := fmt.Sprintf("independentDiskLock:%s", vmHref)
		vcdMutexKV.kvLock(key)
	}
}

// unlockVmsForIndependentDisks release locks to VMs which independent disk is attached
func unlockVmsForIndependentDisks(sliceOfVmsHrefs []string) {
	for _, vmHref := range sliceOfVmsHrefs {
		key := fmt.Sprintf("independentDiskLock:%s", vmHref)
		vcdMutexKV.kvUnlock(key)
	}
}

func detachVms(vcdClient *VCDClient, disk *govcd.Disk, sliceOfVmsHrefs []string) (map[string]types.DiskSettings, diag.Diagnostics) {
	diskDetailsForReAttach := make(map[string]types.DiskSettings)
	var vms []*govcd.VM
	for _, vmHref := range sliceOfVmsHrefs {
		vm, err := vcdClient.Client.GetVMByHref(vmHref)
		if err != nil {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate error fetching attached VM: %s", err)
		}
		vmStatus := types.VAppStatuses[vm.VM.Status]
		if vmStatus != "POWERED_OFF" && busTypesFromValues[disk.Disk.BusType] == "IDE" {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate can not detach disks type `IDE` for VM `%s` which is not powered off", vm.VM.Name)
		}
		vms = append(vms, vm)
	}

	for _, vm := range vms {
		isFoundDiskMatch := false
		if vm.VM != nil && vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.DiskSection != nil && vm.VM.VmSpecSection.DiskSection.DiskSettings != nil {
			for _, diskSettings := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
				if diskSettings.Disk != nil && diskSettings.Disk.HREF == disk.Disk.HREF {
					diskDetailsForReAttach[vm.VM.HREF] = *diskSettings
					isFoundDiskMatch = true
				}
			}
		} else {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate unexpected return from API, missing VmSpecSection or subtype")
		}

		if !isFoundDiskMatch {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate couldn't match Disk with VM disk")
		}
		detachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF}}

		task, err := vm.DetachDisk(detachParams)
		if err != nil {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate error detaching independent disk `%s` to vm %s", disk.Disk.Name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return nil, diag.Errorf("error resourceVcdIndependentDiskUpdate error waiting for task to complete detaching independent disk `%s` to vm %s", disk.Disk.Name, err)
		}
	}
	return diskDetailsForReAttach, nil
}

// attachBackVms reattaches independent disks back to VMs
func attachBackVms(vcdClient *VCDClient, disk *govcd.Disk, diskDetailsForReAttach map[string]types.DiskSettings, sliceOfVmsHrefs []string) diag.Diagnostics {
	for _, vmHref := range sliceOfVmsHrefs {
		vm, err := vcdClient.Client.GetVMByHref(vmHref)
		if err != nil {
			return diag.Errorf("error resourceVcdIndependentDiskUpdate error fetching attached VM: %s", err)
		}
		attachParams := &types.DiskAttachOrDetachParams{Disk: &types.Reference{HREF: disk.Disk.HREF},
			BusNumber:  takeIntPointer(diskDetailsForReAttach[vmHref].BusNumber),
			UnitNumber: takeIntPointer(diskDetailsForReAttach[vmHref].UnitNumber)}

		task, err := vm.AttachDisk(attachParams)
		if err != nil {
			return diag.Errorf("error resourceVcdIndependentDiskUpdate error attaching independent disk `%s` to vm %s", disk.Disk.Name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error resourceVcdIndependentDiskUpdate error waiting for task to complete detaching independent disk `%s` to vm %s", disk.Disk.Name, err)
		}
	}
	return nil
}

func resourceVcdIndependentDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	err = setMainData(d, disk, diskRecord)
	if err != nil {
		diag.FromErr(err)
	}

	log.Printf("[TRACE] Disk read completed.")
	return nil
}

func setMainData(d *schema.ResourceData, disk *govcd.Disk, diskRecord *types.DiskRecordType) error {
	d.SetId(disk.Disk.Id)
	dSet(d, "name", disk.Disk.Name)
	dSet(d, "description", disk.Disk.Description)
	dSet(d, "storage_profile", disk.Disk.StorageProfile.Name)
	dSet(d, "size_in_mb", disk.Disk.SizeMb)
	dSet(d, "bus_type", busTypesFromValues[disk.Disk.BusType])
	if disk.Disk.BusSubType == "vmware.nvme.controller" {
		dSet(d, "bus_type", busTypesFromValues["21"])
	}
	dSet(d, "bus_sub_type", busSubTypesFromValues[disk.Disk.BusSubType])
	dSet(d, "iops", disk.Disk.Iops)
	dSet(d, "owner_name", disk.Disk.Owner.User.Name)
	dSet(d, "datastore_name", diskRecord.DataStoreName)
	dSet(d, "is_attached", diskRecord.IsAttached)
	dSet(d, "encrypted", diskRecord.Encrypted)
	dSet(d, "sharing_type", diskRecord.SharingType)
	dSet(d, "uuid", diskRecord.UUID)

	vmsHrefs, err := disk.GetAttachedVmsHrefs()
	if err != nil {
		return fmt.Errorf("[Independent disk read] error fetching attached VMs IDs: %s ", err)
	}
	var attachedVmIds []string
	for _, vmHref := range vmsHrefs {
		attachedVmIds = append(attachedVmIds, extractUuid(vmHref))
	}
	attachedVmSet := convertStringsTotTypeSet(attachedVmIds)
	err = d.Set("attached_vm_ids", attachedVmSet)
	if err != nil {
		return fmt.Errorf("[Independent disk read] error setting the list of attached VM IDs: %s ", err)
	}
	return nil
}

func resourceVcdIndependentDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("can not remove disk %s as it is attached to vm", diskRecord.Disk.Name)
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
'list@org-name.vdc-name.my-independent-disk-name' or 'list@org-name.vdc-name' to get a list of disks with their IDs`)

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
func resourceVcdIndependentDiskImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	var commandOrgName, orgName, vdcName, diskName, diskId string

	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing vcd_independent_disk resource with provided id %s", d.Id())

	if len(resourceURI) != 3 && len(resourceURI) != 2 {
		return nil, errHelpDiskImport
	}

	if strings.Contains(d.Id(), "list@") {
		commandOrgName, vdcName = resourceURI[0], resourceURI[1]
		if len(resourceURI) == 3 {
			diskName = resourceURI[2]
		}
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

	writer := tabwriter.NewWriter(buf, 0, 8, 1, '\t', tabwriter.AlignRight)

	_, err = fmt.Fprintf(writer, "No\tID\tName\tDescription\tSizeMb\n")
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
	}
	_, err = fmt.Fprintf(writer, "--\t--\t----\t------\t----\n")
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
	}

	if diskName == "" {
		disksRecords, err := vdc.QueryDisks("*")
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve disks in VDC: %s", err)
		}
		for index, disk := range *disksRecords {
			uuid, err := govcd.GetUuidFromHref(disk.HREF, true)
			if err != nil {
				return nil, fmt.Errorf("error parsing disk ID : %s", err)
			}
			_, err = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%d\n", index+1, uuid, disk.Name, disk.Description, disk.SizeMb)
			if err != nil {
				logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
			}
		}
	} else {
		disks, err := vdc.GetDisksByName(diskName, false)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve disks by name: %s", err)
		}
		for index, disk := range *disks {
			_, err = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%d\n", index+1, disk.Disk.Id, disk.Disk.Name, disk.Disk.Description, disk.Disk.SizeMb)
			if err != nil {
				logForScreen("vcd_independent_disk", fmt.Sprintf("error writing to buffer: %s", err))
			}
		}

	}
	err = writer.Flush()
	if err != nil {
		logForScreen("vcd_independent_disk", fmt.Sprintf("error flushing buffer: %s", err))
	}
	return nil, fmt.Errorf("resource was not imported! %s\n%s", errHelpDiskImport, buf.String())
}
