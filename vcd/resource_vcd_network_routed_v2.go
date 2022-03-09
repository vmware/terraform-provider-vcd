package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNetworkRoutedV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNetworkRoutedV2Create,
		ReadContext:   resourceVcdNetworkRoutedV2Read,
		UpdateContext: resourceVcdNetworkRoutedV2Update,
		DeleteContext: resourceVcdNetworkRoutedV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNetworkRoutedV2Import,
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
				Deprecated:  "'vdc' is deprecated and ineffective. Routed networks will inherit VDC setting from parent Edge Gateway",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of VDC or VDC Group",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID in which Routed network should be located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network description",
			},
			"interface_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "internal",
				Description:      "Optional interface type (only for NSX-V networks). One of 'INTERNAL' (default), 'DISTRIBUTED', 'SUBINTERFACE'",
				ValidateFunc:     validation.StringInSlice([]string{"internal", "subinterface", "distributed"}, true),
				DiffSuppressFunc: suppressCase,
			},
			"gateway": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Gateway IP address",
			},
			"prefix_length": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Network prefix",
			},
			"dns1": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS server 1",
			},
			"dns2": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS server 1",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS suffix",
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRange,
			},
		},
	}
}

func resourceVcdNetworkRoutedV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error retrieving Org: %s", err)
	}

	networkType, err := getOpenApiOrgVdcNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	orgNetwork, err := org.CreateOpenApiOrgVdcNetwork(networkType)
	if err != nil {
		return diag.Errorf("[routed network create v2] error creating Org VDC routed network: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return resourceVcdNetworkRoutedV2Read(ctx, d, meta)
}

func resourceVcdNetworkRoutedV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found -
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[routed network update v2] error getting Org VDC network: %s", err)
	}

	networkType, err := getOpenApiOrgVdcNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	// Explicitly add ID to the new type because function `getOpenApiOrgVdcNetworkType` only sets other fields
	networkType.ID = d.Id()

	_, err = orgNetwork.Update(networkType)
	if err != nil {
		return diag.Errorf("[routed network update v2] error updating Org VDC network: %s", err)
	}

	return resourceVcdNetworkRoutedV2Read(ctx, d, meta)
}

func resourceVcdNetworkRoutedV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found - unset ID
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[routed network read v2] error getting Org VDC network: %s", err)
	}

	err = setOpenApiOrgVdcNetworkData(d, orgNetwork.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[routed network read v2] error setting Org VDC network data: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return nil
}

func resourceVcdNetworkRoutedV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[routed network delete v2] error getting Org VDC network: %s", err)
	}

	err = orgNetwork.Delete()
	if err != nil {
		return diag.Errorf("[routed network delete v2] error deleting Org VDC network: %s", err)
	}

	return nil
}

func resourceVcdNetworkRoutedV2Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[routed network import v2] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]
	vcdClient := meta.(*VCDClient)

	// define an interface type to match VDC and VDC Groups
	var vdcOrGroup vdcOrVdcGroupVerifier
	_, vdcOrGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if govcd.ContainsNotFound(err) {
		adminOrg, err := vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrGroup, err = adminOrg.GetVdcGroupByName(vdcName)
		if err != nil {
			return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s': %s", vdcName, err)
		}
	}

	orgNetwork, err := vdcOrGroup.GetOpenApiOrgVdcNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org VDC network '%s': %s", networkName, err)
	}

	if !orgNetwork.IsRouted() {
		return nil, fmt.Errorf("[routed network import v2] Org network with name '%s' found, but is not of type Routed (type is '%s')",
			networkName, orgNetwork.GetType())
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func setOpenApiOrgVdcNetworkData(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {
	dSet(d, "name", orgVdcNetwork.Name)
	dSet(d, "description", orgVdcNetwork.Description)
	dSet(d, "owner_id", orgVdcNetwork.OwnerRef.ID)
	dSet(d, "vdc", orgVdcNetwork.OwnerRef.Name)

	if orgVdcNetwork.Connection != nil {
		dSet(d, "edge_gateway_id", orgVdcNetwork.Connection.RouterRef.ID)
		dSet(d, "interface_type", orgVdcNetwork.Connection.ConnectionType)
	}

	// Only one subnet can be defined although the structure accepts slice
	dSet(d, "gateway", orgVdcNetwork.Subnets.Values[0].Gateway)
	dSet(d, "prefix_length", orgVdcNetwork.Subnets.Values[0].PrefixLength)
	dSet(d, "dns1", orgVdcNetwork.Subnets.Values[0].DNSServer1)
	dSet(d, "dns2", orgVdcNetwork.Subnets.Values[0].DNSServer2)
	dSet(d, "dns_suffix", orgVdcNetwork.Subnets.Values[0].DNSSuffix)

	// If any IP sets are available
	if len(orgVdcNetwork.Subnets.Values[0].IPRanges.Values) > 0 {
		ipRangeSlice := make([]interface{}, len(orgVdcNetwork.Subnets.Values[0].IPRanges.Values))
		for index, ipRange := range orgVdcNetwork.Subnets.Values[0].IPRanges.Values {
			ipRangeMap := make(map[string]interface{})
			ipRangeMap["start_address"] = ipRange.StartAddress
			ipRangeMap["end_address"] = ipRange.EndAddress

			ipRangeSlice[index] = ipRangeMap
		}
		ipRangeSet := schema.NewSet(schema.HashResource(networkV2IpRange), ipRangeSlice)

		err := d.Set("static_ip_pool", ipRangeSet)
		if err != nil {
			return fmt.Errorf("error setting 'static_ip_pool': %s", err)
		}
	}

	return nil
}

func getOpenApiOrgVdcNetworkType(d *schema.ResourceData, vcdClient *VCDClient) (*types.OpenApiOrgVdcNetwork, error) {
	// Must get any type of Edge Gateway because this resource supports NSX-V and NSX-T Routed
	// networks. This resource must inherit OwnerRef.ID from parent Edge Gateway because when
	// migrating NSX-T Edge Gateway to/from VDC Group - routed network migrates together
	// automatically. Because of this reason it is best to avoid requiring Owner ID specification
	// for routed network at all.
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	anyEdgeGateway, err := org.GetAnyEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return nil, fmt.Errorf("error retrieving Edge Gateway structure: %s", err)
	}

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: anyEdgeGateway.EdgeGateway.OwnerRef.ID},

		NetworkType: types.OrgVdcNetworkTypeRouted,

		// Connection is used for "routed" network
		Connection: &types.Connection{
			RouterRef: types.OpenApiReference{
				ID: d.Get("edge_gateway_id").(string),
			},
			// API requires interface type in upper case, but we accept any case
			ConnectionType: strings.ToUpper(d.Get("interface_type").(string)),
		},
		Subnets: types.OrgVdcNetworkSubnets{
			Values: []types.OrgVdcNetworkSubnetValues{
				{
					Gateway:      d.Get("gateway").(string),
					PrefixLength: d.Get("prefix_length").(int),
					DNSServer1:   d.Get("dns1").(string),
					DNSServer2:   d.Get("dns2").(string),
					DNSSuffix:    d.Get("dns_suffix").(string),
					IPRanges: types.OrgVdcNetworkSubnetIPRanges{
						Values: processIpRanges(d.Get("static_ip_pool").(*schema.Set)),
					},
				},
			},
		},
	}

	return orgVdcNetworkConfig, nil
}
