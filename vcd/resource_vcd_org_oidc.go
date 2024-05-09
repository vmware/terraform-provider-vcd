package vcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"log"
	"strings"
	"time"
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
				Description: "ID of the Organization that will have the OpenID Connect settings configured",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client ID to use when talking to the OpenID Connect Identity Provider",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Client Secret to use when talking to the OpenID Connect Identity Provider",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enables or disables OpenID Connect authentication for the specified Organization",
			},
			"wellknown_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Endpoint from the OpenID Connect Identity Provider that serves all the configuration values",
			},
			"issuer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "The issuer identifier of the OpenID Connect Identity Provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained issuer identifier",
				AtLeastOneOf: []string{"issuer_id", "wellknown_endpoint"},
			},
			"user_authorization_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "The user authorization endpoint of the OpenID Connect Identity Provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained user authorization endpoint",
				AtLeastOneOf: []string{"user_authorization_endpoint", "wellknown_endpoint"},
			},
			"access_token_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "The access token endpoint of the OpenID Connect Identity Provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained access token endpoint",
				AtLeastOneOf: []string{"access_token_endpoint", "wellknown_endpoint"},
			},
			"userinfo_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "The user info endpoint of the OpenID Connect Identity Provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained user info endpoint",
				AtLeastOneOf: []string{"userinfo_endpoint", "wellknown_endpoint"},
			},
			"prefer_id_token": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "If you want to combine claims from 'userinfo_endpoint' and the ID Token, set this to 'true'. " +
					"The identity providers do not provide all the required claims set in 'userinfo_endpoint'." +
					"By setting this argument to 'true', VMware Cloud Director can fetch and consume claims from both sources",
			},
			"max_clock_skew_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
				Description: "The maximum clock skew is the maximum allowable time difference between the client and server. " +
					"This time compensates for any small-time differences in the timestamps when verifying tokens",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "A set of scopes to use with the OpenID Connect provider. " +
					"They are used to authorize access to user details, by defining the permissions that the access tokens have to access user information. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained scopes",
			},
			"claims_mapping": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "A single configuration block that specifies the claim mappings to use with the OpenID Connect provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained claim mappings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Email claim mapping",
						},
						"subject": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Subject claim mapping",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Last name claim mapping",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "First name claim mapping",
						},
						"full_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Full name claim mapping",
						},
						"groups": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Groups claim mapping",
						},
						"roles": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Roles claim mapping",
						},
					},
				},
			},
			"key": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "One or more configuration blocks that specify the keys to use with the OIDC provider. " +
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the key",
						},
						"algorithm": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Algorithm of the key, either RSA or EC",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"RSA", "EC"}, false)),
						},
						"certificate": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The certificate contents",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Expiration date for the certificate",
							ValidateDiagFunc: validation.AnyDiag(
								validation.ToDiagFunc(validation.IsRFC3339Time),
								validation.ToDiagFunc(validation.StringIsEmpty)),
						},
					},
				},
				// This function is required because the default hash function makes
				// the Terraform plans to be dirty all the time, trying to remove the given key even if
				// it didn't change.
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					_, err := buf.WriteString(m["id"].(string))
					if err != nil {
						util.Logger.Printf("[ERROR] error writing to string: %s", err)
					}
					_, err = buf.WriteString(m["algorithm"].(string))
					if err != nil {
						util.Logger.Printf("[ERROR] error writing to string: %s", err)
					}
					_, err = buf.WriteString(strings.ReplaceAll(m["certificate"].(string), "\n", ""))
					if err != nil {
						util.Logger.Printf("[ERROR] error writing to string: %s", err)
					}
					if m["expiration_date"].(string) != "" {
						t, err := time.Parse(time.RFC3339, m["expiration_date"].(string))
						if err != nil {
							util.Logger.Printf("[ERROR] error parsing date: %s", err)
						}
						_, err = buf.WriteString(t.Format(time.RFC3339))
						if err != nil {
							util.Logger.Printf("[ERROR] error writing to string: %s", err)
						}
					} else {
						buf.WriteString("nil")
					}
					return hashcodeString(buf.String())
				},
			},
			"key_refresh_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "Endpoint used to refresh the keys. If 'wellknown_endpoint' is set, then this argument" +
					"will override the obtained endpoint",
				RequiredWith: []string{"key_refresh_period_hours", "key_refresh_strategy"},
			},
			"key_refresh_period_hours": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "Defines the frequency of key refresh. Maximum is 720 hours",
				RequiredWith:     []string{"key_refresh_endpoint"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 720)),
			},
			"key_refresh_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Defines the strategy of key refresh",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ADD", "REPLACE", "EXPIRE_AFTER"}, false)),
				RequiredWith:     []string{"key_refresh_endpoint"},
			},
			"key_expire_duration_hours": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "Defines the expiration period of the key, only when 'key_refresh_strategy=EXPIRE_AFTER'. Maximum is 24 hours",
				RequiredWith:     []string{"key_refresh_endpoint", "key_refresh_strategy"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 24)),
			},
			"ui_button_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customizes the label of the UI button of the login screen. Only available since VCD 10.5.1",
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

	// Runtime validations
	isWellKnownEndpointUsed := d.Get("wellknown_endpoint").(string) != ""
	scopes := d.Get("scopes").(*schema.Set).List()
	if !isWellKnownEndpointUsed && len(scopes) == 0 {
		return diag.Errorf("[Organization Open ID Connect %s] 'scopes' cannot be empty when a well-known endpoint is not used", operation)
	}

	if _, ok := d.GetOk("key_expire_duration_hours"); ok && d.Get("key_refresh_strategy") != "EXPIRE_AFTER" {
		return diag.Errorf("[Organization Open ID Connect %s] 'key_expire_duration_hours' can only be used when 'key_refresh_strategy=EXPIRE_AFTER', but key_refresh_strategy=%s", operation, d.Get("key_refresh_strategy"))
	}
	if _, ok := d.GetOk("key_expire_duration_hours"); !ok && d.Get("key_refresh_strategy") == "EXPIRE_AFTER" {
		return diag.Errorf("[Organization Open ID Connect %s] 'key_refresh_strategy=EXPIRE_AFTER' requires 'key_expire_duration_hours' to be set", operation)
	}

	if _, ok := d.GetOk("ui_button_label"); ok && vcdClient.Client.APIVCDMaxVersionIs("< 38.1") {
		return diag.Errorf("[Organization Open ID Connect %s] 'ui_button_label' can only be used since VCD 10.5.1", operation)
	}
	if _, ok := d.GetOk("prefer_id_token"); ok && vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		return diag.Errorf("[Organization Open ID Connect %s] 'prefer_id_token' can only be used since VCD 10.4.1", operation)
	}
	// End of validations

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
		Scope:                      convertTypeListToSliceOfStrings(scopes),
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
				Key:            key["certificate"].(string),
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

	// Attributes that depend on the VCD version.
	// UsePKCE and SendClientCredentialsAsAuthorizationHeader are not used in UI yet
	if vcdClient.Client.APIVCDMaxVersionIs(">= 37.1") {
		settings.EnableIdTokenClaims = addrOf(d.Get("prefer_id_token").(bool))
	}
	if vcdClient.Client.APIVCDMaxVersionIs(">= 38.1") {
		settings.CustomUiButtonLabel = addrOf(d.Get("ui_button_label").(string))
	}

	_, err = org.SetOpenIdConnectSettings(settings)
	if err != nil {
		return diag.Errorf("[Organization Open ID Connect %s] Could not set OIDC settings: %s", operation, err)
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
			key["certificate"] = keyConfig.Key
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
