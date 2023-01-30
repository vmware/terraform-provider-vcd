package vcd

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Provisional implementation - only for data collection

func datasourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvDistributedFirewallRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"rules": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func datasourceVcdNsxvDistributedFirewallRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall DS Read] error retrieving Org: %s", err)
	}

	vdcId := d.Get("vdc_id").(string)
	vdc, err := org.GetVDCById(vdcId, false)
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving VDC: %s", err)
	}

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	enabled, err := dfw.IsEnabled()

	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall state: %s", err)
	}
	if !enabled {
		return diag.Errorf("VDC '%s' does not have distributed firewall enabled", vdc.Vdc.Name)
	}
	util.Logger.Println("[NSXV DFW START]")
	configuration, err := dfw.GetConfiguration()
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall Rules: %s", err)
	}
	util.Logger.Printf("%# v\n", pretty.Formatter(configuration))
	util.Logger.Println("[NSXV DFW END]")
	confText, err := json.MarshalIndent(configuration, " ", " ")
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error encoding configuration into JSON: %s", err)
	}
	dSet(d, "rules", string(confText))
	d.SetId(vdc.Vdc.ID)

	return nil
}
