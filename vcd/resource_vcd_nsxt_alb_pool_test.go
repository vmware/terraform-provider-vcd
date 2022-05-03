//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbPool(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtAlbConfiguration(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"PoolName":           t.Name(),
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":   "DEDICATED",
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"IsActive":           "true",
		"AliasPrivate":       t.Name() + "-cert",
		"Certificate1Path":   testConfig.Certificates.Certificate1Path,
		"Certificate2Path":   testConfig.Certificates.Certificate2Path,
		"CertPrivateKey1":    testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":    testConfig.Certificates.Certificate1Pass,
		"Tags":               "nsxt alb",
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbPoolStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbPoolStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtAlbPoolStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtAlbPoolStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdNsxtAlbPoolStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6 := templateFill(testAccVcdNsxtAlbPoolStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["FuncName"] = t.Name() + "step7"
	configText7 := templateFill(testAccVcdNsxtAlbPoolStep7, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	params["FuncName"] = t.Name() + "step8"
	configText8 := templateFill(testAccVcdNsxtAlbPoolStep8, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	params["FuncName"] = t.Name() + "step9"
	configText9 := templateFill(testAccVcdNsxtAlbPoolStep9, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 9: %s", configText9)

	params["FuncName"] = t.Name() + "step10"
	configText10 := templateFill(testAccVcdNsxtAlbPoolStep10, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 10: %s", configText10)

	params["FuncName"] = t.Name() + "step11"
	configText11 := templateFill(testAccVcdNsxtAlbPoolStep11, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 11: %s", configText11)

	params["FuncName"] = t.Name() + "step12"
	configText12 := templateFill(testAccVcdNsxtAlbPoolStep12, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 12: %s", configText12)

	params["FuncName"] = t.Name() + "step13"
	configText13 := templateFill(testAccVcdNsxtAlbPoolStep13, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 13: %s", configText13)

	params["FuncName"] = t.Name() + "step15"
	configText15 := templateFill(testAccVcdNsxtAlbPoolStep15, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 15: %s", configText15)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbPoolDestroy("vcd_nsxt_alb_pool.test"),
		),

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_CONNECTIONS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "persistence_profile.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText2, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "description", "description text"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "FEWEST_SERVERS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "8443"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is disabled or all of the members are disabled."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "persistence_profile.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "HTTP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "TCP",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText4, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_LOAD"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "9000"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "8"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "4"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "persistence_profile.*", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "false",
						"ip_address":              "192.168.1.1",
						"health_status":           "DISABLED",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "true",
						"ip_address":              "192.168.1.2",
						"health_status":           "DOWN",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "false",
						"ip_address":              "192.168.1.3",
						"port":                    "8320",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "true",
						"ip_address":              "192.168.1.4",
						"port":                    "9200",
						"health_status":           "DOWN",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "false",
						"ip_address":              "192.168.1.5",
						"ratio":                   "3",
						"health_status":           "DISABLED",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "true",
						"ip_address":              "192.168.1.6",
						"ratio":                   "1",
						"health_status":           "DOWN",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "false",
						"ip_address":              "192.168.1.7",
						"ratio":                   "3",
						"port":                    "7000",
						"health_status":           "DISABLED",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":                 "true",
						"ip_address":              "192.168.1.8",
						"ratio":                   "1",
						"port":                    "6000",
						"health_status":           "DOWN",
						"detailed_health_message": "Pool not assigned to any Virtual Service",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText6, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "CORE_AFFINITY"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "80"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "1"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_pool.test", "persistence_profile.0.name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.type", "CLIENT_IP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.value", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "HTTP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "UDP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "PING",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText8, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText9,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_CONNECTIONS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "80"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "1"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_pool.test", "persistence_profile.0.name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.type", "CUSTOM_HTTP_HEADER"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.value", "X-HEADER-PERSISTENCE"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText10, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText11,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_CONNECTIONS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "80"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "true"),
				),
			},
			resource.TestStep{
				Config: configText12, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
			resource.TestStep{
				Config: configText13,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_CONNECTIONS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "80"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "domain_names.*", "domain1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "domain_names.*", "domain2"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_alb_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, params["PoolName"].(string)),
			},
			resource.TestStep{
				Config: configText15, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbPoolDS = `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

data "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
  name            = vcd_nsxt_alb_pool.test.name
}
`

const testAccVcdNsxtAlbPoolStep1 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}
`

const testAccVcdNsxtAlbPoolStep2 = testAccVcdNsxtAlbPoolStep1 + testAccVcdNsxtAlbPoolDS

const testAccVcdNsxtAlbPoolStep3 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  enabled         = false
  description     = "description text"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  algorithm                  = "FEWEST_SERVERS"
  default_port               = "8443"
  graceful_timeout_period    = "2"
  passive_monitoring_enabled = false

  health_monitor {
    type = "HTTP"
  }

  health_monitor {
    type = "TCP"
  }
}
`

const testAccVcdNsxtAlbPoolStep4 = testAccVcdNsxtAlbPoolStep3 + testAccVcdNsxtAlbPoolDS

// testAccVcdNsxtAlbPoolStep5 aims to test out all possible member configurations while also updating some other values
const testAccVcdNsxtAlbPoolStep5 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  algorithm               = "LEAST_LOAD"
  default_port            = "9000"
  graceful_timeout_period = "-1"

  member {
    enabled    = false
    ip_address = "192.168.1.1"
  }

  member {
    enabled    = true
    ip_address = "192.168.1.2"
  }

  member {
    enabled    = false
    ip_address = "192.168.1.3"
    port       = 8320
  }

  member {
    enabled    = true
    ip_address = "192.168.1.4"
    port       = 9200
  }

  member {
    enabled    = false
    ip_address = "192.168.1.5"
    ratio      = 3
  }

  member {
    ip_address = "192.168.1.6"
    ratio      = 1
  }

  member {
    enabled    = false
    ip_address = "192.168.1.7"
    ratio      = 3
    port       = 7000
  }

  member {
    ip_address = "192.168.1.8"
    ratio      = 1
    port       = 6000
  }
}
`

const testAccVcdNsxtAlbPoolStep6 = testAccVcdNsxtAlbPoolStep5 + testAccVcdNsxtAlbPoolDS

// testAccVcdNsxtAlbPoolStep7 tests out many combinations of health_monitors
const testAccVcdNsxtAlbPoolStep7 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  algorithm               = "CORE_AFFINITY"
  graceful_timeout_period = 0

  persistence_profile {
    type = "CLIENT_IP"
  }

  health_monitor {
    type = "HTTP"
  }

  health_monitor {
    type = "TCP"
  }

  health_monitor {
    type = "UDP"
  }

  health_monitor {
    type = "PING"
  }
}
`

const testAccVcdNsxtAlbPoolStep8 = testAccVcdNsxtAlbPoolStep7 + testAccVcdNsxtAlbPoolDS

const testAccVcdNsxtAlbPoolStep9 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  persistence_profile {
    type  = "CUSTOM_HTTP_HEADER"
    value = "X-HEADER-PERSISTENCE"
  }
}
`

const testAccVcdNsxtAlbPoolStep10 = testAccVcdNsxtAlbPoolStep9 + testAccVcdNsxtAlbPoolDS

// testAccVcdNsxtAlbPoolStep11 creates certificates and validates ALB pools can consume them
const testAccVcdNsxtAlbPoolStep11 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_library_certificate" "org-cert-1" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-1"
  certificate            = file("{{.Certificate1Path}}")
  private_key            = file("{{.CertPrivateKey1}}")
  private_key_passphrase = "{{.CertPassPhrase1}}"
}

resource "vcd_library_certificate" "org-cert-2" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-2"
  certificate            = file("{{.Certificate2Path}}")
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  ca_certificate_ids = [vcd_library_certificate.org-cert-1.id, vcd_library_certificate.org-cert-2.id]
  cn_check_enabled   = true
}
`

const testAccVcdNsxtAlbPoolStep12 = testAccVcdNsxtAlbPoolStep11 + testAccVcdNsxtAlbPoolDS

// testAccVcdNsxtAlbPoolStep13 specifies domain names for common name check
const testAccVcdNsxtAlbPoolStep13 = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_library_certificate" "org-cert-1" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-1"
  certificate            = file("{{.Certificate1Path}}")
  private_key            = file("{{.CertPrivateKey1}}")
  private_key_passphrase = "{{.CertPassPhrase1}}"
}

resource "vcd_library_certificate" "org-cert-2" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-2"
  certificate            = file("{{.Certificate2Path}}")
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  ca_certificate_ids = [vcd_library_certificate.org-cert-1.id, vcd_library_certificate.org-cert-2.id]
  cn_check_enabled   = true
  domain_names       = ["domain1", "domain2"]
}
`

const testAccVcdNsxtAlbPoolStep15 = testAccVcdNsxtAlbPoolStep13 + testAccVcdNsxtAlbPoolDS

const testAccVcdNsxtAlbPoolPrereqs = `
data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = {{.IsActive}}

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}

# Local variable is used to avoid direct reference and cover Terraform core bug https://github.com/hashicorp/terraform/issues/29484
# Even changing NSX-T ALB Controller name in UI, plan will cause to recreate all resources depending 
# on vcd_nsxt_alb_importable_cloud data source if this indirect reference (via local) variable is not used.
locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "first-se"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "{{.ReservationModel}}"
}
`

func testAccCheckVcdAlbPoolDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		client := testAccProvider.Meta().(*VCDClient)
		albPool, err := client.GetAlbPoolById(rs.Primary.ID)

		if !govcd.IsNotFound(err) && albPool != nil {
			return fmt.Errorf("ALB Pool (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

func TestAccVcdNsxtAlbPoolOrgUser(t *testing.T) {
	preTestChecks(t)

	// This test cannot run in Short mode because it uses go-vcloud-director SDK to setup prerequisites
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createSystemTemporaryVCDConnection()

	skipNoNsxtAlbConfiguration(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"PoolName":           t.Name(),
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":   "DEDICATED",
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"IsActive":           "true",
		"AliasPrivate":       t.Name() + "-cert",
		"Certificate1Path":   testConfig.Certificates.Certificate1Path,
		"Certificate2Path":   testConfig.Certificates.Certificate2Path,
		"CertPrivateKey1":    testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":    testConfig.Certificates.Certificate1Pass,
		"Tags":               "nsxt alb",
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbPoolStep1OrgUser, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbPoolStep2OrgUser, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtAlbPoolStep2OrgUserAndDs, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	// Setup prerequisites using temporary admin version and defer cleanup
	systemPrerequisites := &albOrgUserPrerequisites{t: t, vcdClient: vcdClient}
	configurePrerequisites := func() {
		fmt.Println("## Setting up prerequisites using System user")
		systemPrerequisites.setupAlbPoolPrerequisites()
		fmt.Println("## Running Terraform test")
	}

	defer func() {
		fmt.Println("## Cleaning up prerequisites")
		systemPrerequisites.teardownAlbPoolPrerequisites()
		fmt.Println("## Finished cleaning up prerequisites")
	}()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbPoolDestroy("vcd_nsxt_alb_pool.test"),
		),

		Steps: []resource.TestStep{
			resource.TestStep{
				PreConfig: configurePrerequisites, // Use temporary System session and setup all prerequisites using SDK
				Config:    configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "LEAST_CONNECTIONS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_message", "The pool is unassigned."),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", "0"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "persistence_profile.*", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "ca_certificate_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "cn_check_enabled", "false"),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "false",
						"ip_address":    "192.168.1.1",
						"health_status": "DISABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "true",
						"ip_address":    "192.168.1.2",
						"health_status": "DOWN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":    "false",
						"ip_address": "192.168.1.3",
						"port":       "8320",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "true",
						"ip_address":    "192.168.1.4",
						"port":          "9200",
						"health_status": "DOWN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "false",
						"ip_address":    "192.168.1.5",
						"ratio":         "3",
						"health_status": "DISABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "true",
						"ip_address":    "192.168.1.6",
						"ratio":         "1",
						"health_status": "DOWN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "false",
						"ip_address":    "192.168.1.7",
						"ratio":         "3",
						"port":          "7000",
						"health_status": "DISABLED",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "member.*", map[string]string{
						"enabled":       "true",
						"ip_address":    "192.168.1.8",
						"ratio":         "1",
						"port":          "6000",
						"health_status": "DOWN",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "health_monitor.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "HTTP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_pool.test", "health_monitor.*", map[string]string{
						"type": "TCP",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "algorithm", "FEWEST_SERVERS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "default_port", "8443"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "graceful_timeout_period", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "member_count", "8"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "up_member_count", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "enabled_member_count", "4"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "passive_monitoring_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.#", "1"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_pool.test", "persistence_profile.0.name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.type", "CLIENT_IP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "persistence_profile.0.value", ""),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_alb_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, params["PoolName"].(string)),
			},
			//
			resource.TestStep{
				Config: configText4, // Test data source
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_pool.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerPool:`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbPoolStep1OrgUser = `
data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}
`

const testAccVcdNsxtAlbPoolStep2OrgUser = `
data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.PoolName}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  algorithm                  = "FEWEST_SERVERS"
  default_port               = "8443"
  graceful_timeout_period    = "2"
  passive_monitoring_enabled = false

  persistence_profile {
    type = "CLIENT_IP"
  }

  health_monitor {
    type = "HTTP"
  }

  health_monitor {
    type = "TCP"
  }

  member {
    enabled    = false
    ip_address = "192.168.1.1"
  }

  member {
    enabled    = true
    ip_address = "192.168.1.2"
  }

  member {
    enabled    = false
    ip_address = "192.168.1.3"
    port       = 8320
  }

  member {
    enabled    = true
    ip_address = "192.168.1.4"
    port       = 9200
  }

  member {
    enabled    = false
    ip_address = "192.168.1.5"
    ratio      = 3
  }

  member {
    ip_address = "192.168.1.6"
    ratio      = 1
  }

  member {
    enabled    = false
    ip_address = "192.168.1.7"
    ratio      = 3
    port       = 7000
  }

  member {
    ip_address = "192.168.1.8"
    ratio      = 1
    port       = 6000
  }
}
`

const testAccVcdNsxtAlbPoolStep2OrgUserAndDs = testAccVcdNsxtAlbPoolStep2OrgUser + `
data "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.PoolName}}"
}
`
