package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxtNetworkContextProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtNetworkContextProfileRead,

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Context ID can be one of VDC, VDC Group, or NSX-T Manager ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "SYSTEM",
				Description:  "'SYSTEM', 'PROVIDER', 'TENANT'",
				ValidateFunc: validation.StringInSlice([]string{"SYSTEM", "PROVIDER", "TENANT"}, false),
			},
		},
	}
}

func datasourceVcdNsxtNetworkContextProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	name := d.Get("name").(string)
	scope := d.Get("scope").(string)
	contextId := d.Get("context_id").(string)

	nsxtNetworkContextProfile, err := govcd.GetNetworkContextProfilesByNameScopeAndContext(&vcdClient.Client, name, scope, contextId)
	if err != nil {
		return diag.Errorf("[Network Context profile DS Read] error finding Network Context Profile with name '%s', scope '%s' and context_id '%s': %s",
			name, scope, contextId, err)
	}

	d.SetId(nsxtNetworkContextProfile.ID)

	return nil
}
