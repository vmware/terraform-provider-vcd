//go:build gateway || alb || nsxt || ALL || functional || vdcGroup || network
// +build gateway alb nsxt ALL functional vdcGroup network

package vcd

import (
	"fmt"
	"os"
)

// testAccVcdVdcGroupNew is a helper definition to setup VDC Group for testing integration with other
// components
// Useful field names:
// * vcd_org_vdc.newVdc.0.id (new VDC)
// * vcd_org_vdc.newVdc.1.id (new VDC)
// * vcd_vdc_group.test1.id (VDC Group ID with two members listed above)

const testAccVcdVdcGroupNew = `
resource "vcd_org_vdc" "newVdc" {
  count = 2

  name = "{{.TestName}}-${count.index}"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "1024"
      limit     = "1024"
    }

    memory {
      allocated = "1024"
      limit     = "1024"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  network_quota = 100

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  include_vm_memory_overhead = true
  elasticity                 = true
}

resource "vcd_vdc_group" "test1" {
  org                   = "{{.Org}}"
  name                  = "{{.Name}}"
  starting_vdc_id       = vcd_org_vdc.newVdc.0.id
  participating_vdc_ids = vcd_org_vdc.newVdc.*.id
  
  dfw_enabled           = "{{.Dfw}}"
  default_policy_status = {{if eq .DefaultPolicy "true" }}true{{else}}false{{end}}
}
`

// overrideDefaultVdcForTest overrides default `vdc` value in `provider` section until it is
// reverted by using the returned func
//
// Simple use case is:
// restoreVdc := overrideDefaultVdcForTest(temporaryVdcFieldValue)
// defer restoreVdc()
//
// Note. When using in tests don't forget to override "PrVdc" field in 'params' for 'templateFill'
// fill function so that binary tests are rendered correctly as well.
func overrideDefaultVdcForTest(temporaryVdcFieldValue string) func() {
	originalVdcValue := os.Getenv("VCD_VDC")
	// testConfigOriginalVdcValue := testConfig.VCD.Vdc

	if vcdTestVerbose {
		fmt.Printf("# Overriding 'vdc' field in provider configuration to be '%s' instead of '%s'\n", temporaryVdcFieldValue, originalVdcValue)
	}

	os.Setenv("VCD_VDC", temporaryVdcFieldValue)
	return func() {
		if vcdTestVerbose {
			fmt.Printf("# Restoring 'vdc' field in provider configuration be '%s'\n", originalVdcValue)
		}
		os.Setenv("VCD_VDC", originalVdcValue)
	}
}
