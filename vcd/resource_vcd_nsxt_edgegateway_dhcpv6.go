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

func resourceVcdNsxtEdgegatewayDhcpV6() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgegatewayDhcpV6Create,
		ReadContext:   resourceVcdNsxtEdgegatewayDhcpV6Read,
		UpdateContext: resourceVcdNsxtEdgegatewayDhcpV6Update,
		DeleteContext: resourceVcdNsxtEdgegatewayDhcpV6Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayDhcpV6Import,
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
				Description: "Edge gateway ID for Rate limiting (DHCPv6) configuration",
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "DHCPv6 configuration mode. One of 'SLAAC', 'DHCPv6', 'DISABLED'",
				ValidateFunc: validation.StringInSlice([]string{"SLAAC", "DHCPv6", "DISABLED"}, true),
			},
			"domain_names": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of domain names (only applicable for 'SLAAC' mode)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dns_servers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of DNS Servers (only applicable for 'SLAAC' mode)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdNsxtEdgegatewayDhcpV6Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDhcpV6CreateUpdate(ctx, d, meta, "create")
}

func resourceVcdNsxtEdgegatewayDhcpV6Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDhcpV6CreateUpdate(ctx, d, meta, "update")
}

func resourceVcdNsxtEdgegatewayDhcpV6CreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] %s", origin, err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error retrieving Edge Gateway: %s", origin, err)
	}

	dhcpv6Config, err := getNsxtEdgeGatewaySlaacProfileType(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error getting DHCPv6 configuration: %s", origin, err)
	}

	_, err = nsxtEdge.UpdateSlaacProfile(dhcpv6Config)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) %s] error updating DHCPv6 configuration: %s", origin, err)
	}

	d.SetId(edgeGatewayId)

	return resourceVcdNsxtEdgegatewayDhcpV6Read(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayDhcpV6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("[dhcpv6 (SLAAC Profile) read] error retrieving NSX-T Edge Gateway DHCPv6 (SLAAC Profile): %s", err)
	}

	slaacProfile, err := nsxtEdge.GetSlaacProfile()
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) read] error retrieving NSX-T Edge Gateway DHCPv6 (SLAAC Profile): %s", err)
	}

	err = setNsxtEdgeGatewaySlaacProfileData(d, slaacProfile)
	if err != nil {
		return diag.Errorf("error storing state: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDhcpV6Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] error retrieving Edge Gateway: %s", err)
	}

	// Disabling DHCPv6 configuration requires at least Mode field
	_, err = nsxtEdge.UpdateSlaacProfile(&types.NsxtEdgeGatewaySlaacProfile{Mode: "DISABLED", Enabled: false})
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) delete] error updating DHCPv6 Profile: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDhcpV6Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway DHCPv6 import initiated")

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

func getNsxtEdgeGatewaySlaacProfileType(d *schema.ResourceData) (*types.NsxtEdgeGatewaySlaacProfile, error) {
	// The API has an ugly behavior that it uses two fields for disabling the service -
	// `mode=DISABLED` and `enabled=false`
	// Asking a user to match both fields looks to be inconvenient therefore we rely on `mode` field
	// and set Enabled to `true` whenever `mode=DISABLED`

	mode := d.Get("mode").(string)

	slaacProfile := &types.NsxtEdgeGatewaySlaacProfile{
		Enabled: mode != "DISABLED", // whenever mode != DISABLED - the service is enabled
		Mode:    d.Get("mode").(string),
	}

	dnsServers := convertSchemaSetToSliceOfStrings(d.Get("dns_servers").(*schema.Set))
	domainNames := convertSchemaSetToSliceOfStrings(d.Get("domain_names").(*schema.Set))

	if len(dnsServers) > 0 || len(domainNames) > 0 {
		slaacProfile.DNSConfig = types.NsxtEdgeGatewaySlaacProfileDNSConfig{
			DNSServerIpv6Addresses: dnsServers,
			DomainNames:            domainNames,
		}
	}

	return slaacProfile, nil
}

func setNsxtEdgeGatewaySlaacProfileData(d *schema.ResourceData, slaacProfile *types.NsxtEdgeGatewaySlaacProfile) error {
	// dSet(d, "enabled", slaacProfile.Enabled)
	dSet(d, "mode", slaacProfile.Mode)

	dnsServerSet := convertStringsToTypeSet(slaacProfile.DNSConfig.DNSServerIpv6Addresses)
	err := d.Set("dns_servers", dnsServerSet)
	if err != nil {
		return fmt.Errorf("error while setting 'dns_servers': %s", err)
	}

	subnetSet := convertStringsToTypeSet(slaacProfile.DNSConfig.DomainNames)
	err = d.Set("domain_names", subnetSet)
	if err != nil {
		return fmt.Errorf("error while setting 'domain_names': %s", err)
	}

	return nil
}
