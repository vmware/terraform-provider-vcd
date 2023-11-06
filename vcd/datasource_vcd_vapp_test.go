//go:build vapp || vm || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// getAvailableVapp collects one available Vapp to use in data source tests
func getAvailableVapp() (*govcd.VApp, error) {
	// Get the data from configuration file. This client is still inactive at this point
	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}
	vdc, err := org.GetVDCByName(testConfig.Nsxt.Vdc, false)
	if err != nil {
		return nil, fmt.Errorf("vdc not found : %s", err)
	}

	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				return vdc.GetVAppByHref(resourceReference.HREF)
			}
		}
	}

	return nil, fmt.Errorf("no vApp found in VDC %s", testConfig.Nsxt.Vdc)
}

// TestAccVcdVappDS tests a vApp data source if a vApp is found in the VDC
func TestAccVcdVappDS(t *testing.T) {
	preTestChecks(t)
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vapp, err := getAvailableVapp()
	if err != nil {
		t.Skip("No suitable vApp found for this test")
		return
	}

	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"VDC":      testConfig.Nsxt.Vdc,
		"VappName": vapp.VApp.Name,
		"FuncName": "TestVappDS",
		"Tags":     "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(datasourceTestVapp, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	statusText, err := vapp.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckOutput("name", vapp.VApp.Name),
					resource.TestCheckOutput("description", vapp.VApp.Description),
					resource.TestCheckOutput("href", vapp.VApp.HREF),
					resource.TestCheckOutput("status_text", statusText),
					resource.TestCheckResourceAttr("data.vcd_vapp."+vapp.VApp.Name, "status", fmt.Sprintf("%d", vapp.VApp.Status)),
				),
			},
		},
	})
	postTestChecks(t)
}

const datasourceTestVapp = `
data "vcd_vapp" "{{.VappName}}" {
  name             = "{{.VappName}}"
  org              = "{{.Org}}"
  vdc              = "{{.VDC}}"
}

output "name" {
  value = data.vcd_vapp.{{.VappName}}.name
}

output "description" {
  value = data.vcd_vapp.{{.VappName}}.description
}

output "href" {
  value = data.vcd_vapp.{{.VappName}}.href
}

output "status" {
  value = data.vcd_vapp.{{.VappName}}.status
}

output "status_text" {
  value = data.vcd_vapp.{{.VappName}}.status_text
}
`

func TestAccVcdVAppInheritedMetadata(t *testing.T) {
	preTestChecks(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 38.1") {
		t.Skipf("This test tests VCD 10.5.1+ (API V38.1+) features. Skipping.")
	}

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.Nsxt.Vdc,
		"Catalog":          testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"VappTemplateName": testConfig.VCD.Catalog.NsxtCatalogItem,
		"OvaPath":          testConfig.Ova.OvaPath,
		"FuncName":         t.Name(),
		"Tags":             "vapp",
	}
	testParamsNotEmpty(t, params)

	config := templateFill(testAccCheckVcdVAppInheritedMetadata, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", config)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vcd_vapp.fetched_vapp", "metadata.#", "0"),
					resource.TestCheckResourceAttr(
						"data.vcd_vapp.fetched_vapp", "metadata.%", "0"),
					resource.TestMatchResourceAttr(
						"data.vcd_vapp.fetched_vapp", "inherited_metadata.0.vapp_origin_id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttrSet(
						"data.vcd_vapp.fetched_vapp", "inherited_metadata.0.vapp_origin_name"),
					resource.TestMatchResourceAttr(
						"data.vcd_vapp.fetched_vapp", "inherited_metadata.0.vapp_origin_type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppInheritedMetadata = `
data "vcd_catalog" "cat" {
 org  = "{{.Org}}"
 name = "{{.Catalog}}"
}

data "vcd_catalog_vapp_template" "vapp_template" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = "{{.VappTemplateName}}"
}

resource "vcd_cloned_vapp" "vapp_from_template" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VappTemplateName}}"
  power_on      = true
  source_id     = data.vcd_catalog_vapp_template.vapp_template.id
  source_type   = "template"
  delete_source = false
}

data "vcd_vapp" "fetched_vapp" {
  name = vcd_cloned_vapp.vapp_from_template.name
  org  = vcd_cloned_vapp.vapp_from_template.org
  vdc  = vcd_cloned_vapp.vapp_from_template.vdc
}
`
