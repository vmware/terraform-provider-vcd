//go:build api || ALL || functional

package vcd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdApiToken(t *testing.T) {
	preTestChecks(t)
	skipTestForServiceAccountAndApiToken(t)

	var params = StringMap{
		"TokenName": t.Name(),
		"FileName":  t.Name(),
	}
	testParamsNotEmpty(t, params)

	filename := params["FileName"].(string)

	configText := templateFill(testAccVcdApiToken, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	t.Cleanup(deleteApiTokenFile(filename, t))
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceName := "vcd_api_token.custom"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckApiTokenDestroy(params["TokenName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					testCheckFileExists(params["FileName"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

// #nosec G101 -- No hardcoded credentials here
const testAccVcdApiToken = `
resource "vcd_api_token" "custom" {
  name = "{{.TokenName}}"		

  file_name        = "{{.FileName}}"
  allow_token_file = true
}
`

// This is a helper function that attempts to remove created API token file no matter of the test outcome
func deleteApiTokenFile(filename string, t *testing.T) func() {
	return func() {
		err := os.Remove(filename)
		if err != nil {
			t.Errorf("Failed to delete file: %s", err)
		}
	}
}

func testCheckFileExists(filename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		filename = filepath.Clean(filename)
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
			if rs.Type != "vcd_api_token" || rs.Primary.Attributes["name"] != tokenName {
				continue
			}

			_, err := conn.GetTokenById(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("error: api token still exists post-destroy")
			}

			return nil
		}

		return nil
	}
}
