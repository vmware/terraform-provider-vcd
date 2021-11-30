package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var participatingOrgVdcsResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"fault_domain_tag": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Represents the fault domain of a given organization vDC.",
		},
		"network_provider_scope": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Specifies the network provider scope of the vDC.",
		},
		"remote_org": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Specifies whether the vDC is local to this VCD site",
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "The status that the vDC can be in e.g. 'SAVING', 'SAVED', 'CONFIGURING'," +
				" 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND'," +
				" 'UNCONFIGURED').",
		},
		"org_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Organization VDC belongs",
		},
		"org_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Organization VDC belongs",
		},
		"site_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Site VDC belongs",
		},
		"site_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Site VDC belongs",
		},
		"vdc_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "VDC name",
		},
		"vdc_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "VDC ID",
		},
	},
}

func resourceDataCenterGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdDataCenterGroupRead,
		CreateContext: resourceVcdDataCenterGroupCreate,
		UpdateContext: resourceVcdDataCenterGroupUpdate,
		DeleteContext: resourceVcdAlbDataCenterGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDataCenterGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of data center group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data center group description",
			},
			"dfw_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Distributed firewall status",
			},
			"starting_vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Starting VDC Id",
			},
			"participating_vdc_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Participating VCD IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"error_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "More detailed error message when datacenter group has error status",
			},
			"local_egress": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Status whether local egress is enabled for a universal router belonging " +
					"to a universal vDC group.",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of network pool to use if creating a local vDC group router.",
			},
			"network_pool_universal_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The network provider’s universal id that is backing the universal network pool.",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the networking provider backing the vDC Group.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The status that the group can be in (e.g. 'SAVING', 'SAVED', 'CONFIGURING'," +
					" 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND'," +
					" 'UNCONFIGURED').",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the group as LOCAL or UNIVERSAL.",
			},
			"universal_networking_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True means that a vDC group router has been created.",
			},
			"participating_org_vdcs": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The list of organization vDCs that are participating in this group.",
				Elem:        participatingOrgVdcsResource,
			},
		},
	}
}

// resourceVcdDataCenterGroupCreate covers Create functionality for resource
func resourceVcdDataCenterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroupConfig := getDataCenterGroupConfigurationType(d)
	createdVdcGroup, err := adminOrg.CreateNsxtVdcGroup(vdcGroupConfig.Name, vdcGroupConfig.Description, vdcGroupConfig.StartingVdcId, vdcGroupConfig.ParticipatingVdcIds)
	if err != nil {
		return diag.Errorf("error creating data center group: %s", err)
	}

	d.SetId(createdVdcGroup.VdcGroup.Id)
	return resourceVcdDataCenterGroupRead(ctx, d, meta)
}

// resourceVcdDataCenterGroupUpdate covers Update functionality for resource
func resourceVcdDataCenterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[data center group update] : %s", err)
	}

	vdcGroupConfig := getDataCenterGroupConfigurationType(d)
	_, err = vdcGroup.Update(vdcGroupConfig.Name, vdcGroupConfig.Description, vdcGroupConfig.ParticipatingVdcIds)
	if err != nil {
		return diag.Errorf("[data center group update] : %s", err)
	}

	return resourceVcdDataCenterGroupRead(ctx, d, meta)
}

func getDataCenterGroupConfigurationType(d *schema.ResourceData) VdcGroupConfig {
	// convert list of VDC IDs to slice of strings
	var vdcIds []string
	participatingVdcsIds := d.Get("participating_vdc_ids").(*schema.Set).List()
	for _, participatingVdcId := range participatingVdcsIds {
		vdcIds = append(vdcIds, participatingVdcId.(string))
	}
	return VdcGroupConfig{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		StartingVdcId:       d.Get("starting_vdc_id").(string),
		ParticipatingVdcIds: vdcIds,
	}
}

func resourceVcdDataCenterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[data center group read] : %s", err)
	}

	setVdcGroupConfigurationData(vdcGroup.VdcGroup, d)

	var participatingVdcIds []interface{}
	for _, participatingVdc := range vdcGroup.VdcGroup.ParticipatingOrgVdcs {
		participatingVdcIds = append(participatingVdcIds, participatingVdc.VdcRef.ID)
	}
	if len(participatingVdcIds) > 0 {
		err = d.Set("participating_vdc_ids", participatingVdcIds)
		if err != nil {
			return diag.Errorf("[data center group read] could not set participating_vdc_ids block: %s", err)
		}
	}
	return nil
}

func setVdcGroupConfigurationData(config *types.VdcGroup, d *schema.ResourceData) error {
	dSet(d, "name", config.Name)
	dSet(d, "description", config.Description)
	dSet(d, "dfw_enabled", config.DfwEnabled)
	dSet(d, "error_message", config.ErrorMessage)
	dSet(d, "local_egress", config.LocalEgress)
	dSet(d, "network_pool_id", config.NetworkPoolId)
	dSet(d, "network_pool_universal_id", config.NetworkPoolUniversalId)
	dSet(d, "network_provider_type", config.NetworkProviderType)
	dSet(d, "status", config.Status)
	dSet(d, "type", config.Type)
	dSet(d, "universal_networking_enabled", config.UniversalNetworkingEnabled)

	var candidateVdcsSlice []interface{}
	if len(config.ParticipatingOrgVdcs) > 0 {
		for _, candidateVdc := range config.ParticipatingOrgVdcs {

			candidateVdcMap := make(map[string]interface{})
			candidateVdcMap["fault_domain_tag"] = candidateVdc.FaultDomainTag
			candidateVdcMap["network_provider_scope"] = candidateVdc.NetworkProviderScope
			candidateVdcMap["remote_org"] = candidateVdc.RemoteOrg
			candidateVdcMap["status"] = candidateVdc.Status
			candidateVdcMap["org_name"] = candidateVdc.OrgRef.Name
			candidateVdcMap["org_id"] = candidateVdc.OrgRef.ID
			candidateVdcMap["site_name"] = candidateVdc.SiteRef.Name
			candidateVdcMap["site_id"] = candidateVdc.SiteRef.ID
			candidateVdcMap["vdc_name"] = candidateVdc.VdcRef.Name
			candidateVdcMap["vdc_id"] = candidateVdc.VdcRef.ID

			candidateVdcsSlice = append(candidateVdcsSlice, candidateVdcMap)
		}
	}

	err := d.Set("participating_org_vdcs", schema.NewSet(schema.HashResource(participatingOrgVdcsResource), candidateVdcsSlice))
	if err != nil {
		return fmt.Errorf("[data center group read] could not set participating_org_vdcs block: %s", err)
	}
	return nil
}

func resourceVcdAlbDataCenterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroupToDelete, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[data center group delete] : %s", err)
	}

	return diag.FromErr(vdcGroupToDelete.Delete())
}

func resourceDataCenterGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.data-center-group-name")
	}
	orgName, vdcGroupName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[data center group import] error retrieving org %s: %s", orgName, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupByName(vdcGroupName)
	if err != nil {
		return nil, fmt.Errorf("error importing data center group item: %s", err)
	}

	d.SetId(vdcGroup.VdcGroup.Id)
	dSet(d, "org", orgName)
	setVdcGroupConfigurationData(vdcGroup.VdcGroup, d)

	return []*schema.ResourceData{d}, nil
}

// VdcGroupConfig is a minimal structure defining a VdcGroup in Organization
type VdcGroupConfig struct {
	Name                string
	Description         string
	ParticipatingVdcIds []string
	StartingVdcId       string
}
