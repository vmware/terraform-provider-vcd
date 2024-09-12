package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbVirtualServiceSecRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbVirtualServiceSecRulesRead,

		Schema: map[string]*schema.Schema{
			"virtual_service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Virtual Service ID",
			},
			"rule": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        dsNsxtAlbVirtualServiceSecRule,
				Description: "A single HTTP Security Rule",
			},
		},
	}
}

var dsNsxtAlbVirtualServiceSecRule = &schema.Resource{
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
			Description: "Rule matching Criteria",
			// Match criteria are the same as for HTTP Request
			Elem: dsNsxtAlbVsReqAndSecRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        dsNsxtAlbVsSecRuleActions,
		},
	},
}

var dsNsxtAlbVsSecRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"redirect_to_https": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Port number that should be redirected to HTTPS",
		},
		"connections": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ALLOW or CLOSE connections",
		},
		"rate_limit": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Apply actions based on rate limits",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Maximum number of connections, requests or packets permitted each period. The count must be between 1 and 1000000000",
					},
					"period": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Time value in seconds to enforce rate count. The period must be between 1 and 1000000000",
					},
					"action_close_connection": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "True if the connection should be closed",
					},
					"action_redirect": {
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
					"action_local_response": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Send custom response",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"content": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Base64 encoded content",
								},
								"content_type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "MIME type for the content",
								},
								"status_code": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "HTTP Status code to send",
								},
							},
						},
					},
				},
			},
		},

		"send_response": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Send custom response",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"content": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Base64 encoded content",
					},
					"content_type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "MIME type for the content",
					},
					"status_code": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "HTTP Status code to send",
					},
				},
			},
		},
	},
}

func datasourceVcdAlbVirtualServiceSecRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceSecRulesRead(ctx, d, meta, "datasource")
}
