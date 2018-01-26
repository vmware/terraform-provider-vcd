package vcd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	types "github.com/ukcloud/govcloudair/types/v56"
)

func expandIPRange(configured []interface{}) types.IPRanges {
	ipRange := make([]*types.IPRange, 0, len(configured))

	for _, ipRaw := range configured {
		data := ipRaw.(map[string]interface{})

		ip := types.IPRange{
			StartAddress: data["start_address"].(string),
			EndAddress:   data["end_address"].(string),
		}

		ipRange = append(ipRange, &ip)
	}

	ipRanges := types.IPRanges{
		IPRange: ipRange,
	}

	return ipRanges
}

func expandFirewallRules(d *schema.ResourceData, gateway *types.EdgeGateway) ([]*types.FirewallRule, error) {
	//firewallRules := make([]*types.FirewallRule, 0, len(configured))
	firewallRules := gateway.Configuration.EdgeGatewayServiceConfiguration.FirewallService.FirewallRule

	rulesCount := d.Get("rule.#").(int)
	for i := 0; i < rulesCount; i++ {
		prefix := fmt.Sprintf("rule.%d", i)

		var protocol *types.FirewallRuleProtocols
		switch d.Get(prefix + ".protocol").(string) {
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

func retryCall(seconds int, f resource.RetryFunc) error {
	return resource.Retry(time.Duration(seconds)*time.Second, f)
}

func convertToStringMap(param map[string]interface{}) map[string]string {
	temp := make(map[string]string)
	for k, v := range param {
		temp[k] = v.(string)
	}
	return temp
}

func interfaceListToMapStringInterface(list []interface{}) []map[string]interface{} {
	newList := make([]map[string]interface{}, len(list))

	for index, item := range list {
		newList[index] = item.(map[string]interface{})
	}

	return newList
}

func interfaceListToStringList(list []interface{}) []string {
	newList := make([]string, len(list))

	for index, item := range list {
		newList[index] = item.(string)
	}

	return newList
}

// Methods borrowed from the vSphere provider project

// DeRef returns the value pointed to by the interface if the interface is a
// pointer and is not nil, otherwise returns nil, or the direct value if it's
// not a pointer.
func DeRef(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	k := reflect.TypeOf(v).Kind()
	if k != reflect.Ptr {
		return v
	}
	if reflect.ValueOf(v) == reflect.Zero(reflect.TypeOf(v)) {
		// All zero-value pointers are nil
		return nil
	}
	return reflect.ValueOf(v).Elem().Interface()
}

// NormalizeValue converts a value to something that is suitable to be set in a
// ResourceData and can be useful in situations where there is not access to
// normal helper/schema functionality, but you still need saved fields to
// behave in the same way.
//
// Specifically, this will run the value through DeRef to dereference any
// pointers first, and then convert numeric primitives, if necessary.
func NormalizeValue(v interface{}) interface{} {
	v = DeRef(v)
	if v == nil {
		return nil
	}
	k := reflect.TypeOf(v).Kind()
	switch {
	case k >= reflect.Int8 && k <= reflect.Uint64:
		v = reflect.ValueOf(v).Convert(reflect.TypeOf(int(0))).Interface()
	case k == reflect.Float32:
		v = reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Interface()
	}
	return v
}

func init() {
	gob.Register(map[string]interface{}{})
}

// Map performs a deep copy of the given map m.
func deepCopyMap(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}
