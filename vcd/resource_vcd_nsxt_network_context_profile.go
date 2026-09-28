package vcd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var networkContextProfileSubAttributeDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "Sub-attribute type (e.g. 'TLS_VERSION', 'TLS_CIPHER_SUITE', 'CIFS_SMB_VERSION')",
		},
		"values": {
			Required:    true,
			MinItems:    1,
			Type:        schema.TypeSet,
			Description: "Set of sub-attribute values",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

var networkContextProfileAttributeDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": {
			Required:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"APP_ID", "DOMAIN_NAME"}, false),
			Description:  "Attribute type - 'APP_ID' or 'DOMAIN_NAME'",
		},
		"values": {
			Required:    true,
			MinItems:    1,
			Type:        schema.TypeSet,
			Description: "Set of attribute values (App IDs or FQDNs known to the backing NSX-T environment)",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"sub_attribute": {
			Optional:    true,
			Type:        schema.TypeSet,
			Description: "Sub-attributes for a single value 'APP_ID' attribute (e.g. SSL or CIFS)",
			Elem:        networkContextProfileSubAttributeDefinition,
		},
	},
}

func resourceVcdNsxtNetworkContextProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtNetworkContextProfileCreate,
		ReadContext:   resourceVcdNsxtNetworkContextProfileRead,
		UpdateContext: resourceVcdNsxtNetworkContextProfileUpdate,
		DeleteContext: resourceVcdNsxtNetworkContextProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtNetworkContextProfileImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of VDC, VDC Group, or NSX-T Manager",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network Context Profile name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network Context Profile description",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      types.ApplicationPortProfileScopeTenant,
				ForceNew:     true,
				Description:  "Scope - 'TENANT' (default) or 'PROVIDER'",
				ValidateFunc: validation.StringInSlice([]string{types.ApplicationPortProfileScopeProvider, types.ApplicationPortProfileScopeTenant}, false),
			},
			"attribute": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "Attributes of the profile. At most one attribute per type",
				Elem:        networkContextProfileAttributeDefinition,
			},
		},
	}
}

func resourceVcdNsxtNetworkContextProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	err := validateNetworkContextProfileScope(d)
	if err != nil {
		return diag.FromErr(err)
	}

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	profileConfig, err := getNsxtNetworkContextProfileType(d, org)
	if err != nil {
		return diag.Errorf("error getting NSX-T Network Context Profile configuration: %s", err)
	}

	createdProfile, err := vcdClient.CreateNetworkContextProfile(profileConfig)
	if err != nil {
		return diag.Errorf("error creating NSX-T Network Context Profile '%s': %s", profileConfig.Name, err)
	}

	d.SetId(createdProfile.NsxtNetworkContextProfile.ID)

	return resourceVcdNsxtNetworkContextProfileRead(ctx, d, meta)
}

func resourceVcdNsxtNetworkContextProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	err := validateNetworkContextProfileScope(d)
	if err != nil {
		return diag.FromErr(err)
	}

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	profile, err := vcdClient.GetNetworkContextProfileById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Network Context Profile: %s", err)
	}

	updateProfileConfig, err := getNsxtNetworkContextProfileType(d, org)
	if err != nil {
		return diag.Errorf("error getting NSX-T Network Context Profile configuration: %s", err)
	}
	// Inject existing ID for update
	updateProfileConfig.ID = d.Id()

	_, err = profile.Update(updateProfileConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T Network Context Profile '%s': %s", updateProfileConfig.Name, err)
	}

	return resourceVcdNsxtNetworkContextProfileRead(ctx, d, meta)
}

func resourceVcdNsxtNetworkContextProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	profile, err := vcdClient.GetNetworkContextProfileById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting NSX-T Network Context Profile with ID '%s': %s", d.Id(), err)
	}

	err = setNsxtNetworkContextProfileData(d, profile.NsxtNetworkContextProfile)
	if err != nil {
		return diag.Errorf("error reading NSX-T Network Context Profile: %s", err)
	}

	return nil
}

func resourceVcdNsxtNetworkContextProfileDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	profile, err := vcdClient.GetNetworkContextProfileById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Network Context Profile: %s", err)
	}

	err = profile.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Network Context Profile: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtNetworkContextProfileImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	// There are two paths of possible import of differently scoped NSX-T Network Context Profiles
	// * PROVIDER (path contains 2 pieces nsxt_manager_name.profile_name)
	// * TENANT (path contains 3 pieces org-name.vdc-or-vdc-group-name.profile_name)

	vcdClient := meta.(*VCDClient)

	var profile *types.NsxtNetworkContextProfile

	switch len(resourceURI) {
	case 2: // PROVIDER scope
		if !vcdClient.Client.IsSysAdmin {
			return nil, errors.New("only System user can modify PROVIDER scope NSX-T Network Context " +
				"Profiles. Please use data source instead")
		}

		nsxtManagerName, profileName := resourceURI[0], resourceURI[1]
		nsxtManagers, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
		if err != nil {
			return nil, fmt.Errorf("could not find NSX-T manager by name '%s': %s", nsxtManagerName, err)
		}
		if len(nsxtManagers) != 1 {
			return nil, fmt.Errorf("%s found %d NSX-T managers with name '%s'",
				govcd.ErrorEntityNotFound, len(nsxtManagers), nsxtManagerName)
		}

		id := extractUuid(nsxtManagers[0].HREF)
		nsxtManagerUrn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", id)
		if err != nil {
			return nil, fmt.Errorf("could not construct URN from id '%s': %s", id, err)
		}

		profile, err = govcd.GetNetworkContextProfilesByNameScopeAndContext(&vcdClient.Client, profileName,
			types.ApplicationPortProfileScopeProvider, nsxtManagerUrn)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Network Context Profile with name '%s' (NSX-T Manager '%s'): %s",
				profileName, nsxtManagerName, err)
		}

		dSet(d, "org", "System")
		dSet(d, "context_id", nsxtManagerUrn)
		dSet(d, "scope", types.ApplicationPortProfileScopeProvider)

	case 3: // TENANT scope
		orgName, vdcOrVdcGroupName, profileName := resourceURI[0], resourceURI[1], resourceURI[2]

		// define an interface type to match VDC and VDC Groups
		var vdcOrVdcGroup vdcOrVdcGroupHandler
		var contextId string
		_, vdcOrVdcGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcOrVdcGroupName)
		if govcd.ContainsNotFound(err) {
			adminOrg, err := vcdClient.GetAdminOrg(orgName)
			if err != nil {
				return nil, fmt.Errorf("error retrieving Admin Org for '%s': %s", orgName, err)
			}

			vdcGroup, err := adminOrg.GetVdcGroupByName(vdcOrVdcGroupName)
			if err != nil {
				return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s': %s", vdcOrVdcGroupName, err)
			}
			vdcOrVdcGroup = vdcGroup
			contextId = vdcGroup.VdcGroup.Id
		} else {
			if err != nil {
				return nil, fmt.Errorf("error retrieving VDC '%s' in Org '%s': %s", vdcOrVdcGroupName, orgName, err)
			}
			vdc, ok := vdcOrVdcGroup.(*govcd.Vdc)
			if !ok {
				return nil, fmt.Errorf("expected to get VDC by name '%s'", vdcOrVdcGroupName)
			}
			contextId = vdc.Vdc.ID
		}

		if !vdcOrVdcGroup.IsNsxt() {
			return nil, errors.New("NSX-T Network Context Profiles are only supported by NSX-T VDCs/VDC Groups")
		}

		profile, err = govcd.GetNetworkContextProfilesByNameScopeAndContext(&vcdClient.Client, profileName,
			types.ApplicationPortProfileScopeTenant, contextId)
		if err != nil {
			return nil, fmt.Errorf("unable to find NSX-T Network Context Profile '%s': %s", profileName, err)
		}

		dSet(d, "org", orgName)
		dSet(d, "context_id", contextId)
		dSet(d, "scope", types.ApplicationPortProfileScopeTenant)

	default:
		return nil, fmt.Errorf("resource path must be specified in one of two formats, based on Network Context Profile scope:\n" +
			"* PROVIDER (path contains 2 pieces nsxt_manager_name.profile_name)\n" +
			"* TENANT (path contains 3 pieces org-name.vdc-name-or-vdc-group-name.profile_name)")
	}

	d.SetId(profile.ID)

	return []*schema.ResourceData{d}, nil
}

func validateNetworkContextProfileScope(d *schema.ResourceData) error {
	scope := d.Get("scope").(string)
	orgName := d.Get("org").(string)

	if scope == types.ApplicationPortProfileScopeProvider && orgName != "" && strings.ToUpper(orgName) != "SYSTEM" {
		return fmt.Errorf("scope 'PROVIDER' requires Org to be \"System\"")
	}

	return nil
}

func getNsxtNetworkContextProfileType(d *schema.ResourceData, org *govcd.Org) (*types.NsxtNetworkContextProfile, error) {
	profileConfig := &types.NsxtNetworkContextProfile{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Scope:       d.Get("scope").(string),
	}

	// context_id can be VDC, VDC Group or NSX-T Manager (called 'network provider' in some docs)
	contextId := d.Get("context_id").(string)
	profileConfig.ContextEntityID = contextId

	// VDC and VDC Group based profiles contain Org references, while NSX-T Manager based ones don't
	if govcd.OwnerIsVdcGroup(contextId) || govcd.OwnerIsVdc(contextId) {
		profileConfig.OrgRef = &types.OpenApiReference{ID: org.Org.ID}
	}

	attributeSet := d.Get("attribute").(*schema.Set)
	attributeSlice := attributeSet.List()
	attributes := make([]types.NsxtNetworkContextProfileAttributes, len(attributeSlice))
	seenAttributeTypes := make(map[string]bool)
	for index, singleAttribute := range attributeSlice {
		attributeMap := singleAttribute.(map[string]interface{})
		attributeType := attributeMap["type"].(string)

		// VCD rejects payloads containing more than one attribute of the same type, but the error
		// only surfaces after task polling. Catching it here gives a quicker and clearer message
		if seenAttributeTypes[attributeType] {
			return nil, fmt.Errorf("only one attribute of type '%s' can be specified, with all values listed in its 'values' set", attributeType)
		}
		seenAttributeTypes[attributeType] = true

		oneAttribute := types.NsxtNetworkContextProfileAttributes{
			Type:   attributeType,
			Values: convertSchemaSetToSliceOfStrings(attributeMap["values"].(*schema.Set)),
		}

		subAttributeSet := attributeMap["sub_attribute"].(*schema.Set)
		if subAttributeSet.Len() > 0 {
			if attributeType != "APP_ID" {
				return nil, fmt.Errorf("sub_attribute blocks are only supported for attributes of type 'APP_ID'")
			}
			subAttributes := make([]types.NsxtNetworkContextProfileSubAttribute, subAttributeSet.Len())
			for subIndex, singleSubAttribute := range subAttributeSet.List() {
				subAttributeMap := singleSubAttribute.(map[string]interface{})
				subAttributes[subIndex] = types.NsxtNetworkContextProfileSubAttribute{
					Type:   subAttributeMap["type"].(string),
					Values: convertSchemaSetToSliceOfStrings(subAttributeMap["values"].(*schema.Set)),
				}
			}
			oneAttribute.SubAttributes = subAttributes
		}

		attributes[index] = oneAttribute
	}
	profileConfig.Attributes = attributes

	return profileConfig, nil
}

// setNsxtNetworkContextProfileData sets Terraform schema from types.NsxtNetworkContextProfile
//
// Note. GET responses do not return 'contextEntityId' (just like Application Port Profiles), so
// 'context_id' cannot be refreshed and remains a configuration-only field
func setNsxtNetworkContextProfileData(d *schema.ResourceData, profile *types.NsxtNetworkContextProfile) error {
	dSet(d, "name", profile.Name)
	dSet(d, "description", profile.Description)
	dSet(d, "scope", profile.Scope)

	attributeSlice := make([]interface{}, len(profile.Attributes))
	for index, attribute := range profile.Attributes {
		attributeMap := make(map[string]interface{})
		attributeMap["type"] = attribute.Type
		attributeMap["values"] = convertStringsToTypeSet(attribute.Values)

		subAttributeSlice := make([]interface{}, len(attribute.SubAttributes))
		for subIndex, subAttribute := range attribute.SubAttributes {
			subAttributeMap := make(map[string]interface{})
			subAttributeMap["type"] = subAttribute.Type
			subAttributeMap["values"] = convertStringsToTypeSet(subAttribute.Values)
			subAttributeSlice[subIndex] = subAttributeMap
		}
		attributeMap["sub_attribute"] = schema.NewSet(schema.HashResource(networkContextProfileSubAttributeDefinition), subAttributeSlice)

		attributeSlice[index] = attributeMap
	}

	attributeSet := schema.NewSet(schema.HashResource(networkContextProfileAttributeDefinition), attributeSlice)
	err := d.Set("attribute", attributeSet)
	if err != nil {
		return fmt.Errorf("error setting attribute set: %s", err)
	}

	return nil
}
