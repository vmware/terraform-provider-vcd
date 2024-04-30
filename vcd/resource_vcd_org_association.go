package vcd

import (
	"context"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdOrgAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOrgAssociationCreate,
		DeleteContext: resourceVcdOrgAssociationDelete,
		ReadContext:   resourceVcdOrgAssociationRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"associated_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the associated Organization",
			},
			"associated_org_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the associated Organization",
			},
			"associated_site_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the site to which the associated Organization belongs",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the association",
			},
			"association_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"association_data", "association_data_file"},
				Description:  "Data needed to associate this Organization to another",
			},
			"association_data_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"association_data", "association_data_file"},
				Description:  "Name of the file filled with association data for this Org",
			},
		},
	}
}

func resourceVcdOrgAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	rawAssociationData := d.Get("association_data").(string)
	associationDataFile := d.Get("association_data_file").(string)
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving Org '%s': %s", orgId, err)
	}
	var associationData types.OrgAssociationMember
	if associationDataFile != "" {
		associationDataPtr, err := govcd.ReadXmlDataFromFile[types.OrgAssociationMember](path.Clean(associationDataFile))
		//rawData, err := os.ReadFile(path.Clean(associationDataFile))
		if err != nil {
			return diag.Errorf("error reading association data from file '%s' : %s", associationDataFile, err)
		}
		associationData = *associationDataPtr
	} else {
		associationDataPtr, err := govcd.RawDataToStructuredXml[types.OrgAssociationMember]([]byte(rawAssociationData))
		if err != nil {
			return diag.Errorf("error decoding association data : %s", err)
		}
		associationData = *associationDataPtr
	}

	err = org.SetOrgAssociation(associationData)
	if err != nil {
		return diag.Errorf("error setting association between Org '%s' and Org '%s': %s", org.AdminOrg.Name, associationData.OrgName, err)
	}
	d.SetId(associationData.OrgID)
	return resourceVcdOrgAssociationRead(ctx, d, meta)
}

func resourceVcdOrgAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgAssociationRead(ctx, d, meta, "resource")
}

func genericVcdOrgAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		if origin == "datasource" {
			return diag.Errorf("error retrieving Org '%s': %s", err)
		}
		d.SetId("")
		return nil
	}
	// In a resource, the associated Org ID should have been set during creation
	associatedOrgId := d.Id()
	if associatedOrgId == "" {
		// In a data source, we need an Org ID to access the data
		associatedOrgId = d.Get("associated_org_id").(string)
	}
	associationData, err := org.GetOrgAssociationByOrgId(associatedOrgId)
	if err != nil {
		return diag.Errorf("association data not found for Org '%s' with org ID '%s': %s", org.AdminOrg.Name, associatedOrgId)
	}
	dSet(d, "associated_org_id", associatedOrgId)
	dSet(d, "associated_org_name", associationData.OrgName)
	dSet(d, "associated_site_id", associationData.SiteID)
	dSet(d, "status", associationData.Status)

	return nil
}

func resourceVcdOrgAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving Org '%s': %s", err)
	}
	associatedOrgId := d.Id()
	associationData, err := org.GetOrgAssociationByOrgId(associatedOrgId)
	if err != nil {
		return diag.Errorf("association data not found for Org '%s' with org ID '%s': %s", org.AdminOrg.Name, associatedOrgId)
	}
	err = org.RemoveOrgAssociation(associationData.Href)
	if err != nil {
		return diag.Errorf("error removing association data for Org '%s' to org '%s': %s", org.AdminOrg.Name, associationData.OrgName, err)
	}
	return nil
}
