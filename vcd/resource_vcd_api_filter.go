package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
)

func resourceVcdApiFilter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdApiFilterCreate,
		ReadContext:   resourceVcdApiFilterRead,
		UpdateContext: resourceVcdApiFilterUpdate,
		DeleteContext: resourceVcdApiFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdApiFilterImport,
		},
		Schema: map[string]*schema.Schema{
			"external_endpoint_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the External Endpoint where this API Filter will process the requests to",
			},
			"url_matcher_pattern": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Request URL pattern, written as a regular expression pattern",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"url_matcher_scope": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Allowed values are EXT_API, EXT_UI_PROVIDER, EXT_UI_TENANT corresponding to /ext-api, /ext-ui/provider, /ext-ui/tenant/<tenant-name>",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"EXT_API", "EXT_UI_PROVIDER", "EXT_UI_TENANT"}, false)),
			},
		},
	}
}

func resourceVcdApiFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	epId := d.Get("external_endpoint_id").(string)
	ep, err := vcdClient.GetExternalEndpointById(epId)
	if err != nil {
		return diag.Errorf("could not retrieve the External Endpoint '%s' to create an API Filter: %s", epId, err)
	}

	createdAf, err := vcdClient.CreateApiFilter(&types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: ep.ExternalEndpoint.Name,
			ID:   ep.ExternalEndpoint.ID,
		},
		UrlMatcher: &types.UrlMatcher{
			UrlPattern: d.Get("url_matcher_pattern").(string),
			UrlScope:   d.Get("url_matcher_scope").(string),
		},
	})
	if err != nil {
		return diag.Errorf("could not create the API Filter: %s", err)
	}
	d.SetId(createdAf.ApiFilter.ID)
	return resourceVcdApiFilterRead(ctx, d, meta)
}

func resourceVcdApiFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdApiFilterRead(ctx, d, meta, "resource")
}

func genericVcdApiFilterRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// ID must be populated during Create, as API filters don't have any other identifier
	af, err := vcdClient.GetApiFilterById(d.Id())
	if govcd.ContainsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find API Filter '%s': %s. Removing from state", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("could not find the API Filter '%s': %s", d.Id(), err)
	}

	if af.ApiFilter.ExternalSystem != nil {
		dSet(d, "external_endpoint_id", af.ApiFilter.ExternalSystem.ID)
	}
	if af.ApiFilter.UrlMatcher != nil {
		dSet(d, "url_matcher_pattern", af.ApiFilter.UrlMatcher.UrlPattern)
		dSet(d, "url_matcher_scope", af.ApiFilter.UrlMatcher.UrlScope)
	}
	d.SetId(af.ApiFilter.ID)
	return nil
}

func resourceVcdApiFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	af, err := vcdClient.GetApiFilterById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve the API Filter '%s' to update it: %s", d.Id(), err)
	}

	epId := d.Get("external_endpoint_id").(string)
	ep, err := vcdClient.GetExternalEndpointById(epId)
	if err != nil {
		return diag.Errorf("could not retrieve the External Endpoint '%s' to update the API Filter '%s': %s", epId, af.ApiFilter.ID, err)
	}

	err = af.Update(types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: ep.ExternalEndpoint.Name,
			ID:   ep.ExternalEndpoint.ID,
		},
		UrlMatcher: &types.UrlMatcher{
			UrlPattern: d.Get("url_matcher_pattern").(string),
			UrlScope:   d.Get("url_matcher_scope").(string),
		},
	})
	if err != nil {
		return diag.Errorf("could not update the API Filter: %s", err)
	}
	return resourceVcdApiFilterRead(ctx, d, meta)
}

func resourceVcdApiFilterDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	af, err := vcdClient.GetApiFilterById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve the API Filter for deletion: %s", err)
	}
	err = af.Delete()
	if err != nil {
		return diag.Errorf("could not delete the API Filter: %s", err)
	}
	return nil
}

// resourceVcdApiFilterImport is responsible for importing the resource.
func resourceVcdApiFilterImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	af, err := vcdClient.GetApiFilterById(d.Id())
	if err != nil {
		return nil, fmt.Errorf("could not find the API Filter with ID '%s': %s", d.Id(), err)
	}
	d.SetId(af.ApiFilter.ID)
	return []*schema.ResourceData{d}, nil
}
