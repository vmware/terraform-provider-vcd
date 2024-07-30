package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdCatalogVappTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCatalogVappTemplateRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the catalog containing the vApp Template. Can't be used if a specific VDC identifier is set",
				ExactlyOneOf: []string{"catalog_id", "vdc_id"},
			},
			"vdc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the VDC to which the vApp Template belongs. Can't be used if a specific Catalog identifier is set",
				ExactlyOneOf: []string{"catalog_id", "vdc_id"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Name of the vApp Template. It is optional when a filter is provided",
				ExactlyOneOf: []string{"name", "filter"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of when the vApp Template was created",
			},
			"catalog_item_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog Item ID of this vApp template",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs from the metadata of the vApp template",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("vApp Template"),
			"inherited_metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "A map that contains metadata that is automatically added by VCD (10.5.1+) and provides details on the origin of the VM",
			},
			"vm_names": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of VM names within the vApp template",
			},
			"lease": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Defines lease parameters for this vApp template",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_lease_in_sec": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "How long the vApp template is available before being automatically deleted or marked as expired. 0 means never expires",
						},
					},
				},
			},
			"filter": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a vApp Template by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"date":       elementDate,
						"earliest":   elementEarliest,
						"latest":     elementLatest,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func datasourceVcdCatalogVappTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogVappTemplateRead(ctx, d, meta, "datasource")
}
