//go:build functional || vapp || ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TestAccVcdVappNetworkChildrenResourceNotFound checks that deletion of parent vApp network
// is correctly handled when resource disappears (remove ID by using d.SetId("") instead of throwing
// error) outside of Terraform control. The following resources are verified here:
// * vcd_vapp_firewall_rules
// * vcd_vapp_nat_rules
// * vcd_vapp_static_routing
func TestAccVcdVappNetworkChildrenResourceNotFound(t *testing.T) {
	preTestChecks(t)

	// This test invokes go-vcloud-director SDK directly therefore it should not run binary tests
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"DefaultAction": "drop",
		"Name1":         "Name1",
		"VmName1":       t.Name() + "-1",
		"NetworkName":   "TestAccVcdVAppVmNet",
		"VappName":      t.Name() + "_vapp",
		"Tags":          "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVappFirewallRulesResourceNotFound, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	cachedvAppId := &testCachedFieldValue{}
	cachedvAppNetworkId := &testCachedFieldValue{}
	firewallResourceName := "vcd_vapp_firewall_rules.fw1"
	natResourceName := "vcd_vapp_nat_rules.nat1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(firewallResourceName, "id"),
					resource.TestCheckResourceAttrSet(natResourceName, "id"),
					cachedvAppId.cacheTestResourceFieldValue("vcd_vapp.vapp1", "id"),
					cachedvAppNetworkId.cacheTestResourceFieldValue("vcd_vapp_network.vappRoutedNet", "id"),

					testAccCheckVcdVappFirewallRulesExists(firewallResourceName, params["Name1"].(string)),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()
					org, err := vcdClient.GetAdminOrgByName(params["Org"].(string))
					if err != nil {
						t.Errorf("error: could not find Org: %s", err)
					}
					vdc, err := org.GetVDCByName(params["Vdc"].(string), false)
					if err != nil {
						t.Errorf("error: could not find VDC: %s", err)
					}

					vapp, err := vdc.GetVAppById(cachedvAppId.fieldValue, false)
					if err != nil {
						t.Errorf("could not find vApp %s: %s", cachedvAppId.fieldValue, err)
					}

					_, err = vapp.RemoveNetwork(cachedvAppNetworkId.fieldValue)
					if err != nil {
						t.Errorf("could not remove vApp network %s: %s", cachedvAppNetworkId.fieldValue, err)
					}

				},
				// Expecting to get a non-empty plan because resource was removed using SDK in
				// PreConfig
				Config:             configText,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVappFirewallRulesResourceNotFound = `
resource "vcd_network_routed" "network_routed" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "vapp1" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "vappRoutedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "vapp-routed-net"
  vapp_name        = vcd_vapp.vapp1.name
  gateway          = "192.168.2.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.network_routed.name
}

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vapp1.name
  org_network_name = vcd_network_routed.network_routed.name
  is_fenced        = true
}

resource "vcd_vapp_firewall_rules" "fw1" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  vapp_id        = vcd_vapp.vapp1.id
  default_action = "{{.DefaultAction}}"
  network_id     = vcd_vapp_network.vappRoutedNet.id

  log_default_action = true

  enabled = true

  rule {
    name             = "{{.Name1}}"
    policy           = "drop"
    protocol         = "udp"
    destination_port = "21"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.0/24"
  }
}

resource "vcd_vapp_nat_rules" "nat1" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.vapp1.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  nat_type   = "ipTranslation"
  enabled    = true

  rule {
    mapping_mode = "automatic" 
    vm_nic_id    = 0
    vm_id        = vcd_vapp_vm.vm1.id
  }
}

resource "vcd_vapp_vm" "vm1" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name   = vcd_vapp.vapp1.name
  name        = "{{.VmName1}}"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1 
  power_on    = false
  
  os_type                        = "sles10_64Guest"
  hardware_version               = "vmx-11"
  computer_name                  = "compName"
  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip                 = "192.168.2.11"
  }
}

resource "vcd_vapp_static_routing" "sr1" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.vapp1.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  enabled     = true

  rule {
    name         = "rule1"
    network_cidr = "10.10.0.0/24"
    next_hop_ip  = "10.10.102.3"
  }
}
`

// TestAccVcdVappChildrenResourceNotFound checks that deletion of parent vApp
// is correctly handled when resource disappears (remove ID by using d.SetId("") instead of throwing
// error) outside of Terraform control. The following resources are verified here:
// * vcd_vapp_network
// * vcd_vapp_org_network
func TestAccVcdVappChildrenResourceNotFound(t *testing.T) {
	preTestChecks(t)

	// This test invokes go-vcloud-director SDK directly therefore it should not run binary tests
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"DefaultAction": "drop",
		"Name1":         "Name1",
		"VmName1":       t.Name() + "-1",
		"NetworkName":   "TestAccVcdVAppVmNet",
		"VappName":      t.Name() + "_vapp",
		"Tags":          "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVappChildrenResourceNotFound, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	cachedvAppId := &testCachedFieldValue{}
	cachedvAppNetworkId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vapp.vapp1", "id"),
					cachedvAppId.cacheTestResourceFieldValue("vcd_vapp.vapp1", "id"),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()
					org, err := vcdClient.GetAdminOrgByName(params["Org"].(string))
					if err != nil {
						t.Errorf("error: could not find Org: %s", err)
					}
					vdc, err := org.GetVDCByName(params["Vdc"].(string), false)
					if err != nil {
						t.Errorf("error: could not find VDC: %s", err)
					}

					vapp, err := vdc.GetVAppById(cachedvAppId.fieldValue, false)
					if err != nil {
						t.Errorf("could not find vApp %s: %s", cachedvAppId.fieldValue, err)
					}

					task, err := vapp.Delete()
					if err != nil {
						t.Errorf("could not trigger vApp removal task %s: %s", cachedvAppNetworkId.fieldValue, err)
					}

					err = task.WaitTaskCompletion()
					if err != nil {
						t.Errorf("vApp removal task failed %s: %s", cachedvAppNetworkId.fieldValue, err)
					}

				},
				// Expecting to get a non-empty plan because resource was removed using SDK in
				// PreConfig
				Config:             configText,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVappChildrenResourceNotFound = `
resource "vcd_network_routed" "network_routed" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_network_routed" "network_routed_2" {
  name         = "{{.NetworkName}}-2"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "50.10.102.1"

  static_ip_pool {
    start_address = "50.10.102.2"
    end_address   = "50.10.102.254"
  }
}

resource "vcd_vapp" "vapp1" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "vappRoutedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "vapp-routed-net"
  vapp_name        = vcd_vapp.vapp1.name
  gateway          = "192.168.2.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.network_routed.name
}

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vapp1.name
  org_network_name = vcd_network_routed.network_routed_2.name
}
`

// TestAccVcdVappAccessControlResourceNotFound checks that deletion of parent vApp
// correctly handles when resource disappears (remove ID by using d.SetId("") instead of throwing
// error) outside of Terraform control. The following resources are verified here
func TestAccVcdVappAccessControlResourceNotFound(t *testing.T) {
	preTestChecks(t)
	skipTestForServiceAccountAndApiToken(t)
	// This test invokes go-vcloud-director SDK directly therefore it should not run binary tests
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"Vdc":                 testConfig.Nsxt.Vdc,
		"SharedToEveryone":    "true",
		"EveryoneAccessLevel": fmt.Sprintf(`everyone_access_level = "%s"`, types.ControlAccessReadWrite),
		"VappName":            t.Name(),
		"Tags":                "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVappAccessControlResourceNotFound, params)
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)

	cachedId := &testCachedFieldValue{}

	resourceName := "vcd_vapp_access_control.ac1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVappAccessControlDestroy(testConfig.VCD.Org, testConfig.Nsxt.Vdc, []string{params["VappName"].(string)}),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappAccessControlExists(resourceName, testConfig.VCD.Org, testConfig.Nsxt.Vdc),
					cachedId.cacheTestResourceFieldValue(resourceName, "id"),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()
					org, err := vcdClient.GetAdminOrgByName(params["Org"].(string))
					if err != nil {
						t.Errorf("error: could not find Org: %s", err)
					}
					vdc, err := org.GetVDCByName(params["Vdc"].(string), false)
					if err != nil {
						t.Errorf("error: could not find VDC: %s", err)
					}

					vapp, err := vdc.GetVAppById(cachedId.fieldValue, false)
					if err != nil {
						t.Errorf("could not find vApp %s: %s", cachedId.fieldValue, err)
					}
					task, err := vapp.Delete()
					if err != nil {
						t.Errorf("could not remove vApp '%s': %s", cachedId.fieldValue, err)
					}

					err = task.WaitTaskCompletion()
					if err != nil {
						t.Errorf("vApp removal task is failing '%s': %s", cachedId.fieldValue, err)
					}
				},
				// Expecting to get a non-empty plan because resource was removed using SDK in
				// PreConfig
				Config:             configText,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccVappAccessControlResourceNotFound = `
resource "vcd_vapp" "test1" {
  name = "{{.VappName}}"
}

resource "vcd_vapp_access_control" "ac1" {
  org      = "{{.Org}}"
  vdc      = "{{.Vdc}}"
  vapp_id  = vcd_vapp.test1.id

  shared_with_everyone    = {{.SharedToEveryone}}
  {{.EveryoneAccessLevel}}
}
`
