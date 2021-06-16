package vcd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func convertToStringMap(param map[string]interface{}) map[string]string {
	temp := make(map[string]string)
	for k, v := range param {
		temp[k] = v.(string)
	}
	return temp
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

func convertStringsToInterfaceSlice(param []string) []interface{} {
	slice := make([]interface{}, len(param))
	for index, value := range param {
		slice[index] = value
	}
	return slice
}

// convertStringsTotTypeSet accepts a slice of strings and returns a *schema.Set suitable for storing in Terraform
// set of strings
func convertStringsTotTypeSet(slice []string) *schema.Set {
	sliceOfInterfaces := convertStringsToInterfaceSlice(slice)
	set := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), sliceOfInterfaces)

	return set
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
