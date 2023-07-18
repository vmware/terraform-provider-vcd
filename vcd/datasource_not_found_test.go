//go:build ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccDataSourceNotFound is using Go sub-tests to ensure that "read" methods for all (current and future) data
// sources defined in this provider always return error and substring 'govcd.ErrorEntityNotFound' in it when an object
// is not found.
func TestAccDataSourceNotFound(t *testing.T) {
	preTestChecks(t)
	// Exit the test early
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Setup temporary client to evaluate versions and conditionally skip tests
	vcdClient := createTemporaryVCDConnection(false)

	// Run a sub-test for each of data source defined in provider
	for _, dataSource := range Provider().DataSources() {
		t.Run(dataSource.Name, testSpecificDataSourceNotFound(dataSource.Name, vcdClient))
	}
	postTestChecks(t)
}

func testSpecificDataSourceNotFound(dataSourceName string, vcdClient *VCDClient) func(*testing.T) {
	return func(t *testing.T) {
		// Skip subtest based on versions
		type skipOnVersion struct {
			skipVersionConstraint string
			datasourceName        string
		}

		skipOnVersionsVersionsOlderThan := []skipOnVersion{
			{
				skipVersionConstraint: "< 36.2",
				datasourceName:        "vcd_nsxt_edgegateway_qos_profile",
			},
			{
				skipVersionConstraint: "< 37.1",
				datasourceName:        "vcd_ip_space",
			},
			{
				skipVersionConstraint: "< 37.1",
				datasourceName:        "vcd_ip_space_custom_quota",
			},
			{
				skipVersionConstraint: "< 37.1",
				datasourceName:        "vcd_ip_space_ip_allocation",
			},
			{
				skipVersionConstraint: "< 37.1",
				datasourceName:        "vcd_ip_space_uplink",
			},
			{
				skipVersionConstraint: "< 37.0",
				datasourceName:        "vcd_nsxt_edgegateway_static_route",
			},
			{
				skipVersionConstraint: "< 37.0",
				datasourceName:        "vcd_service_account",
			},
		}
		// urn:vcloud:ipSpace:2ec12e23-6911-4950-a33f-5602ae72ced2

		for _, constraintSkip := range skipOnVersionsVersionsOlderThan {
			if dataSourceName == constraintSkip.datasourceName && vcdClient.Client.APIVCDMaxVersionIs(constraintSkip.skipVersionConstraint) {
				t.Skipf("This test does not work on API versions %s", constraintSkip.skipVersionConstraint)
			}
		}

		// Skip sub-test if conditions are not met
		dataSourcesRequiringSysAdmin := []string{
			"vcd_external_network",
			"vcd_global_role",
			"vcd_nsxt_edgegateway_bgp_ip_prefix_list",
			"vcd_nsxt_edgegateway_bgp_neighbor",
			"vcd_org_ldap",
			"vcd_org_saml",
			"vcd_portgroup",
			"vcd_provider_vdc",
			"vcd_rights_bundle",
			"vcd_vcenter",
			"vcd_vdc_group",
			"vcd_vm_group",
			"vcd_resource_pool",
			"vcd_network_pool",
			"vcd_nsxt_edgegateway_qos_profile",
		}
		dataSourcesRequiringAlbConfig := []string{
			"vcd_nsxt_alb_cloud",
			"vcd_nsxt_alb_controller",
			"vcd_nsxt_alb_edgegateway_service_engine_group",
			"vcd_nsxt_alb_importable_cloud",
			"vcd_nsxt_alb_pool",
			"vcd_nsxt_alb_service_engine_group",
			"vcd_nsxt_alb_settings",
			"vcd_nsxt_alb_virtual_service",
			"vcd_nsxt_distributed_firewall",
		}
		dataSourcesRequiringNsxtConfig := []string{
			"vcd_external_network_v2",
			"vcd_nsxt_edge_cluster",
			"vcd_nsxt_manager",
			"vcd_nsxt_tier0_router",
		}

		switch {
		case contains(dataSourcesRequiringSysAdmin, dataSourceName) && !usingSysAdmin():
			t.Skip(`Works only with system admin privileges`)
		case contains(dataSourcesRequiringNsxtConfig, dataSourceName) &&
			(testConfig.Nsxt.Manager == "" || testConfig.Nsxt.Tier0router == "" || !usingSysAdmin()):
			t.Skip(`Nsxt.Manager, Nsxt.Tier0route is missing in configuration or not running as System user`)
		case contains(dataSourcesRequiringAlbConfig, dataSourceName):
			skipNoNsxtAlbConfiguration(t)
			if !usingSysAdmin() {
				t.Skip(`Works only with system admin privileges`)
			}
		// vcd_resource_list, vcd_resource_schema, and vcd_nsxv_application_finder don't produce a single entity
		case dataSourceName == "vcd_resource_list" || dataSourceName == "vcd_resource_schema" ||
			dataSourceName == "vcd_nsxv_application_finder":
			t.Skip(`not a real data source`)
		}

		// Get list of mandatory fields in schema for a particular data source
		mandatoryFields := getMandatoryDataSourceSchemaFields(dataSourceName)
		mandatoryRuntimeFields := getMandatoryDataSourceRuntimeFields(dataSourceName)
		mandatoryFields = append(mandatoryFields, mandatoryRuntimeFields...)
		addedParams := addMandatoryParams(dataSourceName, mandatoryFields, t, vcdClient)

		var params = StringMap{
			"DataSourceName":  dataSourceName,
			"MandatoryFields": addedParams,
		}
		if dataSourceName == "vcd_nsxv_distributed_firewall" {
			params["MandatoryFields"] = `vdc_id = "deadbeef-dead-beef-dead-beefdeadbeef"`
		}

		params["FuncName"] = "NotFoundDataSource-" + dataSourceName
		// Adding skip directive as running these tests in binary test mode add no value
		binaryTestSkipText := "# skip-binary-test: data source not found test only works in acceptance tests\n"
		configText := templateFill(binaryTestSkipText+testAccUnavailableDataSource, params)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      configText,
					ExpectError: regexp.MustCompile(`.*` + regexp.QuoteMeta(govcd.ErrorEntityNotFound.Error()) + `.*`),
				},
			},
		})
	}
}

const testAccUnavailableDataSource = `
data "{{.DataSourceName}}" "not-existing" {
  {{.MandatoryFields}}
}
`

// getMandatoryDataSourceSchemaFields checks schema definitions for data sources and return slice of mandatory fields
func getMandatoryDataSourceSchemaFields(dataSourceName string) []string {
	var mandatoryFields []string
	schema := globalDataSourceMap[dataSourceName]
	for fieldName, fieldSchema := range schema.Schema {
		if fieldSchema.Required || (len(fieldSchema.ExactlyOneOf) > 0 && fieldSchema.ExactlyOneOf[0] == fieldName) {
			mandatoryFields = append(mandatoryFields, fieldName)
		}
	}
	return mandatoryFields
}

// getMandatoryDataSourceRuntimeFields contains some exceptions where the schema does not require field, but it is
// validated during runtime and is mandatory
func getMandatoryDataSourceRuntimeFields(dataSourceName string) []string {
	// vcd_independent_disk validates at runtime if `id` or `name` are required
	if dataSourceName == "vcd_independent_disk" {
		return []string{"name"}
	}
	return []string{}
}

func addMandatoryParams(dataSourceName string, mandatoryFields []string, t *testing.T, vcdClient *VCDClient) string {
	var templateFields string

	for fieldIndex := range mandatoryFields {

		// validate that on provider config VDC added
		testParamsNotEmpty(t, StringMap{"VCD.Vdc": testConfig.Nsxt.Vdc})

		// A special case for DHCP relay where only invalid edge_gateway makes sense
		if dataSourceName == "vcd_nsxv_dhcp_relay" && mandatoryFields[fieldIndex] == "edge_gateway" {
			templateFields = templateFields + `edge_gateway = "non-existing"` + "\n"
			return templateFields
		}

		if dataSourceName == "vcd_task" && mandatoryFields[fieldIndex] == "id" {
			templateFields = templateFields + `id = "deadbeef-dead-beef-dead-deadbeefdead"` + "\n"
			return templateFields
		}

		if (dataSourceName == "vcd_nsxt_edgegateway_bgp_configuration" || dataSourceName == "vcd_nsxt_alb_settings" ||
			dataSourceName == "vcd_nsxt_edgegateway_rate_limiting" || dataSourceName == "vcd_nsxt_edgegateway_dhcp_forwarding" ||
			dataSourceName == "vcd_nsxt_firewall" || dataSourceName == "vcd_nsxt_route_advertisement" ||
			dataSourceName == "vcd_nsxt_edgegateway_dhcpv6") &&
			mandatoryFields[fieldIndex] == "edge_gateway_id" {
			// injecting fake Edge Gateway ID
			templateFields = templateFields + `edge_gateway_id = "urn:vcloud:gateway:784feb3d-87e4-4905-202a-bfe9faa5476f"` + "\n"
			return templateFields
		}

		if (dataSourceName == "vcd_org_saml" ||
			dataSourceName == "vcd_org_saml_metadata" ||
			dataSourceName == "vcd_org_ldap" ||
			dataSourceName == "vcd_ip_space_custom_quota" ||
			dataSourceName == "vcd_ip_space_ip_allocation") &&
			mandatoryFields[fieldIndex] == "org_id" {
			// injecting fake Org ID
			templateFields = templateFields + `org_id = "urn:vcloud:org:784feb3d-87e4-4905-202a-bfe9faa5476f"` + "\n"
			// return templateFields
		}

		if dataSourceName == "vcd_ip_space_ip_allocation" && mandatoryFields[fieldIndex] == "type" {
			templateFields = templateFields + `type = "FLOATING_IP"` + "\n"
		}

		// vcd_portgroup requires portgroup  type
		if dataSourceName == "vcd_portgroup" && mandatoryFields[fieldIndex] == "type" {
			templateFields = templateFields + `type = "` + testConfig.Networking.ExternalNetworkPortGroupType + `"` + "\n"
		}

		switch mandatoryFields[fieldIndex] {
		// Fields, which must be valid to satisfy a data source
		case "org": // Some data sources require org - fill it from testConfig
			testParamsNotEmpty(t, StringMap{"VCD.Org": testConfig.VCD.Org})
			templateFields = templateFields + `org = "` + testConfig.VCD.Org + `"` + "\n"
		case "edge_gateway":
			testParamsNotEmpty(t, StringMap{"Networking.EdgeGateway": testConfig.Networking.EdgeGateway})
			templateFields = templateFields + `edge_gateway = "` + testConfig.Networking.EdgeGateway + `"` + "\n"
		case "edge_gateway_id":
			testParamsNotEmpty(t, StringMap{
				"VCD.Org":                             testConfig.VCD.Org,
				"testConfig.Nsxt.VdcGroupEdgeGateway": testConfig.Nsxt.VdcGroupEdgeGateway,
				"Nsxt.VdcGroup":                       testConfig.Nsxt.VdcGroup})

			nsxtEdge, err := getNsxtEdgeGatewayInVdcGroup(vcdClient, testConfig.VCD.Org, testConfig.Nsxt.VdcGroup, testConfig.Nsxt.VdcGroupEdgeGateway)
			if err != nil {
				t.Errorf("error retrieving NSX-T Edge Gateway '%s' in VDC Group '%s': %s", testConfig.Nsxt.VdcGroupEdgeGateway, testConfig.Nsxt.VdcGroup, err)
			}
			templateFields = templateFields + `edge_gateway_id = "` + nsxtEdge.EdgeGateway.ID + `"` + "\n"
		case "catalog":
			testParamsNotEmpty(t, StringMap{"VCD.Catalog.Name": testConfig.VCD.Catalog.Name})
			templateFields = templateFields + `catalog = "` + testConfig.VCD.Catalog.Name + `"` + "\n"
		case "catalog_id":
			testParamsNotEmpty(t, StringMap{
				"VCD.Org":          testConfig.VCD.Org,
				"VCD.Catalog.Name": testConfig.VCD.Catalog.Name})
			org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
			if err != nil {
				t.Skip("No suitable Organization found for this test")
				return ""
			}
			catalog, err := org.GetCatalogByName(testConfig.VCD.Catalog.Name, false)
			if err != nil {
				t.Skip("No suitable Catalog found for this test")
				return ""
			}
			templateFields = templateFields + `catalog_id = "` + catalog.Catalog.ID + `"` + "\n"
		case "vdc_id":
			testParamsNotEmpty(t, StringMap{
				"VCD.Org": testConfig.VCD.Org,
				"VCD.Vdc": testConfig.VCD.Vdc})
			org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
			if err != nil {
				t.Skip("No suitable Organization found for this test")
				return ""
			}
			vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
			if err != nil {
				t.Skip("No suitable VDC found for this test")
				return ""
			}
			templateFields = templateFields + `vdc_id = "` + vdc.Vdc.ID + `"` + "\n"
		case "vapp_name":
			testParamsNotEmpty(t, StringMap{"VCD.Org": testConfig.VCD.Org, "testConfig.Nsxt.Vdc": testConfig.Nsxt.Vdc})
			vapp, err := getAvailableVapp()
			if err != nil {
				t.Skip("No suitable vApp found for this test")
				return ""
			}
			templateFields = templateFields + `vapp_name = "` + vapp.VApp.Name + `"` + "\n"
		case "vcenter_id":
			testParamsNotEmpty(t, StringMap{"Networking.Vcenter": testConfig.Networking.Vcenter})
			vcenter, err := vcdClient.GetVCenterByName(testConfig.Networking.Vcenter)
			if err != nil {
				t.Skip("No suitable Vcenter found for this test")
				return ""
			}
			templateFields = templateFields + `vcenter_id = "` + vcenter.VSphereVCenter.VcId + `"` + "\n"
		case "nsxt_manager_id":
			testParamsNotEmpty(t, StringMap{"Nsxt.Manager": testConfig.Nsxt.Manager})
			// This test needs a valid nsxt_manager_id
			nsxtManager, err := vcdClient.QueryNsxtManagerByName(testConfig.Nsxt.Manager)
			if err != nil {
				t.Skipf("No suitable NSX-T manager found for this test: %s", err)
				return ""
			}
			nsxtManagerUrn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(nsxtManager[0].HREF))
			if err != nil {
				t.Errorf("error building URN for NSX-T manager")
			}
			templateFields = templateFields + `nsxt_manager_id = "` + nsxtManagerUrn + `"` + "\n"
		case "context_id":
			testParamsNotEmpty(t, StringMap{"Nsxt.Manager": testConfig.Nsxt.Manager})
			// This test needs a valid nsxt_manager_id
			nsxtManager, err := vcdClient.QueryNsxtManagerByName(testConfig.Nsxt.Manager)
			if err != nil {
				t.Skipf("No suitable NSX-T manager found for this test: %s", err)
				return ""
			}
			nsxtManagerUrn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(nsxtManager[0].HREF))
			if err != nil {
				t.Errorf("error building URN for NSX-T manager")
			}
			templateFields = templateFields + `context_id = "` + nsxtManagerUrn + `"` + "\n"
			// Invalid fields which are required for some resources for search (usually they are used instead of `name`)
		case "vdc_group_id":
			templateFields = templateFields + `vdc_group_id = "urn:vcloud:vdcGroup:c19ec5b1-3403-4d00-b414-9da50066dc1e"` + "\n"
		case "provider_vdc_id":
			templateFields = templateFields + `provider_vdc_id = "urn:vcloud:providervdc:8453a2e2-1432-4e67-a312-8e713495eabc"` + "\n"
		case "rule_id":
			templateFields = templateFields + `rule_id = "347928347234"` + "\n"
		case "name":
			templateFields = templateFields + `name = "does-not-exist"` + "\n"
		case "alias":
			templateFields = templateFields + `alias = "does-not-exist"` + "\n"
		case "org_network_name":
			templateFields = templateFields + `org_network_name = "does-not-exist"` + "\n"
		// OpenAPI requires org_network_id to be a valid URN - chances of duplicating it are close enough to zero
		case "org_network_id":
			templateFields = templateFields + `org_network_id = "urn:vcloud:network:784feb3d-87e4-4905-202a-bfe9faa5476f"` + "\n"
		case "scope":
			templateFields = templateFields + `scope = "PROVIDER"` + "\n"
		case "controller_id":
			templateFields = templateFields + `controller_id = "urn:vcloud:loadBalancerController:90337fee-f332-40f2-a124-96e890eb1522"` + "\n"
		case "ip_address":
			templateFields = templateFields + `ip_address = "71.58.12.36"` + "\n"
		case "vendor":
			templateFields = templateFields + `vendor = "notexisting"` + "\n"
		case "nss":
			templateFields = templateFields + `nss = "notexisting"` + "\n"
		case "version":
			templateFields = templateFields + `version = "9.9.9"` + "\n"
		case "rde_type_id":
			templateFields = templateFields + `rde_type_id = "urn:vcloud:type:donotexist:donotexist:9.9.9"` + "\n"
		case "rde_interface_id":
			templateFields = templateFields + `rde_interface_id = "urn:vcloud:interface:notexist:notexist:9.9.9"` + "\n"
		case "behavior_id":
			templateFields = templateFields + `behavior_id = "urn:vcloud:behavior-interface:NotExist:notexist:notexist:9.9.9"` + "\n"
		case "ip_space_id":
			templateFields = templateFields + `ip_space_id = "urn:vcloud:ipSpace:90337fee-f332-40f2-a124-96e890eb1522"` + "\n"
		case "external_network_id":
			templateFields = templateFields + `external_network_id = "urn:vcloud:network:74804d82-a58f-4714-be84-75c178751ab0"` + "\n"
		}
	}

	// Inject NSX-V VDC for resources that are known to require it
	switch dataSourceName {
	case "vcd_edgegateway":
		testParamsNotEmpty(t, StringMap{"VCD.Vdc": testConfig.VCD.Vdc})
		templateFields += fmt.Sprintf(`vdc = "%s"`, testConfig.VCD.Vdc)
	case "vcd_nsxv_ip_set":
		testParamsNotEmpty(t, StringMap{"VCD.Vdc": testConfig.VCD.Vdc})
		templateFields += fmt.Sprintf(`vdc = "%s"`, testConfig.VCD.Vdc)
	case "vcd_nsxt_alb_virtual_service":
		testParamsNotEmpty(t, StringMap{"Nsxt.Vdc": testConfig.Nsxt.Vdc})
		templateFields += fmt.Sprintf(`vdc = "%s"`, testConfig.Nsxt.Vdc)
	case "vcd_nsxt_alb_edgegateway_service_engine_group":
		templateFields = templateFields + `service_engine_group_id = "does-not-exist"` + "\n"
	}

	return templateFields
}

func getNsxtEdgeGatewayInVdcGroup(cli *VCDClient, orgName, vdcName, edgeGwName string) (eg *govcd.NsxtEdgeGateway, err error) {
	if edgeGwName == "" {
		return nil, fmt.Errorf("empty NSX-T Edge Gateway name provided")
	}

	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(cli, orgName, vdcName)
	if err != nil {
		return nil, err
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGwName)
	if err != nil {
		return nil, err
	}

	return edge, nil
}
