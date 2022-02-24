package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLibraryCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdLibraryCertificateRead,
		CreateContext: resourceVcdLibraryCertificateCreate,
		UpdateContext: resourceVcdLibraryCertificateUpdate,
		DeleteContext: resourceVcdAlbLibraryCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceLibraryCertificateImport,
		},
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
				Required:    true,
				Description: "Alias of certificate",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Certificate description",
			},
			"certificate": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Certificate content",
			},
			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Certificate private key",
			},
			"private_key_passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Certificate private pass phrase",
			},
		},
	}
}

// resourceVcdLibraryCertificateCreate covers Create functionality for resource
func resourceVcdLibraryCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	certificateConfig := getCertificateConfigurationType(d)
	var createdCertificate *govcd.Certificate
	if isSysOrg(adminOrg) {
		createdCertificate, err = vcdClient.Client.AddCertificateToLibrary(certificateConfig)
		if err != nil {
			return diag.Errorf("error adding certificate library item: %s", err)
		}
	} else {
		createdCertificate, err = adminOrg.AddCertificateToLibrary(certificateConfig)
		if err != nil {
			return diag.Errorf("error adding certificate library item: %s", err)
		}

	}
	d.SetId(createdCertificate.CertificateLibrary.Id)
	return resourceVcdLibraryCertificateRead(ctx, d, meta)
}

func isSysOrg(adminOrg *govcd.AdminOrg) bool {
	return strings.EqualFold(adminOrg.AdminOrg.Name, "system")
}

// resourceVcdLibraryCertificateUpdate covers Update functionality for resource
func resourceVcdLibraryCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var certificate *govcd.Certificate
	if isSysOrg(adminOrg) {
		certificate, err = vcdClient.Client.GetCertificateFromLibraryById(d.Id())
	} else {
		certificate, err = adminOrg.GetCertificateFromLibraryById(d.Id())
	}
	if err != nil {
		return diag.Errorf("[certificate library update] : %s", err)
	}

	certificateConfig := getCertificateConfigurationType(d)
	certificate.CertificateLibrary.Alias = certificateConfig.Alias
	certificate.CertificateLibrary.Description = certificateConfig.Description
	_, err = certificate.Update()
	if err != nil {
		return diag.Errorf("[certificate library update] : %s", err)
	}

	return resourceVcdLibraryCertificateRead(ctx, d, meta)
}

func getCertificateConfigurationType(d *schema.ResourceData) *types.CertificateLibraryItem {
	return &types.CertificateLibraryItem{
		Alias:                d.Get("alias").(string),
		Description:          d.Get("description").(string),
		Certificate:          d.Get("certificate").(string),
		PrivateKey:           d.Get("private_key").(string),
		PrivateKeyPassphrase: d.Get("private_key_passphrase").(string),
	}
}

func resourceVcdLibraryCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var certificate *govcd.Certificate
	if isSysOrg(adminOrg) {
		certificate, err = vcdClient.Client.GetCertificateFromLibraryById(d.Id())
	} else {
		certificate, err = adminOrg.GetCertificateFromLibraryById(d.Id())
	}
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[certificate library read] : %s", err)
	}

	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return nil
}

func setCertificateConfigurationData(config *types.CertificateLibraryItem, d *schema.ResourceData) {
	dSet(d, "alias", config.Alias)
	dSet(d, "description", config.Description)
	dSet(d, "certificate", config.Certificate)
}

func resourceVcdAlbLibraryCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	var certificateToDelete *govcd.Certificate
	if isSysOrg(adminOrg) {
		certificateToDelete, err = vcdClient.Client.GetCertificateFromLibraryById(d.Id())
	} else {
		certificateToDelete, err = adminOrg.GetCertificateFromLibraryById(d.Id())
	}
	if err != nil {
		return diag.Errorf("[certificate library delete] : %s", err)
	}

	return diag.FromErr(certificateToDelete.Delete())
}

func resourceLibraryCertificateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.certificate-name")
	}
	orgName, certificateName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[certificate import] error retrieving org %s: %s", orgName, err)
	}

	var certificate *govcd.Certificate
	if isSysOrg(adminOrg) {
		certificate, err = vcdClient.Client.GetCertificateFromLibraryByName(certificateName)
	} else {
		certificate, err = adminOrg.GetCertificateFromLibraryByName(certificateName)
	}
	if err != nil {
		return nil, fmt.Errorf("error importing certificate library item: %s", err)
	}

	d.SetId(certificate.CertificateLibrary.Id)
	dSet(d, "org", orgName)
	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return []*schema.ResourceData{d}, nil
}
