package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var metadataEntrySchema = schema.Schema{
	Type:        schema.TypeSet,
	Computed:    true,
	Description: "Key and value pairs for Org VDC metadata",
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
		},
	},
}

func getMetadataEntrySchema() *schema.Schema {
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
