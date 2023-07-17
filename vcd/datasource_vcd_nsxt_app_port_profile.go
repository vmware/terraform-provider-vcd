package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				Deprecated:    "Deprecated in favor of 'context_id'",
				ConflictsWith: []string{"context_id"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Application Port Profile name",
			},
			// TODO V4 Change this to required
			"context_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC, VDC Group, or NSX-T Manager. Required if the VCD instance has more than one NSX-T manager",
				ConflictsWith: []string{"nsxt_manager_id", "vdc"},
			},
			"scope": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Scope - 'SYSTEM', 'PROVIDER' or 'TENANT'",
				ValidateFunc: validation.StringInSlice([]string{"SYSTEM", "PROVIDER", "TENANT"}, false),
			},
			"nsxt_manager_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "ID of NSX-T manager. Only required for 'PROVIDER' scope",
				Deprecated:    "Deprecated in favor of 'context_id'",
				ConflictsWith: []string{"context_id"},
			},
			"description": {
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

func datasourceVcdNsxtAppPortProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcName := d.Get("vdc").(string)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)
	contextIdFieldValue := d.Get("context_id").(string)
	nsxtManagerId := d.Get("nsxt_manager_id").(string)

	queryParams := url.Values{}

	var contextId string
	switch {
	// context_id is the preferred method of providing context
	case contextIdFieldValue != "":
		contextId = contextIdFieldValue
	// if vdc attribute is set, use that, as it is also available for every user
	case vdcName != "":
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf("failed to get VDC from resource: %s", err)
		}
		contextId = vdc.Vdc.ID
	// nsxt_manager_id can only be used by sysorg admin
	case nsxtManagerId != "":
		if !vcdClient.Client.IsSysAdmin {
			return diag.Errorf("Only System administrators can provide NSX-T Manager ID")
		}
		contextId = nsxtManagerId
	// if none of previous values are set, don't provide context (usable if only one nsxt manager is used)
	default:
		contextId = ""
	}

	// If contextId is unset, send a request without _context query filter,
	// Works properly only with SYSTEM scope and one NSX-T Manager configured.
	if contextId != "" {
		queryParams.Add("filter", fmt.Sprintf("name==%s;scope==%s;_context==%s", name, scope, contextId))
	} else {
		queryParams.Add("filter", fmt.Sprintf("name==%s;scope==%s", name, scope))
	}

	allAppPortProfiles, err := org.GetAllNsxtAppPortProfiles(queryParams, scope)

	if err != nil {
		return diag.Errorf("error retrieving NSX-T Application Port Profiles: %s", err)
	}

	if len(allAppPortProfiles) == 0 {
		return diag.Errorf("%s NSX-T Application Port Profile not found", govcd.ErrorEntityNotFound)
	}

	if len(allAppPortProfiles) > 1 {
		return diag.Errorf("Expected exactly one NSX-T Application Port Profile. Got '%d'", len(allAppPortProfiles))
	}
	appPortProfile := allAppPortProfiles[0]

	err = setNsxtAppPortProfileData(vcdClient, d, appPortProfile.NsxtAppPortProfile)
	if err != nil {
		return diag.Errorf("error storing NSX-T Application Port Profile schema: %s", err)
	}

	d.SetId(appPortProfile.NsxtAppPortProfile.ID)

	return nil
}
