package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
	"os"
)

// resourceVcdOrgSaml handles Org SAML settings
func resourceVcdOrgSaml() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdOrgSamlRead,
		CreateContext: resourceVcdOrgSamlCreate,
		UpdateContext: resourceVcdOrgSamlUpdate,
		DeleteContext: resourceVcdOrgSamlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOrgSamlImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Enable SAML authentication. When this option is set, authentication is deferred to the SAML identity provider",
			},
			"entity_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Your service provider entity ID. Once you set this field, it cannot be changed back to empty.",
			},
			"identity_provider_metadata_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the file containing the metadata from the identity provider",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional email attribute name",
			},
			"user_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional username attribute name",
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional first name attribute name",
			},
			"surname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional surname attribute name",
			},
			"full_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional full name attribute name",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional group attribute name",
			},
			"role": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional role attribute name",
			},
		},
	}
}

func resourceVcdOrgSamlCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("resource vcd_org_ldap requires System administrator privileges")
	}
	orgId := d.Get("org_id").(string)
	entityId := d.Get("entity_id").(string)
	enabled := d.Get("enabled").(bool)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org SAML %s] error searching for Org %s: %s", origin, orgId, err)
	}

	fileName := d.Get("identity_provider_metadata_file").(string)

	metadataText, err := os.ReadFile(fileName) // #nosec G304 -- We need user input for this file
	if err != nil {
		return diag.Errorf("[ORG SAML %s %s] error reading metadata file %s: %s", origin, adminOrg.AdminOrg.Name, fileName, err)
	}

	settings, err := adminOrg.GetFederationSettings()
	if err != nil {
		return diag.Errorf("[Org SAML %s %s] error reading federation settings values: %s", origin, adminOrg.AdminOrg.Name, err)
	}
	settings.SAMLMetadata = string(metadataText)
	settings.Enabled = enabled
	settings.SamlSPEntityID = entityId

	settings.SamlAttributeMapping.EmailAttributeName = d.Get("email").(string)
	settings.SamlAttributeMapping.UserNameAttributeName = d.Get("user_name").(string)
	settings.SamlAttributeMapping.FirstNameAttributeName = d.Get("first_name").(string)
	settings.SamlAttributeMapping.SurnameAttributeName = d.Get("surname").(string)
	settings.SamlAttributeMapping.FullNameAttributeName = d.Get("full_name").(string)
	settings.SamlAttributeMapping.RoleAttributeName = d.Get("role").(string)
	settings.SamlAttributeMapping.GroupAttributeName = d.Get("group").(string)

	_, err = adminOrg.SetFederationSettings(settings)
	if err != nil {
		return diag.Errorf("[Org SAML %s] error setting org '%s' SAML configuration: %s", origin, adminOrg.AdminOrg.Name, err)
	}
	return resourceVcdOrgSamlRead(ctx, d, meta)
}

func resourceVcdOrgSamlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdOrgSamlCreateOrUpdate(ctx, d, meta, "create")
}
func resourceVcdOrgSamlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdOrgSamlRead(ctx, d, meta, "resource")
}

func genericVcdOrgSamlRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("resource vcd_org_saml requires System administrator privileges")
	}
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgId)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find Organization %s SAML settings: %s. Removing from state", orgId, err)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("unable to find organization %s: %s", orgId, err)
	}

	settings, err := adminOrg.GetFederationSettings()
	if err != nil {
		return diag.Errorf("unable to retrieve organization %s SAML settings: %s", adminOrg.AdminOrg.Name, err)
	}
	dSet(d, "enabled", settings.Enabled)
	dSet(d, "entity_id", settings.SamlSPEntityID)

	dSet(d, "email", settings.SamlAttributeMapping.EmailAttributeName)
	dSet(d, "user_name", settings.SamlAttributeMapping.UserNameAttributeName)
	dSet(d, "first_name", settings.SamlAttributeMapping.FirstNameAttributeName)
	dSet(d, "surname", settings.SamlAttributeMapping.SurnameAttributeName)
	dSet(d, "full_name", settings.SamlAttributeMapping.FullNameAttributeName)
	dSet(d, "role", settings.SamlAttributeMapping.RoleAttributeName)
	dSet(d, "group", settings.SamlAttributeMapping.GroupAttributeName)

	d.SetId(adminOrg.AdminOrg.ID)

	return nil
}

func resourceVcdOrgSamlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdOrgSamlCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcdOrgSamlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("resource vcd_org_saml requires System administrator privileges")
	}
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org SAML delete] error searching for Org %s: %s", orgId, err)
	}
	err = adminOrg.UnsetFederationSettings()
	if err != nil {
		return diag.Errorf("[Org SAML delete] error unsetting SAML settings for Org %s: %s", adminOrg.AdminOrg.Name, err)
	}
	return nil
}

// resourceVcdOrgSamlImport is responsible for importing the resource.
// The only parameter needed is the Org identifier, which could be either the Org name or its ID
func resourceVcdOrgSamlImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	orgName := d.Id()

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, err)
	}

	dSet(d, "org_id", adminOrg.AdminOrg.ID)

	d.SetId(adminOrg.AdminOrg.ID)
	return []*schema.ResourceData{d}, nil
}
