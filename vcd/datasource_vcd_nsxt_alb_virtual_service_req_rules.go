package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbVirtualServiceReqRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbVirtualServiceReqRulesRead,

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
				Elem:        nsxtAlbVirtualServiceReqRule,
				Description: "A single HTTP Request Rule",
			},
		},
	}
}

var dsNsxtAlbVirtualServiceReqRule = &schema.Resource{
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
			Description: "Defines whether to enable logging with headers on rule match or not",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Rule matching Criteria",
			Elem:        dsNsxtAlbVsReqAndSecRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        nsxtAlbVirtualServiceReqRuleActions,
		},
	},
}

var dsNsxtAlbVsReqAndSecRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Client IP Address criteria",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.",
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
			Description: "Service Port criteria",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"ports": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "A set of TCP ports. Allowed values are 1-65535",
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
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"methods": {
						Type:     schema.TypeSet,
						Computed: true,
						// Not validating these options as it might not be finite list and API returns proper explanations
						Description: "HTTP methods to match. Options - GET, PUT, POST, DELETE, HEAD, OPTIONS, TRACE, CONNECT, PATCH, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK",
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
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criteria to use for matching the path in the HTTP request URI. Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH",
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
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the HTTP header whose value is to be matched. Must be non-blank and fewer than 10240 characters",
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
						Optional:    true,
						Description: "Criteria to use for matching cookies in the HTTP request. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the HTTP cookie whose value is to be matched",
					},
					"value": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "String values to match for an HTTP cookie",
					},
				},
			},
		},
	},
}

var dsNsxtAlbVsReqRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"redirect": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Redirect request",
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
						Description: "Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol",
					},
					"status_code": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "One of the redirect status codes - 301, 302, 307",
					},
					"host": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Host to which redirect the request. Default is the original host",
					},
					"path": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Path to which redirect the request. Default is the original path",
					},
				},
			},
		},
		"modify_header": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "A set of header modification rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"action": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "One of the following HTTP header actions. Options - ADD, REMOVE, REPLACE",
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
		"rewrite_url": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "URL rewrite rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_header": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Host to use for the rewritten URL",
					},
					"existing_path": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Path to use for the rewritten URL",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether or not to keep the existing query string when rewriting the URL",
					},
					"query": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Query string to use or append to the existing query string in the rewritten URL",
					},
				},
			},
		},
	},
}

func datasourceVcdAlbVirtualServiceReqRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceReqRulesRead(ctx, d, meta, "datasource")
}
