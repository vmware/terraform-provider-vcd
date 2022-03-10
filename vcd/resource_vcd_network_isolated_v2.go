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
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network description",
			},
			"is_shared": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "NSX-V only - share this network with other VDCs in this organization. Default - false",
			},
			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Gateway IP address",
			},
			"prefix_length": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Network prefix",
			},
			"dns1": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS server 1",
			},
			"dns2": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS server 1",
			},
			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS suffix",
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRange,
			},
		},
	}
}

// resourceVcdNetworkIsolatedV2Create
func resourceVcdNetworkIsolatedV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network v2 create] error retrieving VDC: %s", err)
	}

	networkType, err := getOpenApiOrgVdcIsolatedNetworkType(d, vdc)
	if err != nil {
		return diag.FromErr(err)
	}

	orgNetwork, err := vdc.CreateOpenApiOrgVdcNetwork(networkType)
	if err != nil {
		return diag.Errorf("[isolated network v2 create] error creating Org VDC isolated network: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return resourceVcdNetworkIsolatedV2Read(ctx, d, meta)
}

// resourceVcdNetworkIsolatedV2Update
func resourceVcdNetworkIsolatedV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error retrieving VDC: %s", err)
	}

	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found -
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error getting Org VDC network: %s", err)
	}

	networkType, err := getOpenApiOrgVdcIsolatedNetworkType(d, vdc)
	if err != nil {
		return diag.FromErr(err)
	}

	// Explicitly add ID to the new type because function `getOpenApiOrgVdcIsolatedNetworkType` only sets other fields
	networkType.ID = d.Id()

	_, err = orgNetwork.Update(networkType)
	if err != nil {
		return diag.Errorf("[isolated network v2 update] error updating Org VDC network: %s", err)
	}

	return resourceVcdNetworkIsolatedV2Read(ctx, d, meta)
}

// resourceVcdNetworkIsolatedV2Read
func resourceVcdNetworkIsolatedV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error retrieving VDC: %s", err)
	}

	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkById(d.Id())
	// If object is not found - unset ID
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error getting Org VDC network: %s", err)
	}

	err = setOpenApiOrgVdcIsolatedNetworkData(d, orgNetwork.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[isolated network v2 read] error setting Org VDC network data: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return nil
}

// resourceVcdNetworkIsolatedV2Delete
func resourceVcdNetworkIsolatedV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[isolated network v2 delete] error retrieving VDC: %s", err)
	}

	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[isolated network v2 delete] error getting Org VDC network: %s", err)
	}

	err = orgNetwork.Delete()
	if err != nil {
		return diag.Errorf("[isolated network v2 delete] error deleting Org VDC network: %s", err)
	}

	return nil
}

// resourceVcdNetworkIsolatedV2Import
func resourceVcdNetworkIsolatedV2Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[isolated network v2 import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[isolated network v2 import] unable to find VDC %s: %s ", vdcName, err)
	}

	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("[isolated network v2 import] error reading network with name '%s': %s", networkName, err)
	}

	if !orgNetwork.IsIsolated() {
		return nil, fmt.Errorf("[isolated network v2 import] Org network with name '%s' found, but is not of type Isolated (type is '%s')",
			networkName, orgNetwork.GetType())
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func setOpenApiOrgVdcIsolatedNetworkData(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {

	dSet(d, "name", orgVdcNetwork.Name)
	dSet(d, "description", orgVdcNetwork.Description)

	// Only one subnet can be defined although the structure accepts slice
	dSet(d, "gateway", orgVdcNetwork.Subnets.Values[0].Gateway)
	dSet(d, "prefix_length", orgVdcNetwork.Subnets.Values[0].PrefixLength)
	dSet(d, "dns1", orgVdcNetwork.Subnets.Values[0].DNSServer1)
	dSet(d, "dns2", orgVdcNetwork.Subnets.Values[0].DNSServer2)
	dSet(d, "dns_suffix", orgVdcNetwork.Subnets.Values[0].DNSSuffix)
	dSet(d, "is_shared", orgVdcNetwork.Shared)

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

func getOpenApiOrgVdcIsolatedNetworkType(d *schema.ResourceData, vdc *govcd.Vdc) (*types.OpenApiOrgVdcNetwork, error) {
	orgVdcNetworkConfig := &types.OpenApiOrgVdcNetwork{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		// On v35.0 onwards OrgVdc is not supported anymore. Using OwnerRef instead.
		OwnerRef: &types.OpenApiReference{ID: vdc.Vdc.ID},

		NetworkType: types.OrgVdcNetworkTypeIsolated,
		Shared:      takeBoolPointer(d.Get("is_shared").(bool)),

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
