//go:build network || nsxt || gateway || org || catalog || access_control || networkPool || providerVdc || vm || vapp || standaloneVm || user || multisite || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"strings"
	"testing"
	"time"
)

// updateEdgeGatewayTier0Dedication updates default NSX-T Edge Gateway used for tests to use Tier-0
// gateway as dedicated for the purpose of some tests
func updateEdgeGatewayTier0Dedication(t *testing.T, dedicatedTier0 bool) {
	vcdClient := createSystemTemporaryVCDConnection()
	org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Fatalf("error retrieving Org '%s': %s", testConfig.VCD.Org, err)
	}
	edge, err := org.GetNsxtEdgeGatewayByName(testConfig.Nsxt.EdgeGateway)
	if err != nil {
		t.Fatalf("error retrieving NSX-T Edge Gateway '%s': %s", testConfig.Nsxt.EdgeGateway, err)
	}

	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = dedicatedTier0
	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		t.Fatalf("error updating NSX-T Edge Gateway dedicated Tier 0 gateway usage to '%t': %s", dedicatedTier0, err)
	}
}

func checkNetworkPoolExists(networkPoolName string, wantExisting bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_network_pool" {
				continue
			}
			_, err := conn.GetNetworkPoolByName(networkPoolName)
			if wantExisting {
				if err != nil {
					return fmt.Errorf("network pool %s not found: %s ", networkPoolName, err)
				}
			} else {
				if err == nil {
					return fmt.Errorf("network pool %s not deleted yet", networkPoolName)
				} else {
					return nil
				}
			}
		}
		return nil
	}
}

func testAccCheckOrgDestroy(orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		adminOrg, err := runWithRetry(
			"[testAccCheckOrgDestroy] organization retrieval",
			"[testAccCheckOrgDestroy] error retrieving org",
			10*time.Second,
			nil,
			func() (any, error) {
				return conn.GetAdminOrgByName(orgName)
			},
		)
		if govcd.IsNotFound(err) {
			return fmt.Errorf("org %s was not destroyed", orgName)
		}
		if adminOrg != nil {
			return fmt.Errorf("org %s was found", orgName)
		}
		return nil
	}
}

func checkListForKnownItem(resName, target, unwanted string, isWanted, importing bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// If we didn't indicate any known item, the check is always true, even if no item was returned
		if target == "" {
			return nil
		}

		resourcePath := "data.vcd_resource_list." + resName

		res, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("resource %s not found", resName)
		}

		var list = make([]string, 0)

		for key, value := range res.Primary.Attributes {
			if strings.HasPrefix(key, "list.") {
				list = append(list, value)
			}
		}

		for _, item := range list {
			// if we want ANY item, the comparison is true as long as at least one was found
			found := item == target || target == "*"
			if unwanted != "" && item == unwanted {
				return fmt.Errorf("found unwanted item '%s'", unwanted)
			}
			if importing {
				found = strings.Contains(item, target) || target == "*"
			}
			if found {
				if isWanted {
					return nil
				} else {
					return fmt.Errorf("item '%s' found in '%s'", target, resName)
				}
			}
		}
		if isWanted {
			return fmt.Errorf("item '%s' not found in list %s", target, resourcePath)
		} else {
			return nil
		}
	}
}
