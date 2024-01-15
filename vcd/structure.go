package vcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// getKeys retrieves all the keys from the given map and returns them as a slice
func getKeys[K comparable, V any](input map[K]V) []K {
	result := make([]K, len(input))
	i := 0
	for k := range input {
		result[i] = k
		i++
	}
	return result
}

// traverseMapAndGet traverses the input interface{}, which should be a map of maps, by following the path specified as
// "keyA.keyB.keyC.keyD", doing something similar to, visually speaking, map["keyA"]["keyB"]["keyC"]["keyD"], or in other words,
// it goes inside every inner map, which are inside the initial map, until the given path is finished.
// The final value, "keyD" in the same example, should be of type ResultType, which is a generic type requested during the call
// to this function.
func traverseMapAndGet[ResultType any](input interface{}, path string) (ResultType, error) {
	var nothing ResultType
	if input == nil {
		return nothing, fmt.Errorf("the input is nil")
	}
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nothing, fmt.Errorf("the input is a %T, not a map[string]interface{}", input)
	}
	if len(inputMap) == 0 {
		return nothing, fmt.Errorf("the map is empty")
	}
	pathUnits := strings.Split(path, ".")
	completed := false
	i := 0
	var result interface{}
	for !completed {
		subPath := pathUnits[i]
		traversed, ok := inputMap[subPath]
		if !ok {
			return nothing, fmt.Errorf("key '%s' does not exist in input map", subPath)
		}
		if i < len(pathUnits)-1 {
			traversedMap, ok := traversed.(map[string]interface{})
			if !ok {
				return nothing, fmt.Errorf("key '%s' is a %T, not a map, but there are still %d paths to explore", subPath, traversed, len(pathUnits)-(i+1))
			}
			inputMap = traversedMap
		} else {
			completed = true
			result = traversed
		}
		i++
	}
	resultTyped, ok := result.(ResultType)
	if !ok {
		return nothing, fmt.Errorf("could not convert obtained type %T to requested %T", result, nothing)
	}
	return resultTyped, nil
}

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

// filterVdcId returns a bare UUID if the initial value contains a VDC ID
// otherwise it returns the initial value
func filterVdcId(i interface{}) string {
	s := i.(string)
	if strings.HasPrefix(s, "urn:vcloud:vdc:") {
		return extractUuid(s)
	}
	return s
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

// convertTypeListToSliceOfStrings accepts Terraform's TypeList structure `[]interface{}` and
// converts it to slice of strings.
func convertTypeListToSliceOfStrings(param []interface{}) []string {
	result := make([]string, len(param))
	for i, v := range param {
		result[i] = v.(string)
	}
	return result
}

// convertStringsToTypeSet accepts a slice of strings and returns a *schema.Set suitable for storing in Terraform
// set of strings
func convertStringsToTypeSet(param []string) *schema.Set {
	sliceOfInterfaces := make([]interface{}, len(param))
	for index, value := range param {
		sliceOfInterfaces[index] = value
	}

	set := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), sliceOfInterfaces)
	return set
}

// addrOf is a wantType function to return the address of a variable
// Note. It is mainly meant for converting literal values to pointers (e.g. `addrOf(true)`) or cases
// for converting variables coming out straight from Terraform schema (e.g.
// `addrOf(d.Get("name").(string))`).
func addrOf[T any](variable T) *T {
	return &variable
}

// stringPtrOrNil takes a string and returns a pointer to it, but if the string is empty, returns nil
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// getStringAttributeAsPointer returns a pointer to the value of the given attribute from the current resource data.
// If the attribute is empty, returns a nil pointer.
func getStringAttributeAsPointer(d *schema.ResourceData, attrName string) *string {
	attributeValue := d.Get(attrName).(string)
	if attributeValue == "" {
		return nil
	}
	return &attributeValue
}

// getUuidRegex returns a regular expression that matches UUIDs used by VCD
func getUuidRegex(prefix, suffix string) *regexp.Regexp {
	uuidRegexFormat := fmt.Sprintf("%s[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}%s", prefix, suffix)
	return regexp.MustCompile(uuidRegexFormat)
}

// extractUuid finds an UUID in the input string
// Returns an empty string if no UUID was found
func extractUuid(input string) string {
	reGetID := getUuidRegex("(", ")")
	matchListIds := reGetID.FindAllStringSubmatch(input, -1)
	if len(matchListIds) > 0 && len(matchListIds[0]) > 0 {
		return matchListIds[len(matchListIds)-1][len(matchListIds[0])-1]
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

// extractIdsFromOpenApiReferences extracts []string with IDs from []types.OpenApiReference which contains ID and Names
func extractIdsFromOpenApiReferences(refs []types.OpenApiReference) []string {
	resultStrings := make([]string, len(refs))
	for index := range refs {
		resultStrings[index] = refs[index].ID
	}

	return resultStrings
}

// extractNamesFromOpenApiReferences extracts []string with names from []types.OpenApiReference which contains ID and Names
func extractNamesFromOpenApiReferences(refs []types.OpenApiReference) []string {
	resultStrings := make([]string, len(refs))
	for index := range refs {
		resultStrings[index] = refs[index].Name
	}

	return resultStrings
}

// extractIdsFromReferences extracts []string with IDs from []*types.Reference which contains ID and Names
func extractIdsFromReferences(refs []*types.Reference) []string {
	resultStrings := make([]string, len(refs))
	for index := range refs {
		resultStrings[index] = refs[index].ID
	}

	return resultStrings
}

// convertSliceOfStringsToOpenApiReferenceIds converts []string to []types.OpenApiReference by filling
// types.OpenApiReference.ID fields
func convertSliceOfStringsToOpenApiReferenceIds(ids []string) []types.OpenApiReference {
	resultReferences := make([]types.OpenApiReference, len(ids))
	for i, v := range ids {
		resultReferences[i].ID = v
	}

	return resultReferences
}

// contains returns true if `sliceToSearch` contains `searched`. Returns false otherwise.
func contains(sliceToSearch []string, searched string) bool {
	found := false
	for _, idInSlice := range sliceToSearch {
		if searched == idInSlice {
			found = true
			break
		}
	}
	return found
}

// jsonToCompactString transforms an unmarshalled JSON in form of a map of string->any to a plain string without any spacing.
func jsonToCompactString(inputJson map[string]interface{}) (string, error) {
	rawJson, err := json.Marshal(inputJson)
	if err != nil {
		return "", err
	}
	compactedJson := new(bytes.Buffer)
	err = json.Compact(compactedJson, rawJson)
	if err != nil {
		return "", err
	}
	return compactedJson.String(), nil
}

// areMarshaledJsonEqual compares that two marshaled JSON strings are equal or not. Returns an error if something
// wrong happens during the comparison process.
func areMarshaledJsonEqual(json1, json2 []byte) (bool, error) {
	if !json.Valid(json1) {
		return false, fmt.Errorf("first JSON is not valid: '%s'", json1)
	}
	if !json.Valid(json2) {
		return false, fmt.Errorf("second JSON is not valid: '%s'", json2)
	}

	var unmarshaledJson1, unmarshaledJson2 interface{}
	err := json.Unmarshal(json1, &unmarshaledJson1)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal first JSON '%s': %s", json1, err)
	}
	err = json.Unmarshal(json2, &unmarshaledJson2)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal second JSON '%s': %s", json2, err)
	}
	return reflect.DeepEqual(unmarshaledJson1, unmarshaledJson2), nil
}

// createOrUpdateMetadata creates or updates metadata entries for the given resource and attribute name
// TODO: This function implementation should be replaced with the implementation of `createOrUpdateMetadataEntryInVcd`
// once "metadata" field is removed.
func createOrUpdateMetadata(d *schema.ResourceData, resource metadataCompatible, attributeName string) error {
	// We invoke the new "metadata_entry" metadata creation here to have it centralized and reduce duplication.
	// Ideally, once "metadata" is removed in a new major version, the implementation of `createOrUpdateMetadataEntryInVcd` should
	// just go here in the `createOrUpdateMetadata` body.
	err := createOrUpdateMetadataEntryInVcd(d, resource)
	if err != nil {
		return err
	}

	if d.HasChange(attributeName) && !d.HasChange("metadata_entry") {
		oldRaw, newRaw := d.GetChange(attributeName)
		oldMetadata := oldRaw.(map[string]interface{})
		newMetadata := newRaw.(map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		for k := range oldMetadata {
			if _, ok := newMetadata[k]; !ok {
				toBeRemovedMetadata = append(toBeRemovedMetadata, k)
			}
		}
		for _, k := range toBeRemovedMetadata {
			err = resource.DeleteMetadataEntry(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata: %s", err)
			}
		}
		if len(newMetadata) > 0 {
			err = resource.MergeMetadata(types.MetadataStringValue, newMetadata)
			if err != nil {
				return fmt.Errorf("error adding metadata: %s", err)
			}
		}
	}
	return nil
}

// stringOnNotNil returns the contents of a string pointer
// if the pointer is nil, returns an empty string
func stringOnNotNil(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Checks if a file exists
func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsRegular()
}

// referenceToId is an auxiliary function to be used with ObjectMap
func referenceToId(reference *types.Reference) string {
	if reference.ID != "" {
		return reference.ID
	}
	return extractUuid(reference.HREF)
}

// referenceToName is an auxiliary function to be used with ObjectMap
func referenceToName(reference *types.Reference) string {
	return reference.Name
}

// vimObjectRefToMoref is an auxiliary function to be used with ObjectMap
func vimObjectRefToMoref(input *types.VimObjectRef) string {
	return input.MoRef
}

// ObjectMap extracts an array of wanted elements from an array of complex objects.
// The Input type is the complex object.
// The Output type could be a simple data type, such as a string or a number, but could
// also be a different object.
// The conversion is performed by the f function, which takes one complex input object and
// produces the wanted output.
// examples:
//
//	    ids := ObjectMap[*types.VimObjectRef, string](extendedProviderVdc.VMWProviderVdc.ResourcePoolRefs.VimObjectRef,
//	        vimObjectRefToMoref)
//
//		ids := ObjectMap[*types.Reference, string](extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile,
//			referenceToId)
//
//		names := ObjectMap[*types.Reference, string](extendedProviderVdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile,
//			referenceToName)
func ObjectMap[Input any, Output any](input []Input, f func(Input) Output) []Output {
	var result = make([]Output, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = f(input[i])
	}
	return result
}

// firstNonEmpty returns the first non empty string from a list
// If all arguments are empty, returns an empty string
func firstNonEmpty(args ...string) string {
	for _, s := range args {
		if s != "" {
			return s
		}
	}
	return ""
}
