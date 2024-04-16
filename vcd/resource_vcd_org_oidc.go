package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
