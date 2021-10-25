package vcd

import (
	"context"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceCertificateInLibrary() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCertificateInLibraryRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Alias of certificate",
			},
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Certificate ID",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Certificate description",
			},
			"certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Certificate content",
			},
		},
	}
}

func datasourceVcdCertificateInLibraryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	// get by ID when it's available
	var certificate *govcd.Certificate
	if isSysOrg(adminOrg) {
		alias := d.Get("alias").(string)
		if alias != "" {
			certificate, err = vcdClient.Client.GetCertificateFromLibraryByName(alias)
		} else if d.Get("id").(string) != "" {
			certificate, err = vcdClient.Client.GetCertificateFromLibraryById(d.Get("id").(string))
		} else {
			return diag.Errorf("Id or Alias value is missing %s", err)
		}
	} else {
		alias := d.Get("alias").(string)
		if alias != "" {
			certificate, err = adminOrg.GetCertificateFromLibraryByName(alias)
		} else if d.Get("id").(string) != "" {
			certificate, err = adminOrg.GetCertificateFromLibraryById(d.Get("id").(string))
		} else {
			return diag.Errorf("Id or Alias value is missing %s", err)
		}
	}
	if err != nil {
		return diag.Errorf("[certificate library read] : %s", err)
	}

	d.SetId(certificate.CertificateLibrary.Id)
	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return nil
}
