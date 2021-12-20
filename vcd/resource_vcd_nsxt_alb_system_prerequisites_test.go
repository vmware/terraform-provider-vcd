//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// albOrgUserPrerequisites helps to facilitate prerequisite buildup and teardown so that Org user can test NSX-T
// components (Pools and Virtual Services)
type albOrgUserPrerequisites struct {
	// t testing struct must be injected here because Terraform acceptance test framework does not accept and function
	// parameters
	t         *testing.T
	vcdClient *VCDClient

	albController           *govcd.NsxtAlbController
	albCloud                *govcd.NsxtAlbCloud
	albSeGroup              *govcd.NsxtAlbServiceEngineGroup
	nsxtGw                  *govcd.NsxtEdgeGateway
	nsxtGwSeGroupAssignment *govcd.NsxtAlbServiceEngineGroupAssignment
}

// setupAlbPoolPrerequisites sets up all ALB components that require System user
// The order is as follows:
// * NSX-T ALB Controller (provider section)
// * NSX-T ALB Cloud (provider section)
// * NSX-T ALB Service Engine Group (provider section)
// * Enables ALB On NSX-T Edge Gateway (Org section, but System user required)
// * Assigns Service Engine Group to NSX-T Edge Gateway (Org section, but System user required)
// Teardown must be done in exactly the opposite way so that child/parent relationships are not broken
func (a *albOrgUserPrerequisites) setupAlbPoolPrerequisites() {
	t := a.t
	vcdClient := a.vcdClient

	fmt.Printf("# Creating NSX-T ALB Controller. ")
	// Configure ALB Controller
	newControllerDef := &types.NsxtAlbController{
		Name:        t.Name(),
		Url:         testConfig.Nsxt.NsxtAlbControllerUrl,
		Username:    testConfig.Nsxt.NsxtAlbControllerUser,
		Password:    testConfig.Nsxt.NsxtAlbControllerPassword,
		LicenseType: "ENTERPRISE",
	}

	albController, err := vcdClient.CreateNsxtAlbController(newControllerDef)
	if err != nil {
		t.Logf("error creating NSX-T ALB Controller: %s", err)
	}
	a.albController = albController // Store in struct for teardown
	fmt.Println("Done.")

	// Configure ALB Cloud
	fmt.Printf("# Creating NSX-T ALB Cloud. ")
	importableCloud, err := albController.GetAlbImportableCloudByName(testConfig.Nsxt.NsxtAlbImportableCloud)
	if err != nil {
		t.Logf("error retrieving NSX-T ALB Importable Cloud: %s", err)
	}

	albCloudConfig := &types.NsxtAlbCloud{
		Name: t.Name(),
		LoadBalancerCloudBacking: types.NsxtAlbCloudBacking{
			BackingId: importableCloud.NsxtAlbImportableCloud.ID,
			LoadBalancerControllerRef: types.OpenApiReference{
				ID: albController.NsxtAlbController.ID,
			},
		},
		NetworkPoolRef: &types.OpenApiReference{
			ID: importableCloud.NsxtAlbImportableCloud.NetworkPoolRef.ID,
		},
	}

	createdAlbCloud, err := vcdClient.CreateAlbCloud(albCloudConfig)
	if err != nil {
		t.Logf("error creating NSX-T ALB Cloud: %s", err)
	}
	a.albCloud = createdAlbCloud // Store in struct for teardown
	fmt.Println("Done.")

	// Create Service Engine Group
	fmt.Printf("# Creating NSX-T ALB Service Engine Group. ")
	importableSeGroups, err := vcdClient.GetAllAlbImportableServiceEngineGroups(createdAlbCloud.NsxtAlbCloud.ID, nil)
	if err != nil || len(importableSeGroups) < 1 {
		t.Logf("error retrieving NSX-T ALB Importable Service Engine Groups: %s", err)
	}

	albSeGroup := &types.NsxtAlbServiceEngineGroup{
		Name:            t.Name(),
		ReservationType: "SHARED",
		ServiceEngineGroupBacking: types.ServiceEngineGroupBacking{
			BackingId: importableSeGroups[0].NsxtAlbImportableServiceEngineGroups.ID,
			LoadBalancerCloudRef: &types.OpenApiReference{
				ID: createdAlbCloud.NsxtAlbCloud.ID,
			},
		},
	}

	createdSeGroup, err := vcdClient.CreateNsxtAlbServiceEngineGroup(albSeGroup)
	if err != nil {
		t.Logf("error creating NSX-T ALB Service Engine Group: %s", err)
	}
	a.albSeGroup = createdSeGroup // Store in struct for teardown
	fmt.Println("Done.")

	// EOF Provider part setup

	// NSX-T Edge Gateway configuration
	fmt.Printf("# Enabling ALB on Edge Gateway. ")
	adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Logf("error getting AdminOrg '%s': %s", testConfig.VCD.Org, err)
	}

	vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, false)
	if err != nil {
		t.Logf("error getting NSX-T VDC '%s': %s", testConfig.Nsxt.Vdc, err)
	}

	nsxtEdge, err := vdc.GetNsxtEdgeGatewayByName(testConfig.Nsxt.EdgeGateway)
	if err != nil {
		t.Logf("error retrieving NSX-T Edge Gateway '%s': %s", testConfig.Nsxt.Vdc, err)
	}

	// Update ALB General Settings
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}

	_, err = nsxtEdge.UpdateAlbSettings(albSettingsConfig)
	if err != nil {
		t.Logf("error enabling ALB on NSX-T Edge Gateway: %s", err)
	}
	a.nsxtGw = nsxtEdge // Store in struct for teardown
	fmt.Println("Done.")

	// Assign Service Engine Group to Edge Gateway
	fmt.Printf("# Assigning Service Engine Group to Edge Gateway. ")
	serviceEngineGroupAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: nsxtEdge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: createdSeGroup.NsxtAlbServiceEngineGroup.ID},
		MaxVirtualServices:    takeIntPointer(89),
		MinVirtualServices:    takeIntPointer(20),
	}
	serviceEngineGroupAssignment, err := vcdClient.CreateAlbServiceEngineGroupAssignment(serviceEngineGroupAssignmentConfig)
	if err != nil {
		t.Logf("error assigning Service Engine Group '%s' to NSX-T Edge Gateway: %s", createdSeGroup.NsxtAlbServiceEngineGroup.Name, err)
	}
	a.nsxtGwSeGroupAssignment = serviceEngineGroupAssignment // Store in struct for teardown
	fmt.Println("Done.")
}

// teardownAlbPoolPrerequisites attempts to clean up any existing components
func (a *albOrgUserPrerequisites) teardownAlbPoolPrerequisites() {
	t := a.t
	if a.nsxtGwSeGroupAssignment != nil {
		fmt.Printf("# Cleaning up Service Engine Group assignment to Edge Gateway. ")
		err := a.nsxtGwSeGroupAssignment.Delete()
		if err != nil {
			t.Logf("error removing NSX-T Service Engine Group assignment to Edge Gateway: %s", err)
		} else {
			fmt.Println("Done.")
		}
	}

	if a.nsxtGw != nil {
		fmt.Printf("# Disabling ALB on Edge Gateway. ")
		err := a.nsxtGw.DisableAlb()
		if err != nil {
			t.Logf("error disabling ALB on Edge Gateway: %s", err)
		} else {
			fmt.Println("Done.")
		}
	}

	if a.albSeGroup != nil {
		fmt.Printf("# Removing ALB Service Engine Group. ")
		err := a.albSeGroup.Delete()
		if err != nil {
			t.Logf("error deleting ALB Service Engine Group: %s", err)
		} else {
			fmt.Println("Done.")
		}
	}

	if a.albCloud != nil {
		fmt.Printf("# Removing ALB Cloud. ")
		err := a.albCloud.Delete()
		if err != nil {
			t.Logf("error deleting ALB Cloud: %s", err)
		} else {
			fmt.Println("Done.")
		}
	}

	if a.albController != nil {
		fmt.Printf("# Removing ALB Controller. ")
		err := a.albController.Delete()
		if err != nil {
			t.Logf("error deleting ALB Controller: %s", err)
		} else {
			fmt.Println("Done.")
		}
	}
}
