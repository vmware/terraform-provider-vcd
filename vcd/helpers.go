package vcd

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// onlyHasChange is a schema helper which accepts Terraform schema definition and checks if field
// with `fieldName` is the only one which has change (using d.HasChange)
func onlyHasChange(fieldName string, schema map[string]*schema.Schema, d *schema.ResourceData) bool {
	log.Printf("[DEBUG] [VM update] checking if only field '%s' has change during update", fieldName)
	for schemaFieldName := range schema {
		// Skip checking defined field
		if schemaFieldName == fieldName {
			continue
		}
		if d.HasChange(schemaFieldName) {
			log.Printf("[DEBUG] [VM update] field '%s' has change", schemaFieldName)
			return false
		}
	}
	return true
}
