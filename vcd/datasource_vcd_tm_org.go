package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdTmOrg() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmOrgRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier in the full URL with which users log in to this organization",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Appears in the Cloud application as a human-readable name of the organization",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines is the Org enabled",
			},
			"is_subprovider": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this Organization can manage other organizations",
			},
			"org_vdc_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of VDCs belonging to the Org",
			},
			"catalog_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of catalogs belonging to the Org",
			},
			"vapp_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of vApps in the Org",
			},
			"running_vm_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of running VMs in the Org",
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users in the Org",
			},
			"disk_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of disks in the Org",
			},
			"can_publish": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"directly_managed_org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of directly managed Orgs",
			},
			"is_classic_tenant": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func datasourceVcdTmOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	o, err := vcdClient.GetTmOrgByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	err = setTmOrgData(d, o.TmOrg)
	if err != nil {
		return diag.Errorf("error storing Org data: %s", err)
	}

	d.SetId(o.TmOrg.ID)

	return nil
}
