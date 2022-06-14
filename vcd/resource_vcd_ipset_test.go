//go:build nsxv || gateway || ALL || functional
// +build nsxv gateway ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdIpSet(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":       testConfig.VCD.Org,
		"Vdc":       testConfig.VCD.Vdc,
		"IpSetName": t.Name(),
		"Tags":      "nsxv",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdIpSet, params)

	params["FuncName"] = t.Name() + "-step1"
	params["IpSetName"] = t.Name() + "-changed"
	configText1 := templateFill(testAccVcdIpSetUpdate, params)

	params["FuncName"] = t.Name() + "-step2"
	params["IpSetName"] = t.Name() + "-changed2"
	configText2 := templateFill(testAccVcdIpSetUpdate2, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdIpSetDestroy("vcd_nsxv_ip_set.test-ipset", params["IpSetName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_ip_set.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "is_inheritance_allowed", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "192.168.1.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "192.168.2.1"),

					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_nsxv_ip_set.test-ipset", "data.vcd_nsxv_ip_set.test-ipset", []string{}),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_ip_set.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "name", t.Name()+"-changed"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "is_inheritance_allowed", "true"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "description", "test-ip-set-changed-description"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "10.10.10.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "11.11.11.1"),

					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_nsxv_ip_set.test-ipset", "data.vcd_nsxv_ip_set.test-ipset", []string{}),
				),
			},
			{
				ResourceName:      "vcd_nsxv_ip_set.test-ipset",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, "TestAccVcdIpSet-changed"),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_ip_set.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "name", t.Name()+"-changed2"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "is_inheritance_allowed", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "description", "test-ip-set-changed-description"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "1.1.1.1/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "10.10.10.100-10.10.10.110"),
					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_nsxv_ip_set.test-ipset", "data.vcd_nsxv_ip_set.test-ipset", []string{}),
				),
			},
			{
				Config: configText2,
				Taint:  []string{"vcd_nsxv_ip_set.test-ipset"}, // Force provisioning from scratch instead of update
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_ip_set.test-ipset", "id", regexp.MustCompile(`.*ipset-\d*$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "name", t.Name()+"-changed2"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "is_inheritance_allowed", "false"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "description", "test-ip-set-changed-description"),
					resource.TestCheckResourceAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "1.1.1.1/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_ip_set.test-ipset", "ip_addresses.*", "10.10.10.100-10.10.10.110"),
					// Validate that datasource has all the same fields
					resourceFieldsEqual("vcd_nsxv_ip_set.test-ipset", "data.vcd_nsxv_ip_set.test-ipset", []string{}),
				),
			},
		},
	})
	postTestChecks(t)
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
resource "vcd_nsxv_ip_set" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-description"
  ip_addresses = ["192.168.1.1","192.168.2.1"]
}

data "vcd_nsxv_ip_set" "test-ipset" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
  
	name         = vcd_nsxv_ip_set.test-ipset.name
	depends_on   = [vcd_nsxv_ip_set.test-ipset]
}
`

const testAccVcdIpSetUpdate = `
resource "vcd_nsxv_ip_set" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-changed-description"
  ip_addresses = ["10.10.10.1","11.11.11.1"]
}

data "vcd_nsxv_ip_set" "test-ipset" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
  
	name         = vcd_nsxv_ip_set.test-ipset.name
	depends_on   = [vcd_nsxv_ip_set.test-ipset]
}
`

const testAccVcdIpSetUpdate2 = `
resource "vcd_nsxv_ip_set" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name                   = "{{.IpSetName}}"
  is_inheritance_allowed = false
  description            = "test-ip-set-changed-description"
  ip_addresses           = ["1.1.1.1/24","10.10.10.100-10.10.10.110"]
}

data "vcd_nsxv_ip_set" "test-ipset" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
  
	name         = vcd_nsxv_ip_set.test-ipset.name
	depends_on   = [vcd_nsxv_ip_set.test-ipset]
}
`
