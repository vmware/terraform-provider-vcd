package helper

import (
	types "github.com/ukcloud/govcloudair/types/v56"
)

func ConfigureNetwork(networkConnection map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	// TODO: Network changes
	// name
	// ip
	// ip_allocation_mode
	// is_primary
	// is_connected
	// adapter_type
	return nil, nil
}

func ReadNetwork(networkConnection *types.NetworkConnection, primaryInterfaceIndex int) map[string]interface{} {
	readNetwork := make(map[string]interface{})

	readNetwork["name"] = networkConnection.Network
	readNetwork["ip"] = networkConnection.IPAddress
	readNetwork["ip_allocation_mode"] = networkConnection.IPAddressAllocationMode
	readNetwork["is_primary"] = (primaryInterfaceIndex == networkConnection.NetworkConnectionIndex)
	readNetwork["is_connected"] = networkConnection.IsConnected
	readNetwork["adapter_type"] = networkConnection.NetworkAdapterType

	return readNetwork
}
