package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"log"
	"strings"
)

func resourceVcdExternalEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdExternalEndpointCreate,
		ReadContext:   resourceVcdExternalEndpointRead,
		UpdateContext: resourceVcdExternalEndpointUpdate,
		DeleteContext: resourceVcdExternalEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdExternalEndpointImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the External Endpoint",
				// This should not be needed once VCD validates correctly that "." are not present in the name
				// during creation. Otherwise, this is critical to avoid having incorrect external endpoints that
				// cannot be read, updated nor deleted.
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringDoesNotContainAny(".")),
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Vendor of the External Endpoint",
				// Same as above
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringDoesNotContainAny(".")),
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Version of the External Endpoint",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the External Endpoint is enabled or not",
			},
			"disable_on_removal": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If 'true', the External Endpoint is disabled before deleting the resource",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the External Endpoint",
			},
			"root_url": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The URL which requests will be redirected to. It must be a valid URL using https protocol",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPS),
			},
		},
	}
}

func resourceVcdExternalEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	_, err := vcdClient.CreateExternalEndpoint(&types.ExternalEndpoint{
		Name:        d.Get("name").(string),
		Version:     d.Get("version").(string),
		Vendor:      d.Get("vendor").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
		RootUrl:     d.Get("root_url").(string),
	})
	if err != nil {
		return diag.Errorf("could not create the External Endpoint: %s", err)
	}
	return resourceVcdExternalEndpointRead(ctx, d, meta)
}

func resourceVcdExternalEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdExternalEndpointRead(ctx, d, meta, "resource")
}

func genericVcdExternalEndpointRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	var ep *govcd.ExternalEndpoint
	var err error
	identifier := d.Id()
	if identifier == "" {
		identifier = fmt.Sprintf("urn:vcloud:extensionEndpoint:%s:%s:%s", d.Get("vendor").(string), d.Get("name").(string), d.Get("version").(string))
		ep, err = vcdClient.GetExternalEndpoint(d.Get("vendor").(string), d.Get("name").(string), d.Get("version").(string))
	} else {
		ep, err = vcdClient.GetExternalEndpointById(d.Id())
	}
	if govcd.ContainsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find External Endpoint '%s': %s. Removing from state", identifier, err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("could not find the External Endpoint '%s': %s", identifier, err)
	}

	dSet(d, "name", ep.ExternalEndpoint.Name)
	dSet(d, "vendor", ep.ExternalEndpoint.Vendor)
	dSet(d, "version", ep.ExternalEndpoint.Version)
	dSet(d, "enabled", ep.ExternalEndpoint.Enabled)
	dSet(d, "description", ep.ExternalEndpoint.Description)
	dSet(d, "root_url", ep.ExternalEndpoint.RootUrl)
	d.SetId(ep.ExternalEndpoint.ID)
	return nil
}

func resourceVcdExternalEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !d.HasChangesExcept("disable_on_removal") {
		// This is a utility toggle that does not interact with VCD
		return nil
	}

	vcdClient := meta.(*VCDClient)
	ep, err := vcdClient.GetExternalEndpointById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve the External Endpoint for update: %s", err)
	}
	err = ep.Update(types.ExternalEndpoint{
		Name:        d.Get("name").(string),
		Version:     d.Get("version").(string),
		Vendor:      d.Get("vendor").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
		RootUrl:     d.Get("root_url").(string),
	})
	if err != nil {
		return diag.Errorf("could not update the External Endpoint: %s", err)
	}
	return resourceVcdExternalEndpointRead(ctx, d, meta)
}

func resourceVcdExternalEndpointDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	ep, err := vcdClient.GetExternalEndpointById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return nil // Already deleted
		}
		return diag.Errorf("could not retrieve the External Endpoint for delete: %s", err)
	}
	if d.Get("disable_on_removal").(bool) {
		err = ep.Update(types.ExternalEndpoint{
			Name:        ep.ExternalEndpoint.Name,
			ID:          ep.ExternalEndpoint.ID,
			Version:     ep.ExternalEndpoint.Version,
			Vendor:      ep.ExternalEndpoint.Vendor,
			Enabled:     false,
			Description: ep.ExternalEndpoint.Description,
			RootUrl:     ep.ExternalEndpoint.RootUrl,
		})
		if err != nil {
			return diag.Errorf("could not disable the External Endpoint '%s' for deletion: %s", d.Id(), err)
		}
	}
	err = ep.Delete()
	if err != nil {
		return diag.Errorf("could not delete the External Endpoint '%s': %s", d.Id(), err)
	}
	return nil
}

// resourceVcdExternalEndpointImport is responsible for importing the resource.
// The ID must be vendorVCD_IMPORT_SEPARATORnameVCD_IMPORT_SEPARATORversion
func resourceVcdExternalEndpointImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	var externalEndpoint *govcd.ExternalEndpoint
	var err error
	switch len(resourceURI) {
	case 3: // ie: VCD_IMPORT_SEPARATOR="_" vendor_name_1.2.3
		externalEndpoint, err = vcdClient.GetExternalEndpoint(resourceURI[0], resourceURI[1], resourceURI[2])
		if err != nil {
			return nil, err
		}
	case 5: // ie: VCD_IMPORT_SEPARATOR="."  vendor.name.1.2.3
		externalEndpoint, err = vcdClient.GetExternalEndpoint(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("to import the External Endpoint, the ID must be 'vendor%sname%sversion', but it was '%s'", ImportSeparator, ImportSeparator, d.Id())
	}

	d.SetId(externalEndpoint.ExternalEndpoint.ID)
	return []*schema.ResourceData{d}, nil
}
