package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// metadataEntryDatasourceSchema returns the schema associated to metadata_entry for a given datasource.
// The description will refer to the resource name given as input.
func metadataEntryDatasourceSchema(resourceNameInDescription string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: fmt.Sprintf("Metadata entries from the given %s", resourceNameInDescription),
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
// The description will refer to the resource name given as input.
func metadataEntryResourceSchema(resourceNameInDescription string) *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeSet,
		Optional:      true,
		Description:   fmt.Sprintf("Metadata entries for the given %s", resourceNameInDescription),
		ConflictsWith: []string{"metadata"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Key of this metadata entry",
				},
				"value": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Value of this metadata entry",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      types.MetadataStringValue,
					Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'. Defaults to %s", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue, types.MetadataStringValue),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
				},
				"user_access": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      types.MetadataReadWriteVisibility,
					Description:  fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'. Defaults to %s", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility, types.MetadataReadWriteVisibility),
					ValidateFunc: validation.StringInSlice([]string{types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility}, false),
				},
				"is_system": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL. Defaults to false.",
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
	DeleteMetadataEntry(key string) error
}

// createOrUpdateMetadataInVcd creates or updates metadata entries in VCD for the given resource, only if the attribute
// metadata_entry has been set or updated in the state.
func createOrUpdateMetadataInVcd(d *schema.ResourceData, resource metadataCompatible) error {
	if d.HasChange("metadata_entry") {
		oldRaw, newRaw := d.GetChange("metadata_entry")
		newMetadata := newRaw.(*schema.Set).List()
		// Check if any key in old metadata was removed in new metadata to remove it from VCD
		oldKeySet := getMetadataKeySet(oldRaw.(*schema.Set).List())
		newKeySet := getMetadataKeySet(newMetadata)
		for oldKey := range oldKeySet {
			if _, newKeyPresent := newKeySet[oldKey]; !newKeyPresent {
				err := resource.DeleteMetadataEntry(oldKey)
				if err != nil {
					return fmt.Errorf("error deleting metadata entries: %s", err)
				}
			}
		}
		if len(newMetadata) > 0 {
			err := resource.MergeMetadataWithMetadataValues(convertFromStateToMetadataValues(newMetadata))
			if err != nil {
				return fmt.Errorf("error adding metadata entries: %s", err)
			}
		}
	}
	return nil
}

// updateMetadataInState updates metadata or metadata_entry in the Terraform state for the given receiver object.
// If the origin is "resource", it updates metadata attribute only if it's changed, and the same for metadata_entry.
// Both can never be changed at same time as they conflict with each other in the schema (it relies on ConflictsWith in schema).
// If the origin is "datasource", it updates both metadata and metadata_entry as both are Computed.
//
// The goal of this logic is that metadata and metadata_entry can live together until metadata gets deprecated.
func updateMetadataInState(d *schema.ResourceData, receiverObject metadataCompatible, origin string) error {
	metadata, err := receiverObject.GetMetadata()
	if err != nil {
		return err
	}

	if len(metadata.MetadataEntry) == 0 {
		return nil
	}

	if origin == "datasource" || (origin == "resource" && d.HasChange("metadata_entry")) {
		err = setMetadataEntryInState(d, metadata.MetadataEntry)
		if err != nil {
			return err
		}
	}

	if origin == "datasource" || (origin == "resource" && d.HasChange("metadata")) {
		err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
		if err != nil {
			return err
		}
	}

	return nil
}

// setMetadataEntryInState sets the given metadata entries retrieved from VCD in the Terraform state.
func setMetadataEntryInState(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
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
func convertFromStateToMetadataValues(metadataAttribute []interface{}) map[string]types.MetadataValue {
	metadataValue := map[string]types.MetadataValue{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})

		domain := "GENERAL"
		if metadataEntry["is_system"].(bool) {
			domain = "SYSTEM"
		}
		metadataValue[metadataEntry["key"].(string)] = types.MetadataValue{
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
	return metadataValue
}

// getMetadataKeySet converts the input metadata attribute from Terraform state to a metadata key set.
func getMetadataKeySet(metadataAttribute []interface{}) map[string]bool {
	metadataKeys := map[string]bool{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})
		metadataKeys[metadataEntry["key"].(string)] = true
	}
	return metadataKeys
}
