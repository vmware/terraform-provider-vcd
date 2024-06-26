//go:build api || functional || ALL

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccDse will test both 'vcd_dse_registry_configuration' and 'vcd_dse_solution_publish' so
// that it saves time on Solution Add-On publishing
func TestAccDse(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if testConfig.SolutionAddOn.Org == "" || len(testConfig.SolutionAddOn.DseSolutions) < 6 {
		t.Skipf("Solution Add-On config value not specified")
	}

	vcdClient := createTemporaryVCDConnection(true)
	org, err := vcdClient.GetOrgByName(testConfig.SolutionAddOn.Org)
	if err != nil {
		t.Fatalf("error creating temporary VCD connection: %s", err)
	}

	catalog, err := org.GetCatalogByName(testConfig.SolutionAddOn.Catalog, false)
	if err != nil {
		t.Fatalf("error retrieving catalog: %s", err)
	}

	localAddOnPath, err := fetchCacheFile(catalog, testConfig.SolutionAddOn.AddOnImageDse, t)
	if err != nil {
		t.Fatalf("error finding Solution Add-On cache file: %s", err)
	}

	params := StringMap{
		"Org":     testConfig.SolutionAddOn.Org,
		"VdcName": testConfig.SolutionAddOn.Vdc,

		"TestName":          t.Name(),
		"CatalogName":       testConfig.SolutionAddOn.Catalog,
		"RoutedNetworkName": testConfig.SolutionAddOn.RoutedNetwork,
		"PublishToOrg":      testConfig.Cse.TenantOrg,
		"PublishToOrg2":     testConfig.Cse.SolutionsOrg,
		"AddonIsoPath":      localAddOnPath,

		// Feeding nested configuration map directly into template engine
		"DseSolution": testConfig.SolutionAddOn.DseSolutions,
	}
	testParamsNotEmpty(t, params)

	params["SkipBinary"] = "# skip-binary-test: Will use other step due to long runs" // leave only 1 biggest step to save time

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccDseStep1pre, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccDseStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3DS := templateFill(testAccDseStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3DS)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccDseStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["SkipBinary"] = "# Will run this step"
	params["FuncName"] = t.Name() + "step5"
	configText5DS := templateFill(testAccDseStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_solution_add_on_instance.dse14", "id"),
					resource.TestCheckResourceAttr("vcd_solution_add_on_instance_publish.public", "publish_to_all_tenants", "true"),
				),
			},
			{
				// Data Solutions sometimes lag to appear after Add-On is published. Resetting
				// connection cache and sleeping.
				PreConfig: func() { cachedVCDClients.reset(); time.Sleep(10 * time.Second) },
				Config:    configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_solution_add_on_instance.dse14", "id"),
					resource.TestCheckResourceAttr("vcd_solution_add_on_instance_publish.public", "publish_to_all_tenants", "true"),

					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.dso", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.mongodb-community", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.confluent-platform", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.mongodb", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.postgres", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.rabbit-mq", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_registry_configuration.mysql", "id"),
				),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_dse_registry_configuration.dso", "vcd_dse_registry_configuration.dso", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.mongodb-community", "vcd_dse_registry_configuration.mongodb-community", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.confluent-platform", "vcd_dse_registry_configuration.confluent-platform", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.mongodb", "vcd_dse_registry_configuration.mongodb", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.postgres", "vcd_dse_registry_configuration.postgres", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.rabbit-mq", "vcd_dse_registry_configuration.rabbit-mq", []string{"%", "use_default_value"}),
					resourceFieldsEqual("data.vcd_dse_registry_configuration.mysql", "vcd_dse_registry_configuration.mysql", []string{"%", "use_default_value"}),
				),
			},
			{
				ResourceName:            "vcd_dse_registry_configuration.dso",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "VCD Data Solutions",
				ImportStateVerifyIgnore: []string{"use_default_value"},
			},
			{
				ResourceName:            "vcd_dse_registry_configuration.mongodb-community",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "MongoDB Community",
				ImportStateVerifyIgnore: []string{"use_default_value"},
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.mongodb-community", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.confluent-platform", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.mongodb", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.postgres", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.rabbit-mq", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.mysql", "id"),
					resource.TestCheckResourceAttrSet("vcd_dse_solution_publish.mysqlorg2", "id"),
				),
			},
			{
				Config: configText5DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_dse_solution_publish.mongodb-community", "vcd_dse_solution_publish.mongodb-community", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.confluent-platform", "vcd_dse_solution_publish.confluent-platform", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.confluent-platform-org2", "vcd_dse_solution_publish.confluent-platform-org2", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.mongodb", "vcd_dse_solution_publish.mongodb", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.postgres", "vcd_dse_solution_publish.postgres", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.rabbit-mq", "vcd_dse_solution_publish.rabbit-mq", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.mysql", "vcd_dse_solution_publish.mysql", []string{"%", "confluent_license_key"}),
					resourceFieldsEqual("data.vcd_dse_solution_publish.mysqlorg2", "vcd_dse_solution_publish.mysqlorg2", []string{"%", "confluent_license_key"}),
				),
			},
			{
				ResourceName:      "vcd_dse_solution_publish.mongodb-community",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf(`%s.%s`, "MongoDB Community", params["PublishToOrg"].(string)),
			},
		},
	})
}

const testAccDseStep1pre = testAccSolutionAddonInstanceStep1 + testAccSolutionAddonInstancePublishAll
const testAccDseStep2 = testAccDseStep1pre + testAccDseRegistryConfig
const testAccDseStep3DS = testAccDseStep2 + testAccDseRegistryConfigDS

const testAccDseRegistryConfig = `
{{.SkipBinary}}
resource "vcd_dse_registry_configuration" "dso" {
  name               = "VCD Data Solutions"
  # Using default versions for packages
  use_default_value = true

  container_registry {
    host        = "test.host"
    description = "Test Host that does not work"
    username    = "testuser"
    password    = "invalidtestpassword"
  }

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "mongodb-community" {
  name             = "MongoDB Community"
  chart_repository = "{{ index .DseSolution "MongoDB Community" "chart_repository" }}"
  version          = "{{ index .DseSolution "MongoDB Community" "version" }}"
  package_name     = "{{ index .DseSolution "MongoDB Community" "package_name" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "confluent-platform" {
  name             = "Confluent Platform"
  chart_repository = "{{ index .DseSolution "Confluent Platform" "chart_repository" }}"
  version          = "{{ index .DseSolution "Confluent Platform" "version" }}"
  package_name     = "{{ index .DseSolution "Confluent Platform" "package_name" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "mongodb" {
  name             = "MongoDB"
  chart_repository = "{{ index .DseSolution "MongoDB" "chart_repository" }}"
  version          = "{{ index .DseSolution "MongoDB" "version" }}"
  package_name     = "{{ index .DseSolution "MongoDB" "package_name" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "postgres" {
  name               = "VMware SQL with Postgres"
  package_repository = "{{ index .DseSolution "VMware SQL with Postgres" "package_repository" }}"
  version            = "{{ index .DseSolution "VMware SQL with Postgres" "version" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "rabbit-mq" {
  name               = "VMware RabbitMQ"
  package_repository = "{{ index .DseSolution "VMware RabbitMQ" "package_repository" }}"
  version            = "{{ index .DseSolution "VMware RabbitMQ" "version" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}

resource "vcd_dse_registry_configuration" "mysql" {
  name               = "VMware SQL with MySQL"
  package_repository = "{{ index .DseSolution "VMware SQL with MySQL" "package_repository" }}"
  version            = "{{ index .DseSolution "VMware SQL with MySQL" "version" }}"

  depends_on = [vcd_solution_add_on_instance_publish.public]
}
`

const testAccDseRegistryConfigDS = `
data "vcd_dse_registry_configuration" "dso" {
  name = "VCD Data Solutions"

  depends_on = [vcd_dse_registry_configuration.dso]
}

data "vcd_dse_registry_configuration" "mongodb-community" {
  name = "MongoDB Community"

  depends_on = [vcd_dse_registry_configuration.mongodb-community]
}

data "vcd_dse_registry_configuration" "confluent-platform" {
  name = "Confluent Platform"

  depends_on = [vcd_dse_registry_configuration.confluent-platform]
}

data "vcd_dse_registry_configuration" "mongodb" {
  name = "MongoDB"

  depends_on = [vcd_dse_registry_configuration.mongodb]
}

data "vcd_dse_registry_configuration" "postgres" {
  name = "VMware SQL with Postgres"

  depends_on = [vcd_dse_registry_configuration.postgres]
}

data "vcd_dse_registry_configuration" "rabbit-mq" {
  name = "VMware RabbitMQ"

  depends_on = [vcd_dse_registry_configuration.rabbit-mq]
}

data "vcd_dse_registry_configuration" "mysql" {
  name = "VMware SQL with MySQL"

  depends_on = [vcd_dse_registry_configuration.mysql]
}
`

const testAccDseStep4 = testAccDseStep3DS + `
data "vcd_org" "tenant-org" {
  name = "{{.PublishToOrg}}"
}

data "vcd_org" "solutions-org" {
  name = "{{.PublishToOrg2}}"
}

resource "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = vcd_dse_registry_configuration.mongodb-community.id

  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "confluent-platform" {
  data_solution_id = vcd_dse_registry_configuration.confluent-platform.id

  confluent_license_type = "With License"
  confluent_license_key  = "Fake-key"
  
  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "confluent-platform-org2" {
  data_solution_id = vcd_dse_registry_configuration.confluent-platform.id

  confluent_license_type = "No License"
  
  org_id = data.vcd_org.solutions-org.id
}

resource "vcd_dse_solution_publish" "mongodb" {
  data_solution_id = vcd_dse_registry_configuration.mongodb.id
  
  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "postgres" {
  data_solution_id = vcd_dse_registry_configuration.postgres.id
  
  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "rabbit-mq" {
  data_solution_id = vcd_dse_registry_configuration.rabbit-mq.id
  
  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "mysql" {
  data_solution_id = vcd_dse_registry_configuration.mysql.id
  
  org_id = data.vcd_org.tenant-org.id
}

resource "vcd_dse_solution_publish" "mysqlorg2" {
  data_solution_id = vcd_dse_registry_configuration.mysql.id
  
  org_id = data.vcd_org.solutions-org.id
}
`

const testAccDseStep5DS = testAccDseStep4 + `
data "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = vcd_dse_registry_configuration.mongodb-community.id

  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.mongodb-community]
}

data "vcd_dse_solution_publish" "confluent-platform" {
  data_solution_id = vcd_dse_registry_configuration.confluent-platform.id

  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.confluent-platform]
}

data "vcd_dse_solution_publish" "confluent-platform-org2" {
  data_solution_id = vcd_dse_registry_configuration.confluent-platform.id
  
  org_id = data.vcd_org.solutions-org.id

  depends_on = [vcd_dse_solution_publish.confluent-platform-org2]
}

data "vcd_dse_solution_publish" "mongodb" {
  data_solution_id = vcd_dse_registry_configuration.mongodb.id
  
  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.mongodb]
}

data "vcd_dse_solution_publish" "postgres" {
  data_solution_id = vcd_dse_registry_configuration.postgres.id
  
  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.postgres]
}

data "vcd_dse_solution_publish" "rabbit-mq" {
  data_solution_id = vcd_dse_registry_configuration.rabbit-mq.id
  
  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.rabbit-mq]
}

data "vcd_dse_solution_publish" "mysql" {
  data_solution_id = vcd_dse_registry_configuration.mysql.id
  
  org_id = data.vcd_org.tenant-org.id

  depends_on = [vcd_dse_solution_publish.mysql]
}

data "vcd_dse_solution_publish" "mysqlorg2" {
  data_solution_id = vcd_dse_registry_configuration.mysql.id
  
  org_id = data.vcd_org.solutions-org.id

  depends_on = [vcd_dse_solution_publish.mysqlorg2]
}
`
