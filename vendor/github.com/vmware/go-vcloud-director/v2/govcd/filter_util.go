package govcd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/araddon/dateparse"
)

// Names of the filters allowed in the search engine
const (
	FilterNameRegex = "name_regex" // a name, searched by regular expression
	FilterDate      = "date"       // a date expression (>|<|==|>=|<= date)
	FilterIp        = "ip"         // An IP, searched by regular expression
	FilterLatest    = "latest"     // gets the newest element
)

var (
	// Filters currently supported in the engine
	SupportedFilters = []string{FilterNameRegex, FilterDate, FilterIp, FilterLatest}

	// Metadata types recognized so far. "NONE" is the same as ""
	SupportedMetadataTypes = []string{"NONE", "STRING", "INT", "BOOL"}
)

// Definition of metadata structure
type MetadataDef struct {
	Key      string      // name of the field (addressed as metadata:key)
	Type     string      // Type of the field (one of SupportedMetadataTypes)
	Value    interface{} // contents of the metadata field
	IsSystem bool        // if true, the metadata field will be addressed as metadata@SYSTEM:key
}

// The definition of all the criteria used by the engine to retrieve data
type FilterDef struct {
	// A collection of filters (with keys from SupportedFilters)
	Filters              map[string]string

	// A list of metadata filters
	Metadata             []MetadataDef

	// If true, the query will include metadata fields and search for exact values.
	// Otherwise, the engine will collect metadata fields and search by regexp
	UseMetadataApiFilter bool
}

// Creates a new filter definition
func NewFilterDef() *FilterDef {
	return &FilterDef{
		Filters:  make(map[string]string),
		Metadata: nil,
	}
}

// AddFilter adds a new filter to the criteria
func (fd *FilterDef) AddFilter(key, value string) error {
	for _, allowed := range SupportedFilters {
		if key == allowed {
			fd.Filters[key] = value
			return nil
		}
	}
	return fmt.Errorf("filter '%s' not supported", key)
}

// AddMetadataFilter adds a new metadata filter to an existing set
func (fd *FilterDef) AddMetadataFilter(key, value, valueType string, isSystem, useMetadataApiFilter bool) error {
	if valueType == "" {
		valueType = "NONE"
		useMetadataApiFilter = false
	}
	if useMetadataApiFilter {
		fd.UseMetadataApiFilter = true
	}
	typeSupported := false
	for _, supported := range SupportedMetadataTypes {
		if valueType == supported {
			typeSupported = true
		}
	}
	if !typeSupported {
		return fmt.Errorf("metadata type '%s' not supported", valueType)
	}
	fd.Metadata = append(fd.Metadata, MetadataDef{
		Key:      key,
		Value:    value,
		IsSystem: isSystem,
		Type:     valueType,
	})
	return nil
}

// StringToBool converts a string to a bool
// The following values are recognized as TRUE:
//  t, true, y, yes, ok
func StringToBool(s string) bool {
	switch strings.ToLower(s) {
	case "t", "true", "y", "yes", "ok":
		return true
	default:
		return false
	}
}

// compareDate will get a date from string `got`, and will parse `wanted`
// for an expression containing an operator (>, <, >=, <=, ==) and a date
// (many formats supported, but 'YYYY-MM-DD[ hh:mm[:ss]]' preferred)
// For example:
// got:    "2020-03-09T09:50:51.500Z"
// wanted: ">= 2020-03-08"
// result: true
// got:    "2020-03-09T09:50:51.500Z"
// wanted: "< 02-mar-2020"
// result: false
// See https://github.com/araddon/dateparse for more info
func compareDate(wanted, got string) (bool, error) {

	reExpression := regexp.MustCompile(`(>=|<=|==|<|=|>)\s*(.+)`)

	expList := reExpression.FindAllStringSubmatch(wanted, -1)
	if len(expList) == 0 || len(expList[0]) == 0 {
		return false, fmt.Errorf("expression not found in '%s'", wanted)
	}

	operator := expList[0][1]
	wantedTime, err := dateparse.ParseStrict(expList[0][2])
	if err != nil {
		return false, err
	}

	gotTime, err := dateparse.ParseStrict(got)
	if err != nil {
		return false, err
	}

	wantedSeconds := wantedTime.Unix()
	gotSeconds := gotTime.Unix()

	switch operator {
	case "=", "==":
		return gotSeconds == wantedSeconds, nil
	case ">":
		return gotSeconds > wantedSeconds, nil
	case ">=":
		return gotSeconds >= wantedSeconds, nil
	case "<=":
		return gotSeconds <= wantedSeconds, nil
	case "<":
		return gotSeconds < wantedSeconds, nil
	default:
		return false, fmt.Errorf("unsupported operator '%s'", operator)
	}
}

// conditionText provides a human readable string of searching criteria
func conditionText(criteria *FilterDef) string {
	var result string

	for k, v := range criteria.Filters {
		result += fmt.Sprintf(`("%s" -> "%s") `, k, v)
	}
	for _, m := range criteria.Metadata {
		result += fmt.Sprintf(`m("%s" -> "%s")`, m.Key, m.Value)
	}
	return result
}

