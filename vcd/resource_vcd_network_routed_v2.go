package vcd

import (
	"context"
	"fmt"
	"log"
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
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRange,
			},
			"dual_stack_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Boolean value if Dual-Stack mode should be enabled (default `false`)",
			},
			"secondary_gateway": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "Secondary gateway (can only be IPv6 and requires enabled Dual Stack mode)",
			},
			"secondary_prefix_length": {
				Type:         schema.TypeString, // using TypeString to differentiate between 0 and no value ""
				ForceNew:     true,
				Optional:     true,
				Description:  "Secondary prefix (can only be IPv6 and requires enabled Dual Stack mode)",
				ValidateFunc: IsIntAndAtLeast(0),
			},
			"secondary_static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Secondary IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRange,
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
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Description:   "Key value map of metadata to assign to this network. Key and value can be any string",
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
			},
			"metadata_entry": metadataEntryResourceSchema("Network"),
		},
	}
}

func resourceVcdNetworkRoutedV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on a routed network is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, org, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	networkType, err := getOpenApiOrgVdcRoutedNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	orgNetwork, err := org.CreateOpenApiOrgVdcNetwork(networkType)
	if err != nil {
		return diag.Errorf("[routed network create v2] error creating Routed network: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	err = createOrUpdateOpenApiNetworkMetadata(d, orgNetwork)
	if err != nil {
		return diag.Errorf("[routed network create v2] error adding metadata to Routed network: %s", err)
	}

	return resourceVcdNetworkRoutedV2Read(ctx, d, meta)
}

func resourceVcdNetworkRoutedV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on a routed network is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, org, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[routed network create v2] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found -
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[routed network update v2] error getting Routed network: %s", err)
	}

	networkType, err := getOpenApiOrgVdcRoutedNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	// Explicitly add ID to the new type because function `getOpenApiOrgVdcNetworkType` only sets other fields
	networkType.ID = d.Id()

	_, err = orgNetwork.Update(networkType)
	if err != nil {
		return diag.Errorf("[routed network update v2] error updating Routed network: %s", err)
	}

	err = createOrUpdateOpenApiNetworkMetadata(d, orgNetwork)
	if err != nil {
		return diag.Errorf("[routed network v2 update] error updating Routed network metadata: %s", err)
	}

	return resourceVcdNetworkRoutedV2Read(ctx, d, meta)
}

func resourceVcdNetworkRoutedV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("[routed network read v2] error getting Routed network: %s", err)
	}

	err = setOpenApiOrgVdcRoutedNetworkData(d, orgNetwork.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[routed network read v2] error setting Routed network data: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	// Metadata is not supported when the network is in a VDC Group, although it is still present in the entity.
	// Hence, we skip the read to preserve its value in state.
	var diagErr diag.Diagnostics
	if !govcd.OwnerIsVdcGroup(orgNetwork.OpenApiOrgVdcNetwork.OwnerRef.ID) {
		diagErr = updateMetadataInState(d, vcdClient, "vcd_network_routed_v2", orgNetwork)
	} else if _, ok := d.GetOk("metadata"); !ok {
		// If it's a VDC Group and metadata is not set, we explicitly compute it to empty. Otherwise, its value should
		// be preserved as it is still present in the entity.
		err = d.Set("metadata", StringMap{})
		if err != nil {
			diagErr = diag.FromErr(err)
		}
	}
	if diagErr != nil {
		log.Printf("[DEBUG] Unable to set routed network v2 metadata: %s", err)
		return diagErr
	}

	return nil
}

func resourceVcdNetworkRoutedV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on a routed network is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, org, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[routed network delete v2] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[routed network delete v2] error getting Routed network: %s", err)
	}

	err = orgNetwork.Delete()
	if err != nil {
		return diag.Errorf("[routed network delete v2] error deleting Routed network: %s", err)
	}

	return nil
}

func resourceVcdNetworkRoutedV2Import(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[routed network import v2] resource name must be specified as org-name.vdc-name.network-name or org-name.vdc-group-name.network-name")
	}
	orgName, vdcOrVdcGroupName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]
	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	orgNetwork, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Routed network '%s': %s", networkName, err)
	}

	if !orgNetwork.IsRouted() {
		return nil, fmt.Errorf("[routed network import v2] Org network with name '%s' found, but is not of type Routed (type is '%s')",
			networkName, orgNetwork.GetType())
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcOrVdcGroupName)
	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func setOpenApiOrgVdcRoutedNetworkData(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {
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

	// If any IP ranges are available
	if len(orgVdcNetwork.Subnets.Values[0].IPRanges.Values) > 0 {
		err := setOpenApiOrgVdcNetworkStaticPoolData(d, orgVdcNetwork.Subnets.Values[0].IPRanges.Values, "static_ip_pool")
		if err != nil {
			return err
		}
	}

	if orgVdcNetwork.EnableDualSubnetNetwork != nil && *orgVdcNetwork.EnableDualSubnetNetwork {
		err := setSecondarySubnet(d, orgVdcNetwork)
		if err != nil {
			return fmt.Errorf("error storing Dual-Stack network to schema: %s", err)
		}
	}

	return nil
}

func getOpenApiOrgVdcRoutedNetworkType(d *schema.ResourceData, vcdClient *VCDClient) (*types.OpenApiOrgVdcNetwork, error) {
	// Must get any type of Edge Gateway because this resource supports NSX-V and NSX-T Routed
	// networks. This resource must inherit OwnerRef.ID from parent Edge Gateway because when
	// migrating NSX-T Edge Gateway to/from VDC Group - routed network migrates together
	// automatically. Because of this reason it is best to avoid requiring Owner ID specification
	// for routed network at all.
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(d.Get("edge_gateway_id").(string))
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

	// Handle Dual-Stack configuration (it accepts config address and amends it if required)
	err = getOpenApiOrgVdcSecondaryNetworkType(d, orgVdcNetworkConfig)
	if err != nil {
		return nil, err
	}

	return orgVdcNetworkConfig, nil
}

func getParentEdgeGatewayOwnerId(vcdClient *VCDClient, d *schema.ResourceData) (string, *govcd.Org, error) {
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving Org: %s", err)
	}

	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving Edge Gateway structure: %s", err)
	}

	return anyEdgeGateway.EdgeGateway.OwnerRef.ID, org, nil
}
