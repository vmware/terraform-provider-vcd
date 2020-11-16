// +build vapp vm ALL functional

package vcd

import (
	"fmt"
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
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}
	vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
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

	return nil, fmt.Errorf("no vApp found in VDC %s", testConfig.VCD.Vdc)
}

// TestAccVcdVappDS tests a vApp data source if a vApp is found in the VDC
func TestAccVcdVappDS(t *testing.T) {
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
		"VDC":      testConfig.VCD.Vdc,
		"VappName": vapp.VApp.Name,
		"FuncName": "TestVappDS",
		"Tags":     "vapp",
	}
	configText := templateFill(datasourceTestVapp, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	statusText, err := vapp.GetStatus()
	if err != nil {
		statusText = vAppUnknownStatus
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
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
