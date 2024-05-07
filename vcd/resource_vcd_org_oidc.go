package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"os"
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
				Sensitive:   true,
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Can be obtained with "wellknown_endpoint"
				Description:  "If 'wellknown_endpoint' is set, this attribute overrides the obtained Issuer ID",
				AtLeastOneOf: []string{"issuer_id", "wellknown_endpoint"},
			},
			"user_authorization_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Can be obtained with "wellknown_endpoint"
				Description:  "If 'wellknown_endpoint' is set, this attribute overrides the obtained User Authorization endpoint",
				AtLeastOneOf: []string{"user_authorization_endpoint", "wellknown_endpoint"},
			},
			"access_token_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Can be obtained with "wellknown_endpoint"
				Description:  "If 'wellknown_endpoint' is set, this attribute overrides the obtained Access Token endpoint",
				AtLeastOneOf: []string{"access_token_endpoint", "wellknown_endpoint"},
			},
			"userinfo_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Can be obtained with "wellknown_endpoint"
				Description:  "If 'wellknown_endpoint' is set, this attribute overrides the obtained Userinfo endpoint",
				AtLeastOneOf: []string{"userinfo_endpoint", "wellknown_endpoint"},
			},
			"prefer_id_token": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "",
			},
			"max_clock_skew_seconds": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          60,
				Description:      "",
				ValidateDiagFunc: minimumValue(0, "'max_clock_skew_seconds' must be higher than or equal to 0"),
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "If 'wellknown_endpoint' is set, this attribute overrides the obtained Scopes",
			},
			"claims_mapping": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "If 'wellknown_endpoint' is set, this attribute overrides the obtained Claim mappings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"subject": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"full_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"groups": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
						"roles": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "",
						},
					},
				},
			},
			"key": {
				Type:        schema.TypeSet,
				MinItems:    1,
				Optional:    true,
				Computed:    true, // Can be obtained with "wellknown_endpoint"
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"algorithm": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "",
							ValidateFunc: validation.StringInSlice([]string{"RSA", "EC"}, false),
						},
						"pem_file": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"pem": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
					},
				},
			},
			"key_refresh_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Can be obtained with "wellknown_endpoint"
				Description:  "",
				RequiredWith: []string{"key_refresh_period_hours", "key_refresh_strategy"},
			},
			"key_refresh_period_hours": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "",
				RequiredWith: []string{"key_refresh_endpoint", "key_refresh_strategy"},
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					return nil
					// Maximum 30 days: 24*30
				},
			},
			"key_refresh_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "",
				ValidateFunc: validation.StringInSlice([]string{"ADD", "REPLACE", "EXPIRE_AFTER"}, false),
				RequiredWith: []string{"key_refresh_endpoint", "key_refresh_period_hours"},
			},
			"key_expire_duration_hours": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "",
				RequiredWith: []string{"key_refresh_strategy"},
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					return nil
					// Only if key_refresh_strategy=EXPIRE_AFTER
					// Maximum 1 days: 24
				},
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

	org, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect %s] error searching for Org '%s': %s", operation, orgId, err)
	}

	isWellKnownEndpointUsed := d.Get("wellknown_endpoint").(string) != ""
	scopes := d.Get("scopes").(*schema.Set).List()
	if !isWellKnownEndpointUsed && len(scopes) == 0 {
		return diag.Errorf("[Organization Open ID Connect %s] 'scopes' cannot be empty when a well-known endpoint is not used", operation)
	}

	settings := types.OrgOAuthSettings{
		IssuerId:                   d.Get("issuer_id").(string),
		Enabled:                    d.Get("enabled").(bool),
		ClientId:                   d.Get("client_id").(string),
		ClientSecret:               d.Get("client_secret").(string),
		UserAuthorizationEndpoint:  d.Get("user_authorization_endpoint").(string),
		AccessTokenEndpoint:        d.Get("access_token_endpoint").(string),
		UserInfoEndpoint:           d.Get("userinfo_endpoint").(string),
		MaxClockSkew:               d.Get("max_clock_skew_seconds").(int),
		JwksUri:                    d.Get("key_refresh_endpoint").(string),
		AutoRefreshKey:             d.Get("key_refresh_endpoint").(string) != "" && d.Get("key_refresh_strategy").(string) != "",
		KeyRefreshStrategy:         d.Get("key_refresh_strategy").(string),
		KeyRefreshFrequencyInHours: d.Get("key_refresh_period_hours").(int),
		WellKnownEndpoint:          d.Get("wellknown_endpoint").(string),
		EnableIdTokenClaims:        d.Get("prefer_id_token").(bool),
		CustomUiButtonLabel:        d.Get("ui_button_label").(string),
		Scope:                      convertTypeListToSliceOfStrings(scopes),
		// UsePKCE and SendClientCredentialsAsAuthorizationHeader are not used yet
	}

	// Key configurations: OAuthKeyConfigurations
	keyList := d.Get("key").(*schema.Set).List()
	if len(keyList) == 0 && !isWellKnownEndpointUsed {
		return diag.Errorf("[Organization Open ID Connect %s] error reading keys, either set a 'key' block or set 'wellknown_endpoint' to obtain this information", operation)
	}
	if len(keyList) > 0 {
		oAuthKeyConfigurations := make([]types.OAuthKeyConfiguration, len(keyList))
		for i, k := range keyList {
			key := k.(map[string]interface{})
			oAuthKeyConfigurations[i] = types.OAuthKeyConfiguration{
				KeyId:          key["id"].(string),
				Algorithm:      key["algorithm"].(string),
				ExpirationDate: key["expiration_date"].(string),
			}
			filePath, isSetPemFile := key["pem_file"]
			pem, isSetPem := key["pem"]
			if isSetPemFile && filePath != "" {
				// If there's a PEM file set in the config, we give it priority
				pemContents, err := os.ReadFile(filePath.(string))
				if err != nil {
					return diag.Errorf("[Organization Open ID Connect %s] error reading PEM file '%s': %s", operation, filePath, err)
				}
				oAuthKeyConfigurations[i].Key = string(pemContents)
			} else if isSetPem && pem != "" {
				// Otherwise, the PEM contents may have arrived in the computed field with a well-known endpoint
				oAuthKeyConfigurations[i].Key = pem.(string)
			} else {
				return diag.Errorf("[Organization Open ID Connect %s] a PEM file is required to set up a key", operation)
			}
		}
		settings.OAuthKeyConfigurations = &types.OAuthKeyConfigurationsList{
			OAuthKeyConfiguration: oAuthKeyConfigurations,
		}
	}

	// Claims mapping: OIDCAttributeMapping: Subject, Email, Full name, First name and Last name are mandatory
	claimsMapping := d.Get("claims_mapping").([]interface{})
	if len(claimsMapping) == 0 && !isWellKnownEndpointUsed {
		return diag.Errorf("[Organization Open ID Connect %s] error reading claims, either set a 'claims_mapping' block or set 'wellknown_endpoint' to obtain this information", operation)
	}
	if len(claimsMapping) > 0 {
		var oidcAttributeMapping types.OIDCAttributeMapping
		mappingEntry := claimsMapping[0].(map[string]interface{})
		oidcAttributeMapping.SubjectAttributeName = mappingEntry["subject"].(string)
		oidcAttributeMapping.EmailAttributeName = mappingEntry["email"].(string)
		oidcAttributeMapping.FullNameAttributeName = mappingEntry["full_name"].(string)
		oidcAttributeMapping.FirstNameAttributeName = mappingEntry["first_name"].(string)
		oidcAttributeMapping.LastNameAttributeName = mappingEntry["last_name"].(string)
		oidcAttributeMapping.GroupsAttributeName = mappingEntry["groups"].(string)
		oidcAttributeMapping.RolesAttributeName = mappingEntry["roles"].(string)
		settings.OIDCAttributeMapping = &oidcAttributeMapping
	}

	_, err = org.SetOpenIdConnectSettings(settings)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect create] Could not set OIDC settings: %s", err)
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
	dSet(d, "client_secret", settings.ClientSecret)
	dSet(d, "enabled", settings.Enabled)
	dSet(d, "wellknown_endpoint", settings.WellKnownEndpoint)
	dSet(d, "issuer_id", settings.IssuerId)
	dSet(d, "prefer_id_token", settings.EnableIdTokenClaims)
	dSet(d, "user_authorization_endpoint", settings.UserAuthorizationEndpoint)
	dSet(d, "access_token_endpoint", settings.AccessTokenEndpoint)
	dSet(d, "userinfo_endpoint", settings.UserInfoEndpoint)
	dSet(d, "max_clock_skew_seconds", settings.MaxClockSkew)
	err = d.Set("scopes", settings.Scope)
	if err != nil {
		return diag.FromErr(err)
	}
	if settings.OIDCAttributeMapping != nil {
		claims := make([]interface{}, 1)
		claim := map[string]interface{}{}
		claim["email"] = settings.OIDCAttributeMapping.EmailAttributeName
		claim["subject"] = settings.OIDCAttributeMapping.SubjectAttributeName
		claim["last_name"] = settings.OIDCAttributeMapping.LastNameAttributeName
		claim["first_name"] = settings.OIDCAttributeMapping.FirstNameAttributeName
		claim["full_name"] = settings.OIDCAttributeMapping.FullNameAttributeName
		claim["groups"] = settings.OIDCAttributeMapping.GroupsAttributeName
		claim["roles"] = settings.OIDCAttributeMapping.RolesAttributeName
		claims[0] = claim
		err = d.Set("claims_mapping", claims)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if settings.OAuthKeyConfigurations != nil {
		keyConfigurations := settings.OAuthKeyConfigurations.OAuthKeyConfiguration
		keyConfigs := make([]map[string]interface{}, len(keyConfigurations))
		for i, keyConfig := range keyConfigurations {
			key := map[string]interface{}{}
			key["id"] = keyConfig.KeyId
			key["algorithm"] = keyConfig.Algorithm
			key["pem"] = keyConfig.Key
			key["expiration_date"] = keyConfig.ExpirationDate
			keyConfigs[i] = key
		}
		err = d.Set("key", keyConfigs)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	dSet(d, "key_refresh_endpoint", settings.JwksUri)
	dSet(d, "key_refresh_period_hours", settings.KeyRefreshFrequencyInHours)
	dSet(d, "key_refresh_strategy", settings.KeyRefreshStrategy)
	dSet(d, "key_expire_duration_hours", settings.KeyExpireDurationInHours)
	dSet(d, "ui_button_label", settings.CustomUiButtonLabel)
	dSet(d, "redirect_uri", settings.OrgRedirectUri)
	d.SetId(adminOrg.AdminOrg.ID)

	return nil
}

func resourceVcdOrgOidcDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect delete] error searching for Organization '%s': %s", orgId, err)
	}

	err = adminOrg.DeleteOpenIdConnectSettings()
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect delete] error deleting OIDC settings for Organization '%s': %s", orgId, err)
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