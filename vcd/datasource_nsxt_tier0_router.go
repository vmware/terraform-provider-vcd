package vcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtTier0Router() *schema.Resource {
	return &schema.Resource{
		Read: datasourceNsxtTier0RouterRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Tier-0 router.",
			},
			"nsxt_manager_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of NSX-T manager.",
			},
			"is_assigned": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if Tier-0 router is already assigned to external network.",
			},
		},
	}
}

// datasourceNsxtTier0RouterRead has special behavior. By default `GetImportableNsxtTier0RouterByName` which uses API
// endpoint `1.0.0/nsxTResources/importableTier0Routers` does not return Tier-0 routers when they are used in external
// networks. This causes a problem in regular Terraform flow - when user uses this datasource to reference Tier-0 router
// for external network creation - next "apply" would fail with "Tier 0 router not found error". If original endpoint
// does not find Tier-0 router - then this datasource queries all defined external networks and looks for Tier-0 router
// backing by name.
func datasourceNsxtTier0RouterRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	nsxtManagerId := d.Get("nsxt_manager_id").(string)
	tier0RouterName := d.Get("name").(string)

	tier0Router, err := vcdClient.GetImportableNsxtTier0RouterByName(tier0RouterName, nsxtManagerId)
	if err != nil && !govcd.ContainsNotFound(err) {
		return fmt.Errorf("could not find NSX-T Tier-0 router by name '%s' in NSX-T manager %s: %s",
			tier0RouterName, nsxtManagerId, err)
	}

	// If unused Tier-0 router is found - set the ID and return
	if err == nil {
		d.Set("is_assigned", false)
		d.SetId(tier0Router.NsxtTier0Router.ID)
		return nil
	}

	// API endpoint for Tier-0 routers does not return Tier-0 routers which are already used in external networks
	// therefore we are searching for used Tier-0 router name in external networks. This should not cause any risks as
	// required permissions should be of the same level.
	if govcd.ContainsNotFound(err) {
		// Filtering by network backing is unsupported therefore queryParameters are nil
		extNets, err := govcd.GetAllExternalNetworksV2(vcdClient.VCDClient, nil)
		if err != nil {
			return fmt.Errorf("could not find external networks: %s", err)
		}

		for _, extNetwork := range extNets {
			for _, v := range extNetwork.ExternalNetwork.NetworkBackings.Values {
				// Very odd but when VRF Tier-0 router is used - BackingType can be UNKNOWN
				if v.Name == tier0RouterName &&
					(v.BackingType == types.ExternalNetworkBackingTypeNsxtTier0Router || v.BackingType == "UNKNOWN") {
					d.Set("is_assigned", true)
					d.SetId(v.BackingID)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("%s: could not find NSX-T Tier-0 router by name '%s' in NSX-T manager %s",
		govcd.ErrorEntityNotFound, tier0RouterName, nsxtManagerId)
}
