package vcd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func resourceVcdAlbVirtualServiceSecRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceSecRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceSecRulesRead,
		// Update is the same as create and it does not have any additional details like rule IDs
		// which are important for updates in some cases.
		UpdateContext: resourceVcdAlbVirtualServiceSecRulesCreate,
		DeleteContext: resourceVcdAlbVirtualServiceSecRulesDelete,
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
				Elem:        nsxtAlbVirtualServiceSecRule,
				Description: "A single HTTP Security Rule",
			},
		},
	}
}

var nsxtAlbVirtualServiceSecRule = &schema.Resource{
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
			Description: "Defines is the rule is active or not",
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
			// Match criteria are the same as for HTTP Request
			Elem: nsxtAlbVsReqAndSecRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			MaxItems:    1,
			Required:    true,
			Description: "Actions to perform with the rule that matches",
			Elem:        nsxtAlbVsSecRuleActions,
		},
	},
}

var nsxtAlbVsSecRuleActions = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"redirect_to_https": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Port number that should be redirected to HTTPS",
			ValidateFunc: IsIntAndAtLeast(1), // Using TypeString + validation to be able to distinguish empty value and '0'
		},
		"connections": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "ALLOW or CLOSE connections",
			ValidateFunc: validation.StringInSlice([]string{"ALLOW", "CLOSE"}, false),
		},
		"rate_limit": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Apply actions based on rate limits",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "Maximum number of connections, requests or packets permitted each period. The count must be between 1 and 1000000000",
						ValidateFunc: IsIntAndAtLeast(1), // Using TypeString + validation to be able to distinguish empty value and '0'
					},
					"period": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "Time value in seconds to enforce rate count. The period must be between 1 and 1000000000",
						ValidateFunc: IsIntAndAtLeast(1), // Using TypeString + validation to be able to distinguish empty value and '0'
					},
					"action_close_connection": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Set to true if the connection should be closed",
					},
					"action_redirect": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Redirect based on rate limits",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"protocol": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "HTTP or HTTPS protocol",
									ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, false),
								},
								"port": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "Port to which redirect the request",
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
					"action_local_response": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Send custom response",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"status_code": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "HTTP Status code to send",
									ValidateFunc: IsIntAndAtLeast(1), // Using TypeString + validation to be able to distinguish empty value and '0'
								},
								"content": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Base64 encoded content",
								},
								"content_type": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "MIME type for the content",
								},
							},
						},
					},
				},
			},
		},

		"send_response": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Send custom response",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"status_code": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "HTTP Status code to send",
						ValidateFunc: IsIntAndAtLeast(1), // Using TypeString + validation to be able to distinguish empty value and '0'
					},
					"content": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Base64 encoded content",
					},
					"content_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "MIME type for the content",
					},
				},
			},
		},
	},
}

func resourceVcdAlbVirtualServiceSecRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	cfg, err := getAlbVsHttpSecurityRuleType(d)
	if err != nil {
		return diag.Errorf("error getting HTTP Security Rule type: %s", err)
	}

	_, err = albVirtualService.UpdateHttpSecurityRules(cfg)
	if err != nil {
		return diag.Errorf("error creating HTTP Security Rules: %s", err)
	}

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return resourceVcdAlbVirtualServiceSecRulesRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceSecRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdAlbVirtualServiceSecRulesRead(ctx, d, meta, "resource")
}

func genericVcdAlbVirtualServiceSecRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Get("virtual_service_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) && origin == "resource" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	rules, err := albVirtualService.GetAllHttpSecurityRules(nil)
	if err != nil {
		return diag.Errorf("could not retrieve HTTP Security Rules: %s", err)
	}

	dSet(d, "virtual_service_id", albVirtualService.NsxtAlbVirtualService.ID)
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)
	err = setAlbVsHttpSecuritytRuleData(d, rules)
	if err != nil {
		return diag.Errorf("error storing HTTP Security Rule: %s", err)
	}

	return nil
}

func resourceVcdAlbVirtualServiceSecRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVsId := d.Get("virtual_service_id").(string)
	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
	defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

	_, err = albVirtualService.UpdateHttpSecurityRules(&types.AlbVsHttpSecurityRules{})
	if err != nil {
		return diag.Errorf("error creating HTTP Security Rules: %s", err)
	}

	return nil
}

func getAlbVsHttpSecurityRuleType(d *schema.ResourceData) (*types.AlbVsHttpSecurityRules, error) {
	rules := d.Get("rule").([]interface{})
	rulesType := make([]types.AlbVsHttpSecurityRule, len(rules))

	for ruleIndex, rule := range rules {
		ruleInterface := rule.(map[string]interface{})

		rulesType[ruleIndex].Name = ruleInterface["name"].(string)
		rulesType[ruleIndex].Active = ruleInterface["active"].(bool)
		rulesType[ruleIndex].Logging = ruleInterface["logging"].(bool)
		rulesType[ruleIndex].MatchCriteria = getMatchCriteriaType(ruleInterface["match_criteria"].(*schema.Set))

		redirectToHttpsActions, allowOrCloseConnection, rateLimitAction, localResponseAction, err := getSecurityActionsType(ruleInterface["actions"].(*schema.Set))
		if err != nil {
			return nil, err
		}

		rulesType[ruleIndex].RedirectToHTTPSAction = redirectToHttpsActions
		rulesType[ruleIndex].AllowOrCloseConnectionAction = allowOrCloseConnection
		rulesType[ruleIndex].RateLimitAction = rateLimitAction
		rulesType[ruleIndex].LocalResponseAction = localResponseAction
	}

	structure := &types.AlbVsHttpSecurityRules{
		Values: rulesType,
	}
	return structure, nil
}

func getSecurityActionsType(actions *schema.Set) (*types.AlbVsHttpSecurityRuleRedirectToHTTPSAction, string, *types.AlbVsHttpSecurityRuleRateLimitAction, *types.AlbVsHttpSecurityRuleRateLimitLocalResponseAction, error) {
	if actions.Len() == 0 {
		return nil, "", nil, nil, nil
	}
	schemaSet := actions.List()
	actionsIf := schemaSet[0].(map[string]interface{})

	// 'redirect_to_https'
	redirToHttps := actionsIf["redirect_to_https"].(string)
	var redirToHttpsStruct *types.AlbVsHttpSecurityRuleRedirectToHTTPSAction
	if redirToHttps != "" {
		intPort, err := strconv.Atoi(redirToHttps)
		if err != nil {
			return nil, "", nil, nil, fmt.Errorf("error converting 'redirect_to_https' field to integer: %s", err)
		}
		redirToHttpsStruct = &types.AlbVsHttpSecurityRuleRedirectToHTTPSAction{Port: intPort}
	}

	// 'connections'
	connections := actionsIf["connections"].(string)

	// 'rate_limit'
	rateLimitSlice := actionsIf["rate_limit"].([]interface{})
	var rateLimitType *types.AlbVsHttpSecurityRuleRateLimitAction
	if len(rateLimitSlice) > 0 {
		rateLimitMap := rateLimitSlice[0].(map[string]interface{})

		rateLimitCountStr := rateLimitMap["count"].(string)
		rateLimitCountInt, err := strconv.Atoi(rateLimitCountStr)
		if err != nil {
			return nil, "", nil, nil, fmt.Errorf("error converting 'rate_limit.0.count' to int: %s", err)
		}

		rateLimitPeriodStr := rateLimitMap["period"].(string)
		rateLimitPeriodInt, err := strconv.Atoi(rateLimitPeriodStr)
		if err != nil {
			return nil, "", nil, nil, fmt.Errorf("error converting 'rate_limit.0.period' to int: %s", err)
		}

		rateLimitType = &types.AlbVsHttpSecurityRuleRateLimitAction{
			Count:  rateLimitCountInt,
			Period: rateLimitPeriodInt,
		}

		// Check if any action for rate limit is set
		// 'action_close_connection'
		rateLimitActionCloseConnection := rateLimitMap["action_close_connection"].(bool)
		if rateLimitActionCloseConnection {
			rateLimitType.CloseConnectionAction = "CLOSE" // The only option possible
		}

		// 'action_redirect'
		rateLimitActionRedirect := rateLimitMap["action_redirect"].([]interface{})
		var redir *types.AlbVsHttpRequestRuleRedirectAction
		if len(rateLimitActionRedirect) > 0 {
			redirectStructureMap := rateLimitActionRedirect[0].(map[string]interface{})
			redir = &types.AlbVsHttpRequestRuleRedirectAction{}

			redir.Protocol = redirectStructureMap["protocol"].(string)
			redir.Host = redirectStructureMap["host"].(string)
			if redirectStructureMap["port"] != "" {
				portInt, _ := strconv.Atoi(redirectStructureMap["port"].(string)) // error is ignored because it is checked at field validation level
				redir.Port = &portInt
			}
			redir.StatusCode = redirectStructureMap["status_code"].(int)
			redir.Path = redirectStructureMap["path"].(string)
			redir.KeepQuery = redirectStructureMap["keep_query"].(bool)

			rateLimitType.RedirectAction = redir
		}

		// 'action_local_response'
		rateLimitActionLocalResponse := rateLimitMap["action_local_response"].([]interface{})
		var rateLimitSendResponseType *types.AlbVsHttpSecurityRuleRateLimitLocalResponseAction
		if len(rateLimitActionLocalResponse) > 0 {
			redirectStructureMap := rateLimitActionLocalResponse[0].(map[string]interface{})

			statusCodeStr := redirectStructureMap["status_code"].(string)
			statusCodeInt, err := strconv.Atoi(statusCodeStr)
			if err != nil {
				return nil, "", nil, nil, fmt.Errorf("error converting 'send_response.0.status_code' to int: %s", err)
			}

			rateLimitSendResponseType = &types.AlbVsHttpSecurityRuleRateLimitLocalResponseAction{
				Content:     redirectStructureMap["content"].(string),
				ContentType: redirectStructureMap["content_type"].(string),
				StatusCode:  statusCodeInt,
			}

			rateLimitType.LocalResponseAction = rateLimitSendResponseType
		}

	}

	// 'send_response'
	sendResponse := actionsIf["send_response"].([]interface{})
	var sendResponseType *types.AlbVsHttpSecurityRuleRateLimitLocalResponseAction
	if len(sendResponse) > 0 {
		sendResponseMap := sendResponse[0].(map[string]interface{})

		statusCodeStr := sendResponseMap["status_code"].(string)
		statusCodeInt, err := strconv.Atoi(statusCodeStr)
		if err != nil {
			return nil, "", nil, nil, fmt.Errorf("error converting 'send_response.0.status_code' to int: %s", err)
		}

		sendResponseType = &types.AlbVsHttpSecurityRuleRateLimitLocalResponseAction{
			Content:     sendResponseMap["content"].(string),
			ContentType: sendResponseMap["content_type"].(string),
			StatusCode:  statusCodeInt,
		}
	}

	return redirToHttpsStruct, connections, rateLimitType, sendResponseType, nil
}

func setAlbVsHttpSecuritytRuleData(d *schema.ResourceData, rules []*types.AlbVsHttpSecurityRule) error {
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

		// "singlePath"
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

		// EOF 'match_criteria'

		// 'actions'

		actions := make([]interface{}, 1)
		actionsMap := make(map[string]interface{})

		// 'redirect_to_https'
		if rule.RedirectToHTTPSAction != nil {
			port := rule.RedirectToHTTPSAction.Port
			actionsMap["redirect_to_https"] = strconv.Itoa(port)
		} else {
			actionsMap["redirect_to_https"] = ""
		}

		// 'connections'
		if rule.AllowOrCloseConnectionAction != "" {
			actionsMap["connections"] = rule.AllowOrCloseConnectionAction
		} else {
			actionsMap["connections"] = ""
		}

		// 'rate_limit'
		rateLimit := make([]interface{}, 0)
		if rule.RateLimitAction != nil {
			rateLimitEntry := make(map[string]interface{})

			rateLimitEntry["count"] = strconv.Itoa(rule.RateLimitAction.Count)
			rateLimitEntry["period"] = strconv.Itoa(rule.RateLimitAction.Period)
			if rule.RateLimitAction.CloseConnectionAction != "" {
				rateLimitEntry["action_close_connection"] = true
			}

			singleRedirectActionEntryInterface := make([]interface{}, 0)
			if rule.RateLimitAction.RedirectAction != nil {
				singleRedirectActionEntry := make(map[string]interface{})

				singleRedirectActionEntry["protocol"] = rule.RateLimitAction.RedirectAction.Protocol
				if rule.RateLimitAction.RedirectAction.Port != nil {
					singleRedirectActionEntry["port"] = strconv.Itoa(*rule.RateLimitAction.RedirectAction.Port)
				} else {
					singleRedirectActionEntry["port"] = ""
				}
				singleRedirectActionEntry["status_code"] = rule.RateLimitAction.RedirectAction.StatusCode
				singleRedirectActionEntry["host"] = rule.RateLimitAction.RedirectAction.Host
				singleRedirectActionEntry["path"] = rule.RateLimitAction.RedirectAction.Path
				singleRedirectActionEntry["keep_query"] = rule.RateLimitAction.RedirectAction.KeepQuery

				singleRedirectActionEntryInterface = append(singleRedirectActionEntryInterface, singleRedirectActionEntry)
			}
			rateLimitEntry["action_redirect"] = singleRedirectActionEntryInterface

			singleLocalResponseActionEntryInterface := make([]interface{}, 0)
			if rule.RateLimitAction.LocalResponseAction != nil {
				singleLocalResponseActionEntry := make(map[string]interface{})

				singleLocalResponseActionEntry["content"] = rule.RateLimitAction.LocalResponseAction.Content
				singleLocalResponseActionEntry["content_type"] = rule.RateLimitAction.LocalResponseAction.ContentType
				singleLocalResponseActionEntry["status_code"] = strconv.Itoa(rule.RateLimitAction.LocalResponseAction.StatusCode)

				singleLocalResponseActionEntryInterface = append(singleLocalResponseActionEntryInterface, singleLocalResponseActionEntry)
			}
			rateLimitEntry["action_local_response"] = singleLocalResponseActionEntryInterface

			rateLimit = append(rateLimit, rateLimitEntry)
		}
		actionsMap["rate_limit"] = rateLimit

		// 'send_response'
		sendResponse := make([]interface{}, 0)
		if rule.LocalResponseAction != nil {
			singleEntry := make(map[string]interface{})
			singleEntry["content"] = rule.LocalResponseAction.Content
			singleEntry["content_type"] = rule.LocalResponseAction.ContentType
			singleEntry["status_code"] = strconv.Itoa(rule.LocalResponseAction.StatusCode)

			sendResponse = append(sendResponse, singleEntry)
		}
		actionsMap["send_response"] = sendResponse

		actions[0] = actionsMap
		singleRule["actions"] = actions

		// EOF 'actions'

		allRules[ruleIndex] = singleRule
	}

	return d.Set("rule", allRules)
}
