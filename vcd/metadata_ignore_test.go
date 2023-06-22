//go:build (functional || catalog || vapp || network || org || vm || vdc || disk || standaloneVm || nsxt || ALL) && !skipLong

package vcd

import (
	"fmt"
	"testing"
)

func TestAccVcdOrgMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", id, err)
		}
		return adminOrg, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdOrgMetadata, "vcd_org.test-org",
		testAccCheckVcdOrgMetadataDatasource, "data.vcd_org.test-org-ds",
		"org", getObjectById, nil)
}

func TestAccVcdVdcMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetAdminVDCById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", id, err)
		}
		return vdc, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVdcMetadata, "vcd_org_vdc.test-vdc",
		testAccCheckVcdVdcMetadataDatasource, "data.vcd_org_vdc.test-vdc-ds",
		"vdc", getObjectById, StringMap{
			"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
			"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}

func TestAccVcdCatalogItemMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", testConfig.VCD.Catalog.NsxtBackedCatalogName, err)
		}
		catalogItem, err := catalog.GetCatalogItemById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog Item '%s': %s", id, err)
		}
		return catalogItem, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogItemMetadata, "vcd_catalog_item.test-catalog-item",
		testAccCheckVcdCatalogItemMetadataDatasource, "data.vcd_catalog_item.test-catalog-item-ds",
		"catalogItem", getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}

func TestAccVcdCatalogMediaMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", testConfig.VCD.Catalog.NsxtBackedCatalogName, err)
		}
		media, err := catalog.GetMediaById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Media '%s': %s", id, err)
		}
		return media, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogMediaMetadata, "vcd_catalog_media.test-catalog-media",
		testAccCheckVcdCatalogMediaMetadataDatasource, "data.vcd_catalog_media.test-catalog-media-ds",
		"media", getObjectById, StringMap{
			"Catalog":   testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"MediaPath": testConfig.Media.MediaPath,
		})
}

func TestAccVcdCatalogMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetAdminCatalogById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", id, err)
		}
		return catalog, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogMetadata, "vcd_catalog.test-catalog",
		testAccCheckVcdCatalogMetadataDatasource, "data.vcd_catalog.test-catalog-ds",
		"catalog", getObjectById, nil)
}

func TestAccVcdCatalogVAppTemplateMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", testConfig.VCD.Catalog.NsxtBackedCatalogName, err)
		}
		media, err := catalog.GetVAppTemplateById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp Template '%s': %s", id, err)
		}
		return media, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogVAppTemplateMetadata, "vcd_catalog_vapp_template.test-catalog-vapp-template",
		testAccCheckVcdCatalogVAppTemplateMetadataDatasource, "data.vcd_catalog_vapp_template.test-catalog-vapp-template-ds",
		"vAppTemplate", getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}

func TestAccVcdIndependentDiskMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		disk, err := vdc.GetDiskById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Independent Disk '%s': %s", id, err)
		}
		return disk, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdIndependentDiskMetadata, "vcd_independent_disk.test-independent-disk",
		testAccCheckVcdIndependentDiskMetadataDatasource, "data.vcd_independent_disk.test-independent-disk-ds",
		"disk", getObjectById, StringMap{
			"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}

func TestAccVcdIsolatedNetworkV2MetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		network, err := vdc.GetOpenApiOrgVdcNetworkById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Isolated Network V2 '%s': %s", id, err)
		}
		return network, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdIsolatedNetworkV2Metadata, "vcd_network_isolated_v2.test-network-isolated-v2",
		testAccCheckVcdIsolatedNetworkV2MetadataDatasource, "data.vcd_network_isolated_v2.test-network-isolated-v2-ds",
		"network", getObjectById, nil)
}

func TestAccVcdRoutedNetworkV2MetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		network, err := vdc.GetOpenApiOrgVdcNetworkById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Routed Network V2 '%s': %s", id, err)
		}
		return network, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdRoutedNetworkV2Metadata, "vcd_network_routed_v2.test-network-routed-v2",
		testAccCheckVcdRoutedNetworkV2MetadataDatasource, "data.vcd_network_routed_v2.test-network-routed-v2-ds",
		"network", getObjectById, StringMap{
			"EdgeGateway": testConfig.Nsxt.EdgeGateway,
		})
}

func TestAccVcdDirectNetworkMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.VCD.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.VCD.Vdc, err)
		}
		network, err := vdc.GetOrgVdcNetworkById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Direct Network '%s': %s", id, err)
		}
		return network, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdDirectNetworkMetadata, "vcd_network_direct.test-network-direct",
		testAccCheckVcdDirectNetworkMetadataDatasource, "data.vcd_network_direct.test-network-direct-ds",
		"network", getObjectById, StringMap{
			"ExternalNetwork": testConfig.Networking.ExternalNetwork,
			"Vdc":             testConfig.VCD.Vdc,
		})
}

func TestAccVcdIsolatedNetworkMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.VCD.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.VCD.Vdc, err)
		}
		network, err := vdc.GetOrgVdcNetworkById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Isolated Network '%s': %s", id, err)
		}
		return network, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdIsolatedNetworkMetadata, "vcd_network_isolated.test-network-isolated",
		testAccCheckVcdIsolatedNetworkMetadataDatasource, "data.vcd_network_isolated.test-network-isolated-ds",
		"network", getObjectById, StringMap{
			"Vdc": testConfig.VCD.Vdc,
		})
}

func TestAccVcdRoutedNetworkMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.VCD.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.VCD.Vdc, err)
		}
		network, err := vdc.GetOrgVdcNetworkById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Routed Network '%s': %s", id, err)
		}
		return network, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdRoutedNetworkMetadata, "vcd_network_routed.test-network-routed",
		testAccCheckVcdRoutedNetworkMetadataDatasource, "data.vcd_network_routed.test-network-routed-ds",
		"network", getObjectById, StringMap{
			"Vdc":         testConfig.VCD.Vdc,
			"EdgeGateway": testConfig.Networking.EdgeGateway,
		})
}

func TestAccVcdVAppMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		vApp, err := vdc.GetVAppById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp '%s': %s", id, err)
		}
		return vApp, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVAppMetadata, "vcd_vapp.test-vapp",
		testAccCheckVcdVAppMetadataDatasource, "data.vcd_vapp.test-vapp-ds",
		"vApp", getObjectById, nil)
}

func TestAccVcdVAppVmMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		vApp, err := vdc.GetVAppByName(t.Name(), true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp '%s': %s", t.Name(), err)
		}
		vm, err := vApp.GetVMById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VM '%s': %s", id, err)
		}
		return vm, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVAppVmMetadata, "vcd_vapp_vm.test-vapp-vm",
		testAccCheckVcdVAppVmMetadataDatasource, "data.vcd_vapp_vm.test-vapp-vm-ds",
		"vApp", getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}

func TestAccVcdVmMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		vApp, err := vdc.GetVAppByName(t.Name(), true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp '%s': %s", t.Name(), err)
		}
		vm, err := vApp.GetVMById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VM '%s': %s", id, err)
		}
		return vm, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVmMetadata, "vcd_vm.test-vm",
		testAccCheckVcdVmMetadataDatasource, "data.vcd_vm.test-vm-ds",
		"vApp", getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"Media":   testConfig.Media.NsxtBackedMediaName,
		})
}
