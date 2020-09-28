// +build gateway ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestAccVcdEdgeGatewaySettingsFull(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("Edge Gateway resource tests require system admin privileges")
		return
	}

	fmt.Println("*** This test doesn't run anything directly, but it creates an HCL script")
	fmt.Println("*** (vcd.TestAccVcdEdgeGatewaySettingsFull.tf) that will then run with the binary tests.")

	// The test uses two providers:
	// * one (system administrator) to create external network and edge gateway,
	// * the other (org user) to create the edge gateway settings resource.
	//
	// How to run in isolation:
	//
	//  1. make install && cd vcd
	//  2. go test -tags gateway -v -timeout 0 -run TestAccVcdEdgeGatewaySettingsFull -vcd-add-provider
	//  3. cd test-artifacts
	//  4. ./test-binary.sh clear pause names vcd.TestAccVcdEdgeGatewaySettingsFull.tf
	//  5. -- Exit the execution after apply, and check that lb_enabled is false in edge gateway and true in edge gateway settings
	//  6. cd tmp
	//  7. terraform refresh
	//  8. -- Now the values of edge gateway and edge gateway settings are the same.
	//  9. -- Optionally check the edge gateway in the UI
	// 10. terraform destroy
	// -----------------------------------------------------------------------------------

	if testConfig.TestEnvBuild.OrgUser == "" ||
		testConfig.TestEnvBuild.OrgUserPassword == "" ||
		testConfig.Networking.Vcenter == "" ||
		testConfig.Networking.ExternalNetworkPortGroup == "" ||
		testConfig.Networking.ExternalNetworkPortGroupType == "" ||
		testConfig.VCD.Org == "" ||
		testConfig.VCD.Vdc == "" {
		t.Skip("one or more elements needed for TestAccVcdEdgeGatewaySettingsFull are missing from the configuration file")
	}

	testName := "EdgeGatewaySettingsFull"
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"OrgUser":            testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":    testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":             testConfig.Provider.Url,
		"EdgeGateway":        testName + "-egw",
		"EgwSettings":        testName + "-egw-settings",
		"NewExternalNetwork": testName + "-extnet",
		"Version":            currentProviderVersion,
		"Vcenter":            testConfig.Networking.Vcenter,
		"Type":               testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":          testConfig.Networking.ExternalNetworkPortGroup,
		"Tags":               "gateway",
	}
	configText := templateFill(testAccEdgeGatewaySettingsFull, params)
	debugPrintf("#[DEBUG] %s", configText)
}

func getEdgeGatewayInfo() (*govcd.EdgeGateway, error) {

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}
	vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		return nil, fmt.Errorf("vdc not found : %s", err)
	}
	egw, err := vdc.GetEdgeGatewayByName(testConfig.Networking.EdgeGateway, false)
	if err != nil {
		return nil, fmt.Errorf("edge gateway not found : %s", err)
	}
	return egw, nil
}

func TestAccVcdEdgeGatewaySettingsBasic(t *testing.T) {

	testName := "EdgeGatewaySettingsBasic"
	var existingEgw *govcd.EdgeGateway
	var fwSettings *types.FirewallConfigWithXml
	var lbSettings *types.LbGeneralParamsWithXml
	var err error
	if !vcdShortTest {
		// Gets current settings from the edge gateway
		existingEgw, err = getEdgeGatewayInfo()
		if err != nil {
			t.Errorf("error retrieving edge gateway: %s", err)
		}
		fwSettings, err = existingEgw.GetFirewallConfig()
		if err != nil {
			t.Errorf("error retrieving edge gateway firewall parameters: %s", err)
		}
		lbSettings, err = existingEgw.GetLBGeneralParams()
		if err != nil {
			t.Errorf("error retrieving edge gateway load balancing parameters: %s", err)
		}
		// Restore original values in edge gateway after the test
		defer func() {
			_, err = existingEgw.UpdateLBGeneralParams(lbSettings.Enabled, lbSettings.AccelerationEnabled, lbSettings.Logging.Enable, lbSettings.Logging.LogLevel)
			if err != nil {
				t.Logf("WARNING: restore of LB settings failed: %s\n", err)
			}
			_, err = existingEgw.UpdateFirewallConfig(fwSettings.Enabled, fwSettings.DefaultPolicy.LoggingEnabled, fwSettings.DefaultPolicy.Action)
			if err != nil {
				t.Logf("WARNING: restore of firewall settings failed: %s\n", err)
			}
		}()
	}

	enableLbLogging := "true"
	lbLogLevel := "debug"
	if !usingSysAdmin() {
		// There is a bug in vCD: when using an organization user, load balancer logging settings are ignored.
		// This happens also in the UI, if you *log in* as a tenant (not "open in tenant portal")
		lbLogLevel = "info"
		enableLbLogging = "false"
		if lbSettings != nil {
			enableLbLogging = fmt.Sprintf("%v", lbSettings.Logging.Enable)
			lbLogLevel = lbSettings.Logging.LogLevel
		}
	}

	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           testConfig.Networking.EdgeGateway,
		"LbEnabled":             true,
		"LbAccelerationEnabled": true,
		"LbLoggingEnabled":      enableLbLogging,
		"LbLogLevel":            lbLogLevel,
		"FwEnabled":             true,
		"FwRuleEnabled":         true,
		"FwRuleAction":          "accept",
		"EgwSettings":           testName,

		"Tags": "gateway",
	}
	configText := templateFill(testAccEdgeGatewaySettingsSimple, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	egwSettingsResource := "vcd_edgegateway_settings." + testName
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	// Note: this test can't run in parallel, as it updates the main edge gateway in the vCD
	// and it could interfere with other tests
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		//CheckDestroy: func(s *terraform.State) error {return nil},
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// Check that the resource has the expected values
					resource.TestCheckResourceAttr(egwSettingsResource, "lb_enabled", "true"),
					resource.TestCheckResourceAttr(egwSettingsResource, "lb_acceleration_enabled", "true"),
					resource.TestCheckResourceAttr(egwSettingsResource, "lb_logging_enabled", enableLbLogging),
					resource.TestCheckResourceAttr(egwSettingsResource, "lb_loglevel", lbLogLevel),
					resource.TestCheckResourceAttr(egwSettingsResource, "fw_enabled", "true"),
					resource.TestCheckResourceAttr(egwSettingsResource, "fw_default_rule_logging_enabled", "true"),
					resource.TestCheckResourceAttr(egwSettingsResource, "fw_default_rule_action", "accept"),

					// Check that the edge gateway has the expected values
					checkEdgeGatewaySettingsCorrespondence("lb_enabled", "true"),
					checkEdgeGatewaySettingsCorrespondence("lb_acceleration_enabled", "true"),
					checkEdgeGatewaySettingsCorrespondence("lb_logging_enabled", enableLbLogging),
					checkEdgeGatewaySettingsCorrespondence("lb_loglevel", lbLogLevel),
					checkEdgeGatewaySettingsCorrespondence("fw_enabled", "true"),
					checkEdgeGatewaySettingsCorrespondence("fw_default_rule_logging_enabled", "true"),
					checkEdgeGatewaySettingsCorrespondence("fw_default_rule_action", "accept"),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_edgegateway_settings." + testConfig.Networking.EdgeGateway + "-import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgVdcObject(testConfig, testConfig.Networking.EdgeGateway),
				ImportStateVerifyIgnore: []string{"external_network", "external_networks"},
			},
		},
	})
}

// boolComparisonToErr returns an error if the two provided values don't match
// 'wanted' is a string (as returned from terraform state) and 'got' is the actual value
func boolComparisonToErr(field string, wanted string, got bool) error {
	if fmt.Sprintf("%v", got) == wanted {
		return nil
	}
	return fmt.Errorf("[%s] - wanted %s but got %v", field, wanted, got)
}

// strComparisonToErr returns an error if the two provided values don't match
func strComparisonToErr(field string, wanted, got string) error {
	if got == wanted {
		return nil
	}
	return fmt.Errorf("[%s] - wanted %s but got %s", field, wanted, got)
}

// checkEdgeGatewaySettingsCorrespondence checks that the given field in edge gateway has the wanted value
// This function makes sure that the value defined in vcd_edgegateway_settings are actually set in the edge gateway
func checkEdgeGatewaySettingsCorrespondence(field, value string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		var egw *govcd.EdgeGateway
		var fwSettings *types.FirewallConfigWithXml
		var lbSettings *types.LbGeneralParamsWithXml
		var err error
		egw, err = getEdgeGatewayInfo()
		if err != nil {
			return err
		}
		fwSettings, err = egw.GetFirewallConfig()
		if err != nil {
			return err
		}
		lbSettings, err = egw.GetLBGeneralParams()
		if err != nil {
			return err
		}
		switch field {
		case "lb_enabled":
			return boolComparisonToErr(field, value, lbSettings.Enabled)
		case "lb_acceleration_enabled":
			return boolComparisonToErr(field, value, lbSettings.AccelerationEnabled)
		case "lb_logging_enabled":
			return boolComparisonToErr(field, value, lbSettings.Logging.Enable)
		case "lb_loglevel":
			return strComparisonToErr(field, value, lbSettings.Logging.LogLevel)
		case "fw_enabled":
			return boolComparisonToErr(field, value, fwSettings.Enabled)
		case "fw_default_rule_logging_enabled":
			return boolComparisonToErr(field, value, fwSettings.DefaultPolicy.LoggingEnabled)
		case "fw_default_rule_action":
			return strComparisonToErr(field, value, fwSettings.DefaultPolicy.Action)
		}
		return nil
	}
}

// This test will only run with the binary tests
const testAccEdgeGatewaySettingsFull = `
provider "vcd" {
  alias                = "orguser"
  user                 = "{{.OrgUser}}"
  password             = "{{.OrgUserPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcdUrl}}"
  sysorg               = "{{.Org}}"
  org                  = "{{.Org}}"
  vdc                  = "{{.Vdc}}"
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  version              = "~> {{.Version}}"
  logging              = true
  logging_file         = "go-vcloud-director-org.log"
}

resource "vcd_external_network" "{{.NewExternalNetwork}}" {
  name        = "{{.NewExternalNetwork}}"
  description = "Test External Network"

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.PortGroup}}"
    type    = "{{.Type}}"
  }

  ip_scope {
    gateway      = "192.168.30.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.0.164"
    dns2         = "192.168.0.196"
    dns_suffix   = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }

  retain_net_info_across_deployments = "false"
}

resource "vcd_edgegateway" "egw" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.EdgeGateway}}"
  configuration = "compact"
  advanced      = true

  external_network {
    name = vcd_external_network.{{.NewExternalNetwork}}.name
    subnet {
      gateway = "192.168.30.49"
      netmask = "255.255.255.240"
    }
  }

  # The plan for vcd_edgegateway will fail, because it will have been changed by vcd_edgegateway_settings
  lifecycle {
    ignore_changes = [lb_enabled, lb_acceleration_enabled, lb_logging_enabled, lb_loglevel, fw_enabled, fw_default_rule_logging_enabled, fw_default_rule_action]
  }
}

resource "vcd_edgegateway_settings" "{{.EgwSettings}}" {
  provider                = vcd.orguser
  edge_gateway_id         = vcd_edgegateway.egw.id
  lb_enabled              = true
  lb_acceleration_enabled = true
  lb_logging_enabled      = true
  lb_loglevel             = "info"

  fw_enabled                      = true
  fw_default_rule_logging_enabled = true
  fw_default_rule_action          = "deny"

  # The plan for vcd_edgegateway_settings may fail because of logging fields not being visible to tenants
  lifecycle {
    ignore_changes = [lb_logging_enabled, lb_loglevel]
  }
}

output "egw" {
  value = vcd_edgegateway.egw
}

output "egw_settings" {
  value = vcd_edgegateway_settings.{{.EgwSettings}}
}
`

const testAccEdgeGatewaySettingsSimple = `
# skip-binary-test: would update existing Edge Gateway

data "vcd_edgegateway" "egw" {
  name = "{{.EdgeGateway}}"
}

resource "vcd_edgegateway_settings" "{{.EgwSettings}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  edge_gateway_id         = data.vcd_edgegateway.egw.id
  lb_enabled              = {{.LbEnabled}}
  lb_acceleration_enabled = {{.LbAccelerationEnabled}}
  lb_logging_enabled      = {{.LbLoggingEnabled}}       # only set if provider user is system administrator
  lb_loglevel             = "{{.LbLogLevel}}"           # only set if provider user is system administrator

  fw_enabled                      = {{.FwEnabled}}
  fw_default_rule_logging_enabled = {{.FwRuleEnabled}}
  fw_default_rule_action          = "{{.FwRuleAction}}"
}
`
