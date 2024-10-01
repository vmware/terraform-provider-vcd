package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdTmOrg() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmOrgCreate,
		ReadContext:   resourceVcdTmOrgRead,
		UpdateContext: resourceVcdTmOrgUpdate,
		DeleteContext: resourceVcdTmOrgDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmOrgImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier in the full URL with which users log in to this organization",
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Appears in the Cloud application as a human-readable name of the organization",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Is the org enabled. Defaults to 'true'",
			},
			"is_subprovider": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enables this organization to manage other organizations",
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

func resourceVcdTmOrgCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getTmOrgType(d)
	if err != nil {
		return diag.Errorf("error getting Org Type: %s", err)
	}

	o, err := vcdClient.CreateTmOrg(t)
	if err != nil {
		return diag.Errorf("error creating Org: %s", err)
	}

	d.SetId(o.TmOrg.ID)
	return resourceVcdTmOrgRead(ctx, d, meta)
}

func resourceVcdTmOrgUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	o, err := vcdClient.GetTmOrgById(d.Id())
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	t, err := getTmOrgType(d)
	if err != nil {
		return diag.Errorf("error getting Org Type: %s", err)
	}

	t.ID = o.TmOrg.ID

	/// Injecting some fields that failed requirements
	t.ManagedBy = o.TmOrg.ManagedBy // It is optional in docs, but API rejects

	/// <> Injecting some fields that failed requirements

	_, err = o.Update(t)
	if err != nil {
		return diag.Errorf("error updating Org: %s", err)
	}

	return resourceVcdTmOrgRead(ctx, d, meta)
}

func resourceVcdTmOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	o, err := vcdClient.GetTmOrgById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting Org: %s", err)
	}

	err = setTmOrgData(d, o.TmOrg)
	if err != nil {
		return diag.Errorf("error storing Org data: %s", err)
	}

	return nil
}

func resourceVcdTmOrgDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	o, err := vcdClient.GetTmOrgById(d.Id())
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	if o.TmOrg.IsEnabled {
		err = o.Disable()
		if err != nil {
			return diag.Errorf("error disableing Org for deletion: %s", err)
		}
	}

	err = o.Delete()
	if err != nil {
		return diag.Errorf("error deleting Org: %s", err)
	}

	return nil
}

func resourceVcdTmOrgImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	o, err := vcdClient.GetTmOrgByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	d.SetId(o.TmOrg.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmOrgType(d *schema.ResourceData) (*types.TmOrg, error) {
	t := &types.TmOrg{
		Name:          d.Get("name").(string),
		DisplayName:   d.Get("display_name").(string),
		Description:   d.Get("description").(string),
		IsEnabled:     d.Get("is_enabled").(bool),
		CanManageOrgs: d.Get("is_subprovider").(bool),
	}

	return t, nil

}

func setTmOrgData(d *schema.ResourceData, org *types.TmOrg) error {
	dSet(d, "name", org.Name)
	dSet(d, "display_name", org.DisplayName)
	dSet(d, "description", org.Description)
	dSet(d, "is_enabled", org.IsEnabled)
	dSet(d, "is_subprovider", org.CanManageOrgs)

	// Computed
	dSet(d, "org_vdc_count", org.OrgVdcCount)
	dSet(d, "catalog_count", org.CatalogCount)
	dSet(d, "vapp_count", org.VappCount)
	dSet(d, "running_vm_count", org.RunningVMCount)
	dSet(d, "user_count", org.UserCount)
	dSet(d, "disk_count", org.DiskCount)
	dSet(d, "can_publish", org.CanPublish)
	dSet(d, "directly_managed_org_count", org.DirectlyManagedOrgCount)
	dSet(d, "is_classic_tenant", org.IsClassicTenant)

	return nil
}
