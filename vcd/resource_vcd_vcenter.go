package vcd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

const labelVirtualCenter = "vCenter Server"

func resourceVcdVcenter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVcenterCreate,
		ReadContext:   resourceVcdVcenterRead,
		UpdateContext: resourceVcdVcenterUpdate,
		DeleteContext: resourceVcdVcenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVcenterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("URL including port of %s", labelVirtualCenter),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Username of %s", labelVirtualCenter),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password of %s", labelVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Should the %s be enabled", labelVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVirtualCenter),
			},
			"listener_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVirtualCenter),
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelVirtualCenter),
			},
		},
	}
}

func getTmVcenterType(d *schema.ResourceData) (*types.VSphereVirtualCenter, error) {
	t := &types.VSphereVirtualCenter{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Url:         d.Get("url").(string),
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
		IsEnabled:   d.Get("is_enabled").(bool),
	}

	return t, nil
}

func setTmVcenterData(d *schema.ResourceData, v *govcd.VCenter) error {
	if v == nil || v.VSphereVCenter == nil {
		return fmt.Errorf("nil object for %s", labelVirtualCenter)
	}

	dSet(d, "name", v.VSphereVCenter.Name)
	dSet(d, "description", v.VSphereVCenter.Description)
	dSet(d, "url", v.VSphereVCenter.Url)
	dSet(d, "username", v.VSphereVCenter.Username)
	// dSet(d, "password", v.VSphereVCenter.Password) // password is never returned,
	dSet(d, "is_enabled", v.VSphereVCenter.IsEnabled)

	dSet(d, "has_proxy", v.VSphereVCenter.HasProxy)
	dSet(d, "is_connected", v.VSphereVCenter.IsConnected)
	dSet(d, "mode", v.VSphereVCenter.Mode)
	dSet(d, "listener_state", v.VSphereVCenter.ListenerState)
	dSet(d, "cluster_health_status", v.VSphereVCenter.ClusterHealthStatus)
	dSet(d, "version", v.VSphereVCenter.VcVersion)
	dSet(d, "uuid", v.VSphereVCenter.Uuid)

	d.SetId(v.VSphereVCenter.VcId)

	return nil
}

func resourceVcdVcenterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:      labelVirtualCenter,
		getTypeFunc:      getTmVcenterType,
		stateStoreFunc:   setTmVcenterData,
		createFunc:       vcdClient.CreateVcenter,
		resourceReadFunc: resourceVcdVcenterRead,
		// certificate should be trusted for the vCenter to work
		preCreateHooks: []beforeCreateHook{trustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdVcenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:      labelVirtualCenter,
		getTypeFunc:      getTmVcenterType,
		getEntityFunc:    vcdClient.GetVCenterById,
		resourceReadFunc: resourceVcdVcenterRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterById,
		stateStoreFunc: setTmVcenterData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdVcenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterById,
		preDeleteHooks: []resourceHook[*govcd.VCenter]{disableVcenter}, // vCenter must be disabled before deletion
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdVcenterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	v, err := vcdClient.GetVCenterByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name: %s", labelVirtualCenter, err)
	}

	d.SetId(v.VSphereVCenter.VcId)
	return []*schema.ResourceData{d}, nil
}

// disableVcenter disables vCenter which is usefull before deletion as a non-disabled vCenter cannot
// be removed
func disableVcenter(v *govcd.VCenter) error {
	if v.VSphereVCenter.IsEnabled {
		return v.Disable()
	}
	return nil
}

// trustHostCertificate can automatically add host certificate to trusted ones
// * urlSchemaFieldName - Terraform schema field (TypeString) name that contains URL of entity
// * trustSchemaFieldName - Terraform schema field (TypeBool) name that defines if the certificate should be trusted
// Note. It will not add new entry if the certificate is already trusted
func trustHostCertificate(urlSchemaFieldName, trustSchemaFieldName string) beforeCreateHook {
	return func(vcdClient *VCDClient, d *schema.ResourceData) error {
		shouldExecute := d.Get(trustSchemaFieldName).(bool)
		if !shouldExecute {
			util.Logger.Printf("[DEBUG] Skipping certificate trust execution as '%s' is false", trustSchemaFieldName)
			return nil
		}

		parsedUrl, err := url.Parse(d.Get(urlSchemaFieldName).(string))
		if err != nil {
			return fmt.Errorf("error parsing provided url '%s': %s", d.Get(urlSchemaFieldName).(string), err)
		}

		port, err := strconv.Atoi(parsedUrl.Port())
		if err != nil {
			return fmt.Errorf("error converting '%s' to int: %s", parsedUrl.Port(), err)
		}
		con := types.TestConnection{
			Host:                          parsedUrl.Hostname(),
			Port:                          port,
			Secure:                        addrOf(true),
			Timeout:                       10, // UI timeout value
			HostnameVerificationAlgorithm: "HTTPS",
		}
		res, err := vcdClient.Client.TestConnection(con)
		if err != nil {
			return fmt.Errorf("error testing connection: %s", err)
		}

		// Check if certificate is not trusted yet
		if res != nil && res.TargetProbe != nil && res.TargetProbe.SSLResult != "SUCCESS" {

			if res.TargetProbe.SSLResult == "ERROR_UNTRUSTED_CERTIFICATE" {
				// Need to trust certificate
				cert := res.TargetProbe.CertificateChain
				if cert == "" {
					return fmt.Errorf("error - certificate chain is empty. Connection result: '%s', SSL result: '%s'", res.TargetProbe.ConnectionResult, res.TargetProbe.SSLResult)
				}

				///
				trust := &types.TrustedCertificate{
					Alias:       fmt.Sprintf("%s_%s", parsedUrl.Hostname(), time.Now().UTC().Format(time.RFC3339)),
					Certificate: cert,
				}
				trusted, err := vcdClient.VCDClient.CreateTrustedCertificate(trust)
				if err != nil {
					return fmt.Errorf("error trusting Certificate: %s", err)
				}

				util.Logger.Printf("[DEBUG] Certificate trust established ID - %s, Alias - %s",
					trusted.TrustedCertificate.ID, trusted.TrustedCertificate.Alias)

			} else {
				return fmt.Errorf("SSL verification result - %s", res.TargetProbe.SSLResult)
			}

		}
		return nil
	}
}
