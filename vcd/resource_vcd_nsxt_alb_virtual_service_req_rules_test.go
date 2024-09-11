//go:build nsxt || alb || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbVirtualServicePolicies(t *testing.T) {
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
	configText1 := templateFill(testAccVcdNsxtAlbVirtualServiceHttpRulesStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      resource.ComposeAggregateTestCheckFunc(
		// testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
		// testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
		// testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
		// testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		// testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check:  resource.ComposeAggregateTestCheckFunc(
				// resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
				// resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "name", t.Name()),
				// resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "description", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "application_profile_type", "HTTP"),
				// resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "pool_id", regexp.MustCompile(`^urn:vcloud:`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_engine_group_id", regexp.MustCompile(`^urn:vcloud:`)),
				// resource.TestCheckResourceAttrSet("vcd_nsxt_alb_virtual_service.test", "virtual_ip_address"),
				// resource.TestCheckResourceAttr("vcd_nsxt_alb_virtual_service.test", "service_port.#", "1"),
				// resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_alb_virtual_service.test", "service_port.*", map[string]string{
				// 	"start_port": "80",
				// 	"end_port":   "81",
				// 	"type":       "TCP_PROXY",
				// }),
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
    name   = "criteria-max-modify-header"
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






}

resource "vcd_nsxt_alb_virtual_service_http_sec_rules" "asd" {
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
    logging = true
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
}
`
