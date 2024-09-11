package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdAlbVirtualServiceRespRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceRespRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceRespRulesRead,
		// Update is the same as create and it does not have any additional details like rule IDs
		// which are important for updates in some cases.
		UpdateContext: resourceVcdAlbVirtualServiceRespRulesCreate,
		DeleteContext: resourceVcdAlbVirtualServiceRespRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceImport,
		},

		Schema: map[string]*schema.Schema{
			"virtual_service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Virtual Service ID",
			},
			"rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        nsxtAlbVirtualServiceRespRule,
				Description: "A single HTTP Response Rule",
			},
		},
	}
}

var nsxtAlbVirtualServiceRespRule = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the rule",
		},
		"active": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Defines if the rule is active or not",
		},
		"logging": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Defines whether to enable logging with headers on rule match or not",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			MaxItems:    1,
			Required:    true,
			Description: "Rule matching Criteria",
			Elem:        nsxtAlbVirtualServiceRespRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			MaxItems:    1,
			Required:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        nsxtAlbVirtualServiceRespRuleActions,
		},
	},
}

var nsxtAlbVirtualServiceRespRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.",
					},
					"ip_addresses": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"ports": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "A set of TCP ports. Allowed values are 1-65535",
						Elem: &schema.Schema{
							Type: schema.TypeInt,
						},
					},
				},
			},
		},
		"protocol_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
			Description:  "Protocol to match - 'HTTP' or 'HTTPS'",
		},
		"http_methods": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"methods": {
						Type:     schema.TypeSet,
						Optional: true,
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
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Criteria to use for matching the path in the HTTP request URI. Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH",
					},
					"paths": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			Optional:    true,
			Description: "HTTP request query strings to match",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"request_headers": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "A set of rules for matching request headers",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the HTTP header whose value is to be matched. Must be non-blank and fewer than 10240 characters",
					},
					"values": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			MaxItems:    1,
			Optional:    true,
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

		//
		"location_header": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Criteria to use for matching location header. Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH",
					},
					"values": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			Optional:    true,
			Description: "A set of criteria to match response headers",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the HTTP header whose value is to be matched",
					},
					"values": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			MaxItems:    1,
			Optional:    true,
			Description: "HTTP Status code to match",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.",
					},
					"http_status_code": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Enter a http status code or range",
					},
				},
			},
		},
	},
}

var nsxtAlbVirtualServiceRespRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"rewrite_location_header": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Rewrite location header",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "HTTP or HTTPS protocol",
						ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
					},
					"port": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Port to which redirect the request",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Host to which redirect the request",
					},
					"path": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Port to which redirect the request",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Path to which redirect the request",
					},
				},
			},
		},
		"modify_header": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"action": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "One of the following HTTP header actions. Options - ADD, REMOVE, REPLACE",
						ValidateFunc: validation.StringInSlice([]string{"ADD", "REMOVE", "REPLACE"}, false),
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "HTTP header name",
					},
					"value": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "HTTP header value",
					},
				},
			},
		},
	},
}

func resourceVcdAlbVirtualServiceRespRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	cfg, err := getEdgeVirtualServiceHttpResponseRuleType(d)
	if err != nil {
		return diag.Errorf("error getting HTTP Response Rule type: %s", err)
	}

	_, err = albVirtualService.UpdateHttpResponseRules(cfg)
	if err != nil {
		return diag.Errorf("error creating HTTP Response Rules: %s", err)
	}

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return resourceVcdAlbVirtualServiceRespRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceRespRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceRespRulesRead(ctx, d, meta, "resource")
}

func genericVcdAlbVirtualServiceRespRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Get("virtual_service_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) && origin == "resource" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	rules, err := albVirtualService.GetAllHttpResponseRules(nil)
	if err != nil {
		return diag.Errorf("could not retrieve HTTP Response Rules: %s", err)
	}

	dSet(d, "virtual_service_id", albVirtualService.NsxtAlbVirtualService.ID)
	err = setEdgeVirtualServiceHttpResponsetRuleData(d, rules)
	if err != nil {
		return diag.Errorf("error storing HTTP Response Rule: %s", err)
	}

	return nil
}

func resourceVcdAlbVirtualServiceRespRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	_, err = albVirtualService.UpdateHttpResponseRules(&types.EdgeVirtualServiceHttpResponseRules{})
	if err != nil {
		return diag.Errorf("error creating HTTP Response Rules: %s", err)
	}

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return nil
}

func getEdgeVirtualServiceHttpResponseRuleType(d *schema.ResourceData) (*types.EdgeVirtualServiceHttpResponseRules, error) {

	rules := d.Get("rule").([]interface{})
	rulesType := make([]types.EdgeVirtualServiceHttpResponseRule, len(rules))

	for ruleIndex, rule := range rules {
		ruleInterface := rule.(map[string]interface{})

		rulesType[ruleIndex].Name = ruleInterface["name"].(string)
		rulesType[ruleIndex].Active = ruleInterface["active"].(bool)
		rulesType[ruleIndex].Logging = ruleInterface["logging"].(bool)
		rulesType[ruleIndex].MatchCriteria = getResponseMatchCriteriaType(ruleInterface["match_criteria"].(*schema.Set))
		rulesType[ruleIndex].HeaderActions, rulesType[ruleIndex].RewriteLocationHeaderAction = getRespActionsType(ruleInterface["actions"].(*schema.Set))

	}

	structure := &types.EdgeVirtualServiceHttpResponseRules{
		Values: rulesType,
	}
	return structure, nil
}

func getResponseMatchCriteriaType(matchCriteria *schema.Set) types.EdgeVirtualServiceHttpResponseRuleMatchCriteria {
	if matchCriteria.Len() == 0 {
		return types.EdgeVirtualServiceHttpResponseRuleMatchCriteria{}
	}
	schemaSet := matchCriteria.List()

	allCriteria := schemaSet[0].(map[string]interface{})
	criteria := types.EdgeVirtualServiceHttpResponseRuleMatchCriteria{}

	clientIpAddressCriteria := allCriteria["client_ip_address"].([]interface{})
	if len(clientIpAddressCriteria) > 0 {
		clientIpAddressCriteriaMap := clientIpAddressCriteria[0].(map[string]interface{})
		criteria.ClientIPMatch = &types.EdgeVirtualServiceHttpRequestRuleClientIPMatch{
			MatchCriteria: clientIpAddressCriteriaMap["criteria"].(string),
			Addresses:     convertSchemaSetToSliceOfStrings(clientIpAddressCriteriaMap["ip_addresses"].(*schema.Set)),
		}
	}

	servicePortsCriteria := allCriteria["service_ports"].([]interface{})
	if len(servicePortsCriteria) > 0 {
		servicePortsCriteriaMap := servicePortsCriteria[0].(map[string]interface{})
		criteria.ServicePortMatch = &types.EdgeVirtualServiceHttpRequestRuleServicePortMatch{
			MatchCriteria: servicePortsCriteriaMap["criteria"].(string),
			Ports:         convertSchemaSetToSliceOfInts(servicePortsCriteriaMap["ports"].(*schema.Set)),
		}
	}

	protocolTypeCriteria := allCriteria["protocol_type"].(string)
	if protocolTypeCriteria != "" {
		criteria.Protocol = protocolTypeCriteria
	}

	httpMethodCriteria := allCriteria["http_methods"].([]interface{})
	if len(httpMethodCriteria) > 0 {
		httpMethodCriteriaMap := httpMethodCriteria[0].(map[string]interface{})
		criteria.MethodMatch = &types.EdgeVirtualServiceHttpRequestRuleMethodMatch{
			MatchCriteria: httpMethodCriteriaMap["criteria"].(string),
			Methods:       convertSchemaSetToSliceOfStrings(httpMethodCriteriaMap["methods"].(*schema.Set)),
		}
	}

	pathCriteria := allCriteria["path"].([]interface{})
	if len(pathCriteria) > 0 {
		pathCriteriaMap := pathCriteria[0].(map[string]interface{})
		criteria.PathMatch = &types.EdgeVirtualServiceHttpRequestRulePathMatch{
			MatchCriteria: pathCriteriaMap["criteria"].(string),
			MatchStrings:  convertSchemaSetToSliceOfStrings(pathCriteriaMap["paths"].(*schema.Set)),
		}
	}

	queryCriteria := allCriteria["query"].(*schema.Set)
	if queryCriteria.Len() > 0 {
		criteria.QueryMatch = convertSchemaSetToSliceOfStrings(queryCriteria)
	}

	requestHeaderCriteria := allCriteria["request_headers"].(*schema.Set)
	if requestHeaderCriteria.Len() > 0 {
		newHeaderCriteria := make([]types.EdgeVirtualServiceHttpRequestRuleHeaderMatch, requestHeaderCriteria.Len())
		for requestHeaderIndex, requestHeader := range requestHeaderCriteria.List() {
			requestHeaderMap := requestHeader.(map[string]interface{})

			newHeaderCriteria[requestHeaderIndex] = types.EdgeVirtualServiceHttpRequestRuleHeaderMatch{
				MatchCriteria: requestHeaderMap["criteria"].(string),
				Key:           requestHeaderMap["name"].(string),
				Value:         convertSchemaSetToSliceOfStrings(requestHeaderMap["values"].(*schema.Set)),
			}
		}
		criteria.RequestHeaderMatch = newHeaderCriteria
	}

	cookieCriteria := allCriteria["cookie"].([]interface{})
	if len(cookieCriteria) > 0 {
		cookieCriteriaMap := cookieCriteria[0].(map[string]interface{})
		criteria.CookieMatch = &types.EdgeVirtualServiceHttpRequestRuleCookieMatch{
			MatchCriteria: cookieCriteriaMap["criteria"].(string),
			Key:           cookieCriteriaMap["name"].(string),
			Value:         cookieCriteriaMap["value"].(string),
		}

	}

	locationHeaderCriteria := allCriteria["location_header"].([]interface{})
	if len(locationHeaderCriteria) > 0 {
		locationHeaderCriteriaMap := locationHeaderCriteria[0].(map[string]interface{})
		criteria.LocationHeaderMatch = &types.EdgeVirtualServiceHttpResponseLocationHeaderMatch{
			MatchCriteria: locationHeaderCriteriaMap["criteria"].(string),
			Value:         convertSchemaSetToSliceOfStrings(locationHeaderCriteriaMap["values"].(*schema.Set)),
		}
	}

	responseHeaderCriteria := allCriteria["response_headers"].(*schema.Set)
	if responseHeaderCriteria.Len() > 0 {
		newHeaderCriteria := make([]types.EdgeVirtualServiceHttpRequestRuleHeaderMatch, responseHeaderCriteria.Len())
		for responseHeaderIndex, responseHeader := range responseHeaderCriteria.List() {
			responseHeaderMap := responseHeader.(map[string]interface{})

			newHeaderCriteria[responseHeaderIndex] = types.EdgeVirtualServiceHttpRequestRuleHeaderMatch{
				MatchCriteria: responseHeaderMap["criteria"].(string),
				Key:           responseHeaderMap["name"].(string),
				Value:         convertSchemaSetToSliceOfStrings(responseHeaderMap["values"].(*schema.Set)),
			}
		}
		criteria.ResponseHeaderMatch = newHeaderCriteria
	}

	statusCodeCriteria := allCriteria["status_code"].([]interface{})
	if len(statusCodeCriteria) > 0 {
		statusCodeCriteriaMap := statusCodeCriteria[0].(map[string]interface{})
		criteria.StatusCodeMatch = &types.EdgeVirtualServiceHttpRuleStatusCodeMatch{
			MatchCriteria: statusCodeCriteriaMap["criteria"].(string),
			StatusCodes:   []string{statusCodeCriteriaMap["http_status_code"].(string)},
		}
	}

	return criteria
}

func getRespActionsType(actions *schema.Set) ([]*types.EdgeVirtualServiceHttpRequestRuleHeaderActions, *types.EdgeVirtualServiceHttpRespRuleRewriteLocationHeaderAction) {
	if actions.Len() == 0 {
		return nil, nil
	}
	schemaSet := actions.List()
	actionsIf := schemaSet[0].(map[string]interface{})

	modifyHeaderStructure := actionsIf["modify_header"].(*schema.Set)
	var mod []*types.EdgeVirtualServiceHttpRequestRuleHeaderActions

	rewriteUrlStructure := actionsIf["rewrite_location_header"].([]interface{})
	var rew *types.EdgeVirtualServiceHttpRespRuleRewriteLocationHeaderAction

	// Process any header rewrite cases, if specified
	if modifyHeaderStructure.Len() > 0 {
		newModifyHeaderStructure := make([]*types.EdgeVirtualServiceHttpRequestRuleHeaderActions, modifyHeaderStructure.Len())
		for headerIndex, header := range modifyHeaderStructure.List() {
			headerMap := header.(map[string]interface{})

			newModifyHeaderStructure[headerIndex] = &types.EdgeVirtualServiceHttpRequestRuleHeaderActions{
				Action: headerMap["action"].(string),
				Name:   headerMap["name"].(string),
				Value:  headerMap["value"].(string),
			}
		}
		mod = newModifyHeaderStructure
	}

	// Process any redirection cases, if specified
	if len(rewriteUrlStructure) > 0 {
		rewriteUrlStructure := rewriteUrlStructure[0].(map[string]interface{})
		rew = &types.EdgeVirtualServiceHttpRespRuleRewriteLocationHeaderAction{}
		rew.Protocol = rewriteUrlStructure["protocol"].(string)
		rew.Host = rewriteUrlStructure["host"].(string)
		rew.Port = rewriteUrlStructure["port"].(int)
		rew.Path = rewriteUrlStructure["path"].(string)
		rew.KeepQuery = rewriteUrlStructure["keep_query"].(bool)
	}

	return mod, rew
}

func setEdgeVirtualServiceHttpResponsetRuleData(d *schema.ResourceData, rules []*types.EdgeVirtualServiceHttpResponseRule) error {
	allRules := make([]interface{}, len(rules))

	for ruleIndex, rule := range rules {

		singleRule := make(map[string]interface{})

		singleRule["name"] = rule.Name
		singleRule["active"] = rule.Active
		singleRule["logging"] = rule.Logging

		////////// match_criteria block

		matchCriteria := make([]interface{}, 1)
		matchCriteriaMap := make(map[string]interface{})

		// "client_ip_address"
		ipAddress := make([]interface{}, 0)
		if rule.MatchCriteria.ClientIPMatch != nil {
			singleIpAddress := make(map[string]interface{})
			singleIpAddress["criteria"] = rule.MatchCriteria.ClientIPMatch.MatchCriteria
			singleIpAddress["ip_addresses"] = convertStringsToTypeSet(rule.MatchCriteria.ClientIPMatch.Addresses)
			ipAddress = append(ipAddress, singleIpAddress)
		}
		matchCriteriaMap["client_ip_address"] = ipAddress

		// "service_ports"
		servicePorts := make([]interface{}, 0)
		if rule.MatchCriteria.ServicePortMatch != nil {
			singleServicePorts := make(map[string]interface{})
			singleServicePorts["criteria"] = rule.MatchCriteria.ServicePortMatch.MatchCriteria
			singleServicePorts["ports"] = convertIntsToTypeSet(rule.MatchCriteria.ServicePortMatch.Ports)
			servicePorts = append(servicePorts, singleServicePorts)
		}
		matchCriteriaMap["service_ports"] = servicePorts

		// "protocol_type"
		matchCriteriaMap["protocol_type"] = rule.MatchCriteria.Protocol

		// "http_methods"
		httpMethod := make([]interface{}, 0)
		if rule.MatchCriteria.MethodMatch != nil {
			singleHttpMethod := make(map[string]interface{})
			singleHttpMethod["criteria"] = rule.MatchCriteria.MethodMatch.MatchCriteria
			singleHttpMethod["methods"] = convertStringsToTypeSet(rule.MatchCriteria.MethodMatch.Methods)
			httpMethod = append(httpMethod, singleHttpMethod)
		}
		matchCriteriaMap["http_methods"] = httpMethod

		// "path"
		path := make([]interface{}, 0)
		if rule.MatchCriteria.PathMatch != nil {
			singlePath := make(map[string]interface{})
			singlePath["criteria"] = rule.MatchCriteria.PathMatch.MatchCriteria
			singlePath["paths"] = convertStringsToTypeSet(rule.MatchCriteria.PathMatch.MatchStrings)
			path = append(path, singlePath)
		}
		matchCriteriaMap["path"] = path

		// "query"
		matchCriteriaMap["query"] = convertStringsToTypeSet(rule.MatchCriteria.QueryMatch)

		// "request_headers"
		requestHeaders := make([]interface{}, len(rule.MatchCriteria.RequestHeaderMatch))
		for i, h := range rule.MatchCriteria.RequestHeaderMatch {
			singleHedear := make(map[string]interface{})
			singleHedear["criteria"] = h.MatchCriteria
			singleHedear["name"] = h.Key
			singleHedear["values"] = convertStringsToTypeSet(h.Value)

			requestHeaders[i] = singleHedear
		}
		matchCriteriaMap["request_headers"] = requestHeaders

		// "cookie"
		cookie := make([]interface{}, 0)
		if rule.MatchCriteria.CookieMatch != nil {
			singleCookie := make(map[string]interface{})
			singleCookie["criteria"] = rule.MatchCriteria.CookieMatch.MatchCriteria
			singleCookie["name"] = rule.MatchCriteria.CookieMatch.Key
			singleCookie["value"] = rule.MatchCriteria.CookieMatch.Value
			cookie = append(cookie, singleCookie)
		}
		matchCriteriaMap["cookie"] = cookie

		// "response_headers"
		responseHeaders := make([]interface{}, len(rule.MatchCriteria.ResponseHeaderMatch))
		for i, h := range rule.MatchCriteria.ResponseHeaderMatch {
			singleHedear := make(map[string]interface{})
			singleHedear["criteria"] = h.MatchCriteria
			singleHedear["name"] = h.Key
			singleHedear["values"] = convertStringsToTypeSet(h.Value)

			responseHeaders[i] = singleHedear
		}
		matchCriteriaMap["response_headers"] = responseHeaders

		// "location_header"
		locationHeader := make([]interface{}, 0)
		if rule.MatchCriteria.LocationHeaderMatch != nil {
			singlePath := make(map[string]interface{})
			singlePath["criteria"] = rule.MatchCriteria.LocationHeaderMatch.MatchCriteria
			singlePath["values"] = convertStringsToTypeSet(rule.MatchCriteria.LocationHeaderMatch.Value)
			locationHeader = append(locationHeader, singlePath)
		}
		matchCriteriaMap["location_header"] = locationHeader

		// "status_code"
		statusCode := make([]interface{}, 0)
		if rule.MatchCriteria.StatusCodeMatch != nil {
			singleStatusCode := make(map[string]interface{})
			singleStatusCode["criteria"] = rule.MatchCriteria.StatusCodeMatch.MatchCriteria
			singleStatusCode["http_status_code"] = rule.MatchCriteria.StatusCodeMatch.StatusCodes[0]
			statusCode = append(statusCode, singleStatusCode)
		}
		matchCriteriaMap["status_code"] = statusCode

		// Pack root entry
		matchCriteria[0] = matchCriteriaMap
		singleRule["match_criteria"] = matchCriteria

		////////// EOF match_criteria

		//// 'actions'

		actions := make([]interface{}, 1)
		actionsMap := make(map[string]interface{})

		// 'rewrite_location_header'
		rewriteLocationHeader := make([]interface{}, 0)
		if rule.RewriteLocationHeaderAction != nil {
			singleRedirect := make(map[string]interface{})
			singleRedirect["protocol"] = rule.RewriteLocationHeaderAction.Protocol
			singleRedirect["port"] = rule.RewriteLocationHeaderAction.Port
			singleRedirect["host"] = rule.RewriteLocationHeaderAction.Host
			singleRedirect["path"] = rule.RewriteLocationHeaderAction.Path
			singleRedirect["keep_query"] = rule.RewriteLocationHeaderAction.KeepQuery

			rewriteLocationHeader = append(rewriteLocationHeader, singleRedirect)
		}
		actionsMap["rewrite_location_header"] = rewriteLocationHeader

		// 'modify_header'

		modifyHeader := make([]interface{}, 0)
		if rule.HeaderActions != nil {
			for _, mh := range rule.HeaderActions {
				singleModifyHeader := make(map[string]interface{})
				singleModifyHeader["action"] = mh.Action
				singleModifyHeader["name"] = mh.Name
				singleModifyHeader["value"] = mh.Value

				modifyHeader = append(modifyHeader, singleModifyHeader)
			}
		}
		actionsMap["modify_header"] = modifyHeader

		actions[0] = actionsMap
		singleRule["actions"] = actions

		//// EOF 'actions'

		allRules[ruleIndex] = singleRule
	}

	return d.Set("rule", allRules)
}
