// +build nsxv gateway ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVcdIpSet(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":       testConfig.VCD.Org,
		"Vdc":       testConfig.VCD.Vdc,
		"IpSetName": t.Name(),
		"Tags":      "nsxv",
	}

	configText := templateFill(testAccVcdIpSet, params)

	params["FuncName"] = t.Name() + "-step1"
	params["IpSetName"] = t.Name() + "-changed"
	configText1 := templateFill(testAccVcdIpSetUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdIpSetDestroy("vcd_ipset.test-ipset", params["IpSetName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_ipset.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.2555711295", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.2329977041", "192.168.2.1"),

					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_ipset.test-ipset", "data.vcd_ipset.test-ipset", []string{}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_ipset.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "name", t.Name()+"-changed"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "description", "test-ip-set-changed-description"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.1441693733", "10.10.10.1"),
					resource.TestCheckResourceAttr("vcd_ipset.test-ipset", "ip_addresses.2766637002", "11.11.11.1"),

					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_ipset.test-ipset", "data.vcd_ipset.test-ipset", []string{}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_ipset.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, params["IpSetName"].(string)),
			},
		},
	})
}

func testAccCheckVcdIpSetDestroy(resource, ipSetName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		orgName := rs.Primary.Attributes["org"]
		vdcName := rs.Primary.Attributes["vdc"]

		_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
		if err != nil {
			return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
		}

		ipSet, err := vdc.GetNsxvIpSetByNameOrId(ipSetName)

		if !govcd.IsNotFound(err) || ipSet != nil {
			return fmt.Errorf("IP set (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccVcdIpSet = `
resource "vcd_ipset" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-description"
  ip_addresses = ["192.168.1.1","192.168.2.1"]
}

data "vcd_ipset" "test-ipset" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
  
	name         = vcd_ipset.test-ipset.name
}
`

const testAccVcdIpSetUpdate = `
resource "vcd_ipset" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-changed-description"
  ip_addresses = ["10.10.10.1","11.11.11.1"]
}

data "vcd_ipset" "test-ipset" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
  
	name         = vcd_ipset.test-ipset.name
}
`
