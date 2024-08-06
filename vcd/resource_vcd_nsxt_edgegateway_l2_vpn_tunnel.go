package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtEdgegatewayL2VpnTunnel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgegatewayL2VpnTunnelCreate,
		ReadContext:   resourceVcdNsxtEdgegatewayL2VpnTunnelRead,
		UpdateContext: resourceVcdNsxtEdgegatewayL2VpnTunnelUpdate,
		DeleteContext: resourceVcdNsxtEdgegatewayL2VpnTunnelDestroy,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayL2VpnTunnelImport,
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
				Description: "Edge Gateway ID for the tunnel",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the L2 VPN Tunnel session",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the L2 VPN Tunnel session",
			},
			"session_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"CLIENT", "SERVER"}, false),
				ForceNew:     true,
				Description:  "Mode of the tunnel session, must be CLIENT or SERVER",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Status of the L2 VPN Tunnel session. Always set to `true` for CLIENT sessions. Defaults to true.",
			},
			"local_endpoint_ip": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Local endpoint IP of the tunnel session, the IP must be sub-allocated to the Edge Gateway",
				ValidateFunc: validation.IsIPAddress,
			},
			"remote_endpoint_ip": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The IP address of the remote endpoint, which corresponds to the device" +
					"on the remote site terminating the VPN tunnel.",
				ValidateFunc: validation.IsIPAddress,
			},
			"tunnel_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Network CIDR block over which the session interfaces. Only relevant if " +
					"`session_mode` is set to `SERVER`",
				ValidateFunc: validation.IsCIDR,
			},
			"connector_initiation_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Connector initation mode of the session describing how a connection is made. " +
					"Needs to be set only if `session_mode` is set to `SERVER`",
				ValidateFunc: validation.StringInSlice([]string{"INITIATOR", "RESPOND_ONLY", "ON_DEMAND"}, false),
			},
			"pre_shared_key": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Pre-shared key used for authentication, needs to be provided only for" +
					"`SERVER` sessions.",
			},
			"peer_code": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Base64 encoded string of the full configuration of the tunnel provided by the SERVER session. " +
					"It is a computed field for SERVER sessions and is a required field for CLIENT sessions.",
			},
			"stretched_network": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Org VDC networks that are attached to the L2 VPN tunnel",
				Elem:        stretchedNetwork,
			},
		},
	}
}

var stretchedNetwork = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"network_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ID of the Org VDC network",
		},
		"tunnel_id": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			Description:  "Tunnel ID of the network for the tunnel. Read-only for `SERVER` sessions.",
			ValidateFunc: validation.IntBetween(1, 4093),
		},
	},
}

func resourceVcdNsxtEdgegatewayL2VpnTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel create] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel create] error retrieving Edge Gateway: %s", err)
	}

	tunnelConfig, err := readL2VpnTunnelFromSchema(d)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel create] %s", err)
	}
	tunnel, err := nsxtEdge.CreateL2VpnTunnel(tunnelConfig)
	if err != nil {
		if strings.Contains(err.Error(), "or the target entity is invalid") {
			if err2 := doesNotWorkWithDistributedOnlyEdgeGateway("vcd_nsxt_edgegateway_l2_vpn_tunnel", vcdClient, nsxtEdge); err2 != nil {
				return diag.Errorf(err.Error() + "\n\n" + err2.Error())
			}
		}
		return diag.Errorf("[L2 VPN Tunnel create] error creating L2 VPN Tunnel: %s", err)
	}
	d.SetId(tunnel.NsxtL2VpnTunnel.ID)

	return resourceVcdNsxtEdgegatewayL2VpnTunnelRead(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayL2VpnTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericNsxtEdgegatewayL2VpnTunnelRead(ctx, d, meta, "resource")
}

func genericNsxtEdgegatewayL2VpnTunnelRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)
	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel read] error retrieving Edge Gateway: %s", err)
	}

	tunnelName := d.Get("name").(string)
	tunnelId := d.Id()
	if origin == "datasource" {
		tunnel, err := nsxtEdge.GetL2VpnTunnelByName(tunnelName)
		if err != nil {
			return diag.Errorf("[L2 VPN Tunnel DS read] error retrieving L2 VPN Tunnel: %s", err)
		}
		// We need to first read by name then by ID for data sources, as GET by name doesn't return all the information
		// (peer_code, pre_shared_key)
		tunnelId = tunnel.NsxtL2VpnTunnel.ID
	}

	tunnelConfig, err := nsxtEdge.GetL2VpnTunnelById(tunnelId)
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] L2 VPN Tunnel no longer exists. Removing from tfstate")
		return nil
	}
	err = readL2VpnTunnelToSchema(tunnelConfig.NsxtL2VpnTunnel, d, vcdClient)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel read] error reading retrieved tunnel into schema: %s", err)
	}
	return nil
}

func resourceVcdNsxtEdgegatewayL2VpnTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel update] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)
	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel update] error retrieving Edge Gateway: %s", err)
	}

	tunnel, err := nsxtEdge.GetL2VpnTunnelById(d.Id())
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel update] error retrieving L2 VPN Tunnel: %s", err)
	}

	tunnelUpdatedConfig, err := readL2VpnTunnelFromSchema(d)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel update] error reading L2 VPN Tunnel config from schema: %s", err)
	}

	_, err = tunnel.Update(tunnelUpdatedConfig)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel update] error updating L2 VPN Tunnel: %s", err)
	}

	return resourceVcdNsxtEdgegatewayL2VpnTunnelRead(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayL2VpnTunnelDestroy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel destroy] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)
	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel destroy] error retrieving Edge Gateway: %s", err)
	}

	tunnel, err := nsxtEdge.GetL2VpnTunnelById(d.Id())
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel destroy] error retrieving L2 VPN Tunnel: %s", err)
	}

	// There is an unexpected error in all versions of VCD up to 10.5.0, it happens
	// when a L2 VPN Tunnel is created in CLIENT mode, has atleast one Org VDC
	// network attached, and is updated in any way. After that, to delete the tunnel,
	// one needs to send a DELETE request the amount of times the tunnel was updated
	// or de-attach the Org Networks from the Tunnel and send the DELETE request
	//
	// De-attach all the networks and update the Tunnel
	tunnel.NsxtL2VpnTunnel.StretchedNetworks = nil
	updatedTunnel, err := tunnel.Update(tunnel.NsxtL2VpnTunnel)
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel destroy] error de-attaching networks: %s", err)
	}

	err = updatedTunnel.Delete()
	if err != nil {
		return diag.Errorf("[L2 VPN Tunnel destroy] error deleting L2 VPN Tunnel: %s", err)
	}

	return nil
}

func readL2VpnTunnelFromSchema(d *schema.ResourceData) (*types.NsxtL2VpnTunnel, error) {
	id := d.Id()
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	sessionMode := d.Get("session_mode").(string)
	enabled := d.Get("enabled").(bool)
	localEndpointIp := d.Get("local_endpoint_ip").(string)
	remoteEndpointIp := d.Get("remote_endpoint_ip").(string)
	tunnelInterface := d.Get("tunnel_interface").(string)
	connectorInitiationMode := d.Get("connector_initiation_mode").(string)
	preSharedKey := d.Get("pre_shared_key").(string)
	peerCode := d.Get("peer_code").(string)
	stretchedNetworksSet := d.Get("stretched_network").(*schema.Set)
	stretchedNetworks := make([]types.EdgeL2VpnStretchedNetwork, len(stretchedNetworksSet.List()))
	for rangeIndex, network := range stretchedNetworksSet.List() {
		networkDefinition := network.(map[string]interface{})
		oneNetwork := types.EdgeL2VpnStretchedNetwork{
			NetworkRef: types.OpenApiReference{
				ID: networkDefinition["network_id"].(string),
			},
		}
		// SERVER sessions auto-assign tunnel IDs
		if sessionMode == "CLIENT" {
			if networkDefinition["tunnel_id"].(int) == 0 {
				return nil, fmt.Errorf("tunnel ID must be set for CLIENT sessions")
			}
			oneNetwork.TunnelID = networkDefinition["tunnel_id"].(int)
		}
		stretchedNetworks[rangeIndex] = oneNetwork
	}

	tunnel := &types.NsxtL2VpnTunnel{
		ID:                id,
		Name:              name,
		Description:       description,
		SessionMode:       sessionMode,
		LocalEndpointIp:   localEndpointIp,
		RemoteEndpointIp:  remoteEndpointIp,
		StretchedNetworks: stretchedNetworks,
	}

	// Server and Client tunnel sessions require and provide different parameters
	if sessionMode == "SERVER" {
		tunnel.TunnelInterface = tunnelInterface
		tunnel.Enabled = enabled
		if connectorInitiationMode == "" {
			return nil, fmt.Errorf("connector initiation mode must be set for `SERVER` sessions")
		}
		tunnel.ConnectorInitiationMode = connectorInitiationMode
		if preSharedKey == "" {
			return nil, fmt.Errorf("pre-shared key must be set for `SERVER` sessions")
		}
		tunnel.PreSharedKey = preSharedKey
	}

	if sessionMode == "CLIENT" {
		// There is a known bug with CLIENT mode sessions up to 10.5.0, they are always active and can't be disabled.
		if !enabled {
			return nil, fmt.Errorf("`enabled` must always be set to `true` for `CLIENT` sessions")
		}
		tunnel.Enabled = true
		if peerCode == "" {
			return nil, fmt.Errorf("peer code must be set for `CLIENT` sessions")
		}
		tunnel.PeerCode = peerCode

		// Set peer_code of CLIENT Sessions only on CREATE/UPDATE
		dSet(d, "peer_code", peerCode)
	}

	return tunnel, nil
}

func readL2VpnTunnelToSchema(tunnel *types.NsxtL2VpnTunnel, d *schema.ResourceData, vcdClient *VCDClient) error {
	d.SetId(tunnel.ID)
	dSet(d, "name", tunnel.Name)
	dSet(d, "description", tunnel.Description)
	dSet(d, "session_mode", tunnel.SessionMode)
	dSet(d, "enabled", tunnel.Enabled)
	dSet(d, "local_endpoint_ip", tunnel.LocalEndpointIp)
	dSet(d, "remote_endpoint_ip", tunnel.RemoteEndpointIp)

	if tunnel.SessionMode == "SERVER" {
		dSet(d, "tunnel_interface", tunnel.TunnelInterface)
		dSet(d, "connector_initiation_mode", tunnel.ConnectorInitiationMode)
		dSet(d, "peer_code", tunnel.PeerCode)
	}

	stretchedNetworkSlice := make([]interface{}, len(tunnel.StretchedNetworks))
	for rangeIndex, stretchedNetwork := range tunnel.StretchedNetworks {
		stretchedNetworkMap := make(map[string]interface{})
		stretchedNetworkMap["network_id"] = stretchedNetwork.NetworkRef.ID
		stretchedNetworkMap["tunnel_id"] = stretchedNetwork.TunnelID

		stretchedNetworkSlice[rangeIndex] = stretchedNetworkMap
	}

	// Hash and return stretched networks
	err := d.Set("stretched_network", schema.NewSet(schema.HashResource(stretchedNetwork), stretchedNetworkSlice))
	if err != nil {
		return err
	}

	return nil
}

func resourceVcdNsxtEdgegatewayL2VpnTunnelImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway L2 VPN Tunnel import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as " +
			"org-name.vdc-name.nsxt-edge-gw-name.l2-vpn-tunnel-name or " +
			"org-name.vdc-group-name.nsxt-edge-gw-name.l2-vpn-tunnel-name")
	}
	orgName, vdcOrVdcGroupName, edgeName, tunnelName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("edge gateway not backed by NSX-T")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", edgeName, err)
	}

	tunnel, err := edge.GetL2VpnTunnelByName(tunnelName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve L2 VPN Tunnel name with ID '%s': %s", tunnelName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)

	// Storing VPN Tunnel ID and Read will retrieve all other data
	d.SetId(tunnel.NsxtL2VpnTunnel.ID)

	return []*schema.ResourceData{d}, nil
}
