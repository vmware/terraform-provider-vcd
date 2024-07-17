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
				ForceNew:    true, // TODO: Check
				Description: "ID of the External Endpoint where this API Filter will process the requests to",
			},
			"url_matcher_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "TODO",
				RequiredWith: []string{"url_matcher_scope"},
				ExactlyOneOf: []string{"url_matcher_pattern, response_content_type"},
			},
			"url_matcher_scope": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Allowed values are EXT_API, EXT_UI_PROVIDER, EXT_UI_TENANT corresponding to /ext-api, /ext-ui/provider, /ext-ui/tenant/<tenant-name>",
				RequiredWith:     []string{"url_matcher_pattern"},
				ExactlyOneOf:     []string{"url_matcher_scope, response_content_type"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"EXT_API", "EXT_UI_PROVIDER", "EXT_UI_TENANT"}, false)),
			},
			"response_content_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "TODO",
				ExactlyOneOf: []string{"url_matcher_pattern, response_content_type"},
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

	af := &types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: ep.ExternalEndpoint.Name,
			ID:   ep.ExternalEndpoint.ID,
		},
	}

	if p, ok := d.GetOk("url_matcher_pattern"); ok {
		af.UrlMatcher = &types.UrlMatcher{
			UrlPattern: p.(string),
			UrlScope:   d.Get("url_matcher_scope").(string), // Guaranteed by schema constraints
		}
	}
	if r, ok := d.GetOk("response_content_type"); ok {
		af.ResponseContentType = addrOf(r.(string))
	}

	createdAf, err := vcdClient.CreateApiFilter(af)
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
	if af.ApiFilter.ResponseContentType != nil {
		dSet(d, "response_content_type", af.ApiFilter.ResponseContentType)
	}
	d.SetId(af.ApiFilter.ID)
	return nil
}

func resourceVcdApiFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// TODO: Can it be updated?
	af, err := vcdClient.GetApiFilterById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve the API Filter '%s' to update it: %s", d.Id(), err)
	}

	epId := d.Get("external_endpoint_id").(string)
	ep, err := vcdClient.GetExternalEndpointById(epId)
	if err != nil {
		return diag.Errorf("could not retrieve the External Endpoint '%s' to create an API Filter: %s", epId, err)
	}

	updatePayload := types.ApiFilter{
		ExternalSystem: &types.OpenApiReference{
			Name: ep.ExternalEndpoint.Name,
			ID:   ep.ExternalEndpoint.ID,
		},
	}

	if p, ok := d.GetOk("url_matcher_pattern"); ok {
		updatePayload.UrlMatcher = &types.UrlMatcher{
			UrlPattern: p.(string),
			UrlScope:   d.Get("url_matcher_scope").(string), // Guaranteed by schema constraints
		}
	}
	if r, ok := d.GetOk("response_content_type"); ok {
		updatePayload.ResponseContentType = addrOf(r.(string))
	}

	err = af.Update(updatePayload)
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
