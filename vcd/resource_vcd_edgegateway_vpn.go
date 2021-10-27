package vcd

//lint:file-ignore SA1019 ignore deprecated functions
//lint:file-ignore U1000 ignore because it is a working example

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var edgeVpnLocalSubnetResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"local_subnet_name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},

		"local_subnet_gateway": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},

		"local_subnet_mask": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

var edgeVpnPeerSubnetResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"peer_subnet_name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},

		"peer_subnet_gateway": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},

		"peer_subnet_mask": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

func resourceVcdEdgeGatewayVpn() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdEdgeGatewayVpnCreate,
		Read:   resourceVcdEdgeGatewayVpnRead,
		Delete: resourceVcdEdgeGatewayVpnDelete,

		Schema: map[string]*schema.Schema{

			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"encryption_protocol": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"local_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"local_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"mtu": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"peer_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"peer_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"shared_secret": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},

			"local_subnets": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     edgeVpnLocalSubnetResource,
			},

			"peer_subnets": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     edgeVpnPeerSubnetResource,
			},
		},
	}
}

func resourceVcdEdgeGatewayVpnCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] CLIENT: %#v", vcdClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	localSubnetsList := d.Get("local_subnets").(*schema.Set).List()
	peerSubnetsList := d.Get("peer_subnets").(*schema.Set).List()

	localSubnets := make([]*types.IpsecVpnSubnet, len(localSubnetsList))
	peerSubnets := make([]*types.IpsecVpnSubnet, len(peerSubnetsList))

	for i, s := range localSubnetsList {
		ls := s.(map[string]interface{})
		localSubnets[i] = &types.IpsecVpnSubnet{
			Name:    ls["local_subnet_name"].(string),
			Gateway: ls["local_subnet_gateway"].(string),
			Netmask: ls["local_subnet_mask"].(string),
		}
	}

	for i, s := range peerSubnetsList {
		ls := s.(map[string]interface{})
		peerSubnets[i] = &types.IpsecVpnSubnet{
			Name:    ls["peer_subnet_name"].(string),
			Gateway: ls["peer_subnet_gateway"].(string),
			Netmask: ls["peer_subnet_mask"].(string),
		}
	}

	tunnel := &types.GatewayIpsecVpnTunnel{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		IpsecVpnLocalPeer: &types.IpsecVpnLocalPeer{
			ID:   "",
			Name: "",
		},
		EncryptionProtocol: d.Get("encryption_protocol").(string),
		LocalIPAddress:     d.Get("local_ip_address").(string),
		LocalID:            d.Get("local_id").(string),
		LocalSubnet:        localSubnets,
		Mtu:                d.Get("mtu").(int),
		PeerID:             d.Get("peer_id").(string),
		PeerIPAddress:      d.Get("peer_ip_address").(string),
		PeerSubnet:         peerSubnets,
		SharedSecret:       d.Get("shared_secret").(string),
		IsEnabled:          true,
	}

	tunnels := make([]*types.GatewayIpsecVpnTunnel, 1)
	tunnels[0] = tunnel

	ipsecVPNConfig := &types.EdgeGatewayServiceConfiguration{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: true,
			Tunnel:    tunnels,
		},
	}

	log.Printf("[INFO] ipsecVPNConfig: %#v", ipsecVPNConfig)

	err = edgeGateway.Refresh()
	if err != nil {
		log.Printf("[INFO] Error refreshing edge gateway: %#v", err)
		return fmt.Errorf("error refreshing edge gateway: %#v", err)
	}
	task, err := edgeGateway.AddIpsecVPN(ipsecVPNConfig)
	if err != nil {
		log.Printf("[INFO] Error setting ipsecVPNConfig rules: %s", err)
		return fmt.Errorf("error setting ipsecVPNConfig rules: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf(errorCompletingTask, err)
	}

	d.SetId(d.Get("edge_gateway").(string))

	return resourceVcdEdgeGatewayVpnRead(d, meta)
}

func resourceVcdEdgeGatewayVpnDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	log.Printf("[TRACE] CLIENT: %#v", vcdClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	ipsecVPNConfig := &types.EdgeGatewayServiceConfiguration{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: false,
		},
	}

	log.Printf("[INFO] ipsecVPNConfig: %#v", ipsecVPNConfig)

	err = edgeGateway.Refresh()
	if err != nil {
		log.Printf("[INFO] Error refreshing edge gateway: %#v", err)
		return fmt.Errorf("error refreshing edge gateway: %#v", err)
	}
	task, err := edgeGateway.AddIpsecVPN(ipsecVPNConfig)
	if err != nil {
		log.Printf("[INFO] Error setting ipsecVPNConfig rules: %s", err)
		return fmt.Errorf("error setting ipsecVPNConfig rules: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf(errorCompletingTask, err)
	}

	d.SetId(d.Get("edge_gateway").(string))

	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	return nil
}

func resourceVcdEdgeGatewayVpnRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	egsc := edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.GatewayIpsecVpnService

	if len(egsc.Tunnel) == 0 {
		d.SetId("")
		return nil
	}

	if len(egsc.Tunnel) == 1 {
		tunnel := egsc.Tunnel[0]
		dSet(d, "name", tunnel.Name)
		dSet(d, "description", tunnel.Description)
		dSet(d, "encryption_protocol", tunnel.EncryptionProtocol)
		dSet(d, "local_ip_address", tunnel.LocalIPAddress)
		dSet(d, "local_id", tunnel.LocalID)
		dSet(d, "mtu", tunnel.Mtu)
		dSet(d, "peer_ip_address", tunnel.PeerIPAddress)
		dSet(d, "peer_id", tunnel.PeerID)

		// Read for local_subnets and peer_subnets never worked and it is impossible to fix it with current resource
		// design because not all data can be retrieved. Detailed explanation and code demonstrating read problems is
		// below;.
		//
		// local_subnets and peer_subnets cannot be read because API does not return all values that are sent
		// Also some values are returned different than sent.
		// Example POST:
		// <LocalSubnet>
		//    <Name>WEB_EAST</Name>
		//    <Gateway>10.150.192.1</Gateway>
		//    <Netmask>255.255.255.0</Netmask>
		// </LocalSubnet>
		// Example GET after that:
		// 	<LocalSubnet>
		//    <Name>10.150.192.0/24</Name>
		//    <Gateway>10.150.192.0</Gateway>
		//    <Netmask>255.255.255.0</Netmask>
		//  </LocalSubnet>
		//
		// In the example above - one can see that only Netmask is returned in the same format as sent. Name is ignored
		// at all. Because it is TypeSet - we cannot set "partial" data as it would immediately cause Diff.

		// Uncomment code below to enable "READ" functionality and witness problems described above.
		//
		// err := convertAndSet("local_subnets", "local", edgeVpnLocalSubnetResource, tunnel.LocalSubnet, d)
		// if err != nil {
		// 	return fmt.Errorf("error setting 'local_subnets': %s", err)
		// }
		// err = convertAndSet("peer_subnets", "peer", edgeVpnPeerSubnetResource, tunnel.PeerSubnet, d)
		// if err != nil {
		// 	return fmt.Errorf("error setting 'peer_subnets': %s", err)
		// }

	} else {
		return fmt.Errorf("multiple tunnels not currently supported")
	}

	return nil
}

func convertAndSet(key, prefix string, hashObejct *schema.Resource, subNets []*types.IpsecVpnSubnet, d *schema.ResourceData) error {
	var items []interface{}

	for _, subNet := range subNets {
		item := map[string]interface{}{
			prefix + "_subnet_name":    subNet.Name,
			prefix + "_subnet_gateway": subNet.Gateway,
			prefix + "_subnet_mask":    subNet.Netmask,
		}
		items = append(items, item)
	}

	itemsHash := schema.NewSet(schema.HashResource(hashObejct), items)

	return d.Set(key, itemsHash)
}
