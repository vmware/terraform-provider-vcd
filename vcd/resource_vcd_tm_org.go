package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmOrg = "TM Organization"

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
				Description: fmt.Sprintf("The unique identifier in the full URL with which users log in to this %s", labelTmOrg),
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Appears in the Cloud application as a human-readable name of the %s", labelTmOrg),
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
				Description: fmt.Sprintf("Defines if the %s enabled. Defaults to 'true'", labelTmOrg),
			},
			"is_subprovider": {
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Description: fmt.Sprintf("Enables this organization to manage other %ss", labelTmOrg),
			},
			// TODO: TM: validate if all of these computed attributes are effective
			"org_vdc_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of VDCs belonging to the %s", labelTmOrg),
			},
			"catalog_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of catalog belonging to the %s", labelTmOrg),
			},
			"vapp_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of vApps belonging to the %s", labelTmOrg),
			},
			"running_vm_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of running VMs in the %s", labelTmOrg),
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of users in the %s", labelTmOrg),
			},
			"disk_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of disks in the %s", labelTmOrg),
			},
			"can_publish": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines whether the %s can publish catalogs externally", labelTmOrg),
			},
			"directly_managed_org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of directly managed %ss", labelTmOrg),
			},
			"is_classic_tenant": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines whether the %s is a classic VRA-style tenant", labelTmOrg),
			},
		},
	}
}

func resourceVcdTmOrgCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:      labelTmOrg,
		getTypeFunc:      getTmOrgType,
		stateStoreFunc:   setTmOrgData,
		createFunc:       vcdClient.CreateTmOrg,
		resourceReadFunc: resourceVcdTmOrgRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmOrgUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:      labelTmOrg,
		getTypeFunc:      getTmOrgType,
		getEntityFunc:    vcdClient.GetTmOrgById,
		resourceReadFunc: resourceVcdTmOrgRead,
		// TODO: TM: review if ID and ManagedBy should always be submitted on update
		preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmOrg, *types.TmOrg]{resubmitIdAndManagedByFields},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:    labelTmOrg,
		getEntityFunc:  vcdClient.GetTmOrgById,
		stateStoreFunc: setTmOrgData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmOrgDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:    labelTmOrg,
		getEntityFunc:  vcdClient.GetTmOrgById,
		preDeleteHooks: []outerEntityHook[*govcd.TmOrg]{disableTmOrg}, // Org must be disabled before deletion
	}

	return deleteResource(ctx, d, meta, c)
}

// disableTmOrg disables Org which is usefull before deletion as a non-disabled Org cannot be
// removed
func disableTmOrg(t *govcd.TmOrg) error {
	if t.TmOrg.IsEnabled {
		return t.Disable()
	}
	return nil
}

func resubmitIdAndManagedByFields(o *govcd.TmOrg, i *types.TmOrg) error {
	// TODO: TM: review if ManagedBy should always be submitted
	i.ID = o.TmOrg.ID

	i.ManagedBy = o.TmOrg.ManagedBy // It is optional in docs, but API rejects update without it
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

func setTmOrgData(d *schema.ResourceData, org *govcd.TmOrg) error {
	if org == nil || org.TmOrg == nil {
		return fmt.Errorf("cannot save state for nil Org")
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "name", org.TmOrg.Name)
	dSet(d, "display_name", org.TmOrg.DisplayName)
	dSet(d, "description", org.TmOrg.Description)
	dSet(d, "is_enabled", org.TmOrg.IsEnabled)
	dSet(d, "is_subprovider", org.TmOrg.CanManageOrgs)

	// Computed in resource
	dSet(d, "org_vdc_count", org.TmOrg.OrgVdcCount)
	dSet(d, "catalog_count", org.TmOrg.CatalogCount)
	dSet(d, "vapp_count", org.TmOrg.VappCount)
	dSet(d, "running_vm_count", org.TmOrg.RunningVMCount)
	dSet(d, "user_count", org.TmOrg.UserCount)
	dSet(d, "disk_count", org.TmOrg.DiskCount)
	dSet(d, "can_publish", org.TmOrg.CanPublish)
	dSet(d, "directly_managed_org_count", org.TmOrg.DirectlyManagedOrgCount)
	dSet(d, "is_classic_tenant", org.TmOrg.IsClassicTenant)

	return nil
}
