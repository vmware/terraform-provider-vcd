package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdVappOrgNetwork() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVappOrgNetworkRead,
		Schema: map[string]*schema.Schema{
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "vApp name",
			},
			"org_network_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization network name to which vApp network is connected to",
			},
			"is_fenced": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp",
			},
			"retain_ip_mac_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments.",
			},
		},
	}
}

func datasourceVappOrgNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return genericVappOrgNetworkRead(d, meta, "datasource")
}
