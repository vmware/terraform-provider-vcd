//go:build ALL || functional
// +build ALL functional

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
	vcdClient := createTemporaryVCDConnection()

	// Run a sub-test for each of data source defined in provider
	for _, dataSource := range Provider().DataSources() {
		t.Run(dataSource.Name, testSpecificDataSourceNotFound(t, dataSource.Name, vcdClient))
	}
	postTestChecks(t)
}

func testSpecificDataSourceNotFound(t *testing.T, dataSourceName string, vcdClient *VCDClient) func(*testing.T) {
	return func(t *testing.T) {

		// Skip sub-test if conditions are not met
		switch {
		case (dataSourceName == "vcd_external_network" || dataSourceName == "vcd_vcenter" ||
			dataSourceName == "vcd_portgroup" || dataSourceName == "vcd_global_role" ||
			dataSourceName == "vcd_rights_bundle") &&
			!usingSysAdmin():
			t.Skip(`Works only with system admin privileges`)
		case dataSourceName == "vcd_external_network_v2" && vcdClient.Client.APIVCDMaxVersionIs("< 33"):
			t.Skip("External network V2 requires at least API version 33 (VCD 10.0+)")
		case (dataSourceName == "vcd_nsxt_edgegateway" || dataSourceName == "vcd_nsxt_edge_cluster" ||
			dataSourceName == "vcd_nsxt_security_group" || dataSourceName == "vcd_nsxt_nat_rule" ||
			dataSourceName == "vcd_nsxt_app_port_profile" || dataSourceName == "vcd_nsxt_ip_set") &&
			(vcdClient.Client.APIVCDMaxVersionIs("< 34") || testConfig.Nsxt.Vdc == ""):
			t.Skip("this datasource requires at least API version 34 (VCD 10.1+) and NSX-T VDC specified in configuration")
		case (dataSourceName == "vcd_nsxt_tier0_router" || dataSourceName == "vcd_external_network_v2" ||
			dataSourceName == "vcd_nsxt_manager" || dataSourceName == "vcd_nsxt_edge_cluster") &&
			(testConfig.Nsxt.Manager == "" || testConfig.Nsxt.Tier0router == "" || !usingSysAdmin()):
			t.Skip(`No NSX-T configuration detected or not running as System user`)
		case dataSourceName == "vcd_nsxt_alb_controller" || dataSourceName == "vcd_nsxt_alb_cloud" ||
			dataSourceName == "vcd_nsxt_alb_importable_cloud" || dataSourceName == "vcd_nsxt_alb_service_engine_group":
			skipNoNsxtAlbConfiguration(t)
			if !usingSysAdmin() {
				t.Skip(`Works only with system admin privileges`)
			}
		// vcd_resource_list and vcd_resource_schema don't search for real entities
		case dataSourceName == "vcd_resource_list" || dataSourceName == "vcd_resource_schema":
			t.Skip(`not a real data source`)
		}

		// Get list of mandatory fields in schema for a particular data source
		mandatoryFields := getMandatoryDataSourceSchemaFields(dataSourceName)
		mandatoryRuntimeFields := getMandatoryDataSourceRuntimeFields(dataSourceName)
		mandatoryFields = append(mandatoryFields, mandatoryRuntimeFields...)
		addedParams := addMandatoryParams(dataSourceName, mandatoryFields, t, vcdClient)

		// Inject NSX-T VDC for resources that are known to require it
		switch dataSourceName {
		case "vcd_nsxt_edgegateway":
			addedParams += fmt.Sprintf(`vdc = "%s"`, testConfig.Nsxt.Vdc)
		}

		var params = StringMap{
			"DataSourceName":  dataSourceName,
			"MandatoryFields": addedParams,
		}

		params["FuncName"] = "NotFoundDataSource-" + dataSourceName
		// Adding skip directive as running these tests in binary test mode add no value
		binaryTestSkipText := "# skip-binary-test: data source not found test only works in acceptance tests\n"
		configText := templateFill(binaryTestSkipText+testAccUnavailableDataSource, params)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
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

		// A special case for DHCP relay where only invalid edge_gateway makes sense
		if dataSourceName == "vcd_nsxv_dhcp_relay" && mandatoryFields[fieldIndex] == "edge_gateway" {
			templateFields = templateFields + `edge_gateway = "non-existing"` + "\n"
			return templateFields
		}

		// vcd_portgroup requires portgroup  type
		if dataSourceName == "vcd_portgroup" && mandatoryFields[fieldIndex] == "type" {
			templateFields = templateFields + `type = "` + testConfig.Networking.ExternalNetworkPortGroupType + `"` + "\n"
		}

		switch mandatoryFields[fieldIndex] {
		// Fields, which must be valid to satisfy a data source
		case "org": // Some data sources require org - fill it from testConfig
			templateFields = templateFields + `org = "` + testConfig.VCD.Org + `"` + "\n"
		case "edge_gateway":
			templateFields = templateFields + `edge_gateway = "` + testConfig.Networking.EdgeGateway + `"` + "\n"
		case "edge_gateway_id":
			nsxtEdgeGw, err := vcdClient.GetNsxtEdgeGateway(testConfig.VCD.Org, testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway)
			if err != nil {
				t.Skipf("Unable to lookup NSX-T Edge Gateway '%s' : %s", testConfig.Nsxt.EdgeGateway, err)
				return ""
			}
			templateFields = templateFields + `edge_gateway_id = "` + nsxtEdgeGw.EdgeGateway.ID + `"` + "\n"
		case "catalog":
			templateFields = templateFields + `catalog = "` + testConfig.VCD.Catalog.Name + `"` + "\n"
		case "vapp_name":
			vapp, err := getAvailableVapp()
			if err != nil {
				t.Skip("No suitable vApp found for this test")
				return ""
			}
			templateFields = templateFields + `vapp_name = "` + vapp.VApp.Name + `"` + "\n"
		case "nsxt_manager_id":
			skipNoNsxtConfiguration(t)
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
			// Invalid fields which are required for some resources for search (usually they are used instead of `name`)
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
			templateFields = templateFields + `scope = "TENANT"` + "\n"
		case "controller_id":
			templateFields = templateFields + `controller_id = "urn:vcloud:loadBalancerController:90337fee-f332-40f2-a124-96e890eb1522"` + "\n"
		}

	}
	return templateFields
}
