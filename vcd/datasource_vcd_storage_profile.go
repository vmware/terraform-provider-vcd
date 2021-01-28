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
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of storage profile",
			},
		},
	}
}

func datasourceVcdStorageProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return nil
}
