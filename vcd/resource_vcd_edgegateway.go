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
				Optional: true,
				ForceNew: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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
			"advanced": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "True if the gateway uses advanced networking. (Set by default in vCD 9.7+)",
			},
			"gateway_configuration": &schema.Schema{
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

	// Making sure the parent entities are available
	orgName := vcdClient.getOrgName(d)
	vdcName := vcdClient.getVdcName(d)

	var missing []string
	if orgName == "" {
		missing = append(missing, "org")
	}
	if vdcName == "" {
		missing = append(missing, "vdc")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing properties. %v should be given either in the resource or at provider level", missing)
	}

	org, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return err
	}
	if (org == govcd.Org{}) || org.Org.HREF == "" || org.Org.ID == "" || org.Org.Name == "" {
		return fmt.Errorf("no valid Organization named '%s' was found", orgName)
	}
	if (vdc == govcd.Vdc{}) || vdc.Vdc.HREF == "" || vdc.Vdc.ID == "" || vdc.Vdc.Name == "" {
		return fmt.Errorf("no valid VDC named '%s' was found", vdcName)
	}

	var gwCreation = govcd.EdgeGatewayCreation{
		ExternalNetworks:          externalNetworks,
		Name:                      d.Get("name").(string),
		OrgName:                   orgName,
		VdcName:                   vdcName,
		Description:               d.Get("description").(string),
		BackingConfiguration:      d.Get("gateway_configuration").(string),
		AdvancedNetworkingEnabled: d.Get("advanced").(bool),
		DefaultGateway:            d.Get("default_gateway").(string),
		DistributedRoutingEnabled: d.Get("distributed_routing").(bool),
		HAEnabled:                 d.Get("ha_enabled").(bool),
	}

	// In version 9.7+ the advanced property is true by default
	if vcdClient.APIVCDMaxVersionIs(">= 32.0") {
		if !gwCreation.AdvancedNetworkingEnabled {
			return fmt.Errorf("'advanced' property for vCD 9.7+ must be set to 'true'")
		}
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

// Convenience function to fill edge gateway values from resource data
func setEdgeGatewayValues(d *schema.ResourceData, egw govcd.EdgeGateway) error {

	d.SetId(egw.EdgeGateway.ID)
	err := d.Set("name", egw.EdgeGateway.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", egw.EdgeGateway.Description)
	if err != nil {
		return err
	}
	err = d.Set("gateway_configuration", egw.EdgeGateway.Configuration.GatewayBackingConfig)
	if err != nil {
		return err
	}
	var networks []string
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if net.InterfaceType == "uplink" {
			networks = append(networks, net.Network.Name)
		}
	}
	err = d.Set("external_networks", networks)
	if err != nil {
		return err
	}

	err = d.Set("advanced", egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled)
	if err != nil {
		return err
	}
	err = d.Set("ha_enabled", egw.EdgeGateway.Configuration.HaEnabled)
	if err != nil {
		return err
	}
	err = d.Set("distributed_routing", egw.EdgeGateway.Configuration.DistributedRoutingEnabled)
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

	vcdClient.lockEdgeGateway(d)
	defer vcdClient.unlockEdgeGateway(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error fetching edge gateway details %#v", err)
	}

	err = edgeGateway.Delete(true, true)

	log.Printf("[TRACE] external edge gateway completed\n")
	return err
}
