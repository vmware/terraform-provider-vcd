package vcd

import (
	"context"
	"os"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdMultiSiteSiteData() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSiteMultiSiteRead,
		Schema: map[string]*schema.Schema{
			"association_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data needed to associate this site to another",
			},
			"download_to_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the file to be filled with association data for this site",
			},
		},
	}
}

func datasourceVcdSiteMultiSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	downloadToFile := d.Get("download_to_file").(string)
	rawData, err := client.Client.GetSiteRawAssociationData()
	if err != nil {
		return diag.Errorf("error retrieving association data for current site: %s", err)
	}
	dSet(d, "association_data", string(rawData))
	if downloadToFile != "" {
		err = os.WriteFile(path.Clean(downloadToFile), rawData, 0600)
		if err != nil {
			return diag.Errorf("error writing site association data to file '%s' : %s", downloadToFile, err)
		}
	}
	return nil
}
