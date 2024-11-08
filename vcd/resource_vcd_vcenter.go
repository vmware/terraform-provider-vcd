package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
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
				ForceNew:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVirtualCenter),
			},
			"refresh_vcenter_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should be refreshed on every read operation", labelVirtualCenter),
			},
			"refresh_policies_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should refresh Policies on every read operation", labelVirtualCenter),
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

func getTmVcenterType(_ *VCDClient, d *schema.ResourceData) (*types.VSphereVirtualCenter, error) {
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
		preCreateHooks: []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdVcenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// return immediately if only flags are updated
	if !d.HasChangesExcept("refresh_vcenter_on_read", "refresh_policies_on_read") {
		return nil
	}

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

	// TODO: TM: remove this block and use the commented one within crudConfig below.
	// Retrieval endpoints by Name and by ID return differently formated url (the by Id one returns
	// URL with port http://host:443, while the one by name - doesn't). Using the same getByName to
	// match format everywhere
	fakeGetById := func(id string) (*govcd.VCenter, error) {
		vc, err := vcdClient.GetVCenterById(id)
		if err != nil {
			return nil, fmt.Errorf("error retrieving vCenter by Id: %s", err)
		}

		return vcdClient.GetVCenterByName(vc.VSphereVCenter.Name)
	}

	shouldRefresh := d.Get("refresh_vcenter_on_read").(bool)
	shouldRefreshPolicies := d.Get("refresh_policies_on_read").(bool)
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel: labelVirtualCenter,
		// getEntityFunc:  vcdClient.GetVCenterById,// TODO: TM: use this function
		getEntityFunc:  fakeGetById, // TODO: TM: remove this function
		stateStoreFunc: setTmVcenterData,
		readHooks: []outerEntityHook[*govcd.VCenter]{
			refreshVcenter(shouldRefresh),               // vCenter read can optionally trigger "refresh" operation
			refreshVcenterPolicy(shouldRefreshPolicies), // vCenter read can optionally trigger "refresh policies" operation
		},
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdVcenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterById,
		preDeleteHooks: []outerEntityHook[*govcd.VCenter]{disableVcenter}, // vCenter must be disabled before deletion
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

// refreshVcenter triggers refresh on vCenter which is useful for reloading some of the vCenter
// components like Supervisors
func refreshVcenter(execute bool) outerEntityHook[*govcd.VCenter] {
	return func(v *govcd.VCenter) error {
		if execute {
			err := v.Refresh()
			if err != nil {
				return fmt.Errorf("error refreshing vCenter: %s", err)
			}
		}
		return nil
	}
}

// refreshVcenterPolicy triggers refresh on vCenter which is useful for reloading some of the
// vCenter components like Supervisors
func refreshVcenterPolicy(execute bool) outerEntityHook[*govcd.VCenter] {
	return func(v *govcd.VCenter) error {
		if execute {
			err := v.RefreshStorageProfiles()
			if err != nil {
				return fmt.Errorf("error refreshing Storage Policies: %s", err)
			}
		}
		return nil
	}
}

// autoTrustHostCertificate can automatically add host certificate to trusted ones
// * urlSchemaFieldName - Terraform schema field (TypeString) name that contains URL of entity
// * trustSchemaFieldName - Terraform schema field (TypeBool) name that defines if the certificate should be trusted
// Note. It will not add new entry if the certificate is already trusted
func autoTrustHostCertificate(urlSchemaFieldName, trustSchemaFieldName string) schemaHook {
	return func(vcdClient *VCDClient, d *schema.ResourceData) error {
		shouldExecute := d.Get(trustSchemaFieldName).(bool)
		if !shouldExecute {
			util.Logger.Printf("[DEBUG] Skipping certificate trust execution as '%s' is false", trustSchemaFieldName)
			return nil
		}
		schemaUrl := d.Get(urlSchemaFieldName).(string)
		parsedUrl, err := url.Parse(schemaUrl)
		if err != nil {
			return fmt.Errorf("error parsing provided url '%s': %s", schemaUrl, err)
		}

		_, err = vcdClient.AutoTrustCertificate(parsedUrl)
		if err != nil {
			return fmt.Errorf("error trusting '%s' certificate: %s", schemaUrl, err)
		}

		return nil
	}
}
