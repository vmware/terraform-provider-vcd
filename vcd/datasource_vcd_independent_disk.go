package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcIndependentDisk() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVcdIndependentDiskRead,
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
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size_in_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "size in MB",
			},
			"bus_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bus_sub_type": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for disk metadata",
			},
		},
	}
}

func dataSourceVcdIndependentDiskRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc("", d.Get("vdc").(string))
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	idValue := d.Get("id").(string)
	nameValue := d.Get("name").(string)

	if idValue == "" && nameValue == "" {
		return diag.Errorf("`id` and `name` are empty. At least one is needed")
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
			return diag.Errorf("unable to find disk with ID %s: %s", identifier, err)
		}
	} else {
		identifier = nameValue
		disks, err := vdc.GetDisksByName(identifier, true)
		if err != nil {
			return diag.Errorf("unable to find disk with name %s: %s", identifier, err)
		}
		if len(*disks) > 1 {
			var diskIds []string
			for _, disk := range *disks {
				diskIds = append(diskIds, disk.Disk.Id)
			}
			return diag.Errorf("found more than one disk with name %s. Disk ids are: %s. Please use `id` property", identifier, diskIds)
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
