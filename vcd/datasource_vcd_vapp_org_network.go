package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdVappOrgNetwork() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVappOrgNetworkRead,
		Schema: map[string]*schema.Schema{
			"vapp_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
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
			"org_network": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization network name to which vApp network connected to",
			},
			"is_fenced": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp",
			},
			"retain_ip_mac_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default - false",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "firewall service enabled or disabled. Default - true",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NAT service enabled or disabled. Default - true",
			},
		},
	}
}

func datasourceVappOrgNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return genericVappOrgNetworkRead(d, meta, "datasource")
}
