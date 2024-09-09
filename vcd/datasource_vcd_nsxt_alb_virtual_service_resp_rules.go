package vcd

/*
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
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        dsNsxtAlbVirtualServiceReqRule,
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
			Description: "Defines is the rule is active or not",
		},
		"logging": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Defines whether to enable logging with headers on rule match or not",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Rule matching criterion",
			Elem:        dsNsxtAlbVirtualServiceReqRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        dsNsxtAlbVirtualServiceReqRuleActions,
		},
	},
}

var dsNsxtAlbVirtualServiceReqRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criterion to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.",
					},
					"ip_addresses": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "Enter IPv4 or IPv6 address, range or CIDR",
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
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criterion to use for port matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"ports": {
						Type:        schema.TypeSet,
						Computed:    true,
						Description: "Listening TCP ports. Allowed values are 1-65535",
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
		"http_method": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criterion to use for matching the method in the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"method": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP methods to match. Options - GET, PUT, POST, DELETE, HEAD, OPTIONS, TRACE, CONNECT, PATCH, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK",
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
						Description: "Criterion to use for matching the path in the HTTP request URI. Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH",
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
			Description: "HTTP request query strings in key=value format",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"request_headers": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criterion to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the HTTP header whose value is to be matched. Must be non-blank and fewer than 10240 characters",
					},
					"value": {

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
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Criterion to use for matching cookies in the HTTP request. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
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
	},
}

var dsNsxtAlbVirtualServiceReqRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"redirect": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "",
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
			Description: "",
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
			Description: "",
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

func datasourceVcdAlbVirtualServiceRespRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceRespRulesRead(ctx, d, meta, "datasource")
}
*/
