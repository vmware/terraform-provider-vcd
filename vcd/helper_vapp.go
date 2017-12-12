package vcd

import (
	"fmt"
	"log"
	"math"

	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func readVApp(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Should be fetched by ID/HREF
	vapp, err := vcdClient.OrgVdc.FindVAppByName(d.Id())

	if err != nil {
		return fmt.Errorf("Error finding VApp: %#v", err)
	}

	err = vapp.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing vApp before running customization: %#v", err)
	}

	// Reading networks defined on the vApp

	networksFromState := interfaceListToStringList(
		d.Get("network").([]interface{}))

	// Read collection
	newNetworks := make([]string,
		int(math.Max(
			float64(len(vapp.VApp.NetworkConfigSection.NetworkConfig)),
			float64(len(networksFromState)))))

	// Order is not guarenteed, so we will have to check what we have first
	// First look for networks that we already have
	for index, network := range networksFromState {
		vAppNetwork, err := vapp.GetNetworkByName(network)

		if err != nil {
			return err
		}

		if vAppNetwork != nil {
			newNetworks[index] = network
		}
	}

	log.Printf("[TRACE] (%s) Networks read after step 1: %#v", vapp.VApp.Name, newNetworks)

	// Second, look for new networks added on remote site.
	for _, network := range vapp.VApp.NetworkConfigSection.NetworkConfig {
		if !networkInList(newNetworks, network) {
			newNetworks = append(newNetworks, network.NetworkName)
		}
	}

	log.Printf("[TRACE] Networks defined for vApp (%s) is: %#v", vapp.VApp.Name, newNetworks)

	d.Set("network", newNetworks)

	return nil

}

func networkInList(networks []string, vAppNetwork *types.VAppNetworkConfiguration) bool {
	for _, network := range networks {
		if network == vAppNetwork.NetworkName {
			return true
		}
	}
	return false
}
