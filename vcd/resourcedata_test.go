package vcd

import (
	"regexp"
	"strconv"
)

// schemaResourceData implements an interface vcdResourceDataInterface and allows to mimic Terraform's native
// schema.Resourcedata (usually consumed as 'd' variable). This allows to reuse functions created for Terraform
// to work properly. One of such cases is the need to be able to delete an existing object.
type schemaResourceData struct {
	id         string
	org        string
	vdc        string
	configText string
}

// Get is a copy of schema.Resourcedata.Get
func (d schemaResourceData) Get(key string) interface{} {
	if key == "id" {
		return d.id
	}

	if key == "org" {
		return d.org
	}

	if key == "vdc" {
		return d.vdc
	}

	// When the above specific fields are covered, let's use a generic function which looks for a field value inside
	// configText

	// This regex extract HCLs key = value (without quotes if they were present in definition)
	// In the below example configText it would extract unquoted 'key1' for match1 and 'value1' for match2
	// resource "vcd_catalog" "catalog" {
	//    key1 = value1
	//    "key1" = value1
	//    key1 = "value1"
	//    "key1" = value1
	// }
	re := regexp.MustCompile(`\s*"?([a-zA-Z\-_.]+)"?\s+=\s+"?([a-zA-Z\-_.]+)"?`)

	matches := re.FindAllStringSubmatch(d.configText, -1)
	for _, match := range matches {
		if match[1] == key {

			// For boolean type variables value must be returned as boolean
			boolValue, err := strconv.ParseBool(match[2])
			if err == nil {
				return boolValue
			}

			return match[2]
		}
	}

	return ""
}

// SetId is a copy of schema.Resourcedata.SetId
func (d schemaResourceData) SetId(v string) {
	d.id = v
}

// Id is a copy of schema.Resourcedata.Id
func (d schemaResourceData) Id() string {
	return d.id
}
