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
			"role_id": {
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
				Default:     false,
				Description: "Status of the service account.",
			},
			"file_name": {
				Type:        schema.TypeString,
				Required:    true,
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
	roleId := d.Get("role_id").(string)
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
	uuid := extractUuid(sa.ServiceAccount.ID)
	d.SetId(uuid)

	active := d.Get("active").(bool)
	useragent := vcdClient.Client.UserAgent
	err = updateServiceAccountStatus(sa, active, filename, useragent)
	if err != nil {
		return diag.Errorf("[Service Account create] error changing Service Account status: %s", err)
	}

	return resourceVcdServiceAccountRead(ctx, d, meta)
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

	// Role needs to be sent in URN format, and the role name needs to be percent-encoded
	roleId := d.Get("role_id").(string)

	softwareId := d.Get("software_id").(string)
	softwareVersion := d.Get("software_version").(string)
	uri := d.Get("uri").(string)
	filename := d.Get("file_name").(string)
	active := d.Get("active").(bool)

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

	if d.HasChange("active") {
		useragent := vcdClient.Client.UserAgent
		err = updateServiceAccountStatus(sa, active, filename, useragent)
		if err != nil {
			return diag.Errorf("[Service Account update] error updating Service Account status: %s", err)
		}
	}

	return resourceVcdServiceAccountRead(ctx, d, meta)
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

	dSet(d, "name", sa.ServiceAccount.Name)
	dSet(d, "software_id", sa.ServiceAccount.SoftwareID)
	dSet(d, "software_version", sa.ServiceAccount.SoftwareVersion)
	dSet(d, "role_id", sa.ServiceAccount.Role.ID)
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

	dSet(d, "name", sa.ServiceAccount.Name)
	dSet(d, "role", sa.ServiceAccount.Role.Name)
	dSet(d, "software_id", sa.ServiceAccount.SoftwareID)
	dSet(d, "software_version", sa.ServiceAccount.SoftwareVersion)
	dSet(d, "uri", sa.ServiceAccount.URI)
	dSet(d, "status", sa.ServiceAccount.Status)
	d.SetId(extractUuid(sa.ServiceAccount.ID))

	return []*schema.ResourceData{d}, nil
}
