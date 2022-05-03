//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbVdcGroupIntegration(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtAlbConfiguration(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"VirtualServiceName":        t.Name(),
		"ControllerName":            t.Name(),
		"ControllerUrl":             testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername":        testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword":        testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":           testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":          "DEDICATED",
		"Org":                       testConfig.VCD.Org,
		"NsxtVdc":                   testConfig.Nsxt.Vdc,
		"EdgeGw":                    testConfig.Nsxt.EdgeGateway,
		"IsActive":                  "true",
		"AliasPrivate":              t.Name() + "-cert",
		"Certificate1Path":          testConfig.Certificates.Certificate1Path,
		"CertPrivateKey1":           testConfig.Certificates.Certificate1PrivateKeyPath,
		"CertPassPhrase1":           testConfig.Certificates.Certificate1Pass,
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"NsxtImportSegment":         testConfig.Nsxt.NsxtImportSegment,
		"Name":                      t.Name(),
		"TestName":                  t.Name(),
		"NsxtExternalNetworkName":   testConfig.Nsxt.ExternalNetwork,

		"Tags": "nsxt alb vdcGroup",
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdNsxtAlbVdcGroupIntegration3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	parentVdcGroupName := t.Name()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
			testAccCheckVcdAlbVirtualServiceDestroy("vcd_nsxt_alb_virtual_service.test"),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_virtual_service.test", "id", regexp.MustCompile(`^urn:vcloud:loadBalancerVirtualService:`)),
				),
			},
			// Test ALB resource imports using VDC Group name in lookup path. (Parent NSX-T Edge Gateway is in VDC Group)
			{
				ResourceName:            "vcd_nsxt_alb_settings.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(testConfig, parentVdcGroupName, t.Name()),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_alb_edgegateway_service_engine_group.assignment",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), t.Name()+"-service-engine-group"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_alb_virtual_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), t.Name()+"-vs"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:            "vcd_nsxt_alb_pool.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(parentVdcGroupName, t.Name(), t.Name()+"-pool"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

// Config merges required prerequisites for ALB and VDC Group creation
const testAccVcdNsxtAlbVdcGroupIntegration1 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "{{.NsxtExternalNetworkName}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name = "{{.Name}}"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}


locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "{{.Name}}-service-engine-group"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "DEDICATED"
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name            = "{{.Name}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.newVdc.0.name

  name            = "{{.Name}}-vs"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`

const testAccVcdNsxtAlbVdcGroupIntegration2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "nsxt-ext-net" {
  name = "{{.NsxtExternalNetworkName}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "{{.Org}}"
  owner_id = vcd_vdc_group.test1.id

  name = "{{.Name}}"

  external_network_id = data.vcd_external_network_v2.nsxt-ext-net.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.nsxt-ext-net.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_alb_settings" "test" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}


locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "{{.Name}}-service-engine-group"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "DEDICATED"
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "{{.Org}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "{{.Org}}"

  name            = "{{.Name}}-pool"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "{{.Org}}"

  name            = "{{.Name}}-vs"
  edge_gateway_id = vcd_nsxt_alb_settings.test.edge_gateway_id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = tolist(vcd_nsxt_edgegateway.nsxt-edge.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    end_port   = 81
    type       = "TCP_PROXY"
  }
}
`
