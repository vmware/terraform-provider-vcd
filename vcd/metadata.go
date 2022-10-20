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
		Computed:      true, // Just to live together with deprecated "metadata" attribute.
		Description:   fmt.Sprintf("Metadata entries for the given %s", objectNameInDescription),
		ConflictsWith: []string{"metadata"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Optional:    true, // Should be required, but it is not as specifying empty entry is the only way to delete metadata_entry as it's Computed
					Description: "Key of this metadata entry. Required if the metadata entry is not empty",
				},
				"value": {
					Type:        schema.TypeString,
					Optional:    true, // Should be required, but it is not as specifying empty entry is the only way to delete metadata_entry as it's Computed.
					Description: "Value of this metadata entry. Required if the metadata entry is not empty",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					// Default:      types.MetadataStringValue, // Can't be set like this as we allow empty metadata entries to be able to delete metadata
					Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
				},
				"user_access": {
					Type:         schema.TypeString,
					Optional:     true,
					// Default:      types.MetadataReadWriteVisibility, // Can't be set like this as we allow empty metadata entries to be able to delete metadata
					Description:  fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility}, false),
				},
				"is_system": {
					Type:        schema.TypeBool,
					Optional:    true,
					// Default:     false,  // Can't be set like this as we allow empty metadata entries to be able to delete metadata
					Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL",
				},
			},
		},
	}
}

// getMetadataEntrySchema returns a schema for the "metadata_entry" attribute, that can be used to
// build data sources (isDatasource=true) or resources (isDatasource=false). The description of the
// attribute will refer to the input resource name.
func getMetadataEntrySchema(resourceNameInDescription string, isDatasource bool) *schema.Schema {
	if isDatasource {
		return metadataEntryDatasourceSchema(resourceNameInDescription)
	}
	return metadataEntryResourceSchema(resourceNameInDescription)
}

// metadataCompatible allows to consider all structs that implement metadata handling to be the same type
type metadataCompatible interface {
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntry(typedValue, key, value string) error // Deprecated
	AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error
	MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error
	MergeMetadata(typedValue string, metadata map[string]interface{}) error // Deprecated
	DeleteMetadataEntry(key string) error // Deprecated
	DeleteMetadataEntryWithDomain(key string, isSystem bool) error
}

// createOrUpdateMetadataInVcd creates or updates metadata entries in VCD for the given resource, only if the attribute
// metadata_entry has been set or updated in the state.
func createOrUpdateMetadataInVcd(d *schema.ResourceData, resource metadataCompatible) error {
	if d.HasChange("metadata_entry") {
		oldRaw, newRaw := d.GetChange("metadata_entry")
		newMetadata := newRaw.(*schema.Set).List()
		// Check if any key in old metadata was removed in new metadata to remove it from VCD
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
	}
	return nil
}

// updateMetadataInState updates metadata and metadata_entry in the Terraform state for the given receiver object.
// If the origin is "resource", it updates metadata attribute only if it's changed, and the same for metadata_entry.
// Both can never be changed at same time as they conflict with each other in the schema (it relies on ConflictsWith in schema).
// If the origin is "datasource", it updates both metadata and metadata_entry as both are Computed.
//
// The goal of this logic is that metadata and metadata_entry can live together until metadata gets deprecated.
func updateMetadataInState(d *schema.ResourceData, receiverObject metadataCompatible) error {
	metadata, err := receiverObject.GetMetadata()
	if err != nil {
		return err
	}

	err = setMetadataEntryInState(d, metadata.MetadataEntry)
	if err != nil {
		return err
	}

	// Deprecated, just for compatibility reasons
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return err
	}

	return nil
}

// setMetadataEntryInState sets the given metadata entries retrieved from VCD in the Terraform state.
func setMetadataEntryInState(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
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

		// This is a workaround for metadata_entry deletion, as one needs to set "metadata_entry {}" to be able
		// to delete metadata. Hence, if all fields are empty we consider that the entry must be ignored.
		// This must be done as long as metadata_entry is Computed and key, value, etc; are Optional.
		if !allMetadataEntryFieldsAreSet(metadataEntry) {
			continue
		}
		// Same as above, a workaround as we can't have Required sub-attributes inside metadata_entry to allow deletion.
		// Hence, we check here that all are populated.
		if someMetadataEntryFieldsAreEmpty(metadataEntry) {
			return nil, fmt.Errorf("all fields in a metadata_entry are required, but got some empty: %v", metadataEntry)
		}

		domain := "GENERAL"
		if metadataEntry["is_system"] != nil && metadataEntry["is_system"].(bool) {
			domain = "SYSTEM"
		}
		// These could be done by the Terraform SDK "Default" option, but we can't as metadata_entry is Computed and
		// needs empty values for deletion of metadata entries.
		// This must be done as long as these fields don't have a Default option and are Computed.

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

		// This is a workaround for metadata_entry deletion, as one needs to set "metadata_entry {}".
		// This must be done as long as metadata_entry is Computed and key and is_system are Optional.
		if !allMetadataEntryFieldsAreSet(metadataEntry) {
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

// allMetadataEntryFieldsAreSet serves as a workaround to allow empty `metadata_entry`, to delete metadata from VCD.
// This function can be used to detect this case.
func allMetadataEntryFieldsAreSet(metadataEntry map[string]interface{}) bool {
	expectedFilledEntries := len(metadataEntry)
	actualFilledEntries := 0
	for _, v := range metadataEntry {
		if v != "" {
			actualFilledEntries++
		}
	}
	return expectedFilledEntries == actualFilledEntries
}

// someMetadataEntryFieldsAreEmpty serves as a workaround to allow empty `metadata_entry`, to delete metadata from VCD.
// This function can be used to complement allMetadataEntryFieldsAreSet to detect a metadata_entry which is fully
// set.
func someMetadataEntryFieldsAreEmpty(metadataEntry map[string]interface{}) bool {
	for _, v := range metadataEntry {
		if v == "" {
			return true
		}
	}
	return false
}