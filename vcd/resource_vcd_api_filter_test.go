//go:build functional || openapi || ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdApiFilter(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Vendor":            "vmware",
		"Name":              t.Name(),
		"Version1":          "1.0.0",
		"Version2":          "1.0.1",
		"Enabled":           false, // This makes things easier as it can be destroyed at any time
		"RootUrl":           "https://www.broadcom.com",
		"ExternalEndpoint":  "vcd_external_endpoint.ep1.id",
		"UrlMatcherPattern": "/test/.*",
		"UrlMatcherScope":   "EXT_API",
		"FuncName":          t.Name() + "Step1",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccCheckVcdApiFilter, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s", configText1)

	params["FuncName"] = t.Name() + "Step2"
	params["ExternalEndpoint"] = "vcd_external_endpoint.ep2.id"
	params["UrlMatcherPattern"] = "/test/update/.*"
	params["UrlMatcherScope"] = "EXT_UI_PROVIDER"
	configText2 := templateFill(testAccCheckVcdApiFilter, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s", configText2)

	params["FuncName"] = t.Name() + "Step3"
	configText3 := templateFill(testAccCheckVcdApiFilterDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_api_filter.af"
	cachedId := testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckApiFilterDestroy(cachedId.String()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`urn:vcloud:apiFilter:.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "external_endpoint_id", "vcd_external_endpoint.ep1", "id"),
					resource.TestCheckResourceAttr(resourceName, "url_matcher_pattern", "/test/.*"),
					resource.TestCheckResourceAttr(resourceName, "url_matcher_scope", "EXT_API"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "external_endpoint_id", "vcd_external_endpoint.ep2", "id"),
					resource.TestCheckResourceAttr(resourceName, "url_matcher_pattern", "/test/update/.*"),
					resource.TestCheckResourceAttr(resourceName, "url_matcher_scope", "EXT_UI_PROVIDER"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(resourceName, "id"),
					resourceFieldsEqual(resourceName, "data.vcd_api_filter.af-ds", []string{"api_filter_id", "%"}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cachedId.fieldValue, nil
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ExpectError:       regexp.MustCompile(`.*1.+urn:vcloud:apiFilter:.+EXT_UI_PROVIDER.+/test/update/.*`),
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("list@%s%s%s%s%s", params["Vendor"], ImportSeparator, params["Name"], ImportSeparator, params["Version2"]), nil
				},
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckApiFilterDestroy(id string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_api_filter" && rs.Primary.ID != id {
				continue
			}

			conn := testAccProvider.Meta().(*VCDClient)
			_, err := conn.GetApiFilterById(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("API Filter '%s' still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

const testAccCheckVcdApiFilter = `
resource "vcd_external_endpoint" "ep1" {
  vendor      = "{{.Vendor}}"
  name        = "{{.Name}}"
  version     = "{{.Version1}}"
  enabled     = {{.Enabled}}
  root_url    = "{{.RootUrl}}"
}
resource "vcd_external_endpoint" "ep2" {
  vendor      = "{{.Vendor}}"
  name        = "{{.Name}}"
  version     = "{{.Version2}}"
  enabled     = {{.Enabled}}
  root_url    = "{{.RootUrl}}"
}

resource "vcd_api_filter" "af" {
  external_endpoint_id = {{.ExternalEndpoint}}
  url_matcher_pattern  = "{{.UrlMatcherPattern}}"
  url_matcher_scope    = "{{.UrlMatcherScope}}"
}
`

const testAccCheckVcdApiFilterDS = testAccCheckVcdApiFilter + `
data "vcd_api_filter" "af-ds" {
  api_filter_id = vcd_api_filter.af.id
}
`
