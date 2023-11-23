//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// TestAccVcdNsxtEdgeGateway tests out creating and updating edge gateway using existing external network
// testConfig.Nsxt.ExternalNetwork which is expected to be correctly configured.
func TestAccVcdNsxtEdgeGateway(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoConfiguration(t, StringMap{"Nsxt.ExternalNetwork": testConfig.Nsxt.ExternalNetwork})
	vcdClient := createTemporaryVCDConnection(false)

	nsxtExtNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, testConfig.Nsxt.ExternalNetwork)
	if err != nil {
		t.Skipf("%s - could not retrieve external network", t.Name())
	}

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd": "nsxt-edge-test",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,
		"Tags":               "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	ifPossibleAddClusterId(t, vcdClient, params)

	configText := templateFill(testAccNsxtEdgeGateway, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccNsxtEdgeGatewayUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),

					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "0"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "description", "Updated-Description"),
					resource.TestMatchResourceAttr(
						"vcd_nsxt_edgegateway.nsxt-edge", "edge_cluster_id", params["EdgeClusterForAssert"].(*regexp.Regexp)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "0"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(params["NsxtEdgeGatewayVcd"].(string)),
			},
		},
	})
	postTestChecks(t)
}

// When test run in CDS then cluster ID isn't accessible.
// You will get error: Forbidden: User is not authorized to perform this operation on the application. Please contact the system administrator to get access., error code 401
// This function adds correct params if cluster ID found or not.
func ifPossibleAddClusterId(t *testing.T, vcdClient *VCDClient, params StringMap) {
	clusterId, err := lookupAvailableEdgeClusterId(t, vcdClient)
	if err != nil {
		t.Logf("\nWARNING: cluster id fetch failed, test will continue withouth cluster id. Error: %s", err)
		// adding regular expr param to map to use in Assertion
		params["EdgeClusterForAssert"] = getUuidRegex("", "$")
		params["EdgeClusterId"] = ""
		params["EdgeClusterKey"] = ""
		params["equalsChar"] = ""
	} else {
		params["EdgeClusterId"] = "\"" + clusterId + "\""
		params["EdgeClusterKey"] = "edge_cluster_id"
		params["equalsChar"] = "="
		// adding regular expr param to map to use in Assertion
		params["EdgeClusterForAssert"] = regexp.MustCompile(`^` + clusterId + `$`)
	}
}

const testAccNsxtEdgeGatewayDataSources = `
data "vcd_external_network_v2" "existing-extnet" {
  name = "{{.ExternalNetwork}}"
}
`

const testAccNsxtEdgeGateway = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const testAccNsxtEdgeGatewayUpdate = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
  description             = "Updated-Description"
  {{.EdgeClusterKey}}         {{.equalsChar}} {{.EdgeClusterId}}

  external_network_id = data.vcd_external_network_v2.existing-extnet.id
  dedicate_external_network = false

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length
     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

func testAccCheckVcdNsxtEdgeGatewayDestroy(edgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, rs := range s.RootModule().Resources {
			edgeGatewayName := rs.Primary.Attributes["name"]
			if rs.Type != "vcd_edgegateway" {
				continue
			}
			if edgeGatewayName != edgeName {
				continue
			}
			conn := testAccProvider.Meta().(*VCDClient)
			orgName := rs.Primary.Attributes["org"]
			vdcName := rs.Primary.Attributes["vdc"]

			org, _, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
			}

			_, err = org.GetNsxtEdgeGatewayByName(edgeName)
			if err == nil {
				return fmt.Errorf("NSX-T edge gateway %s was not removed", edgeName)
			}
		}

		return nil
	}
}

func TestAccVcdNsxtEdgeGatewayVdcGroup(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"Name":                      "TestAccVcdVdcGroupResource",
		"Description":               "myDescription",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"TestName":                  t.Name(),

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(edgeVdcGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3DS"
	configText3 := templateFill(edgeVdcGroupDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(edgeVdcGroup2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// initialize new VDC, this done separately as otherwise randomly fail due choose wrong connection
			{
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Ignoring total field count '%' and 'starting_vdc_id' because it does not make sense for data source
					resourceFieldsEqual("vcd_nsxt_edgegateway.nsxt-edge", "data.vcd_nsxt_edgegateway.ds", []string{"starting_vdc_id", "%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject("TestAccVcdVdcGroupResource", params["NsxtEdgeGatewayVcd"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const edgeVdcGroup = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const edgeVdcGroup2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const edgeVdcGroupDS = edgeVdcGroup + `
data "vcd_nsxt_edgegateway" "ds" {
  org = "{{.Org}}"

  name     = vcd_nsxt_edgegateway.nsxt-edge.name
  owner_id = vcd_vdc_group.test1.id
}
`

// TestAccVcdNsxtEdgeGatewayVdcGroupMigration has the main goal to test migration path from
// deprecated `vdc` to `owner_id`. It does so in the following steps:
// Step 1 - sets up prerequisites (a VDC Group with 2 VDCs in it)
// Step 2 - creates an Edge Gateway in a VDC using deprecated `vdc` field
// Step 3 - updates the Edge Gateway to use `owner_id` field instead of `vdc` field (keeping the same VDC)
// Step 4 - migrates the Edge Gateway to a VDC Group
// Step 5 - migrates the Edge Gateway to a different VDC than the starting one
func TestAccVcdNsxtEdgeGatewayVdcGroupMigration(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"Description":               "myDescription",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"TestName":                  t.Name(),

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(edgeVdcGroupMigration, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(edgeVdcGroupMigration2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(edgeVdcGroupMigration3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(edgeVdcGroupMigration4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				// Setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", fmt.Sprintf("%s-0", t.Name())),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const edgeVdcGroupMigration = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const edgeVdcGroupMigration2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.0.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const edgeVdcGroupMigration3 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

}
`

const edgeVdcGroupMigration4 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

// TestAccVcdNsxtEdgeGatewayVdcUpdateFails checks that it is impossible to update `vdc` field unless it is
// set to empty (in case of migration to `owner_id` field)
// After an expected failure it will just use the same VDC using `owner_id` instead of `vdc` field.
func TestAccVcdNsxtEdgeGatewayVdcUpdateFails(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd":        "nsxt-edge-test",
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"Name":                      "TestAccVcdVdcGroupResource",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccNsxtEdgeGateway, params)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccNsxtEdgeGatewayVdcSwitch, params)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccNsxtEdgeGatewayVdcSwitch2, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				Config:      configText2,
				ExpectError: regexp.MustCompile(`changing 'vdc' field value is not supported`),
			},
			{
				// Switch directly from `vdc` to the same VDC using `owner_id` field
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayVdcSwitch = testAccNsxtEdgeGatewayDataSources + `
# skip-binary-test: This test is expected to fail
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  vdc         = vcd_org_vdc.newVdc.name
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  depends_on = [vcd_org_vdc.newVdc]
}
resource "vcd_org_vdc" "newVdc" {
	name = "newVdc"
	org  = "{{.Org}}"
  
	allocation_model  = "Flex"
	network_pool_name = "{{.NetworkPool}}"
	provider_vdc_name = "{{.ProviderVdc}}"
  
	compute_capacity {
	  cpu {
		allocated = "1024"
		limit     = "1024"
	  }
  
	  memory {
		allocated = "1024"
		limit     = "1024"
	  }
	}
  
	storage_profile {
	  name    = "{{.ProviderVdcStorageProfile}}"
	  enabled = true
	  limit   = 10240
	  default = true
	}
  
	enabled                    = true
	enable_thin_provisioning   = true
	enable_fast_provisioning   = true
	delete_force               = true
	delete_recursive           = true
	elasticity      		   = true
	include_vm_memory_overhead = true
}
`

const testAccNsxtEdgeGatewayVdcSwitch2 = testAccNsxtEdgeGatewayDataSources + `
data "vcd_org_vdc" "test" {
	org  = "{{.Org}}"
	name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org = "{{.Org}}"

  owner_id    = data.vcd_org_vdc.test.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

func lookupAvailableEdgeClusterId(t *testing.T, vcdClient *VCDClient) (string, error) {
	// Lookup available Edge Clusters to explicitly specify for edge gateway
	_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
	if err != nil {
		t.Errorf("error retrieving vdc: %s", err)
		t.FailNow()
	}

	eClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if len(eClusters) < 1 {
		return "", fmt.Errorf("no edge clusters found: %s", err)
	}

	return eClusters[0].NsxtEdgeCluster.ID, nil
}

func TestAccVcdNsxtEdgeGatewayCreateInVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"Description":               "myDescription",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccNsxtEdgeGatewayInVdc, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccNsxtEdgeGatewayInVdcDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "data.vcd_org_vdc.test", "id"),
					// Comparing data source and resource fields. Ignoring total field count '%' because data source does not have `starting_vdc_id`
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(params["NsxtEdgeGatewayVcd"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayInVdc = `
data "vcd_external_network_v2" "existing-extnet" {
  name = "{{.ExternalNetwork}}"
}

data "vcd_org_vdc" "test" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"

  owner_id = data.vcd_org_vdc.test.id
  name     = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`
const testAccNsxtEdgeGatewayInVdcDS = testAccNsxtEdgeGatewayInVdc + `
# skip-binary-test: Cannot have resource and data source in the same file
data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "{{.Org}}"
  owner_id = vcd_nsxt_edgegateway.nsxt-edge.owner_id
  name     = vcd_nsxt_edgegateway.nsxt-edge.name
}
`

func TestAccVcdNsxtEdgeGatewayT0AndExternalNetworkUplink(t *testing.T) {
	testAccVcdNsxtEdgeGatewayExternalNetworkUplink(t, testConfig.Nsxt.Tier0router)
}

func TestAccVcdNsxtEdgeGatewayT0VrfAndExternalNetworkUplink(t *testing.T) {
	testAccVcdNsxtEdgeGatewayExternalNetworkUplink(t, testConfig.Nsxt.Tier0routerVrf)
}

func testAccVcdNsxtEdgeGatewayExternalNetworkUplink(t *testing.T, t0GatewayName string) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		t.Skipf("Segment Backed NSX-T Edge Gateway Uplinks require at least VCD 10.4.1+ (API v37.1+)")
	}

	skipNoConfiguration(t, StringMap{"Nsxt.ExternalNetwork": testConfig.Nsxt.ExternalNetwork})
	vcdClient := createTemporaryVCDConnection(false)

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"TestName":            t.Name(),
		"T0Gateway":           t0GatewayName,
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtSegment":         testConfig.Nsxt.NsxtImportSegment,
		"NsxtSegment2":        testConfig.Nsxt.NsxtImportSegment2,
		"NsxtEdgeGatewayVcd":  t.Name(),
		"ExternalNetwork":     testConfig.Nsxt.ExternalNetwork,
		"ExternalNetworkName": t.Name() + "-segment-backed",
		"Tags":                "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	ifPossibleAddClusterId(t, vcdClient, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4ds"
	configText4DS := templateFill(testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// This fields must reports counts for Tier0 gateway backed uplink, therefore it must not be impacted by additional uplinks
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "external_network_allocated_ip_count", "8"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "6"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "14.14.14.1",
						"primary_ip":         "14.14.14.10",
						"prefix_length":      "24",
						"allocated_ip_count": "4",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "15.14.14.1",
						"primary_ip":         "15.14.14.12",
						"prefix_length":      "24",
						"allocated_ip_count": "4",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// This fields must reports counts for Tier0 gateway backed uplink, therefore it must not be impacted by additional uplinks
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "external_network_allocated_ip_count", "10"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "3"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "8"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "14.14.14.1",
						"primary_ip":         "14.14.14.10",
						"prefix_length":      "24",
						"allocated_ip_count": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "15.14.14.1",
						"primary_ip":         "15.14.14.12",
						"prefix_length":      "24",
						"allocated_ip_count": "8",
					}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					// This fields must reports counts for Tier0 gateway backed uplink, therefore it must not be impacted by additional uplinks
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "external_network_allocated_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "14.14.14.1",
						"primary_ip":         "14.14.14.10",
						"prefix_length":      "24",
						"allocated_ip_count": "1",
					}),
				),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdNsxtEdgeGatewayExternalNetworkUplinkShared = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.T0Gateway}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt-t0" {
  name        = "{{.TestName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "54.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "54.14.14.10"
      end_address   = "54.14.14.15"
    }

    static_ip_pool {
      start_address = "54.14.14.20"
      end_address   = "54.14.14.25"
    }
  }
}

data "vcd_org_vdc" "nsxt" {
  name = "{{.NsxtVdc}}"
}

resource "vcd_external_network_v2" "segment-backed" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.19"
    }
    
    static_ip_pool {
      start_address = "14.14.14.22"
      end_address   = "14.14.14.29"
    }
  }
}

resource "vcd_external_network_v2" "segment-backed2" {
  name = "{{.ExternalNetworkName}}-2"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment2}}"
  }

  ip_scope {
    gateway       = "15.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "15.14.14.10"
      end_address   = "15.14.14.19"
    }
  }
}
`

const testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep1 = testAccVcdNsxtEdgeGatewayExternalNetworkUplinkShared + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.nsxt.id
  name      = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt-t0.id

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].prefix_length
     primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address

     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 4
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed2.id
    gateway             = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].prefix_length
    allocated_ip_count  = 4
    primary_ip          = "15.14.14.12"
  }
}
`

const testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep2 = testAccVcdNsxtEdgeGatewayExternalNetworkUplinkShared + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.nsxt.id
  name      = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt-t0.id

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].prefix_length
     primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address

     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 2
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed2.id
    gateway             = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].prefix_length
    allocated_ip_count  = 8
    primary_ip          = "15.14.14.12"
  }
}
`

const testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep3 = testAccVcdNsxtEdgeGatewayExternalNetworkUplinkShared + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.nsxt.id
  name      = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt-t0.id

  subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].prefix_length
    primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address

    allocated_ips {
      start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
      end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
    }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 1
  }
}
`

const testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep4DS = testAccVcdNsxtEdgeGatewayExternalNetworkUplinkStep3 + `
data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = vcd_nsxt_edgegateway.nsxt-edge.name
}
`

func TestAccVcdNsxtEdgeGatewayVdcGroupExternalUplink(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		t.Skipf("Segment Backed NSX-T Edge Gateway Uplinks require at least VCD 10.4.1+ (API v37.1+)")
	}
	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      t.Name(),
		"Description":               "myDescription",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"T0Gateway":                 testConfig.Nsxt.Tier0router,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"NsxtManager":               testConfig.Nsxt.Manager,
		"NsxtSegment":               testConfig.Nsxt.NsxtImportSegment,
		"NsxtSegment2":              testConfig.Nsxt.NsxtImportSegment2,
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"TestName":                  t.Name(),

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		"Tags": "vdcGroup gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNsxtEdgeGatewayVdcGroupExternalUplink, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgeGatewayVdcGroupExternalUplink = testAccVcdVdcGroupNew + `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.T0Gateway}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt-t0" {
  name        = "{{.TestName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "54.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "54.14.14.10"
      end_address   = "54.14.14.15"
    }

    static_ip_pool {
      start_address = "54.14.14.20"
      end_address   = "54.14.14.25"
    }
  }
}

resource "vcd_external_network_v2" "segment-backed" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.19"
    }
    
    static_ip_pool {
      start_address = "14.14.14.22"
      end_address   = "14.14.14.29"
    }
  }
}

resource "vcd_external_network_v2" "segment-backed2" {
  name = "{{.ExternalNetworkName}}-2"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment2}}"
  }

  ip_scope {
    gateway       = "15.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "15.14.14.10"
      end_address   = "15.14.14.19"
    }
  }
}

data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = vcd_external_network_v2.ext-net-nsxt-t0.id

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt-t0.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 2
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed2.id
    gateway             = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed2.ip_scope)[0].prefix_length
    allocated_ip_count  = 8
    primary_ip          = "15.14.14.12"
  }
}
`
