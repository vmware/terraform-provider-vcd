//go:build catalog || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdSubscribedCatalog(t *testing.T) {
	preTestChecks(t)

	var (
		publisherDescription  = "test publisher catalog"
		publisherCatalog      = "test-publisher"
		subscriberCatalog     = "test-subscriber"
		publisherOrg          = testConfig.VCD.Org
		subscriberOrg         = testConfig.VCD.Org + "-1"
		numberOfVappTemplates = 2
		numberOfMediaItems    = 3
		params                = StringMap{
			"PublisherOrg":            publisherOrg,
			"PublisherVdc":            testConfig.Nsxt.Vdc,
			"PublisherCatalog":        publisherCatalog,
			"PublisherDescription":    publisherDescription,
			"PublisherStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
			"Password":                "superUnknown",
			"SubscriberOrg":           subscriberOrg,
			"SubscriberCatalog":       subscriberCatalog,
			"VappTemplateBaseName":    "test-vt",
			"MediaItemBaseName":       "test-media",
			"MakeLocalCopy":           false,
			"OvaPath":                 testConfig.Ova.OvaPath,
			"MediaPath":               testConfig.Media.MediaPath,
			"NumberOfVappTemplates":   numberOfVappTemplates,
			"NumberOfMediaItems":      numberOfMediaItems,
			"Tags":                    "catalog subscribe",
			"FuncName":                t.Name(),
		}
	)

	configText := templateFill(testAccVcdPublisherCatalogCreation+
		testAccVcdPublisherCatalogItems, params)

	params["FuncName"] = t.Name() + "-subscriber"
	subscriberConfigText := templateFill(testAccVcdPublisherCatalogCreation+
		testAccVcdPublisherCatalogItems+
		testAccSubscribedCatalogCreation, params)

	params["FuncName"] = t.Name() + "-subscriber-update"
	subscriberConfigTextUpdate := templateFill(testAccVcdPublisherCatalogCreation+
		testAccVcdPublisherCatalogItems+
		testAccSubscribedCatalogUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION publisher: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION subscriber: %s", subscriberConfigText)
	debugPrintf("#[DEBUG] CONFIGURATION subscriber update: %s", subscriberConfigTextUpdate)

	resourcePublisher := "vcd_catalog." + publisherCatalog
	resourceSubscriber := "vcd_subscribed_catalog." + subscriberCatalog
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckCatalogDestroy(publisherOrg, publisherCatalog),
			testCheckCatalogDestroy(subscriberOrg, subscriberCatalog),
		),
		Steps: []resource.TestStep{
			{
				Config:       configText,
				ResourceName: resourcePublisher,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdCatalogExists(resourcePublisher),
					resource.TestCheckResourceAttr(resourcePublisher, "name", publisherCatalog),
					resource.TestCheckResourceAttr(resourcePublisher, "description", publisherDescription),
					resource.TestCheckResourceAttr(resourcePublisher, "publish_subscription_type", "PUBLISHED"),
					resource.TestMatchResourceAttr(resourcePublisher, "publish_subscription_url", regexp.MustCompile(`^https://\S+$`)),
				),
			},
			{
				Config:       subscriberConfigText,
				ResourceName: resourceSubscriber,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourcePublisher),
					testAccCheckVcdCatalogExists(resourceSubscriber),
					resource.TestCheckResourceAttr(resourceSubscriber, "name", resourceSubscriber),
					resource.TestCheckResourceAttr(resourceSubscriber, "description", publisherDescription),
					resource.TestCheckResourceAttr(
						resourcePublisher, "number_of_vapp_templates", fmt.Sprintf("%d", numberOfVappTemplates)),
					resource.TestCheckResourceAttr(
						resourcePublisher, "number_of_media", fmt.Sprintf("%d", numberOfMediaItems)),
					resource.TestCheckResourceAttr(
						resourceSubscriber, "number_of_vapp_templates", fmt.Sprintf("%d", numberOfVappTemplates)),
					resource.TestCheckResourceAttr(
						resourceSubscriber, "number_of_media", fmt.Sprintf("%d", numberOfMediaItems)),
				),
			},
		},
	})
	postTestChecks(t)
}

func testCheckCatalogDestroy(orgName, catalogName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_catalog" &&
				rs.Primary.Attributes["org"] != orgName &&
				rs.Primary.Attributes["name"] != catalogName {
				continue
			}

			adminOrg, err := conn.GetAdminOrg(orgName)
			if err != nil {
				return fmt.Errorf(errorRetrievingOrg, orgName+" and error: "+err.Error())
			}

			_, err = adminOrg.GetCatalogById(rs.Primary.ID, false)

			if err == nil {
				return fmt.Errorf("catalog %s still exists", catalogName)
			}
		}
		return nil
	}
}

const testAccVcdPublisherCatalogCreation = `
data "vcd_storage_profile" "storage_profile" {
  org  = "{{.PublisherOrg}}"
  vdc  = "{{.PublisherVdc}}"
  name = "{{.PublisherStorageProfile}}"
}

resource "vcd_catalog" "{{.PublisherCatalog}}" {
  org                = "{{.PublisherOrg}}"
  name               = "{{.PublisherCatalog}}"
  description        = "{{.PublisherDescription}}"
  storage_profile_id = data.vcd_storage_profile.storage_profile.id

  delete_force     = "true"
  delete_recursive = "true"

  publish_enabled               = "true"
  cache_enabled                 = "true"
  preserve_identity_information = "false"
  password                      = "{{.Password}}"
}
`

const testAccVcdPublisherCatalogItems = `
resource "vcd_catalog_item" "test_vt" {
  count = {{.NumberOfVappTemplates}}
  org     = "{{.PublisherOrg}}"
  catalog = vcd_catalog.{{.PublisherCatalog}}.name

  name                 = "{{.VappTemplateBaseName}}-${count.index}"
  description          = "test vapp template {{.VappTemplateBaseName}}-${count.index}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = 5
}

resource "vcd_catalog_media" "test_media" {
  count   = {{.NumberOfMediaItems}}
  org     = "{{.PublisherOrg}}"
  catalog = vcd_catalog.{{.PublisherCatalog}}.name

  name                 = "{{.MediaItemBaseName}}-${count.index}"
  description          = "test media item {{.MediaItemBaseName}}-${count.index}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = 5
}
`

const testAccSubscribedCatalogCreation = `
resource "vcd_subscribed_catalog" "{{.SubscriberCatalog}}" {
  org  = "{{.SubscriberOrg}}"
  name = "{{.SubscriberCatalog}}"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = vcd_catalog.{{.PublisherCatalog}}.publish_subscription_url
  make_local_copy       = {{.MakeLocalCopy}}
  subscription_password = "{{.Password}}"

  sync_on_refresh = true
}
`

const testAccSubscribedCatalogUpdate = `
resource "vcd_subscribed_catalog" "{{.SubscriberCatalog}}" {
  org  = "{{.SubscriberOrg}}"
  name = "{{.SubscriberCatalog}}"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = vcd_catalog.{{.PublisherCatalog}}.publish_subscription_url
  make_local_copy       = {{.MakeLocalCopy}}
  subscription_password = "{{.Password}}"

  sync_on_refresh = true
  sync_all        = true
}
`
