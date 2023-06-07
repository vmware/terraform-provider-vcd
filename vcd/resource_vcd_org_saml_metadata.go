package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/util"
	"os"
)

func resourceVcdOrgSamlMetadata() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOrgSamlMetadataCRU,
		ReadContext:   resourceVcdOrgSamlMetadataCRU,
		UpdateContext: resourceVcdOrgSamlMetadataCRU,
		DeleteContext: resourceVcdOrgSamlMetadataDelete,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Org from which we want the SAML metadata",
			},
			"file_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "File name where to store the metadata",
			},
			"metadata_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "text of the metadata retrieved from the remote server URL",
			},
		},
	}
}

func resourceVcdOrgSamlMetadataCRU(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org SAML metadata creation] error searching for Org %s: %s", orgId, err)
	}

	util.Logger.Printf("--- start retrieving\n")
	metadataText, err := adminOrg.RetrieveServiceProviderSamlMetadata()

	if err != nil {
		return diag.Errorf("error retrieving metadata from Org %s: %s", adminOrg.AdminOrg.Name, err)
	}
	d.SetId(adminOrg.AdminOrg.ID)
	dSet(d, "metadata_text", metadataText)
	fileName := d.Get("file_name").(string)
	err = os.WriteFile(fileName, []byte(metadataText), 0o600)
	if err != nil {
		d.SetId("")
		return diag.Errorf("[Org SAML metadata creation] error writing to file %s: %s", fileName, err)
	}
	return nil
}

func resourceVcdOrgSamlMetadataDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	fileName := d.Get("file_name").(string)
	return diag.FromErr(os.Remove(fileName))
}
