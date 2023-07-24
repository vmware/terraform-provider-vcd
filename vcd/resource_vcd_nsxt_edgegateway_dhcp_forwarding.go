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

func resourceVcdNsxtEdgegatewayDhcpForwarding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgegatewayDhcpForwardingCreate,
		UpdateContext: resourceVcdNsxtEdgegatewayDhcpForwardingUpdate,
		ReadContext:   resourceVcdNsxtEdgegatewayDhcpForwardingRead,
		DeleteContext: resourceVcdNsxtEdgegatewayDhcpForwardingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayDhcpForwardingImport,
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
				Description: "Edge gateway ID for DHCP forwarding configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Status of DHCP Forwarding for the Edge Gateway",
			},
			"dhcp_servers": {
				Type:     schema.TypeSet,
				Required: true,
				// DHCP forwarding supports up to 8 IP addresses
				MaxItems:    8,
				Description: "IP addresses of the DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdNsxtEdgegatewayDhcpForwardingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDhcpForwardingCreateUpdate(ctx, d, meta, "create")
}

func resourceVcdNsxtEdgegatewayDhcpForwardingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDhcpForwardingCreateUpdate(ctx, d, meta, "update")
}

func resourceVcdNsxtEdgegatewayDhcpForwardingCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, method string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[DHCP forwarding %s] %s", method, err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[DHCP forwarding %s] error retrieving Edge Gateway: %s", method, err)
	}

	dhcpForwardConfig := &types.NsxtEdgeGatewayDhcpForwarder{
		Enabled:     d.Get("enabled").(bool),
		DhcpServers: convertSchemaSetToSliceOfStrings(d.Get("dhcp_servers").(*schema.Set)),
	}

	_, err = nsxtEdge.UpdateDhcpForwarder(dhcpForwardConfig)
	if err != nil {
		return diag.Errorf("[DHCP forwarding %s] error updating DHCP forwarding configuration: %s", method, err)
	}

	d.SetId(edgeGatewayId)

	var diags diag.Diagnostics
	if !dhcpForwardConfig.Enabled && d.HasChange("dhcp_servers") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "DHCP forwarding IP addresses will not be changed if the service is disabled",
		})
	}

	// As there may be warnings in the CreateUpdate function, we need to append them
	// to the read function, as we don't want to exit the program if there is only
	// a warning.
	return append(diags, resourceVcdNsxtEdgegatewayDhcpForwardingRead(ctx, d, meta)...)
}

func resourceVcdNsxtEdgegatewayDhcpForwardingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxtEdgegatewayDhcpForwardingRead(ctx, d, meta, "resource")
}

func genericVcdNsxtEdgegatewayDhcpForwardingRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if govcd.ContainsNotFound(err) {
		// When parent Edge Gateway is not found - this resource is also not found and should be
		// removed from state
		if origin == "datasource" {
			return diag.Errorf("[DHCP forwarding DS read] error retrieving NSX-T Edge Gateway DHCP forwarding: %s", err)
		}
		d.SetId("")
		log.Printf("[DEBUG] Edge gateway no longer exists. Removing from tfstate")
		return nil
	}

	if err != nil {
		return diag.Errorf("[DHCP forwarding read] error: %s", err)
	}

	dhcpForwardConfig, err := nsxtEdge.GetDhcpForwarder()
	if err != nil {
		return diag.Errorf("[DHCP forwarding read] error retrieving NSX-T Edge Gateway DHCP forwarding: %s", err)
	}

	// DHCP forwarding does not have its own ID - it is a part of Edge Gateway
	d.SetId(edgeGatewayId)
	dSet(d, "enabled", dhcpForwardConfig.Enabled)

	err = d.Set("dhcp_servers", convertStringsToTypeSet(dhcpForwardConfig.DhcpServers))
	if err != nil {
		return diag.Errorf("error setting dhcp_servers attribute: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDhcpForwardingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[DHCP forwarding delete] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[DHCP forwarding delete] error retrieving Edge Gateway: %s", err)
	}

	// There is no "delete" for DHCP forwarding. It can only be updated to empty values (disabled)
	_, err = nsxtEdge.UpdateDhcpForwarder(&types.NsxtEdgeGatewayDhcpForwarder{})
	if err != nil {
		return diag.Errorf("[DHCP forwarding delete] error updating DHCP forwarding: %s", err)
	}

	return nil
}

// resourceVcdNsxtEdgegatewayDhcpForwardingImport imports DHCP forwarding configuration for NSX-T
// Edge Gateway.
// The import path for this resource is Edge Gateway. ID of the field is also Edge Gateway ID as
// DHCP forwarding is a property of Edge Gateway, not a separate entity.
func resourceVcdNsxtEdgegatewayDhcpForwardingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway DHCP forwarding import initiated")

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
