package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxtEdgeGatewayStaticRoute() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayStaticRouteRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway ID for Static Route configuration",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Static Route",
			},
			"network_cidr": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Network CIDR (e.g. 192.168.1.1/24) for Static Route",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Static Route",
			},
			"next_hop": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of next hops used within the static route",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP Address of next hop",
						},
						"admin_distance": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Admin distance of next hop",
						},
						"scope": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of Scope element",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of Scope element",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
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

func datasourceVcdNsxtEdgeGatewayStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route DS read] error retrieving Edge Gateway: %s", err)
	}

	searchName := d.Get("name").(string)
	searchNetworkCidr := d.Get("network_cidr").(string)

	// API does not support filtering yet therefore all Static Routes need to be retrieved and filtered manually
	// There is also a problem that 'name' must not be unique as well as 'network_cidr', however the
	// goal is to attempt 2 pass filtering - 'name' is mandatory, but if it happens it is not
	// unique, one can attempt to make the search more precise by specifying 'network_cidr'
	allStaticRoutes, err := nsxtEdge.GetAllStaticRoutes(nil)
	if err != nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route DS read] failed to get all Static Routes: %s", err)
	}

	var foundResult *govcd.NsxtEdgeGatewayStaticRoute

	filteredByName := make([]*govcd.NsxtEdgeGatewayStaticRoute, 0)
	// First - filter by name
	for _, sr := range allStaticRoutes {
		if sr.NsxtEdgeGatewayStaticRoute.Name == searchName {
			filteredByName = append(filteredByName, sr)
		}
	}

	if len(filteredByName) == 0 {
		return diag.Errorf("%s no NSX-T Edge Gateway Static Routes with Name '%s' found",
			govcd.ErrorEntityNotFound, searchName)
	}

	if len(filteredByName) == 1 {
		foundResult = filteredByName[0]
	}

	if len(filteredByName) > 1 && searchNetworkCidr == "" {
		return diag.Errorf("found more than one '%d' NSX-T Edge Gateway Static Routes. Please specify 'network_cidr' to make the search more accurate", len(filteredByName))
	}

	if searchNetworkCidr != "" {
		filteredByNetworkCidr := make([]*govcd.NsxtEdgeGatewayStaticRoute, 0)
		for _, sr := range filteredByName {
			if sr.NsxtEdgeGatewayStaticRoute.NetworkCidr == searchNetworkCidr {
				filteredByNetworkCidr = append(filteredByNetworkCidr, sr)
			}
		}

		if len(filteredByNetworkCidr) == 0 {
			return diag.Errorf("[NSX-T Edge Gateway Static Route DS read] %s no NSX-T Edge Gateway Static Routes found with Name '%s' and Network CIDR '%s'",
				govcd.ErrorEntityNotFound, searchName, searchNetworkCidr)
		}

		if len(filteredByNetworkCidr) > 1 {
			return diag.Errorf("[NSX-T Edge Gateway Static Route DS read] cannot identify single result. '%d' NSX-T Edge Gateway Static Routes found.", len(filteredByNetworkCidr))
		}

		// found exactly one value
		foundResult = filteredByNetworkCidr[0]
	}

	if foundResult == nil {
		return diag.Errorf("[NSX-T Edge Gateway Static Route DS read] %s no NSX-T Edge Gateway Static Routes found", govcd.ErrorEntityNotFound)
	}

	d.SetId(foundResult.NsxtEdgeGatewayStaticRoute.ID)
	err = setStaticRouteData(foundResult.NsxtEdgeGatewayStaticRoute, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
