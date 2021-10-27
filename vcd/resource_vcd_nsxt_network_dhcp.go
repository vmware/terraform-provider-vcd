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
		"start_address": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "Start address of DHCP pool IP range",
		},
		"end_address": &schema.Schema{
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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},

			"org_network_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network ID",
			},
			"pool": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "IP ranges used for DHCP pool allocation in the network",
				Elem:        nsxtDhcpPoolSetSchema,
			},
		},
	}
}

func resourceVcdOpenApiDhcpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool set] error retrieving VDC: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := vdc.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool create] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	if !orgVdcNet.IsRouted() || !vdc.IsNsxt() {
		return diag.Errorf("[NSX-T DHCP pool set] DHCP configuration is only supported for Routed NSX-T networks: %s", err)
	}

	dhcpType := getOpenAPIOrgVdcNetworkDhcpType(d)
	_, err = vdc.UpdateOpenApiOrgVdcNetworkDhcp(orgNetworkId, dhcpType)
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

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool read] error retrieving VDC: %s", err)
	}

	orgNetworkId := d.Id()
	// There may be cases when parent Org VDC network is no longer present. In that case we want to report that
	// DHCP pool no longer exists without breaking Terraform read.
	_, err = vdc.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}

		return diag.Errorf("[NSX-T DHCP pool read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	pool, err := vdc.GetOpenApiOrgVdcNetworkDhcp(d.Id())
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

	// VCD versions < 10.2 do not allow to execute "DELETE" therefore we emit warning and "return success" to prevent
	// destroy errors breaking Terraform flow.
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		dumpFprint(getTerraformStdout(), "vcd_nsxt_network_dhcp WARNING: for VCD versions < 10.2 DHCP pool "+
			"removal is not supported. Destroy is a NO-OP for VCD versions < 10.2. "+
			"Please recreate parent network to remove DHCP pools.\n")
		return nil
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool delete] error retrieving VDC: %s", err)
	}

	err = vdc.DeleteOpenApiOrgVdcNetworkDhcp(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool delete] error removing DHCP pool for Org network ID '%s': %s", d.Id(), err)
	}

	return nil
}

func resourceVcdOpenApiDhcpImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.org_network_name")
	}
	orgName, vdcName, orgVdcNetworkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", orgName, err)
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
	}

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := vdc.GetOpenApiOrgVdcNetworkByName(orgVdcNetworkName)
	if err != nil {
		return nil, fmt.Errorf("[NSX-T DHCP pool import] error retrieving Org VDC network with name '%s': %s", orgVdcNetworkName, err)
	}

	if !orgVdcNet.IsRouted() || !vdc.IsNsxt() {
		return nil, fmt.Errorf("[NSX-T DHCP pool import] DHCP configuration is only supported for Routed NSX-T networks: %s", err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
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

	return nil
}
