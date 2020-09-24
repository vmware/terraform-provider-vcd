package vcd

import (
	"fmt"
	"log"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdExternalNetworkV2() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdExternalNetworkV2Read,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_scope": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of IP scopes for the network",
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
							Description: "The vCenter server name",
						},
						"portgroup_id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the port group",
						},
					},
				},
			},
			"nsxt_network": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				MaxItems:    1,
				ForceNew:    true,
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
							Description: "ID of NSX-T Tier-0 router",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdExternalNetworkV2Read(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 read initiated")

	name := d.Get("name").(string)

	extNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, name)
	if err != nil {
		return fmt.Errorf("could not find external network by name '%s': %s", name, err)
	}

	d.SetId(extNet.ExternalNetwork.ID)

	return setExternalNetworkV2Data(d, extNet.ExternalNetwork)
}
