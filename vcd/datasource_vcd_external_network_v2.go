package vcd

import (
	"fmt"
	"log"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdExternalNetworkV2() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdExternalNetworkV2Read,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_scope": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP scopes for the network",
				Elem:        networkV2IpScope,
			},
			"vsphere_network": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcenter_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vCenter server ID",
						},
						"portgroup_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The portgroup ID",
						},
					},
				},
			},
			"nsxt_network": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nsxt_manager_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of NSX-T manager",
						},
						"nsxt_tier0_router_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of NSX-T Tier-0 router (for T0 gateway backed external network)",
						},
						"nsxt_segment_name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of NSX-T segment (for NSX-T segment backed external network)",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdExternalNetworkV2Read(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 data source read initiated")

	name := d.Get("name").(string)

	extNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, name)
	if err != nil {
		return fmt.Errorf("could not find external network V2 by name '%s': %s", name, err)
	}

	d.SetId(extNet.ExternalNetwork.ID)

	return setExternalNetworkV2Data(d, extNet.ExternalNetwork, vcdClient)
}
