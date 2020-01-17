/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type responseEdgeDhcpLeases struct {
	XMLName   xml.Name            `xml:"dhcpLeases"`
	TimeStamp string              `xml:"timeStamp"`
	DhcpLease types.EdgeDhcpLease `xml:"dhcpLeaseInfo"`
}

// GetNsxvActiveDhcpLeaseByMac finds active DHCP lease for a given hardware address (MAC)
func (egw *EdgeGateway) GetNsxvActiveDhcpLeaseByMac(mac string) (*types.EdgeDhcpLeaseInfo, error) {
	if mac == "" {
		return nil, fmt.Errorf("MAC address must be provided to lookup DHCP lease")
	}
	dhcpLeases, err := egw.GetAllNsxvDhcpLeases()
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] Looking up active DHCP lease for MAC: %s", mac)
	for _, lease := range dhcpLeases {
		util.Logger.Printf("[DEBUG] Checking DHCP lease: %#+v", lease)
		if lease.BindingState == "active" && lease.MacAddress == mac {
			return lease, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetAllNsxvDhcpLeases retrieves all DHCP leases defined in NSX-V edge gateway
func (egw *EdgeGateway) GetAllNsxvDhcpLeases() ([]*types.EdgeDhcpLeaseInfo, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpLeasePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	dhcpLeases := &responseEdgeDhcpLeases{}

	// This query returns all DHCP leases
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read DHCP leases: %s", nil, dhcpLeases)
	if err != nil {
		return nil, err
	}

	if dhcpLeases != nil && len(dhcpLeases.DhcpLease.DhcpLeaseInfos) == 0 {
		util.Logger.Printf("[DEBUG] GetAllNsxvDhcpLeases found 0 leases available")
		return nil, ErrorEntityNotFound
	}

	return dhcpLeases.DhcpLease.DhcpLeaseInfos, nil
}
