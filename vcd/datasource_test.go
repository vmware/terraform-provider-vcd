// +build ALL functional

package vcd

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccDataSourceNotFound is using Go sub-tests to ensure that "read" methods for all (current and future) data
// sources defined in this provider always return error and substring 'govcd.ErrorEntityNotFound' in itwhen an object is
// not found.
func TestAccDataSourceNotFound(t *testing.T) {
	// Exit the test early
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Run a sub-test for each of data source defined in provider
	allDataSources := getAllDatasourceNames()
	for _, dataSourceName := range allDataSources {
		t.Run(dataSourceName, testSpecificDataSourceNotFound(t, dataSourceName))
	}
}

func testSpecificDataSourceNotFound(t *testing.T, dataSourceName string) func(*testing.T) {
	return func(t *testing.T) {
		// Get list of mandatory fields in schema for a particular data source
		mandatoryFields := getMandatoryDataSourceSchemaFields(dataSourceName)
		mandatoryRuntimeFields := getMandatoryDataSourceRuntimeFields(dataSourceName)
		mandatoryFields = append(mandatoryFields, mandatoryRuntimeFields...)
		addedParams := addMandatoryParams(dataSourceName, mandatoryFields, t)

		var params = StringMap{
			"DataSourceName":  dataSourceName,
			"MandatoryFields": addedParams,
		}

		// Generate template without using `templateFill` because we do not need to store the config
		var configText bytes.Buffer
		configTemplate := template.Must(template.New("letter").Parse(testAccUnavailableDataSource))
		err := configTemplate.Execute(&configText, params)
		if err != nil {
			t.Errorf("could not generate template for %s data source", dataSourceName)
		}

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resource.Test(t, resource.TestCase{
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config:      configText.String(),
					ExpectError: regexp.MustCompile(`.*` + regexp.QuoteMeta(govcd.ErrorEntityNotFound.Error()) + `.*`),
				},
			},
		})

		return
	}
}

const testAccUnavailableDataSource = `
data "{{.DataSourceName}}" "not-existing" {
  {{.MandatoryFields}}
}
`

// getAllDatasourceNames returns all data source names defined in provider
func getAllDatasourceNames() []string {
	var allDataSources []string
	for _, ds := range Provider().DataSources() {
		allDataSources = append(allDataSources, ds.Name)
	}

	return allDataSources
}

// getMandatoryDataSourceSchemaFields checks schema definitions for data sources and return slice of mandatory fields
func getMandatoryDataSourceSchemaFields(dataSourceName string) []string {
	var mandatoryFields []string
	schema := dataSourceMap[dataSourceName]
	for fieldName, fieldSchema := range schema.Schema {
		if fieldSchema.Required {
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

func addMandatoryParams(dataSourceName string, mandatoryFields []string, t *testing.T) string {
	var templateFields string
	for fieldIndex := range mandatoryFields {

		// A special case for DHCP relay where only invalid edge_gateway makes sense
		if dataSourceName == "vcd_nsxv_dhcp_relay" && mandatoryFields[fieldIndex] == "edge_gateway" {
			templateFields = templateFields + `edge_gateway = "non-existing"` + "\n"
			return templateFields
		}

		switch mandatoryFields[fieldIndex] {
		// Fields, which must be valid to satisfy a data source
		case "org": // Some data sources require org - fill it from testConfig
			templateFields = templateFields + `org = "` + testConfig.VCD.Org + `"` + "\n"
		case "edge_gateway":
			templateFields = templateFields + `edge_gateway = "` + testConfig.Networking.EdgeGateway + `"` + "\n"
		case "catalog":
			templateFields = templateFields + `catalog = "` + testConfig.VCD.Catalog.Name + `"` + "\n"
		case "vapp_name":
			vapp, err := getAvailableVapp()
			if err != nil {
				t.Skip("No suitable vApp found for this test")
				return ""
			}
			templateFields = templateFields + `vapp_name = "` + vapp.VApp.Name + `"` + "\n"

			// Invalid fields which are required for some resources for search (usually they are used instead of `name`)
		case "rule_id":
			templateFields = templateFields + `rule_id = "347928347234"` + "\n"
		case "name":
			templateFields = templateFields + `name = "does-not-exist"` + "\n"
		}

	}
	return templateFields
}
