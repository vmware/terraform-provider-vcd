//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxtEdgeRateLimiting(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.2") {
		t.Skipf("This test tests VCD 10.3.2+ (API V36.2+) features. Skipping.")
	}

	qosPolicyName, err := findQosPolicy(vcdClient)
	if err != nil {
		t.Fatalf("error finding QoS profile: %s", err)
	}

	if qosPolicyName == "" {
		t.Skip("No QoS profile found. Skipping test")
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
		"NsxtQosPolicyName":    qosPolicyName,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeRateLimitingStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtEdgeRateLimitingStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

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
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_rate_limiting.testing-in-vdc-group", "id"),
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
		},
	})
}

const testAccVcdNsxtEdgeRateLimitingStep1 = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
  name = "{{.NsxtQosPolicyName}}"
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

		qosPolicy, err := edge.GetQoS()
		if err != nil {
			return fmt.Errorf("unable to get Qos profile: %s", err)
		}

		if qosPolicy.EgressProfile != nil && qosPolicy.EgressProfile.ID != "" {
			return fmt.Errorf("QoS Egress policy still exists")
		}

		if qosPolicy.IngressProfile != nil && qosPolicy.IngressProfile.ID != "" {
			return fmt.Errorf("QoS Ingress policy still exists")
		}

		return nil
	}
}

func findQosPolicy(vcdClient *VCDClient) (string, error) {
	nsxtManagers, err := vcdClient.QueryNsxtManagerByName(testConfig.Nsxt.Manager)
	if err != nil {
		return "", fmt.Errorf("unable to find NSX-T manager: %s", err)
	}

	id := extractUuid(nsxtManagers[0].HREF)
	nsxtManagerUrn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", id)
	if err != nil {
		return "", fmt.Errorf("could not construct URN from id '%s': %s", id, err)
	}

	allQosPolicies, err := vcdClient.GetAllNsxtEdgeGatewayQosProfiles(nsxtManagerUrn, nil)
	if err != nil {
		return "", fmt.Errorf("unable to find Qos profile: %s", err)
	}

	if len(allQosPolicies) == 0 {
		return "", nil
	}

	return allQosPolicies[0].NsxtEdgeGatewayQosProfile.DisplayName, nil
}
