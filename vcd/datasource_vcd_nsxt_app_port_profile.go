package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var appPortDefinitionComputed = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"protocol": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"port": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Set of ports or ranges",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func datasourceVcdNsxtAppPortProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtAppPortProfileRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Application Port Profile name",
			},
			"scope": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Scope - 'SYSTEM', 'PROVIDER', or 'TENANT'",
				ValidateFunc: validation.StringInSlice([]string{"SYSTEM", "PROVIDER", "TENANT"}, false),
			},
			"nsxt_manager_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of NSX-T manager. Only required for 'PROVIDER' scope",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application Port Profile description",
			},
			"app_port": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     appPortDefinitionComputed,
			},
		},
	}
}

func datasourceVcdNsxtAppPortProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)

	appPortProfile, err := org.GetNsxtAppPortProfileByName(name, scope)
	if err != nil {
		return diag.Errorf("error getting NSX-T Application Port Profile with Name '%s' (scope '%s'): %s",
			name, scope, err)
	}

	err = setNsxtAppPortProfileData(d, appPortProfile.NsxtAppPortProfile)
	if err != nil {
		return diag.Errorf("error reading NSX-T Application Port Profile: %s", err)
	}

	d.SetId(appPortProfile.NsxtAppPortProfile.ID)

	return nil
}
