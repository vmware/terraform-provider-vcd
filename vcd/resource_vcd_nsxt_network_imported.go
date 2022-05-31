package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtNetworkImported() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtNetworkImportedCreate,
		ReadContext:   resourceVcdNsxtNetworkImportedRead,
		UpdateContext: resourceVcdNsxtNetworkImportedUpdate,
		DeleteContext: resourceVcdNsxtNetworkImportedDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtNetworkImportedImport,
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
			"nsxt_logical_switch_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of existing NSX-T Logical Switch",
			},
			"nsxt_logical_switch_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of existing NSX-T Logical Switch",
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

func resourceVcdNsxtNetworkImportedCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("[nsxt imported network create] only System Administrator can operate NSX-T Imported networks")
	}

	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt imported network create] error retrieving Org: %s", err)
	}

	// Validate if VDC or VDC Group is NSX-T backed
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)

	err = validateIfVdcOrVdcGroupIsNsxt(org, inheritedVdcField, vdcField, ownerIdField)
	if err != nil {
		return diag.Errorf("[nsxt imported network create] this resource supports only NSX-T: %s", err)
	}

	networkType, err := getOpenApiOrgVdcImportedNetworkType(d, vcdClient, true)
	if err != nil {
		return diag.FromErr(err)
	}

	orgNetwork, err := org.CreateOpenApiOrgVdcNetwork(networkType)
	if err != nil {
		return diag.Errorf("[nsxt imported network create] error creating Org VDC imported network: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return resourceVcdNsxtNetworkImportedRead(ctx, d, meta)
}

func resourceVcdNsxtNetworkImportedUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("[nsxt imported network update] only System Administrator can operate NSX-T Imported networks")
	}

	// `vdc` field is deprecated. `vdc` value should not be changed unless it is removal of the
	// field at all to allow easy migration to `owner_id` path
	if _, new := d.GetChange("vdc"); d.HasChange("vdc") && new.(string) != "" {
		return diag.Errorf("changing 'vdc' field value is not supported. It can only be removed. " +
			"Please use `owner_id` field for moving network to/from VDC Group")
	}

	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt imported network update] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found -
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[nsxt imported network update] error getting Org VDC network: %s", err)
	}

	networkType, err := getOpenApiOrgVdcImportedNetworkType(d, vcdClient, false)
	if err != nil {
		return diag.FromErr(err)
	}

	// Feed in backing network ID, because it cannot be looked up after assignment to importable network
	networkType.BackingNetworkId = orgNetwork.OpenApiOrgVdcNetwork.BackingNetworkId

	// Explicitly add ID to the new type because function `getOpenApiOrgVdcNetworkType` only sets other fields
	networkType.ID = d.Id()

	_, err = orgNetwork.Update(networkType)
	if err != nil {
		return diag.Errorf("[nsxt imported network update] error updating Org VDC network: %s", err)
	}

	return resourceVcdNsxtNetworkImportedRead(ctx, d, meta)
}
func resourceVcdNsxtNetworkImportedRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("[nsxt imported network read] only System Administrator can operate NSX-T Imported networks")
	}

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt imported network create] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found - unset ID
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[nsxt imported network read] error getting Org VDC network: %s", err)
	}

	err = setOpenApiOrgVdcImportedNetworkData(d, orgNetwork.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[nsxt imported network read] error setting Org VDC network data: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return nil
}

func resourceVcdNsxtNetworkImportedDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("[nsxt imported network delete] only System Administrator can operate NSX-T Imported networks")
	}

	vcdClient.lockIfOwnerIsVdcGroup(d)
	defer vcdClient.unLockIfOwnerIsVdcGroup(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt imported network create] error retrieving Org: %s", err)
	}

	orgNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt imported network delete] error getting Org VDC network: %s", err)
	}

	err = orgNetwork.Delete()
	if err != nil {
		return diag.Errorf("[nsxt imported network delete] error deleting Org VDC network: %s", err)
	}

	return nil
}

func resourceVcdNsxtNetworkImportedImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[nsxt imported network import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcOrVdcGroupName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("imported networks are only supported in NSX-T backed environments")
	}

	orgNetwork, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("[nsxt imported network import] error reading network with name '%s': %s", networkName, err)
	}

	if !orgNetwork.IsImported() {
		return nil, fmt.Errorf("[nsxt imported network import] Org network with name '%s' found, but is not of type Imported (OPAQUE) (type is '%s')",
			networkName, orgNetwork.GetType())
	}

	dSet(d, "org", orgName)
	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func setOpenApiOrgVdcImportedNetworkData(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {
	// Note. VCD does not export `nsxt_logical_switch_name` and there is no API to retrieve it once
	// consumed (assigned to imported network) therefore there is no way to read name once it is set

	dSet(d, "name", orgVdcNetwork.Name)
	dSet(d, "description", orgVdcNetwork.Description)
	dSet(d, "nsxt_logical_switch_id", orgVdcNetwork.BackingNetworkId)
	dSet(d, "owner_id", orgVdcNetwork.OwnerRef.ID)
	dSet(d, "vdc", orgVdcNetwork.OwnerRef.Name)

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

func getOpenApiOrgVdcImportedNetworkType(d *schema.ResourceData, vcdClient *VCDClient, isCreate bool) (*types.OpenApiOrgVdcNetwork, error) {
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

		// 'OPAQUE' type is used for imported network
		NetworkType: types.OrgVdcNetworkTypeOpaque,

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

	// Lookup NSX-T logical switch in Create phase only, because there is no API to return the
	// network after it is consumed
	if isCreate {
		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Org: %s", err)
		}

		vdcOrVdcGroup, err := getVdcOrVdcGroupVerifierByOwnerId(org, ownerId)
		if err != nil {
			return nil, fmt.Errorf("error identifying VDC or VDC Group by Owner ID '%s' :%s", ownerId, err)
		}

		nsxtImportableSwitch, err := vdcOrVdcGroup.GetNsxtImportableSwitchByName(d.Get("nsxt_logical_switch_name").(string))
		if err != nil {
			return nil, fmt.Errorf("unable to find NSX-T logical switch: %s", err)
		}

		orgVdcNetworkConfig.BackingNetworkId = nsxtImportableSwitch.NsxtImportableSwitch.ID
	}

	return orgVdcNetworkConfig, nil
}
