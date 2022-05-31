package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
	"strings"
)

func resourceVcdNsxtRouteAdvertisement() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		ReadContext:   resourceVcdNsxtRouteAdvertisementRead,
		UpdateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		DeleteContext: resourceVcdNsxtRouteAdvertisementDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtRouteAdvertisementImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T Edge Gateway ID in which route advertisement is located",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Define if route advertisement is active",
			},
			"subnets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Set of subnets that will be advertised to Tier-0 gateway. Leaving it empty means none",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdNsxtRouteAdvertisementCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	var subnets []string
	enableRouteAdvertisement := d.Get("enabled").(bool)
	subnetsFromSchema, subnetsArgumentSet := d.GetOk("subnets")

	if subnetsArgumentSet {
		subnets = convertSchemaSetToSliceOfStrings(subnetsFromSchema.(*schema.Set))
	}

	if !enableRouteAdvertisement && len(subnets) > 0 {
		return diag.Errorf("if enable is set to false, no subnets must be passed")
	}

	_, edgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "route advertisement")
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = edgeGateway.UpdateNsxtRouteAdvertisement(enableRouteAdvertisement, subnets, true)
	if err != nil {
		return diag.Errorf("error when creating/updating route advertisement - %s", err)
	}

	d.SetId(edgeGateway.EdgeGateway.ID)

	return resourceVcdNsxtRouteAdvertisementRead(ctx, d, meta)
}

func resourceVcdNsxtRouteAdvertisementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	routeAdvertisement, err := nsxtEdge.GetNsxtRouteAdvertisement(true)
	if err != nil {
		return diag.Errorf("error while retrieving route advertisement - %s", err)
	}

	dSet(d, "enabled", routeAdvertisement.Enable)

	subnetSet := convertStringsToTypeSet(routeAdvertisement.Subnets)
	err = d.Set("subnets", subnetSet)
	if err != nil {
		return diag.Errorf("error while setting subnets argument")
	}

	return nil
}

func resourceVcdNsxtRouteAdvertisementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	orgName := d.Get("org").(string)
	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	err = nsxtEdge.DeleteNsxtRouteAdvertisement(true)
	if err != nil {
		return diag.Errorf("error while deleting route advertisement - %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdNsxtRouteAdvertisementImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway Firewall Rule import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)

	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}
