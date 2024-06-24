package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sort"
)

func datasourceVcdMultisiteSite() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdMultisiteSiteRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the site",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the site",
			},
			"number_of_associations": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How many associations for this site",
			},
			"associations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of associations",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceVcdMultisiteSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	site, err := client.Client.GetSite()

	if err != nil {
		return diag.Errorf("error retrieving site: %s", err)
	}

	d.SetId(site.Id)
	dSet(d, "name", site.Name)
	dSet(d, "description", site.Description)
	dSet(d, "number_of_associations", len(site.SiteAssociations.SiteAssociations))
	var associations []string
	for _, a := range site.SiteAssociations.SiteAssociations {
		associations = append(associations, a.SiteName)
	}
	if len(associations) > 0 {
		sort.Strings(associations)
		err = d.Set("associations", associations)
		if err != nil {
			return diag.Errorf("error setting value of 'associations': %s", err)
		}
	}
	return nil
}
