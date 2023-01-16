//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdNsxtDynamicSecurityGroupVdcGroupEmpty
//
// Dynamic security groups __can only be scoped to VDC Groups__ and a self explanatory error is
// returned if they are scoped to Edge Gateway in a VDC only:
//
// Error: [nsxt dynamic security group create] error creating NSX-T dynamic security group
// 'test-dynamic-security-group': error creating NSX-T Firewall Group: error in HTTP POST request:
// BAD_REQUEST - [ 27e7eea1-0f65-4147-93e6-b6ecc7ecc61f ] Firewall Group test-dynamic-security-group
// cannot have type VM_CRITERIA unless it is scoped to a VDC Group
func TestAccVcdNsxtDynamicSecurityGroupVdcGroupEmpty(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"VdcGroup":    testConfig.Nsxt.VdcGroup,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxtDynamicSecurityGroupEmpty, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccNsxtDynamicSecurityGroupEmpty2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtDynamicSecurityGroupEmpty2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-dynamic-security-group", types.FirewallGroupTypeVmCriteria),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group-changed", types.FirewallGroupTypeVmCriteria),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "name", "test-dynamic-security-group"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "description", "test-security-group-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "member_vms.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "criteria.#", "0"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "name", "test-dynamic-security-group-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "member_vms.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "criteria.#", "0"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_dynamic_security_group.group1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(testConfig.Nsxt.VdcGroup, "test-dynamic-security-group-changed"),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group1", "vcd_nsxt_dynamic_security_group.group1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtDynamicSecurityGroupPrereqsEmpty = `
data "vcd_vdc_group" "group1" {
  org  = "{{.Org}}"
  name = "{{.VdcGroup}}"
}
`

const testAccNsxtDynamicSecurityGroupEmpty = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "test-dynamic-security-group"
  description = "test-security-group-description"
}
`

const testAccNsxtDynamicSecurityGroupEmpty2 = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name            = "test-dynamic-security-group-changed"
}
`

const testAccNsxtDynamicSecurityGroupEmpty2DS = testAccNsxtDynamicSecurityGroupEmpty2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name            = "test-dynamic-security-group-changed"
}
`

// TestAccVcdNsxtDynamicSecurityGroupMaximumCriteria Tests our maximum configuration for the NSX-T
// dynamic firewall group - 3 criteria each containing 4 rules.
func TestAccVcdNsxtDynamicSecurityGroupVdcGroupMaximumCriteria(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"NsxtVdc":  testConfig.Nsxt.Vdc,
		"VdcGroup": testConfig.Nsxt.VdcGroup,
		"EdgeGw":   testConfig.Nsxt.VdcGroupEdgeGateway,
		"TestName": t.Name(),
		"Tags":     "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxtDynamicSecurityGroupMaximumCriteria, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configTextDS := templateFill(testAccNsxtDynamicSecurityGroupMaximumCriteriaDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configTextDS)

	params["FuncName"] = t.Name() + "-step2"
	configText1 := templateFill(testAccNsxtDynamicSecurityGroupMaximumCriteria2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group", types.FirewallGroupTypeSecurityGroup),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group-changed", types.FirewallGroupTypeSecurityGroup),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "name", "test-dynamic-security-group"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "member_vms.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "criteria.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_TAG",
						"operator": "EQUALS",
						"value":    "tag-equals",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_TAG",
						"operator": "CONTAINS",
						"value":    "tag-contains",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_TAG",
						"operator": "STARTS_WITH",
						"value":    "starts_with",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_TAG",
						"operator": "ENDS_WITH",
						"value":    "ends_with",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "CONTAINS",
						"value":    "name-contains2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "STARTS_WITH",
						"value":    "starts_with2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "CONTAINS",
						"value":    "name-contains22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "STARTS_WITH",
						"value":    "starts_with22",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "CONTAINS",
						"value":    "name-contains3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "STARTS_WITH",
						"value":    "starts_with3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "CONTAINS",
						"value":    "name-contains33",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "STARTS_WITH",
						"value":    "starts_with33",
					}),
				),
			},
			{
				Config: configTextDS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group1", "vcd_nsxt_dynamic_security_group.group1", nil),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "name", "test-dynamic-security-group-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "member_vms.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "STARTS_WITH",
						"value":    "starts_with3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_dynamic_security_group.group1", "criteria.*.rule.*", map[string]string{
						"type":     "VM_NAME",
						"operator": "CONTAINS",
						"value":    "name-contains33",
					}),
				),
			},
			{
				ResourceName:      "vcd_nsxt_dynamic_security_group.group1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(testConfig.Nsxt.VdcGroup, "test-dynamic-security-group-changed"),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtDynamicSecurityGroupMaximumCriteria = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "test-dynamic-security-group"

  criteria {
	rule {
	  type = "VM_TAG"
	  operator = "EQUALS"
	  value = "tag-equals"
	}

	rule {
	  type = "VM_TAG"
	  operator = "CONTAINS"
	  value = "tag-contains"
	}

	rule {
	  type = "VM_TAG"
	  operator = "STARTS_WITH"
	  value = "starts_with"
	}

	rule {
	  type = "VM_TAG"
	  operator = "ENDS_WITH"
	  value = "ends_with"
	}
  }

  criteria {
	rule {
	  type     = "VM_NAME"
	  operator = "CONTAINS"
	  value    = "name-contains2"
	}

	rule {
	  type     = "VM_NAME"
	  operator = "STARTS_WITH"
	  value    = "starts_with2"
	}

	rule {
		type     = "VM_NAME"
		operator = "CONTAINS"
		value    = "name-contains22"
	  }
  
	rule {
		type     = "VM_NAME"
		operator = "STARTS_WITH"
		value    = "starts_with22"
	}
  }

  criteria {
	rule {
	  type     = "VM_NAME"
	  operator = "CONTAINS"
	  value    = "name-contains3"
	}

	rule {
	  type     = "VM_NAME"
	  operator = "STARTS_WITH"
	  value    = "starts_with3"
	}

	rule {
	  type     = "VM_NAME"
	  operator = "CONTAINS"
	  value    = "name-contains33"
	}

	rule {
	  type     = "VM_NAME"
	  operator = "STARTS_WITH"
	  value    = "starts_with33"
	}
  }
}
`

const testAccNsxtDynamicSecurityGroupMaximumCriteriaDS = testAccNsxtDynamicSecurityGroupMaximumCriteria + `
# skip-binary-test: Data Source test
data "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "test-dynamic-security-group"
}
`

const testAccNsxtDynamicSecurityGroupMaximumCriteria2 = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "test-dynamic-security-group-changed"

  criteria {
	rule {
	  type     = "VM_NAME"
	  operator = "CONTAINS"
	  value    = "name-contains33"
	}

	rule {
	  type     = "VM_NAME"
	  operator = "STARTS_WITH"
	  value    = "starts_with3"
	}
  }
}
`

// TestAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVms tests out the dynamic security groups
// matching existing VMs both - based on Tags and on VM names.
func TestAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVms(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"NsxtVdc":  testConfig.Nsxt.Vdc,
		"VdcGroup": testConfig.Nsxt.VdcGroup,
		"EdgeGw":   testConfig.Nsxt.VdcGroupEdgeGateway,
		"TestName": "DSGVdcGroupCritWithVms", // Shortened name instead of t.Name() as it could be the reason sometimes VdcGroup.GetNsxtFirewallGroupByName fails complaining about the filter
		"Tags":     "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsTags, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsTagsDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsNames, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4DS := templateFill(testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsNamesDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group", types.FirewallGroupTypeSecurityGroup),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-security-group-changed", types.FirewallGroupTypeSecurityGroup),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group2", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group3", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group4", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
				),
			},
			{
				// VM membership is not immediately updated by VCD, therefore we apply the same step to check that VM counts are updated
				Config:    configText2DS,
				PreConfig: func() { time.Sleep(time.Second * 15) }, // Sleeping additional 15 seconds to be sure Member VMs are populated by VCD
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group2", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group3", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group4", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "member_vms.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group2", "member_vms.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group3", "member_vms.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group4", "member_vms.#", "1"),

					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group1", "vcd_nsxt_dynamic_security_group.group1", nil),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group2", "vcd_nsxt_dynamic_security_group.group2", nil),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group3", "vcd_nsxt_dynamic_security_group.group3", nil),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group4", "vcd_nsxt_dynamic_security_group.group4", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group5", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group6", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group5", "member_vms.#", "1"),
					// resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group6", "member_vms.#", "2"),
				),
			},
			{
				// Apply the same config twice to give time for `member_vms` to be updated
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group5", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group6", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group5", "member_vms.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group6", "member_vms.#", "2"),

					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group5", "vcd_nsxt_dynamic_security_group.group5", nil),
					resourceFieldsEqual("data.vcd_nsxt_dynamic_security_group.group6", "vcd_nsxt_dynamic_security_group.group6", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsPrereqs = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_network_isolated_v2" "nsxt-backed" {
  org = "{{.Org}}"
  owner_id = data.vcd_vdc_group.group1.id

  name        = "{{.TestName}}-network"
  description = "Isolated network for dynamic security group membership tests"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.40"
  }
}

resource "vcd_vm" "emptyVM" {
  org = "{{.Org}}"
  vdc = tolist(data.vcd_vdc_group.group1.participating_org_vdcs)[0].vdc_name

  name             = "{{.TestName}}-standalone"
  computer_name    = "emptyVM"
  power_on         = false
  memory           = 512
  cpus             = 1
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  network {
    type               = "org"
    name               = vcd_network_isolated_v2.nsxt-backed.name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }
}

resource "vcd_vapp" "web" {
  org  = "{{.Org}}"
  vdc  = tolist(data.vcd_vdc_group.group1.participating_org_vdcs)[0].vdc_name
  name = "{{.TestName}}-vapp"
}

resource "vcd_vapp_org_network" "vappOrgNet" {
  org  = "{{.Org}}"
  vdc  = tolist(data.vcd_vdc_group.group1.participating_org_vdcs)[0].vdc_name

  vapp_name        = vcd_vapp.web.name
  org_network_name = vcd_network_isolated_v2.nsxt-backed.name
}

resource "vcd_vapp_vm" "emptyVM" {
  org  = "{{.Org}}"
  vdc  = tolist(data.vcd_vdc_group.group1.participating_org_vdcs)[0].vdc_name

  vapp_name        = vcd_vapp.web.name
  name             = "{{.TestName}}-vapp-vm"
  computer_name    = "emptyVM"
  power_on         = false
  memory           = 512
  cpus             = 1
  cpu_cores        = 1
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappOrgNet.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }
}
`

const testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsTags = testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsPrereqs + `
resource "vcd_security_tag" "tag1" {
  org    = "{{.Org}}"
  name   = "tag1"
  vm_ids = [vcd_vm.emptyVM.id]
}

resource "vcd_security_tag" "tag2" {
  org    = "{{.Org}}"
  name   = "tag2"
  vm_ids = [vcd_vm.emptyVM.id, vcd_vapp_vm.emptyVM.id]
}

# This group should match single tag and contain 1 VM
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-1"

  criteria {
	rule {
	  type     = "VM_TAG"
	  operator = "EQUALS"
	  value    = vcd_security_tag.tag1.name
	}
  }
}

# This group should match both tags and contain 2 VMs
resource "vcd_nsxt_dynamic_security_group" "group2" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-2"

  criteria {
    rule {
      type     = "VM_TAG"
      operator = "CONTAINS"
      value    = "ag"
	}
  }
}

# This group should match tag1, tag2 and contain 2 VMs
resource "vcd_nsxt_dynamic_security_group" "group3" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-3"

  criteria {
    rule {
      type     = "VM_TAG"
      operator = "STARTS_WITH"
      value    = "t"
	}
  }
}

# This group should match only tag1 and contain 1 VMs
resource "vcd_nsxt_dynamic_security_group" "group4" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-4"

  criteria {
    rule {
      type     = "VM_TAG"
      operator = "ENDS_WITH"
      value    = "1"
	}
  }
}
`

const testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsTagsDS = testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsTags + `
# skip-binary-test: Data Source test
data "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-1"
}

data "vcd_nsxt_dynamic_security_group" "group2" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-2"
}

data "vcd_nsxt_dynamic_security_group" "group3" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-3"
}

data "vcd_nsxt_dynamic_security_group" "group4" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-4"
}
`

const testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsNames = testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsPrereqs + `
# This group should match 1 VM
resource "vcd_nsxt_dynamic_security_group" "group5" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-5"

  criteria {
	rule {
	  type     = "VM_NAME"
	  operator = "CONTAINS"
	  value    = "standalone"
	}
  }
}

# This group should match 2 VMs
resource "vcd_nsxt_dynamic_security_group" "group6" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-6"

  criteria {
    rule {
      type     = "VM_NAME"
      operator = "STARTS_WITH"
      value    = "{{.TestName}}"
	}
  }
}
`

const testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsNamesDS = testAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVmsNames + `
# skip-binary-test: Data Source test
# This group should match 1 VM
data "vcd_nsxt_dynamic_security_group" "group5" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-5"
}

data "vcd_nsxt_dynamic_security_group" "group6" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-6"
}
`

// TestAccVcdNsxtDynamicSecurityGroupIntegartion tests out how Dynamic security groups integrate
// with firewalls - both - distributed and simple ones.
//
// Note. Dynamic security groups can only be created when an NSX-T Edge Gateway is a member of VDC
// Group, but when it is - Dynamic Security Groups can be used in regular Edge Gateway firewalls.
func TestAccVcdNsxtDynamicSecurityGroupIntegration(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"NsxtVdc":  testConfig.Nsxt.Vdc,
		"VdcGroup": testConfig.Nsxt.VdcGroup,
		"EdgeGw":   testConfig.Nsxt.VdcGroupEdgeGateway,
		"TestName": t.Name(),
		"Tags":     "network nsxt",
	}
	testParamsNotEmpty(t, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	configText := templateFill(testAccVcdNsxtDynamicSecurityGroupIntegration, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-dynamic-security-group", types.FirewallGroupTypeVmCriteria),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_dynamic_security_group.group1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_dynamic_security_group.group1", "name", "TestAccVcdNsxtDynamicSecurityGroupIntegration-group1"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtDynamicSecurityGroupIntegration = testAccNsxtDynamicSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_dynamic_security_group" "group1" {
  org          = "{{.Org}}"
  vdc_group_id = data.vcd_vdc_group.group1.id

  name = "{{.TestName}}-group1"
  criteria {
	rule {
		type     = "VM_NAME"
		operator = "STARTS_WITH"
		value    = "{{.TestName}}"
	}
  }
}

resource "vcd_nsxt_distributed_firewall" "t1" {
	org          = "{{.Org}}"
	vdc_group_id = data.vcd_vdc_group.group1.id
  
	rule {
	  name        = "rule1"
	  action      = "REJECT"
	  description = "description"
  
	  source_ids = [vcd_nsxt_dynamic_security_group.group1.id]
	}
  
	rule {
	  name        = "rule3"
	  action      = "DROP"
	  ip_protocol = "IPV4"
	}
}

data "vcd_nsxt_edgegateway" "edgegw" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.group1.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.edgegw.id

  rule {
    action      = "ALLOW"
    name        = "test_rule"
    direction   = "IN"
    ip_protocol = "IPV4"
    source_ids  = [vcd_nsxt_dynamic_security_group.group1.id]
    enabled     = true
  }
}
`
