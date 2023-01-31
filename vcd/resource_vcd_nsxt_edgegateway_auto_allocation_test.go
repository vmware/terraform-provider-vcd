//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtEdgeGatewayAutoAllocationTotal(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoConfiguration(t, StringMap{"Nsxt.ExternalNetwork": testConfig.Nsxt.ExternalNetwork})

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd":  t.Name(),
		"ExternalNetwork":     testConfig.Nsxt.ExternalNetwork,
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),

		"Tags": "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtEdgeGatewayAutoAllocationEmpty, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtEdgeGatewayAutoAllocationEmptyPrimaryIp, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "100"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					// stateDumper(),
				),
			},
			{
				Config: configText2,
				Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
				Check: resource.ComposeTestCheckFunc(
					// sleepTester(4*time.Minute),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "204"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					// stateDumper(),
				),
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdNsxtEdgeGatewayAutoAllocation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoConfiguration(t, StringMap{"Nsxt.ExternalNetwork": testConfig.Nsxt.ExternalNetwork})

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd":  t.Name(),
		"ExternalNetwork":     testConfig.Nsxt.ExternalNetwork,
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),

		"Tags": "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxtEdgeGatewayAutoAllocationEmpty, params)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccNsxtEdgeGatewayAutoAllocationTotal, params)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccNsxtEdgeGatewayAutoAllocationSubnet, params)
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
					// sleepTester(4*time.Minute),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "1"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					// stateDumper(),
				),
			},
			{
				Config: configText2,
				Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
				Check: resource.ComposeTestCheckFunc(
					// sleepTester(4*time.Minute),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "60"), // TODO - reenable
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
						// "primary_ip":    "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
						// "primary_ip":    "",
					}),
					// stateDumper(),
					// sleepTester(4*time.Minute),
				),
			},
			{
				Config: configText3,
				Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
				Check: resource.ComposeTestCheckFunc(
					// sleepTester(4*time.Minute),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "60"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
						// "total_ip_count": "5",
						// "primary_ip":    "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":        "14.14.14.1",
						"prefix_length":  "24",
						"total_ip_count": "5",
						// "primary_ip":    "",
					}),
					stateDumper(),
					sleepTester(4*time.Minute),
				),
			},
			// {
			// 	Config: configText1,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "description", "Updated-Description"),
			// 		resource.TestMatchResourceAttr(
			// 			"vcd_nsxt_edgegateway.nsxt-edge", "edge_cluster_id", params["EdgeClusterForAssert"].(*regexp.Regexp)),
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
			// 			"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
			// 			"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
			// 			"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
			// 		}),
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
			// 			"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
			// 			"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
			// 		}),
			// 	),
			// },
			// {
			// 	ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	ImportStateIdFunc: importStateIdOrgNsxtVdcObject(params["NsxtEdgeGatewayVcd"].(string)),
			// },
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayAutoAllocationPrerequisites = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = true
    gateway       = "93.0.0.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "93.0.0.10"
      end_address   = "93.0.0.100"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }

	static_ip_pool {
		start_address = "14.14.14.100"
		end_address   = "14.14.14.200"
	  }
  }
}
`

const testAccNsxtEdgeGatewayAutoAllocationEmpty = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = 100 # all IPs in the external network
  auto_subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
  }

  auto_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
  }
}
`

const testAccNsxtEdgeGatewayAutoAllocationEmptyPrimaryIp = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = 204 # all IPs in the external network
  auto_subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
	# primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
  }

  auto_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
  }
}
`

const testAccNsxtEdgeGatewayAutoAllocationTotal = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = 60
  auto_subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
	 #primary_ip            = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
  }

  auto_subnet {
	gateway               = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length         = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
  }
}
`

const testAccNsxtEdgeGatewayAutoAllocationSubnet = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  # total_allocated_ip_count = 60
  auto_subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

	 # primary_ip     = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
	 total_ip_count = 5
  }

  auto_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length

	#primary_ip     = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].static_ip_pool)[0].end_address

	// total_ip_count = 5
  }
}
`

const testAccNsxtEdgeGatewayUpdateAutoAllocation = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id       = vcd_external_network_v2.ext-net-nsxt.id
  dedicate_external_network = false

  total_allocated_ip_count = 10

  auto_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

	#primary_ip            = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
  }

  
}
`

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}
