package vcd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func resourceVcdAlbVirtualServiceReqRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceReqRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceReqRulesRead,
		// Update is the same as create and it does not have any additional details like rule IDs
		// which are important for updates in some cases.
		UpdateContext: resourceVcdAlbVirtualServiceReqRulesCreate,
		DeleteContext: resourceVcdAlbVirtualServiceReqRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceHttpPolicyImport,
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
				Required:    true,
				Elem:        nsxtAlbVirtualServiceReqRule,
				Description: "A single HTTP Request Rule",
			},
		},
	}
}

var nsxtAlbVirtualServiceReqRule = &schema.Resource{
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
			Description: "Defines whether to enable logging with headers on rule match or not",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			MaxItems:    1,
			Required:    true,
			Description: "Rule matching Criteria",
			Elem:        nsxtAlbVsReqAndSecRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			MaxItems:    1,
			Required:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        nsxtAlbVirtualServiceReqRuleActions,
		},
	},
}

var nsxtAlbVsReqAndSecRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Client IP Address criteria",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN.",
					},
					"ip_addresses": {
						Type:        schema.TypeSet,
						Required:    true,
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
			Description: "Service Port criteria",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"ports": {
						Type:        schema.TypeSet,
						Required:    true,
						Description: "A set of TCP ports. Allowed values are 1-65535",
						Elem: &schema.Schema{
							Type:             schema.TypeInt,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 65535)),
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
			Description: "HTTP methods that should be matched",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
						Description:  "Criteria to use for IP address matching the HTTP request. Options - IS_IN, IS_NOT_IN",
					},
					"methods": {
						Type:     schema.TypeSet,
						Required: true,
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
			Description: "Request path criteria",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:     schema.TypeString,
						Required: true,
						// Validation does return options as opposed to the cases where we only have IS_IN, IS_NOT_IN
						Description: "Criteria to use for matching the path in the HTTP request URI. Options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH",
					},
					"paths": {
						Type:        schema.TypeSet,
						Required:    true,
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
						Required:    true,
						Description: "Criteria to use for matching headers and cookies in the HTTP request amd response. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the HTTP header whose value is to be matched",
					},
					"values": {
						Type:        schema.TypeSet,
						Required:    true,
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
			Description: "Criteria for matching cookie",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Criteria to use for matching cookies in the HTTP request. Options - EXISTS, DOES_NOT_EXIST, BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL",
					},
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the HTTP cookie whose value is to be matched",
					},
					"value": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "String values to match for an HTTP cookie",
					},
				},
			},
		},
	},
}

var nsxtAlbVirtualServiceReqRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"redirect": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Redirect request",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "HTTP or HTTPS protocol",
						ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
					},
					"port": {
						Type:         schema.TypeString, // using TypeString to distinguish unset value
						Optional:     true,
						Description:  "Port to which redirect the request. Default is 80 for HTTP and 443 for HTTPS protocol",
						ValidateFunc: IsIntAndAtLeast(1),
					},
					"status_code": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "One of the redirect status codes - 301, 302, 307",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Host to which redirect the request",
					},
					"path": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Path to which redirect the request",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "Should the query part be preserved",
					},
				},
			},
		},
		"modify_header": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "A set of header modification rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"action": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "One of the following HTTP header actions. Options - ADD, REMOVE, REPLACE",
						ValidateFunc: validation.StringInSlice([]string{"ADD", "REMOVE", "REPLACE"}, false),
					},
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "HTTP header name",
					},
					"value": {
						Type:        schema.TypeString,
						Optional:    true, // 'REMOVE' action does not require value
						Description: "HTTP header value",
					},
				},
			},
		},
		"rewrite_url": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "URL rewrite rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_header": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Host to use for the rewritten URL",
					},
					"existing_path": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Path to use for the rewritten URL",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "Whether or not to keep the existing query string when rewriting the URL",
					},
					"query": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Query string to use or append to the existing query string in the rewritten URL",
					},
				},
			},
		},
	},
}

func resourceVcdAlbVirtualServiceReqRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	cfg, err := getAlbVsHttpRequestRuleType(d)
	if err != nil {
		return diag.Errorf("error getting HTTP Request Rule type: %s", err)
	}

	_, err = albVirtualService.UpdateHttpRequestRules(cfg)
	if err != nil {
		return diag.Errorf("error creating HTTP Request Rules: %s", err)
	}

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return resourceVcdAlbVirtualServiceReqRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceReqRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceReqRulesRead(ctx, d, meta, "resource")
}

func genericVcdAlbVirtualServiceReqRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Get("virtual_service_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) && origin == "resource" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	rules, err := albVirtualService.GetAllHttpRequestRules(nil)
	if err != nil {
		return diag.Errorf("could not retrieve HTTP Request Rules: %s", err)
	}

	dSet(d, "virtual_service_id", albVirtualService.NsxtAlbVirtualService.ID)
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)
	err = setAlbVsHttpRequestRuleData(d, rules)
	if err != nil {
		return diag.Errorf("error storing HTTP Request Rule: %s", err)
	}

	return nil
}

func resourceVcdAlbVirtualServiceReqRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	_, err = albVirtualService.UpdateHttpRequestRules(&types.AlbVsHttpRequestRules{})
	if err != nil {
		return diag.Errorf("error creating HTTP Request Rules: %s", err)
	}

	return nil
}

func getAlbVsHttpRequestRuleType(d *schema.ResourceData) (*types.AlbVsHttpRequestRules, error) {
	rules := d.Get("rule").([]interface{})
	rulesType := make([]types.AlbVsHttpRequestRule, len(rules))

	for ruleIndex, rule := range rules {
		ruleInterface := rule.(map[string]interface{})

		rulesType[ruleIndex].Name = ruleInterface["name"].(string)
		rulesType[ruleIndex].Active = ruleInterface["active"].(bool)
		rulesType[ruleIndex].Logging = ruleInterface["logging"].(bool)
		rulesType[ruleIndex].MatchCriteria = getMatchCriteriaType(ruleInterface["match_criteria"].(*schema.Set))
		rulesType[ruleIndex].RedirectAction, rulesType[ruleIndex].HeaderActions, rulesType[ruleIndex].RewriteURLAction = getActionsType(ruleInterface["actions"].(*schema.Set))

	}

	structure := &types.AlbVsHttpRequestRules{
		Values: rulesType,
	}
	return structure, nil
}

func getMatchCriteriaType(matchCriteria *schema.Set) types.AlbVsHttpRequestAndSecurityRuleMatchCriteria {
	if matchCriteria.Len() == 0 {
		return types.AlbVsHttpRequestAndSecurityRuleMatchCriteria{}
	}
	schemaSet := matchCriteria.List()

	allCriteria := schemaSet[0].(map[string]interface{})
	criteria := types.AlbVsHttpRequestAndSecurityRuleMatchCriteria{}

	clientIpAddressCriteria := allCriteria["client_ip_address"].([]interface{})
	if len(clientIpAddressCriteria) > 0 {
		clientIpAddressCriteriaMap := clientIpAddressCriteria[0].(map[string]interface{})
		criteria.ClientIPMatch = &types.AlbVsHttpRequestRuleClientIPMatch{
			MatchCriteria: clientIpAddressCriteriaMap["criteria"].(string),
			Addresses:     convertSchemaSetToSliceOfStrings(clientIpAddressCriteriaMap["ip_addresses"].(*schema.Set)),
		}
	}

	servicePortsCriteria := allCriteria["service_ports"].([]interface{})
	if len(servicePortsCriteria) > 0 {
		servicePortsCriteriaMap := servicePortsCriteria[0].(map[string]interface{})
		criteria.ServicePortMatch = &types.AlbVsHttpRequestRuleServicePortMatch{
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
		criteria.MethodMatch = &types.AlbVsHttpRequestRuleMethodMatch{
			MatchCriteria: httpMethodCriteriaMap["criteria"].(string),
			Methods:       convertSchemaSetToSliceOfStrings(httpMethodCriteriaMap["methods"].(*schema.Set)),
		}
	}

	pathCriteria := allCriteria["path"].([]interface{})
	if len(pathCriteria) > 0 {
		pathCriteriaMap := pathCriteria[0].(map[string]interface{})
		criteria.PathMatch = &types.AlbVsHttpRequestRulePathMatch{
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
		newHeaderCriteria := make([]types.AlbVsHttpRequestRuleHeaderMatch, requestHeaderCriteria.Len())
		for requestHeaderIndex, requestHeader := range requestHeaderCriteria.List() {
			requestHeaderMap := requestHeader.(map[string]interface{})

			newHeaderCriteria[requestHeaderIndex] = types.AlbVsHttpRequestRuleHeaderMatch{
				MatchCriteria: requestHeaderMap["criteria"].(string),
				Key:           requestHeaderMap["name"].(string),
				Value:         convertSchemaSetToSliceOfStrings(requestHeaderMap["values"].(*schema.Set)),
			}
		}
		criteria.HeaderMatch = newHeaderCriteria
	}

	cookieCriteria := allCriteria["cookie"].([]interface{})
	if len(cookieCriteria) > 0 {
		cookieCriteriaMap := cookieCriteria[0].(map[string]interface{})
		criteria.CookieMatch = &types.AlbVsHttpRequestRuleCookieMatch{
			MatchCriteria: cookieCriteriaMap["criteria"].(string),
			Key:           cookieCriteriaMap["name"].(string),
			Value:         cookieCriteriaMap["value"].(string),
		}
	}

	return criteria
}

func getActionsType(actions *schema.Set) (*types.AlbVsHttpRequestRuleRedirectAction, []*types.AlbVsHttpRequestRuleHeaderActions, *types.AlbVsHttpRequestRuleRewriteURLAction) {
	if actions.Len() == 0 {
		return nil, nil, nil
	}
	schemaSet := actions.List()
	actionsIf := schemaSet[0].(map[string]interface{})

	redirectStructure := actionsIf["redirect"].([]interface{})
	var redir *types.AlbVsHttpRequestRuleRedirectAction
	modifyHeaderStructure := actionsIf["modify_header"].(*schema.Set)
	var mod []*types.AlbVsHttpRequestRuleHeaderActions

	rewriteUrlStructure := actionsIf["rewrite_url"].([]interface{})
	var rew *types.AlbVsHttpRequestRuleRewriteURLAction

	// Process any redirection cases, if specified
	if len(redirectStructure) > 0 {
		redirectStructureMap := redirectStructure[0].(map[string]interface{})
		redir = &types.AlbVsHttpRequestRuleRedirectAction{}

		redir.Protocol = redirectStructureMap["protocol"].(string)
		redir.Host = redirectStructureMap["host"].(string)
		if redirectStructureMap["port"].(string) != "" {
			portInt, _ := strconv.Atoi(redirectStructureMap["port"].(string)) // error is ignored because it is checked at field validation level
			redir.Port = &portInt
		}
		redir.StatusCode = redirectStructureMap["status_code"].(int)
		redir.Path = redirectStructureMap["path"].(string)
		redir.KeepQuery = redirectStructureMap["keep_query"].(bool)
	}

	// Process any header rewrite cases, if specified
	if modifyHeaderStructure.Len() > 0 {
		newModifyHeaderStructure := make([]*types.AlbVsHttpRequestRuleHeaderActions, modifyHeaderStructure.Len())
		for headerIndex, header := range modifyHeaderStructure.List() {
			headerMap := header.(map[string]interface{})

			newModifyHeaderStructure[headerIndex] = &types.AlbVsHttpRequestRuleHeaderActions{
				Action: headerMap["action"].(string),
				Name:   headerMap["name"].(string),
				Value:  headerMap["value"].(string),
			}
		}
		mod = newModifyHeaderStructure
	}

	// Process any rewrite_url cases if any
	if len(rewriteUrlStructure) > 0 {
		rewriteUrlStructureMap := rewriteUrlStructure[0].(map[string]interface{})
		rew = &types.AlbVsHttpRequestRuleRewriteURLAction{}
		rew.Host = rewriteUrlStructureMap["host_header"].(string)
		rew.Path = rewriteUrlStructureMap["existing_path"].(string)
		rew.KeepQuery = rewriteUrlStructureMap["keep_query"].(bool)
		rew.Query = rewriteUrlStructureMap["query"].(string)

	}

	return redir, mod, rew
}

func setAlbVsHttpRequestRuleData(d *schema.ResourceData, rules []*types.AlbVsHttpRequestRule) error {
	allRules := make([]interface{}, len(rules))
	for ruleIndex, rule := range rules {

		singleRule := make(map[string]interface{})

		singleRule["name"] = rule.Name
		singleRule["active"] = rule.Active
		singleRule["logging"] = rule.Logging

		// 'match_criteria' block

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
		requestHeaders := make([]interface{}, len(rule.MatchCriteria.HeaderMatch))
		for i, h := range rule.MatchCriteria.HeaderMatch {
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

		// Pack root entry
		matchCriteria[0] = matchCriteriaMap
		singleRule["match_criteria"] = matchCriteria

		// EOF 'match_criteria' block

		// 'actions' block

		actions := make([]interface{}, 1)
		actionsMap := make(map[string]interface{})

		// 'redirect'
		redirect := make([]interface{}, 0)
		if rule.RedirectAction != nil {
			singleRedirect := make(map[string]interface{})
			singleRedirect["protocol"] = rule.RedirectAction.Protocol
			if rule.RedirectAction.Port != nil {
				singleRedirect["port"] = strconv.Itoa(*rule.RedirectAction.Port)
			} else {
				singleRedirect["port"] = ""
			}
			singleRedirect["status_code"] = rule.RedirectAction.StatusCode
			singleRedirect["host"] = rule.RedirectAction.Host
			singleRedirect["path"] = rule.RedirectAction.Path
			singleRedirect["keep_query"] = rule.RedirectAction.KeepQuery

			redirect = append(redirect, singleRedirect)
		}
		actionsMap["redirect"] = redirect

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

		// 'rewrite_url'
		rewriteUrl := make([]interface{}, 0)
		if rule.RewriteURLAction != nil {
			singleRewriteUrl := make(map[string]interface{})
			singleRewriteUrl["host_header"] = rule.RewriteURLAction.Host
			singleRewriteUrl["existing_path"] = rule.RewriteURLAction.Path
			singleRewriteUrl["keep_query"] = rule.RewriteURLAction.KeepQuery
			singleRewriteUrl["query"] = rule.RewriteURLAction.Query

			rewriteUrl = append(rewriteUrl, singleRewriteUrl)
		}
		actionsMap["rewrite_url"] = rewriteUrl

		actions[0] = actionsMap
		singleRule["actions"] = actions

		// EOF 'actions' block
		allRules[ruleIndex] = singleRule
	}

	return d.Set("rule", allRules)
}

func resourceVcdAlbVirtualServiceHttpPolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Virtual Service HTTP Policy import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.nsxt-edge-gw-name.virtual_service_name")
	}
	orgName, vdcOrVdcGroupName, edgeName, virtualServiceName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("ALB Virtual Services are only supported on NSX-T. Please use 'vcd_lb_virtual_server' for NSX-V load balancers")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	albVirtualService, err := vcdClient.GetAlbVirtualServiceByName(edge.EdgeGateway.ID, virtualServiceName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T ALB Virtual Service '%s': %s", virtualServiceName, err)
	}

	dSet(d, "virtual_service_id", albVirtualService.NsxtAlbVirtualService.ID)
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return []*schema.ResourceData{d}, nil
}
