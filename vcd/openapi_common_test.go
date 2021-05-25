// +build network nsxt ALL functional

package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckOpenApiVcdNetworkDestroy(vdcName, networkName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetOpenApiOrgVdcNetworkByName(networkName)
		if err == nil {
			return fmt.Errorf("network %s still exists", networkName)
		}

		return nil
	}
}

func testAccCheckOpenApiNsxtAppPortDestroy(appPortProfileName, scope string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, appPortProfileName, testConfig.VCD.Org, err)
		}

		_, err = org.GetNsxtAppPortProfileByName(appPortProfileName, scope)
		if err == nil {
			return fmt.Errorf("'%s' Application Port Profile still exists", appPortProfileName)
		}

		return nil
	}
}
