package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func readVApp(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.GetVAppByHREF(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vApp before running customization: %#v", err)
	}

	// Reading networks defined on the vApp

	organizationNetworksFromState := interfaceListToStringList(
		d.Get("organization_network").([]interface{}))
	vAppNetworksFromState := interfaceListToMapStringInterface(
		d.Get("vapp_network").([]interface{}))

	// Read collection
	readOrgNetworks := make([]string, 0)
	readVAppNetworks := make([]map[string]interface{}, 0)

	// // Order is not guarenteed, so we will have to check what we have first
	// // First look for networks that we already have
	for _, network := range organizationNetworksFromState {
		vAppNetwork, _ := vapp.GetNetworkByName(network)
		if err != nil {
			return err
		}

		if vAppNetwork != nil {
			readOrgNetworks = append(readOrgNetworks, network)
		}
	}

	for index := range vAppNetworksFromState {
		vAppNetworkResource := NewVAppNetworkSubresource(vAppNetworksFromState[index], nil)
		vAppNetwork, _ := vapp.GetNetworkByName(vAppNetworkResource.Get("name").(string))
		if err != nil {
			return err
		}

		if vAppNetwork != nil {
			vAppNetworkResource.Set("name",
				vAppNetwork.NetworkName)
			// vAppNetworkResource.Set("description", vAppnetwork.Configuration)
			vAppNetworkResource.Set("gateway",
				vAppNetwork.Configuration.IPScopes.IPScope.Gateway)
			vAppNetworkResource.Set("netmask",
				vAppNetwork.Configuration.IPScopes.IPScope.Netmask)
			vAppNetworkResource.Set("dns1",
				vAppNetwork.Configuration.IPScopes.IPScope.DNS1)
			vAppNetworkResource.Set("dns2",
				vAppNetwork.Configuration.IPScopes.IPScope.DNS2)
			vAppNetworkResource.Set("start",
				vAppNetwork.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].StartAddress)
			vAppNetworkResource.Set("end",
				vAppNetwork.Configuration.IPScopes.IPScope.IPRanges.IPRange[0].EndAddress)

			if vAppNetwork.Configuration.ParentNetwork != nil {
				vAppNetworkResource.Set("parent",
					vAppNetwork.Configuration.ParentNetwork.Name)
			}

			if vAppNetwork.Configuration.Features != nil {
				if vAppNetwork.Configuration.Features.NatService != nil {
					vAppNetworkResource.Set("nat",
						vAppNetwork.Configuration.Features.NatService.IsEnabled)
				}

				if vAppNetwork.Configuration.Features.DhcpService != nil {
					vAppNetworkResource.Set("dhcp",
						vAppNetwork.Configuration.Features.DhcpService.IsEnabled)
				}
			}

			readVAppNetworks = append(readVAppNetworks, vAppNetworkResource.Data())
		}
	}

	log.Printf("[TRACE] Org Networks defined for vApp (%s) is: %#v", vapp.VApp.Name, readOrgNetworks)
	log.Printf("[TRACE] vApp Networks defined for vApp (%s) is: %#v", vapp.VApp.Name, readOrgNetworks)

	d.Set("organization_network", readOrgNetworks)
	d.Set("vapp_network", readVAppNetworks)

	return nil

}

func createNetworkConfiguration(d *schema.ResourceData, meta interface{}) ([]*types.VAppNetworkConfiguration, error) {
	vcdClient := meta.(*VCDClient)

	// Organization Network
	organizationNetworks := d.Get("organization_network").([]interface{})
	log.Printf("[TRACE] Networks from state: %#v", organizationNetworks)

	orgnetworks := make([]*types.VAppNetworkConfiguration, len(organizationNetworks))
	for index, network := range organizationNetworks {
		orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(network.(string))
		if err != nil {
			return nil, fmt.Errorf("Error finding vdc org network: %s, %#v", network, err)
		}
		orgnetworks[index] = orgVDCNetworkToNetworkConfiguration(orgnetwork.OrgVDCNetwork)
	}

	// vApp Network
	vAppNetworksInterfaceList := d.Get("vapp_network").([]interface{})
	vAppNetworks := interfaceListToMapStringInterface(vAppNetworksInterfaceList)

	vAppNetworkConfigurations := make([]*types.VAppNetworkConfiguration, len(vAppNetworks))
	for index := range vAppNetworks {
		vAppNetwork := NewVAppNetworkSubresource(vAppNetworks[index], nil)

		configuration := &types.NetworkConfiguration{
			FenceMode: types.FenceModeIsolated,
			Features:  &types.NetworkFeatures{},
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
					IsInherited: false,
					Gateway:     vAppNetwork.Get("gateway").(string),
					Netmask:     vAppNetwork.Get("netmask").(string),
					DNS1:        vAppNetwork.Get("dns1").(string),
					DNS2:        vAppNetwork.Get("dns2").(string),
					IsEnabled:   true,
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{&types.IPRange{
							StartAddress: vAppNetwork.Get("start").(string),
							EndAddress:   vAppNetwork.Get("end").(string),
						}},
					},
				},
			},
		}

		if vAppNetwork.Get("dhcp").(bool) {
			configuration.Features.DhcpService = &types.DhcpService{
				IsEnabled: true,
				IPRange: &types.IPRange{
					StartAddress: vAppNetwork.Get("dhcp_start").(string),
					EndAddress:   vAppNetwork.Get("dhcp_end").(string),
				},
				PrimaryNameServer:   configuration.IPScopes.IPScope.DNS1,
				SecondaryNameServer: configuration.IPScopes.IPScope.DNS1,
				SubMask:             configuration.IPScopes.IPScope.Netmask,
				RouterIP:            configuration.IPScopes.IPScope.Gateway,
			}
		}

		if vAppNetwork.Get("nat").(bool) {
			configuration.Features.NatService = &types.NatService{
				IsEnabled: true,
				NatType:   "ipTranslation",
				Policy:    "allowTrafficIn",
				// We need to set parent
			}

			orgnetwork, err := vcdClient.OrgVdc.FindVDCNetwork(vAppNetwork.Get("parent").(string))

			if err != nil {
				return nil, fmt.Errorf("Error finding vdc org network: %s, %#v", vAppNetwork.Get("parent").(string), err)
			}
			configuration.ParentNetwork = &types.Reference{
				HREF: orgnetwork.OrgVDCNetwork.HREF,
				ID:   orgnetwork.OrgVDCNetwork.ID,
				Name: orgnetwork.OrgVDCNetwork.Name,
			}

			configuration.FenceMode = types.FenceModeNAT
		}

		vAppNetworkConfigurations[index] = &types.VAppNetworkConfiguration{
			Configuration: configuration,
			NetworkName:   vAppNetwork.Get("name").(string),
		}
	}

	networks := append(orgnetworks, vAppNetworkConfigurations...)
	return networks, nil
}

func orgVDCNetworkToNetworkConfiguration(orgnetwork *types.OrgVDCNetwork) *types.VAppNetworkConfiguration {
	return &types.VAppNetworkConfiguration{
		NetworkName: orgnetwork.Name,
		Configuration: &types.NetworkConfiguration{
			FenceMode: types.FenceModeBridged,
			ParentNetwork: &types.Reference{
				HREF: orgnetwork.HREF,
				Name: orgnetwork.Name,
				Type: orgnetwork.Type,
			},
		},
	}
}

func networkInList(networks []string, vAppNetwork *types.VAppNetworkConfiguration) bool {
	for _, network := range networks {
		if network == vAppNetwork.NetworkName {
			return true
		}
	}
	return false
}
