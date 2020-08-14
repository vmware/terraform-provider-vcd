package vcd

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcIndependentDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVcdIndependentDiskRead,
		Schema: map[string]*schema.Schema{
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
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// we enable this when when we solve https://github.com/vmware/terraform-provider-vcd/issues/355
			/*			"size_in_bytes": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "size in bytes",
					},*/
			"bus_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"bus_sub_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceVcdIndependentDiskRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc("", d.Get("vdc").(string))
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	idValue := d.Get("id").(string)
	nameValue := d.Get("name").(string)

	if idValue == "" && nameValue == "" {
		return errors.New("`id` and `name` are empty. At least one is needed")
	}

	identifier := idValue
	var disk *govcd.Disk
	if identifier != "" {
		disk, err = vdc.GetDiskById(identifier, true)
		if govcd.IsNotFound(err) {
			log.Printf("unable to find disk with ID %s: %s. Removing from state", identifier, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("unable to find disk with ID %s: %s", identifier, err)
		}
	} else {
		identifier = nameValue
		disks, err := vdc.GetDisksByName(identifier, true)
		if err != nil {
			return fmt.Errorf("unable to find disk with name %s: %s", identifier, err)
		}
		if len(*disks) > 1 {
			var diskIds []string
			for _, disk := range *disks {
				diskIds = append(diskIds, disk.Disk.Id)
			}
			return fmt.Errorf("found more than one disk with name %s. Disk ids are: %s. Please use `id` property", identifier, diskIds)
		}
		disk = &(*disks)[0]
	}

	diskRecords, err := vdc.QueryDisks(disk.Disk.Name)
	if err != nil {
		return fmt.Errorf("unable to query disk with name %s: %s", identifier, err)
	}

	var diskRecord *types.DiskRecordType
	for _, entity := range *diskRecords {
		if entity.HREF == disk.Disk.HREF {
			diskRecord = entity
		}
	}

	if diskRecord == nil {
		return fmt.Errorf("unable to find queried disk with name %s: and href: %s, %s", identifier, disk.Disk.HREF, err)
	}

	setMainData(d, disk)
	_ = d.Set("datastore_name", diskRecord.DataStoreName)
	_ = d.Set("is_attached", diskRecord.IsAttached)

	log.Printf("[TRACE] Disk read completed.")
	return nil
}
