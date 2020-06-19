package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func expandIPRange(configured []interface{}) (types.IPRanges, error) {
	ipRange := make([]*types.IPRange, 0, len(configured))

	for _, ipRaw := range configured {
		data := ipRaw.(map[string]interface{})

		startAddress := data["start_address"].(string)
		endAddress := data["end_address"].(string)
		ip := types.IPRange{
			StartAddress: startAddress,
			EndAddress:   endAddress,
		}

		ipRange = append(ipRange, &ip)
	}

	ipRanges := types.IPRanges{
		IPRange: ipRange,
	}

	return ipRanges, nil
}

func expandFirewallRules(d *schema.ResourceData, gateway *types.EdgeGateway) ([]*types.FirewallRule, error) {
	//firewallRules := make([]*types.FirewallRule, 0, len(configured))
	firewallRules := gateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule

	rulesCount := d.Get("rule.#").(int)
	for i := 0; i < rulesCount; i++ {
		prefix := fmt.Sprintf("rule.%d", i)

		var protocol *types.FirewallRuleProtocols
		// Allow upper and lower case protocol names
		switch strings.ToLower(d.Get(prefix + ".protocol").(string)) {
		case "tcp":
			protocol = &types.FirewallRuleProtocols{
				TCP: true,
			}
		case "udp":
			protocol = &types.FirewallRuleProtocols{
				UDP: true,
			}
		case "icmp":
			protocol = &types.FirewallRuleProtocols{
				ICMP: true,
			}
		default:
			protocol = &types.FirewallRuleProtocols{
				Any: true,
			}
		}
		rule := &types.FirewallRule{
			//ID: strconv.Itoa(len(configured) - i),
			IsEnabled:            true,
			MatchOnTranslate:     false,
			Description:          d.Get(prefix + ".description").(string),
			Policy:               d.Get(prefix + ".policy").(string),
			Protocols:            protocol,
			Port:                 getNumericPort(d.Get(prefix + ".destination_port")),
			DestinationPortRange: d.Get(prefix + ".destination_port").(string),
			DestinationIP:        d.Get(prefix + ".destination_ip").(string),
			SourcePort:           getNumericPort(d.Get(prefix + ".source_port")),
			SourcePortRange:      d.Get(prefix + ".source_port").(string),
			SourceIP:             d.Get(prefix + ".source_ip").(string),
			EnableLogging:        false,
		}
		firewallRules = append(firewallRules, rule)
	}

	return firewallRules, nil
}

func getProtocol(protocol types.FirewallRuleProtocols) string {
	if protocol.TCP {
		return "tcp"
	}
	if protocol.TCP && protocol.UDP {
		return "tcp&udp"
	}
	if protocol.UDP {
		return "udp"
	}
	if protocol.ICMP {
		return "icmp"
	}
	return "any"
}

func getNumericPort(portrange interface{}) int {
	i, err := strconv.Atoi(portrange.(string))
	if err != nil {
		return -1
	}
	return i
}

func getPortString(port int) string {
	if port == -1 {
		return "any"
	}
	portstring := strconv.Itoa(port)
	return portstring
}

func convertToStringMap(param map[string]interface{}) map[string]string {
	temp := make(map[string]string)
	for k, v := range param {
		temp[k] = v.(string)
	}
	return temp
}

func convertToStringSlice(sliceOfInterfaces []interface{}) []string {
	var result []string
	for _, en := range sliceOfInterfaces {
		result = append(result, en.(string))
	}
	return result
}

// convertSchemaSetToSliceOfStrings accepts Terraform's *schema.Set object and converts it to slice
// of strings.
// This is useful for extracting values from a set of strings
func convertSchemaSetToSliceOfStrings(param *schema.Set) []string {
	paramList := param.List()
	result := make([]string, len(paramList))
	for index, value := range paramList {
		result[index] = fmt.Sprint(value)
	}

	return result
}

func convertToTypeSet(param []string) []interface{} {
	slice := make([]interface{}, len(param))
	for index, value := range param {
		slice[index] = value
	}
	return slice
}

// takeBoolPointer accepts a boolean and returns a pointer to this value.
func takeBoolPointer(value bool) *bool {
	return &value
}

// takeIntPointer accepts an int and returns a pointer to this value.
func takeIntPointer(x int) *int {
	return &x
}

// extractUuid finds an UUID in the input string
// Returns an empty string if no UUID was found
func extractUuid(input string) string {
	reGetID := regexp.MustCompile(`([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`)
	matchListId := reGetID.FindAllStringSubmatch(input, -1)
	if len(matchListId) > 0 && len(matchListId[0]) > 0 {
		return matchListId[0][1]
	}
	return ""
}

// normalizeId checks if the ID contains a wanted prefix
// If it does, the function returns the original ID.
// Otherwise, it returns the prefix + the ID
func normalizeId(prefix, id string) string {
	if strings.Contains(id, prefix) {
		return id
	}
	return prefix + id
}

// haveSameUuid compares two IDs (or HREF)
// and returns true if the UUID part of the two input strings are the same.
// This is useful when comparing a HREF to a ID, or a HREF from an admin path
// to a HREF from a regular user path.
func haveSameUuid(s1, s2 string) bool {
	return extractUuid(s1) == extractUuid(s2)
}
