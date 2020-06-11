package vcd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// DataSources is a public function which allows to filter and access all defined data sources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func DataSources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcdSchemaFilter(globalDataSourceMap, nameRegexp, includeDeprecated)
}

// Resources is a public function which allows to filter and access all defined resources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func Resources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcdSchemaFilter(globalResourceMap, nameRegexp, includeDeprecated)
}

var globalDataSourceMap = map[string]*schema.Resource{
	"vcd_org":                datasourceVcdOrg(),              // 2.5
	"vcd_org_vdc":            datasourceVcdOrgVdc(),           // 2.5
	"vcd_catalog":            datasourceVcdCatalog(),          // 2.5
	"vcd_catalog_media":      datasourceVcdCatalogMedia(),     // 2.5
	"vcd_catalog_item":       datasourceVcdCatalogItem(),      // 2.5
	"vcd_edgegateway":        datasourceVcdEdgeGateway(),      // 2.5
	"vcd_external_network":   datasourceVcdExternalNetwork(),  // 2.5
	"vcd_independent_disk":   datasourceVcIndependentDisk(),   // 2.5
	"vcd_network_routed":     datasourceVcdNetworkRouted(),    // 2.5
	"vcd_network_direct":     datasourceVcdNetworkDirect(),    // 2.5
	"vcd_network_isolated":   datasourceVcdNetworkIsolated(),  // 2.5
	"vcd_vapp":               datasourceVcdVApp(),             // 2.5
	"vcd_vapp_vm":            datasourceVcdVAppVm(),           // 2.6
	"vcd_lb_service_monitor": datasourceVcdLbServiceMonitor(), // 2.4
	"vcd_lb_server_pool":     datasourceVcdLbServerPool(),     // 2.4
	"vcd_lb_app_profile":     datasourceVcdLBAppProfile(),     // 2.4
	"vcd_lb_app_rule":        datasourceVcdLBAppRule(),        // 2.4
	"vcd_lb_virtual_server":  datasourceVcdLbVirtualServer(),  // 2.4
	"vcd_nsxv_dnat":          datasourceVcdNsxvDnat(),         // 2.5
	"vcd_nsxv_snat":          datasourceVcdNsxvSnat(),         // 2.5
	"vcd_nsxv_firewall_rule": datasourceVcdNsxvFirewallRule(), // 2.5
	"vcd_nsxv_dhcp_relay":    datasourceVcdNsxvDhcpRelay(),    // 2.6
	"vcd_nsxv_ip_set":        datasourceVcdIpSet(),            // 2.6
	"vcd_vapp_network":       datasourceVcdVappNetwork(),      // 2.7
	"vcd_vapp_org_network":   datasourceVcdVappOrgNetwork(),   // 2.7
}

var globalResourceMap = map[string]*schema.Resource{
	"vcd_network":            resourceVcdNetwork(),          // 1.0 DEPRECATED: replaced by vcd_network_routed
	"vcd_network_routed":     resourceVcdNetworkRouted(),    // 2.0
	"vcd_network_direct":     resourceVcdNetworkDirect(),    // 2.0
	"vcd_network_isolated":   resourceVcdNetworkIsolated(),  // 2.0
	"vcd_vapp_network":       resourceVcdVappNetwork(),      // 2.1
	"vcd_vapp":               resourceVcdVApp(),             // 1.0
	"vcd_firewall_rules":     resourceVcdFirewallRules(),    // 1.0 DEPRECATED: Use only for non-advanced edge gateway. Replaced by vcd_nsxv_firewall_rule
	"vcd_dnat":               resourceVcdDNAT(),             // 1.0 DEPRECATED: Use only for non-advanced edge gateway. Replaced by vcd_nsxv_dnat
	"vcd_snat":               resourceVcdSNAT(),             // 1.0 DEPRECATED: Use only for non-advanced edge gateway. Replaced by vcd_nsxv_snat
	"vcd_edgegateway":        resourceVcdEdgeGateway(),      // 2.4
	"vcd_edgegateway_vpn":    resourceVcdEdgeGatewayVpn(),   // 1.0
	"vcd_vapp_vm":            resourceVcdVAppVm(),           // 1.0
	"vcd_org":                resourceOrg(),                 // 2.0
	"vcd_org_vdc":            resourceVcdOrgVdc(),           // 2.2
	"vcd_org_user":           resourceVcdOrgUser(),          // 2.4
	"vcd_catalog":            resourceVcdCatalog(),          // 2.0
	"vcd_catalog_item":       resourceVcdCatalogItem(),      // 2.0
	"vcd_catalog_media":      resourceVcdCatalogMedia(),     // 2.0
	"vcd_inserted_media":     resourceVcdInsertedMedia(),    // 2.1
	"vcd_independent_disk":   resourceVcdIndependentDisk(),  // 2.1
	"vcd_external_network":   resourceVcdExternalNetwork(),  // 2.2
	"vcd_lb_service_monitor": resourceVcdLbServiceMonitor(), // 2.4
	"vcd_lb_server_pool":     resourceVcdLBServerPool(),     // 2.4
	"vcd_lb_app_profile":     resourceVcdLBAppProfile(),     // 2.4
	"vcd_lb_app_rule":        resourceVcdLBAppRule(),        // 2.4
	"vcd_lb_virtual_server":  resourceVcdLBVirtualServer(),  // 2.4
	"vcd_nsxv_dnat":          resourceVcdNsxvDnat(),         // 2.5
	"vcd_nsxv_snat":          resourceVcdNsxvSnat(),         // 2.5
	"vcd_nsxv_firewall_rule": resourceVcdNsxvFirewallRule(), // 2.5
	"vcd_nsxv_dhcp_relay":    resourceVcdNsxvDhcpRelay(),    // 2.6
	"vcd_nsxv_ip_set":        resourceVcdIpSet(),            // 2.6
	"vcd_vm_internal_disk":   resourceVmInternalDisk(),      // 2.7
	"vcd_vapp_org_network":   resourceVcdVappOrgNetwork(),   // 2.7
}

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_USER", nil),
				Description: "The user name for VCD API operations.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_PASSWORD", nil),
				Description: "The user password for VCD API operations.",
			},
			"auth_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("VCD_AUTH_TYPE", "integrated"),
				Description:  "'integrated', 'saml_adfs', and 'token' are the only supported now. 'integrated' is default.",
				ValidateFunc: validation.StringInSlice([]string{"integrated", "saml_adfs", "token"}, false),
			},
			"saml_adfs_rpt_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_SAML_ADFS_RPT_ID", nil),
				Description: "Allows to specify custom Relaying Party Trust Identifier for auth_type=saml_adfs",
			},

			"token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_TOKEN", nil),
				Description: "The token used instead of username/password for VCD API operations.",
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
			"import_separator": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_IMPORT_SEPARATOR", "."),
				Description: "Defines the import separation string to be used with 'terraform import'",
			},
		},
		ResourcesMap:   globalResourceMap,
		DataSourcesMap: globalDataSourceMap,
		ConfigureFunc:  providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	maxRetryTimeout := d.Get("max_retry_timeout").(int)

	if err := validateProviderSchema(d); err != nil {
		return nil, fmt.Errorf("[provider validation] :%s", err)
	}

	// If sysOrg is defined, we use it for authentication.
	// Otherwise, we use the default org defined for regular usage
	connectOrg := d.Get("sysorg").(string)
	if connectOrg == "" {
		connectOrg = d.Get("org").(string)
	}

	config := Config{
		User:            d.Get("user").(string),
		Password:        d.Get("password").(string),
		Token:           d.Get("token").(string),
		SysOrg:          connectOrg,            // Connection org
		Org:             d.Get("org").(string), // Default org for operations
		Vdc:             d.Get("vdc").(string), // Default vdc
		Href:            d.Get("url").(string),
		MaxRetryTimeout: maxRetryTimeout,
		InsecureFlag:    d.Get("allow_unverified_ssl").(bool),
	}

	// auth_type dependent configuration
	authType := d.Get("auth_type").(string)
	switch authType {
	case "saml_adfs":
		config.UseSamlAdfs = true
		config.CustomAdfsRptId = d.Get("saml_adfs_rpt_id").(string)
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

	separator := os.Getenv("VCD_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	} else {
		ImportSeparator = d.Get("import_separator").(string)
	}
	return config.Client()
}

// vcdSchemaFilter is a function which allows to filters and export type 'map[string]*schema.Resource' which may hold
// Terraform's native resource or data source list
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func vcdSchemaFilter(schemaMap map[string]*schema.Resource, nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	var (
		err error
		re  *regexp.Regexp
	)
	filteredResources := make(map[string]*schema.Resource)

	// validate regex if it was provided
	if nameRegexp != "" {
		re, err = regexp.Compile(nameRegexp)
		if err != nil {
			return nil, fmt.Errorf("unable to compile regexp: %s", err)
		}
	}

	// copy the map with filtering out unwanted object
	for resourceName, schemaResource := range schemaMap {

		// Skip deprecated resources if it was requested so
		if !includeDeprecated && schemaResource.DeprecationMessage != "" {
			continue
		}
		// If regex was defined - try to filter based on it
		if re != nil {
			// if it does not match regex - skip it
			doesNotmatchRegex := !re.MatchString(resourceName)
			if doesNotmatchRegex {
				continue
			}

		}

		filteredResources[resourceName] = schemaResource
	}

	return filteredResources, nil
}

func validateProviderSchema(d *schema.ResourceData) error {

	// Validate org and sys org
	sysOrg := d.Get("sysorg").(string)
	org := d.Get("org").(string)
	if sysOrg == "" && org == "" {
		return fmt.Errorf(`both "org" and "sysorg" properties are empty`)
	}

	return nil
}
