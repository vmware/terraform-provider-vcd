//go:build nsxt || alb || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbVirtualService(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"TestName":           t.Name(),
		"VirtualServiceName": t.Name(),
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
		"CertPrivateKey1":    testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":    testConfig.Certificates.Certificate1Pass,
		"Tags":               "nsxt alb",
	}
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVirtualServiceStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtAlbVirtualServiceStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtAlbVirtualServiceStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdNsxtAlbVirtualServiceStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6 := templateFill(testAccVcdNsxtAlbVirtualServiceStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["FuncName"] = t.Name() + "step7"
	configText7 := templateFill(testAccVcdNsxtAlbVirtualServiceStep7, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	params["FuncName"] = t.Name() + "step8"
	configText8 := templateFill(testAccVcdNsxtAlbVirtualServiceStep8, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	params["FuncName"] = t.Name() + "step9"
	configText9 := templateFill(testAccVcdNsxtAlbVirtualServiceStep9, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 9: %s", configText9)

	params["FuncName"] = t.Name() + "step11"
	configText11 := templateFill(testAccVcdNsxtAlbVirtualServiceStep11, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 11: %s", configText11)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"end_port":   "81",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText2, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),

					// Check ALB Pool attributes that represent which Virtual Services consume this pool
					resource.TestCheckResourceAttr("vcd_nsxt_alb_pool.test", "associated_virtual_service_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_alb_pool.test", "associated_virtual_services.*", t.Name()),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText4, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "L4"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"type":       "TCP_PROXY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "84",
						"end_port":   "85",
						"type":       "TCP_PROXY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "87",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText6, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
				),
			},
			{
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTPS"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port":  "80",
						"type":        "TCP_PROXY",
						"ssl_enabled": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port":  "84",
						"end_port":    "85",
						"type":        "TCP_PROXY",
						"ssl_enabled": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "87",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText8, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
				),
			},
			{
				Config: configText9,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "L4_TLS"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port":  "80",
						"type":        "TCP_PROXY",
						"ssl_enabled": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port":  "84",
						"end_port":    "85",
						"type":        "TCP_PROXY",
						"ssl_enabled": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "87",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_virtual_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, params["VirtualServiceName"].(string)),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				Config: configText11, // Datasource check
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbVirtualServicePrereqs = testAccVcdNsxtAlbPoolPrereqs + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}
`

const testAccVcdNsxtAlbVirtualServiceDS = `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

data "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
  name            = vcd_nsxt_alb_virtual_service.test.name
}
`

const testAccVcdNsxtAlbVirtualServiceStep1 = testAccVcdNsxtAlbVirtualServicePrereqs + `
resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`
const testAccVcdNsxtAlbVirtualServiceStep2 = testAccVcdNsxtAlbVirtualServiceStep1 + testAccVcdNsxtAlbVirtualServiceDS

const testAccVcdNsxtAlbVirtualServiceStep3 = testAccVcdNsxtAlbVirtualServicePrereqs + `
resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  description     = "description"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep4 = testAccVcdNsxtAlbVirtualServiceStep3 + testAccVcdNsxtAlbVirtualServiceDS

const testAccVcdNsxtAlbVirtualServiceStep5 = testAccVcdNsxtAlbVirtualServicePrereqs + `
resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "L4"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }

  service_port {
    start_port = 84
    end_port   = 85
    type       = "TCP_PROXY"
  }

  service_port {
    start_port = 87
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep6 = testAccVcdNsxtAlbVirtualServiceStep5 + testAccVcdNsxtAlbVirtualServiceDS

const testAccVcdNsxtAlbVirtualServiceStep7 = testAccVcdNsxtAlbVirtualServicePrereqs + `
resource "vcd_library_certificate" "org-cert-1" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-1"
  certificate            = file("{{.Certificate1Path}}")
  private_key            = file("{{.CertPrivateKey1}}")
  private_key_passphrase = "{{.CertPassPhrase1}}"
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  ca_certificate_id        = vcd_library_certificate.org-cert-1.id
  application_profile_type = "HTTPS"
  service_port {
    start_port  = 80
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port  = 84
    end_port    = 85
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port = 87
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep8 = testAccVcdNsxtAlbVirtualServiceStep7 + testAccVcdNsxtAlbVirtualServiceDS

const testAccVcdNsxtAlbVirtualServiceStep9 = testAccVcdNsxtAlbVirtualServicePrereqs + `
resource "vcd_library_certificate" "org-cert-1" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}-1"
  certificate            = file("{{.Certificate1Path}}")
  private_key            = file("{{.CertPrivateKey1}}")
  private_key_passphrase = "{{.CertPassPhrase1}}"
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  ca_certificate_id        = vcd_library_certificate.org-cert-1.id
  application_profile_type = "L4_TLS"
  service_port {
    start_port  = 80
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port  = 84
    end_port    = 85
    type        = "TCP_PROXY"
    ssl_enabled = true
  }

  service_port {
    start_port = 87
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep11 = testAccVcdNsxtAlbVirtualServiceStep9 + testAccVcdNsxtAlbVirtualServiceDS

func TestAccVcdNsxtAlbVirtualServiceOrgUser(t *testing.T) {
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
		"TestName":           t.Name(),
		"VirtualServiceName": t.Name(),
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
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceStep1OrgUser, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVirtualServiceStep2OrgUser, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtAlbVirtualServiceStep2OrgUserAndDS, params)
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
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbPoolDestroy("vcd_nsxt_alb_pool.test"),
		),

		Steps: []resource.TestStep{
			{
				PreConfig: configurePrerequisites, // Use temporary System session and setup all prerequisites using SDK
				Config:    configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"end_port":   "81",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "L4"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"type":       "TCP_PROXY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "84",
						"end_port":   "85",
						"type":       "TCP_PROXY",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "87",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_virtual_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, params["VirtualServiceName"].(string)),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				Config: configText4, // Test data source
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbVirtualServiceStep1OrgUser = `
data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

# This is not really needed in this example, but checking that the 
# data source can run with Org user
data "vcd_nsxt_alb_settings" "gw" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_alb_settings.gw.id

  # This name comes from prerequisite setup
  service_engine_group_name = "{{.TestName}}"
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}-pool"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = data.vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep2OrgUser = `
data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

# This is not really needed in this example, but checking that the 
# data source can run with Org user
data "vcd_nsxt_alb_settings" "gw" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_alb_settings.gw.id

  # This name comes from prerequisite setup
  service_engine_group_name = "{{.TestName}}"
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}-pool"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = data.vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "L4"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }

  service_port {
    start_port = 84
    end_port   = 85
    type       = "TCP_PROXY"
  }

  service_port {
    start_port = 87
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceStep2OrgUserAndDS = testAccVcdNsxtAlbVirtualServiceStep2OrgUser + `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

data "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = vcd_nsxt_alb_virtual_service.test.name
}
`

func testAccCheckVcdAlbVirtualServiceDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		client := testAccProvider.Meta().(*VCDClient)
		albVirtualService, err := client.GetAlbVirtualServiceById(rs.Primary.ID)

		if !govcd.IsNotFound(err) && albVirtualService != nil {
			return fmt.Errorf("ALB Virtual Service (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

// TestAccVcdNsxtAlbVirtualServiceTransparentMode tests two explicitly 10.4.1+ features:
// * Transparent mode (uses Legacy Active-Standby Service Engine Group in AVI as it is the only way
// to support Transparent mode)
// * ALB Pools with member groups (IP Set)
func TestAccVcdNsxtAlbVirtualServiceTransparentMode(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"IpSetName":          t.Name(),
		"VirtualServiceName": t.Name(),
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		// A Service Engine Group in Legacy Active Standby mode must be used
		// so that Transparent mode can be tested
		"ServiceEngineGroupName": testConfig.Nsxt.NsxtAlbServiceEngineGroup,
		"ReservationModel":       "DEDICATED",
		"Org":                    testConfig.VCD.Org,
		"NsxtVdc":                testConfig.Nsxt.Vdc,
		"EdgeGw":                 testConfig.Nsxt.EdgeGateway,
		"IsActive":               "true",
		"AliasPrivate":           t.Name() + "-cert",
		"Certificate1Path":       testConfig.Certificates.Certificate1Path,
		"CertPrivateKey1":        testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":        testConfig.Certificates.Certificate1Pass,
		"Tags":                   "nsxt alb",
	}
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceTransparentMode1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVirtualServiceTransparentMode1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"end_port":   "81",
						"type":       "TCP_PROXY",
					}),
				),
			},
			{
				Config: configText2, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_alb_virtual_service.test", "vcd_nsxt_alb_virtual_service.test", nil),
					resourceFieldsEqual("data.vcd_nsxt_alb_pool.test", "vcd_nsxt_alb_pool.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbVirtualServiceTransparentMode1 = `
# Provider configuration

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  {{.LicenseType}}
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
  importable_service_engine_group_name = "{{.ServiceEngineGroupName}}"
  reservation_model                    = "{{.ReservationModel}}"
  {{.SupportedFeatureSet}}
}

# Tenant Configuration

data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id             = data.vcd_nsxt_edgegateway.existing.id
  is_active                   = {{.IsActive}}
  is_transparent_mode_enabled = true
  {{.SupportedFeatureSet}}

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

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}


resource "vcd_nsxt_ip_set" "pool-members" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  name        = "{{.IpSetName}}"
  description = "IP Set for NSX-T ALB Pool with Transparent Virtual Service"

  ip_addresses = [
    "10.10.10.0/24",
  ]
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
  member_group_id = vcd_nsxt_ip_set.pool-members.id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name                        = "{{.VirtualServiceName}}"
  edge_gateway_id             = vcd_nsxt_alb_settings.test.edge_gateway_id
  is_transparent_mode_enabled = true

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVirtualServiceTransparentMode1DS = testAccVcdNsxtAlbVirtualServiceTransparentMode1 + `
# skip-binary-test: data source test
data "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id 
  name = vcd_nsxt_alb_virtual_service.test.name
}

data "vcd_nsxt_alb_pool" "test" {
  org             = "{{.Org}}"
  vdc             = "{{.NsxtVdc}}"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
  name            = vcd_nsxt_alb_pool.test.name
}
`

// TestAccVcdNsxtAlbVirtualServiceIPv6 tests that IPv6 Virtual Service IP can be set on VCD 10.4.0+
func TestAccVcdNsxtAlbVirtualServiceIPv6(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		t.Skipf("This test tests VCD 10.4.0+ (API V37.0+) features. Skipping.")
	}

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"IpSetName":          t.Name(),
		"VirtualServiceName": t.Name(),
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		// A Service Engine Group in Legacy Active Standby mode must be used
		// so that Transparent mode can be tested
		"ServiceEngineGroupName": testConfig.Nsxt.NsxtAlbServiceEngineGroup,
		"ReservationModel":       "DEDICATED",
		"Org":                    testConfig.VCD.Org,
		"NsxtVdc":                testConfig.Nsxt.Vdc,
		"EdgeGw":                 testConfig.Nsxt.EdgeGateway,
		"IsActive":               "true",
		"AliasPrivate":           t.Name() + "-cert",
		"Certificate1Path":       testConfig.Certificates.Certificate1Path,
		"CertPrivateKey1":        testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":        testConfig.Certificates.Certificate1Pass,
		"Tags":                   "nsxt alb",
	}
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceIpv6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "ipv6_virtual_ip_address", "2002:0:0:1234:abcd:ffff:c0a8:103"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
						"start_port": "80",
						"end_port":   "81",
						"type":       "TCP_PROXY",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbVirtualServiceIpv6 = `
# Provider configuration

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  {{.LicenseType}}
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
  importable_service_engine_group_name = "{{.ServiceEngineGroupName}}"
  reservation_model                    = "{{.ReservationModel}}"
  {{.SupportedFeatureSet}}
}

# Tenant Configuration

data "vcd_nsxt_edgegateway" "existing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  mode = "SLAAC"
}

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id                    = data.vcd_nsxt_edgegateway.existing.id
  is_active                          = {{.IsActive}}
  service_network_specification      = "10.10.255.225/27"
  ipv6_service_network_specification = "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"

  {{.SupportedFeatureSet}}

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

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name            = "{{.VirtualServiceName}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
  member {
    enabled    = true
    ip_address = "192.168.1.2"
  }
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name                        = "{{.VirtualServiceName}}"
  edge_gateway_id             = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  ipv6_virtual_ip_address  = "2002:0:0:1234:abcd:ffff:c0a8:103"
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`
