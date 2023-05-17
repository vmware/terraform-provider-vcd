package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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

// metadataEntryIgnoreSchema returns the schema associated to metadata_entry_ignore for a given resource.
// The description will refer to the object name given as input.
func metadataEntryIgnoreSchema(objectNameInDescription string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: fmt.Sprintf("Metadata entries to ignore for %s", objectNameInDescription),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Metadata entry key to ignore. It can be a regular expression",
				},
				"value": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Metadata entry value to ignore. It can be a regular expression",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  fmt.Sprintf("Type of the metadata entry to ignore. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
				},
				"user_access": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  fmt.Sprintf("User access level of the metadata entry to ignore. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility}, false),
				},
				"is_system": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Domain of the metadata entry to ignore. true if it belongs to SYSTEM, false if it belongs to GENERAL",
				},
			},
		},
	}
}

// metadataCompatible allows to consider all structs that implement metadata handling to be the same type
type metadataCompatible interface {
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
	if err != nil {
		return fmt.Errorf("error adding metadata entries: %s", err)
	}
	return nil
}

// updateMetadataInState updates metadata and metadata_entry in the Terraform state for the given receiver object.
// This can be done as both are Computed, for compatibility reasons.
func updateMetadataInState(d *schema.ResourceData, receiverObject metadataCompatible) error {
	metadata, err := receiverObject.GetMetadata()
	if err != nil {
		return err
	}

	err = setMetadataEntryInState(d, metadata.MetadataEntry)
	if err != nil {
		return err
	}

	// Set deprecated metadata attribute, just for compatibility reasons
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return err
	}

	return nil
}

// setMetadataEntryInState sets the given metadata entries retrieved from VCD in the Terraform state.
func setMetadataEntryInState(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
	// This early return guarantees that if we try to delete metadata with `metadata_entry {}`, we don't
	// set an empty attribute in state, which would taint it and ask for an update all the time.
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
