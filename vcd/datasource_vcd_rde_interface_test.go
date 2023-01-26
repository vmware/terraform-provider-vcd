//go:build rde || ALL || functional
// +build rde ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// TestAccVcdRdeDefinedInterfaceDS tests the vcd_rde_interface as both System Administrator and tenant user.
func TestAccVcdRdeDefinedInterfaceDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"ProviderVcdSystem": providerVcdSystem,
		"ProviderVcdOrg1":   providerVcdOrg1,

		// This is a Defined Interface that comes with VCD out of the box
		"InterfaceNamespace": "k8s",
		"InterfaceVersion":   "1.0.0",
		"InterfaceVendor":    "vmware",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeDefinedInterfaceDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	sysadminInterfaceName := "data.vcd_rde_interface.sysadmin_interface_ds"
	tenantInterfaceName := "data.vcd_rde_interface.tenant_interface_ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sysadminInterfaceName, "namespace", params["InterfaceNamespace"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "version", params["InterfaceVersion"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "vendor", params["InterfaceVendor"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "name", "Kubernetes"), // Name is always the same
					resource.TestCheckResourceAttr(sysadminInterfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["InterfaceVendor"].(string), params["InterfaceNamespace"].(string), params["InterfaceVersion"].(string))),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "readonly", "false"),

					resource.TestCheckResourceAttrPair(tenantInterfaceName, "namespace", sysadminInterfaceName, "namespace"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "version", sysadminInterfaceName, "version"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "vendor", sysadminInterfaceName, "vendor"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "name", sysadminInterfaceName, "name"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "id", sysadminInterfaceName, "id"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "readonly", sysadminInterfaceName, "readonly"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterfaceDS = `
data "vcd_rde_interface" "sysadmin_interface_ds" {
  provider = {{.ProviderVcdSystem}}

  namespace = "{{.InterfaceNamespace}}"
  version   = "{{.InterfaceVersion}}"
  vendor    = "{{.InterfaceVendor}}"
}

data "vcd_rde_interface" "tenant_interface_ds" {
  provider = {{.ProviderVcdOrg1}}

  namespace = "{{.InterfaceNamespace}}"
  version   = "{{.InterfaceVersion}}"
  vendor    = "{{.InterfaceVendor}}"
}
`
