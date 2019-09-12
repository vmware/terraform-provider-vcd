// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/terraform"
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

		checkDeepValues := func(key, parent, parentHash, child, childHash, label string, result *string) bool {

			expected := attr[key]
			value := vdcResource.Primary.Attributes[fmt.Sprintf("%s.%s.%s.%s.%s", parent, parentHash, child, childHash, label)]

			if expected != value {
				*result = fmt.Sprintf("%s.0.%s.0.%s is %v; want %v", parent, child, label, attr[key], expected)
				return false
			}
			return true
		}

		mainHashValue, cpuHashInternalValue, err := getHashValuesFromKey(vdcResource.Primary.Attributes, "compute_capacity", "cpu")
		if err != nil {
			return err
		}
		_, memoryHashInternalValue, err := getHashValuesFromKey(vdcResource.Primary.Attributes, "compute_capacity", "memory")
		if err != nil {
			return err
		}

		type testInfo struct {
			child     string
			childHash string
			label     string
		}
		var errorMsg string

		var testData = []testInfo{
			{"cpu", cpuHashInternalValue, "allocated"},
			{"cpu", cpuHashInternalValue, "limit"},
			{"cpu", cpuHashInternalValue, "overhead"},
			{"cpu", cpuHashInternalValue, "reserved"},
			{"cpu", cpuHashInternalValue, "used"},
			{"memory", memoryHashInternalValue, "allocated"},
			{"memory", memoryHashInternalValue, "limit"},
			{"memory", memoryHashInternalValue, "overhead"},
			{"memory", memoryHashInternalValue, "reserved"},
			{"memory", memoryHashInternalValue, "used"},
		}
		for _, td := range testData {
			key := fmt.Sprintf("compute_capacity.0.%s.0.%s", td.child, td.label)
			if !checkDeepValues(key, "compute_capacity", mainHashValue, td.child, td.childHash, td.label, &errorMsg) {
				return fmt.Errorf("%s", errorMsg)
			}
		}

		return nil
	}
}

func TestGetHashValuesFromKey(t *testing.T) {

	type testInfo struct {
		key      string
		parent   string
		child    string
		expected []string
	}
	var testData = []testInfo{
		{
			"first.1234.second.5678.third",
			"first",
			"second",
			[]string{"1234", "5678"},
		},
		{
			"compute_capacity.315866465.memory.508945747.limit",
			"compute_capacity",
			"memory",
			[]string{"315866465", "508945747"},
		},
		{
			"compute_capacity.315866465.cpu.798465156.limit",
			"compute_capacity",
			"cpu",
			[]string{"315866465", "798465156"},
		},
	}
	for _, td := range testData {
		testMap := map[string]string{
			td.key: "",
		}
		first, second, err := getHashValuesFromKey(testMap, td.parent, td.child)
		if err != nil {
			t.Logf("processing key '%s' got error %s", td.key, err)
			t.Fail()
		}
		if first != td.expected[0] {
			t.Logf("Expected result from key '%s' was '%s' - Got '%s' instead", td.key, td.parent, td.expected[0])
			t.Fail()
		}
		if second != td.expected[1] {
			t.Logf("Expected result from key '%s' was '%s' - Got '%s' instead", td.key, td.child, td.expected[1])
			t.Fail()
		}
	}
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
