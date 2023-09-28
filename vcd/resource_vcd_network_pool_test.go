//go:build ALL || functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"testing"
)

type networkPoolData struct {
	name              string
	description       string
	poolType          string
	networkProviderId string
	backingType       string
	backingName       string
}

func TestAccVcdResourceNetworkPool(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient == nil {
		t.Skipf("error getting client configuration")
	}
	vCenter, err := vcdClient.GetVCenterByName(testConfig.Networking.Vcenter)
	if err != nil {
		t.Skipf("error getting vCenter '%s': %s", testConfig.Networking.Vcenter, err)
	}

	nsxtManagers, err := vcdClient.QueryNsxtManagerByName(testConfig.Nsxt.Manager)
	if err != nil || len(nsxtManagers) == 0 {
		t.Skipf("error getting NSX-T manager '%s': %s", testConfig.Nsxt.Manager, err)
	}
	nsxtManager := nsxtManagers[0]
	nsxtManagerId := "urn:vcloud:nsxtmanager:" + extractUuid(nsxtManager.HREF)

	// Collecting transport zones, distributed switches, and importable port groups.
	// Note that while we could do the same using vcd_resource_list, the plan would fail
	// after that, as the retrieved names would not be available. Thus, we need to cache
	// the retrieved names outside of Terraform
	transportZones, err := vcdClient.GetAllNsxtTransportZones(nsxtManagerId, nil)
	if err != nil {
		t.Skipf("error getting NSX-T transport zones: %s", err)
	}
	distributedSwitches, err := vcdClient.GetAllVcenterDistributedSwitches(vCenter.VSphereVCenter.VcId, nil)
	if err != nil {
		t.Skipf("error getting distributed switches: %s", err)
	}
	var params = make(url.Values)
	params.Set("virtualCenter.id", vCenter.VSphereVCenter.VcId)
	portGroups, err := vcdClient.GetAllVcenterImportableDvpgs(params)
	if err != nil {
		t.Skipf("error getting port groups: %s", err)
	}

	for _, tz := range transportZones {
		if tz.AlreadyImported {
			continue
		}
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-geneve-name",
			description:       t.Name() + "-geneve-name description",
			poolType:          types.NetworkPoolGeneveType,
			networkProviderId: nsxtManagerId,
			backingType:       "transport_zone",
			backingName:       tz.Name,
		})
	}
	if len(transportZones) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-geneve-none",
			description:       t.Name() + "-geneve-none description",
			poolType:          types.NetworkPoolGeneveType,
			networkProviderId: nsxtManagerId,
			backingType:       "none",
			backingName:       "none",
		})
	}
	for _, ds := range distributedSwitches {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-vlan-name",
			description:       t.Name() + "-vlan-name description",
			poolType:          types.NetworkPoolVlanType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "distributed_switches",
			backingName:       ds.BackingRef.Name,
		})
	}
	if len(distributedSwitches) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-vlan-none",
			description:       t.Name() + "-vlan-none description",
			poolType:          types.NetworkPoolVlanType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "none",
			backingName:       "none",
		})
	}
	for _, pg := range portGroups {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-pg-name",
			description:       t.Name() + "-pg-name description",
			poolType:          types.NetworkPoolPortGroupType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "port_groups",
			backingName:       pg.VcenterImportableDvpg.BackingRef.Name,
		})
	}
	if len(portGroups) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-pg-none",
			description:       t.Name() + "-pg-none description",
			poolType:          types.NetworkPoolPortGroupType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "none",
			backingName:       "none",
		})
	}

	postTestChecks(t)
}

func runNetworkPoolTest(t *testing.T, npData networkPoolData) {

	tmpl := testAccNetworkPoolWithBacking
	if npData.backingType == "none" {
		tmpl = testAccNetworkPoolNoBacking
	}

	var data = StringMap{
		"SkipMessage":            " ",
		"NetworkPoolName":        npData.name,
		"NetworkPoolDescription": npData.description,
		"NetworkProviderId":      npData.networkProviderId,
		"PoolType":               npData.poolType,
		"BackingType":            npData.backingType,
		"BackingName":            npData.backingName,
		"RangeIds":               " ",
		"FuncName":               t.Name() + "-" + npData.poolType + "-" + npData.backingName,
	}

	if npData.poolType == types.NetworkPoolVlanType {
		data["RangeIds"] = testAccRangeIds
	}
	testParamsNotEmpty(t, data)

	configText := templateFill(tmpl, data)
	debugPrintf("#[DEBUG] Configuration: %s", configText)

	updatedName := npData.name + "-updated"
	updatedDescription := npData.description + " updated"
	data["FuncName"] = data["FuncName"].(string) + "-update"
	data["NetworkPoolName"] = updatedName
	data["NetworkPoolDescription"] = updatedDescription

	updatedText := templateFill(tmpl, data)
	debugPrintf("#[DEBUG] Update: %s", updatedText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	t.Run(npData.poolType+"-"+npData.backingName, func(t *testing.T) {

		resourceDef := "vcd_network_pool.npool"
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				checkNetworkPoolExists(npData.name, false),
			),
			Steps: []resource.TestStep{
				// step 0 - Create network pool
				{
					Config: configText,
					Check: resource.ComposeTestCheckFunc(
						checkNetworkPoolExists(npData.name, true),
						resource.TestCheckResourceAttr(resourceDef, "name", npData.name),
						resource.TestCheckResourceAttr(resourceDef, "description", npData.description),
						resource.TestCheckResourceAttr(resourceDef, "status", "REALIZED"),
					),
				},
				// step 1 - update network pool
				{
					Config: updatedText,
					Check: resource.ComposeTestCheckFunc(
						checkNetworkPoolExists(updatedName, true),
						resource.TestCheckResourceAttr(resourceDef, "name", updatedName),
						resource.TestCheckResourceAttr(resourceDef, "description", updatedDescription),
						resource.TestCheckResourceAttr(resourceDef, "status", "REALIZED"),
					),
				},
				// step 2 - import
				{
					ResourceName:      resourceDef,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importStateIdTopHierarchy(updatedName),
				},
			},
		})
	})
}

const testAccNetworkPoolWithBacking = `
{{.SkipMessage}}
resource "vcd_network_pool" "npool" {
  name                = "{{.NetworkPoolName}}"
  description         = "{{.NetworkPoolDescription}}"
  network_provider_id = "{{.NetworkProviderId}}"
  type                = "{{.PoolType}}"

  backing {
    {{.BackingType}} {
      name = "{{.BackingName}}"
    }
    {{.RangeIds}}
  }
}
`

const testAccNetworkPoolNoBacking = `
{{.SkipMessage}}
resource "vcd_network_pool" "npool" {
  name                = "{{.NetworkPoolName}}"
  description         = "{{.NetworkPoolDescription}}"
  network_provider_id = "{{.NetworkProviderId}}"
  type                = "{{.PoolType}}"
  backing {
    {{.RangeIds}}
  }
}
`
const testAccRangeIds = `
    range_ids {
      start_id = 101
      end_id   = 200
    }
`
