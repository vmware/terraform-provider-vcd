//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdIpSpacePublic(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

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
		CheckDestroy:      testAccCheckVcdNsxtIpSpacesDestroy(params["TestName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PUBLIC"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
				),
			},
			{
				Config: configText3,
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
				Config: configText4,
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
				Config: configText5,
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

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 2

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = -1

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
    }
  }
}
`

const testAccVcdIpSpacePublicStep5 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "PUBLIC"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 2

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }

    prefix {
      first_ip      = "192.168.1.200"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = -1

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
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

  depends_on = [vcd_ip_space.space1]
}
`

func TestAccVcdIpSpaceShared(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

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
		CheckDestroy:      testAccCheckVcdNsxtIpSpacesDestroy(params["TestName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "SHARED_SERVICES"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "SHARED_SERVICES"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "SHARED_SERVICES"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "SHARED_SERVICES"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "SHARED_SERVICES"),
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

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 0 # no quota

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = 0 # no quota

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
    }
  }
}
`

const testAccVcdIpSpaceSharedStep5 = `
resource "vcd_ip_space" "space1" {
  name        = "{{.TestName}}"
  description = "added description"
  type        = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 0 # no quota

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }

    prefix {
      first_ip      = "192.168.1.200"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = 0 # no quota

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
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

  depends_on = [vcd_ip_space.space1]
}
`

func TestAccVcdIpSpacePrivate(t *testing.T) {
	preTestChecks(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

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
		CheckDestroy:      testAccCheckVcdNsxtIpSpacesDestroy(params["TestName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PRIVATE"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PRIVATE"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "1"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", "8.8.8.0/23"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PRIVATE"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PRIVATE"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_ip_space.space1", "internal_scope.*", "11.11.11.0/24"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "external_scope", ""),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "route_advertisement_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "type", "PRIVATE"),
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

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = -1 # unlimited

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = -1 # unlimited

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
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

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = -1 # unlimited

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }

    prefix {
      first_ip      = "192.168.1.200"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = -1 # unlimited

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
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

  depends_on = [vcd_ip_space.space1]
}
`

func testAccCheckVcdNsxtIpSpacesDestroy(ipSpaceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			ipSpaceResourceName := rs.Primary.Attributes["name"]
			if rs.Type != "vcd_ip_space" {
				continue
			}
			if ipSpaceResourceName != ipSpaceName {
				continue
			}
			conn := testAccProvider.Meta().(*VCDClient)
			_, err := conn.GetIpSpaceByName(ipSpaceResourceName)
			if err == nil {
				return fmt.Errorf("IP Space %s was not removed", ipSpaceName)
			}

		}

		return nil
	}
}
