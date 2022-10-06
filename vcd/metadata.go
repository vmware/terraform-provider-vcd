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
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s', '%s'", types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue),
				ValidateFunc: validation.StringInSlice([]string{types.MetadataStringValue, types.MetadataNumberValue, types.MetadataBooleanValue, types.MetadataDateTimeValue}, false),
			},
			"user_access": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("User access level for this metadata entry. One of: '%s', '%s', '%s'", types.MetadataReadWriteVisibility, types.MetadataReadOnlyVisibility, types.MetadataHiddenVisibility),
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

// metadataEntryCompatible allows to consider all resources that implement the "metadata_entry" schema to be the same type.
type metadataEntryCompatible interface {
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error
	MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error
	DeleteMetadataEntry(key string) error
}

// createOrUpdateMetadataInVcd creates or updates metadata entries for the given resource.
func createOrUpdateMetadataInVcd(d *schema.ResourceData, resource metadataEntryCompatible) error {
	if d.HasChange("metadata_entry") {
		oldRaw, newRaw := d.GetChange("metadata_entry")
		oldMetadata := oldRaw.([]map[string]interface{})
		newMetadata := newRaw.([]map[string]interface{})
		var toBeRemovedMetadata []string
		// Check if any key in old metadata was removed in new metadata.
		// Creates a list of keys to be removed.
		for _, oldEntry := range oldMetadata {
			for _, newEntry := range newMetadata {
				if oldEntry["key"] == newEntry["key"] {
					toBeRemovedMetadata = append(toBeRemovedMetadata, oldEntry["key"].(string))
				}
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
			"key":         metadataEntryFromVcd.Key,
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
			Domain:     &types.MetadataDomainTag{
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