package vcd

import (
	"context"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdStorageProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdStorageProfileRead,
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of storage profile",
			},
			"limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of storage bytes (scaled by 'units' field) allocated for this profile. `0` means `maximum possible`",
			},
			"used_storage": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Storage used, in Megabytes, by the storage profile",
			},
			"default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this storage profile is enabled for use in the VDC",
			},
			"iops_allocated": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total IOPS currently allocated to this storage profile",
			},
			"units": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Scale used to define Limit",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata retrieved from this storage profile. Key and value can be any string.",
			},
			"iops_settings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "IOPs related settings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops_limiting_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if this storage profile is IOPS-based placement enabled",
						},
						"maximum_disk_iops": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum IOPS value that this storage profile is permitted to deliver. Value of 0 means this max setting is disabled and there is no max disk IOPS restriction",
						},
						"default_disk_iops": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Value of 0 for disk IOPS means that no IOPS would be reserved or provisioned for that virtual disk",
						},
						"disk_iops_per_gb_max": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum disk IOPs per GB value that this storage profile is permitted to deliver. A value of 0 means there is no per GB IOPS restriction",
						},
						"iops_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum number of IOPs that can be allocated for this profile. `0` means `maximum possible`",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdStorageProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error reading Org and VDC: %s", err)
	}

	name := d.Get("name").(string)
	storageProfileReference, err := vdc.FindStorageProfileReference(name)
	if err != nil {
		return diag.Errorf("%s: error finding Storage Profile '%s' in VDC '%s': %s",
			govcd.ErrorEntityNotFound, name, vdc.Vdc.Name, err)
	}
	d.SetId(storageProfileReference.ID)

	storageProfile, err := vcdClient.GetStorageProfileByHref(storageProfileReference.HREF)
	if err != nil {
		return diag.Errorf("%s: error fetching additional details for Storage Profile '%s' in VDC '%s': %s",
			govcd.ErrorEntityNotFound, name, vdc.Vdc.Name, err)
	}

	dSet(d, "limit", storageProfile.Limit)
	dSet(d, "used_storage", storageProfile.StorageUsedMB)
	dSet(d, "default", storageProfile.Default)
	dSet(d, "enabled", storageProfile.Enabled)
	dSet(d, "iops_allocated", storageProfile.IopsAllocated)
	dSet(d, "units", storageProfile.Units)
	if storageProfile.IopsSettings != nil {

		var iopsSettings = make(map[string]interface{})

		iopsSettings["iops_limit"] = storageProfile.IopsSettings.StorageProfileIopsLimit
		iopsSettings["iops_limiting_enabled"] = storageProfile.IopsSettings.Enabled
		iopsSettings["default_disk_iops"] = storageProfile.IopsSettings.DiskIopsDefault
		iopsSettings["maximum_disk_iops"] = storageProfile.IopsSettings.DiskIopsMax
		iopsSettings["disk_iops_per_gb_max"] = storageProfile.IopsSettings.DiskIopsPerGbMax

		err = d.Set("iops_settings", []map[string]interface{}{iopsSettings})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	metadata, err := vcdClient.GetMetadataByHref(storageProfileReference.HREF)
	if err != nil {
		return diag.Errorf("Unable to find metadata for storage profile: %s", err)
	}
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
