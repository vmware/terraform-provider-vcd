package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
)

func datasourceVcdOrgSamlMetadata() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgSamlMetadataRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Org from which we want the SAML metadata",
			},
			"file_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional file name where to store the metadata",
			},
			"metadata_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "text of the metadata retrieved from the remote server URL",
			},
		},
	}
}

func datasourceVcdOrgSamlMetadataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org SAML metadata retrieval] error searching for Org %s: %s", orgId, err)
	}

	metadataText, err := adminOrg.RetrieveServiceProviderSamlMetadata()

	if err != nil {
		return diag.Errorf("error retrieving metadata from Org %s: %s", adminOrg.AdminOrg.Name, err)
	}
	d.SetId(adminOrg.AdminOrg.ID)
	dSet(d, "metadata_text", metadataText)
	fileName := d.Get("file_name").(string)
	if fileName != "" {
		if fileExists(fileName) {
			// If the file exists, compare its contents with the retrieved metadata,
			// and overwrite only if the contents differ
			fileContents, err := os.ReadFile(fileName) // #nosec G304 -- We need user input for this file
			if err != nil {
				return diag.Errorf("[Org SAML metadata retrieval] error reading from file %s: %s", fileName, err)
			}
			if string(fileContents) == metadataText {
				return nil
			}
		}
		err = os.WriteFile(fileName, []byte(metadataText), 0o600)
		if err != nil {
			d.SetId("")
			return diag.Errorf("[Org SAML metadata retrieval] error writing to file %s: %s", fileName, err)
		}
	}
	return nil
}
