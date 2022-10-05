package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"user_access": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User access level for this metadata entry",
			},
			"is_system": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Domain for this metadata entry. true if it belongs to SYSTEM, false if it belongs to GENERAL",
			},
		},
	},
}

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

func setMetadataEntries(d *schema.ResourceData, metadataFromVcd []*types.MetadataEntry) error {
	metadataSet := make([]interface{}, len(metadataFromVcd))
	for i, metadataEntryFromVcd := range metadataFromVcd {
		metadataEntry := map[string]interface{}{
			"key":         metadataEntryFromVcd.Key,
			"value":       metadataEntryFromVcd.TypedValue.Value,
			"user_access": metadataEntryFromVcd.Domain,
		}
		metadataSet[i] = metadataEntry
	}
	err := d.Set("metadata_entry", metadataSet)
	return err
}
