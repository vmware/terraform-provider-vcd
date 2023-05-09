package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtEdgegatewayRateLimiting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgegatewayRateLimitingCreateUpdate,
		UpdateContext: resourceVcdNsxtEdgegatewayRateLimitingCreateUpdate,
		ReadContext:   resourceVcdNsxtEdgegatewayRateLimitingRead,
		DeleteContext: resourceVcdNsxtEdgegatewayRateLimitingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayRateLimitingImport,
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
				Description: "Edge gateway ID for Rate limiting (QoS) configuration",
			},
			"ingress_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Ingress profile ID for Rate limiting (QoS) configuration",
			},
			"egress_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Egress profile ID for Rate limiting (QoS) configuration",
			},
		},
	}
}

func resourceVcdNsxtEdgegatewayRateLimitingCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) create/update] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) create/update] error retrieving Edge Gateway: %s", err)
	}

	qosConfig, err := getNsxtEdgeGatewayRateLimitingType(d)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) create/update] error getting QoS configuration: %s", err)
	}

	_, err = nsxtEdge.UpdateQoS(qosConfig)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) create/update] error updating QoS configuration: %s", err)
	}

	d.SetId(edgeGatewayId)

	return resourceVcdNsxtEdgegatewayRateLimitingRead(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayRateLimitingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			// When parent Edge Gateway is not found - this resource is also not found and should be
			// removed from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("[rate limiting (QoS) read] error retrieving NSX-T Edge Gateway rate limiting (QoS): %s", err)
	}

	qosConfig, err := nsxtEdge.GetQoS()
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) read] error retrieving NSX-T Edge Gateway rate limiting (QoS): %s", err)
	}

	setNsxtEdgeGatewayRateLimitingData(d, qosConfig)

	return nil
}

func resourceVcdNsxtEdgegatewayRateLimitingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) delete] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) delete] error retrieving Edge Gateway: %s", err)
	}

	// There is no real "delete" for QoS. It can only be updated to empty values (unlimited)
	_, err = nsxtEdge.UpdateQoS(&types.NsxtEdgeGatewayQos{})
	if err != nil {
		return diag.Errorf("[rate limiting (QoS) delete] error updating QoS Profile: %s", err)
	}

	return nil
}

// resourceVcdNsxtEdgegatewayRateLimitingImport imports rate limiting (QoS) configuration for NSX-T
// Edge Gateway.
// The import path for this resource is Edge Gateway. ID of the field is also Edge Gateway ID as
// rate limiting is a property of Edge Gateway, not a separate entity.
func resourceVcdNsxtEdgegatewayRateLimitingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway Rate limiting (QoS) import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name or org-name.vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)

	// Storing Edge Gateway ID and Read will retrieve all other data
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtEdgeGatewayRateLimitingType(d *schema.ResourceData) (*types.NsxtEdgeGatewayQos, error) {
	qosType := &types.NsxtEdgeGatewayQos{}
	ingressProfileId := d.Get("ingress_profile_id").(string)
	egressProfileId := d.Get("egress_profile_id").(string)

	if ingressProfileId != "" {
		qosType.IngressProfile = &types.OpenApiReference{
			ID: ingressProfileId,
		}
	}

	if egressProfileId != "" {
		qosType.EgressProfile = &types.OpenApiReference{
			ID: egressProfileId,
		}
	}

	return qosType, nil
}

func setNsxtEdgeGatewayRateLimitingData(d *schema.ResourceData, qosType *types.NsxtEdgeGatewayQos) {
	if qosType.IngressProfile != nil {
		dSet(d, "ingress_profile_id", qosType.IngressProfile.ID)
	} else {
		dSet(d, "ingress_profile_id", "") // Empty means `unlimited`
	}

	if qosType.EgressProfile != nil {
		dSet(d, "egress_profile_id", qosType.EgressProfile.ID)
	} else {
		dSet(d, "egress_profile_id", "") // Empty means `unlimited`
	}
}
