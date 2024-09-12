package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbVirtualServiceRespRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbVirtualServiceRespRulesRead,

		Schema: map[string]*schema.Schema{
			"virtual_service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Virtual Service ID",
			},
			"rule": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        dsNsxtAlbVirtualServiceRespRule,
				Description: "HTTP Response Rules",
			},
		},
	}
}

var dsNsxtAlbVirtualServiceRespRule = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the rule",
		},
		"active": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Defines if the rule is active or not",
		},
		"logging": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Defines whether logging with headers on rule match is enabled or not",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Rule matching Criteria",
			Elem:        dsNsxtAlbVirtualServiceRespRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        dsNsxtAlbVirtualServiceRespRuleActions,
		},
	},
}

var dsNsxtAlbVirtualServiceRespRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Criteria for matching client IP Address",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for IP address matching the HTTP request",
					},
					"ip_addresses": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "A set of IP addresses",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"service_ports": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Criteria for matching service ports",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for service ports matching the HTTP request",
					},
					"ports": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "A set of TCP ports",
						Elem: &schema.Schema{
							Type: schema.TypeInt,
						},
					},
				},
			},
		},
		"protocol_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Protocol to match - 'HTTP' or 'HTTPS'",
		},
		"http_methods": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Criteria to match HTTP methods",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for HTTP methods matching the request",
					},
					"methods": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "HTTP methods to match",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"path": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Criteria for matching request paths",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching the path in the HTTP request URI",
					},
					"paths": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "String values to match the path",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"query": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "HTTP request query strings to match",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"request_headers": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "A set of rules for matching request headers",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the HTTP header whose value is to be matched",
					},
					"values": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "String values to match for an HTTP header",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"cookie": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Rule for matching cookie",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching cookies in the HTTP request",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the HTTP cookie whose value is to be matched",
					},
					"value": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "String values to match for an HTTP cookie",
					},
				},
			},
		},
		// in addition to the same rules that are available for HTTP requests
		"location_header": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A matching criteria for Location header",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching location header",
					},
					"values": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "A set of values to match for criteria",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"response_headers": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "A set of criteria to match response headers",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the HTTP header whose value is to be matched",
					},
					"values": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "A set of values to match for an HTTP header",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"status_code": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "HTTP Status code to match",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for status code matching the HTTP request.",
					},
					"http_status_code": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP status code or range of this rule matching",
					},
				},
			},
		},
	},
}

var dsNsxtAlbVirtualServiceRespRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"rewrite_location_header": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Rewrite location header",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP or HTTPS protocol",
					},
					"port": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Port to which redirect the request",
					},
					"host": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Host to which redirect the request",
					},
					"path": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Port to which redirect the request",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Path to which redirect the request",
					},
				},
			},
		},
		"modify_header": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"action": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "One of the following HTTP header actions",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP header name",
					},
					"value": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP header value",
					},
				},
			},
		},
	},
}

func datasourceVcdAlbVirtualServiceRespRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceRespRulesRead(ctx, d, meta, "datasource")
}
