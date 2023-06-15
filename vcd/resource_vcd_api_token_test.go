//go:build api || ALL || functional

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdApiToken(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TokenName": t.Name(),
		"FileName":  t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdApiToken, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceName := "vcd_api_token.custom"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApiTokenDestroy(params["TokenName"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					testCheckFileExists(t.Name()),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdApiToken = `
resource "vcd_api_token" "custom" {
  name = "{{.TokenName}}"		

  file_name = "{{.FileName}}"
}
`

func testCheckFileExists(filename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		return nil
	}
}

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
