//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtEdgeGatewayAutoSubnetAllocation tests
// * subnet_with_total_ip_count allocation with total_allocated_ip_count
func TestAccVcdNsxtEdgeGatewayAutoSubnetAllocation(t *testing.T) {
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
		"IpCount":             "100",

		"Tags": "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeGatewayAutoSubnetAllocation, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["IpCount"] = "204"
	configText2 := templateFill(testAccVcdNsxtEdgeGatewayAutoSubnetAllocation, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	params["IpCount"] = "199"
	configText3 := templateFill(testAccVcdNsxtEdgeGatewayAutoSubnetAllocation, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{ // create with subnet_with_total_ip_count
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "100"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "99"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
				),
			},
			{ // increase total_allocated_ip_count with subnet_with_total_ip_count
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "204"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "203"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
				),
			},
			{ // decrease total_allocated_ip_count with subnet_with_total_ip_count
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "199"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "198"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "93.0.0.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayAutoAllocationPrerequisites = `
data "vcd_org_vdc" "test" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

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

const testAccVcdNsxtEdgeGatewayAutoSubnetAllocation = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.test.id
  name      = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = {{.IpCount}}
  subnet_with_total_ip_count {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
    primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].start_address
  }

  subnet_with_total_ip_count {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
  }
}
`

// TestAccVcdNsxtEdgeGatewayAutoAllocatedSubnet
// subnet_with_ip_count
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
		"Subnet1Count":        "9",
		"Subnet2Count":        "10",

		"Tags": "gateway nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtEdgeGatewayAutoAllocatedSubnet, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["Subnet1Count"] = "12"
	params["Subnet2Count"] = "13"
	configText2 := templateFill(testAccNsxtEdgeGatewayAutoAllocatedSubnet, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	params["Subnet1Count"] = "5"
	params["Subnet2Count"] = "4"
	configText3 := templateFill(testAccNsxtEdgeGatewayAutoAllocatedSubnet, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			{ // create
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "18"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "19"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "14.14.14.1",
						"prefix_length":      "24",
						"allocated_ip_count": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "93.0.0.1",
						"prefix_length":      "24",
						"allocated_ip_count": "10",
					}),
				),
			},
			{ // update with increase
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "25"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "14.14.14.1",
						"prefix_length":      "24",
						"allocated_ip_count": "12",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "93.0.0.1",
						"prefix_length":      "24",
						"allocated_ip_count": "13",
					}),
				),
			},
			{ // update with decrease
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "8"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "9"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "14.14.14.1",
						"prefix_length":      "24",
						"allocated_ip_count": "5",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_ip_count.*", map[string]string{
						"gateway":            "93.0.0.1",
						"prefix_length":      "24",
						"allocated_ip_count": "4",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayAutoAllocatedSubnet = testAccNsxtEdgeGatewayAutoAllocationPrerequisites + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.test.id
  name      = "{{.NsxtEdgeGatewayVcd}}"


  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  subnet_with_ip_count {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

    primary_ip         = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
    allocated_ip_count = "{{.Subnet1Count}}"
  }

  subnet_with_ip_count {
	gateway            = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].gateway
	prefix_length      = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[1].prefix_length
	allocated_ip_count = "{{.Subnet2Count}}"
  }
}
`

// TestAccVcdNsxtEdgeGatewayAutoAllocationUsedAndUnusedIps tests that unused and used IPs are
// calculated correctly when having a huge subnet allocated to Edge Gateway. A subnet of 1.0.0.1/8
// that makes up a total of 16777213 IPs. It should be way bigger than any Edge Gateway can
// handle.
func TestAccVcdNsxtEdgeGatewayAutoAllocationUsedAndUnusedIps(t *testing.T) {
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

	configText1 := templateFill(testAccVcdNsxtEdgeGatewayAutoAllocationUsedAndUnusedIps, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "total_allocated_ip_count", "16777213"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "16777212"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet_with_total_ip_count.*", map[string]string{
						"gateway":       "1.0.0.1",
						"prefix_length": "8",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgeGatewayAutoAllocationUsedAndUnusedIps = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_org_vdc" "test" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
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
    gateway       = "1.0.0.1"
    prefix_length = "8"

    static_ip_pool {
      start_address = "1.0.0.2"
      end_address   = "1.255.255.254"
    }
  }
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org       = "{{.Org}}"
  owner_id  = data.vcd_org_vdc.test.id
  name      = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  total_allocated_ip_count = 16777213 
  subnet_with_total_ip_count {
    gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
    prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
    primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
  }
}
`
