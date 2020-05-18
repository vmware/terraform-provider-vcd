package vcd

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func datasourceVcdEdgeGateway() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdEdgeGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "name of the edge gateway. (Optional when 'filter' is used)",
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"vdc": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"advanced": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the gateway uses advanced networking. (Enabled by default)",
			},
			"configuration": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Configuration of the vShield edge VM for this gateway. One of: compact, full ("Large"), full4 ("Quad Large"), x-large`,
			},
			"ha_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable high availability on this edge gateway",
			},
			"external_networks": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of external networks to be used by the edge gateway",
				Deprecated:  "Please use the more advanced 'external_network' block(s)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_gateway_network": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "Please use the more advanced 'external_network' block(s)",
				Description: "External network to be used as default gateway. Its name must be included in 'external_networks'. An empty value will skip the default gateway",
			},
			"default_external_network_ip": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of edge gateway interface which is used as default.",
			},
			"external_network_ips": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "List of IP addresses set on edge gateway external network interfaces",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"distributed_routing": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If advanced networking enabled, also enable distributed routing",
			},
			"lb_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable load balancing. (Disabled by default)",
			},
			"lb_acceleration_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable load balancer acceleration. (Disabled by default)",
			},
			"lb_logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable load balancer logging. (Disabled by default)",
			},
			"lb_loglevel": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "Log level. One of 'emergency', 'alert', 'critical', 'error', " +
					"'warning', 'notice', 'info', 'debug'. ('info' by default)",
			},
			"fw_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable firewall. Default 'true'",
			},
			"fw_default_rule_logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable logging for default rule. Default 'false'",
			},
			"fw_default_rule_action": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "'accept' or 'deny'. Default 'deny'",
			},
			"fips_mode_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable FIPS mode. FIPS mode turns on the cipher suites that comply with FIPS. (False by default)",
			},
			"use_default_route_for_dns_relay": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If true, default gateway will be used for the edge gateways' default routing and DNS forwarding.(False by default)",
			},
			"external_network": {
				Type:        schema.TypeSet,
				Description: "One or more blocks with external network information to be attached to this gateway's interface",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "External network name",
						},
						"enable_rate_limit": {
							Computed:    true,
							Type:        schema.TypeBool,
							Description: "Enable rate limiting",
						},
						"incoming_rate_limit": {
							Computed:    true,
							Type:        schema.TypeFloat,
							Description: "Incoming rate limit (Mbps)",
						},
						"outgoing_rate_limit": {
							Computed:    true,
							Type:        schema.TypeFloat,
							Description: "Outgoing rate limit (Mbps)",
						},
						"subnet": {
							Computed: true,
							Type:     schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gateway": {
										Computed:    true,
										Description: "Gateway address for a subnet",
										Type:        schema.TypeString,
									},
									"netmask": {
										Computed:    true,
										Description: "Netmask address for a subnet",
										Type:        schema.TypeString,
									},
									"ip_address": {
										Computed:    true,
										Type:        schema.TypeString,
										Description: "IP address on the edge gateway - will be auto-assigned if not defined",
									},
									"use_for_default_route": {
										Computed:    true,
										Type:        schema.TypeBool,
										Description: "Defines if this subnet should be used as default gateway for edge",
									},
									"suballocate_pool": {
										Type:        schema.TypeSet,
										Computed:    true,
										Description: "Define zero or more blocks to sub-allocate pools on the edge gateway",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start_address": {
													Computed: true,
													Type:     schema.TypeString,
												},
												"end_address": {
													Computed: true,
													Type:     schema.TypeString,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"filter": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving an edge gateway by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
					},
				},
			},
		},
	}
}

func datasourceVcdEdgeGatewayRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdEdgeGatewayRead(d, meta, "datasource")
}
