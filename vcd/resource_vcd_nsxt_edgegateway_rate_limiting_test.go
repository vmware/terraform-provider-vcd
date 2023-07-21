//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtEdgeRateLimiting(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 36.2") {
		t.Skipf("This test tests VCD 10.3.2+ (API V36.2+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"NsxtEdgeGw":           testConfig.Nsxt.EdgeGateway,
		"TestName":             t.Name(),
		"NsxtManager":          testConfig.Nsxt.Manager,
		"NsxtQosProfileName":   testConfig.Nsxt.GatewayQosProfile,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeRateLimitingStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtEdgeRateLimitingStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtEdgeRateLimitingStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtEdgeRateLimitDestroy(params["NsxtVdc"].(string), params["NsxtEdgeGw"].(string)),
			testAccCheckNsxtEdgeRateLimitDestroy(params["NsxtVdcGroup"].(string), params["NsxtEdgeGwInVdcGroup"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "ingress_profile_id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "egress_profile_id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "ingress_profile_id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "egress_profile_id"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(params["NsxtEdgeGw"].(string)),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(params["NsxtVdcGroup"].(string), params["NsxtEdgeGwInVdcGroup"].(string)),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "id"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "id"),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "id"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "ingress_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc", "egress_profile_id", ""),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "ingress_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "egress_profile_id", ""),
				),
			},
		},
	})
}

const testAccVcdNsxtEdgeRateLimitingShared = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
  name            = "{{.NsxtQosProfileName}}"
}

data "vcd_vdc_group" "g1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc-group" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.g1.id

  name = "{{.NsxtEdgeGwInVdcGroup}}"
}

data "vcd_org_vdc" "v1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.v1.id

  name = "{{.NsxtEdgeGw}}"
}
`

const testAccVcdNsxtEdgeRateLimitingStep1 = testAccVcdNsxtEdgeRateLimitingShared + `
resource "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  ingress_profile_id = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
  egress_profile_id  = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
}

resource "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id

  ingress_profile_id = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
  egress_profile_id  = data.vcd_nsxt_edgegateway_qos_profile.qos-1.id
}
`

const testAccVcdNsxtEdgeRateLimitingStep2DS = testAccVcdNsxtEdgeRateLimitingStep1 + `
# skip-binary-test: datasource test will fail when run together with resource
data "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id
}

data "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id
}
`

const testAccVcdNsxtEdgeRateLimitingStep3 = testAccVcdNsxtEdgeRateLimitingShared + `
resource "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id
}

resource "vcd_nsxt_edgegateway_rate_limiting" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id
}
`

// testAccCheckNsxtEdgeRateLimitDestroy checks if the resource was destroyed.
// "destroy" means that Edge Gateway Qos profile was removed (is unlimited).
func testAccCheckNsxtEdgeRateLimitDestroy(vdcOrVdcGroupName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		vdcOrVdcGroup, err := lookupVdcOrVdcGroup(conn, testConfig.VCD.Org, vdcOrVdcGroupName)
		if err != nil {
			return fmt.Errorf("unable to find VDC or VDC group %s: %s", vdcOrVdcGroupName, err)
		}

		edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		qosConfig, err := edge.GetQoS()
		if err != nil {
			return fmt.Errorf("unable to get Qos profile: %s", err)
		}

		if qosConfig.EgressProfile != nil && qosConfig.EgressProfile.ID != "" {
			return fmt.Errorf("QoS Egress profile still exists")
		}

		if qosConfig.IngressProfile != nil && qosConfig.IngressProfile.ID != "" {
			return fmt.Errorf("QoS Ingress profile still exists")
		}

		return nil
	}
}
