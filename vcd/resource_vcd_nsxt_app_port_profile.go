package vcd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var appPortDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"protocol": {
			Required:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"ICMPv4", "ICMPv6", "TCP", "UDP"}, false),
		},
		"port": {
			Optional:    true,
			Type:        schema.TypeSet,
			Description: "Set of ports or ranges",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func resourceVcdNsxtAppPortProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtAppPortProfileCreate,
		ReadContext:   resourceVcdNsxtAppPortProfileRead,
		UpdateContext: resourceVcdNsxtAppPortProfileUpdate,
		DeleteContext: resourceVcdNsxtAppPortProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtAppPortProfileImport,
		},

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
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application Port Profile description",
			},
			"scope": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Scope - 'PROVIDER' or 'TENANT'",
				ValidateFunc: validation.StringInSlice([]string{"PROVIDER", "TENANT"}, false),
			},
			"nsxt_manager_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of NSX-T manager. Only required for 'PROVIDER' scope",
			},
			"app_port": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     appPortDefinition,
			},
		},
	}
}

func resourceVcdNsxtAppPortProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	scope := d.Get("scope").(string)
	err := validateScope(scope, d.Get("nsxt_manager_id").(string), d.Get("org").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	appPortProfile, err := getNsxtAppPortProfileType(d, org, vcdClient)
	if err != nil {
		return diag.Errorf("error getting NSX-T Application Port Profile configuration: %s", err)
	}

	createdAppPortProfile, err := org.CreateNsxtAppPortProfile(appPortProfile)
	if err != nil {
		return diag.Errorf("error creating NSX-T Application Port Profile '%s': %s", appPortProfile.Name, err)
	}

	d.SetId(createdAppPortProfile.NsxtAppPortProfile.ID)

	return resourceVcdNsxtAppPortProfileRead(ctx, d, meta)
}

func resourceVcdNsxtAppPortProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	scope := d.Get("scope").(string)
	err := validateScope(scope, d.Get("nsxt_manager_id").(string), d.Get("org").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	appPortProfile, err := org.GetNsxtAppPortProfileById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Application Port Profile: %s", err)
	}

	updateappPortProfile, err := getNsxtAppPortProfileType(d, org, vcdClient)
	if err != nil {
		return diag.Errorf("error getting NSX-T Application Port Profile configuration: %s", err)
	}
	// Inject existing ID for update
	updateappPortProfile.ID = d.Id()

	_, err = appPortProfile.Update(updateappPortProfile)
	if err != nil {
		return diag.Errorf("error updating NSX-T Application Port Profile '%s': %s", updateappPortProfile.Name, err)
	}

	return resourceVcdNsxtAppPortProfileRead(ctx, d, meta)
}

func resourceVcdNsxtAppPortProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	appPortProfile, err := org.GetNsxtAppPortProfileById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting NSX-T Application Port Profile with ID '%s': %s", d.Id(), err)
	}

	err = setNsxtAppPortProfileData(d, appPortProfile.NsxtAppPortProfile)
	if err != nil {
		return diag.Errorf("error reading NSX-T Application Port Profile: %s", err)
	}

	return nil
}

func resourceVcdNsxtAppPortProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	appPortProfile, err := org.GetNsxtAppPortProfileById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Application Port Profile: %s", err)
	}

	err = appPortProfile.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Application Port Profile: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtAppPortProfileImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	// There are two paths of possible import of differently scoped NSX-T Application Port Profiles
	// * PROVIDER (path contains 2 pieces nsxt_manager_name.app_port_profile_name)
	// * TENANT (path contains 3 pieces org-name.vdc-name.app_port_profile_name)

	vcdClient := meta.(*VCDClient)

	var nsxtAppPortProfile *govcd.NsxtAppPortProfile

	switch len(resourceURI) {
	case 2: // PROVIDER scope
		if !vcdClient.Client.IsSysAdmin {
			return nil, errors.New("only System user can modify PROVIDER scope NSX-T Application Port " +
				"Profiles. Please use data source instead")
		}

		org, err := vcdClient.GetOrgByName("System")
		if err != nil {
			return nil, fmt.Errorf("error retrieving System Org: %s", err)
		}

		nsxtManagerName, appPortProfileName := resourceURI[0], resourceURI[1]
		nsxtManagers, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
		if err != nil {
			return nil, fmt.Errorf("could not find NSX-T manager by name '%s': %s", nsxtManagerName, err)
		}
		if len(nsxtManagers) == 0 {
			return nil, fmt.Errorf("%s found %d NSX-T managers with name '%s'",
				govcd.ErrorEntityNotFound, len(nsxtManagers), nsxtManagerName)
		}

		if len(nsxtManagers) > 1 {
			return nil, fmt.Errorf("found %d NSX-T managers with name '%s'", len(nsxtManagers), nsxtManagerName)
		}

		id := extractUuid(nsxtManagers[0].HREF)
		nsxtManagerUrn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", id)
		if err != nil {
			return nil, fmt.Errorf("could not construct URN from id '%s': %s", id, err)
		}

		// Inject NSX-T Manager ID into _context filter
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("_context==%s;name==%s", nsxtManagerUrn, appPortProfileName))
		allNsxtAppPortProfiles, err := org.GetAllNsxtAppPortProfiles(queryParams, types.ApplicationPortProfileScopeProvider)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Application Port Profile with name %s (NSX-T Manager Name '%s'): %s",
				appPortProfileName, nsxtManagerName, err)
		}
		nsxtAppPortProfile = allNsxtAppPortProfiles[0]

		dSet(d, "org", org.Org.Name)
		dSet(d, "nsxt_manager_id", nsxtManagerUrn)

	case 3: // TENANT scope
		orgName, vdcName, appPortProfileName := resourceURI[0], resourceURI[1], resourceURI[2]

		org, err := vcdClient.GetOrgByName(orgName)
		if err != nil {
			return nil, fmt.Errorf("unable to find Org %s: %s", orgName, err)
		}
		vdc, err := org.GetVDCByName(vdcName, false)
		if err != nil {
			return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
		}

		if !vdc.IsNsxt() {
			return nil, errors.New("security groups are only supported by NSX-T VDCs")
		}

		nsxtAppPortProfile, err = org.GetNsxtAppPortProfileByName(appPortProfileName, types.ApplicationPortProfileScopeTenant)
		if err != nil {
			return nil, fmt.Errorf("unable to find Application Port Profile '%s': %s", appPortProfileName, err)
		}

		dSet(d, "org", orgName)
		dSet(d, "vdc", vdcName)

	default:
		return nil, fmt.Errorf("resource path must be specified in one of two formats, based on Application Port Profile scope:\n" +
			"* PROVIDER (path contains 2 pieces nsxt_manager_name.app_port_profile_name)\n" +
			"* TENANT (path contains 3 pieces org-name.vdc-name.app_port_profile_name)")
	}

	d.SetId(nsxtAppPortProfile.NsxtAppPortProfile.ID)

	return []*schema.ResourceData{d}, nil
}

func validateScope(scope, nsxtManagerId, orgName string) error {
	if scope == types.ApplicationPortProfileScopeProvider && nsxtManagerId == "" {
		return fmt.Errorf("scope 'PROVIDER' requires NSX-T Manager ID")
	}

	if scope == types.ApplicationPortProfileScopeProvider && strings.ToUpper(orgName) != "SYSTEM" {
		return fmt.Errorf("scope 'PROVIDER' requires Org to be \"System\"")
	}

	return nil
}

func getNsxtAppPortProfileType(d *schema.ResourceData, org *govcd.Org, vcdClient *VCDClient) (*types.NsxtAppPortProfile, error) {
	appPortProfileConfig := &types.NsxtAppPortProfile{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Scope:       d.Get("scope").(string),
	}

	switch strings.ToUpper(appPortProfileConfig.Scope) {
	case types.ApplicationPortProfileScopeProvider:
		nsxtManagerUrn := d.Get("nsxt_manager_id").(string)
		appPortProfileConfig.ContextEntityId = nsxtManagerUrn
	case types.ApplicationPortProfileScopeTenant:
		appPortProfileConfig.OrgRef = &types.OpenApiReference{ID: org.Org.ID}
		// Tenant scope requires VDC
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
		}
		appPortProfileConfig.ContextEntityId = vdc.Vdc.ID
	}

	appPortSet := d.Get("app_port").(*schema.Set)
	if appPortSet != nil {
		appPortSlice := appPortSet.List()
		applicationPorts := make([]types.NsxtAppPortProfilePort, len(appPortSlice))
		for index, singlePort := range appPortSlice {
			appPortMap := singlePort.(map[string]interface{})
			onePortDef := types.NsxtAppPortProfilePort{
				Protocol:         appPortMap["protocol"].(string),
				DestinationPorts: convertSchemaSetToSliceOfStrings(appPortMap["port"].(*schema.Set)),
			}
			applicationPorts[index] = onePortDef
		}
		appPortProfileConfig.ApplicationPorts = applicationPorts
	}

	return appPortProfileConfig, nil
}

// setNsxtAppPortProfileData sets Terraform schema from types.NsxtAppPortProfile
//
// Note. GET queries do not return nsxt_manager_ir for SYSTEM scope therefore it cannot be read.
func setNsxtAppPortProfileData(d *schema.ResourceData, appPortProfile *types.NsxtAppPortProfile) error {
	dSet(d, "name", appPortProfile.Name)
	dSet(d, "description", appPortProfile.Description)
	dSet(d, "scope", appPortProfile.Scope)

	if appPortProfile.ApplicationPorts != nil && len(appPortProfile.ApplicationPorts) > 0 {

		resultSet := make([]interface{}, len(appPortProfile.ApplicationPorts))

		for index, value := range appPortProfile.ApplicationPorts {
			appPortMap := make(map[string]interface{})
			appPortMap["protocol"] = value.Protocol

			desitnationPortSet := convertStringsTotTypeSet(value.DestinationPorts)
			appPortMap["port"] = desitnationPortSet

			resultSet[index] = appPortMap

		}

		appPortSet := schema.NewSet(schema.HashResource(appPortDefinition), resultSet)
		err := d.Set("app_port", appPortSet)
		if err != nil {
			return fmt.Errorf("error setting Application Port Profile: %s", err)
		}
	}

	return nil
}
