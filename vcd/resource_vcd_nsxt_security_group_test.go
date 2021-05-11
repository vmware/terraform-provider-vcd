// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdNsxtSecurityGroupEmpty tests out capabilities to setup Security Groups without
// attaching member networks
func TestAccVcdNsxtSecurityGroupEmpty(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccNsxtSecurityGroupEmpty, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccNsxtSecurityGroupEmpty2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group", types.FirewallGroupTypeSecurityGroup),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group-changed", types.FirewallGroupTypeSecurityGroup),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "name", "test-security-group"),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "description", "test-security-group-description"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_org_network_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_vm_ids"),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "name", "test-security-group-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "description", ""),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_org_network_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_vm_ids"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_security_group.group1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-security-group-changed"),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtSecurityGroupPrereqsEmpty = `
data "vcd_nsxt_edgegateway" "existing" {
	org  = "{{.Org}}"
	vdc  = "{{.NsxtVdc}}"

	name = "{{.EdgeGw}}"
}
`

const testAccNsxtSecurityGroupEmpty = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_security_group" "group1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name = "test-security-group"
  description = "test-security-group-description"
}
`

const testAccNsxtSecurityGroupEmpty2 = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_security_group" "group1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name = "test-security-group-changed"
}
`

// TestAccVcdNsxtSecurityGroup is similar to TestAccVcdNsxtFirewallGroupEmpty, but it also creates
// Org VDC networks and attaches them to security group.

// Additionally it tests `vcd_nsxt_security_group` datasource to save testing time and avoid creating
// the same prerequisite resources.
func TestAccVcdNsxtSecurityGroup(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccNsxtSecurityGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText1 := templateFill(testAccNsxtSecurityGroupDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	delete(params, "SkipTest")
	params["FuncName"] = t.Name() + "-step3"
	configText2 := templateFill(testAccNsxtSecurityGroupStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group", types.FirewallGroupTypeSecurityGroup),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group-changed", types.FirewallGroupTypeSecurityGroup),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "name", "test-security-group"),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "description", "test-security-group-description"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_security_group.group1", "member_vms.*", map[string]string{
						"vm_name":   "vapp-vm",
						"vapp_name": "web",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_security_group.group1", "member_vms.*", map[string]string{
						"vm_name":   "standalone-VM",
						"vapp_name": "",
					}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "name", "test-security-group"),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "description", "test-security-group-description"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_security_group.group1", "member_vms.*", map[string]string{
						"vm_name":   "vapp-vm",
						"vapp_name": "web",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_security_group.group1", "member_vms.*", map[string]string{
						"vm_name":   "standalone-VM",
						"vapp_name": "",
					}),
					// Ensure datasource has all the fields
					resourceFieldsEqual("vcd_nsxt_security_group.group1", "data.vcd_nsxt_security_group.group1", []string{}),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "name", "test-security-group-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_security_group.group1", "description", ""),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_org_network_ids"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_security_group.group1", "member_vm_ids"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_security_group.group1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-security-group-changed"),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtSecurityGroupPrereqs = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_network_routed_v2" "nsxt-backed" {
  # This value could be larger to test more members, but left 2 for the sake of testing speed
  count = 2

  org         = "{{.Org}}"
  vdc         = "{{.NsxtVdc}}"
  name        = "nsxt-routed-${count.index}"
  description = "My routed Org VDC network backed by NSX-T"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "212.1.${count.index}.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "212.1.${count.index}.10"
    end_address   = "212.1.${count.index}.20"
  }
}

# Create stanadlone VM to check membership
resource "vcd_vm" "emptyVM" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  power_on      = false
  name          = "standalone-VM"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  network {
    type               = "org"
    name               = (vcd_network_routed_v2.nsxt-backed[0].id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed[0].name)
    ip_allocation_mode = "POOL"
    is_primary         = true
  }

  depends_on = [vcd_network_routed_v2.nsxt-backed]
}

# Create a vApp and VM
resource "vcd_vapp" "web" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name     = "web"
  power_on = false
}

resource "vcd_vapp_org_network" "vappOrgNet" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  vapp_name        = vcd_vapp.web.name
  org_network_name = (vcd_network_routed_v2.nsxt-backed[1].id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed[1].name)

  depends_on = [vcd_vapp.web, vcd_network_routed_v2.nsxt-backed]
}

resource "vcd_vapp_vm" "emptyVM" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  power_on      = false
  vapp_name     = vcd_vapp.web.name
  name          = "vapp-vm"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  network {
    type               = "org"
    name               = (vcd_vapp_org_network.vappOrgNet.id == "always-not-equal" ? null : vcd_vapp_org_network.vappOrgNet.org_network_name)
    ip_allocation_mode = "POOL"
    is_primary         = true
  }

  depends_on = [vcd_vapp_org_network.vappOrgNet, vcd_network_routed_v2.nsxt-backed]
}
`

const testAccNsxtSecurityGroup = testAccNsxtSecurityGroupPrereqs + `
resource "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-security-group"
  description = "test-security-group-description"

  member_org_network_ids = vcd_network_routed_v2.nsxt-backed.*.id

  depends_on = [vcd_vapp_vm.emptyVM, vcd_vm.emptyVM]
}
`

const testAccNsxtSecurityGroupStep3 = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name = "test-security-group-changed"
}
`

const testAccNsxtSecurityGroupDatasource = testAccNsxtSecurityGroup + `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file
data "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "test-security-group"
}
`

// TestAccVcdNsxtSecurityGroupInvalidConfigs is expected to fail when:
// * NSX-V Edge Gateway ID is supplied
// * Invalid (non existent) Edge Gateway ID is presented
// * Isolated Org Vdc network added as a member
func TestAccVcdNsxtSecurityGroupInvalidConfigs(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// This test is meant to fail
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Networking.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccVcdNsxtSecurityGroupIncorrectEdgeGateway, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText1 := templateFill(testAccVcdNsxtSecurityGroupIncorrectEdgeGatewayStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText2 := templateFill(testAccVcdNsxtSecurityGroupIncorrectEdgeGatewayStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group", types.FirewallGroupTypeSecurityGroup),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText,
				ExpectError: regexp.MustCompile(`please use 'vcd_nsxt_edgegateway' for NSX-T backed VDC`),
			},
			resource.TestStep{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`error creating NSX-T Security Group`),
			},
			resource.TestStep{
				Config:      configText2,
				ExpectError: regexp.MustCompile(`Not all member network IDs reference to Routed Org networks`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtSecurityGroupIncorrectEdgeGateway = `
data "vcd_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_edgegateway.existing.id

  name        = "test-security-group"
  description = "test-security-group-description"
}
`

const testAccVcdNsxtSecurityGroupIncorrectEdgeGatewayStep2 = `
resource "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  # A correct syntax of non existing NSX-T Edge Gateway
  edge_gateway_id = "urn:vcloud:gateway:71df3e4b-6da9-404d-8e44-1111111c1c38"

  name        = "test-security-group"
  description = "test-security-group-description"
}
`

const testAccVcdNsxtSecurityGroupIncorrectEdgeGatewayStep3 = `
resource "vcd_network_isolated_v2" "nsxt-backed" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name        = "nsxt-isolated-test"
  description = "My isolated Org VDC network backed by NSX-T"

  gateway       = "52.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "52.1.1.10"
    end_address   = "52.1.1.20"
  }

}
resource "vcd_nsxt_security_group" "group1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  # A correct syntax of non existing NSX-T Edge Gateway
  edge_gateway_id = "urn:vcloud:gateway:71df3e4b-6da9-404d-8e44-1111111c1c38"

  name        = "test-security-group"
  description = "test-security-group-description"

  member_org_network_ids = [vcd_network_isolated_v2.nsxt-backed.id]
}
`

func testAccCheckNsxtFirewallGroupDestroy(vdcName, firewalGroupName, firewallGroupType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetNsxtFirewallGroupByName(firewalGroupName, firewallGroupType)
		if err == nil {
			return fmt.Errorf("firewall group '%s' of type '%s' still exists", firewalGroupName, firewallGroupType)
		}

		return nil
	}
}
