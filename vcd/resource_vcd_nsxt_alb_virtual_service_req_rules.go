package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdAlbVirtualServiceReqRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceReqRulesCreate,
		ReadContext:   resourceVcdAlbVirtualServiceReqRulesRead,
		UpdateContext: resourceVcdAlbVirtualServiceReqRulesCreate,
		DeleteContext: resourceVcdAlbVirtualServiceReqRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceReqRulesImport,
		},

		Schema: map[string]*schema.Schema{
			"virtual_service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Virtual Service ID",
			},
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     nsxtAlbVirtualServiceReqRule,
			},
		},
	}
}

var nsxtAlbVirtualServiceReqRule = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"state": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "",
		},
		"match_criteria": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "",
			Elem:        nsxtAlbVirtualServiceReqRuleMatchCriteria,
		},
		"actions": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "",
			Elem:        nsxtAlbVirtualServiceReqRuleActions,
		},
	},
}

var nsxtAlbVirtualServiceReqRuleMatchCriteria = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"client_ip_address": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:     schema.TypeString,
						Optional: true,
						// ValidateFunc: validation.StringInSlice([]string{"IS_IN", "IS_NOT_IN"}, false),
					},
					"ip_addresses": {
						Type:        schema.TypeSet,
						Optional:    true,
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
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"ports": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Enter IPv4 or IPv6 address, range or CIDR",
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
		},
		"http_method": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"method": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
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
						Type:     schema.TypeString,
						Optional: true,
					},
					"paths": {
						// Type:     schema.TypeString,
						// Optional: true,
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Enter IPv4 or IPv6 address, range or CIDR",
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
			Description: "Enter text contained in the query string",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"request_headers": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Enter a port or a comma separated port list. The port numbers must be between 1 and 65535 (Example: 80 or 80,90,443)",
					},
					"value": {
						// Type:     schema.TypeString,
						// Optional: true,
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
		"cookie": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"criteria": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"value": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
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
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"port": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "",
					},
					"status_code": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"path": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "",
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
						Type:     schema.TypeString,
						Optional: true,
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"value": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
				},
			},
		},
		"rewrite_url": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_header": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"existing_path": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
					},
					"keep_query": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "",
					},
					"query": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "",
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

	cfg, err := getEdgeVirtualServiceHttpRequestRuleType(d)
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

/*
	func resourceVcdAlbVirtualServiceReqRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		vcdClient := meta.(*VCDClient)

		albVsId := d.Get("virtual_service_id").(string)
		albVirtualService, err := vcdClient.GetAlbVirtualServiceById(albVsId)
		if err != nil {
			return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
		}

		vcdMutexKV.kvLock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)
		defer vcdMutexKV.kvUnlock(albVirtualService.NsxtAlbVirtualService.GatewayRef.ID)

		return resourceVcdAlbVirtualServiceReqRulesRead(ctx, d, meta)
	}
*/
func resourceVcdAlbVirtualServiceReqRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	rules, err := albVirtualService.GetAllHttpRequestRules(nil)
	if err != nil {
		return diag.Errorf("could not retrieve HTTP Request Rules: %s", err)
	}

	err = setEdgeVirtualServiceHttpRequestRuleData(d, rules)
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

	_, err = albVirtualService.UpdateHttpRequestRules(&types.EdgeVirtualServiceHttpRequestRules{})
	if err != nil {
		return diag.Errorf("error creating HTTP Request Rules: %s", err)
	}

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return nil
}

func resourceVcdAlbVirtualServiceReqRulesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func getEdgeVirtualServiceHttpRequestRuleType(d *schema.ResourceData) (*types.EdgeVirtualServiceHttpRequestRules, error) {

	rules := d.Get("rule").(*schema.Set)
	rulesType := make([]types.EdgeVirtualServiceHttpRequestRule, rules.Len())

	for ruleIndex, rule := range rules.List() {
		ruleInterface := rule.(map[string]interface{})

		rulesType[ruleIndex].Name = ruleInterface["name"].(string)
		rulesType[ruleIndex].Active = ruleInterface["state"].(bool)
		rulesType[ruleIndex].MatchCriteria = getMatchCriteriaType(ruleInterface["match_criteria"].(*schema.Set))
		rulesType[ruleIndex].RedirectAction, rulesType[ruleIndex].HeaderActions, rulesType[ruleIndex].RewriteURLAction = getActionsType(ruleInterface["actions"].(*schema.Set))

	}

	structure := &types.EdgeVirtualServiceHttpRequestRules{
		Values: rulesType,
	}
	return structure, nil
}

func getMatchCriteriaType(matchCriteria *schema.Set) types.EdgeVirtualServiceHttpRequestRuleMatchCriteria {
	if matchCriteria.Len() == 0 {
		return types.EdgeVirtualServiceHttpRequestRuleMatchCriteria{}
	}
	schemaSet := matchCriteria.List()

	allCriteria := schemaSet[0].(map[string]interface{})
	criteria := types.EdgeVirtualServiceHttpRequestRuleMatchCriteria{}

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

	httpMethodCriteria := allCriteria["http_method"].([]interface{})
	if len(httpMethodCriteria) > 0 {
		httpMethodCriteriaMap := httpMethodCriteria[0].(map[string]interface{})
		criteria.MethodMatch = &types.EdgeVirtualServiceHttpRequestRuleMethodMatch{
			MatchCriteria: httpMethodCriteriaMap["criteria"].(string),
			Methods:       []string{httpMethodCriteriaMap["method"].(string)},
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
				Value:         convertSchemaSetToSliceOfStrings(requestHeaderMap["value"].(*schema.Set)),
			}
		}
		criteria.HeaderMatch = newHeaderCriteria
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

	return criteria
}

func getActionsType(actions *schema.Set) (*types.EdgeVirtualServiceHttpRequestRuleRedirectAction, []*types.EdgeVirtualServiceHttpRequestRuleHeaderActions, *types.EdgeVirtualServiceHttpRequestRuleRewriteURLAction) {
	if actions.Len() == 0 {
		return nil, nil, nil
	}
	schemaSet := actions.List()
	actionsIf := schemaSet[0].(map[string]interface{})

	redirectStructure := actionsIf["redirect"].([]interface{})
	var redir *types.EdgeVirtualServiceHttpRequestRuleRedirectAction
	modifyHeaderStructure := actionsIf["modify_header"].(*schema.Set)
	var mod []*types.EdgeVirtualServiceHttpRequestRuleHeaderActions

	rewriteUrlStructure := actionsIf["rewrite_url"].([]interface{})
	var rew *types.EdgeVirtualServiceHttpRequestRuleRewriteURLAction

	// Process any redirection cases, if specified
	if len(redirectStructure) > 0 {
		redirectStructureMap := redirectStructure[0].(map[string]interface{})
		redir = &types.EdgeVirtualServiceHttpRequestRuleRedirectAction{}

		redir.Protocol = redirectStructureMap["protocol"].(string)
		redir.Host = redirectStructureMap["host"].(string)
		redir.Port = redirectStructureMap["port"].(int)
		redir.StatusCode = redirectStructureMap["status_code"].(int)
		redir.Path = redirectStructureMap["path"].(string)
		redir.KeepQuery = redirectStructureMap["keep_query"].(bool)
	}

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

	// Process any rewrite_url cases if any
	if len(rewriteUrlStructure) > 0 {
		rewriteUrlStructureMap := rewriteUrlStructure[0].(map[string]interface{})
		rew = &types.EdgeVirtualServiceHttpRequestRuleRewriteURLAction{}
		rew.Host = rewriteUrlStructureMap["host_header"].(string)
		rew.Path = rewriteUrlStructureMap["existing_path"].(string)
		rew.KeepQuery = rewriteUrlStructureMap["keep_query"].(bool)
		rew.Query = rewriteUrlStructureMap["query"].(string)

	}

	return redir, mod, rew
}

func setEdgeVirtualServiceHttpRequestRuleData(d *schema.ResourceData, rules []*types.EdgeVirtualServiceHttpRequestRule) error {
	allRules := make([]interface{}, len(rules))

	for ruleIndex, rule := range rules {

		singleRule := make(map[string]interface{})

		singleRule["name"] = rule.Name
		singleRule["state"] = rule.Active

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

		// "http_method"
		httpMethod := make([]interface{}, 0)
		if rule.MatchCriteria.MethodMatch != nil {
			singleHttpMethod := make(map[string]interface{})
			singleHttpMethod["criteria"] = rule.MatchCriteria.MethodMatch.MatchCriteria
			singleHttpMethod["method"] = rule.MatchCriteria.MethodMatch.Methods[0]
			httpMethod = append(httpMethod, singleHttpMethod)
		}
		matchCriteriaMap["http_method"] = httpMethod

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
			singleHedear["value"] = convertStringsToTypeSet(h.Value)

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

		////////// EOF match_criteria

		//// 'actions'

		actions := make([]interface{}, 1)
		actionsMap := make(map[string]interface{})

		// 'redirect'
		redirect := make([]interface{}, 0)
		if rule.RedirectAction != nil {
			singleRedirect := make(map[string]interface{})
			singleRedirect["protocol"] = rule.RedirectAction.Protocol
			singleRedirect["port"] = rule.RedirectAction.Port
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

		//// EOF 'actions'

		allRules[ruleIndex] = singleRule
	}

	return d.Set("rule", allRules)
}
