package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdEdgeGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdEdgeGatewayCreate,
		Delete: resourceVcdEdgeGatewayDelete,
		Read:   resourceVcdEdgeGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vdc": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"delete_force": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "TBD: no description in the API, and undefined behavior in practice",
			},
			"delete_recursive": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "TBD: no description in the API, and undefined behavior in practice",
			},
			"advanced_networking": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "True if the gateway uses advanced networking",
			},
			"backing_config": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Configuration of the vShield edge VM for this gateway. One of: compact, full",
			},
			"ha_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Enable high availability on this edge gateway",
			},
			"external_networks": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "A list of external networks to be used by the edge gateway",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "External network to be used as default gateway. Its name must be included in 'external_networks'. An empty value will skip the default gateway",
			},
			"distributed_routing": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If advanced networking enabled, also enable distributed routing",
			},
		},
	}
}

// Creates a new edge gateway from a resource definition
func resourceVcdEdgeGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway creation initiated")

	vcdClient := meta.(*VCDClient)

	rawExternalNetworks := d.Get("external_networks").([]interface{})
	var externalNetworks []string
	for _, en := range rawExternalNetworks {
		externalNetworks = append(externalNetworks, en.(string))
	}

	var gwCreation = govcd.EdgeGatewayCreation{
		ExternalNetworks:          externalNetworks,
		Name:                      d.Get("name").(string),
		OrgName:                   d.Get("org").(string),
		VdcName:                   d.Get("vdc").(string),
		Description:               d.Get("description").(string),
		BackingConfiguration:      d.Get("backing_config").(string),
		AdvancedNetworkingEnabled: d.Get("advanced_networking").(bool),
		DefaultGateway:            d.Get("default_gateway").(string),
		DistributedRoutingEnabled: d.Get("distributed_routing").(bool),
		HAEnabled:                 d.Get("ha_enabled").(bool),
	}

	edge, err := govcd.CreateEdgeGateway(vcdClient.VCDClient, gwCreation)
	if err != nil {
		log.Printf("[DEBUG] Error creating edge gateway: %#v", err)
		return fmt.Errorf("error creating edge gateway: %#v", err)
	}

	d.SetId(d.Get("name").(string))
	log.Printf("[TRACE] edge gateway created: %#v", edge.EdgeGateway.Name)
	return resourceVcdEdgeGatewayRead(d, meta)
}

func setEdgeGatewayValues(d *schema.ResourceData, egw govcd.EdgeGateway) error {

	d.SetId(egw.EdgeGateway.HREF)
	err := d.Set("name", egw.EdgeGateway.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", egw.EdgeGateway.Description)
	if err != nil {
		return err
	}
	err = d.Set("backing_config", egw.EdgeGateway.Configuration.GatewayBackingConfig)
	if err != nil {
		return err
	}
	var networks []string
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		networks = append(networks, net.Network.Name)
	}

	err = d.Set("external_networks", networks)
	if err != nil {
		return err
	}
	return nil
}

// Fetches information about an existing edge gateway for a data definition
func resourceVcdEdgeGatewayRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway read initiated")

	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error fetching edge gateway details %#v", err)
	}

	err = setEdgeGatewayValues(d, edgeGateway)

	log.Printf("[TRACE] edge gateway read completed: %#v", edgeGateway.EdgeGateway)
	return err
}

// Deletes a edge gateway, optionally removing all objects in it as well
func resourceVcdEdgeGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway delete started")

	vcdClient := meta.(*VCDClient)

	force := d.Get("delete_force").(bool)
	recursive := d.Get("delete_recursive").(bool)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error fetching edge gateway details %#v", err)
	}

	err = edgeGateway.Delete(force, recursive)

	log.Printf("[TRACE] external edge gateway completed\n")
	return err
}
