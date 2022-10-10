package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var baseMetadataEntrySchema = schema.Schema{
	Type: schema.TypeSet,
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
				Type:         schema.TypeString,
				Computed:     true,
				Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
				ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
			},
			"user_access": {
				Type:         schema.TypeString,
				Computed:     true,
				Description:  fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
				ValidateFunc: validation.StringInSlice([]string{types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility}, false),
			},
			"is_system": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL",
			},
		},
	},
}

// getMetadataEntrySchema returns a schema for the "metadata_entry" attribute, that can be used to
// build data sources (isDatasource=true) or resources (isDatasource=false). The description of the
// attribute will refer to the input resource name.
func getMetadataEntrySchema(resourceNameInDescription string, isDatasource bool) *schema.Schema {
	metadataEntrySchema := baseMetadataEntrySchema
	metadataEntrySchema.Description = fmt.Sprintf("Key and value pairs for %s metadata", resourceNameInDescription)
	if isDatasource {
		metadataEntrySchema.Computed = true
	} else {
		metadataEntrySchema.Optional = true
	}
	return &metadataEntrySchema
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

// createOrUpdateMetadataInVcd creates or updates metadata entries for the given resource.
func createOrUpdateMetadataInVcd(d *schema.ResourceData, resource metadataCompatible) error {
	if d.HasChange("metadata_entry") {
		oldRaw, newRaw := d.GetChange("metadata_entry")
		newMetadata := newRaw.([]map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		oldKeySet := getMetadataKeySet(oldRaw.([]map[string]interface{}))
		newKeySet := getMetadataKeySet(newMetadata)
		for oldKey := range oldKeySet {
			if _, newKeyPresent := newKeySet[oldKey]; !newKeyPresent {
				toBeRemovedMetadata = append(toBeRemovedMetadata, oldKey)
			}
		}
		for _, k := range toBeRemovedMetadata {
			err := resource.DeleteMetadataEntry(k)
			if err != nil {
				return fmt.Errorf("error deleting metadata entries: %s", err)
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

// setMetadataEntryInState sets the given metadata entries in the Terraform state.
func setMetadataEntryInState(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
	metadataSet := make([]map[string]interface{}, len(metadataFromVcd))

	for i, metadataEntryFromVcd := range metadataFromVcd {
		metadataEntry := map[string]interface{}{
			"key": metadataEntryFromVcd.Key,
		}
		if metadataEntryFromVcd.TypedValue != nil {
			metadataEntry["type"] = metadataEntryFromVcd.TypedValue.XsiType
			metadataEntry["value"] = metadataEntryFromVcd.TypedValue.Value
		}
		if metadataEntryFromVcd.Domain != nil {
			metadataEntry["is_system"] = metadataEntryFromVcd.Domain.Domain == "SYSTEM"
			metadataEntry["visibility"] = metadataEntryFromVcd.Domain.Visibility
		}
		metadataSet[i] = metadataEntry
	}

	err := d.Set("metadata_entry", metadataSet)
	return err
}

// convertFromStateToMetadataValues converts the structure retrieved from Terraform state to a structure compatible
// with the Go SDK.
func convertFromStateToMetadataValues(metadataAttribute []map[string]interface{}) map[string]types.MetadataValue {
	metadataValue := map[string]types.MetadataValue{}
	for _, metadataEntry := range metadataAttribute {
		domain := "GENERAL"
		if metadataEntry["is_system"].(bool) {
			domain = "SYSTEM"
		}
		metadataValue[metadataEntry["key"].(string)] = types.MetadataValue{
			Domain: &types.MetadataDomainTag{
				Visibility: metadataEntry["visibility"].(string),
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

// getMetadataKeySet gives the metadata key set associated to the input metadata attribute from Terraform state.
func getMetadataKeySet(metadataAttribute []map[string]interface{}) map[string]bool {
	metadataKeys := map[string]bool{}
	for _, metadataEntry := range metadataAttribute {
		metadataKeys[metadataEntry["key"].(string)] = true
	}
	return metadataKeys
}
