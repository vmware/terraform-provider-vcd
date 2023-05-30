//go:build network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdIpSpacePublic(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName": t.Name(),

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpacePublicStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpacePublicStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdIpSpacePublicStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdIpSpacePublicStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdIpSpacePublicStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6DS := templateFill(testAccVcdIpSpacePublicStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText2, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
				),
			},
			{
				Config: configText3, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText4, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText5, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				ResourceName:      "vcd_ip_space.space1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     t.Name(),
			},
			{
				Config: configText6DS,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resourceFieldsEqual("data.vcd_ip_space.space1", "vcd_ip_space.space1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpacePublicStep1 = `
resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpacePublicStep2 = `
resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24"]
  external_scope = "8.8.8.0/23"

  route_advertisement_enabled = true
}
`

const testAccVcdIpSpacePublicStep3 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PUBLIC"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpacePublicStep4 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PUBLIC"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	default_quota = 2

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = -1

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }
}
`

const testAccVcdIpSpacePublicStep5 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PUBLIC"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	default_quota = 2

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}

	prefix {
		first_ip = "192.168.1.200"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = -1

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }

  ip_range {
	start_address = "11.11.11.100"
	end_address   = "11.11.11.110"
  }

  ip_range {
	start_address = "11.11.11.120"
	end_address   = "11.11.11.123"
  }
}
`

const testAccVcdIpSpacePublicStep5DS = testAccVcdIpSpacePublicStep5 + `
data "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
}
`

func TestAccVcdIpSpaceShared(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName": t.Name(),

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpaceSharedStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpaceSharedStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdIpSpaceSharedStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdIpSpaceSharedStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdIpSpaceSharedStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6DS := templateFill(testAccVcdIpSpaceSharedStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText2, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText3, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText4, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText5, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
					// sleepTester(2*time.Minute),
				),
			},
			{
				ResourceName:      "vcd_ip_space.space1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     t.Name(),
			},
			{
				Config: configText6DS,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resourceFieldsEqual("data.vcd_ip_space.space1", "vcd_ip_space.space1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpaceSharedStep1 = `
resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpaceSharedStep2 = `
resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24"]
  external_scope = "8.8.8.0/23"

  route_advertisement_enabled = true
}
`

const testAccVcdIpSpaceSharedStep3 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpaceSharedStep4 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	default_quota = 0 # no quota

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = 0 # no quota

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }
}
`

const testAccVcdIpSpaceSharedStep5 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	 default_quota = 0 # no quota

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}

	prefix {
		first_ip = "192.168.1.200"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = 0 # no quota

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }

  ip_range {
	start_address = "11.11.11.100"
	end_address   = "11.11.11.110"
  }

  ip_range {
	start_address = "11.11.11.120"
	end_address   = "11.11.11.123"
  }
}
`

const testAccVcdIpSpaceSharedStep5DS = testAccVcdIpSpaceSharedStep5 + `
data "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
}
`

func TestAccVcdIpSpacePrivate(t *testing.T) {
	preTestChecks(t)
	// skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName": t.Name(),
		"Org":      testConfig.VCD.Org,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpacePrivateStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpacePrivateStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdIpSpacePrivateStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdIpSpacePrivateStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdIpSpacePrivateStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6DS := templateFill(testAccVcdIpSpacePrivateStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText2, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText3, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText4, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText5, // minimal
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					// resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					// resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					// resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
					// sleepTester(2*time.Minute),
				),
			},
			{
				ResourceName:      "vcd_ip_space.space1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, t.Name()),
			},
			{
				Config: configText6DS,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resourceFieldsEqual("data.vcd_ip_space.space1", "vcd_ip_space.space1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpacePrivateShared = `
data "vcd_org" "org1" {
  name = "{{.Org}}"
}
`

const testAccVcdIpSpacePrivateStep1 = testAccVcdIpSpacePrivateShared + `
resource "vcd_ip_space" "space1" {
  name   = "{{.TestName}}"
  type   = "PRIVATE"
  org_id = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpacePrivateStep2 = testAccVcdIpSpacePrivateShared + `
resource "vcd_ip_space" "space1" {
  name   = "{{.TestName}}"
  type   = "PRIVATE"
  org_id = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24"]
  external_scope = "8.8.8.0/23"

  route_advertisement_enabled = true
}
`

const testAccVcdIpSpacePrivateStep3 = testAccVcdIpSpacePrivateShared + `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PRIVATE"
  org_id      = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false
}
`

const testAccVcdIpSpacePrivateStep4 = testAccVcdIpSpacePrivateShared + `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PRIVATE"
  org_id      = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	default_quota = -1 # unlimited

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = -1 # unlimited

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }
}
`

const testAccVcdIpSpacePrivateStep5 = testAccVcdIpSpacePrivateShared + `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PRIVATE"
  org_id      = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	 default_quota = -1 # unlimited

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}

	prefix {
		first_ip = "192.168.1.200"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = -1 # unlimited

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
	}
  }

  ip_range {
	start_address = "11.11.11.100"
	end_address   = "11.11.11.110"
  }

  ip_range {
	start_address = "11.11.11.120"
	end_address   = "11.11.11.123"
  }
}
`

const testAccVcdIpSpacePrivateStep5DS = testAccVcdIpSpacePrivateStep5 + `
data "vcd_ip_space" "space1" {
  org_id = data.vcd_org.org1.id
  name   = "{{.TestName}}"
}
`
