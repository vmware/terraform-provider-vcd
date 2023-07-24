package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNetworkIsolatedV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNetworkIsolatedV2Create,
		ReadContext:   resourceVcdNetworkIsolatedV2Read,
		UpdateContext: resourceVcdNetworkIsolatedV2Update,
		DeleteContext: resourceVcdNetworkIsolatedV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNetworkIsolatedV2Import,
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				ConflictsWith: []string{"owner_id"},
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC Group IDs",
			},
			"owner_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC or VDC Group",
				ConflictsWith: []string{"vdc"},
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
			"is_shared": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "NSX-V only - share this network with other VDCs in this organization. Default - false",
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

func resourceVcdNetworkIsolatedV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Only when a network is in VDC Group - it must lock parent VDC Group. It doesn't cause lock
	// issues when created in VDC.
	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network create v2] error retrieving Org: %s", err)
	}

	networkType, err := getOpenApiOrgVdcIsolatedNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	orgNetwork, err := org.CreateOpenApiOrgVdcNetwork(networkType)
	if err != nil {
		return diag.Errorf("[isolated network v2 create] error creating Isolated network: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	err = createOrUpdateOpenApiNetworkMetadata(d, orgNetwork)
	if err != nil {
		return diag.Errorf("[isolated network v2 create] error adding metadata to Isolated network: %s", err)
	}

	return resourceVcdNetworkIsolatedV2Read(ctx, d, meta)
}

func resourceVcdNetworkIsolatedV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// `vdc` field is deprecated. `vdc` value should not be changed unless it is removal of the
	// field at all to allow easy migration to `owner_id` path
	if _, newValue := d.GetChange("vdc"); d.HasChange("vdc") && newValue.(string) != "" {
		return diag.Errorf("changing 'vdc' field value is not supported. It can only be removed. " +
			"Please use `owner_id` field for moving network to/from VDC Group")
	}

	// Only when a network is in VDC Group - it must lock parent VDC Group. It doesn't cause lock
	// issues when created in VDC.
	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network create v2] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found -
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error getting Isolated network: %s", err)
	}

	networkType, err := getOpenApiOrgVdcIsolatedNetworkType(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	// Explicitly add ID to the new type because function `getOpenApiOrgVdcIsolatedNetworkType` only sets other fields
	networkType.ID = d.Id()

	_, err = orgNetwork.Update(networkType)
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error updating Isolated network: %s", err)
	}

	err = createOrUpdateOpenApiNetworkMetadata(d, orgNetwork)
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error updating Isolated network metadata: %s", err)
	}

	return resourceVcdNetworkIsolatedV2Read(ctx, d, meta)
}

func resourceVcdNetworkIsolatedV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error retrieving VDC: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found - unset ID
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error getting Isolated network: %s", err)
	}

	err = setOpenApiOrgVdcIsolatedNetworkData(d, orgNetwork.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error setting Isolated network data: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	// Metadata is not supported when the network is in a VDC Group, although it is still present in the entity.
	// Hence, we skip the read to preserve its value in state.
	var diagErr diag.Diagnostics
	if !govcd.OwnerIsVdcGroup(orgNetwork.OpenApiOrgVdcNetwork.OwnerRef.ID) {
		diagErr = updateMetadataInState(d, vcdClient, "vcd_network_isolated_v2", orgNetwork)
	} else if _, ok := d.GetOk("metadata"); !ok {
		// If it's a VDC Group and metadata is not set, we explicitly compute it to empty. Otherwise, its value should
		// be preserved as it is still present in the entity.
		err = d.Set("metadata", StringMap{})
		if err != nil {
			diagErr = diag.FromErr(err)
		}
	}
	if diagErr != nil {
		log.Printf("[DEBUG] Unable to set isolated network v2 metadata: %s", err)
		return diagErr
	}

	return nil
}

func resourceVcdNetworkIsolatedV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Only when a network is in VDC Group - it must lock parent VDC Group. It doesn't cause lock
	// issues when created in VDC.
	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network create v2] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[isolated network v2 delete] error getting Isolated network: %s", err)
	}

	err = orgNetwork.Delete()
	if err != nil {
		return diag.Errorf("[isolated network v2 delete] error deleting Isolated network: %s", err)
	}

	return nil
}

func resourceVcdNetworkIsolatedV2Import(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[isolated network v2 import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcOrVdcGroupName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]
	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	orgNetwork, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Isolated network '%s': %s", networkName, err)
	}

	if !orgNetwork.IsIsolated() {
		return nil, fmt.Errorf("[isolated network v2 import] Org network with name '%s' found, but is not of type Isolated (type is '%s')",
			networkName, orgNetwork.GetType())
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcOrVdcGroupName)
	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func setOpenApiOrgVdcIsolatedNetworkData(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {
	dSet(d, "name", orgVdcNetwork.Name)
	dSet(d, "description", orgVdcNetwork.Description)

	dSet(d, "owner_id", orgVdcNetwork.OwnerRef.ID)
	dSet(d, "vdc", orgVdcNetwork.OwnerRef.Name)

	// Only one subnet can be defined although the structure accepts slice
	dSet(d, "gateway", orgVdcNetwork.Subnets.Values[0].Gateway)
	dSet(d, "prefix_length", orgVdcNetwork.Subnets.Values[0].PrefixLength)
	dSet(d, "dns1", orgVdcNetwork.Subnets.Values[0].DNSServer1)
	dSet(d, "dns2", orgVdcNetwork.Subnets.Values[0].DNSServer2)
	dSet(d, "dns_suffix", orgVdcNetwork.Subnets.Values[0].DNSSuffix)
	dSet(d, "is_shared", orgVdcNetwork.Shared)

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

func getOpenApiOrgVdcIsolatedNetworkType(d *schema.ResourceData, vcdClient *VCDClient) (*types.OpenApiOrgVdcNetwork, error) {
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)

	ownerId, err := getOwnerId(d, vcdClient, ownerIdField, vdcField, inheritedVdcField)
	if err != nil {
		return nil, fmt.Errorf("error finding owner reference: %s", err)
	}

	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: ownerId},

		NetworkType: types.OrgVdcNetworkTypeIsolated,
		Shared:      addrOf(d.Get("is_shared").(bool)),

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

func createOrUpdateOpenApiNetworkMetadata(d *schema.ResourceData, network *govcd.OpenApiOrgVdcNetwork) error {
	log.Printf("[TRACE] adding/updating metadata to Network V2")

	// Metadata is not supported when the network is in a VDC Group
	if govcd.OwnerIsVdcGroup(network.OpenApiOrgVdcNetwork.OwnerRef.ID) {
		return nil
	}

	return createOrUpdateMetadata(d, network, "metadata")
}
