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

var nsxtDhcpPoolSetSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Start address of DHCP pool IP range",
		},
		"end_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "End address of DHCP pool IP range",
		},
	},
}

func resourceVcdOpenApiDhcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOpenApiDhcpCreate,
		ReadContext:   resourceVcdOpenApiDhcpRead,
		UpdateContext: resourceVcdOpenApiDhcpUpdate,
		DeleteContext: resourceVcdOpenApiDhcpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOpenApiDhcpImport,
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
				Deprecated:  "Org network will be looked up based on 'org_network_id' field",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network ID",
			},
			"pool": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "IP ranges used for DHCP pool allocation in the network",
				Elem:        nsxtDhcpPoolSetSchema,
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The DNS server IPs to be assigned by this DHCP service. 2 values maximum.",
				MaxItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdOpenApiDhcpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool set] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool create] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	dhcpType := getOpenAPIOrgVdcNetworkDhcpType(d)

	// DnsServers is a feature added from API 36.1. If API is lower, this attribute is set to empty to avoid sending it
	_, ok := d.GetOk("dns_servers")
	if ok && vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		return diag.Errorf("`dns_servers` is supported from VCD 10.3.1+ version")
	}

	_, err = orgVdcNet.UpdateDhcp(dhcpType)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool set] error setting DHCP pool for Org VDC network ID '%s': %s",
			orgNetworkId, err)
	}
	// ID is in fact Org VDC network ID because DHCP pools do not have their own ID, only Org Network ID in API path
	d.SetId(orgNetworkId)

	return resourceVcdOpenApiDhcpRead(ctx, d, meta)
}

// resourceVcdOpenApiDhcpUpdate is exactly the same as resourceVcdOpenApiDhcpCreate because there is no "create"
// operation in this endpoint, only update.
func resourceVcdOpenApiDhcpUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdOpenApiDhcpCreate(ctx, d, meta)
}

func resourceVcdOpenApiDhcpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool read] error retrieving VDC: %s", err)
	}

	orgNetworkId := d.Id()
	// There may be cases when parent Org VDC network is no longer present. In that case we want to report that
	// DHCP pool no longer exists without breaking Terraform read.
	orgVdcNetwork, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}

		return diag.Errorf("[NSX-T DHCP pool read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	pool, err := orgVdcNetwork.GetOpenApiOrgVdcNetworkDhcp()
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool read] error retrieving DHCP pools for Org network ID '%s': %s",
			d.Id(), err)
	}

	err = setOpenAPIOrgVdcNetworkDhcpData(d.Id(), pool.OpenApiOrgVdcNetworkDhcp, d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool read] error setting DHCP pool data for Org network ID '%s': %s",
			orgNetworkId, err)
	}

	return nil
}

func resourceVcdOpenApiDhcpDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool delete] error retrieving Org: %s", err)
	}

	orgVdcNetwork, err := org.GetOpenApiOrgVdcNetworkById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool delete] error retrieving parent Org VDC Network: %s", err)
	}

	err = orgVdcNetwork.DeletNetworkDhcp()
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool delete] error removing DHCP pool for Org network ID '%s': %s", d.Id(), err)
	}

	return nil
}

func resourceVcdOpenApiDhcpImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-org-vdc-group-name.org_network_name")
	}
	orgName, vdcOrVdcGroupName, orgVdcNetworkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	// define an interface type to match VDC and VDC Groups
	var vdcOrVdcGroup vdcOrVdcGroupHandler
	_, vdcOrVdcGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcOrVdcGroupName)
	if govcd.ContainsNotFound(err) {
		adminOrg, err := vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrVdcGroup, err = adminOrg.GetVdcGroupByName(vdcOrVdcGroupName)
		if err != nil {
			return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s': %s", vdcOrVdcGroupName, err)
		}
	}

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(orgVdcNetworkName)
	if err != nil {
		return nil, fmt.Errorf("[NSX-T DHCP pool import] error retrieving Org VDC network with name '%s': %s", orgVdcNetworkName, err)
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("[NSX-T DHCP pool import] DHCP configuration is only supported for Routed NSX-T networks: %s", err)
	}

	dSet(d, "org", orgName)
	d.SetId(orgVdcNet.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func getOpenAPIOrgVdcNetworkDhcpType(d *schema.ResourceData) *types.OpenApiOrgVdcNetworkDhcp {
	orgVdcNetDhcp := &types.OpenApiOrgVdcNetworkDhcp{
		DhcpPools: nil,
	}

	dhcpPool := d.Get("pool")
	if dhcpPool == nil {
		return orgVdcNetDhcp
	}

	dhcpPoolSet := dhcpPool.(*schema.Set)
	dhcpPoolList := dhcpPoolSet.List()

	if len(dhcpPoolList) > 0 {
		dhcpPools := make([]types.OpenApiOrgVdcNetworkDhcpPools, len(dhcpPoolList))
		for index, pool := range dhcpPoolList {
			poolMap := pool.(map[string]interface{})
			onePool := types.OpenApiOrgVdcNetworkDhcpPools{
				IPRange: types.OpenApiOrgVdcNetworkDhcpIpRange{
					StartAddress: poolMap["start_address"].(string),
					EndAddress:   poolMap["end_address"].(string),
				},
			}
			dhcpPools[index] = onePool
		}

		// Inject data into main structure
		orgVdcNetDhcp.DhcpPools = dhcpPools
	}

	dnsServers, ok := d.GetOk("dns_servers")
	if ok {
		dnsServerSet := make([]string, len(dnsServers.([]interface{})))
		for i, v := range dnsServers.([]interface{}) {
			dnsServerSet[i] = v.(string)
		}
		orgVdcNetDhcp.DnsServers = dnsServerSet
	}

	return orgVdcNetDhcp
}

func setOpenAPIOrgVdcNetworkDhcpData(orgNetworkId string, orgVdc *types.OpenApiOrgVdcNetworkDhcp, d *schema.ResourceData) error {
	dSet(d, "org_network_id", orgNetworkId)
	if len(orgVdc.DhcpPools) > 0 {
		poolInterfaceSlice := make([]interface{}, len(orgVdc.DhcpPools))

		for index, pool := range orgVdc.DhcpPools {
			onePool := make(map[string]interface{})
			onePool["start_address"] = pool.IPRange.StartAddress
			onePool["end_address"] = pool.IPRange.EndAddress

			poolInterfaceSlice[index] = onePool
		}

		dhcpPoolSet := schema.NewSet(schema.HashResource(nsxtDhcpPoolSetSchema), poolInterfaceSlice)
		err := d.Set("pool", dhcpPoolSet)
		if err != nil {
			return err
		}
	}

	if len(orgVdc.DnsServers) > 0 {
		dSet(d, "dns_servers", orgVdc.DnsServers)
	}

	return nil
}
