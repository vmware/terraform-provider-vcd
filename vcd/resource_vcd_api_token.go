package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdApiToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdApiTokenCreate,
		UpdateContext: resourceVcdApiTokenUpdate,
		ReadContext:   resourceVcdApiTokenRead,
		DeleteContext: resourceVcdApiTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdApiTokenImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of API token",
			},
			"file_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the file that the API token will be saved to",
			},
			"allow_token_file": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Set this to true if you understand the security risks of using" +
					" API token files and would like to suppress the warnings",
			},
		},
	}
}

func resourceVcdApiTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// System Admin can't create API tokens outside SysOrg,
	// just as Org admins can't create API tokens in other Orgs
	org := vcdClient.SysOrg
	if org == "" {
		org = vcdClient.Org
	}

	tokenName := d.Get("name").(string)
	token, err := vcdClient.CreateToken(org, tokenName)
	if err != nil {
		return diag.Errorf("[API token create] error creating API token: %s", err)
	}
	d.SetId(token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	if err != nil {
		return diag.Errorf("[API token create] error getting refresh token from API token: %s", err)
	}

	filename := d.Get("file_name").(string)
	allowTokenFile := d.Get("allow_token_file").(bool)

	var diagnostics diag.Diagnostics
	if !allowTokenFile {
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + filename + " should be considered sensitive information.",
			Detail: "The file " + filename + " contains an API token which is " +
				"usable by anyone to run operations to the current VCD. As such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_token_file = true\n\nto the provider settings.",
		})
	}

	if filename == "" {
		return diag.Errorf("[API token create] file_name must be set on creation")
	}

	err = govcd.SaveApiTokenToFile(filename, vcdClient.Client.UserAgent, apiToken)
	if err != nil {
		return diag.Errorf("[API token create] error saving API token to file: %s", err)
	}

	return append(diagnostics, resourceVcdApiTokenRead(ctx, d, meta)...)
}

// There are no fields that can be updated after creation
func resourceVcdApiTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdApiTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	token, err := vcdClient.GetTokenById(d.Id())
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] API token no longer exists. Removing from tfstate")
	}
	if err != nil {
		return diag.Errorf("[API token read] error getting API token: %s", err)
	}

	d.SetId(token.Token.ID)
	dSet(d, "name", token.Token.Name)

	return nil
}

func resourceVcdApiTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	token, err := vcdClient.GetTokenById(d.Id())
	if err != nil {
		return diag.Errorf("[API token delete] error getting API token: %s", err)
	}

	err = token.Delete()
	if err != nil {
		return diag.Errorf("[API token delete] error deleting API token: %s", err)
	}
	d.SetId("")

	return nil
}

func resourceVcdApiTokenImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] API token import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 1 {
		return nil, fmt.Errorf("resource name must be specified as token-name")
	}
	tokenName := resourceURI[0]

	vcdClient := meta.(*VCDClient)

	sessionInfo, err := vcdClient.Client.GetSessionInfo()
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("error getting username: %s", err)
	}

	token, err := vcdClient.GetTokenByNameAndUsername(tokenName, sessionInfo.User.Name)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("error getting token by name: %s", err)
	}

	d.SetId(token.Token.ID)
	dSet(d, "name", token.Token.Name)

	return []*schema.ResourceData{d}, nil
}
