package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// natRuleImporter works as a shared structure for both dnat and snat rule resource
func natRuleImporter(natType string) func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		resourceURI := strings.Split(d.Id(), ".")
		if len(resourceURI) != 4 {
			return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw.rule-id")
		}
		orgName, vdcName, edgeName, natRuleId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

		vcdClient := meta.(*VCDClient)
		edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
		if err != nil {
			return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		readNatRule, err := edgeGateway.GetNsxvNatRuleById(natRuleId)
		if err != nil {
			return []*schema.ResourceData{}, fmt.Errorf("unable to find NAT rule with id %s: %s",
				d.Id(), err)
		}

		if readNatRule.Action != natType {
			return []*schema.ResourceData{}, fmt.Errorf("NAT rule with id %s is of type %s. Expected type %s. Please use correct resource",
				readNatRule.ID, readNatRule.Action, natType)
		}

		_ = d.Set("org", orgName)
		_ = d.Set("vdc", vdcName)
		_ = d.Set("edge_gateway", edgeName)

		d.SetId(readNatRule.ID)
		return []*schema.ResourceData{d}, nil
	}
}

// natRuleDeleter
func natRuleDeleter(natType string) func(d *schema.ResourceData, meta interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		vcdClient := meta.(*VCDClient)
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)

		edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		err = edgeGateway.DeleteNsxvNatRuleById(d.Id())
		if err != nil {
			return fmt.Errorf("error deleting NAT rule of type %s: %s", natType, err)
		}

		d.SetId("")
		return nil
	}
}
