//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdRdeDefinedInterfaceDS tests the vcd_rde_interface as both System Administrator and tenant user.
func TestAccVcdRdeDefinedInterfaceDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"ProviderVcdSystem": providerVcdSystem,
		"ProviderVcdOrg1":   providerVcdOrg1,

		// This is a Defined Interface that comes with VCD out of the box
		"InterfaceNss":     "k8s",
		"InterfaceVersion": "1.0.0",
		"InterfaceVendor":  "vmware",
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
					resource.TestCheckResourceAttr(sysadminInterfaceName, "nss", params["InterfaceNss"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "version", params["InterfaceVersion"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "vendor", params["InterfaceVendor"].(string)),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "name", "Kubernetes"), // Name is always the same
					resource.TestCheckResourceAttr(sysadminInterfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["InterfaceVendor"].(string), params["InterfaceNss"].(string), params["InterfaceVersion"].(string))),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "readonly", "false"),
					resource.TestCheckResourceAttr(sysadminInterfaceName, "behavior.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(sysadminInterfaceName, "behavior.*", map[string]string{
						"name":           "createKubeConfig",
						"id":             "urn:vcloud:behavior-interface:createKubeConfig:vmware:k8s:1.0.0",
						"ref":            "urn:vcloud:behavior-interface:createKubeConfig:vmware:k8s:1.0.0",
						"description":    "Creates and returns a kubeconfig",
						"execution.%":    "2",
						"execution.id":   "CreateKubeConfigActivity",
						"execution.type": "Activity",
					}),

					resource.TestCheckResourceAttrPair(tenantInterfaceName, "nss", sysadminInterfaceName, "nss"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "version", sysadminInterfaceName, "version"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "vendor", sysadminInterfaceName, "vendor"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "name", sysadminInterfaceName, "name"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "id", sysadminInterfaceName, "id"),
					resource.TestCheckResourceAttrPair(tenantInterfaceName, "readonly", sysadminInterfaceName, "readonly"),
					resource.TestCheckResourceAttr(tenantInterfaceName, "behavior.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterfaceDS = `
data "vcd_rde_interface" "sysadmin_interface_ds" {
  provider = {{.ProviderVcdSystem}}

  nss     = "{{.InterfaceNss}}"
  version = "{{.InterfaceVersion}}"
  vendor  = "{{.InterfaceVendor}}"
}

data "vcd_rde_interface" "tenant_interface_ds" {
  provider = {{.ProviderVcdOrg1}}

  nss     = "{{.InterfaceNss}}"
  version = "{{.InterfaceVersion}}"
  vendor  = "{{.InterfaceVendor}}"
}
`
