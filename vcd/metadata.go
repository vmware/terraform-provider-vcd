package vcd

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"regexp"
	"strings"
)

var (
	// IgnoreMetadataChangesErrorLevel can be "error", "warn" or "ignore", and tells Terraform whether it should
	// give an error or just a warning if any Metadata Entry configured in HCL is affected by the 'ignore_metadata_changes'
	// configuration.
	IgnoreMetadataChangesErrorLevel string
)

// ignoreMetadataSchema returns the schema associated to ignore_metadata_changes for the provider configuration.
func ignoreMetadataSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Defines a set of `metadata_entry` that need to be ignored by this provider. All filters on this attribute are computed with a logical AND",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"resource_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Ignores metadata from the specific resource type",
					ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Errorf("expected type of '%v' to be string", i)
						}
						found := false
						for k := range resourceMetadataApiRelation {
							if v == k {
								found = true
								break
							}
						}
						if !found {
							return diag.Errorf("can't ignore metadata of resource type '%s'", i)
						}
						return nil
					},
				},
				"object_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Ignores metadata from the specific object in VCD named like this argument",
				},
				"key_regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Regular expression of the metadata entry keys to ignore. Either `key_regex` or `value_regex` is required",
					// Note: This one should have AtLeastOneOf, but it can't be used inside Sets
				},
				"value_regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Regular expression of the metadata entry values to ignore. Either `key_regex` or `value_regex` is required",
					// Note: This one should have AtLeastOneOf, but it can't be used inside Sets
				},
			},
		},
		RequiredWith: []string{"ignore_metadata_changes_error_level"},
	}
}

// Transforms the metadata to ignore from schema to the []govcd.IgnoredMetadata structure.
func getIgnoredMetadata(d *schema.ResourceData, ignoredMetadataAttribute string) ([]govcd.IgnoredMetadata, error) {
	ignoreMetadataRaw := d.Get(ignoredMetadataAttribute).(*schema.Set).List()
	if len(ignoreMetadataRaw) == 0 {
		return []govcd.IgnoredMetadata{}, nil
	}

	result := make([]govcd.IgnoredMetadata, len(ignoreMetadataRaw))
	for i, ignoredEntryRaw := range ignoreMetadataRaw {
		ignoredEntry := ignoredEntryRaw.(map[string]interface{})
		result[i] = govcd.IgnoredMetadata{}
		if ignoredEntry["key_regex"].(string) != "" {
			regex, err := regexp.Compile(ignoredEntry["key_regex"].(string))
			if err != nil {
				return nil, err
			}
			result[i].KeyRegex = regex
		}
		if ignoredEntry["value_regex"].(string) != "" {
			regex, err := regexp.Compile(ignoredEntry["value_regex"].(string))
			if err != nil {
				return nil, err
			}
			result[i].ValueRegex = regex
		}
		if result[i].KeyRegex == nil && result[i].ValueRegex == nil {
			return nil, fmt.Errorf("either `key_regex` or `value_regex` is required inside the `ignore_metadata_changes` attribute")
		}
		if ignoredEntry["object_name"].(string) != "" {
			result[i].ObjectName = addrOf(ignoredEntry["object_name"].(string))
		}
		if ignoredEntry["resource_type"].(string) != "" {
			result[i].ObjectType = addrOf(ignoredEntry["resource_type"].(string))
		}
	}
	return result, nil
}

// This map is used by mapTerraformIgnoredMetadata and the Schema validation. It links a Terraform
// resource type (how the resource was named) with a Metadata API endpoint object present in
// https://developer.vmware.com/apis/1601/vmware-cloud-director
var resourceMetadataApiRelation = map[string]string{
	"vcd_catalog":               "catalog",
	"vcd_catalog_item":          "catalogItem",
	"vcd_catalog_media":         "media",
	"vcd_catalog_vapp_template": "vAppTemplate",
	"vcd_independent_disk":      "disk",
	"vcd_network_direct":        "network",
	"vcd_network_isolated":      "network",
	"vcd_network_isolated_v2":   "network",
	"vcd_network_routed":        "network",
	"vcd_network_routed_v2":     "network",
	"vcd_org":                   "org",
	"vcd_org_vdc":               "vdc",
	"vcd_provider_vdc":          "providervdc",
	"vcd_storage_profile":       "vdcStorageProfile",
	"vcd_vapp":                  "vApp",
	"vcd_vapp_vm":               "vApp",
	"vcd_vm":                    "vApp",
}

// mapTerraformIgnoredMetadata transforms the ObjectType property on every input entry, which references a Terraform
// resource by its name, to the corresponding VCD object that the Go SDK can understand.
func mapTerraformIgnoredMetadata(ignoredMetadata []govcd.IgnoredMetadata) []govcd.IgnoredMetadata {
	terraformMetadataRelation := func(resourceName *string) *string {
		if resourceName == nil {
			return nil
		}
		result, ok := resourceMetadataApiRelation[*resourceName]
		if !ok {
			return nil
		}
		return &result
	}

	result := make([]govcd.IgnoredMetadata, len(ignoredMetadata))
	for i, entry := range ignoredMetadata {
		result[i] = govcd.IgnoredMetadata{
			ObjectType: terraformMetadataRelation(entry.ObjectType),
			ObjectName: entry.ObjectName,
			KeyRegex:   entry.KeyRegex,
			ValueRegex: entry.ValueRegex,
		}
	}
	return result
}

// metadataEntryDatasourceSchema returns the schema associated to metadata_entry for a given data source.
// The description will refer to the object name given as input.
func metadataEntryDatasourceSchema(objectNameInDescription string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: fmt.Sprintf("Metadata entries from the given %s", objectNameInDescription),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Key of this metadata entry",
				},
				"value": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Value of this metadata entry",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
				},
				"user_access": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
				},
				"is_system": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL",
				},
			},
		},
	}
}

// metadataEntryResourceSchema returns the schema associated to metadata_entry for a given resource.
// The description will refer to the object name given as input.
func metadataEntryResourceSchema(objectNameInDescription string) *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeSet,
		Optional:      true,
		Computed:      true, // This is required for `metadata_entry` to live together with deprecated `metadata`.
		Description:   fmt.Sprintf("Metadata entries for the given %s", objectNameInDescription),
		ConflictsWith: []string{"metadata"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Optional:    true, // Should be required, but it is not as specifying an empty `metadata_entry` is the only way to delete metadata in VCD, as metadata_entry is Computed (see above comment)
					Description: "Key of this metadata entry. Required if the metadata entry is not empty",
				},
				"value": {
					Type:        schema.TypeString,
					Optional:    true, // Should be required, but it is not as specifying an empty `metadata_entry` is the only way to delete metadata in VCD, as metadata_entry is Computed (see above comment)
					Description: "Value of this metadata entry. Required if the metadata entry is not empty",
				},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					// Default:      types.MetadataStringValue, // Can't be set like this as we must allow empty `metadata_entry`, to be able to delete metadata (see above comment)
					Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
				},
				"user_access": {
					Type:     schema.TypeString,
					Optional: true,
					// Default:      types.MetadataReadWriteVisibility, // Can't be set like this as we must allow empty `metadata_entry`, to be able to delete metadata (see above comment)
					Description:  fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility}, false),
				},
				"is_system": {
					Type:     schema.TypeBool,
					Optional: true,
					// Default:     false,  // Can't be set like this as we must allow empty `metadata_entry`, to be able to delete metadata (see above comment)
					Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL",
				},
			},
		},
	}
}

// metadataCompatible allows to consider all structs that implement metadata handling to be the same type
type metadataCompatible interface {
	GetMetadataByKey(key string, isSystem bool) (*types.MetadataValue, error)
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntry(typedValue, key, value string) error // Deprecated
	AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error
	MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error
	MergeMetadata(typedValue string, metadata map[string]interface{}) error // Deprecated
	DeleteMetadataEntry(key string) error                                   // Deprecated
	DeleteMetadataEntryWithDomain(key string, isSystem bool) error
}

// createOrUpdateMetadataEntryInVcd creates or updates metadata entries in VCD for the given resource, only if the attribute
// metadata_entry has been set or updated in the state.
func createOrUpdateMetadataEntryInVcd(d *schema.ResourceData, resource metadataCompatible) error {
	if !d.HasChange("metadata_entry") {
		return nil
	}

	// Delete old metadata from VCD
	oldRaw, newRaw := d.GetChange("metadata_entry")
	newMetadata := newRaw.(*schema.Set).List()
	oldKeyMapWithDomain := getMetadataKeyWithDomainMap(oldRaw.(*schema.Set).List())
	newKeyMapWithDomain := getMetadataKeyWithDomainMap(newMetadata)
	for oldKey, isSystem := range oldKeyMapWithDomain {
		if _, newKeyPresent := newKeyMapWithDomain[oldKey]; !newKeyPresent {
			err := resource.DeleteMetadataEntryWithDomain(oldKey, isSystem)
			if err != nil {
				return fmt.Errorf("error deleting metadata entry corresponding to key %s: %s", oldKey, err)
			}
		}
	}

	// Update metadata
	if len(newMetadata) == 0 {
		return nil
	}
	metadataToMerge, err := convertFromStateToMetadataValues(newMetadata)
	if err != nil {
		return err
	}
	if len(metadataToMerge) == 0 {
		return nil
	}
	err = resource.MergeMetadataWithMetadataValues(metadataToMerge)
	if err != nil && !strings.Contains(err.Error(), "after filtering metadata, there is no metadata to merge") {
		return fmt.Errorf("error adding metadata entries: %s", err)
	}
	return nil
}

// checkIgnoredMetadataConflicts checks that no `metadata_entry` managed by Terraform is ignored due to being filtered out
// in any `ignore_metadata_changes` block and errors/warns if so, depending on the value of `ignore_metadata_changes_error_level`.
func checkIgnoredMetadataConflicts(d *schema.ResourceData, vcdClient *VCDClient, resourceType string) diag.Diagnostics {
	metadataEntryList := d.Get("metadata_entry").(*schema.Set).List()
	if len(metadataEntryList) == 0 {
		return nil
	}
	for _, entryRaw := range metadataEntryList {
		entry := entryRaw.(map[string]interface{})
		for _, ignoredMetadata := range vcdClient.Client.IgnoredMetadata {

			if (ignoredMetadata.ObjectType == nil || strings.TrimSpace(*ignoredMetadata.ObjectType) == "" || *ignoredMetadata.ObjectType == resourceMetadataApiRelation[resourceType]) &&
				(ignoredMetadata.ObjectName == nil || strings.TrimSpace(*ignoredMetadata.ObjectName) == "" || strings.TrimSpace(d.Get("name").(string)) == "" || *ignoredMetadata.ObjectName == d.Get("name").(string)) &&
				(ignoredMetadata.KeyRegex == nil || ignoredMetadata.KeyRegex.MatchString(entry["key"].(string))) &&
				(ignoredMetadata.ValueRegex == nil || ignoredMetadata.ValueRegex.MatchString(entry["value"].(string))) {
				util.Logger.Printf("[DEBUG] trying to update metadata entry with key '%s' and value '%v' when it is being being ignored with '%v'", entry["key"].(string), entry["value"].(string), ignoredMetadata)

				var severity diag.Severity
				switch IgnoreMetadataChangesErrorLevel {
				case "error":
					severity = diag.Error
				case "warn":
					severity = diag.Warning
				case "ignore":
					return nil
				default:
					return diag.Errorf("unknown value for 'ignore_metadata_changes_error_level': %s", IgnoreMetadataChangesErrorLevel)
				}
				return diag.Diagnostics{
					diag.Diagnostic{
						Severity: severity,
						Summary:  "Found a conflict between 'ignore_metadata_changes' and 'metadata_entry'",
						Detail: fmt.Sprintf("There is an 'ignored_metadata' block: %s\n"+
							"and there is a 'metadata_entry' with key '%s' and value '%s' in your Terraform configuration that matches the criteria and will be ignored.\n"+
							"This will cause that the entry will be present in Terraform state but it won't have any effect in VCD, causing an inconsistency.\n"+
							"Please use a more fine-grained 'ignore_metadata_changes' configuration or change your metadata entry.", ignoredMetadata, entry["key"], entry["value"]),
						AttributePath: cty.Path{},
					},
				}
			}
		}
	}
	return nil
}

// updateMetadataInState updates metadata and metadata_entry in the Terraform state for the given receiver object.
// This can be done as both are Computed, for compatibility reasons.
func updateMetadataInState(d *schema.ResourceData, vcdClient *VCDClient, resourceType string, receiverObject metadataCompatible) diag.Diagnostics {

	diagErr := checkIgnoredMetadataConflicts(d, vcdClient, resourceType)
	if diagErr != nil {
		return diagErr
	}

	metadata, err := receiverObject.GetMetadata()
	if err != nil {
		return diag.Errorf("error getting metadata to save in state: %s", err)
	}

	err = setMetadataEntryInState(d, metadata.MetadataEntry)
	if err != nil {
		return diag.Errorf("error setting metadata entry in state: %s", err)
	}

	// Set deprecated metadata attribute, just for compatibility reasons
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("error setting metadata in state: %s", err)
	}

	return nil
}

// setMetadataEntryInState sets the given metadata entries retrieved from VCD in the Terraform state.
func setMetadataEntryInState(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
	// A consequence of having metadata_entry computed is that to remove the entries one needs to write `metadata_entry {}`.
	// This snippet guarantees that if we try to delete metadata with `metadata_entry {}`, we don't
	// set an empty Set as attribute in state, which would taint it and ask for an update all the time.
	if len(metadataFromVcd) == 0 {
		return nil
	}

	metadataSet := make([]map[string]interface{}, len(metadataFromVcd))
	for i, metadataEntryFromVcd := range metadataFromVcd {
		metadataEntry := map[string]interface{}{
			"key":         metadataEntryFromVcd.Key,
			"is_system":   false,                             // Default value unless it comes populated from VCD
			"user_access": types.MetadataReadWriteVisibility, // Default value unless it comes populated from VCD
		}
		if metadataEntryFromVcd.TypedValue != nil {
			metadataEntry["type"] = metadataEntryFromVcd.TypedValue.XsiType
			metadataEntry["value"] = metadataEntryFromVcd.TypedValue.Value
		}
		if metadataEntryFromVcd.Domain != nil {
			metadataEntry["is_system"] = metadataEntryFromVcd.Domain.Domain == "SYSTEM"
			metadataEntry["user_access"] = metadataEntryFromVcd.Domain.Visibility
		}
		metadataSet[i] = metadataEntry
	}

	err := d.Set("metadata_entry", metadataSet)
	return err
}

// convertFromStateToMetadataValues converts the structure retrieved from Terraform state to a structure compatible
// with the Go SDK.
func convertFromStateToMetadataValues(metadataAttribute []interface{}) (map[string]types.MetadataValue, error) {
	metadataValues := map[string]types.MetadataValue{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})

		// This is a workaround for metadata_entry deletion, as one needs to set `metadata_entry {}` to be able
		// to delete metadata. Here, if all fields are empty we consider that the entry must be ignored.
		// This must be done as long as metadata_entry is Computed and key, value, etc; are Optional.
		//
		// The len(metadataEntry)-1 is because Terraform doesn't have a tri-valued TypeBool, hence an empty value in `is_system`
		// is always "false", which makes the count of empty attributes wrong by 1.
		metadataEmptyAttributes := getMetadataEmptySubAttributes(metadataEntry)
		if len(metadataEntry)-1 == metadataEmptyAttributes {
			continue
		}
		// On the other hand, if some fields are empty but not all of them, it is not that we set "metadata_entry {}",
		// is that the metadata_entry is malformed, which is an error.
		// This validation would not be needed with Default values, but then trying to delete metadata with "metadata_entry {}"
		// would produce strange state attributes.
		if metadataEmptyAttributes > 0 {
			return nil, fmt.Errorf("all fields in a metadata_entry are required, but got some empty: %v", metadataEntry)
		}

		domain := "GENERAL"
		if metadataEntry["is_system"] != nil && metadataEntry["is_system"].(bool) {
			domain = "SYSTEM"
		}

		metadataValues[metadataEntry["key"].(string)] = types.MetadataValue{
			Domain: &types.MetadataDomainTag{
				Visibility: metadataEntry["user_access"].(string),
				Domain:     domain,
			},
			TypedValue: &types.MetadataTypedValue{
				XsiType: metadataEntry["type"].(string),
				Value:   metadataEntry["value"].(string),
			},
		}
	}
	return metadataValues, nil
}

// getMetadataKeyWithDomainMap converts the input metadata attribute from Terraform state to a map that associates keys
// with their domain (true if is from SYSTEM domain, false if GENERAL).
func getMetadataKeyWithDomainMap(metadataAttribute []interface{}) map[string]bool {
	metadataKeys := map[string]bool{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})

		// This is a workaround for metadata_entry deletion, as one needs to set `metadata_entry {}` to be able
		// to delete metadata. Hence, if all fields are empty we consider that the entry must be ignored.
		// This must be done as long as metadata_entry is Computed and key, value, etc; are Optional.
		//
		// The len(metadataEntry)-1 is because Terraform doesn't have a tri-valued TypeBool, hence an empty value in `is_system`
		// is always "false", which makes the count of empty attributes wrong by 1.
		metadataEmptyAttributes := getMetadataEmptySubAttributes(metadataEntry)
		if len(metadataEntry)-1 == metadataEmptyAttributes {
			continue
		}
		isSystem := false
		if metadataEntry["is_system"] != nil && metadataEntry["is_system"].(bool) {
			isSystem = true
		}

		metadataKeys[metadataEntry["key"].(string)] = isSystem
	}
	return metadataKeys
}

// getMetadataEmptySubAttributes returns the number of empty attributes inside one metadata_entry.
// Returned value can be at most len(metadataEntry).
func getMetadataEmptySubAttributes(metadataEntry map[string]interface{}) int {
	emptySubAttributes := 0
	for _, v := range metadataEntry {
		if v == "" {
			emptySubAttributes++
		}
	}
	return emptySubAttributes
}
