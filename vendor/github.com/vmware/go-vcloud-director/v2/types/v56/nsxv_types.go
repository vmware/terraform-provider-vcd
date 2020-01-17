/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import "encoding/xml"

// FirewallConfigWithXml allows to enable/disable firewall on a specific edge gateway
// Reference: vCloud Director API for NSX Programming Guide
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
//
// Warning. It nests all firewall rules because Edge Gateway API is done so that if this data is not
// sent while enabling it would wipe all firewall rules. InnerXML type field is used with struct tag
//`innerxml` to prevent any manipulation of configuration and sending it verbatim
type FirewallConfigWithXml struct {
	XMLName       xml.Name              `xml:"firewall"`
	Enabled       bool                  `xml:"enabled"`
	DefaultPolicy FirewallDefaultPolicy `xml:"defaultPolicy"`

	// Each configuration change has a version number
	Version string `xml:"version,omitempty"`

	// The below field has `innerxml` tag so that it is not processed but instead
	// sent verbatim
	FirewallRules InnerXML `xml:"firewallRules,omitempty"`
	GlobalConfig  InnerXML `xml:"globalConfig,omitempty"`
}

// FirewallDefaultPolicy represent default rule
type FirewallDefaultPolicy struct {
	LoggingEnabled bool   `xml:"loggingEnabled"`
	Action         string `xml:"action"`
}

// LbGeneralParamsWithXml allows to enable/disable load balancing capabilities on specific edge gateway
// Reference: vCloud Director API for NSX Programming Guide
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
//
// Warning. It nests all components (LbMonitor, LbPool, LbAppProfile, LbAppRule, LbVirtualServer)
// because Edge Gateway API is done so that if this data is not sent while enabling it would wipe
// all load balancer configurations. InnerXML type fields are used with struct tag `innerxml` to
// prevent any manipulation of configuration and sending it verbatim
type LbGeneralParamsWithXml struct {
	XMLName             xml.Name   `xml:"loadBalancer"`
	Enabled             bool       `xml:"enabled"`
	AccelerationEnabled bool       `xml:"accelerationEnabled"`
	Logging             *LbLogging `xml:"logging"`

	// This field is not used anywhere but needs to be passed through
	EnableServiceInsertion bool `xml:"enableServiceInsertion"`
	// Each configuration change has a version number
	Version string `xml:"version,omitempty"`

	// The below fields have `innerxml` tag so that they are not processed but instead
	// sent verbatim
	VirtualServers []InnerXML `xml:"virtualServer,omitempty"`
	Pools          []InnerXML `xml:"pool,omitempty"`
	AppProfiles    []InnerXML `xml:"applicationProfile,omitempty"`
	Monitors       []InnerXML `xml:"monitor,omitempty"`
	AppRules       []InnerXML `xml:"applicationRule,omitempty"`
}

// LbLogging represents logging configuration for load balancer
type LbLogging struct {
	Enable   bool   `xml:"enable"`
	LogLevel string `xml:"logLevel"`
}

// InnerXML is meant to be used when unmarshaling a field into text rather than struct
// It helps to avoid missing out any fields which may not have been specified in the struct.
type InnerXML struct {
	Text string `xml:",innerxml"`
}

// LbMonitor defines health check parameters for a particular type of network traffic
// Reference: vCloud Director API for NSX Programming Guide
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type LbMonitor struct {
	XMLName    xml.Name `xml:"monitor"`
	ID         string   `xml:"monitorId,omitempty"`
	Type       string   `xml:"type"`
	Interval   int      `xml:"interval,omitempty"`
	Timeout    int      `xml:"timeout,omitempty"`
	MaxRetries int      `xml:"maxRetries,omitempty"`
	Method     string   `xml:"method,omitempty"`
	URL        string   `xml:"url,omitempty"`
	Expected   string   `xml:"expected,omitempty"`
	Name       string   `xml:"name,omitempty"`
	Send       string   `xml:"send,omitempty"`
	Receive    string   `xml:"receive,omitempty"`
	Extension  string   `xml:"extension,omitempty"`
}

type LbMonitors []LbMonitor

// LbPool represents a load balancer server pool as per "vCloud Director API for NSX Programming Guide"
// Type: LBPoolHealthCheckType
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type LbPool struct {
	XMLName             xml.Name      `xml:"pool"`
	ID                  string        `xml:"poolId,omitempty"`
	Name                string        `xml:"name"`
	Description         string        `xml:"description,omitempty"`
	Algorithm           string        `xml:"algorithm"`
	AlgorithmParameters string        `xml:"algorithmParameters,omitempty"`
	Transparent         bool          `xml:"transparent"`
	MonitorId           string        `xml:"monitorId,omitempty"`
	Members             LbPoolMembers `xml:"member,omitempty"`
}

type LbPools []LbPool

// LbPoolMember represents a single member inside LbPool
type LbPoolMember struct {
	ID          string `xml:"memberId,omitempty"`
	Name        string `xml:"name"`
	IpAddress   string `xml:"ipAddress"`
	Weight      int    `xml:"weight,omitempty"`
	MonitorPort int    `xml:"monitorPort,omitempty"`
	Port        int    `xml:"port"`
	MaxConn     int    `xml:"maxConn,omitempty"`
	MinConn     int    `xml:"minConn,omitempty"`
	Condition   string `xml:"condition,omitempty"`
}

type LbPoolMembers []LbPoolMember

// LbAppProfile represents a load balancer application profile as per "vCloud Director API for NSX
// Programming Guide"
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type LbAppProfile struct {
	XMLName                       xml.Name                  `xml:"applicationProfile"`
	ID                            string                    `xml:"applicationProfileId,omitempty"`
	Name                          string                    `xml:"name,omitempty"`
	SslPassthrough                bool                      `xml:"sslPassthrough"`
	Template                      string                    `xml:"template,omitempty"`
	HttpRedirect                  *LbAppProfileHttpRedirect `xml:"httpRedirect,omitempty"`
	Persistence                   *LbAppProfilePersistence  `xml:"persistence,omitempty"`
	InsertXForwardedForHttpHeader bool                      `xml:"insertXForwardedFor"`
	ServerSslEnabled              bool                      `xml:"serverSslEnabled"`
}

type LbAppProfiles []LbAppProfile

// LbAppProfilePersistence defines persistence profile settings in LbAppProfile
type LbAppProfilePersistence struct {
	XMLName    xml.Name `xml:"persistence"`
	Method     string   `xml:"method,omitempty"`
	CookieName string   `xml:"cookieName,omitempty"`
	CookieMode string   `xml:"cookieMode,omitempty"`
	Expire     int      `xml:"expire,omitempty"`
}

// LbAppProfileHttpRedirect defines http redirect settings in LbAppProfile
type LbAppProfileHttpRedirect struct {
	XMLName xml.Name `xml:"httpRedirect"`
	To      string   `xml:"to,omitempty"`
}

// LbAppRule represents a load balancer application rule as per "vCloud Director API for NSX
// Programming Guide"
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type LbAppRule struct {
	XMLName xml.Name `xml:"applicationRule"`
	ID      string   `xml:"applicationRuleId,omitempty"`
	Name    string   `xml:"name,omitempty"`
	Script  string   `xml:"script,omitempty"`
}

type LbAppRules []LbAppRule

// LbVirtualServer represents a load balancer virtual server as per "vCloud Director API for NSX
// Programming Guide"
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type LbVirtualServer struct {
	XMLName              xml.Name `xml:"virtualServer"`
	ID                   string   `xml:"virtualServerId,omitempty"`
	Name                 string   `xml:"name,omitempty"`
	Description          string   `xml:"description,omitempty"`
	Enabled              bool     `xml:"enabled"`
	IpAddress            string   `xml:"ipAddress"`
	Protocol             string   `xml:"protocol"`
	Port                 int      `xml:"port"`
	AccelerationEnabled  bool     `xml:"accelerationEnabled"`
	ConnectionLimit      int      `xml:"connectionLimit,omitempty"`
	ConnectionRateLimit  int      `xml:"connectionRateLimit,omitempty"`
	ApplicationProfileId string   `xml:"applicationProfileId,omitempty"`
	DefaultPoolId        string   `xml:"defaultPoolId,omitempty"`
	ApplicationRuleIds   []string `xml:"applicationRuleId,omitempty"`
}

// EdgeNatRule contains shared structure for SNAT and DNAT rule configuration using
// NSX-V proxied edge gateway endpoint
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type EdgeNatRule struct {
	XMLName           xml.Name `xml:"natRule"`
	ID                string   `xml:"ruleId,omitempty"`
	RuleType          string   `xml:"ruleType,omitempty"`
	RuleTag           string   `xml:"ruleTag,omitempty"`
	Action            string   `xml:"action"`
	Vnic              *int     `xml:"vnic,omitempty"`
	OriginalAddress   string   `xml:"originalAddress"`
	TranslatedAddress string   `xml:"translatedAddress"`
	LoggingEnabled    bool     `xml:"loggingEnabled"`
	Enabled           bool     `xml:"enabled"`
	Description       string   `xml:"description,omitempty"`
	Protocol          string   `xml:"protocol,omitempty"`
	OriginalPort      string   `xml:"originalPort,omitempty"`
	TranslatedPort    string   `xml:"translatedPort,omitempty"`
	IcmpType          string   `xml:"icmpType,omitempty"`
}

// EdgeFirewall holds data for creating firewall rule using proxied NSX-V API
// https://code.vmware.com/docs/6900/vcloud-director-api-for-nsx-programming-guide
type EdgeFirewallRule struct {
	XMLName         xml.Name                `xml:"firewallRule" `
	ID              string                  `xml:"id,omitempty"`
	Name            string                  `xml:"name,omitempty"`
	RuleType        string                  `xml:"ruleType,omitempty"`
	RuleTag         string                  `xml:"ruleTag,omitempty"`
	Source          EdgeFirewallEndpoint    `xml:"source" `
	Destination     EdgeFirewallEndpoint    `xml:"destination"`
	Application     EdgeFirewallApplication `xml:"application"`
	MatchTranslated *bool                   `xml:"matchTranslated,omitempty"`
	Direction       string                  `xml:"direction,omitempty"`
	Action          string                  `xml:"action,omitempty"`
	Enabled         bool                    `xml:"enabled"`
	LoggingEnabled  bool                    `xml:"loggingEnabled"`
}

// EdgeFirewallEndpoint can contains slices of objects for source or destination in EdgeFirewall
type EdgeFirewallEndpoint struct {
	Exclude           bool     `xml:"exclude"`
	VnicGroupIds      []string `xml:"vnicGroupId,omitempty"`
	GroupingObjectIds []string `xml:"groupingObjectId,omitempty"`
	IpAddresses       []string `xml:"ipAddress,omitempty"`
}

// EdgeFirewallApplication Wraps []EdgeFirewallApplicationService for multiple protocol/port specification
type EdgeFirewallApplication struct {
	ID       string                           `xml:"applicationId,omitempty"`
	Services []EdgeFirewallApplicationService `xml:"service,omitempty"`
}

// EdgeFirewallApplicationService defines port/protocol details for one service in EdgeFirewallRule
type EdgeFirewallApplicationService struct {
	Protocol   string `xml:"protocol,omitempty"`
	Port       string `xml:"port,omitempty"`
	SourcePort string `xml:"sourcePort,omitempty"`
}

// EdgeIpSet defines a group of IP addresses that you can add as the source or destination in a
// firewall rule or in DHCP relay configuration. The object itself has more fields in API response,
// however vCD UI only uses the below mentioned. It looks as if the other fields are used in NSX
// internally and are simply proxied back.
//
// Note. Only advanced edge gateways support IP sets
type EdgeIpSet struct {
	XMLName xml.Name `xml:"ipset"`
	// ID holds composite ID of IP set which is formatted as
	// 'f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-4' where the first segment before colon is vDC id
	// and the second one is IP set ID
	ID string `xml:"objectId,omitempty"`
	// Name is mandatory and must be unique
	Name string `xml:"name"`
	// Description - optional
	Description string `xml:"description,omitempty"`
	// IPAddresses is a mandatory field with comma separated values. The API is known to re-order
	// data after submiting and may shuffle components even if re-submitted as it was return from
	// API itself
	// (eg: "192.168.200.1,192.168.200.1/24,192.168.200.1-192.168.200.24")
	IPAddresses string `xml:"value"`
	// InheritanceAllowed defines visibility at underlying scopes
	InheritanceAllowed *bool `xml:"inheritanceAllowed"`
	// Revision is a "version" of IP set configuration. During read current revision is being
	// returned and when update is performed this latest version must be sent as it validates if no
	// updates ocurred in between. When not the latest version is being sent during update one can
	// expect similar error response from API: "The object ipset-27 used in this operation has an
	// older version 0 than the current system version 1. Refresh UI or fetch the latest copy of the
	// object and retry operation."
	Revision *int `xml:"revision,omitempty"`
}

// EdgeIpSets is a slice of pointers to EdgeIpSet
type EdgeIpSets []*EdgeIpSet

// EdgeGatewayVnics is a data structure holding information of vNic configuration in NSX-V edge
// gateway using "/network/edges/edge_id/vnics" endpoint
type EdgeGatewayVnics struct {
	XMLName xml.Name `xml:"vnics"`
	Vnic    []struct {
		Label         string `xml:"label"`
		Name          string `xml:"name"`
		AddressGroups struct {
			AddressGroup struct {
				PrimaryAddress     string `xml:"primaryAddress,omitempty"`
				SecondaryAddresses struct {
					IpAddress []string `xml:"ipAddress,omitempty"`
				} `xml:"secondaryAddresses,omitempty"`
				SubnetMask         string `xml:"subnetMask,omitempty"`
				SubnetPrefixLength string `xml:"subnetPrefixLength,omitempty"`
			} `xml:"addressGroup,omitempty"`
		} `xml:"addressGroups,omitempty"`
		Mtu                 string `xml:"mtu,omitempty"`
		Type                string `xml:"type,omitempty"`
		IsConnected         string `xml:"isConnected,omitempty"`
		Index               *int   `xml:"index"`
		PortgroupId         string `xml:"portgroupId,omitempty"`
		PortgroupName       string `xml:"portgroupName,omitempty"`
		EnableProxyArp      string `xml:"enableProxyArp,omitempty"`
		EnableSendRedirects string `xml:"enableSendRedirects,omitempty"`
		SubInterfaces       struct {
			SubInterface []struct {
				IsConnected         string `xml:"isConnected,omitempty"`
				Label               string `xml:"label,omitempty"`
				Name                string `xml:"name,omitempty"`
				Index               *int   `xml:"index,omitempty"`
				TunnelId            string `xml:"tunnelId,omitempty"`
				LogicalSwitchId     string `xml:"logicalSwitchId,omitempty"`
				LogicalSwitchName   string `xml:"logicalSwitchName,omitempty"`
				EnableSendRedirects string `xml:"enableSendRedirects,omitempty"`
				Mtu                 string `xml:"mtu,omitempty"`
				AddressGroups       struct {
					AddressGroup struct {
						PrimaryAddress     string `xml:"primaryAddress,omitempty"`
						SubnetMask         string `xml:"subnetMask,omitempty"`
						SubnetPrefixLength string `xml:"subnetPrefixLength,omitempty"`
					} `xml:"addressGroup,omitempty"`
				} `xml:"addressGroups,omitempty"`
			} `xml:"subInterface,omitempty"`
		} `xml:"subInterfaces,omitempty"`
	} `xml:"vnic,omitempty"`
}

// EdgeGatewayInterfaces is a data structure holding information of vNic configuration in NSX-V edge
// gateway using "/network/edges/edge_id/vdcNetworks" endpoint
type EdgeGatewayInterfaces struct {
	XMLName       xml.Name `xml:"edgeInterfaces"`
	EdgeInterface []struct {
		Name             string `xml:"name"`
		Type             string `xml:"type"`
		Index            *int   `xml:"index"`
		NetworkReference struct {
			ID   string `xml:"id"`
			Name string `xml:"name"`
			Type string `xml:"type"`
		} `xml:"networkReference"`
		AddressGroups struct {
			AddressGroup struct {
				PrimaryAddress     string `xml:"primaryAddress"`
				SubnetMask         string `xml:"subnetMask"`
				SubnetPrefixLength string `xml:"subnetPrefixLength"`
				SecondaryAddresses struct {
					IpAddress []string `xml:"ipAddress"`
				} `xml:"secondaryAddresses"`
			} `xml:"addressGroup"`
		} `xml:"addressGroups"`
		PortgroupId   string `xml:"portgroupId"`
		PortgroupName string `xml:"portgroupName"`
		IsConnected   string `xml:"isConnected"`
		TunnelId      string `xml:"tunnelId"`
	} `xml:"edgeInterface"`
}

// EdgeDhcpRelay - Dynamic Host Configuration Protocol (DHCP) relay enables you to leverage your
// existing DHCP infrastructure from within NSX without any interruption to the IP address
// management in your environment. DHCP messages are relayed from virtual machine(s) to the
// designated DHCP server(s) in the physical world. This enables IP addresses within NSX to continue
// to be in sync with IP addresses in other environments.
type EdgeDhcpRelay struct {
	XMLName xml.Name `xml:"relay"`
	// RelayServer specifies external relay server(s) to which DHCP messages are to be relayed to.
	// The relay server can be an IP set, IP address block, domain, or a combination of all of
	// these. Messages are relayed to each listed DHCP server.
	RelayServer *EdgeDhcpRelayServer `xml:"relayServer"`
	// EdgeDhcRelayAgents  specifies a list of edge gateway interfaces (vNics) from which DHCP
	// messages are to be relayed to the external DHCP relay server(s) with optional gateway
	// interface addresses.
	RelayAgents *EdgeDhcpRelayAgents `xml:"relayAgents"`
}

type EdgeDhcpRelayServer struct {
	// GroupingObjectIds is a general concept in NSX which allows to pass in many types of objects
	// (like VM IDs, IP set IDs, org networks, security groups) howether in this case it accepts
	// only IP sets which have IDs specified as 'f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-2' where
	// first part is vDC ID and the second part is unique IP set ID
	GroupingObjectId []string `xml:"groupingObjectId,omitempty"`
	// IpAddresses holds a list of IP addresses for DHCP servers
	IpAddress []string `xml:"ipAddress,omitempty"`
	// Fqdn holds a list of FQDNs (fully qualified domain names)
	Fqdns []string `xml:"fqdn,omitempty"`
}

// EdgeDhcpRelayAgent specifies which edge gateway interface (vNic) from which DHCP messages are to
// be relayed to the external DHCP relay server(s) with an optional gateway interface address.
type EdgeDhcpRelayAgent struct {
	// VnicIndex must specify vNic adapter index on the edge gateway
	VnicIndex *int `xml:"vnicIndex"`
	// GatewayInterfaceAddress holds a gateway interface address. Optional, defaults to the vNic
	// primary address.
	GatewayInterfaceAddress string `xml:"giAddress,omitempty"`
}

// EdgeDhcpRelayAgents holds a slice of EdgeDhcpRelayAgent
type EdgeDhcpRelayAgents struct {
	Agents []EdgeDhcpRelayAgent `xml:"relayAgent"`
}

// EdgeDhcpLease holds a list of EdgeDhcpLeaseInfo
type EdgeDhcpLease struct {
	XMLName        xml.Name             `xml:"dhcpLeaseInfo"`
	DhcpLeaseInfos []*EdgeDhcpLeaseInfo `xml:"leaseInfo"`
}

// EdgeDhcpLeaseInfo contains information about DHCP leases provided by NSX-V edge gateway
type EdgeDhcpLeaseInfo struct {
	// Uid statement records the client identifier used by the client to acquire the lease. Clients
	// are not required to send client identifiers, and this statement only appears if the client
	// did in fact send one.
	Uid string `xml:"uid"`
	// MacAddress holds hardware (MAC) address of requester (e.g. "00:50:56:01:29:c8")
	MacAddress string `xml:"macAddress"`
	// IpAddress holds the IP address assigned to a particular MAC address (e.g. "10.10.10.100")
	IpAddress string `xml:"ipAddress"`
	// ClientHostname Most DHCP clients will send their hostname in the host-name option. If a
	// client sends its hostname in this way, the hostname is recorded on the lease with a
	// client-hostname statement. This is not required by the protocol, however, so many specialized
	// DHCP clients do not send a host-name option.
	ClientHostname string `xml:"clientHostname"`
	// BindingState declares the lease’s binding state. When the DHCP server is not configured to
	// use the failover protocol, a lease’s binding state may be active, free or abandoned. The
	// failover protocol adds some additional transitional states, as well as the backup state,
	// which indicates that the lease is available for allocation by the failover secondary
	BindingState string `xml:"bindingState"`
	// NextBindingState statement indicates what state the lease will move to when the current state
	// expires. The time when the current state expires is specified in the ends statement.
	NextBindingState string `xml:"nextBindingState"`
	// Cltt holds value of clients last transaction time (format is "weekday year/month/day
	// hour:minute:second", e.g. "2 2019/12/17 06:12:03")
	Cltt string `xml:"cltt"`
	// Starts holds the start time of a lease (format is "weekday year/month/day
	// hour:minute:second", e.g. "2 2019/12/17 06:12:03")
	Starts string `xml:"starts"`
	// Ends holds the end time of a lease (format is "weekday year/month/day hour:minute:second",
	// e.g. "3 2019/12/18 06:12:03")
	Ends string `xml:"ends"`
	// HardwareType holds ... (e.g. "ethernet")
	HardwareType string `xml:"hardwareType"`
}
