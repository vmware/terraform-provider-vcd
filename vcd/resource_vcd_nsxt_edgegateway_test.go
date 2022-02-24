//go:build gateway || nsxt || ALL || functional
// +build gateway nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdNsxtEdgeGateway tests out creating and updating edge gateway using existing external network
// testConfig.Nsxt.ExternalNetwork which is expected to be correctly configured.
func TestAccVcdNsxtEdgeGateway(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1+)")
	}

	nsxtExtNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, testConfig.Nsxt.ExternalNetwork)
	if err != nil {
		t.Skipf("%s - could not retrieve external network", t.Name())
	}

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd": "nsxt-edge-test",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,
		"EdgeClusterId":      lookupAvailableEdgeClusterId(t, vcdClient),
		"Tags":               "gateway nsxt",
	}
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
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
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "description", "Updated-Description"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "edge_cluster_id", params["EdgeClusterId"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NsxtEdgeGatewayVcd"].(string)),
			},
		},
	})
	postTestChecks(t)
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
  description             = "Description"

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
  edge_cluster_id         = "{{.EdgeClusterId}}"

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

func lookupAvailableEdgeClusterId(t *testing.T, vcdClient *VCDClient) string {

	// Lookup available Edge Clusters to explicitly specify for edge gateway
	_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
	if err != nil {
		t.Errorf("error retrieving vdc: %s", err)
		t.FailNow()
	}

	eClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if len(eClusters) < 1 {
		t.Errorf("no edge clusters found: %s", err)
		t.FailNow()
	}

	return eClusters[0].NsxtEdgeCluster.ID
}

func TestAccVcdNsxtEdgeGatewayVdcGroup(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	//if vcdShortTest {
	//	t.Skip(acceptanceTestsSkipped)
	//	return
	//}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"Name":                      "TestAccVcdVdcGroupResource",
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"Description":               "myDescription",
		"DescriptionUpdate":         "myDescription updated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"OrgUser":                   testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":           testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":                    testConfig.Provider.Url,
		"OrgUserProvider":           "",
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"DfwUpdated":                "true",
		"DefaultPolicyUpdated":      "true",
		"DfwUpdated2":               "false",
		"DefaultPolicyUpdated2":     "false",
		"DfwUpdated3":               "true",
		"DefaultPolicyUpdated3":     "false",
		"DfwUpdated4":               "false",
		"DefaultPolicyUpdated4":     "true",
		"DfwUpdated5":               "true",
		"DefaultPolicyUpdated5":     "true",

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,
		"EdgeClusterId":      lookupAvailableEdgeClusterId(t, vcdClient),

		"Tags": "vdcGroup gateway nsxt",
	}

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
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			// initialize new VDC, this done separately as otherwise randomly fail due choose wrong connection
			resource.TestStep{
				Config: configTextPre,
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					//resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
					//	"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
					//	"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
					//	"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					//}),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Ignoring total field count '%' and 'starting_vdc_id' because it does not make sense for data source
					resourceFieldsEqual("vcd_nsxt_edgegateway.nsxt-edge", "data.vcd_nsxt_edgegateway.ds", []string{"starting_vdc_id", "%"}),
					//stateDumper(),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					//resourceFieldsEqual("data.vcd_nsxt_edgegateway.ds", "vcd_nsxt_edgegateway.nsxt-edge", nil),
					//stateDumper(),
				),
			},
		},
	})
	postTestChecks(t)
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}

const edgeVdcGroup = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org          = "{{.Org}}"
  //vdc          = vcd_org_vdc.newVdc.name
  owner_id     = vcd_vdc_group.test1.id
  name         = "{{.NsxtEdgeGatewayVcd}}"
  description  = "Description"

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

const edgeVdcGroup2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org          = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id
  name         = "{{.NsxtEdgeGatewayVcd}}"
  description  = "Description"

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

const edgeVdcGroupDS = edgeVdcGroup + `
data "vcd_nsxt_edgegateway" "ds" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.name

  name     = vcd_nsxt_edgegateway.nsxt-edge.name
  owner_id = vcd_vdc_group.test1.id
}
`

// TestAccVcdNsxtEdgeGatewayVdcGroupMigration tests the following migration scenario
// Step 1 - sets up prerequisites
// Step 2 - creates an Edge Gateway in a VDC using deprecated `vdc` field
// Step 3 - updates the Edge Gateway to use `owner_id` field instead of `vdc` field (keeping the same VDC)
// Step 4 - migrates the Edge Gateway to a VDC group
func TestAccVcdNsxtEdgeGatewayVdcGroupMigration(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	//if vcdShortTest {
	//	t.Skip(acceptanceTestsSkipped)
	//	return
	//}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"Name":                      "TestAccVcdVdcGroupResource",
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"Description":               "myDescription",
		"DescriptionUpdate":         "myDescription updated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"OrgUser":                   testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":           testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":                    testConfig.Provider.Url,
		"OrgUserProvider":           "",
		"Dfw":                       "false",
		"DefaultPolicy":             "false",

		"NsxtEdgeGatewayVcd": t.Name() + "-edge",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,
		"EdgeClusterId":      lookupAvailableEdgeClusterId(t, vcdClient),

		"Tags": "vdcGroup gateway nsxt",
	}

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

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			// initialize new VDC, this done separately as otherwise randomly fail due choose wrong connection
			resource.TestStep{
				Config: configTextPre,
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", "newVdc"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					//resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
					//	"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
					//	"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
					//	"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					//}),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					//resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
					//	"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
					//	"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
					//	"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					//}),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					//resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					//resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
					//	"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
					//	"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
					//	"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					//}),
				),
			},
			//resource.TestStep{
			//	Config: configText3,
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		// Ignoring total field count '%' and 'starting_vdc_id' because it does not make sense for data source
			//		resourceFieldsEqual("vcd_nsxt_edgegateway.nsxt-edge", "data.vcd_nsxt_edgegateway.ds", []string{"starting_vdc_id", "%"}),
			//		//stateDumper(),
			//	),
			//},
			//resource.TestStep{
			//	Config: configText4,
			//	Check: resource.ComposeAggregateTestCheckFunc(
			//		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", ""),
			//		//resourceFieldsEqual("data.vcd_nsxt_edgegateway.ds", "vcd_nsxt_edgegateway.nsxt-edge", nil),
			//		//stateDumper(),
			//	),
			//},
		},
	})
	postTestChecks(t)
}

const edgeVdcGroupMigration = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org          = "{{.Org}}"
  vdc          = vcd_org_vdc.newVdc.name
  //owner_id     = vcd_vdc_group.test1.id
  name         = "{{.NsxtEdgeGatewayVcd}}"
  description  = "Description"

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

const edgeVdcGroupMigration2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org          = "{{.Org}}"
  owner_id     = vcd_org_vdc.newVdc.id
  name         = "{{.NsxtEdgeGatewayVcd}}"
  description  = "Description"

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

// TestAccVcdNsxtEdgeGatewayVdcUpdateFails checks that it is impossible to update `vdc` field unless it is
// set to empty (in case of migration to `owner_id` field)
func TestAccVcdNsxtEdgeGatewayVdcUpdateFails(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection(false)

	nsxtExtNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, testConfig.Nsxt.ExternalNetwork)
	if err != nil {
		t.Skipf("%s - could not retrieve external network", t.Name())
	}

	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd":        "nsxt-edge-test",
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"EdgeClusterId":             lookupAvailableEdgeClusterId(t, vcdClient),
		"Name":                      "TestAccVcdVdcGroupResource",
		"Description":               "myDescription",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"OrgUser":                   testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":           testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":                    testConfig.Provider.Url,
		"OrgUserProvider":           "",

		"Tags": "vdcGroup gateway nsxt",
	}
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
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
				),
			},
			resource.TestStep{
				Config:      configText2,
				ExpectError: regexp.MustCompile(`changing 'vdc' field value is not supported`),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayVdcSwitch = testAccNsxtEdgeGatewayDataSources + `
# skip-binary-test: This test is expected to fail
resource "vcd_org_vdc" "newVdc" {
	provider = vcd
  
	name = "newVdc"
	org  = "{{.Org}}"
  
	allocation_model  = "Flex"
	network_pool_name = "{{.NetworkPool}}"
	provider_vdc_name = "{{.ProviderVdc}}"
  
	compute_capacity {
	  cpu {
		allocated = "{{.Allocated}}"
		limit     = "{{.Limit}}"
	  }
  
	  memory {
		allocated = "{{.Allocated}}"
		limit     = "{{.Limit}}"
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
	elasticity      			 = true
	include_vm_memory_overhead = true
  }

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
`

const testAccNsxtEdgeGatewayVdcSwitch2 = testAccNsxtEdgeGatewayDataSources + `
data "vcd_org_vdc" "test" {
	org  = "{{.Org}}"
	name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"

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
