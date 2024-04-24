package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdSolutionAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionAddonCreate,
		ReadContext:   resourceVcdSolutionAddonRead,
		UpdateContext: resourceVcdSolutionAddonUpdate,
		DeleteContext: resourceVcdSolutionAddonDelete,
		Importer:      &schema.ResourceImporter{
			// StateContext: resourceVcdSolutionAddonImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog_item_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "absolute or relative path to Solution Add-on ISO file",
			},
			"addon_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "absolute or relative path to Solution Add-on ISO file",
			},
			// Trust certificate - should we untrust (remove the certificate) in "update"?
			"trust_certificate": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "",
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "",
			},
			"state": {
				Type:        schema.TypeString,
				Description: "State reports RDE state",
				Computed:    true,
			},
		},
	}
}

func resourceVcdSolutionAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !d.Get("accept_eula").(bool) {
		return diag.Errorf("cannot create Solution Add-on without accepting EULA")
	}

	createCfg := govcd.SolutionAddOnConfig{
		IsoFilePath:          d.Get("addon_path").(string),
		User:                 "administrator",
		CatalogItemId:        d.Get("catalog_item_id").(string),
		AcceptEula:           d.Get("accept_eula").(bool),
		AutoTrustCertificate: d.Get("trust_certificate").(bool),
	}
	addon, err := vcdClient.CreateSolutionAddOn(createCfg)

	if err != nil {
		return diag.Errorf("error configuring Solution Add-on: %s", err)
	}

	d.SetId(addon.DefinedEntity.DefinedEntity.ID)

	return resourceVcdSolutionAddonRead(ctx, d, meta)
}

func resourceVcdSolutionAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if d.HasChange("catalog_item_id") {
		addon, err := vcdClient.GetSolutionAddonById(d.Id())
		if err != nil {
			return diag.Errorf("error retrieving ID: %s", err)
		}

		addon.SolutionEntity.Origin.CatalogItemId = d.Get("catalog_item_id").(string)

		_, err = addon.Update(addon.SolutionEntity)
		if err != nil {
			return diag.Errorf("error updating Solution Add-On: %s", err)
		}
	}

	return resourceVcdSolutionAddonRead(ctx, d, meta)
}

func resourceVcdSolutionAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionAddonById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-on: %s", err)
	}

	// dSet(d, "user", slz.SolutionEntity.Origin.AcceptedBy)
	dSet(d, "state", slz.DefinedEntity.DefinedEntity.State)
	dSet(d, "catalog_item_id", slz.SolutionEntity.Origin.CatalogItemId)

	return nil
}

func resourceVcdSolutionAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	entity, err := vcdClient.GetSolutionAddonById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving ID: %s", err)
	}
	err = entity.Delete()
	if err != nil {
		return diag.Errorf("error deleting Solution Add-on RDE: %s", err)
	}

	return nil
}

/* func resourceVcdSolutionAddonImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetExactlyOneSolutionAddon()
	if err != nil {
		return nil, fmt.Errorf("error finding Solution Landing Zone: %s", err)
	}

	if slz != nil && slz.SolutionAddonType != nil {
		dSet(d, "org", slz.SolutionAddonType.Name)
	}
	d.SetId(slz.Id())

	return []*schema.ResourceData{d}, nil
}
*/
