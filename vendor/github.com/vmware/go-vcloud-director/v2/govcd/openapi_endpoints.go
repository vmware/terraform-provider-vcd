package govcd

import "fmt"

var endpointMinApiVersions = map[string]string{
	"1.0.0/roles/": "31.0",
}

// checkOpenApiEndpointCompatibility checks if required VCD version is sufficient to work with this endpoint and returns
// either error, either version
func (client *Client) checkOpenApiEndpointCompatibility(endpoint string) (string, error) {
	minimumApiVersion, ok := endpointMinApiVersions[endpoint]
	if !ok {
		return "", fmt.Errorf("minimum API version for endopoint '%s' is not defined", endpoint)
	}

	if client.APIVCDMaxVersionIs("< " + minimumApiVersion) {
		maxSupportedVersion, err := client.maxSupportedVersion()
		if err != nil {
			return "", fmt.Errorf("error reading maximum supported API version: %s", err)
		}
		return "", fmt.Errorf("endpoint '%s' requires API version to support at least '%s'. Maximum supported version in this instance: '%s'",
			endpoint, minimumApiVersion, maxSupportedVersion)
	}

	return minimumApiVersion, nil
}
