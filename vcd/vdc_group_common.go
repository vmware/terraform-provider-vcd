package vcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// vdcOrVdcGroupHandler is an interface to access some common methods on VDC or VDC Group without
// explicitly handling exact types
type vdcOrVdcGroupHandler interface {
	IsNsxt() bool
	GetNsxtEdgeGatewayByName(name string) (*govcd.NsxtEdgeGateway, error)
	GetOpenApiOrgVdcNetworkByName(name string) (*govcd.OpenApiOrgVdcNetwork, error)
	GetNsxtImportableSwitchByName(name string) (*govcd.NsxtImportableSwitch, error)
	GetNsxtFirewallGroupByName(name, firewallGroupType string) (*govcd.NsxtFirewallGroup, error)
	GetNsxtFirewallGroupById(id string) (*govcd.NsxtFirewallGroup, error)
	GetNsxtAppPortProfileByName(name, scope string) (*govcd.NsxtAppPortProfile, error)
	GetOpenApiOrgVdcNetworkById(id string) (*govcd.OpenApiOrgVdcNetwork, error)
}

// getVdcOrVdcGroupVerifierByOwnerId helps to find VDC or VDC Group by ownerId field and returns an
// interface type `vdcOrVdcGroupHandler` so that some functions can be called directly without
// careing if the object is VDC or VDC Group
func getVdcOrVdcGroupVerifierByOwnerId(org *govcd.Org, ownerId string) (vdcOrVdcGroupHandler, error) {
	var vdcOrVdcGroup vdcOrVdcGroupHandler
	var err error
	switch {
	case govcd.OwnerIsVdc(ownerId):
		vdcOrVdcGroup, err = org.GetVDCById(ownerId, false)
		if err != nil {
			return nil, err
		}
	case govcd.OwnerIsVdcGroup(ownerId):
		vdcOrVdcGroup, err = org.GetVdcGroupById(ownerId)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("error determining VDC type by ID '%s'", ownerId)
	}

	return vdcOrVdcGroup, nil
}

// validateIfVdcOrVdcGroupIsNsxt evaluates VDC field priority using pickVdcIdByPriority and then
// checks if that VDC or VDC Group is an NSX-T one and returns an error if not
func validateIfVdcOrVdcGroupIsNsxt(org *govcd.Org, inheritedVdcField, vdcField, ownerIdField string) error {
	usedFieldId, _, err := pickVdcIdByPriority(org, inheritedVdcField, vdcField, ownerIdField)
	if err != nil {
		return fmt.Errorf("error finding VDC ID: %s", err)
	}

	isNsxt, err := isBackedByNsxt(org, usedFieldId)
	if err != nil {
		return fmt.Errorf("error checking if VDC or VDC Group is backed by NSX-T: %s", err)
	}

	if !isNsxt {
		return fmt.Errorf("this resource does not support NSX-V")
	}

	return nil
}

// pickVdcIdByPriority picks primary field to be used from the specified ones. The priority is such
// * `owner_id`
// * `vdc` at resource level
// * `vdc` inherited from provider configuration
func pickVdcIdByPriority(org *govcd.Org, inheritedVdcField, vdcField, ownerIdField string) (string, *govcd.Vdc, error) {
	if ownerIdField != "" {
		return ownerIdField, nil, nil
	}

	if vdcField != "" {
		vdc, err := org.GetVDCByName(vdcField, false)
		if err != nil {
			return "", nil, fmt.Errorf("error finding VDC '%s': %s", vdcField, err)
		}
		return vdc.Vdc.ID, vdc, nil
	}

	if inheritedVdcField != "" {
		vdc, err := org.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", nil, fmt.Errorf("error finding VDC '%s': %s", inheritedVdcField, err)
		}
		return vdc.Vdc.ID, vdc, nil
	}

	return "", nil, fmt.Errorf("none of the fields `owner_id`, `vdc` and provider inherited `vdc`")
}

// isBackedByNsxt accepts VDC or VDC Group ID and checks if it is backed by NSX-T
func isBackedByNsxt(org *govcd.Org, vdcOrVdcGroupId string) (bool, error) {
	var vdcOrVdcGroup vdcOrVdcGroupVerifier
	var err error

	switch {
	case govcd.OwnerIsVdc(vdcOrVdcGroupId):
		vdcOrVdcGroup, err = org.GetVDCById(vdcOrVdcGroupId, false)
		if err != nil {
			return false, err
		}
	case govcd.OwnerIsVdcGroup(vdcOrVdcGroupId):
		vdcOrVdcGroup, err = org.GetVdcGroupById(vdcOrVdcGroupId)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("error determining VDC type by ID '%s'", vdcOrVdcGroupId)
	}

	return vdcOrVdcGroup.IsNsxt(), nil
}

// lookupVdcOrVdcGroup is an import helper for entities whose parent can be:
// * org-name.vdc-name
// * org-name.vdc-group-name
// It will return an interface type of `vdcOrVdcGroupHandler` which has common methods needed to
// handle further actions in the Import functions for Terraform resources
func lookupVdcOrVdcGroup(vcdClient *VCDClient, orgName, vdcOrVdcGroupName string) (vdcOrVdcGroupHandler, error) {
	var vdcOrVdcGroup vdcOrVdcGroupHandler
	var err error
	_, vdcOrVdcGroup, err = vcdClient.GetOrgAndVdc(orgName, vdcOrVdcGroupName)
	if err != nil && !govcd.ContainsNotFound(err) {
		return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s' in Org %s: %s", vdcOrVdcGroupName, orgName, err)
	}
	if govcd.ContainsNotFound(err) {
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrVdcGroup, err = adminOrg.GetVdcGroupByName(vdcOrVdcGroupName)
		if err != nil {
			return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s': %s", vdcOrVdcGroupName, err)
		}
	}

	return vdcOrVdcGroup, nil
}
