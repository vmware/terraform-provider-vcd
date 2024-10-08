package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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
			"network_provider_scope": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network Provider Scope for NSX-T Manager",
			},
			"status": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Status of NSX-T Manager",
			},
		},
	}
}

func resourceVcdTmNsxtManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getTmNsxtManagerType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T Manager type: %s")
	}

	nsxtManager, err := vcdClient.CreateTmNsxtManager(t)
	if err != nil {
		return diag.Errorf("error creating NSX-T Manager: %s")
	}

	d.SetId(nsxtManager.TmNsxtManager.ID)
	return resourceVcdTmNsxtManagerRead(ctx, d, meta)
}

func resourceVcdTmNsxtManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getTmNsxtManagerType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T Manager type: %s")
	}

	nsxtManager, err := vcdClient.GetTmNsxtManagerById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Manager: %s")
	}

	_, err = nsxtManager.Update(t)
	if err != nil {
		return diag.Errorf("error updating NSX-T Manager: %s", err)
	}

	return resourceVcdTmNsxtManagerRead(ctx, d, meta)
}

func resourceVcdTmNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	nsxtManager, err := vcdClient.GetTmNsxtManagerById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Manager: %s")
	}

	err = setTmNsxtManagerData(d, nsxtManager.TmNsxtManager)
	if err != nil {
		return diag.Errorf("error storing NSX-T Manager to state: %s", err)
	}

	return nil
}

func resourceVcdTmNsxtManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	nsxtManager, err := vcdClient.GetTmNsxtManagerById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Manager: %s", err)
	}

	err = nsxtManager.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Manager: %s", err)
	}

	return nil
}

func resourceVcdTmNsxtManagerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	nsxtManager, err := vcdClient.GetTmNsxtManagerByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Manager '%s': %s", d.Id(), err)
	}
	d.SetId(nsxtManager.TmNsxtManager.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmNsxtManagerType(d *schema.ResourceData) (*types.TmNsxtManager, error) {
	t := &types.TmNsxtManager{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Username:             d.Get("username").(string),
		Password:             d.Get("password").(string),
		URL:                  d.Get("url").(string),
		NetworkProviderScope: d.Get("network_provider_scope").(string),
	}

	return t, nil
}

func setTmNsxtManagerData(d *schema.ResourceData, t *types.TmNsxtManager) error {
	dSet(d, "name", t.Name)
	dSet(d, "description", t.Description)
	dSet(d, "username", t.Username)
	dSet(d, "password", t.Password)
	dSet(d, "url", t.URL)
	dSet(d, "network_provider_scope", t.NetworkProviderScope)
	dSet(d, "status", t.Status)

	return nil
}
