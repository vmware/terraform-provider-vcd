//go:build network || nsxt || gateway || org || catalog || access_control || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
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

// TODO investigate the need for delay on Org removal
func testAccCheckOrgDestroy(orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		var org *govcd.AdminOrg
		var err error
		for N := 0; N < 30; N++ {
			org, err = conn.GetAdminOrgByName(orgName)
			if err != nil && org == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("org %s was not destroyed", orgName)
		}
		if org != nil {
			return fmt.Errorf("org %s was found", orgName)
		}
		return nil
	}
}

// testCheckCatalogAndItemsExist checks that a catalog exists, and optionally that it has as many items as expected
// * checkItems defines whether we count the items or not
// * expectedItems is the total number of catalog items (includes both vApp templates and media items)
// * expectedTemplates is the number of vApp templates
// expectedMedia is the number of Media
func testCheckCatalogAndItemsExist(orgName, catalogName string, checkItems bool, expectedItems, expectedTemplates, expectedMedia int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if testAccProvider == nil || testAccProvider.Meta() == nil {
			return fmt.Errorf("testAccProvider is not initialized")
		}
		conn := testAccProvider.Meta().(*VCDClient)
		catalog, err := conn.Client.GetAdminCatalogByName(orgName, catalogName)
		if err != nil {
			return fmt.Errorf("error retrieving catalog %s/%s: %s", orgName, catalogName, err)
		}
		if !checkItems {
			return nil
		}

		if catalog.AdminCatalog.Tasks != nil {
			err = catalog.WaitForTasks()
			if err != nil {
				return err
			}
		}

		items, err := catalog.QueryCatalogItemList()
		if err != nil {
			return fmt.Errorf("error retrieving catalog item list: %s", err)
		}
		vappTemplates, err := catalog.QueryVappTemplateList()
		if err != nil {
			return fmt.Errorf("error retrieving vApp templates list: %s", err)
		}
		mediaItems, err := catalog.QueryMediaList()
		if err != nil {
			return fmt.Errorf("error retrieving media items list: %s", err)
		}
		if len(items) != expectedItems {
			return fmt.Errorf("catalog '%s' -expected %d items - found %d", catalogName, expectedItems, len(items))
		}
		if len(vappTemplates) != expectedTemplates {
			return fmt.Errorf("catalog '%s' -expected %d vApp templates - found %d", catalogName, expectedTemplates, len(vappTemplates))
		}
		if len(mediaItems) != expectedMedia {
			return fmt.Errorf("catalog '%s' -expected %d media items - found %d", catalogName, expectedMedia, len(mediaItems))
		}
		return nil
	}
}
