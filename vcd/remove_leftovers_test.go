package vcd

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// This file contains routines that clean up the test suite after failed tests

// entityDef is the definition of an entity (to be either deleted or kept)
// with an optional comment
type entityDef struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Comment string `json:"comment,omitempty"`
}

// entityList is a collection of entityDef
type entityList []entityDef

// doNotDelete contains a list of entities that should not be deleted,
// despite having a name that starts with `Test` or `test`
// Use only one of the following type identifiers:
// vcd_org
// vcd_provider_vdc
// vcd_org_vdc
// vcd_catalog
// vcd_catalog_media
// vcd_catalog_vapp_template
// vcd_vapp
// vcd_vm
// vcd_network (for any kind of org network)
var doNotDelete = entityList{
	{Type: "vcd_catalog_media", Name: "test_media", Comment: "loaded with provisioning"},
	{Type: "vcd_catalog_media", Name: "test_media_nsxt", Comment: "loaded with provisioning"},
	{Type: "vcd_vapp", Name: "TestVapp", Comment: "loaded with provisioning"},
	{Type: "vcd_vapp", Name: "Test_EmptyVmVapp1", Comment: "created by test, but to be preserved"},
	{Type: "vcd_vapp", Name: "Test_EmptyVmVapp2", Comment: "created by test, but to be preserved"},
	{Type: "vcd_vapp", Name: "Test_EmptyVmVapp3", Comment: "created by test, but to be preserved"},
}

// alsoDelete contains a list of entities that should be removed , in addition to the ones
// found by name matching
// Add to this list if you ever get an entity left behind by a test
// Note: these names were captured running 'go run get_test_resource_names.go' in ./vcd/test-artifacts
var alsoDelete = entityList{
	{Type: "vcd_catalog", Name: "Catalog-AC-0", Comment: "from vcd.TestAccVcdCatalogAccessControl-update.tf: Catalog-AC-0"},
	{Type: "vcd_catalog", Name: "Catalog-AC-1", Comment: "from vcd.TestAccVcdCatalogAccessControl-update.tf: Catalog-AC-1"},
	{Type: "vcd_catalog", Name: "Catalog-AC-2", Comment: "from vcd.TestAccVcdCatalogAccessControl-update.tf: Catalog-AC-2"},
	{Type: "vcd_catalog", Name: "Catalog-AC-3", Comment: "from vcd.TestAccVcdCatalogAccessControl-update.tf: Catalog-AC-3"},
	{Type: "vcd_catalog", Name: "Catalog-AC-4", Comment: "from vcd.TestAccVcdCatalogAccessControl-update.tf: Catalog-AC-4"},
	{Type: "vcd_org", Name: "datacloud-clone", Comment: "from vcd.TestAccVcdDatasourceOrg.tf: datacloud-clone"},
	{Type: "vcd_network", Name: "direct-net", Comment: "from vcd.TestAccVcdExternalNetworkV2NsxtSegmentIntegration-step2.tf: direct-net"},
	{Type: "vcd_network", Name: "nsxt-routed-test-initial", Comment: "from vcd.TestAccVcdNetworkRoutedV2Nsxt.tf: nsxt-routed-test-initial"},
	{Type: "vcd_org_vdc", Name: "newVdc", Comment: "from vcd.TestAccVcdNsxtEdgeGatewayVdcUpdateFailsstep2.tf: newVdc"},
	{Type: "vcd_network", Name: "nsxt-routed-0", Comment: "from vcd.TestAccVcdNsxtSecurityGroup-step2.tf: nsxt-backed (nsxt-routed-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "nsxt-routed-1", Comment: "from vcd.TestAccVcdNsxtSecurityGroup-step2.tf: nsxt-backed (nsxt-routed-${count.index} count = 2)"},
	{Type: "vcd_vm", Name: "standalone-VM", Comment: "from vcd.TestAccVcdNsxtSecurityGroup-step2.tf: standalone-VM"},
	{Type: "vcd_vapp", Name: "web", Comment: "from vcd.TestAccVcdNsxtSecurityGroup-step2.tf: web"},
	{Type: "vcd_network", Name: "nsxt-routed-vdcGroup-0", Comment: "from vcd.TestAccVcdNsxtSecurityGroupOwnerVdcGroup-step2.tf: nsxt-backed-vdc-group (nsxt-routed-vdcGroup-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "nsxt-routed-vdcGroup-1", Comment: "from vcd.TestAccVcdNsxtSecurityGroupOwnerVdcGroup-step2.tf: nsxt-backed-vdc-group (nsxt-routed-vdcGroup-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "multinic-net", Comment: "from vcd.TestAccVcdNsxtStandaloneEmptyVm.tf: multinic-net"},
	{Type: "vcd_network", Name: "multinic-net2", Comment: "from vcd.TestAccVcdNsxtStandaloneEmptyVm.tf: multinic-net2"},
	{Type: "vcd_network", Name: "dhcp-relay-0", Comment: "from vcd.TestAccVcdNsxvDhcpRelay-step1.tf: test-routed (dhcp-relay-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "dhcp-relay-1", Comment: "from vcd.TestAccVcdNsxvDhcpRelay-step1.tf: test-routed (dhcp-relay-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "firewall-test-0", Comment: "from vcd.TestAccVcdNsxvEdgeFirewallRule-step3.tf: test-routed (firewall-test-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "firewall-test-1", Comment: "from vcd.TestAccVcdNsxvEdgeFirewallRule-step3.tf: test-routed (firewall-test-${count.index} count = 2)"},
	{Type: "vcd_network", Name: "fw-routed-net", Comment: "from vcd.TestAccVcdNsxvEdgeFirewallRuleVms-step1.tf: fw-routed-net"},
	{Type: "vcd_vapp", Name: "fw-test", Comment: "from vcd.TestAccVcdNsxvEdgeFirewallRuleVms-step1.tf: fw-test"},
	{Type: "vcd_network", Name: "nsxt-routed-dhcp", Comment: "from vcd.TestAccVcdOpenApiDhcpNsxtRouted-step1.tf: nsxt-routed-dhcp"},
	{Type: "vcd_org", Name: "org1", Comment: "from vcd.TestAccVcdOrgFull_org1.tf: org1"},
	{Type: "vcd_org", Name: "org2", Comment: "from vcd.TestAccVcdOrgFull_org2.tf: org2"},
	{Type: "vcd_org", Name: "org3", Comment: "from vcd.TestAccVcdOrgFull_org3.tf: org3"},
	{Type: "vcd_org", Name: "org4", Comment: "from vcd.TestAccVcdOrgFull_org4.tf: org4"},
	{Type: "vcd_org", Name: "org5", Comment: "from vcd.TestAccVcdOrgFull_org5.tf: org5"},
	{Type: "vcd_catalog_vapp_template", Name: "4cpu-4cores", Comment: "from vcd.TestAccVcdStandaloneVmShrinkCpu.tf: 4cpu-4cores"},
	{Type: "vcd_vapp", Name: "Vapp-AC-0", Comment: "from vcd.TestAccVcdVappAccessControl-update.tf: Vapp-AC-0"},
	{Type: "vcd_vapp", Name: "Vapp-AC-1", Comment: "from vcd.TestAccVcdVappAccessControl-update.tf: Vapp-AC-1"},
	{Type: "vcd_vapp", Name: "Vapp-AC-2", Comment: "from vcd.TestAccVcdVappAccessControl-update.tf: Vapp-AC-2"},
	{Type: "vcd_vapp", Name: "Vapp-AC-3", Comment: "from vcd.TestAccVcdVappAccessControl-update.tf: Vapp-AC-3"},
	{Type: "vcd_org_vdc", Name: "ForInternalDiskTest", Comment: "from vcd.TestAccVcdVmInternalDisk-CreateALl.tf: ForInternalDiskTest"},
}

// isTest is a regular expression that tells if an entity needs to be deleted
var isTest = regexp.MustCompile(`^[Tt]est`)

// alwaysShow lists the resources that will always be shown
var alwaysShow = []string{"vcd_provider_vdc", "vcd_org", "vcd_catalog", "vcd_org_vdc", "vcd_nsxt_alb_controller"}

func removeLeftovers(govcdClient *govcd.VCDClient, verbose bool) error {
	if verbose {
		fmt.Printf("Start leftovers removal\n")
	}

	// NSX-T ALB configuration Hierarchical cleanup is separate from main hierarchy as even if the
	// Org is going to be deleted - NSX-T ALB configuration must be cleaned up first
	// Only System user can control ALB resources
	if govcdClient.Client.IsSysAdmin {
		err := removeLeftoversNsxtAlb(govcdClient, verbose)
		if err != nil {
			return fmt.Errorf("error removing NSX-T ALB leftovers: %s", err)
		}
	}

	// --------------------------------------------------------------
	// Provider VDCs
	// --------------------------------------------------------------
	if govcdClient.Client.IsSysAdmin {
		providerVdcs, err := govcdClient.QueryProviderVdcs()
		if err != nil {
			return fmt.Errorf("error retrieving provider VDCs: %s", err)
		}
		for _, pvdcRec := range providerVdcs {
			pvdc, err := govcdClient.GetProviderVdcExtendedByName(pvdcRec.Name)
			if err != nil {
				return fmt.Errorf("error retrieving provider VDC '%s': %s", pvdcRec.Name, err)
			}
			tobeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, pvdcRec.Name, "vcd_provider_vdc", 0, verbose)
			if tobeDeleted {
				fmt.Printf("\t REMOVING Provider VDC %s\n", pvdcRec.Name)
				err = pvdc.Disable()
				if err != nil {
					return fmt.Errorf("error disabling provider VDC '%s': %s", pvdcRec.Name, err)
				}
				task, err := pvdc.Delete()
				if err != nil {
					return fmt.Errorf("error deleting provider VDC '%s': %s", pvdcRec.Name, err)
				}
				err = task.WaitTaskCompletion()
				if err != nil {
					return fmt.Errorf("error finishing deletion of provider VDC '%s': %s", pvdcRec.Name, err)
				}
			}
		}
	}
	// traverses the VCD hierarchy, starting at the Org level
	orgs, err := govcdClient.GetOrgList()
	if err != nil {
		return fmt.Errorf("error retrieving Orgs list: %s", err)
	}
	// --------------------------------------------------------------
	// organizations
	// --------------------------------------------------------------
	for _, orgRef := range orgs.Org {
		org, err := govcdClient.GetOrgById("urn:vcloud:org:" + extractUuid(orgRef.HREF))
		if err != nil {
			return fmt.Errorf("error retrieving org %s: %s", orgRef.Name, err)
		}
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, orgRef.Name, "vcd_org", 0, verbose)
		if toBeDeleted {
			fmt.Printf("\t REMOVING org %s\n", org.Org.Name)
			adminOrg, err := govcdClient.GetAdminOrgById("urn:vcloud:org:" + extractUuid(orgRef.HREF))
			if err != nil {
				return fmt.Errorf("error retrieving org %s: %s", orgRef.Name, err)
			}
			err = adminOrg.Delete(true, true)
			if err != nil {
				return fmt.Errorf("error removing org %s: %s", orgRef.Name, err)
			}
			continue
		}
		// --------------------------------------------------------------
		// catalogs
		// --------------------------------------------------------------

		catalogs, err := org.QueryCatalogList()
		if err != nil {
			return fmt.Errorf("error retrieving catalog list: %s", err)
		}
		for _, catRec := range catalogs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, catRec.Name, "catalog", 1, verbose)
			catalog, err := org.GetCatalogByHref(catRec.HREF)
			if err != nil {
				return fmt.Errorf("error retrieving catalog '%s': %s", catRec.Name, err)
			}
			if toBeDeleted {
				fmt.Printf("\t\t REMOVING catalog %s/%s\n", org.Org.Name, catalog.Catalog.Name)
				err = catalog.Delete(true, true)
				if err != nil {
					return fmt.Errorf("error deleting catalog '%s': %s", catRec.Name, err)
				}
				continue
			}
			// --------------------------------------------------------------
			// vApp templates
			// --------------------------------------------------------------
			templates, err := catalog.QueryVappTemplateList()
			if err != nil {
				return fmt.Errorf("error retrieving catalog '%s' vApp template list: %s", catalog.Catalog.Name, err)
			}
			for _, templateRec := range templates {
				toBeDeleted = shouldDeleteEntity(alsoDelete, doNotDelete, templateRec.Name, "vcd_catalog_vapp_template", 2, verbose)
				if toBeDeleted {
					template, err := catalog.GetVappTemplateByHref(templateRec.HREF)
					if err != nil {
						return fmt.Errorf("error retrieving vapp template '%s': %s", templateRec.Name, err)
					}
					fmt.Printf("\t\t REMOVING vApp template %s/%s\n", catalog.Catalog.Name, template.VAppTemplate.Name)
					err = template.Delete()
					if err != nil {
						return fmt.Errorf("error deleting vApp template '%s': %s", templateRec.Name, err)
					}
				}
			}
			// --------------------------------------------------------------
			// media items
			// --------------------------------------------------------------
			mediaItems, err := catalog.QueryMediaList()
			if err != nil {
				return fmt.Errorf("error retrieving catalog '%s' media items list: %s", catalog.Catalog.Name, err)
			}
			for _, mediaRec := range mediaItems {
				toBeDeleted = shouldDeleteEntity(alsoDelete, doNotDelete, mediaRec.Name, "vcd_catalog_media", 2, verbose)
				if toBeDeleted {
					err = deleteMediaItem(catalog, mediaRec)
					if err != nil {
						return fmt.Errorf("error deleting media item '%s': %s", mediaRec.Name, err)
					}
				}
			}
		}
		// --------------------------------------------------------------
		// VDCs
		// --------------------------------------------------------------
		vdcs, err := org.QueryOrgVdcList()
		if err != nil {
			return fmt.Errorf("error retrieving VDC list: %s", err)
		}
		for _, vdcRec := range vdcs {
			vdc, err := org.GetVDCByName(vdcRec.Name, false)
			if err != nil {
				return fmt.Errorf("error retrieving VDC %s: %s", vdcRec.Name, err)
			}
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, vdc.Vdc.Name, "vcd_org_vdc", 1, verbose)
			if toBeDeleted {
				err = deleteVdc(org, vdc)
				if err != nil {
					return fmt.Errorf("error deleting VDC '%s': %s", vdc.Vdc.Name, err)
				}
				continue
			}
			// --------------------------------------------------------------
			// vApps
			// --------------------------------------------------------------
			vapps := vdc.GetVappList()
			for _, vappRef := range vapps {
				toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, vappRef.Name, "vcd_vapp", 2, verbose)
				if toBeDeleted {
					err = deleteVapp(vdc, vappRef)
					if err != nil {
						return fmt.Errorf("error deleting vApp %s: %s", vappRef.Name, err)
					}
				}
			}
			// --------------------------------------------------------------
			// VMs
			// --------------------------------------------------------------
			vms, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
			if err != nil {
				return fmt.Errorf("error retrieving VM list: %s", err)
			}
			for _, vmRec := range vms {
				// If not a standalone VM, we'll skip it, as it should be handled (or skipped) by vApp deletion
				if !vmRec.AutoNature {
					continue
				}
				toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, vmRec.Name, "vcd_vm", 2, verbose)
				if toBeDeleted {
					vm, err := govcdClient.Client.GetVMByHref(vmRec.HREF)
					if err != nil {
						return fmt.Errorf("error retrieving VM %s: %s", vmRec.Name, err)
					}
					fmt.Printf("\t\t REMOVING VM %s/%s\n", vdc.Vdc.Name, vm.VM.Name)
					err = vm.Delete()
					if err != nil {
						return fmt.Errorf("error deleting VM %s: %s", vmRec.Name, err)
					}
				}

			}
			// --------------------------------------------------------------
			// Networks
			// --------------------------------------------------------------
			networks, err := vdc.GetAllOpenApiOrgVdcNetworks(nil)
			if err != nil {
				return fmt.Errorf("error retrieving Org VDC network list: %s", err)
			}
			for _, netRef := range networks {
				toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, netRef.OpenApiOrgVdcNetwork.Name, "vcd_network", 2, verbose)
				if toBeDeleted {
					err = netRef.Delete()
					if err != nil {
						return fmt.Errorf("error deleting Org VDC '%s': %s", netRef.OpenApiOrgVdcNetwork.Name, err)
					}
				}
			}

			// --------------------------------------------------------------
			// NSX-T Edge Gateways
			// --------------------------------------------------------------
			edgeGateways, err := vdc.GetAllNsxtEdgeGateways(nil)
			if err != nil {
				return fmt.Errorf("error retrieving NSX-T Edge Gateway list: %s", err)
			}
			for _, edgeGw := range edgeGateways {
				toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, edgeGw.EdgeGateway.Name, "vcd_nsxt_edgegateway", 2, verbose)
				if toBeDeleted {
					err = edgeGw.Delete()
					if err != nil {
						return fmt.Errorf("error deleting NSX-T Edge Gateway '%s': %s", edgeGw.EdgeGateway.Name, err)
					}
				}
			}
		}
	}
	// --------------------------------------------------------------
	// RDE Types
	// --------------------------------------------------------------
	rdeTypes, err := govcdClient.GetAllRdeTypes(nil)
	if err != nil {
		return fmt.Errorf("error retrieving RDE Types: %s", err)
	}
	for _, rdeType := range rdeTypes {
		rdes, err := rdeType.GetAllRdes(nil)
		if err != nil {
			return fmt.Errorf("error retrieving RDEs of type %s: %s", rdeType.DefinedEntityType.ID, err)
		}
		// --------------------------------------------------------------
		// RDEs
		// --------------------------------------------------------------
		for _, rde := range rdes {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, rde.DefinedEntity.Name, "vcd_rde", 2, verbose)
			if toBeDeleted {
				err = deleteRde(rde)
				if err != nil {
					return fmt.Errorf("error deleting RDE '%s' of type '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.EntityType, err)
				}
			}
		}
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, rdeType.DefinedEntityType.Name, "vcd_rde_type", 1, verbose)
		if toBeDeleted {
			err = deleteRdeType(rdeType)
			if err != nil {
				return fmt.Errorf("error deleting RDE Type '%s': %s", rdeType.DefinedEntityType.ID, err)
			}
		}
	}
	// --------------------------------------------------------------
	// RDE Interfaces
	// --------------------------------------------------------------
	definedInterfaces, err := govcdClient.GetAllDefinedInterfaces(nil)
	if err != nil {
		return fmt.Errorf("error retrieving RDE Types: %s", err)
	}
	for _, di := range definedInterfaces {
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, di.DefinedInterface.Name, "vcd_rde_interface", 1, verbose)
		if toBeDeleted {
			err = deleteRdeInterface(di)
			if err != nil {
				return fmt.Errorf("error deleting RDE Interface '%s': %s", di.DefinedInterface.Name, err)
			}
		}
	}

	// --------------------------------------------------------------
	// External Networks and Provider Gateways (only for SysAdmin)
	// --------------------------------------------------------------
	if govcdClient.Client.IsSysAdmin {
		externalNetworks, err := govcd.GetAllExternalNetworksV2(govcdClient, nil)
		if err != nil {
			return fmt.Errorf("error retrieving External Network list: %s", err)
		}
		for _, extNet := range externalNetworks {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, extNet.ExternalNetwork.Name, "vcd_external_network_v2", 0, verbose)
			if toBeDeleted {
				err = extNet.Delete()
				if err != nil {
					return fmt.Errorf("error deleting External Network '%s': %s", extNet.ExternalNetwork.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// IP Allocations and IP Spaces (VCD 10.4.1+)
	// --------------------------------------------------------------
	if govcdClient.Client.APIVCDMaxVersionIs(">= 37.1") {
		err = removeLeftoversIpSpacesAndAllocations(govcdClient, true)
		if err != nil {
			return err
		}
	}

	// --------------------------------------------------------------
	// Plugins
	// --------------------------------------------------------------
	uiPlugins, err := govcdClient.GetAllUIPlugins()
	if err != nil {
		return fmt.Errorf("error retrieving UI Plugins: %s", err)
	}
	for _, uiPlugin := range uiPlugins {
		// This will delete all UI Plugins that match the `isTest` regex.
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, uiPlugin.UIPluginMetadata.PluginName, "vcd_ui_plugin", 1, verbose)
		if toBeDeleted {
			err = deleteUIPlugin(uiPlugin)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// removeLeftoversNsxtAlb is responsible for cleanup up NSX-T ALB leftovers
// Note. Due to the fact that we need to test Controller creation and deletion - there should never
// remain any NSX-T ALB leftovers after the test.
//
// Hierarchy of NSX-T ALB Configurations is:
// Provider part (Infrastructure configuration)
// 1. NSX-T ALB Controller (vcd_nsxt_alb_controller)
// 2. NSX-T Clouds (vcd_nsxt_alb_cloud)
// 3. NSX-T Service Engine Groups (vcd_nsxt_alb_service_engine_group)
// Tenant part (Edge Gateway children)
// 4. Edge Gateway settings (vcd_nsxt_alb_settings)
// 5. Edge Gateway Service Engine Group assignment (vcd_nsxt_alb_edgegateway_service_engine_group)
// 6. NSX-T ALB Pools (vcd_nsxt_alb_pool)
// 7. NSX-T ALB Virtual Services (vcd_nsxt_alb_virtual_service)
// This structure must be cleaned up in reverse order (starting from 7 and ending
// with 1) so that no dependency constraints are violated
func removeLeftoversNsxtAlb(govcdClient *govcd.VCDClient, verbose bool) error {
	// --------------------------------------------------------------
	// vcd_nsxt_alb_controller
	// --------------------------------------------------------------
	albControllers, err := govcdClient.GetAllAlbControllers(nil)
	if err != nil {
		return fmt.Errorf("error retrieving NSX-T ALB Controllers list: %s", err)
	}
	for _, albController := range albControllers {
		albControllertoBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, albController.NsxtAlbController.Name, "vcd_nsxt_alb_controller", 0, verbose)

		// --------------------------------------------------------------
		// vcd_nsxt_alb_cloud
		// --------------------------------------------------------------
		albClouds, err := govcdClient.GetAllAlbClouds(nil)
		if err != nil {
			return fmt.Errorf("error retrieving NSX-T ALB Clouds list: %s", err)
		}
		for _, albCloud := range albClouds {
			albCloudToBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, albCloud.NsxtAlbCloud.Name, "vcd_nsxt_alb_cloud", 1, verbose)

			// --------------------------------------------------------------
			// vcd_nsxt_alb_service_engine_group
			// --------------------------------------------------------------
			albServiceEngineGroups, err := govcdClient.GetAllAlbServiceEngineGroups("", nil)

			if err != nil {
				return fmt.Errorf("error retrieving NSX-T ALB Service Engine Groups list: %s", err)
			}
			for _, albServiceEngineGroup := range albServiceEngineGroups {
				albServiceEngineGroupToBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, albServiceEngineGroup.NsxtAlbServiceEngineGroup.Name, "vcd_nsxt_alb_service_engine_group", 2, verbose)
				if albServiceEngineGroupToBeDeleted {

					// --------------------------------------------------------------
					// NSX-T ALB Tenant Configuration cleanup
					// --------------------------------------------------------------
					err = removeLeftoversNsxtAlbTenant(govcdClient, verbose)
					if err != nil {
						return fmt.Errorf("error removing NSX-T ALB Tenant leftovers: %s", err)
					}

					// vcd_nsxt_alb_service_engine_group deletion
					err = albServiceEngineGroup.Delete()
					if err != nil {
						return fmt.Errorf("error deleting NSX-T ALB Service Engine Group '%s': %s", albServiceEngineGroup.NsxtAlbServiceEngineGroup.Name, err)
					}
				}
			}

			// vcd_nsxt_alb_cloud deletion
			if albCloudToBeDeleted {
				err = albCloud.Delete()
				if err != nil {
					return fmt.Errorf("error deleting NSX-T ALB Cloud '%s': %s", albCloud.NsxtAlbCloud.Name, err)
				}
			}
		}

		// vcd_nsxt_alb_controller deletion
		if albControllertoBeDeleted {
			err = albController.Delete()
			if err != nil {
				return fmt.Errorf("error deleting NSX-T ALB Controller '%s': %s", albController.NsxtAlbController.Name, err)
			}
		}
	}
	return nil
}

// removeLeftoversNsxtAlbTenant is responsible for cleanup up NSX-T ALB leftovers in Edge Gateways
// (tenant configuration)
// 1. Edge Gateway settings (vcd_nsxt_alb_settings)
// 2. Edge Gateway Service Engine Group assignment (vcd_nsxt_alb_edgegateway_service_engine_group)
// 3. NSX-T ALB Pools (vcd_nsxt_alb_pool)
// 4. NSX-T ALB Virtual Services (vcd_nsxt_alb_virtual_service)
func removeLeftoversNsxtAlbTenant(govcdClient *govcd.VCDClient, verbose bool) error {
	// Iterate over Edge Gateways and ensure no tenant ALB configurations remain
	edgeGateways, err := govcdClient.GetAllNsxtEdgeGateways(nil)
	if err != nil {
		return fmt.Errorf("error retrieving NSX-T Edge Gateway list: %s", err)
	}

	for _, edgeGateway := range edgeGateways {

		// --------------------------------------------------------------
		// vcd_nsxt_alb_settings
		// --------------------------------------------------------------
		albSetting, err := edgeGateway.GetAlbSettings()
		if err != nil {
			return fmt.Errorf("error retrieving NSX-T ALB Edge Gateway settings: %s", err)
		}

		if albSetting.Enabled {
			// Real name of Edge Gateway would not be matched, but ALB needs to be deleted therefore adding a prefix Test
			edgeNameWithPrefix := "Test" + edgeGateway.EdgeGateway.Name
			toBeDisabledAlbSettings := shouldDeleteEntity(alsoDelete, doNotDelete, edgeNameWithPrefix, "vcd_nsxt_alb_settings", 3, verbose)

			// --------------------------------------------------------------
			// vcd_nsxt_alb_edgegateway_service_engine_group
			// --------------------------------------------------------------
			filterOnEdgeGateway := url.Values{}
			filterOnEdgeGateway.Add("filter", fmt.Sprintf("(gatewayRef.id==%s)", edgeGateway.EdgeGateway.ID))
			albEdgeGatewayServiceEngineGroups, err := govcdClient.GetAllAlbServiceEngineGroupAssignments(filterOnEdgeGateway)
			if err != nil {
				return fmt.Errorf("error retrieving NSX-T ALB Edge Gateway Service Engine Group assignments list: %s", err)
			}

			for _, albEdgeGatewayServiceEngineGroup := range albEdgeGatewayServiceEngineGroups {
				// Edge Gateway Service Engine Group Assignment does not have names, therefore we use
				// Service Engine Group name as a name for the resource
				edgeServiceEngineGroupAssignmentName := albEdgeGatewayServiceEngineGroup.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name
				toBeDeletedAlbEdgeGatewayServiceEngineGroup := shouldDeleteEntity(alsoDelete, doNotDelete, edgeServiceEngineGroupAssignmentName, "vcd_nsxt_alb_edgegateway_service_engine_group", 4, verbose)

				// --------------------------------------------------------------
				// vcd_nsxt_alb_pool
				// --------------------------------------------------------------
				albPools, err := govcdClient.GetAllAlbPools(edgeGateway.EdgeGateway.ID, nil)
				if err != nil {
					return fmt.Errorf("error retrieving NSX-T ALB Pools list: %s", err)
				}

				for _, albPool := range albPools {
					toBeDeletedAlbPool := shouldDeleteEntity(alsoDelete, doNotDelete, albPool.NsxtAlbPool.Name, "vcd_nsxt_alb_pool", 5, verbose)

					// --------------------------------------------------------------
					// vcd_nsxt_alb_virtual_service
					// --------------------------------------------------------------
					albVirtualServices, err := govcdClient.GetAllAlbVirtualServices(edgeGateway.EdgeGateway.ID, nil)
					if err != nil {
						return fmt.Errorf("error retrieving NSX-T ALB Virtual Services list: %s", err)
					}

					for _, albVs := range albVirtualServices {
						toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, albVs.NsxtAlbVirtualService.Name, "vcd_nsxt_alb_virtual_service", 6, verbose)
						if toBeDeleted {
							err = albVs.Delete()
							if err != nil {
								return fmt.Errorf("error deleting NSX-T ALB Virtual Service '%s': %s", albVs.NsxtAlbVirtualService.Name, err)
							}
						}
					}

					// vcd_nsxt_alb_pool deletion
					if toBeDeletedAlbPool {
						err = albPool.Delete()
						if err != nil {
							return fmt.Errorf("error deleting NSX-T ALB Pool '%s': %s", albPool.NsxtAlbPool.Name, err)
						}
					}
				}

				// vcd_nsxt_alb_edgegateway_service_engine_group deletion
				if toBeDeletedAlbEdgeGatewayServiceEngineGroup {
					err = albEdgeGatewayServiceEngineGroup.Delete()
					if err != nil {
						return fmt.Errorf("error deleting NSX-T ALB Edge Gateway Service Engine Group assignment '%s': %s", edgeServiceEngineGroupAssignmentName, err)
					}
				}
			}

			// vcd_nsxt_alb_settings
			if toBeDisabledAlbSettings {
				err = edgeGateway.DisableAlb()
				if err != nil {
					return fmt.Errorf("error disabling NSX-T ALB Edge Gateway settings: %s", err)
				}
			}
		}
	}

	return nil
}

func removeLeftoversIpSpacesAndAllocations(govcdClient *govcd.VCDClient, verbose bool) error {
	allIpSpaces, err := govcdClient.GetAllIpSpaceSummaries(nil)
	if err != nil {
		return fmt.Errorf("error retrieving IP spaces: %s", err)
	}

	// Remove all IP allocations before removing IP Space itself
	for _, ipSpace := range allIpSpaces {
		// IP Space itself
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, ipSpace.IpSpace.Name, "vcd_ip_space", 0, verbose)
		if toBeDeleted {

			// Removing Floating IP Allocations
			err = removeLeftoversIpAllocations(ipSpace, "FLOATING_IP", verbose)
			if err != nil {
				return err
			}
			// Removing IP Prefix Allocations
			err = removeLeftoversIpAllocations(ipSpace, "IP_PREFIX", verbose)
			if err != nil {
				return err
			}

			fmt.Printf("REMOVING IP Space IP  %s\n", ipSpace.IpSpace.Name)
			err = ipSpace.Delete()
			if err != nil {
				return fmt.Errorf("error deleting IP Space '%s': %s", ipSpace.IpSpace.Name, err)
			}
		}
	}

	return nil
}

func removeLeftoversIpAllocations(ipSpace *govcd.IpSpace, ipSpaceType string, verbose bool) error {
	allIpSpaceIpAllocations, err := ipSpace.GetAllIpSpaceAllocations(ipSpaceType, nil)
	if err != nil {
		return fmt.Errorf("error retrieving Floating IP allocations for IP Space '%s'('%s'): %s",
			ipSpace.IpSpace.Name, ipSpace.IpSpace.ID, err)
	}

	for _, ipAllocation := range allIpSpaceIpAllocations {
		fmt.Printf("\tREMOVING IP Space IP Allocation %s\n", ipAllocation.IpSpaceIpAllocation.Value)
		err = ipAllocation.Delete()
		if err != nil {
			return fmt.Errorf("error deleting IP Allocation '%s' in IP Space %s: %s", ipAllocation.IpSpaceIpAllocation.ID, ipSpace.IpSpace.Name, err)
		}
	}

	return nil
}

// shouldDeleteEntity checks whether a given entity is to be deleted, either by its name
// or by its inclusion in one of the entity lists
func shouldDeleteEntity(alsoDelete, doNotDelete entityList, name, entityType string, level int, verbose bool) bool {
	inclusion := ""
	exclusion := ""
	// 1. First requirement to be deleted: the entity name starts with 'Test' or 'test'
	toBeDeleted := isTest.MatchString(name)
	if inList(alsoDelete, name, entityType) {
		toBeDeleted = true
		// 2. If the entity was in the additional deletion list, regardless of the name,
		// it is marked for deletion, with a "+", indicating that it was selected for deletion because of the
		// deletion list
		inclusion = " +"
	}
	if inList(doNotDelete, name, entityType) {
		toBeDeleted = false
		// 3. If a file, normally marked for deletion, is found in the keep list,
		// its deletion status is revoked, and it is marked with a "-", indicating that it was excluded
		// for deletion because of the keep list
		exclusion = " -"
	}
	tabs := strings.Repeat("\t", level)
	format := tabs + "[%s] %s (%s%s%s)\n"
	deletionText := "DELETE"
	if !toBeDeleted {
		deletionText = "keep"
	}

	// 4. Show the entity. If it is to be deleted, it will always be shown
	if toBeDeleted || contains(alwaysShow, entityType) {
		if verbose {
			fmt.Printf(format, entityType, name, deletionText, inclusion, exclusion)
		}
	}
	return toBeDeleted
}

// inList shows whether a given entity is included in an entityList
func inList(list entityList, name, entityType string) bool {
	for _, element := range list {
		if element.Name == name && element.Type == entityType {
			return true
		}
	}
	return false
}

func deleteVdc(org *govcd.Org, vdc *govcd.Vdc) error {
	fmt.Printf("\t REMOVING VDC %s/%s\n", org.Org.Name, vdc.Vdc.Name)
	task, err := vdc.Delete(true, true)
	if err != nil {
		return fmt.Errorf("error initiating VDC '%s' deletion: %s", vdc.Vdc.Name, err)
	}
	return task.WaitTaskCompletion()
}

func deleteVapp(vdc *govcd.Vdc, vappRef *types.ResourceReference) error {
	vapp, err := vdc.GetVAppByHref(vappRef.HREF)
	if err != nil {
		return fmt.Errorf("error retrieving vApp %s: %s", vappRef.Name, err)
	}
	fmt.Printf("\t\t REMOVING vApp %s/%s\n", vdc.Vdc.Name, vapp.VApp.Name)
	vappStatus, err := vapp.GetStatus()
	if err != nil {
		return fmt.Errorf("error reading vApp status '%s': %s", vappRef.Name, err)
	}
	if vappStatus != "POWERED_OFF" {
		task, err := vapp.Undeploy()
		if err != nil {
			return fmt.Errorf("error initiating vApp '%s' undeploy: %s", vappRef.Name, err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error undeploying vApp '%s': %s", vappRef.Name, err)
		}
	}

	task, err := vapp.RemoveAllNetworks()
	if err != nil {
		return fmt.Errorf("error initiating vApp '%s' network removal: %s", vappRef.Name, err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error removing networks from vApp '%s': %s", vappRef.Name, err)
	}
	task, err = vapp.Delete()
	if err != nil {
		return fmt.Errorf("error initiating vApp '%s' deletion: %s", vappRef.Name, err)
	}
	return task.WaitTaskCompletion()
}

func deleteMediaItem(catalog *govcd.Catalog, mediaRec *types.MediaRecordType) error {
	media, err := catalog.GetMediaByHref(mediaRec.HREF)
	if err != nil {
		return fmt.Errorf("error retrieving media item '%s': %s", mediaRec.Name, err)
	}
	fmt.Printf("\t\t REMOVING media item %s/%s\n", catalog.Catalog.Name, media.Media.Name)
	task, err := media.Delete()
	if err != nil {
		return fmt.Errorf("error initiating media item '%s' deletion: %s", mediaRec.Name, err)
	}
	return task.WaitTaskCompletion()
}

func deleteRde(rde *govcd.DefinedEntity) error {
	if rde.DefinedEntity.State != nil && *rde.DefinedEntity.State == "PRE_CREATED" {
		err := rde.Resolve()
		if err != nil {
			return fmt.Errorf("error resolving RDE '%s' before deletion: %s", rde.DefinedEntity.Name, err)
		}
	}
	fmt.Printf("\t\t REMOVING RDE %s WITH TYPE %s\n", rde.DefinedEntity.Name, rde.DefinedEntity.EntityType)
	err := rde.Delete()
	if err != nil {
		return fmt.Errorf("error deleting RDE '%s' with type '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.EntityType, err)
	}
	return nil
}

func deleteRdeType(rdeType *govcd.DefinedEntityType) error {
	fmt.Printf("\t\t REMOVING RDE TYPE %s\n", rdeType.DefinedEntityType.ID)
	err := rdeType.Delete()
	if err != nil {
		return fmt.Errorf("error deleting RDE type '%s': %s", rdeType.DefinedEntityType.ID, err)
	}
	return nil
}

func deleteUIPlugin(uiPlugin *govcd.UIPlugin) error {
	fmt.Printf("\t\t REMOVING UI PLUGIN %s\n", uiPlugin.UIPluginMetadata.ID)
	err := uiPlugin.Delete()
	if err != nil {
		return fmt.Errorf("error deleting UI Plugin '%s': %s", uiPlugin.UIPluginMetadata.ID, err)
	}
	return nil
}

func deleteRdeInterface(di *govcd.DefinedInterface) error {
	fmt.Printf("\t\t REMOVING RDE INTERFACE %s\n", di.DefinedInterface.ID)
	err := di.Delete()
	if err != nil {
		return fmt.Errorf("error deleting RDE Interface '%s': %s", di.DefinedInterface.ID, err)
	}
	return nil
}
