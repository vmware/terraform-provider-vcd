package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdMultisiteOrgAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOrgAssociationCreate,
		DeleteContext: resourceVcdOrgAssociationDelete,
		UpdateContext: resourceVcdOrgAssociationUpdate,
		ReadContext:   resourceVcdOrgAssociationRead,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOrgAssociationImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"associated_org_id": {
				Type:        schema.TypeString,
				Optional:    true,
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
			"connection_timeout_mins": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "How many minutes to keep checking for connection (0=no check)",
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
	// Note: "connection_timeout_mins" will only be used in UPDATE operations
	return resourceVcdOrgAssociationRead(ctx, d, meta)
}

func resourceVcdOrgAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgAssociationRead(ctx, d, meta, "resource")
}

func genericVcdOrgAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	associationDataFile := d.Get("association_data_file").(string)

	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		if origin == "datasource" {
			return diag.Errorf("error retrieving Org '%s': %s", orgId, err)
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
	// If no associated Org ID was supplied, the last attempt is to get it through an associated data file
	if associatedOrgId == "" && associationDataFile != "" {
		associatedRawData, err := os.ReadFile(path.Clean(associationDataFile))
		if err != nil {
			return diag.Errorf("error reading from 'association_data_file' %s: %s", associationDataFile, err)
		}
		// Note: this data is only a convenient way of retrieving the associated Org ID. It does not mean that an association
		// exists. For that, the operation below will determine the truth.
		associationData, err := govcd.RawDataToStructuredXml[types.OrgAssociationMember](associatedRawData)
		if err != nil {
			return diag.Errorf("error decoding data from 'association_data_file' %s: %s", associationDataFile, err)
		}
		associatedOrgId = associationData.OrgID
	}
	if associatedOrgId == "" {
		return diag.Errorf("no site ID found in either d.Id() or 'associated_org_id' field")
	}
	// Note: the data retrieved by the operation below only exists if an association has already been established.
	// The existence of an XML file containing an associated Org is just a convenient way of retrieving the Org ID.
	associationData, err := org.GetOrgAssociationByOrgId(associatedOrgId)
	if err != nil {
		return diag.Errorf("association data not found for Org '%s' with org ID '%s': %s", org.AdminOrg.Name, associatedOrgId,
			fmt.Errorf("%s: %s", err, govcd.ErrorEntityNotFound))
	}
	dSet(d, "associated_org_id", associatedOrgId)
	dSet(d, "associated_org_name", associationData.OrgName)
	dSet(d, "associated_site_id", associationData.SiteID)
	dSet(d, "status", associationData.Status)
	d.SetId(associatedOrgId)

	return nil
}

// resourceVcdOrgAssociationUpdate only deals with optional check of association status, as there is nothing else that can be updated
func resourceVcdOrgAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	associatedOrgId := d.Id()
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving Org '%s': %s", orgId, err)
	}
	connectionCheckMinutes := d.Get("connection_timeout_mins").(int)
	if d.HasChange("connection_timeout_mins") && connectionCheckMinutes > 0 {
		status, elapsed, err := org.CheckOrgAssociation(associatedOrgId, time.Minute*time.Duration(connectionCheckMinutes))
		if err != nil {
			return diag.Errorf("error checking for org connection after %s - detected status '%s': %s", elapsed, status, err)
		}
	}
	return resourceVcdOrgAssociationRead(ctx, d, meta)
}

func resourceVcdOrgAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving Org '%s': %s", orgId, err)
	}
	associatedOrgId := d.Id()
	if associatedOrgId == "" {
		associatedOrgId = d.Get("associated_org_id").(string)
	}
	if associatedOrgId == "" {
		return diag.Errorf("no site ID found in either d.Id() or 'associated_org_id' field")
	}
	associationData, err := org.GetOrgAssociationByOrgId(associatedOrgId)
	if err != nil {
		return diag.Errorf("association data not found for Org '%s' with org ID '%s': %s", org.AdminOrg.Name, associatedOrgId, err)
	}
	err = org.RemoveOrgAssociation(associationData.Href)
	if err != nil {
		return diag.Errorf("error removing association data for Org '%s' to org '%s': %s", org.AdminOrg.Name, associationData.OrgName, err)
	}
	return nil
}

func resourceVcdOrgAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*VCDClient)
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("[org association import] org association must be provided as local-org-id.remote-org-id")
	}
	orgId := resourceURI[0]
	associatedOrgId := resourceURI[1]

	org, err := client.GetAdminOrgById(orgId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving org '%s': %s", orgId, err)
	}

	associationData, err := org.GetOrgAssociationByOrgId(associatedOrgId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving association data for org ID '%s': %s", associatedOrgId, err)
	}
	d.SetId(associationData.OrgID)
	dSet(d, "org_id", orgId)
	dSet(d, "associated_org_id", associatedOrgId)
	dSet(d, "associated_org_name", associationData.OrgName)
	return []*schema.ResourceData{d}, nil
}
