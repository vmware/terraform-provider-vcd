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
	FilterEarliest  = "earliest"   // gets the oldest element
	FilterParent    = "parent"     // matches the entity parent
	FilterParentId  = "parent_id"  // matches the entity parent ID
)

var (
	// Filters currently supported in the engine, available to users
	supportedFilters = []string{
		FilterNameRegex,
		FilterDate,
		FilterIp,
		FilterLatest,
		FilterEarliest,
		FilterParent,
		FilterParentId,
	}

	// Metadata types recognized so far. "NONE" is the same as ""
	SupportedMetadataTypes = []string{"NONE", "STRING", "NUMBER", "BOOLEAN", "DATETIME"}

	// retrievedMetadataTypes maps the internal value of metadata type with the
	// string needed when searching for a metadata field in the API
	retrievedMetadataTypes = map[string]string{
		"MetadataBooleanValue":  "BOOLEAN",
		"MetadataStringValue":   "STRING",
		"MetadataNumberValue":   "NUMBER",
		"MetadataDateTimeValue": "STRING", // values for DATETIME can't be passed as such in a query.
	}
)

// Definition of metadata structure
type MetadataDef struct {
	Key      string      // name of the field (addressed as metadata:key)
	Type     string      // Type of the field (one of SupportedMetadataTypes)
	Value    interface{} // contents of the metadata field
	IsSystem bool        // if true, the metadata field will be addressed as metadata@SYSTEM:key
}

// matchResult stores the result of a condition evaluation
// Used to build the human readable description of the engine operations
type matchResult struct {
	Name       string
	Type       string
	Definition string
	Result     bool
}

// The definition of all the criteria used by the engine to retrieve data
type FilterDef struct {
	// A collection of filters (with keys from SupportedFilters)
	Filters map[string]string

	// A list of metadata filters
	Metadata []MetadataDef

	// If true, the query will include metadata fields and search for exact values.
	// Otherwise, the engine will collect metadata fields and search by regexp
	UseMetadataApiFilter bool
}

// NewFilterDef builds a new filter definition
func NewFilterDef() *FilterDef {
	return &FilterDef{
		Filters:  make(map[string]string),
		Metadata: nil,
	}
}

// validateMetadataType checks that a metadata type is within supported types
func validateMetadataType(valueType string) error {
	typeSupported := false
	for _, supported := range SupportedMetadataTypes {
		if valueType == supported {
			typeSupported = true
		}
	}
	if !typeSupported {
		return fmt.Errorf("metadata type '%s' not supported", valueType)
	}
	return nil
}

// AddFilter adds a new filter to the criteria
func (fd *FilterDef) AddFilter(key, value string) error {
	for _, allowed := range supportedFilters {
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
	err := validateMetadataType(valueType)
	if err != nil {
		return err
	}
	fd.Metadata = append(fd.Metadata, MetadataDef{
		Key:      key,
		Value:    value,
		IsSystem: isSystem,
		Type:     valueType,
	})
	return nil
}

// stringToBool converts a string to a bool
// The following values are recognized as TRUE:
//  t, true, y, yes, ok
func stringToBool(s string) bool {
	switch strings.ToLower(s) {
	case "t", "true", "y", "yes", "ok":
		return true
	default:
		return false
	}
}

// CompareDate will get a date from string `got`, and will parse `wanted`
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
func CompareDate(wanted, got string) (bool, error) {

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

	wantedSeconds := wantedTime.UnixNano()
	gotSeconds := gotTime.UnixNano()

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
	result := "criteria: "

	for k, v := range criteria.Filters {
		result += fmt.Sprintf(`("%s" -> "%s") `, k, v)
	}
	for _, m := range criteria.Metadata {
		marker := "meta"
		if criteria.UseMetadataApiFilter {
			marker = "metaApi"
		}
		result += fmt.Sprintf(`%s("%s" -> "%s") `, marker, m.Key, m.Value)
	}
	return result
}

// matchesToText provides a human readable string of search operations results
func matchesToText(matches []matchResult) string {
	result := ""
	for _, item := range matches {
		result += fmt.Sprintf("name: %s; type: %s definition: %s; result: %v\n", item.Name, item.Type, item.Definition, item.Result)
	}
	return result
}
