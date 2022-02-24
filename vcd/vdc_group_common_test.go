//go:build vdcGroup || nsxt || ALL || functional
// +build vdcGroup nsxt ALL functional

package vcd

// testAccVcdVdcGroupNew is a helper definition to setup VDC Group for testing integration with other
// components
// Useful field names:
// * vcd_org_vdc.newVdc.id (new VDC)
// * vcd_vdc_group.test1.id (VDC Group ID with vcd_org_vdc.newVdc.id being a single member)

const testAccVcdVdcGroupNew = `
  resource "vcd_org_vdc" "newVdc" {
  provider = vcd

  name = "newVdc"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity      			 = true
  include_vm_memory_overhead = true
  }

resource "vcd_vdc_group" "test1" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.Name}}"
  description           = "{{.Description}}"
  starting_vdc_id       = vcd_org_vdc.newVdc.id
  participating_vdc_ids = [vcd_org_vdc.newVdc.id]
  
  dfw_enabled           = "{{.Dfw}}"
}
`
