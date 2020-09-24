package types

import (
	"encoding/json"
	"fmt"
)

// OpenApiPages unwraps pagination for "Get All" endpoints in OpenAPI. Values kept in json.RawMessage helps to decouple
// marshalling paging related information from exact type related information. Paging can be handled dynamically this
// way while values can be marshaled into exact types.
type OpenApiPages struct {
	// ResultTotal reports total results available
	ResultTotal int `json:"resultTotal,omitempty"`
	// PageCount reports total result pages available
	PageCount int `json:"pageCount,omitempty"`
	// Page reports current page of result
	Page int `json:"page,omitempty"`
	// PageSize reports page size
	PageSize int `json:"pageSize,omitempty"`
	// Associations ...
	Associations interface{} `json:"associations,omitempty"`
	// Values holds types depending on the endpoint therefore `json.RawMessage` is used to dynamically unmarshal into
	// specific type as required
	Values json.RawMessage `json:"values,omitempty"`
}

// OpenApiError helps to marshal and provider meaningful `Error` for
type OpenApiError struct {
	MinorErrorCode string `json:"minorErrorCode"`
	Message        string `json:"message"`
	StackTrace     string `json:"stackTrace"`
}

// Error method implements Go's default `error` interface for CloudAPI errors formats them for human readable output.
func (openApiError OpenApiError) Error() string {
	return fmt.Sprintf("%s - %s", openApiError.MinorErrorCode, openApiError.Message)
}

// ErrorWithStack is the same as `Error()`, but also includes stack trace returned by API which is usually lengthy.
func (openApiError OpenApiError) ErrorWithStack() string {
	return fmt.Sprintf("%s - %s. Stack: %s", openApiError.MinorErrorCode, openApiError.Message,
		openApiError.StackTrace)
}

// Role defines access roles in VCD
type Role struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BundleKey   string `json:"bundleKey"`
	ReadOnly    bool   `json:"readOnly"`
}

// NsxtTier0Router defines NSX-T Tier 0 router
type NsxtTier0Router struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
}

// ExternalNetworkV2 defines a struct for OpenAPI endpoint
type ExternalNetworkV2 struct {
	ID              string                    `json:"id,omitempty"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	Subnets         ExternalNetworkV2Subnets  `json:"subnets"`
	NetworkBackings ExternalNetworkV2Backings `json:"networkBackings"`
}

// ExternalNetworkV2IPRange defines allocated IP pools for a subnet in external network
type ExternalNetworkV2IPRange struct {
	StartAddress string `json:"startAddress"`
	EndAddress   string `json:"endAddress"`
}

// ExternalNetworkV2IPRanges contains slice of ExternalNetworkV2IPRange
type ExternalNetworkV2IPRanges struct {
	Values []ExternalNetworkV2IPRange `json:"values"`
}

// ExternalNetworkV2Subnets contains slice of ExternalNetworkV2Subnet
type ExternalNetworkV2Subnets struct {
	Values []ExternalNetworkV2Subnet `json:"values"`
}

// ExternalNetworkV2Subnet defines one subnet for external network with assigned static IP ranges
type ExternalNetworkV2Subnet struct {
	Gateway      string                    `json:"gateway"`
	PrefixLength int                       `json:"prefixLength"`
	DNSSuffix    string                    `json:"dnsSuffix"`
	DNSServer1   string                    `json:"dnsServer1"`
	DNSServer2   string                    `json:"dnsServer2"`
	IPRanges     ExternalNetworkV2IPRanges `json:"ipRanges"`
	Enabled      bool                      `json:"enabled"`
	UsedIPCount  int                       `json:"usedIpCount,omitempty"`
	TotalIPCount int                       `json:"totalIpCount,omitempty"`
}

type ExternalNetworkV2Backings struct {
	Values []ExternalNetworkV2Backing `json:"values"`
}

// ExternalNetworkV2Backing defines which networking subsystem is used for external network (NSX-T or NSX-V)
type ExternalNetworkV2Backing struct {
	// BackingID must contain either Tier-0 router ID for NSX-T or PortGroup ID for NSX-V
	BackingID string `json:"backingId"`
	Name      string `json:"name,omitempty"`
	// BackingType can be either ExternalNetworkBackingTypeNsxtTier0Router in case of NSX-T or one of
	// ExternalNetworkBackingTypeNetwork or ExternalNetworkBackingDvPortgroup in case of NSX-V
	BackingType     string                  `json:"backingType"`
	NetworkProvider NetworkProviderProvider `json:"networkProvider"`
}

// NetworkProvider can be NSX-T manager or vCenter
type NetworkProviderProvider struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id"`
}
