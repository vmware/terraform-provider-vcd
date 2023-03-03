//go:build catalog || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	checkTimeout = 10 * time.Second
	checkDelay   = 200 * time.Millisecond
)

func TestAccVcdVsphereSubscriber(t *testing.T) {
	preTestChecks(t)

	if testConfig.VCD.Catalog.VSphereSubscribedCatalog == "" {
		t.Skip("vSphereSubscribedCatalog was not defined")
	}

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		t.Skipf("This test tests VCD 10.4.0+ (API V37.0+) features. Skipping.")
	}

	subscriberCatalog := t.Name()
	subscriberOrg := testConfig.VCD.Org
	var params = StringMap{
		"SubscriberOrg":            subscriberOrg,
		"VsphereSubscriberCatalog": testConfig.VCD.Catalog.VSphereSubscribedCatalog,
		"SubscriberCatalog":        subscriberCatalog,
		"Tags":                     "catalog subscribe",
		"FuncName":                 t.Name(),
	}
	configText := templateFill(testAccSubscribedCatalogVSphere, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION subscriber: %s", configText)

	dataSourceVsphereSubscriber := "data.vcd_subscribed_catalog." + testConfig.VCD.Catalog.VSphereSubscribedCatalog
	resourceSubscriber := "vcd_subscribed_catalog." + subscriberCatalog
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckCatalogDestroy(subscriberOrg, subscriberCatalog),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceSubscriber),
					testCheckCatalogAndItemsExist(subscriberOrg, subscriberCatalog, true, 0, 0, 0),
					resource.TestCheckResourceAttr(resourceSubscriber, "name", subscriberCatalog),
					resource.TestCheckResourceAttrPair(resourceSubscriber, "subscription_url", dataSourceVsphereSubscriber, "subscription_url"),
					checkWithTimeout(checkTimeout, checkDelay,
						resource.TestCheckResourceAttrPair(resourceSubscriber, "subscription_url", dataSourceVsphereSubscriber, "subscription_url")),
				),
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdSubscribedCatalog(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		t.Skipf("This test tests VCD 10.4.0+ (API V37.0+) features. Skipping.")
	}

	var (
		publisherDescription  = "test publisher catalog"
		publisherCatalog      = "test-publisher"
		subscriberCatalog     = "test-subscriber"
		testVm                = "testVmSubscribed"
		publisherOrg          = testConfig.VCD.Org
		subscriberOrg         = testConfig.VCD.Org + "-1"
		subscriberVdc         = testConfig.Nsxt.Vdc + "-1"
		numberOfVappTemplates = 2
		numberOfMediaItems    = 3
		skipMessage           = "# skip-binary-test: not suitable for binary tests due to timing considerations"
	)

	var localCopyBehavior = map[bool]string{
		true:  "sync_catalog = true",
		false: "sync_all = true",
	}

	for makeLocalCopy, syncWhat := range localCopyBehavior {
		testName := fmt.Sprintf("%s-with-local-copy", t.Name())
		if !makeLocalCopy {
			testName = fmt.Sprintf("%s-no-local-copy", t.Name())
		}
		t.Run(fmt.Sprintf("make_local_copy=%v", makeLocalCopy), func(t *testing.T) {
			var params = StringMap{
				"ProviderVcdSystem":       providerVcdSystem,
				"ProviderVcdOrg1":         providerVcdOrg1,
				"ProviderVcdOrg2":         providerVcdOrg2,
				"SkipMessage":             skipMessage,
				"PublisherOrg":            publisherOrg,
				"PublisherVdc":            testConfig.Nsxt.Vdc,
				"PublisherCatalog":        publisherCatalog,
				"PublisherDescription":    publisherDescription,
				"PublisherStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
				"VmName":                  testVm,
				"Password":                "superUnknown",
				"SubscriberOrg":           subscriberOrg,
				"SubscriberCatalog":       subscriberCatalog,
				"SubscriberVdc":           subscriberVdc,
				"VappTemplateBaseName":    "test-vt",
				"MediaItemBaseName":       "test-media",
				"MakeLocalCopy":           makeLocalCopy,
				"SyncWhat":                syncWhat,
				"OvaPath":                 testConfig.Ova.OvaPath,
				"MediaPath":               testConfig.Media.MediaPath,
				"NumberOfVappTemplates":   numberOfVappTemplates,
				"NumberOfMediaItems":      numberOfMediaItems,
				"Tags":                    "catalog subscribe",
				"FuncName":                testName,
			}

			configText := templateFill(testAccVcdPublisherCatalogCreation+
				testAccVcdPublisherCatalogItems, params)

			params["FuncName"] = testName + "-subscriber"
			subscriberConfigText := templateFill(testAccVcdPublisherCatalogCreation+
				testAccVcdPublisherCatalogItems+
				testAccSubscribedCatalogCreation, params)

			params["FuncName"] = testName + "-subscriber-update"
			// Enable binary tests only for this stage, as the others would have timing problems
			params["SkipMessage"] = ""
			subscriberConfigTextUpdate := templateFill(testAccVcdPublisherCatalogCreation+
				testAccVcdPublisherCatalogItems+
				testAccSubscribedCatalogUpdate, params)

			params["FuncName"] = testName + "-subscriber-sync"
			params["SkipMessage"] = skipMessage
			subscriberConfigTextSync := templateFill(testAccVcdPublisherCatalogCreation+
				testAccVcdPublisherCatalogItems+
				testAccSubscribedCatalogUpdate+
				testAccSubscribedCatalogSync, params)

			if vcdShortTest {
				t.Skip(acceptanceTestsSkipped)
				return
			}
			debugPrintf("#[DEBUG] CONFIGURATION publisher: %s", configText)
			debugPrintf("#[DEBUG] CONFIGURATION subscriber: %s", subscriberConfigText)
			debugPrintf("#[DEBUG] CONFIGURATION subscriber update: %s", subscriberConfigTextUpdate)
			debugPrintf("#[DEBUG] CONFIGURATION subscriber sync: %s", subscriberConfigTextSync)

			resourcePublisher := "vcd_catalog." + publisherCatalog
			resourceSubscriber := "vcd_subscribed_catalog." + subscriberCatalog
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preRunChecks(t) },
				ProviderFactories: buildMultipleProviders(),
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
							testCheckCatalogAndItemsExist(publisherOrg, publisherCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							resource.TestCheckResourceAttr(resourcePublisher, "name", publisherCatalog),
							resource.TestCheckResourceAttr(resourcePublisher, "description", publisherDescription),
							resource.TestCheckResourceAttr(resourcePublisher, "publish_subscription_type", "PUBLISHED"),
							resource.TestMatchResourceAttr(resourcePublisher, "publish_subscription_url", regexp.MustCompile(`^https://\S+$`)),
						),
					},
					{
						Config:       subscriberConfigText,
						ResourceName: resourceSubscriber,
						Check: resource.ComposeAggregateTestCheckFunc(
							testCheckCatalogAndItemsExist(publisherOrg, publisherCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							testCheckCatalogAndItemsExist(subscriberOrg, subscriberCatalog, false, 0, 0, 0),
							resource.TestCheckResourceAttr(resourceSubscriber, "name", subscriberCatalog),
							resource.TestCheckResourceAttr(
								resourcePublisher, "number_of_vapp_templates", fmt.Sprintf("%d", numberOfVappTemplates)),
							resource.TestCheckResourceAttr(
								resourcePublisher, "number_of_media", fmt.Sprintf("%d", numberOfMediaItems)),
							resource.TestCheckResourceAttr(resourceSubscriber, "number_of_vapp_templates", "0"),
							resource.TestCheckResourceAttr(resourceSubscriber, "number_of_media", "0"),
						),
					},
					{
						Config:       subscriberConfigTextUpdate,
						ResourceName: resourceSubscriber,
						Check: resource.ComposeAggregateTestCheckFunc(
							testCheckCatalogAndItemsExist(publisherOrg, publisherCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							testCheckCatalogAndItemsExist(subscriberOrg, subscriberCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr(resourceSubscriber, "name", subscriberCatalog)),
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr(resourceSubscriber, "description", publisherDescription)),
						),
					},
					{
						PreConfig: func() {
							time.Sleep(10 * time.Second)
						},
						Config:       subscriberConfigTextSync,
						ResourceName: resourceSubscriber,
						Check: resource.ComposeAggregateTestCheckFunc(
							testCheckCatalogAndItemsExist(publisherOrg, publisherCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							testCheckCatalogAndItemsExist(subscriberOrg, subscriberCatalog, true, numberOfVappTemplates+numberOfMediaItems, numberOfVappTemplates, numberOfMediaItems),
							testAccCheckVcdStandaloneVmExists(testVm, "vcd_vm."+testVm, subscriberOrg, subscriberVdc),
							resource.TestCheckResourceAttr(resourceSubscriber, "name", subscriberCatalog),

							// A subscribed catalog gets its description and metadata from the publisher
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr(resourceSubscriber, "description", publisherDescription)),
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr(resourceSubscriber, "metadata.identity", "published catalog")),

							// Subscribed catalog items also get their metadata from the corresponding published items
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr("data.vcd_catalog_vapp_template.test-vt-1", "metadata.ancestors", fmt.Sprintf("%s.%s", publisherOrg, publisherCatalog))),
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr("data.vcd_catalog_media.test-media-1", "metadata.ancestors", fmt.Sprintf("%s.%s", publisherOrg, publisherCatalog))),
							checkWithTimeout(checkTimeout, checkDelay,
								resource.TestCheckResourceAttr("data.vcd_catalog_item.test-vt-1", "metadata.ancestors", fmt.Sprintf("%s.%s", publisherOrg, publisherCatalog))),

							// If these VM exist, it means that the corresponding vApp template and Media items are fully functional
							resource.TestCheckResourceAttr("vcd_vm."+testVm, "name", testVm),
							resource.TestCheckResourceAttr("vcd_vm."+testVm+"2", "name", testVm+"2"),
							resource.TestCheckResourceAttr("vcd_vm."+testVm+"3", "name", testVm+"3"),
						),
					},
					{
						ResourceName:      resourceSubscriber,
						ImportState:       true,
						ImportStateVerify: true,
						ImportStateIdFunc: importStateIdOrgObject(subscriberOrg, subscriberCatalog),
						// These fields can't be retrieved from catalog data
						ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive",
							"sync_catalog", "sync_all", "sync_on_refresh", "subscription_password",
							"cancel_failed_tasks", "store_tasks", "sync_all_vapp_templates",
							"sync_vapp_templates", "sync_all_media_items", "tasks_file_name",
							"sync_media_items",
						},
					},
				},
			})
		})
	}
	postTestChecks(t)
}

// testCheckCatalogAndItemsExist checks that a catalog exists, and optionally that it has as many items as expected
// * checkItems defines whether we count the items or not
// * expectedItems is the total number of catalog items (includes both vApp templates and media items)
// * expectedTemplates is the number of vApp templates
// expectedMedia is the number of Media
func testCheckCatalogAndItemsExist(orgName, catalogName string, checkItems bool, expectedItems, expectedTemplates, expectedMedia int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if testAccProvider == nil || testAccProvider.Meta() == nil {
			return fmt.Errorf("testAccProvider is not initialized")
		}
		conn := testAccProvider.Meta().(*VCDClient)
		catalog, err := conn.Client.GetAdminCatalogByName(orgName, catalogName)
		if err != nil {
			return fmt.Errorf("error retrieving catalog %s/%s: %s", orgName, catalogName, err)
		}
		if !checkItems {
			return nil
		}

		if catalog.AdminCatalog.Tasks != nil {
			err = catalog.WaitForTasks()
			if err != nil {
				return err
			}
		}

		items, err := catalog.QueryCatalogItemList()
		if err != nil {
			return fmt.Errorf("error retrieving catalog item list: %s", err)
		}
		vappTemplates, err := catalog.QueryVappTemplateList()
		if err != nil {
			return fmt.Errorf("error retrieving vApp templates list: %s", err)
		}
		mediaItems, err := catalog.QueryMediaList()
		if err != nil {
			return fmt.Errorf("error retrieving media items list: %s", err)
		}
		if len(items) != expectedItems {
			return fmt.Errorf("catalog '%s' -expected %d items - found %d", catalogName, expectedItems, len(items))
		}
		if len(vappTemplates) != expectedTemplates {
			return fmt.Errorf("catalog '%s' -expected %d vApp templates - found %d", catalogName, expectedTemplates, len(vappTemplates))
		}
		if len(mediaItems) != expectedMedia {
			return fmt.Errorf("catalog '%s' -expected %d media items - found %d", catalogName, expectedMedia, len(mediaItems))
		}
		return nil
	}
}

func testCheckCatalogDestroy(orgName, catalogName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		_, err := conn.Client.GetAdminCatalogByName(orgName, catalogName)
		if err == nil {
			return fmt.Errorf("catalog %s still exists", catalogName)
		}
		return nil
	}
}

// checkWithTimeout runs the wanted check function "f" until it succeeds or the given "timeout" expires.
// The check pauses for the amount of time specified by "delay"
func checkWithTimeout(timeout, delay time.Duration, f resource.TestCheckFunc) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		start := time.Now()
		for time.Since(start) < timeout {
			time.Sleep(delay)
			err := f(state)
			if err == nil {
				return nil
			}
		}
		return f(state)
	}
}

// testAccVcdPublisherCatalogCreation contains the creation of the publishing catalog
const testAccVcdPublisherCatalogCreation = `
{{.SkipMessage}}
data "vcd_storage_profile" "storage_profile" {
  provider = {{.ProviderVcdSystem}}

  org  = "{{.PublisherOrg}}"
  vdc  = "{{.PublisherVdc}}"
  name = "{{.PublisherStorageProfile}}"
}

resource "vcd_catalog" "{{.PublisherCatalog}}" {
  provider = {{.ProviderVcdOrg1}}

  org                = "{{.PublisherOrg}}"
  name               = "{{.PublisherCatalog}}"
  description        = "{{.PublisherDescription}}"
  storage_profile_id = data.vcd_storage_profile.storage_profile.id

  delete_force     = "true"
  delete_recursive = "true"

  metadata = {
    identity = "published catalog"
    parent   = "{{.PublisherOrg}}"
  }
  publish_enabled               = "true"
  cache_enabled                 = "true"
  preserve_identity_information = "false"
  password                      = "{{.Password}}"
}
`

// testAccVcdPublisherCatalogItems creates all items that depend on the publishing catalog
const testAccVcdPublisherCatalogItems = `
resource "vcd_catalog_vapp_template" "{{.VappTemplateBaseName}}" {
  provider = {{.ProviderVcdOrg1}}

  count      = {{.NumberOfVappTemplates}}
  org        = "{{.PublisherOrg}}"
  catalog_id = vcd_catalog.{{.PublisherCatalog}}.id

  name              = "{{.VappTemplateBaseName}}-${count.index}"
  description       = "test vapp template {{.VappTemplateBaseName}}-${count.index}"
  ova_path          = "{{.OvaPath}}"
  upload_piece_size = 5

  metadata = {
    identity  = "published vApp template {{.VappTemplateBaseName}}-${count.index}"
    ancestors = "{{.PublisherOrg}}.{{.PublisherCatalog}}"
  }
}

resource "vcd_catalog_media" "{{.MediaItemBaseName}}" {
  provider = {{.ProviderVcdOrg1}}

  count   = {{.NumberOfMediaItems}}
  org     = "{{.PublisherOrg}}"
  catalog = vcd_catalog.{{.PublisherCatalog}}.name

  name              = "{{.MediaItemBaseName}}-${count.index}"
  description       = "test media item {{.MediaItemBaseName}}-${count.index}"
  media_path        = "{{.MediaPath}}"
  upload_piece_size = 5

  metadata = {
    identity  = "published media item {{.MediaItemBaseName}}-${count.index}"
    ancestors = "{{.PublisherOrg}}.{{.PublisherCatalog}}"
  }
}
`

// testAccSubscribedCatalogCreation creates the subscribed catalog (in a different Org)
const testAccSubscribedCatalogCreation = `
resource "vcd_subscribed_catalog" "{{.SubscriberCatalog}}" {
  provider = {{.ProviderVcdOrg2}}

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

// testAccSubscribedCatalogUpdate adds parameters to the subscribed catalog to handle synchronisation
const testAccSubscribedCatalogUpdate = `
resource "vcd_subscribed_catalog" "{{.SubscriberCatalog}}" {
  provider = {{.ProviderVcdOrg2}}

  org  = "{{.SubscriberOrg}}"
  name = "{{.SubscriberCatalog}}"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = vcd_catalog.{{.PublisherCatalog}}.publish_subscription_url
  make_local_copy       = {{.MakeLocalCopy}}
  subscription_password = "{{.Password}}"

  sync_on_refresh = true
  {{.SyncWhat}}
}
`

// testAccSubscribedCatalogSync adds data sources for a few catalog items
// and a VM that uses one of the subscribed items
const testAccSubscribedCatalogSync = `
data "vcd_catalog_item" "{{.VappTemplateBaseName}}-1" {
  provider = {{.ProviderVcdOrg2}}

  org     = "{{.SubscriberOrg}}"
  catalog = vcd_subscribed_catalog.{{.SubscriberCatalog}}.name
  name    = "{{.VappTemplateBaseName}}-1"
}

data "vcd_catalog_vapp_template" "{{.VappTemplateBaseName}}-1" {
  provider = {{.ProviderVcdOrg2}}

  org        = "{{.SubscriberOrg}}"
  catalog_id = vcd_subscribed_catalog.{{.SubscriberCatalog}}.id
  name       = "{{.VappTemplateBaseName}}-1"
}

data "vcd_catalog_media" "{{.MediaItemBaseName}}-1" {
  provider = {{.ProviderVcdOrg2}}

  org     = "{{.SubscriberOrg}}"
  catalog = vcd_subscribed_catalog.{{.SubscriberCatalog}}.name
  name    = "{{.MediaItemBaseName}}-1"
}

resource "vcd_vm" "{{.VmName}}" {
  provider = {{.ProviderVcdOrg2}}

  org           = "{{.SubscriberOrg}}"
  vdc           = "{{.SubscriberVdc}}"
  name          = "{{.VmName}}"
  catalog_name  = vcd_subscribed_catalog.{{.SubscriberCatalog}}.name
  template_name = data.vcd_catalog_item.{{.VappTemplateBaseName}}-1.name
  description   = "test standalone VM"
  power_on      = false
}

resource "vcd_vm" "{{.VmName}}2" {
  provider = {{.ProviderVcdOrg2}}

  org              = "{{.SubscriberOrg}}"
  vdc              = "{{.SubscriberVdc}}"
  name             = "{{.VmName}}2"
  vapp_template_id = data.vcd_catalog_vapp_template.{{.VappTemplateBaseName}}-1.id
  description      = "test standalone VM 2"
  power_on         = false
}

resource "vcd_vm" "{{.VmName}}3" {
  provider = {{.ProviderVcdOrg2}}

  org              = "{{.SubscriberOrg}}"
  vdc              = "{{.SubscriberVdc}}"
  name             = "{{.VmName}}3"
  boot_image_id    = data.vcd_catalog_media.{{.MediaItemBaseName}}-1.id
  description      = "test standalone VM 3"
  computer_name    = "standalone"
  cpus             = 1
  memory           = 1024
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  power_on         = false
}
`

const testAccSubscribedCatalogVSphere = `
data "vcd_subscribed_catalog" "{{.VsphereSubscriberCatalog}}" {
  org  = "{{.SubscriberOrg}}"
  name = "{{.VsphereSubscriberCatalog}}"
}

resource "vcd_subscribed_catalog" "{{.SubscriberCatalog}}" {
  org  = "{{.SubscriberOrg}}"
  name = "{{.SubscriberCatalog}}"

  delete_force     = true
  delete_recursive = true

  subscription_url = data.vcd_subscribed_catalog.{{.VsphereSubscriberCatalog}}.subscription_url
  make_local_copy  = true
  sync_on_refresh  = true
}
`
