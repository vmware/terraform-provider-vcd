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

// TestAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet tests
// * auto_subnet allocation with total_allocated_ip_count
func TestAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet(t *testing.T) {
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

	configText1 := templateFill(testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet2, params)
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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ips.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ips.#", "99"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
				),
			},
			{
				Config: configText2,
				Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
				Check: resource.ComposeTestCheckFunc(
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
					stateDumper(),
				),
			},
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

const testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet1 = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
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

const testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet2 = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
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

func TestAccVcdNsxtEdgeGatewayAutoAllocatedSubnet(t *testing.T) {
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

	configText1 := templateFill(testAccNsxtEdgeGatewayAutoAllocatedSubnetStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtEdgeGatewayAutoAllocatedSubnetStep2, params)
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
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "100"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_allocated_subnet.*", map[string]string{
						"gateway":            "93.0.0.1",
						"prefix_length":      "24",
						"allocated_ip_count": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_allocated_subnet.*", map[string]string{
						"gateway":            "14.14.14.1",
						"prefix_length":      "24",
						"allocated_ip_count": "9",
					}),
				),
			},
			{
				Config: configText2,
				// Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
				Check: resource.ComposeTestCheckFunc(
					stateDumper(),

					// sleepTester(4*time.Minute),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "204"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_allocated_subnet.*", map[string]string{
						"gateway":            "93.0.0.1",
						"prefix_length":      "24",
						"allocated_ip_count": "13",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_allocated_subnet.*", map[string]string{
						"gateway":            "14.14.14.1",
						"prefix_length":      "24",
						"allocated_ip_count": "12",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayAutoAllocatedSubnetStep1 = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  auto_allocated_subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

	primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
	allocated_ip_count = 9
  }

  auto_allocated_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
	allocated_ip_count = 9
  }
}
`

const testAccNsxtEdgeGatewayAutoAllocatedSubnetStep2 = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  auto_allocated_subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

	primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
	allocated_ip_count = 12
  }

  auto_allocated_subnet {
	gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
	allocated_ip_count = 13
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

//// TypeSet benchmark

func TestAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnetBenchmark(t *testing.T) {
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

	configText1 := templateFill(testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet1Benchmark, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	// params["FuncName"] = t.Name() + "-step2"
	// configText2 := templateFill(testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet2, params)
	// debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

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
					stateDumper(),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "100"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ips.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ips.#", "99"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
						"gateway":       "1.1.1.1",
						"prefix_length": "1",
					}),
					// resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
					// 	"gateway":       "14.14.14.1",
					// 	"prefix_length": "24",
					// }),
				),
			},
			// {
			// 	Config: configText2,
			// 	Taint:  []string{"vcd_nsxt_edgegateway.nsxt-edge"},
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "204"),
			// 		resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
			// 			"gateway":       "93.0.0.1",
			// 			"prefix_length": "24",
			// 		}),
			// 		resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "auto_subnet.*", map[string]string{
			// 			"gateway":       "14.14.14.1",
			// 			"prefix_length": "24",
			// 		}),
			// 		stateDumper(),
			// 	),
			// },
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgeGatewayAutoAllocationAutoSubnet1Benchmark = `
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
    gateway       = "1.1.1.1"
    prefix_length = "1"

    static_ip_pool {
      start_address = "0.0.0.1"
      end_address   = "127.255.255.254"
    }
  }

}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = 2147483646 # all IPs in the external network
  auto_subnet {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
  }
}
`
