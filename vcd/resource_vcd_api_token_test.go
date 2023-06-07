//go:build api || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdApiToken(t *testing.T) {
	skipIfNotSysAdmin(t)

	preTestChecks(t)

	var params = StringMap{
		"Org":   testConfig.Provider.SysOrg,
		"Token": t.Name(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      testAccCheckApiTokenDestroy(params["Token"].(string)),
		Steps: []resource.TestStep{
			{},
		},
	})
}

// const testAccVcdApiToken_sysorg = `
// resource

// `

func testAccCheckApiTokenDestroy(tokenName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_api_token" && rs.Primary.Attributes["name"] != tokenName {
				continue
			}

			_, err := conn.GetTokenById(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Token still exist")
			}

			return nil
		}

		return nil
	}
}
