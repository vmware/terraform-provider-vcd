package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

func datasourceVcdNsxtManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtManagerRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T manager.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HREF of NSX-T manager.",
			},
		},
	}
}

func datasourceNsxtManagerRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	nsxtManagerName := d.Get("name").(string)

	nsxtManagers, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
	if err != nil {
		return diag.Errorf("could not find NSX-T manager by name '%s': %s", nsxtManagerName, err)
	}

	if len(nsxtManagers) == 0 {
		return diag.Errorf("%s found %d NSX-T managers with name '%s'",
			govcd.ErrorEntityNotFound, len(nsxtManagers), nsxtManagerName)
	}

	if len(nsxtManagers) > 1 {
		return diag.Errorf("found %d NSX-T managers with name '%s'", len(nsxtManagers), nsxtManagerName)
	}

	// We try to keep IDs clean
	id := extractUuid(nsxtManagers[0].HREF)
	urn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", id)
	if err != nil {
		return diag.Errorf("could not construct URN from id '%s': %s", id, err)
	}
	dSet(d, "name", nsxtManagers[0].Name)
	dSet(d, "href", nsxtManagers[0].HREF)
	d.SetId(urn)

	return nil
}
