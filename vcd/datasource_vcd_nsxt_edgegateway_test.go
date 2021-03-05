// +build functional gateway nsxt ALL

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtEdgeGatewayMultipleSubnets test creates its own external network with many subnets and tests if edge
// gateway resource can correctly consume these multiple subnets
func TestAccVcdNsxtEdgeGatewayMultipleSubnetsAndDS(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1+)")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd": "nsxt-edge-test-multi-subnet",
		"NsxtManager":        testConfig.Nsxt.Manager,
		"Tier0Router":        testConfig.Nsxt.Tier0router,
		"EdgeClusterId":      lookupAvailableEdgeClusterId(t, vcdClient),
		"Tags":               "gateway nsxt",
	}
	configText := templateFill(testAccNsxtEdgeGatewayMultipleSubnets, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccNsxtEdgeGatewayMultipleSubnetsDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", "99.99.99.23"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "external_network_id"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       "88.88.88.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": "88.88.88.91",
						"end_address":   "88.88.88.92",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": "88.88.88.94",
						"end_address":   "88.88.88.95",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": "88.88.88.97",
						"end_address":   "88.88.88.98",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       "99.99.99.1",
						"prefix_length": "25",
						"primary_ip":    "99.99.99.23",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": "99.99.99.22",
						"end_address":   "99.99.99.24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       "77.77.77.1",
						"prefix_length": "26",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": "77.77.77.10",
						"end_address":   "77.77.77.12",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NsxtEdgeGatewayVcd"].(string)),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "name", "data.vcd_nsxt_edgegateway.egw-ds", "name"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "external_network_id", "data.vcd_nsxt_edgegateway.egw-ds", "external_network_id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "subnet", "data.vcd_nsxt_edgegateway.egw-ds", "subnet"),
					// Ensure all attributes are available on data source as on the resource itself
					resourceFieldsEqual("vcd_nsxt_edgegateway.nsxt-edge", "data.vcd_nsxt_edgegateway.egw-ds", []string{}),
				),
			},
		},
	})
}

const testAccNsxtEdgeGatewayMultipleSubnets = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.Tier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "test-nsxt-external-network"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "88.88.88.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }

  ip_scope {
    gateway       = "99.99.99.1"
    prefix_length = "25"

    static_ip_pool {
      start_address = "99.99.99.10"
      end_address   = "99.99.99.15"
    }

    static_ip_pool {
      start_address = "99.99.99.20"
      end_address   = "99.99.99.25"
    }
  }

  ip_scope {
    gateway       = "77.77.77.1"
    prefix_length = "26"

    static_ip_pool {
      start_address = "77.77.77.10"
      end_address   = "77.77.77.15"
    }

    static_ip_pool {
      start_address = "77.77.77.20"
      end_address   = "77.77.77.25"
    }
  }

  ip_scope {
    gateway       = "66.66.66.1"
    prefix_length = "27"

    static_ip_pool {
      start_address = "66.66.66.5"
      end_address   = "66.66.66.7"
    }

    static_ip_pool {
      start_address = "66.66.66.9"
      end_address   = "66.66.66.10"
    }
  }
}


resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  subnet {
     gateway       = "88.88.88.1"
     prefix_length = "24"

     allocated_ips {
       start_address = "88.88.88.91"
       end_address   = "88.88.88.92"
     }

     allocated_ips {
       start_address = "88.88.88.94"
       end_address   = "88.88.88.95"
     }

     allocated_ips {
       start_address = "88.88.88.97"
       end_address   = "88.88.88.98"
     }
  }

  subnet {
     gateway       = "99.99.99.1"
     prefix_length = "25"
	 # primary_ip should fall into defined "allocated_ips" as otherwise next apply will report additional range of
	 # "allocated_ips" with the range containing single "primary_ip" and will cause non-empty plan.
     primary_ip    = "99.99.99.23"

     allocated_ips {
       start_address = "99.99.99.22"
       end_address   = "99.99.99.24"
     }
  }

  subnet {
     gateway       = "77.77.77.1"
     prefix_length = "26"

     allocated_ips {
       start_address = "77.77.77.10"
       end_address   = "77.77.77.12"
	 }
  }
}
`
const testAccNsxtEdgeGatewayMultipleSubnetsDS = testAccNsxtEdgeGatewayMultipleSubnets + `
# skip-binary-test: resource and data source cannot refer itself in a single file
data "vcd_nsxt_edgegateway" "egw-ds" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
}
`

// TestAccVcdNsxtEdgeGatewayDSDoesNotAcceptNsxv expects to get an error because it tries to lookup NSX-V edge gateway
// using NSX-T datasource. There is a validator inside `vcd_nsxt_edgegateway` which is supposed to refer to
// `vcd_edgegateway` when VDC is NSX-V
func TestAccVcdNsxtEdgeGatewayDSDoesNotAcceptNsxv(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1+)")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxvVdc":             testConfig.VCD.Vdc,
		"NsxvEdgeGatewayName": testConfig.Networking.EdgeGateway,
		"Tags":                "gateway nsxt",
	}

	configText := templateFill(testAccVcdNsxtEdgeGatewayDSDoesNotAcceptNsxv, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText,
				ExpectError: regexp.MustCompile("please use 'vcd_edgegateway' for NSX-V backed VDC"),
			},
		},
	})
}

const testAccVcdNsxtEdgeGatewayDSDoesNotAcceptNsxv = `
# skip-binary-test: should fail on purpose because NSX-T datasource should not accept NSX-V edge gateway
data "vcd_nsxt_edgegateway" "nsxv-try" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxvVdc}}"
  name                    = "{{.NsxvEdgeGatewayName}}"
}
`
