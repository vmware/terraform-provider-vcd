package vcd

import (
	"fmt"
	"net/netip"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func getOpenApiOrgVdcSecondaryNetworkType(d *schema.ResourceData, orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) error {
	isDualStackEnabled := d.Get("dual_stack_enabled").(bool)
	orgVdcNetworkConfig.EnableDualSubnetNetwork = &isDualStackEnabled

	secondaryGateway := d.Get("secondary_gateway").(string)
	secondaryPrefixLength := d.Get("secondary_prefix_length").(string)

	// `secondary_gateway` can only be IPv6 that works in Dual Stack mode only
	if isDualStackEnabled && (secondaryGateway == "" || secondaryPrefixLength == "") {
		return fmt.Errorf("'secondary_gateway' and 'secondary_prefix_length' must be set when 'dual_stack_enabled' is enabled")
	}

	if secondaryGateway != "" && secondaryPrefixLength != "" {
		parsedSecondaryGateway, err := netip.ParseAddr(secondaryGateway)
		if err != nil {
			return fmt.Errorf("error parsing 'secondary_gateway' %s: %s", secondaryGateway, err)
		}
		if !parsedSecondaryGateway.Is6() {
			return fmt.Errorf("'secondary_gateway' can only be IPv6 address")
		}

		// Ignoring conversion error evaluation because it is done in schema
		secondaryPrefixLengthInt, _ := strconv.Atoi(secondaryPrefixLength)

		orgVdcNetworkConfig.Subnets.Values = append(orgVdcNetworkConfig.Subnets.Values,
			types.OrgVdcNetworkSubnetValues{
				Gateway:      secondaryGateway,
				PrefixLength: secondaryPrefixLengthInt,
				IPRanges: types.OrgVdcNetworkSubnetIPRanges{
					Values: processIpRanges(d.Get("secondary_static_ip_pool").(*schema.Set)),
				},
				// Not setting DNS fields here as they are being set only for the first entry of
				// 'types.OrgVdcNetworkSubnetValues'
				// DNSServer1:   d.Get("dns1").(string),
				// DNSServer2:   d.Get("dns2").(string),
				// DNSSuffix:    d.Get("dns_suffix").(string),
			},
		)
	}

	return nil
}

func setSecondarySubnet(d *schema.ResourceData, orgVdcNetwork *types.OpenApiOrgVdcNetwork) error {
	dSet(d, "dual_stack_enabled", *orgVdcNetwork.EnableDualSubnetNetwork)

	if len(orgVdcNetwork.Subnets.Values) > 1 {
		dSet(d, "secondary_gateway", orgVdcNetwork.Subnets.Values[1].Gateway)
		dSet(d, "secondary_prefix_length", strconv.Itoa(orgVdcNetwork.Subnets.Values[1].PrefixLength))

		if len(orgVdcNetwork.Subnets.Values[1].IPRanges.Values) > 0 {
			err := setOpenApiOrgVdcNetworkStaticPoolData(d, orgVdcNetwork.Subnets.Values[1].IPRanges.Values, "secondary_static_ip_pool")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func setOpenApiOrgVdcNetworkStaticPoolData(d *schema.ResourceData, ipRanges []types.OpenApiIPRangeValues, schemaFieldName string) error {
	ipRangeSlice := make([]interface{}, len(ipRanges))
	for index, ipRange := range ipRanges {
		ipRangeMap := make(map[string]interface{})
		ipRangeMap["start_address"] = ipRange.StartAddress
		ipRangeMap["end_address"] = ipRange.EndAddress

		ipRangeSlice[index] = ipRangeMap
	}
	ipRangeSet := schema.NewSet(schema.HashResource(networkV2IpRange), ipRangeSlice)

	err := d.Set(schemaFieldName, ipRangeSet)
	if err != nil {
		return fmt.Errorf("error setting '%s': %s", schemaFieldName, err)
	}

	return nil
}
