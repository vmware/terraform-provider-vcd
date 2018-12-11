package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_USER", nil),
				Description: "The user name for VCD API operations.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_PASSWORD", nil),
				Description: "The user password for VCD API operations.",
			},

			"sysorg": &schema.Schema{
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_SYS_ORG", nil),
				Description: "The VCD Org for user authentication",
			},

			"org": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_ORG", nil),
				Description: "The VCD Org for API operations",
			},

			"vdc": &schema.Schema{
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_VDC", nil),
				Description: "The VDC for API operations",
			},

			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_URL", nil),
				Description: "The VCD url for VCD API operations.",
			},

			"max_retry_timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_MAX_RETRY_TIMEOUT", 60),
				Description: "Max num seconds to wait for successful response when operating on resources within vCloud (defaults to 60)",
			},

			"allow_unverified_ssl": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, VCDClient will permit unverifiable SSL certificates.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"vcd_network":          resourceVcdNetwork(), // DEPRECATED: replaced by vcd_network_routed
			"vcd_network_routed":   resourceVcdNetworkRouted(),
			"vcd_network_direct":   resourceVcdNetworkDirect(),
			"vcd_network_isolated": resourceVcdNetworkIsolated(),
			"vcd_vapp":             resourceVcdVApp(),
			"vcd_firewall_rules":   resourceVcdFirewallRules(),
			"vcd_dnat":             resourceVcdDNAT(),
			"vcd_snat":             resourceVcdSNAT(),
			"vcd_edgegateway_vpn":  resourceVcdEdgeGatewayVpn(),
			"vcd_vapp_vm":          resourceVcdVAppVm(),
			"vcd_org":              resourceOrg(),
			"vcd_catalog":          resourceVcdCatalog(),
			"vcd_catalog_item":     resourceVcdCatalogItem(),
			"vcd_catalog_media":    resourceVcdCatalogMedia(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	maxRetryTimeout := d.Get("max_retry_timeout").(int)

	// If sysOrg is defined, we use it for authentication.
	// Otherwise, we use the default org defined for regular usage
	connectOrg := d.Get("sysorg").(string)
	if connectOrg == "" {
		connectOrg = d.Get("org").(string)
	}
	config := Config{
		User:            d.Get("user").(string),
		Password:        d.Get("password").(string),
		SysOrg:          connectOrg,            // Connection org
		Org:             d.Get("org").(string), // Default org for operations
		Vdc:             d.Get("vdc").(string), // Default vdc
		Href:            d.Get("url").(string),
		MaxRetryTimeout: maxRetryTimeout,
		InsecureFlag:    d.Get("allow_unverified_ssl").(bool),
	}

	return config.Client()
}
