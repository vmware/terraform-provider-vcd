package vcd

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelTmVirtualCenter = "Tenant Manager vCenter Server"

func resourceVcdTmVcenter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmVcenterCreate,
		ReadContext:   resourceVcdTmVcenterRead,
		UpdateContext: resourceVcdTmVcenterUpdate,
		DeleteContext: resourceVcdTmVcenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmVcenterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("URL including port of %s", labelTmVirtualCenter),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelTmVirtualCenter),
			},
			"refresh_vcenter_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should be refreshed on every read operation", labelTmVirtualCenter),
			},
			"refresh_policies_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should refresh Policies on every read operation", labelTmVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Username of %s", labelTmVirtualCenter),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password of %s", labelTmVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Should the %s be enabled", labelTmVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelTmVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelTmVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelTmVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelTmVirtualCenter),
			},
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelTmVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelTmVirtualCenter),
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelTmVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelTmVirtualCenter),
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s hostname", labelTmVirtualCenter),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
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

func setTmVcenterData(_ *VCDClient, d *schema.ResourceData, v *govcd.VCenter) error {
	if v == nil || v.VSphereVCenter == nil {
		return fmt.Errorf("nil object for %s", labelTmVirtualCenter)
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
	dSet(d, "connection_status", v.VSphereVCenter.ListenerState)
	dSet(d, "cluster_health_status", v.VSphereVCenter.ClusterHealthStatus)
	dSet(d, "vcenter_version", v.VSphereVCenter.VcVersion)
	dSet(d, "uuid", v.VSphereVCenter.Uuid)
	host, err := url.Parse(v.VSphereVCenter.Url)
	if err != nil {
		return fmt.Errorf("error parsing URL for storing 'vcenter_host': %s", err)
	}
	dSet(d, "vcenter_host", host.Host)

	// Status is a derivative value that was present in XML Query API, but is no longer maintained
	// The value was derived from multiple fields based on a complex logic. Instead, evaluating if
	// vCenter is ready for operations, would be to rely on `is_enabled`, `is_connected` and
	// optionally `cluster_health_status` fields.
	//
	// The `status` is a rough approximation of this value
	dSet(d, "status", "NOT_READY")
	if v.VSphereVCenter.IsConnected && v.VSphereVCenter.ListenerState == "CONNECTED" {
		dSet(d, "status", "READY")
	}

	d.SetId(v.VSphereVCenter.VcId)

	return nil
}

func resourceVcdTmVcenterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:      labelTmVirtualCenter,
		getTypeFunc:      getTmVcenterType,
		stateStoreFunc:   setTmVcenterData,
		createFunc:       vcdClient.CreateVcenter,
		resourceReadFunc: resourceVcdTmVcenterRead,
		// certificate should be trusted for the vCenter to work
		preCreateHooks: []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmVcenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// return immediately if only flags are updated
	if !d.HasChangesExcept("refresh_vcenter_on_read", "refresh_policies_on_read") {
		return nil
	}

	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:      labelTmVirtualCenter,
		getTypeFunc:      getTmVcenterType,
		getEntityFunc:    vcdClient.GetVCenterById,
		resourceReadFunc: resourceVcdTmVcenterRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		entityLabel: labelTmVirtualCenter,
		// getEntityFunc:  vcdClient.GetVCenterById,// TODO: TM: use this function
		getEntityFunc:  fakeGetById, // TODO: TM: remove this function
		stateStoreFunc: setTmVcenterData,
		readHooks: []outerEntityHook[*govcd.VCenter]{
			// TODO: TM ensure that the vCenter listener state is "CONNECTED"  before triggering
			// refresh as it will fail otherwise. At the moment it has a delay before it becomes
			// CONNECTED after creation task succeeds. It should not be needed once vCenter creation
			// task ensures that the listener is connected.
			waitForListenerStatusConnected,

			refreshVcenter(shouldRefresh),               // vCenter read can optionally trigger "refresh" operation
			refreshVcenterPolicy(shouldRefreshPolicies), // vCenter read can optionally trigger "refresh policies" operation
		},
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmVcenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelTmVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterById,
		preDeleteHooks: []outerEntityHook[*govcd.VCenter]{disableVcenter}, // vCenter must be disabled before deletion
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmVcenterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	v, err := vcdClient.GetVCenterByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name: %s", labelTmVirtualCenter, err)
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
			err := v.RefreshVcenter()
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

// TODO: TM: should not be required because a successful vCenter creation task should work
func waitForListenerStatusConnected(v *govcd.VCenter) error {
	for c := 0; c < 20; c++ {
		err := v.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing vCenter: %s", err)
		}

		if v.VSphereVCenter.ListenerState == "CONNECTED" {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("failed waiting for listener state to become 'CONNECTED', got '%s'", v.VSphereVCenter.ListenerState)
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
