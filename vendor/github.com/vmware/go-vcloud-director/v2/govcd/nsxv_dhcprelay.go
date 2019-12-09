/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// UpdateDhcpRelay updates DHCP relay settings for a particular edge gateway and returns them. The
// feature itself enables you to leverage your existing DHCP infrastructure from within NSX without
// any interruption to the IP address management in your environment. DHCP messages are relayed from
// virtual machine(s) to the designated DHCP server(s) in the physical world. This enables IP
// addresses within NSX to continue to be in sync with IP addresses in other environments.
func (egw *EdgeGateway) UpdateDhcpRelay(dhcpRelayConfig *types.EdgeDhcpRelay) (*types.EdgeDhcpRelay, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP relay")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpRelayPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusNoContent or if not an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error setting DHCP relay settings: %s", dhcpRelayConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	return egw.GetDhcpRelay()
}

// GetDhcpRelay retrieves a structure of *types.EdgeDhcpRelay with all DHCP relay settings present
// on a particular edge gateway.
func (egw *EdgeGateway) GetDhcpRelay() (*types.EdgeDhcpRelay, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP relay")
	}
	response := &types.EdgeDhcpRelay{}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpRelayPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// This query Edge gaateway DHCP relay using proxied NSX-V API
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read edge gateway DHCP relay configuration: %s", nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ResetDhcpRelay removes all configuration by sending a DELETE request for DHCP relay configuration
// endpoint
func (egw *EdgeGateway) ResetDhcpRelay() error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support DHCP relay")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpRelayPath)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Send a DELETE request to DHCP relay configuration endpoint
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to reset edge gateway DHCP relay configuration: %s", nil, &types.NSXError{})
	return err
}
