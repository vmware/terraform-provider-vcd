package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"path"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdMultisiteSiteAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSiteAssociationCreate,
		ReadContext:   resourceVcdSiteAssociationRead,
		UpdateContext: resourceVcdSiteAssociationUpdate,
		DeleteContext: resourceVcdSiteAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSiteAssociationImport,
		},
		Schema: map[string]*schema.Schema{
			"associated_site_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the site to which the associated site belongs",
			},
			"associated_site_name": {
				Type:     schema.TypeString,
				Computed: true,
				// Note: The API allows customising this field, and it will apply it, but the name will be silently
				// changed back to its default (host name) immediately after creation. Hence, we keep it computed and immutable
				Description: "Name of the site to which the associated site belongs",
			},
			"associated_site_href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the associated site",
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
				ExactlyOneOf: []string{"association_data_file", "association_data"},
				Description:  "Data needed to associate this site to another",
			},
			"association_data_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"association_data_file", "association_data"},
				Description:  "Name of the file to be filled with association data for this site",
			},
		},
	}
}

func resourceVcdSiteAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	rawAssociationData := d.Get("association_data").(string)
	associationDataFile := d.Get("association_data_file").(string)
	var associationData types.SiteAssociationMember
	if associationDataFile != "" {
		associationDataPtr, err := govcd.ReadXmlDataFromFile[types.SiteAssociationMember](associationDataFile)
		if err != nil {
			return diag.Errorf("error retrieving site association data from file '%s': %s", associationDataFile, err)
		}
		associationData = *associationDataPtr
	} else {
		associationDataPtr, err := govcd.RawDataToStructuredXml[types.SiteAssociationMember]([]byte(rawAssociationData))
		if err != nil {
			return diag.Errorf("error retrieving site association data from 'association_data' field: %s", err)
		}
		associationData = *associationDataPtr
	}

	err := client.Client.SetSiteAssociation(associationData)
	if err != nil {
		return diag.Errorf("error setting site association: %s", err)
	}

	d.SetId(associationData.SiteID)
	_, err = client.Client.GetSiteAssociationBySiteId(associationData.SiteID)
	if err != nil {
		return diag.Errorf("no association found for site '%s' after setting: %s", associationData.SiteName, err)
	}
	// Note: connection_timeout_mins will only be used in UPDATE operations
	return resourceVcdSiteAssociationRead(ctx, d, meta)
}

func resourceVcdSiteAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdSiteAssociationRead(ctx, d, meta, "resource")
}

func genericVcdSiteAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	client := meta.(*VCDClient)
	associatedSiteId := d.Id()
	associationDataFile := d.Get("association_data_file").(string)
	if associatedSiteId == "" {
		// In a data source, we need a site ID to access the data
		associatedSiteId = d.Get("associated_site_id").(string)
	}
	// If no associated Site ID was supplied, the last attempt is to get it through an associated data file
	if associatedSiteId == "" && associationDataFile != "" {
		associatedRawData, err := os.ReadFile(path.Clean(associationDataFile))
		if err != nil {
			return diag.Errorf("error reading from 'association_data_file' %s: %s", associationDataFile, err)
		}
		// Note: this data is only a convenient way of retrieving the associated Site ID. It does not mean that an association
		// exists. For that, the operation below will determine the truth.
		associationData, err := govcd.RawDataToStructuredXml[types.SiteAssociationMember](associatedRawData)
		if err != nil {
			return diag.Errorf("error decoding data from 'association_data_file' %s: %s", associationDataFile, err)
		}
		associatedSiteId = associationData.SiteID
	}
	if associatedSiteId == "" {
		return diag.Errorf("no site ID found in either d.Id() or 'associated_site_id' field")
	}
	// Note: the data retrieved by the operation below only exists if an association has already been established.
	// The existence of an XML file containing an associated site is just a convenient way of retrieving the site ID.
	associationData, err := client.Client.GetSiteAssociationBySiteId(associatedSiteId)
	if err != nil {
		if origin == "datasource" {
			return diag.Errorf("error retrieving association data for site ID '%s': %s", associatedSiteId, err)
		}
		d.SetId("")
		return nil
	}
	dSet(d, "associated_site_id", associationData.SiteID)
	dSet(d, "associated_site_name", associationData.SiteName)
	dSet(d, "associated_site_href", associationData.Href)
	dSet(d, "status", associationData.Status)
	d.SetId(associationData.SiteID)

	return nil
}

// resourceVcdSiteAssociationUpdate will only update "connection_timeout_mins"
func resourceVcdSiteAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	associatedSiteId := d.Id()
	connectionCheckMinutes := d.Get("connection_timeout_mins").(int)
	if d.HasChange("connection_timeout_mins") && connectionCheckMinutes > 0 {
		status, elapsed, err := client.Client.CheckSiteAssociation(associatedSiteId, time.Minute*time.Duration(connectionCheckMinutes))
		if err != nil {
			return diag.Errorf("error checking for site connection after %s - detected status '%s': %s", elapsed, status, err)
		}
	}
	return resourceVcdSiteAssociationRead(ctx, d, meta)
}

func resourceVcdSiteAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*VCDClient)
	associatedSiteId := d.Id()
	if associatedSiteId == "" {
		associatedSiteId = d.Get("associated_site_id").(string)
	}
	if associatedSiteId == "" {
		return diag.Errorf("no site ID found in either d.Id() or 'associated_site_id' field")
	}
	associationData, err := client.Client.GetSiteAssociationBySiteId(associatedSiteId)
	if err != nil {
		return diag.Errorf("error retrieving association data for site ID '%s': %s", associatedSiteId, err)
	}
	err = client.Client.RemoveSiteAssociation(associationData.Href)
	if err != nil {
		return diag.Errorf("error removing site '%s': %s", associationData.SiteName, err)
	}
	return nil
}

func resourceVcdSiteAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*VCDClient)

	associatedSiteId := d.Id()
	if associatedSiteId == "" {
		associatedSiteId = d.Get("associated_site_id").(string)
	}

	if associatedSiteId == "" {
		return nil, fmt.Errorf("[site association import] no site ID found in either d.Id() or 'associated_site_id' field")
	}
	associationData, err := client.Client.GetSiteAssociationBySiteId(associatedSiteId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving association data for site ID '%s': %s", associatedSiteId, err)
	}
	d.SetId(associationData.SiteID)
	dSet(d, "associated_site_name", associationData.SiteName)
	return []*schema.ResourceData{d}, nil
}
