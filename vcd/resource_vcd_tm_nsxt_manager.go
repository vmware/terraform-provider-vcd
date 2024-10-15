package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const labelTmNsxtManager = "NSX-T Manager"

func resourceVcdTmNsxtManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmNsxtManagerCreate,
		ReadContext:   resourceVcdTmNsxtManagerRead,
		UpdateContext: resourceVcdTmNsxtManagerUpdate,
		DeleteContext: resourceVcdTmNsxtManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmNsxtManagerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Manager",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of NSX-T Manager",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for authenticating to NSX-T Manager",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password for authenticating to NSX-T Manager ",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL of NSX-T Manager",
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVirtualCenter),
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network Provider Scope for NSX-T Manager",
			},
			"status": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of NSX-T Manager",
			},
		},
	}
}

func resourceVcdTmNsxtManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelTmNsxtManager,
		getTypeFunc:      getTmNsxtManagerType,
		stateStoreFunc:   setTmNsxtManagerData,
		createFunc:       vcdClient.CreateNsxtManagerOpenApi,
		resourceReadFunc: resourceVcdTmNsxtManagerRead,
		preCreateHooks:   []beforeCreateHook{trustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmNsxtManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelTmNsxtManager,
		getTypeFunc:      getTmNsxtManagerType,
		getEntityFunc:    vcdClient.GetNsxtManagerOpenApiById,
		resourceReadFunc: resourceVcdTmNsxtManagerRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelTmNsxtManager,
		getEntityFunc:  vcdClient.GetNsxtManagerOpenApiById,
		stateStoreFunc: setTmNsxtManagerData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmNsxtManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:   labelTmNsxtManager,
		getEntityFunc: vcdClient.GetNsxtManagerOpenApiById,
		// preDeleteHooks: []resourceHook[*govcd.VCenter]{disableBeforeDelete},
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmNsxtManagerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	nsxtManager, err := vcdClient.GetNsxtManagerOpenApiByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Manager '%s': %s", d.Id(), err)
	}
	d.SetId(nsxtManager.NsxtManagerOpenApi.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmNsxtManagerType(d *schema.ResourceData) (*types.NsxtManagerOpenApi, error) {
	t := &types.NsxtManagerOpenApi{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Username:             d.Get("username").(string),
		Password:             d.Get("password").(string),
		Url:                  d.Get("url").(string),
		NetworkProviderScope: d.Get("network_provider_scope").(string),
	}

	return t, nil
}

func setTmNsxtManagerData(d *schema.ResourceData, t *govcd.NsxtManagerOpenApi) error {
	if t == nil || t.NsxtManagerOpenApi == nil {
		return fmt.Errorf("nil object for %s", labelTmNsxtManager)
	}
	n := t.NsxtManagerOpenApi

	d.SetId(n.ID)
	dSet(d, "name", n.Name)
	dSet(d, "description", n.Description)
	dSet(d, "username", n.Username)
	dSet(d, "password", n.Password)
	dSet(d, "url", n.Url)
	dSet(d, "network_provider_scope", n.NetworkProviderScope)
	dSet(d, "status", n.Status)

	return nil
}
