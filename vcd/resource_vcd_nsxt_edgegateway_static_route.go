package vcd

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtEdgeGatewayStaticRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgeGatewayStaticRouteCreate,
		ReadContext:   resourceVcdNsxtEdgeGatewayStaticRouteRead,
		UpdateContext: resourceVcdNsxtEdgeGatewayStaticRouteUpdate,
		DeleteContext: resourceVcdNsxtEdgeGatewayStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgeGatewayStaticRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID for Static Route configuration",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Static Route",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of Static Route",
			},
			"network_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network CIDR (e.g. 192.168.1.1/24) for Static Route",
			},
			"next_hop": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "A set of next hops to use within the static route",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "IP Address of next hop",
						},
						"admin_distance": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Admin distance of next hop",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"scope": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "ID of Scope element",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of Scope element",
									},
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Scope type - One of 'NETWORK', 'SYSTEM_OWNED'",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxtEdgeGatewayStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route create] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route create] error retrieving Edge Gateway: %s", err)
	}

	staticRouteConfig := getStaticRouteType(d)

	createdStaticRoute, err := nsxtEdge.CreateStaticRoute(staticRouteConfig)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route create] error creating Static Route: %s", err)
	}

	d.SetId(createdStaticRoute.NsxtEdgeGatewayStaticRoute.ID)

	return resourceVcdNsxtEdgeGatewayStaticRouteRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route update] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route update] error retrieving Edge Gateway: %s", err)
	}

	staticRoute, err := nsxtEdge.GetStaticRouteById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route update] error retrieving existing Static Route by ID: %s", err)
	}

	staticRouteConfig := getStaticRouteType(d)
	staticRouteConfig.ID = staticRoute.NsxtEdgeGatewayStaticRoute.ID
	staticRouteConfig.Version = staticRoute.NsxtEdgeGatewayStaticRoute.Version

	_, err = staticRoute.Update(staticRouteConfig)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route update] error updating Static Route: %s", err)
	}

	return resourceVcdNsxtEdgeGatewayStaticRouteRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
		}
		return diag.Errorf("[NSX-T Edge Gateway Static Route read] error retrieving Edge Gateway: %s", err)
	}

	staticRoute, err := nsxtEdge.GetStaticRouteById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route read]: failed to get Static Route By ID %s", err)
	}

	err = setStaticRouteData(staticRoute.NsxtEdgeGatewayStaticRoute, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVcdNsxtEdgeGatewayStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route delete] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route delete] error retrieving Edge Gateway: %s", err)
	}

	staticRoute, err := nsxtEdge.GetStaticRouteById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route delete]: %s", err)
	}

	err = staticRoute.Delete()
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route delete] error deleting Static Route: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgeGatewayStaticRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.SplitN(d.Id(), ImportSeparator, 4)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.edge_gateway_name.static_route_name or "+
			"'org-name.vdc-or-vdc-group-name.edge_gateway_name.name', got '%s'", d.Id())
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, staticRouteCidrOrName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("[NSX-T Edge Gateway Static Route import] unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	_, _, err = net.ParseCIDR(staticRouteCidrOrName)
	isNetworkCidr := err == nil // when error is nil, this is a CIDR

	var staticRoute *govcd.NsxtEdgeGatewayStaticRoute

	if isNetworkCidr {
		staticRoute, err = edgeGateway.GetStaticRouteByNetworkCidr(staticRouteCidrOrName)
		if err != nil {
			return nil, fmt.Errorf("[NSX-T Edge Gateway Static Route import] unable to find Static Route with Network CIDR '%s': %s", staticRouteCidrOrName, err)
		}
	} else { // by name
		staticRoute, err = edgeGateway.GetStaticRouteByName(staticRouteCidrOrName)
		if err != nil {
			return nil, fmt.Errorf("[NSX-T Edge Gateway Static Route import] unable to find Static Route with Name '%s': %s", staticRouteCidrOrName, err)
		}
	}

	if staticRoute == nil {
		return nil, fmt.Errorf("NSX-T Edge Gateway Static Route import unable to find Static Route by the following ID: %s", d.Id())
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(staticRoute.NsxtEdgeGatewayStaticRoute.ID)

	return []*schema.ResourceData{d}, nil
}

func getStaticRouteType(d *schema.ResourceData) *types.NsxtEdgeGatewayStaticRoute {

	srConfig := &types.NsxtEdgeGatewayStaticRoute{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		NetworkCidr: d.Get("network_cidr").(string),
		NextHops:    []types.NsxtEdgeGatewayStaticRouteNextHops{},
	}

	nextHopSet := d.Get("next_hop").(*schema.Set)
	nextHopList := nextHopSet.List()

	for _, singleHop := range nextHopList {
		nextHopMap := singleHop.(map[string]interface{})

		nextHop := types.NsxtEdgeGatewayStaticRouteNextHops{
			IPAddress:     nextHopMap["ip_address"].(string),
			AdminDistance: nextHopMap["admin_distance"].(int),
		}

		// next hop scope (Optional TypeList, MaxItems 1)
		nextHopScopeSlice := nextHopMap["scope"].([]interface{})
		if len(nextHopScopeSlice) > 0 {
			nextHopScope := nextHopScopeSlice[0]
			nextHopScopeMap := convertToStringMap(nextHopScope.(map[string]interface{}))
			nextHop.Scope = &types.NsxtEdgeGatewayStaticRouteNextHopScope{
				ID:        nextHopScopeMap["id"],
				Name:      nextHopScopeMap["name"],
				ScopeType: nextHopScopeMap["type"],
			}
		}

		srConfig.NextHops = append(srConfig.NextHops, nextHop)

	}

	return srConfig
}

func setStaticRouteData(staticRoute *types.NsxtEdgeGatewayStaticRoute, d *schema.ResourceData) error {
	dSet(d, "name", staticRoute.Name)
	dSet(d, "description", staticRoute.Description)
	dSet(d, "network_cidr", staticRoute.NetworkCidr)

	// Construct 'next_hop' structure
	nextHopInterface := make([]interface{}, len(staticRoute.NextHops))
	for nextHopIndex, nextHop := range staticRoute.NextHops {
		singleHop := make(map[string]interface{})
		singleHop["ip_address"] = nextHop.IPAddress
		singleHop["admin_distance"] = nextHop.AdminDistance

		// Construct next_hop.X.scope
		if nextHop.Scope != nil {
			singleHopScope := make(map[string]interface{})
			singleHopScope["id"] = nextHop.Scope.ID
			singleHopScope["name"] = nextHop.Scope.Name
			singleHopScope["type"] = nextHop.Scope.ScopeType

			singleHopScopeInterface := make([]interface{}, 1)
			singleHopScopeInterface[0] = singleHopScope

			singleHop["scope"] = singleHopScopeInterface
		}

		nextHopInterface[nextHopIndex] = singleHop
	}

	err := d.Set("next_hop", nextHopInterface)
	if err != nil {
		return fmt.Errorf("error storing 'next_hop' into state: %s", err)
	}

	return nil
}
