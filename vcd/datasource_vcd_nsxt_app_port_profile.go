package vcd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
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
			"context_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC, VDC Group, or NSX-T Manager",
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

func datasourceVcdNsxtAppPortProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)

	queryParams := url.Values{}
	// For `TENANT` scope Org and VDC or the specified `context_id` matter. It would set _context
	// filter to be searching for App Port Profiles in specific
	if strings.EqualFold(scope, types.ApplicationPortProfileScopeTenant) {
		contextIdField := d.Get("context_id").(string)
		dataSourceVdcField := d.Get("vdc").(string)
		inheritedVdcField := vcdClient.Vdc
		contextId, err := pickAppPortProfileContextFilterByPriority(vcdClient, d, inheritedVdcField, dataSourceVdcField, contextIdField)
		if err != nil {
			return diag.Errorf("error identifying correct context filter: %s", err)
		}
		queryParams.Add("filter", fmt.Sprintf("name==%s;scope==%s;_context==%s", name, scope, contextId))
	} else {
		queryParams.Add("filter", fmt.Sprintf("name==%s;scope==%s", name, scope))
	}
	allAppPortProfiles, err := org.GetAllNsxtAppPortProfiles(queryParams, scope)

	if err != nil {
		return diag.Errorf("error retrieving Application Port Profiles: %s", err)
	}

	if len(allAppPortProfiles) == 0 {
		return diag.Errorf("%s NSX-T Application Port Profile not found", govcd.ErrorEntityNotFound)
	}

	if len(allAppPortProfiles) > 1 {
		return diag.Errorf("Expected exactly one NSX-T Application Port Profile. Got '%d'", len(allAppPortProfiles))
	}
	appPortProfile := allAppPortProfiles[0]

	err = setNsxtAppPortProfileData(d, appPortProfile.NsxtAppPortProfile)
	if err != nil {
		return diag.Errorf("error storing NSX-T Application Port Profile schema: %s", err)
	}

	d.SetId(appPortProfile.NsxtAppPortProfile.ID)

	return nil
}

// pickAppPortProfileContextFilterByPriority will evaluate 3 fields - 'context_id', 'vdc' in
// resource and 'vdc' in provider section. It will pick the right one based on priority:
// * Priority 1 -> 'context_id' field
// * Priority 2 -> 'vdc' field in data source
// * Priority 3 -> 'vdc' field inherited from provider configuration
func pickAppPortProfileContextFilterByPriority(vcdClient *VCDClient, d *schema.ResourceData, inheritedVdcField, dataSourceVdcField, contextIdField string) (string, error) {
	// Context ID can be returned directly, VDC must be looked up to return its ID
	if contextIdField != "" {
		return contextIdField, nil
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return "", fmt.Errorf("error retrieving Org and VDC: %s", err)
	}

	return vdc.Vdc.ID, nil
}
