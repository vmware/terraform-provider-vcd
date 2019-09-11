// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdVdcDatasource(t *testing.T) {
	validateConfiguration(t)

	vdcName := TestAccVcdVdc + "ForDataSourceTest"

	var params = StringMap{
		"ExistingVdcName": testConfig.VCD.Vdc,
		"VdcName":         vdcName,
		"OrgName":         testConfig.VCD.Org,
		"FuncName":        "TestAccVcdVdcDatasource",
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcDatasource requires system admin privileges")
		return
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	configText := templateFill(testAccCheckVcdVdcDatasource_basic, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceVdc := "vcd_org_vdc.existingVdc"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+vdcName),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "org", "vcd_org_vdc."+vdcName, "org"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "allocation_model", "vcd_org_vdc."+vdcName, "allocation_model"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "network_pool_name", "vcd_org_vdc."+vdcName, "network_pool_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "provider_vdc_name", "vcd_org_vdc."+vdcName, "provider_vdc_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enabled", "vcd_org_vdc."+vdcName, "enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enable_thin_provisioning", "vcd_org_vdc."+vdcName, "enable_thin_provisioning"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.enabled", "vcd_org_vdc."+vdcName, "storage_profile.0.enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.default", "vcd_org_vdc."+vdcName, "storage_profile.0.default"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "metadata.vdc_metadata", "VDC Metadata"),
					resource.ComposeTestCheckFunc(testAccDataSourceVcdOrgVdc("data."+datasourceVdc, vdcName)),
				),
			},
		},
	})
}

func testAccDataSourceVcdOrgVdc(name, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vdcResource, ok := s.RootModule().Resources["vcd_org_vdc."+vdcName]
		if !ok {
			return fmt.Errorf("can't find vcd_org_vdc.%s in state", vdcName)
		}

		attr := resources.Primary.Attributes

		mainHashValue, cpuHashInternalValue, memoryHashInternalValue, err := getHashValues(vdcResource.Primary.Attributes)
		if err != nil {
			return err
		}
		if attr["compute_capacity.0.cpu.0.allocated"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.allocated", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.allocated is %#v; want %#v", attr["compute_capacity.0.cpu.0.allocated"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.allocated", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.limit"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.limit", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.limit is %#v; want %#v", attr["compute_capacity.0.cpu.0.limit"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.limit", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.overhead"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.overhead", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.overhead is %#v; want %#v", attr["compute_capacity.0.cpu.0.overhead"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.overhead", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.reserved"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.reserved", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.reserved is %#v; want %#v", attr["compute_capacity.0.cpu.0.reserved"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.reserved", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.used"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.used", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.used is %#v; want %#v", attr["compute_capacity.0.cpu.0.used"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.cpu.%s.used", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.allocated"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.allocated", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.allocated is %#v; want %#v", attr["compute_capacity.0.memory.0.allocated"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.allocated", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.limit"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.limit", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.limit is %#v; want %#v", attr["compute_capacity.0.memory.0.limit"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.limit", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.overhead"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.overhead", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.overhead is %#v; want %#v", attr["compute_capacity.0.memory.0.overhead"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.overhead", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.reserved"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.reserved", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.reserved is %#v; want %#v", attr["compute_capacity.0.memory.0.reserved"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.reserved", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.used"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.used", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.used is %#v; want %#v", attr["compute_capacity.0.memory.0.used"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%s.memory.%s.used", mainHashValue, memoryHashInternalValue)])
		}

		return nil
	}
}

// Returns the hash part of key
// From "compute_capacity.315866465.cpu.798465156.limit",
// From "compute_capacity.315866465.memory.508945747.limit",
// will return "315866465", "798465156", "508945747"
func getHashValues(stateFileMap map[string]string) (string, string, string, error) {

	var cpuKey string
	var memoryKey string
	for k, _ := range stateFileMap {
		matched, err := regexp.MatchString(`compute_capacity+\.(\d+)\.cpu+\.(\d+)\.\w+`, k)
		if err != nil {
			return "", "", "", fmt.Errorf("error extracting hashes err %s", err)
		}
		if matched {
			cpuKey = k
		}
		matched, err = regexp.MatchString(`compute_capacity+\.(\d+)\.memory+\.(\d+)\.\w+`, k)
		if err != nil {
			return "", "", "", fmt.Errorf("error extracting hashes err %s", err)
		}
		if matched {
			memoryKey = k
		}
	}

	firstValue, secondValue, err := getHashesFromKey(cpuKey)
	if err != nil {
		return "", "", "", fmt.Errorf("error extracting hashes from '%s', err %s", cpuKey, err)
	}

	_, thirdValue, err := getHashesFromKey(memoryKey)
	if err != nil {
		return "", "", "", fmt.Errorf("error extracting hashes from '%s', err %s", memoryKey, err)
	}

	return firstValue, secondValue, thirdValue, nil
}

// Returns the hash part of key
// From "compute_capacity.315866465.memory.508945747.limit",
// will return "315866465", "508945747"
func getHashesFromKey(key string) (string, string, error) {

	// Regular expression to match key: compute_capacity.315866465.memory.508945747.limit
	reGetID := regexp.MustCompile(`\w+\.(\d+)\.\w+\.(\d+)\.\w+`)
	matchList := reGetID.FindAllStringSubmatch(key, -1)

	// matchList has the format
	// [][]string{[]string{"TOTAL MATCHED STRING", "CAPTURED STRING", "CAPTURED STRING"}}
	// such as
	// [][]string{[]string{"compute_capacity.315866465.memory.508945747.limit", "315866465", "508945747"}}
	if len(matchList) == 0 || len(matchList[0]) < 2 {
		return "", "", fmt.Errorf("error extracting ID from '%s'", key)
	}
	return matchList[0][1], matchList[0][2], nil
}

const testAccCheckVcdVdcDatasource_basic = `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "${data.vcd_org_vdc.existingVdc.allocation_model}"
  network_pool_name = "${data.vcd_org_vdc.existingVdc.network_pool_name}"
  provider_vdc_name = "${data.vcd_org_vdc.existingVdc.provider_vdc_name}"

  compute_capacity {
    cpu {
     allocated = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].cpu)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].cpu)[0].limit}"
    }

    memory {
     allocated = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].memory)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].memory)[0].limit}"
    }
  }

  storage_profile {
    name    = "${data.vcd_org_vdc.existingVdc.storage_profile[0].name}"
    enabled = "${data.vcd_org_vdc.existingVdc.storage_profile[0].enabled}"
    limit   = "${data.vcd_org_vdc.existingVdc.storage_profile[0].limit}"
    default = "${data.vcd_org_vdc.existingVdc.storage_profile[0].default}"
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = "${data.vcd_org_vdc.existingVdc.enabled}"
  enable_thin_provisioning = "${data.vcd_org_vdc.existingVdc.enable_thin_provisioning}"
  enable_fast_provisioning = "${data.vcd_org_vdc.existingVdc.enable_fast_provisioning}"
  delete_force             = true
  delete_recursive         = true
}
`
