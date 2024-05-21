package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcdOrgOidc defines the data source that reads Open ID Connect (OIDC) settings from an Organization
func datasourceVcdOrgOidc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgOidcRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization ID that has the OpenID Connect settings",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Client ID used when talking to the OpenID Connect Identity Provider",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Client Secret used when talking to the OpenID Connect Identity Provider",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether OpenID Connect authentication for the specified Organization is enabled or disabled",
			},
			"wellknown_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint from the OpenID Connect Identity Provider that serves all the configuration values",
			},
			"issuer_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The issuer identifier of the OpenID Connect Identity Provider",
			},
			"user_authorization_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user authorization endpoint of the OpenID Connect Identity Provider",
			},
			"access_token_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The access token endpoint of the OpenID Connect Identity Provider",
			},
			"userinfo_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user info endpoint of the OpenID Connect Identity Provider",
			},
			"prefer_id_token": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the claims from 'userinfo_endpoint' and the ID Token are combined (true) or not (false)",
			},
			"max_clock_skew_seconds": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum clock skew is the maximum allowable time difference between the client and server",
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "A set of scopes used with the OpenID Connect provider",
			},
			"claims_mapping": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A single configuration block that contains the claim mappings used with the OpenID Connect provider",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Email claim mapping",
						},
						"subject": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subject claim mapping",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last name claim mapping",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "First name claim mapping",
						},
						"full_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Full name claim mapping",
						},
						"groups": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Groups claim mapping",
						},
						"roles": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Roles claim mapping",
						},
					},
				},
			},
			"key": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "One or more configuration blocks that contain the keys used with the OIDC provider",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the key",
						},
						"algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Algorithm of the key",
						},
						"certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The certificate contents",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expiration date for the certificate",
						},
					},
				},
			},
			"key_refresh_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint used to refresh the keys",
			},
			"key_refresh_period_hours": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The frequency of key refresh",
			},
			"key_refresh_strategy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the strategy of key refresh",
			},
			"key_expire_duration_hours": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The expiration period of the key, only available if 'key_refresh_strategy=EXPIRE_AFTER'",
			},
			"ui_button_label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The label of the UI button of the login screen. Only available since VCD 10.5.1",
			},
			"redirect_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Redirect URI for this org",
			},
		},
	}
}

func datasourceVcdOrgOidcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgOidcRead(ctx, d, meta, "datasource")
}
