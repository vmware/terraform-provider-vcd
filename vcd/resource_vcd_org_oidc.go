package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
)

// resourceVcdOrgOidc defines the resource that manages Open ID Connect (OIDC) settings for an existing Organization
func resourceVcdOrgOidc() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdOrgOidcRead,
		CreateContext: resourceVcdOrgOidcCreate,
		UpdateContext: resourceVcdOrgOidcUpdate,
		DeleteContext: resourceVcdOrgOidcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOrgOidcImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client ID to use when talking to the OIDC Identity Provider",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client Secret to use when talking to the OIDC Identity Provider",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enables or disables OIDC Authentication for the Organization specified in 'org_id'",
			},
			"wellknown_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Endpoint from the OIDC Identity Provider that serves all the configuration values",
			},
			"issuer_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"user_authorization_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"access_endpoint_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"userinfo_endpoint_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"max_clock_skew": {
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true, // Can be obtained with "wellknown_endpoint"
				Description:      "",
				ValidateDiagFunc: minimumValue(0, "'max_clock_skew' must be higher than or equal to 0"),
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"claims_mapping": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					// Check that keys match only 'subject', 'last_name', etc
					return nil
				},
			},
			"key": {
				Type:        schema.TypeSet,
				MinItems:    1,
				Required:    true, // FIXME: Can be obtained with "wellknown_endpoint"
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"pem_file": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expiration": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"key_refresh_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
			},
			"key_refresh_period": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"key_refresh_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "",
				ValidateFunc: validation.StringInSlice([]string{"add", "replace", "expire_after"}, false),
			},
			"ui_button_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"redirect_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Redirect URI for this org",
			},
		},
	}
}

func resourceVcdOrgOidcCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)

	_, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect %s] error searching for Org '%s': %s", operation, orgId, err)
	}
	return resourceVcdOrgOidcRead(ctx, d, meta)
}
func resourceVcdOrgOidcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdOrgOidcCreateOrUpdate(ctx, d, meta, "create")
}
func resourceVcdOrgOidcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdOrgOidcCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcdOrgOidcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgOidcRead(ctx, d, meta, "resource")
}

func genericVcdOrgOidcRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgId)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find Organization '%s' Open ID Connect settings: %s. Removing from state", orgId, err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect read] unable to find Organization '%s': %s", orgId, err)
	}

	settings, err := adminOrg.GetOpenIdConnectSettings()
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect read] unable to read Organization '%s' OIDC settings: %s", orgId, err)
	}

	dSet(d, "client_id", settings.ClientId)
	d.SetId(adminOrg.AdminOrg.ID)

	return nil
}

func resourceVcdOrgOidcDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	_, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect delete] error searching for Organization '%s': %s", orgId, err)
	}

	return nil
}

// resourceVcdOrgOidcImport is responsible for importing the resource.
// The only parameter needed is the Org identifier, which could be either the Org name or its ID
func resourceVcdOrgOidcImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	orgNameOrId := d.Id()

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgNameOrId)
	if err != nil {
		return nil, fmt.Errorf("[Organization Open ID Connect import] error searching for Organization '%s': %s", orgNameOrId, err)
	}

	dSet(d, "org_id", adminOrg.AdminOrg.ID)

	d.SetId(adminOrg.AdminOrg.ID)
	return []*schema.ResourceData{d}, nil
}
