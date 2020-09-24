package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdPortgroup() *schema.Resource {
	return &schema.Resource{
		Read: datasourcePortgroupRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Tier-0 router.",
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Portgroup type. One of 'NETWORK', 'DV_PORTGROUP'",
				ValidateFunc: validation.StringInSlice([]string{types.ExternalNetworkBackingTypeNetwork, types.ExternalNetworkBackingDvPortgroup}, false),
			},
			"vcenter_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter ID.",
			},
		},
	}
}

func datasourcePortgroupRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	portGroupType := d.Get("type").(string)
	portGroupName := d.Get("name").(string)

	var err error
	var pgs []*types.PortGroupRecordType

	switch portGroupType {
	case types.ExternalNetworkBackingTypeNetwork:
		pgs, err = govcd.QueryNetworkPortGroup(vcdClient.VCDClient, portGroupName)
	case types.ExternalNetworkBackingDvPortgroup:
		pgs, err = govcd.QueryDistributedPortGroup(vcdClient.VCDClient, portGroupName)
	default:
		return fmt.Errorf("unrecognized portgroup_type: %s", portGroupType)
	}

	if err != nil {
		return fmt.Errorf("error querying for portgroups '%s' of type '%s': %s", portGroupName, portGroupType, err)
	}

	if len(pgs) == 0 {
		return fmt.Errorf("%s: expected to get exactly one portgroup with name '%s' of type '%s', got %d",
			govcd.ErrorEntityNotFound, portGroupName, portGroupType, len(pgs))
	}

	if len(pgs) > 1 {
		return fmt.Errorf("expected to get exactly one portgroup with name '%s' of type '%s', got %d",
			portGroupName, portGroupType, len(pgs))
	}

	d.SetId(pgs[0].MoRef)
	return nil
}
