package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/util"
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

			"logging": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_LOGGING", false),
				Description: "If set, it will enable logging of API requests and responses",
			},

			"logging_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_LOGGING_FILE", "go-vcloud-director.log"),
				Description: "Defines the full name of the logging file for API calls (requires 'logging')",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"vcd_network":            resourceVcdNetwork(),          // 1.0 DEPRECATED: replaced by vcd_network_routed
			"vcd_network_routed":     resourceVcdNetworkRouted(),    // 2.0
			"vcd_network_direct":     resourceVcdNetworkDirect(),    // 2.0
			"vcd_network_isolated":   resourceVcdNetworkIsolated(),  // 2.0
			"vcd_vapp_network":       resourceVcdVappNetwork(),      // 2.1
			"vcd_vapp":               resourceVcdVApp(),             // 1.0
			"vcd_firewall_rules":     resourceVcdFirewallRules(),    // 1.0
			"vcd_dnat":               resourceVcdDNAT(),             // 1.0
			"vcd_snat":               resourceVcdSNAT(),             // 1.0
			"vcd_edgegateway":        resourceVcdEdgeGateway(),      // 2.4
			"vcd_edgegateway_vpn":    resourceVcdEdgeGatewayVpn(),   // 1.0
			"vcd_vapp_vm":            resourceVcdVAppVm(),           // 1.0
			"vcd_org":                resourceOrg(),                 // 2.0
			"vcd_org_vdc":            resourceVcdOrgVdc(),           // 2.2
			"vcd_catalog":            resourceVcdCatalog(),          // 2.0
			"vcd_catalog_item":       resourceVcdCatalogItem(),      // 2.0
			"vcd_catalog_media":      resourceVcdCatalogMedia(),     // 2.0
			"vcd_inserted_media":     resourceVcdInsertedMedia(),    // 2.1
			"vcd_independent_disk":   resourceVcdIndependentDisk(),  // 2.1
			"vcd_external_network":   resourceVcdExternalNetwork(),  // 2.2
			"vcd_lb_service_monitor": resourceVcdLbServiceMonitor(), // 2.4
			"vcd_lb_server_pool":     resourceVcdLBServerPool(),     // 2.4
			"vcd_lb_app_profile":     resourceVcdLBAppProfile(),     // 2.4
		},

		DataSourcesMap: map[string]*schema.Resource{
			"vcd_lb_service_monitor": datasourceVcdLbServiceMonitor(), // 2.4
			"vcd_lb_server_pool":     datasourceVcdLbServerPool(),     // 2.4
			"vcd_lb_app_profile":     datasourceVcdLBAppProfile(),     // 2.4
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
	// If the provider includes logging directives,
	// it will activate logging from upstream go-vcloud-director
	logging := d.Get("logging").(bool)
	// Logging is disabled by default.
	// If enabled, we set the log file name and invoke the upstream logging set-up
	if logging {
		loggingFile := d.Get("logging_file").(string)
		if loggingFile != "" {
			util.EnableLogging = true
			util.ApiLogFileName = loggingFile
			util.InitLogging()
		}
	}

	return config.Client()
}
