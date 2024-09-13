//go:build nsxt || alb || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbVirtualServicePolicies(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 38.0") {
		t.Skipf("This test tests VCD 10.5.0+ (API V38.0+) features - ALB VS HTTP Policies. Skipping.")
	}

	skipNoNsxtAlbConfiguration(t)

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
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceHttpRulesStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVirtualServiceHttpRulesStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(

					// Request rule
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_alb_virtual_service.test", "id", "vcd_nsxt_alb_virtual_service_http_req_rules.test1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.#", "6"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.name", "criteria-max-rewrite"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.criteria", "IS_NOT_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.protocol_type", "HTTP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.cookie.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.cookie.0.criteria", "DOES_NOT_END_WITH"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.cookie.0.name", "does-not-name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.cookie.0.value", "does-not-value"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.http_methods.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.http_methods.0.criteria", "IS_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.http_methods.0.methods.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.path.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.path.0.criteria", "CONTAINS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.path.0.paths.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.query.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.request_headers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.match_criteria.0.service_ports.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.actions.0.modify_header.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.actions.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.0.actions.0.rewrite_url.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.1.name", "criteria-max-modify-header"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.1.active", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.1.logging", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.1.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.1.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.2.name", "criteria-max-rewrite-url"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.2.active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.2.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.2.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.2.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.3.name", "one-criteria"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.4.name", "criteria-max-min-fields-redirect"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.4.actions.0.redirect.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.4.actions.0.redirect.0.protocol", "302"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.5.name", "criteria-max-min-fields-rewrite"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.5.actions.0.rewrite_url.0.host_header", "asd"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "rule.5.actions.0.rewrite_url.0.existing_path", "/existing"),

					// Response rule
					resource.TestCheckResourceAttrPair("vcd_nsxt_alb_virtual_service.test", "id", "vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.#", "5"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.name", "criteria-max-rewrite"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.criteria", "IS_NOT_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.protocol_type", "HTTP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.cookie.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.cookie.0.criteria", "DOES_NOT_END_WITH"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.cookie.0.name", "does-not-name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.cookie.0.value", "does-not-value"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.http_methods.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.http_methods.0.criteria", "IS_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.http_methods.0.methods.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.path.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.path.0.criteria", "CONTAINS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.path.0.paths.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.query.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.request_headers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.service_ports.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.location_header.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.location_header.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.response_headers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.match_criteria.0.status_code.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.0.actions.0.rewrite_location_header.#", "1"),

					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.1.name", "criteria-max-modify-header"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.1.active", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.1.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.1.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.1.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.2.name", "one-criteria"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.3.name", "criteria-min-rewrite"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.3.actions.0.rewrite_location_header.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.3.actions.0.rewrite_location_header.0.port", "443"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.4.name", "criteria-min-modify-header"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "rule.4.actions.0.modify_header.#", "1"),

					// Security rule
					resource.TestCheckResourceAttrPair("vcd_nsxt_alb_virtual_service.test", "id", "vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.#", "12"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.name", "max-sec-redirect-to-https"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.logging", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.actions.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.criteria", "IS_NOT_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.client_ip_address.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.protocol_type", "HTTP"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.cookie.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.cookie.0.criteria", "DOES_NOT_END_WITH"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.cookie.0.name", "does-not-name"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.cookie.0.value", "does-not-value"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.http_methods.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.http_methods.0.criteria", "IS_IN"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.http_methods.0.methods.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.path.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.path.0.criteria", "CONTAINS"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.path.0.paths.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.query.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.request_headers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.0.match_criteria.0.service_ports.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.1.name", "max-sec-connection-allow"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.1.logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.2.name", "max-sec-connection-close"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.3.name", "max-sec-response"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.4.name", "max-sec-rate-limit-report-only"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.5.name", "max-sec-rate-limit-close-connection"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.6.name", "max-sec-rate-limit-redirect"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.7.name", "max-sec-rate-limit-local-resp"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.8.name", "one-criteria"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.9.name", "one-criteria-action-min1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.10.name", "one-criteria-action-min2"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "rule.11.name", "one-criteria-action-min3"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_alb_virtual_service_http_req_rules.test1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, params["VirtualServiceName"].(string)),
			},
			{
				ResourceName:      "vcd_nsxt_alb_virtual_service_http_resp_rules.test1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, params["VirtualServiceName"].(string)),
			},
			{
				ResourceName:      "vcd_nsxt_alb_virtual_service_http_sec_rules.test1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, params["VirtualServiceName"].(string)),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_alb_virtual_service_http_req_rules.test1", "data.vcd_nsxt_alb_virtual_service_http_req_rules.test1", nil),
					resourceFieldsEqual("vcd_nsxt_alb_virtual_service_http_resp_rules.test1", "data.vcd_nsxt_alb_virtual_service_http_resp_rules.test1", nil),
					resourceFieldsEqual("vcd_nsxt_alb_virtual_service_http_sec_rules.test1", "data.vcd_nsxt_alb_virtual_service_http_sec_rules.test1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbVirtualServiceHttpRulesStep1 = testAccVcdNsxtAlbVirtualServicePrereqs + `
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

resource "vcd_nsxt_alb_virtual_service_http_req_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id

  rule {
    name   = "criteria-max-rewrite"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      rewrite_url {
        host_header   = "X-HOST-HEADER"
        existing_path = "/123"
        keep_query    = true
        query         = "rewrite"
      }
    }
  }

  rule {
    name    = "criteria-max-modify-header"
    active  = false
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }

      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      modify_header {
        action = "REMOVE"
        name   = "X-REMOVE-HEADER"
      }
      modify_header {
        action = "ADD"
        name   = "X-ADDED-HEADER"
        value  = "value"
      }

      modify_header {
        action = "REPLACE"
        name   = "X-EXISTING-HEADER"
        value  = "new-value"
      }

    }
  }

  rule {
    name   = "criteria-max-rewrite-url"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rewrite_url {
        host_header   = "host-header"
        existing_path = "/123"
        keep_query    = true
        query         = "new"
      }
    }
  }

  rule {
    name   = "one-criteria"
    active = true
    match_criteria {
      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }
    }

    actions {
      rewrite_url {
        host_header   = "X-HOST-HEADER"
        existing_path = "/123"
        keep_query    = true
        query         = "rewrite"
      }
    }
  }

  rule {
    name   = "criteria-max-min-fields-redirect"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      redirect {
        protocol    = "HTTP"
        status_code = 302
      }
    }
  }

  rule {
    name   = "criteria-max-min-fields-rewrite"
    active = true
    match_criteria {
      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"
    }

    actions {
      rewrite_url {
        host_header   = "asd"
        existing_path = "/existing"
      }
    }
  }
}


resource "vcd_nsxt_alb_virtual_service_http_resp_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id

  rule {
    name   = "criteria-max-rewrite"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

      location_header {
        criteria = "DOES_NOT_EQUAL"
        values   = ["one", "two"]
      }

      response_headers {
        criteria = "CONTAINS"
        name     = "X-CONTAINS-RESP-HEADER"
        values   = ["vone", "vtwo"]
      }

      response_headers {
        criteria = "DOES_NOT_END_WITH"
        name     = "-END"
        values   = ["asd", "bsd"]
      }

      status_code {
        criteria         = "IS_NOT_IN"
        http_status_code = "200"
      }
    }

    actions {
      rewrite_location_header {
        protocol   = "HTTP"
        port       = 443
        host       = "another-host"
        path       = "/"
        keep_query = true
      }
    }
  }

  rule {
    name   = "criteria-max-modify-header"
    active = false
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

      location_header {
        criteria = "DOES_NOT_EQUAL"
        values   = ["one", "two"]
      }

      response_headers {
        criteria = "CONTAINS"
        name     = "X-CONTAINS-RESP-HEADER"
        values   = ["vone", "vtwo"]
      }

      response_headers {
        criteria = "DOES_NOT_END_WITH"
        name     = "-END"
        values   = ["asd", "bsd"]
      }

      status_code {
        criteria         = "IS_NOT_IN"
        http_status_code = "200"
      }

    }

    actions {
      modify_header {
        action = "REMOVE"
        name   = "X-REMOVE-HEADER"
      }
      modify_header {
        action = "ADD"
        name   = "X-ADDED-HEADER"
        value  = "value"
      }

      modify_header {
        action = "REPLACE"
        name   = "X-EXISTING-HEADER"
        value  = "new-value"
      }
    }
  }

  rule {
    name   = "one-criteria"
    active = true
    match_criteria {
      protocol_type = "HTTP"
    }

    actions {
      rewrite_location_header {
        protocol   = "HTTP"
        port       = 443
        host       = "another-host"
        path       = "/"
        keep_query = true
      }
    }
  }

  rule {
    name   = "criteria-min-rewrite"
    active = true
    match_criteria {

      status_code {
        criteria         = "IS_NOT_IN"
        http_status_code = "200"
      }
    }

    actions {
      rewrite_location_header {
        protocol = "HTTP"
        port     = 443
      }
    }
  }

  rule {
    name   = "criteria-min-modify-header"
    active = true
    match_criteria {

      status_code {
        criteria         = "IS_NOT_IN"
        http_status_code = "200"
      }
    }

    actions {
      modify_header {
        action = "REMOVE"
        name   = "X-HEADER-ONE"
      }
    }
  }
}

resource "vcd_nsxt_alb_virtual_service_http_sec_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id

  rule {
    name    = "max-sec-redirect-to-https"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      redirect_to_https = "80"
    }
  }

  rule {
    name    = "max-sec-connection-allow"
    active  = true
    logging = false
    match_criteria {
      client_ip_address {
        criteria     = "IS_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      connections = "ALLOW"
    }
  }

  rule {
    name    = "max-sec-connection-close"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      connections = "CLOSE"
    }
  }

  rule {
    name    = "max-sec-response"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      send_response {
        content      = base64encode("PERMISSION DENIED")
        content_type = "text/plain"
        status_code  = "403"
      }
    }
  }

  rule {
    name    = "max-sec-rate-limit-report-only"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
      }
    }
  }

  rule {
    name    = "max-sec-rate-limit-close-connection"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rate_limit {
        count                   = "10000"
        period                  = "2000"
        action_close_connection = true
      }
    }
  }

  rule {
    name    = "max-sec-rate-limit-redirect"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_redirect {
          protocol    = "HTTPS"
          port        = 80
          status_code = 302
          host        = "other-host"
          path        = "/"
          keep_query  = true
        }
      }
    }
  }

  rule {
    name    = "max-sec-rate-limit-local-resp"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_local_response {
          content      = base64encode("PERMISSION DENIED")
          content_type = "text/plain"
          status_code  = "403"
        }
      }
    }
  }

  rule {
    name    = "one-criteria"
    active  = true
    logging = true
    match_criteria {
      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      redirect_to_https = "80"
    }
  }

  rule {
    name    = "one-criteria-action-min1"
    active  = true
    logging = true
    match_criteria {
      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_redirect {
          protocol    = "HTTPS"
          port        = 443
          status_code = 301
        }
      }
    }
  }

  rule {
    name    = "one-criteria-action-min2"
    active  = true
    logging = true
    match_criteria {
      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_local_response {
          status_code  = "403"
        }
      }
    }
  }

  rule {
    name    = "one-criteria-action-min3"
    active  = true
    logging = true
    match_criteria {
      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      send_response {
        status_code = 403
      }
    }
  }
}
`

const testAccVcdNsxtAlbVirtualServiceHttpRulesStep2DS = testAccVcdNsxtAlbVirtualServiceHttpRulesStep1 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_alb_virtual_service_http_req_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id
}

data "vcd_nsxt_alb_virtual_service_http_resp_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id
}

data "vcd_nsxt_alb_virtual_service_http_sec_rules" "test1" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id
}
`
