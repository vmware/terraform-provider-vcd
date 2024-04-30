package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"path"
	"sort"
)

func datasourceVcdOrgMultiSite() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgMultiSiteRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization ID",
			},
			"number_of_associations": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How many associations for this Org",
			},
			"associations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of associations",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"association_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data needed to associate this Organization to another",
			},
			"association_data_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the file to be filled with association data for this Org",
			},
		},
	}
}

func datasourceVcdOrgMultiSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	associationDataFile := d.Get("association_data_file").(string)
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving Org '%s': %s", orgId, err)
	}

	rawData, err := org.GetOrgRawAssociationData()
	if err != nil {
		return diag.Errorf("error retrieving Org '%s' association data : %s", org.AdminOrg.Name, err)
	}
	d.SetId(orgId)
	dSet(d, "association_data", string(rawData))
	if associationDataFile != "" {
		err = os.WriteFile(path.Clean(associationDataFile), rawData, 0600)
		if err != nil {
			return diag.Errorf("error writing Org '%s' association data to file '%s' : %s", org.AdminOrg.Name, associationDataFile, err)
		}
	}
	associations, err := org.GetOrgAssociations()
	if err != nil {
		return diag.Errorf("error retrieving associations for Org '%s': %s", org.AdminOrg.Name, err)
	}
	dSet(d, "number_of_associations", len(associations))
	var orgAssociations []string
	for _, a := range associations {
		orgAssociations = append(orgAssociations, a.OrgName+" "+a.OrgID)
	}
	if len(orgAssociations) > 0 {
		sort.Strings(orgAssociations)
		err = d.Set("associations", orgAssociations)
		if err != nil {
			return diag.Errorf("error setting list of associations for org '%s': %s", org.AdminOrg.Name, err)
		}
	}
	return nil
}
