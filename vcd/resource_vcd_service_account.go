package vcd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdServiceAccountCreate,
		UpdateContext: resourceVcdServiceAccountUpdate,
		ReadContext:   resourceVcdServiceAccountRead,
		DeleteContext: resourceVcdServiceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdServiceAccountImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service account",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"software_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of software, e.g: 12345678-1234-5678-90ab-1234567890ab",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role ID of service account",
			},
			"software_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Version of software using the service account",
			},
			"uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URI of the client using the service account",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Status of the service account.",
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

func resourceVcdServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Service Account create] error getting Org from resource: %s", err)
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Service Account create] error getting AdminOrg from resource: %s", err)
	}

	saName := d.Get("name").(string)
	softwareId := d.Get("software_id").(string)
	softwareVersion := d.Get("software_version").(string)
	uri := d.Get("uri").(string)
	filename := d.Get("file_name").(string)

	// Role needs to be sent in URN format, and the role name needs to be percent-encoded
	// e.g urn:vcloud:role:Organization%20Administrator
	roleId := d.Get("role").(string)
	role, err := adminOrg.GetRoleById(roleId)
	if err != nil {
		return diag.Errorf("[Service Account create] error getting Role: %s", err)
	}
	escapedRoleName := url.PathEscape(role.Role.Name)
	formattedRole := "urn:vcloud:role:" + escapedRoleName

	sa, err := vcdClient.CreateServiceAccount(org.Org.Name, saName, formattedRole, softwareId, softwareVersion, uri)
	if err != nil {
		return diag.Errorf("[Service Account create] error creating Service Account: %s", err)
	}
	d.SetId(sa.ServiceAccount.ID)

	active := d.Get("active").(bool)
	allowTokenFile := d.Get("allow_token_file").(bool)
	useragent := vcdClient.Client.UserAgent
	err = updateServiceAccountStatus(sa, active, filename, useragent)
	if err != nil {
		return diag.Errorf("[Service Account create] error changing Service Account status: %s", err)
	}

	var diagnostics diag.Diagnostics
	if active && !allowTokenFile {
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + filename + " should be considered sensitive information.",
			Detail: "The file " + filename + " containing the initial service account API " +
				"HAS BEEN UPDATED with a freshly generated token. The initial token was invalidated and the " +
				"token currently in the file will be invalidated at the next usage. In the meantime, it is " +
				"usable by anyone to run operations to the current VCD. As such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_token_file = true\n\nto the provider settings.",
		})
	}
	return append(resourceVcdServiceAccountRead(ctx, d, meta), diagnostics...)
}

func resourceVcdServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Service Account update] error getting Org from resource: %s", err)
	}

	sa, err := org.GetServiceAccountById(d.Id())
	if err != nil {
		return diag.Errorf("[Service Account update] error getting Service Account: %s", err)
	}

	roleId := d.Get("role").(string)
	softwareId := d.Get("software_id").(string)
	softwareVersion := d.Get("software_version").(string)
	uri := d.Get("uri").(string)
	active := d.Get("active").(bool)
	allowTokenFile := d.Get("allow_token_file").(bool)
	filename := d.Get("file_name").(string)

	saConfig := &types.ServiceAccount{
		SoftwareID:      softwareId,
		SoftwareVersion: softwareVersion,
		URI:             uri,
		Role: &types.OpenApiReference{
			ID: roleId,
		},
	}

	sa, err = sa.Update(saConfig)
	if err != nil {
		return diag.Errorf("[Service Account update] error updating Service Account: %s", err)
	}

	var diagnostics diag.Diagnostics
	if d.HasChange("active") {
		if active && filename == "" {
			return diag.Errorf("[Service Account update] filename must be set on account activation")
		}
		useragent := vcdClient.Client.UserAgent
		err = updateServiceAccountStatus(sa, active, filename, useragent)
		if err != nil {
			return diag.Errorf("[Service Account update] error updating Service Account status: %s", err)
		}

		// Output a warning during update if a service account was set to active with allow_token_file
		// set to false
		if active && !allowTokenFile {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "The file " + filename + " should be considered sensitive information.",
				Detail: "The file " + filename + " containing the initial service account API " +
					"HAS BEEN UPDATED with a freshly generated token. The initial token was invalidated and the " +
					"token currently in the file will be invalidated at the next usage. In the meantime, it is " +
					"usable by anyone to run operations to the current VCD. As such, it should be considered SENSITIVE INFORMATION. " +
					"If you would like to remove this warning, add\n\n" + "	allow_token_file = true\n\nto the provider settings.",
			})
		}
	}

	return append(resourceVcdServiceAccountRead(ctx, d, meta), diagnostics...)
}

func updateServiceAccountStatus(sa *govcd.ServiceAccount, active bool, filename, useragent string) error {
	if active {
		err := sa.Authorize()
		if err != nil {
			return fmt.Errorf("error authorizing Service Account: %s", err)
		}
		err = sa.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing Service Account: %s", err)
		}
		err = sa.Grant()
		if err != nil {
			return fmt.Errorf("error granting Service Account: %s", err)
		}
		err = sa.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing Service Account: %s", err)
		}
		initialApiToken, err := sa.GetInitialApiToken()
		if err != nil {
			return fmt.Errorf("error activating Service Account: %s", err)
		}
		err = sa.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing Service Account: %s", err)
		}
		err = govcd.SaveServiceAccountToFile(filename, useragent, initialApiToken)
		if err != nil {
			return fmt.Errorf("error saving service account api token to file: %s", err)
		}
	} else {
		err := sa.Revoke()
		if err != nil {
			return fmt.Errorf("error revoking Service Account: %s", err)
		}
		err = sa.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing Service Account: %s", err)
		}
	}

	return nil
}

func resourceVcdServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdServiceAccountRead(ctx, d, meta, "resource")
}

func genericVcdServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}
	var sa *govcd.ServiceAccount
	if d.Id() != "" {
		sa, err = org.GetServiceAccountById(d.Id())
	} else {
		saName := d.Get("name").(string)
		sa, err = org.GetServiceAccountByName(saName)
	}
	if govcd.ContainsNotFound(err) {
		// When parent Edge Gateway is not found - this resource is also not found and should be
		// removed from state
		if origin == "datasource" {
			return diag.Errorf("[Service Account DS read] error retrieving Service Account: %s", err)
		}
		d.SetId("")
		log.Printf("[DEBUG] Service Account no longer exists. Removing from tfstate")
		return nil
	}

	if err != nil {
		return diag.Errorf("[Service Account read] error: %s", err)
	}

	d.SetId(sa.ServiceAccount.ID)
	dSet(d, "name", sa.ServiceAccount.Name)
	dSet(d, "software_id", sa.ServiceAccount.SoftwareID)
	dSet(d, "software_version", sa.ServiceAccount.SoftwareVersion)
	dSet(d, "role", sa.ServiceAccount.Role.ID)
	dSet(d, "uri", sa.ServiceAccount.URI)
	if sa.ServiceAccount.Status == "ACTIVE" {
		dSet(d, "active", true)
	} else {
		dSet(d, "active", false)
	}

	return nil
}

func resourceVcdServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}
	sa, err := org.GetServiceAccountById(d.Id())
	if err != nil {
		return diag.Errorf("[Service Account delete] error retrieving Service Account: %s", err)
	}

	err = sa.Delete()
	if err != nil {
		return diag.Errorf("[Service Account delete] error deleting Service Account: %s", err)
	}

	return resourceVcdServiceAccountRead(ctx, d, meta)
}

func resourceVcdServiceAccountImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] API token import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.service-account-name")
	}
	orgName := resourceURI[0]
	saName := resourceURI[1]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving org: %s", err)
	}

	sa, err := org.GetServiceAccountByName(saName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving service account: %s", err)
	}

	d.SetId(sa.ServiceAccount.ID)
	dSet(d, "name", sa.ServiceAccount.Name)
	dSet(d, "role", sa.ServiceAccount.Role.Name)
	dSet(d, "software_id", sa.ServiceAccount.SoftwareID)
	dSet(d, "software_version", sa.ServiceAccount.SoftwareVersion)
	dSet(d, "uri", sa.ServiceAccount.URI)
	dSet(d, "role", sa.ServiceAccount.Role.ID)

	return []*schema.ResourceData{d}, nil
}
