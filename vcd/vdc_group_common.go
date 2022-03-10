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
}

// getVdcOrVdcGroupVerifierByOwnerId helps to find VDC or VDC Group by ownerId field and returns an
// interface type `vdcOrVdcGroupHandler` so that some functions can be called directly without
// careing if the object is VDC or VDC Group
func getVdcOrVdcGroupVerifierByOwnerId(org *govcd.Org, ownerId string) (vdcOrVdcGroupHandler, error) {
	var vdcOrGroup vdcOrVdcGroupHandler
	var err error
	switch {
	case govcd.OwnerIsVdc(ownerId):
		vdcOrGroup, err = org.GetVDCById(ownerId, false)
		if err != nil {
			return nil, err
		}
	case govcd.OwnerIsVdcGroup(ownerId):
		vdcOrGroup, err = org.GetVdcGroupById(ownerId)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("error determining VDC type by ID '%s'", ownerId)
	}

	return vdcOrGroup, nil
}
