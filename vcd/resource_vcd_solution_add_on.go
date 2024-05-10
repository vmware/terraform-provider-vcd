package vcd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"text/tabwriter"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionAddonImport,
		},

		Schema: map[string]*schema.Schema{
			"catalog_item_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Solution Add-On Catalog Item ID ",
			},
			"addon_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Absolute or relative path to Solution Add-On ISO file available locally",
			},
			"trust_certificate": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Defines if the resource should automatically trust Solution Add-On certificate",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Description: "Parent RDE state",
				Computed:    true,
			},
		},
	}
}

func resourceVcdSolutionAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	createCfg := govcd.SolutionAddOnConfig{
		IsoFilePath:          d.Get("addon_path").(string),
		User:                 "administrator",
		CatalogItemId:        d.Get("catalog_item_id").(string),
		AutoTrustCertificate: d.Get("trust_certificate").(bool),
	}
	addon, err := vcdClient.CreateSolutionAddOn(createCfg)

	if err != nil {
		return diag.Errorf("error creating Solution Add-On: %s", err)
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
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	dSet(d, "rde_state", slz.DefinedEntity.DefinedEntity.State)
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
		return diag.Errorf("error deleting Solution Add-On RDE: %s", err)
	}

	return nil
}

func resourceVcdSolutionAddonImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	if strings.Contains(d.Id(), "list@") {
		addOnList, err := listSolutionAddons(vcdClient)
		if err != nil {
			return nil, fmt.Errorf("error listing Solution Add-Ons: %s", err)
		}

		return nil, fmt.Errorf("resource was not imported! \n%s", addOnList)
	}

	log.Printf("[DEBUG] importing vcd_solution_add_on resource with provided id %s", d.Id())

	if strings.HasPrefix(d.Id(), "urn:vcloud:entity:") { // Import by id
		addOnById, err := vcdClient.GetSolutionAddonById(d.Id())
		if err != nil {
			addOnTable, err2 := listSolutionAddons(vcdClient)
			if err2 != nil {
				return nil, fmt.Errorf("error finding Solution Add-On by ID '%s' and couldn't retrieve list: %s", d.Id(), err2)
			}

			return nil, fmt.Errorf("error finding Solution Add-On by ID '%s': %s\n Available Add-Ons:\n %s", d.Id(), err, addOnTable)
		}

		d.SetId(addOnById.Id())
	} else {
		addOnByName, err := vcdClient.GetSolutionAddonByName(d.Id())
		if err != nil {
			addOnTable, err2 := listSolutionAddons(vcdClient)
			if err2 != nil {
				return nil, fmt.Errorf("error finding Solution Add-On by ID '%s' and couldn't retrieve list: %s", d.Id(), err2)
			}
			return nil, fmt.Errorf("error finding Solution Add-On by ID '%s': %s\n Available Add-Ons:\n %s", d.Id(), err, addOnTable)
		}

		d.SetId(addOnByName.Id())
	}

	return []*schema.ResourceData{d}, nil

}

func listSolutionAddons(vcdClient *VCDClient) (string, error) {
	buf := new(bytes.Buffer)
	writer := tabwriter.NewWriter(buf, 0, 8, 1, '\t', tabwriter.AlignRight)

	_, err := fmt.Fprintln(writer, "No\tID\tName\tStatus\tExtension Name\tVersion")
	if err != nil {
		return "", fmt.Errorf("error writing to buffer: %s", err)
	}
	_, err = fmt.Fprintln(writer, "--\t--\t-------\t------\t------\t------")
	if err != nil {
		return "", fmt.Errorf("error writing to buffer: %s", err)
	}

	addOns, err := vcdClient.GetAllSolutionAddons(nil)
	if err != nil {
		return "", fmt.Errorf("error retrieving all Solution Add-Ons")
	}

	for index, addon := range addOns {
		_, err = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\t%s\n", index+1,
			addon.Id(),
			addon.DefinedEntity.DefinedEntity.Name,
			addon.SolutionEntity.Status,
			addon.SolutionEntity.Manifest["name"].(string),
			addon.SolutionEntity.Manifest["version"].(string),
		)
		if err != nil {
			return "", fmt.Errorf("error writing to buffer: %s", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return "", fmt.Errorf("error flusher buffer: %s", err)
	}

	return buf.String(), nil
}
