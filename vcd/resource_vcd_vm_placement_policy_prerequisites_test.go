//go:build vdc || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// vdcPlacementPolicyOrgUserPrerequisites helps to facilitate prerequisite buildup and teardown so that Org user can test placement policy data source
// It creates and cleans up the following:
// * VM Placement Policy
// * Org VDC
// * Assign the new VM Placement Policy to VDC
type vdcPlacementPolicyOrgUserPrerequisites struct {
	// t testing struct must be injected here because Terraform acceptance test framework does not accept and function
	// parameters
	t         *testing.T
	vcdClient *VCDClient

	placementPolicy *govcd.VdcComputePolicyV2
	vdc             *govcd.Vdc
}

func (v *vdcPlacementPolicyOrgUserPrerequisites) setup() {
	t := v.t
	vcdClient := v.vcdClient

	// Lookup Org
	adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Logf("error retrieving Org: %s", err)
	}

	pVdc, err := vcdClient.GetProviderVdcByName(testConfig.VCD.NsxtProviderVdc.Name)
	if err != nil {
		t.Logf("error retrieving Provider VDC: %s", err)
	}

	// Lookup storage profile
	storageProfile, err := vcdClient.QueryProviderVdcStorageProfileByName(testConfig.VCD.NsxtProviderVdc.StorageProfile, pVdc.ProviderVdc.HREF)
	if err != nil {
		t.Logf("error retrieving Storage Profile: %s", err)
	}

	networkPools, err := govcd.QueryNetworkPoolByName(vcdClient.VCDClient, testConfig.VCD.NsxtProviderVdc.NetworkPool)
	if err != nil || len(networkPools) < 1 {
		t.Logf("error retrieving Network Pool HREF: %s", err)
	}
	networkPoolHref := networkPools[0].HREF

	// Create VM Placement Policy
	// We also need the VM Group to create a VM Placement Policy
	fmt.Printf("# Creating VM Placement Policy. ")
	vmGroup, err := vcdClient.GetVmGroupByNameAndProviderVdcUrn(testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup, pVdc.ProviderVdc.ID)
	if err != nil {
		t.Logf("error retrieving VM Group: %s", err)
	}

	// Create a new VDC Compute Policy (VM Placement Policy)
	newComputePolicy := &govcd.VdcComputePolicyV2{
		VdcComputePolicyV2: &types.VdcComputePolicyV2{
			VdcComputePolicy: types.VdcComputePolicy{
				Name: t.Name(),
			},
			PolicyType: "VdcVmPolicy",
			PvdcNamedVmGroupsMap: []types.PvdcNamedVmGroupsMap{
				{
					NamedVmGroups: []types.OpenApiReferences{
						{
							types.OpenApiReference{
								Name: vmGroup.VmGroup.Name,
								ID:   fmt.Sprintf("%s:%s", "urn:vcloud:namedVmGroup", vmGroup.VmGroup.NamedVmGroupId),
							},
						},
					},
					Pvdc: types.OpenApiReference{
						Name: pVdc.ProviderVdc.Name,
						ID:   pVdc.ProviderVdc.ID,
					},
				},
			},
		},
	}

	createdPlacementPolicy, err := vcdClient.CreateVdcComputePolicyV2(newComputePolicy.VdcComputePolicyV2)
	if err != nil {
		t.Logf("error creating VM Placement Policy: %s", err)
	}

	v.placementPolicy = createdPlacementPolicy
	fmt.Printf("Done\n")

	// Create VDC and assign VM placement policy
	fmt.Printf("# Creating VDC. ")
	vdcConfiguration := &types.VdcConfiguration{
		Name:            t.Name(),
		AllocationModel: "AllocationPool",

		ComputeCapacity: []*types.ComputeCapacity{
			{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Units:     "MB",
					Allocated: 1024,
					Limit:     1024,
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{{
			Enabled: addrOf(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: storageProfile.HREF,
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: networkPoolHref,
		},
		ProviderVdcReference: &types.Reference{
			HREF: pVdc.ProviderVdc.HREF,
		},
		IsEnabled:            true,
		IsThinProvision:      true,
		UsesFastProvisioning: true,
	}

	err = adminOrg.CreateVdcWait(vdcConfiguration)
	if err != nil {
		t.Logf("error creating VDC: %s", err)
	}

	createdAdminVdc, err := adminOrg.GetAdminVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		t.Logf("error retrieving created Admin VDC: %s", err)
	}

	createdVdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		t.Logf("error retrieving created VDC: %s", err)
	}
	v.vdc = createdVdc
	fmt.Printf("Done\n")

	fmt.Printf("# Assigning placement policy to VDC. ")

	// Fetching existing default compute policy for VDC as it is autocreated and must persist
	existingComputePolicies, err := createdAdminVdc.GetAllAssignedVdcComputePolicies(nil)
	if err != nil {
		t.Logf("error retrieving default VDC Compute Policy: %s", err)
	}

	existingComputePolicyHref, _ := vcdClient.Client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies, extractUuid(existingComputePolicies[0].VdcComputePolicy.ID))
	newComputePolicyHref, _ := vcdClient.Client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies, createdPlacementPolicy.VdcComputePolicyV2.ID)
	cpReferences := types.VdcComputePolicyReferences{
		VdcComputePolicyReference: []*types.Reference{
			{ // Default one, must persist in VDC
				HREF: existingComputePolicyHref.String(),
			},
			{
				HREF: newComputePolicyHref.String(),
				ID:   createdPlacementPolicy.VdcComputePolicyV2.ID,
				Name: createdPlacementPolicy.VdcComputePolicyV2.Name,
				Type: createdPlacementPolicy.VdcComputePolicyV2.PolicyType,
			},
		},
	}

	_, err = createdAdminVdc.SetAssignedComputePolicies(cpReferences)
	if err != nil {
		t.Logf("error assigning Placement Policy to VDC: %s", err)
	}
	fmt.Printf("Done\n")

}

func (v *vdcPlacementPolicyOrgUserPrerequisites) teardown() {
	t := v.t

	if v.vdc != nil {
		fmt.Printf("# Cleaning up created VDC. ")
		err := v.vdc.DeleteWait(true, true)
		if err != nil {
			t.Logf("error removing created VDC '%s': %s", v.vdc.Vdc.Name, err)
		} else {
			fmt.Println("Done.")
		}
	}

	if v.placementPolicy != nil {
		fmt.Printf("# Cleaning up created Placement Policy. ")
		err := v.placementPolicy.Delete()
		if err != nil {
			t.Logf("error removing created Placement Policy '%s': %s", v.placementPolicy.VdcComputePolicyV2.Name, err)
		} else {
			fmt.Println("Done.")
		}
	}

}
