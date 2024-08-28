package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtTier0Interface() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtTier0InterfaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Tier-0 router.",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of External network (Provider Gateway)",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Tier-0 assigned interface",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of 'EXTERNAL', 'SERVICE', 'LOOPBACK'",
			},
		},
	}
}

func datasourceNsxtTier0InterfaceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	externalNetworkId := d.Get("external_network_id").(string)
	name := d.Get("name").(string)

	t0Interface, err := vcdClient.GetTier0RouterInterfaceByName(externalNetworkId, name)
	if err != nil {
		return diag.Errorf("error retrieving Tier-0 router interface by name '%s': %s", name, err)
	}

	d.SetId(t0Interface.ID)
	dSet(d, "description", t0Interface.Description)
	dSet(d, "type", t0Interface.InterfaceType)

	return nil
}
