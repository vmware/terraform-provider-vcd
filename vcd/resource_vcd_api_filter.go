package vcd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"net/url"
	"strings"
	"text/tabwriter"
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
	helpError := fmt.Errorf(`resource id must be specified in one of these formats:
'api-filter-id' to import by API Filter ID
'list@vendor%sname%sversion' to get a list of API Filters related to the External Endpoint identified by vendor, name and version`, ImportSeparator, ImportSeparator)

	id := strings.Split(d.Id(), "@")
	switch len(id) {
	case 1:
		af, err := vcdClient.GetApiFilterById(d.Id())
		if err != nil {
			return nil, fmt.Errorf("could not find the API Filter with ID '%s': %s", d.Id(), err)
		}
		d.SetId(af.ApiFilter.ID)
		return []*schema.ResourceData{d}, nil
	case 2:
		externalEndpointId := strings.Split(id[1], ImportSeparator)
		var externalEndpoint *govcd.ExternalEndpoint
		var err error
		switch len(externalEndpointId) {
		case 3: // ie: VCD_IMPORT_SEPARATOR="_" vendor_name_1.2.3
			externalEndpoint, err = vcdClient.GetExternalEndpoint(externalEndpointId[0], externalEndpointId[1], externalEndpointId[2])
			if err != nil {
				return nil, err
			}
		case 5: // ie: VCD_IMPORT_SEPARATOR="."  vendor.name.1.2.3
			externalEndpoint, err = vcdClient.GetExternalEndpoint(externalEndpointId[0], externalEndpointId[1], fmt.Sprintf("%s.%s.%s", externalEndpointId[2], externalEndpointId[3], externalEndpointId[4]))
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("to list the API Filters, the External Endpoint ID must be 'vendor%sname%sversion', but it was '%s'", ImportSeparator, ImportSeparator, externalEndpointId)
		}
		queryParameters := url.Values{}
		queryParameters.Add("filter", fmt.Sprintf("externalSystem.id==%s", externalEndpoint.ExternalEndpoint.ID))
		apiFilters, err := vcdClient.GetAllApiFilters(queryParameters)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve all API Filters to list them: %s", err)
		}
		buf := new(bytes.Buffer)
		_, err = fmt.Fprintln(buf, "Retrieving all API Filters that use "+externalEndpoint.ExternalEndpoint.ID+" as External Endpoint")
		if err != nil {
			return nil, fmt.Errorf("could not list API Filters: %s", err)
		}
		writer := tabwriter.NewWriter(buf, 0, 8, 1, '\t', tabwriter.AlignRight)
		_, err = fmt.Fprintf(writer, "No\tID\tScope\tPattern\n")
		if err != nil {
			return nil, fmt.Errorf("could not list API Filters: %s", err)
		}
		_, err = fmt.Fprintf(writer, "--\t--\t-----\t-------\n")
		if err != nil {
			return nil, fmt.Errorf("could not list API Filters: %s", err)
		}
		for i, af := range apiFilters {
			if af.ApiFilter.UrlMatcher == nil {
				continue
			}
			_, err = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", i+1, af.ApiFilter.ID, af.ApiFilter.UrlMatcher.UrlScope, af.ApiFilter.UrlMatcher.UrlPattern)
			if err != nil {
				return nil, fmt.Errorf("could not list API Filters: %s", err)
			}
		}
		err = writer.Flush()
		if err != nil {
			return nil, fmt.Errorf("could not list API Filters: %s", err)
		}
		return nil, fmt.Errorf("resource was not imported! %s\n%s", helpError, buf.String())
	default:
		return nil, helpError
	}

}
