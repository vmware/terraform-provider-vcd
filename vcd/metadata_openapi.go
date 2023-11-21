package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"reflect"
	"strconv"
)

// openApiMetadataEntryDatasourceSchema returns the schema associated to the OpenAPI metadata_entry for a given data source.
// The description will refer to the object type given as input.
func openApiMetadataEntryDatasourceSchema(resourceType string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: fmt.Sprintf("Metadata entries from the given %s", resourceType),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ID of the metadata entry",
				},
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
					Description: fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s'", types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry),
				},
				"readonly": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "True if the metadata entry is read only",
				},
				"domain": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Only meaningful for providers. Allows them to share entries with their tenants. One of: `TENANT`, `PROVIDER`",
				},
				"namespace": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Namespace of the metadata entry",
				},
				"persistent": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Persistent metadata entries can be copied over on some entity operation",
				},
			},
		},
	}
}

// openApiMetadataEntryResourceSchema returns the schema associated to the OpenAPI metadata_entry for a given resource.
// The description will refer to the object name given as input.
func openApiMetadataEntryResourceSchema(resourceType string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: fmt.Sprintf("Metadata entries for the given %s", resourceType),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ID of the metadata entry",
				},
				"key": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Key of this metadata entry. Required if the metadata entry is not empty",
				},
				"value": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Value of this metadata entry. Required if the metadata entry is not empty",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      types.OpenApiMetadataStringEntry,
					Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s'", types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry),
					ValidateFunc: validation.StringInSlice([]string{types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry}, false),
				},
				"readonly": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "True if the metadata entry is read only",
				},
				"domain": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "TENANT",
					Description:  "Only meaningful for providers. Allows them to share entries with their tenants. Currently, accepted values are: `TENANT`, `PROVIDER`",
					ValidateFunc: validation.StringInSlice([]string{"TENANT", "PROVIDER"}, false),
				},
				"namespace": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Namespace of the metadata entry",
				},
				"persistent": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Persistent metadata entries can be copied over on some entity operation",
				},
			},
		},
	}
}

// openApiMetadataCompatible allows to consider all structs that implement OpenAPI metadata handling to be the same type
type openApiMetadataCompatible interface {
	GetMetadata() ([]*types.OpenApiMetadataEntry, error)
	GetMetadataByKey(key string) (*types.OpenApiMetadataEntry, error)
	AddMetadata(metadataEntry types.OpenApiMetadataEntry) (*types.OpenApiMetadataEntry, error)
	UpdateMetadata(key string, value interface{}) (*types.OpenApiMetadataEntry, error)
	DeleteMetadata(key string) error
}

// createOrUpdateOpenApiMetadataEntryInVcd creates or updates OpenAPI metadata entries in VCD for the given resource, only if the attribute
// metadata_entry has been set or updated in the state.
func createOrUpdateOpenApiMetadataEntryInVcd(d *schema.ResourceData, resource openApiMetadataCompatible) error {
	if !d.HasChange("metadata_entry") {
		return nil
	}

	oldRaw, newRaw := d.GetChange("metadata_entry")
	metadataToAdd, metadataToUpdate, metadataToDelete, err := getMetadataOperations(oldRaw.(*schema.Set).List(), newRaw.(*schema.Set).List())
	if err != nil {
		return fmt.Errorf("could not calculate the needed metadata operations: %s", err)
	}

	for _, metadataKey := range metadataToDelete {
		err := resource.DeleteMetadata(metadataKey)
		if err != nil {
			return fmt.Errorf("error deleting metadata: %s", err)
		}
	}

	for metadataKey, metadataEntry := range metadataToUpdate {
		_, err := resource.UpdateMetadata(metadataKey, metadataEntry.KeyValue.Value.Value)
		if err != nil {
			return fmt.Errorf("error updating metadata: %s", err)
		}
	}

	for _, metadataEntry := range metadataToAdd {
		_, err := resource.AddMetadata(metadataEntry)
		if err != nil {
			return fmt.Errorf("error adding metadata entry: %s", err)
		}
	}
	return nil
}

// getMetadataOperations retrieves the metadata that needs to be added, to be updated and to be deleted depending
// on the old and new attribute values from Terraform state.
func getMetadataOperations(oldMetadata []interface{}, newMetadata []interface{}) ([]types.OpenApiMetadataEntry, map[string]types.OpenApiMetadataEntry, []string, error) {
	oldMetadataEntries, err := getOpenApiMetadataEntryMap(oldMetadata)
	if err != nil {
		return nil, nil, nil, err
	}
	newMetadataEntries, err := getOpenApiMetadataEntryMap(newMetadata)
	if err != nil {
		return nil, nil, nil, err
	}

	var metadataToRemove []string
	for oldKey := range oldMetadataEntries {
		if _, ok := newMetadataEntries[oldKey]; !ok {
			metadataToRemove = append(metadataToRemove, oldKey)
		}
	}

	metadataToUpdate := map[string]types.OpenApiMetadataEntry{}
	for newKey, newEntry := range newMetadataEntries {
		if oldEntry, ok := oldMetadataEntries[newKey]; ok {
			if reflect.DeepEqual(oldEntry, newEntry) {
				continue
			}
			metadataToUpdate[newKey] = newEntry
		}
	}

	var metadataToCreate []types.OpenApiMetadataEntry
	for newKey, newEntry := range newMetadataEntries {
		_, alreadyExisting := oldMetadataEntries[newKey]
		_, beingUpdated := metadataToUpdate[newKey]
		if !alreadyExisting && !beingUpdated {
			metadataToCreate = append(metadataToCreate, newEntry)
		}
	}

	return metadataToCreate, metadataToUpdate, metadataToRemove, nil
}

// getOpenApiMetadataKeySet converts the input metadata attribute from Terraform state to a map composed by metadata keys and their values.
func getOpenApiMetadataEntryMap(metadataAttribute []interface{}) (map[string]types.OpenApiMetadataEntry, error) {
	metadataMap := map[string]types.OpenApiMetadataEntry{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})

		namespace := ""
		if _, ok := metadataEntry["namespace"]; ok {
			namespace = metadataEntry["namespace"].(string)
		}

		value, err := convertOpenApiMetadataValue(metadataEntry["type"].(string), metadataEntry["value"].(string))
		if err != nil {
			return nil, fmt.Errorf("error parsing the 'value' attribute '%s' from state: %s", metadataEntry["value"].(string), err)
		}

		metadataMap[metadataEntry["key"].(string)] = types.OpenApiMetadataEntry{
			IsReadOnly:   metadataEntry["readonly"].(bool),   // It is always populated as it has a default value
			IsPersistent: metadataEntry["persistent"].(bool), // It is always populated as it has a default value
			KeyValue: types.OpenApiMetadataKeyValue{
				Domain: metadataEntry["domain"].(string), // It is always populated as it has a default value
				Key:    metadataEntry["key"].(string),    // It is always populated as it is required
				Value: types.OpenApiMetadataTypedValue{
					Value: value,
					Type:  metadataEntry["type"].(string), // It is always populated as it has a default value
				},
				Namespace: namespace,
			},
		}
	}
	return metadataMap, nil
}

// updateOpenApiMetadataInState updates metadata and metadata_entry in the Terraform state for the given receiver object.
// This can be done as both are Computed, for compatibility reasons.
func updateOpenApiMetadataInState(d *schema.ResourceData, receiverObject openApiMetadataCompatible) error {
	allMetadata, err := receiverObject.GetMetadata()
	if err != nil {
		return err
	}

	if len(allMetadata) == 0 {
		return nil
	}

	metadata := make([]interface{}, len(allMetadata))
	for i, metadataEntryFromVcd := range allMetadata {
		// We need to set the correct type, otherwise saving the state will fail
		value := ""
		switch metadataEntryFromVcd.KeyValue.Value.Type {
		case types.OpenApiMetadataBooleanEntry:
			value = fmt.Sprintf("%t", metadataEntryFromVcd.KeyValue.Value.Value.(bool))
		case types.OpenApiMetadataNumberEntry:
			// OpenAPI doesn't
			value = fmt.Sprintf("%.0f", metadataEntryFromVcd.KeyValue.Value.Value.(float64))
		case types.OpenApiMetadataStringEntry:
			value = metadataEntryFromVcd.KeyValue.Value.Value.(string)
		default:
			return fmt.Errorf("not supported metadata type %s", metadataEntryFromVcd.KeyValue.Value.Type)
		}

		metadataEntry := map[string]interface{}{
			"id":         metadataEntryFromVcd.ID,
			"key":        metadataEntryFromVcd.KeyValue.Key,
			"readonly":   metadataEntryFromVcd.IsReadOnly,
			"domain":     metadataEntryFromVcd.KeyValue.Domain,
			"namespace":  metadataEntryFromVcd.KeyValue.Namespace,
			"type":       metadataEntryFromVcd.KeyValue.Value.Type,
			"value":      value,
			"persistent": metadataEntryFromVcd.IsPersistent,
		}
		metadata[i] = metadataEntry
	}

	err = d.Set("metadata_entry", metadata)
	return err
}

// convertOpenApiMetadataValue converts a metadata value from plain string to a correct typed value that can be sent
// in OpenAPI payloads.
func convertOpenApiMetadataValue(valueType, value string) (interface{}, error) {
	var convertedValue interface{}
	var err error
	switch valueType {
	case types.OpenApiMetadataStringEntry:
		convertedValue = value
	case types.OpenApiMetadataNumberEntry:
		convertedValue, err = strconv.ParseFloat(value, 64)
	case types.OpenApiMetadataBooleanEntry:
		convertedValue, err = strconv.ParseBool(value)
	default:
		return nil, fmt.Errorf("unrecognized metadata type %s", valueType)
	}
	if err != nil {
		return nil, fmt.Errorf("error parsing the value '%v': %s", value, err)
	}
	return convertedValue, nil
}
