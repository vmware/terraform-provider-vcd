//go:build ALL || network || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/url"
	"testing"
)

type networkPoolData struct {
	name              string
	description       string
	poolType          string
	networkProviderId string
	backingType       string
	backingNames      []string
	backingConstraint types.BackingUseConstraint
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
	params.Set("filter", fmt.Sprintf("virtualCenter.id==%s", vCenter.VSphereVCenter.VcId))
	portGroups, err := vcdClient.GetAllVcenterImportableDvpgs(params)
	if err != nil {
		t.Skipf("error getting port groups: %s", err)
	}

	var usableTransportZones []string
	for _, tz := range transportZones {
		if tz.AlreadyImported {
			continue
		}
		usableTransportZones = append(usableTransportZones, tz.Name)
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-geneve-name",
			description:       t.Name() + "-geneve-name description",
			poolType:          types.NetworkPoolGeneveType,
			networkProviderId: nsxtManagerId,
			backingType:       "transport_zone",
			backingNames:      []string{tz.Name},
		})
	}
	if len(usableTransportZones) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-geneve-none-first",
			description:       t.Name() + "-geneve-none-first description",
			poolType:          types.NetworkPoolGeneveType,
			networkProviderId: nsxtManagerId,
			backingType:       "none",
			backingNames:      []string{"none"},
			backingConstraint: types.BackingUseFirstAvailable,
		})
	}
	if len(usableTransportZones) == 1 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-geneve-none-one",
			description:       t.Name() + "-geneve-none-one description",
			poolType:          types.NetworkPoolGeneveType,
			networkProviderId: nsxtManagerId,
			backingType:       "none",
			backingNames:      []string{"none"},
			backingConstraint: types.BackingUseWhenOnlyOne,
		})
	}
	for _, ds := range distributedSwitches {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-vlan-name",
			description:       t.Name() + "-vlan-name description",
			poolType:          types.NetworkPoolVlanType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "distributed_switch",
			backingNames:      []string{ds.BackingRef.Name},
		})
	}
	if len(distributedSwitches) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-vlan-none",
			description:       t.Name() + "-vlan-none description",
			poolType:          types.NetworkPoolVlanType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "none",
			backingNames:      []string{"none"},
			backingConstraint: types.BackingUseFirstAvailable,
		})
	}
	var compatiblePgs []string
	for _, pg := range portGroups {

		compatible := false
		for _, pg1 := range portGroups {
			if pg.VcenterImportableDvpg.BackingRef.ID == pg1.VcenterImportableDvpg.BackingRef.ID {
				continue
			}
			if pg.UsableWith(pg1) {
				if !contains(compatiblePgs, pg.VcenterImportableDvpg.BackingRef.Name) {
					compatiblePgs = append(compatiblePgs, pg.VcenterImportableDvpg.BackingRef.Name)
				}
				if !contains(compatiblePgs, pg1.VcenterImportableDvpg.BackingRef.Name) {
					compatiblePgs = append(compatiblePgs, pg1.VcenterImportableDvpg.BackingRef.Name)
				}
				compatible = true
			}
		}
		if compatible {
			continue
		}
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-pg-name",
			description:       t.Name() + "-pg-name description",
			poolType:          types.NetworkPoolPortGroupType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "port_group",
			backingNames:      []string{pg.VcenterImportableDvpg.BackingRef.Name},
		})
	}
	if len(compatiblePgs) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-pg-multi",
			description:       t.Name() + "-pg-multi description",
			poolType:          types.NetworkPoolPortGroupType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "port_group",
			backingNames:      compatiblePgs,
		})
	}
	if len(portGroups) > 0 {
		runNetworkPoolTest(t, networkPoolData{
			name:              t.Name() + "-pg-none",
			description:       t.Name() + "-pg-none description",
			poolType:          types.NetworkPoolPortGroupType,
			networkProviderId: vCenter.VSphereVCenter.VcId,
			backingType:       "none",
			backingNames:      []string{"none"},
			backingConstraint: types.BackingUseFirstAvailable,
		})
	}

	postTestChecks(t)
}

func runNetworkPoolTest(t *testing.T, npData networkPoolData) {

	tmpl := testAccNetworkPoolWithBacking
	if npData.backingType == "none" {
		tmpl = testAccNetworkPoolNoBacking
	}
	backingElements := " "
	if len(npData.backingNames) > 1 {
		tmpl = testAccNetworkPoolMulti
		for _, name := range npData.backingNames {
			backingElements = fmt.Sprintf("%s\n%s", backingElements, fmt.Sprintf(portGroupBacking, name))
		}
	}
	testName := npData.poolType + "-" + npData.backingNames[0]
	if len(npData.backingNames) > 1 {
		testName = npData.poolType + "-multi"
	}
	if npData.backingNames[0] == "none" {
		testName = npData.poolType + "-" + npData.name
	}

	var data = StringMap{
		"SkipMessage":            " ",
		"NetworkPoolName":        npData.name,
		"NetworkPoolDescription": npData.description,
		"NetworkProviderId":      npData.networkProviderId,
		"PoolType":               npData.poolType,
		"BackingType":            npData.backingType,
		"BackingName":            npData.backingNames[0],
		"BackingConstraint":      npData.backingConstraint,
		"BackingElements":        backingElements,
		"RangeIds":               " ",
		"FuncName":               t.Name() + "-" + testName,
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
	data["SkipMessage"] = "# skip-binary-test: only for update"
	updatedText := templateFill(tmpl, data)
	debugPrintf("#[DEBUG] Update: %s", updatedText)

	t.Run(testName, func(t *testing.T) {
		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}
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
						checkWithCondition(npData.backingType != "none",
							resource.TestCheckResourceAttr(resourceDef, "backing.0."+npData.backingType+".#",
								fmt.Sprintf("%d", len(npData.backingNames)))),
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
					ResourceName:            resourceDef,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateIdFunc:       importStateIdTopHierarchy(updatedName),
					ImportStateVerifyIgnore: []string{"backing_selection_constraint"},
				},
			},
		})
	})
}

// checkWithCondition runs a check only if the leading condition is true
func checkWithCondition(condition bool, f resource.TestCheckFunc) resource.TestCheckFunc {
	if !condition {
		return func(s *terraform.State) error {
			return nil
		}
	}
	return f
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

const portGroupBacking = `
    port_group {
      name = "%s"
    }
`

const testAccNetworkPoolMulti = `
{{.SkipMessage}}
resource "vcd_network_pool" "npool" {
  name                = "{{.NetworkPoolName}}"
  description         = "{{.NetworkPoolDescription}}"
  network_provider_id = "{{.NetworkProviderId}}"
  type                = "{{.PoolType}}"

  backing {
    {{.BackingElements}}
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

  backing_selection_constraint = "{{.BackingConstraint}}"

  backing {
    {{.RangeIds}}
  }
}
`
const testAccRangeIds = `
    range_id {
      start_id = 101
      end_id   = 200
    }
`
