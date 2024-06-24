package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdDsePublish() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdDsePublishCreate,
		ReadContext:   resourceVcdDsePublishRead,
		UpdateContext: resourceVcdDsePublishUpdate,
		DeleteContext: resourceVcdDsePublishDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdDsePublishImport,
		},

		Schema: map[string]*schema.Schema{
			"data_solution_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of Data Solution",
			},

			"org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of Organization IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dse_right_bundle_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "vmware:dataSolutionsRightsBundle",
				Description: "The ID of Data Solution",
			},
			"data_solutions_package_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "VCD Data Solutions",
				Description: "The name of Data Solutions configuration entry",
			},
			//
			// "vdc": {
			// 	Type:        schema.TypeString,
			// 	Optional:    true,
			// 	ForceNew:    true,
			// 	Description: "The name of VDC to use, optional if defined at provider level",
			// },
			// "edge_gateway": {
			// 	Type:        schema.TypeString,
			// 	Required:    true,
			// 	ForceNew:    true,
			// 	Description: "Edge gateway name in which NAT Rule is located",
			// },
			// "network_name": {
			// 	Type:        schema.TypeString,
			// 	Required:    true,
			// 	Description: "Org or external network name",
			// },
		},
	}
}

func resourceVcdDsePublishCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] DSE publishing started")
	vcdClient := meta.(*VCDClient)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving DSE Configuration: %s", err)
	}

	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))
	err = dataSolution.Publish(orgIds)
	if err != nil {
		return diag.Errorf("error publishing Data Solution '%s' to Orgs with IDs '%s': %s",
			dataSolution.Name(), strings.Join(orgIds, ","), err)
	}

	d.SetId(dataSolution.RdeId())

	return nil
}

// func resourceVcdDsePublishCreate2(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	util.Logger.Printf("[TRACE] DSE publishing started")
// 	vcdClient := meta.(*VCDClient)

// 	// Check if the data solution provided exists at all
// 	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
// 	if err != nil {
// 		return diag.Errorf("error retrieving DSE Configuration: %s", err)
// 	}

// 	// Publish operation does 3 things for given tenants:
// 	// 1. Publishes DSE right bundle to tenant
// 	// IF the Data Solution has a requirement for additional configuration - this can be stored
// 	// * Post to DSO Org Config (only for custom configuration)

// 	//

// 	// 2. Set access control for a given tenant - Specific Data Solution and always VCD Data Solutions
// 	// 3. Set access controls for all instance templates

// 	// 1. Publish rights bundle to given tenants
// 	rightsBundleName := d.Get("dse_right_bundle_name").(string)
// 	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))
// 	err = publishRightsBundleToTenants(vcdClient, rightsBundleName, orgIds)
// 	if err != nil {
// 		return diag.Errorf("error publishing rights bundle '%s': %s", rightsBundleName, err)
// 	}

// 	//////// Publish DS Org Config
// 	///
// 	///
// 	///
// 	/* {
// 	  "name": "urn:vcloud:org:c722b866-1ce2-4ebd-ae16-5a1d33ff9d16",
// 	  "entity": {
// 	    "apiVersion": "vcloud.vmware.com/v1alpha1",
// 	    "kind": "DSOrgConfig",
// 	    "metadata": {
// 	      "name": "data-solutions-confluent-platform-org-config",
// 	      "generation": 1
// 	    },
// 	    "spec": {
// 	      "solutionType": "Confluent Platform",
// 	      "privateData": {},
// 	      "data": {
// 	        "LicenseType": "No License"
// 	      },
// 	      "privateSecureData": {}
// 	    }
// 	  }
// 	} */

// 	////// Only for Confluent Platform (at least for now)
// 	/* for _, orgId := range orgIds {
// 		cfg := &types.DsOrgConfig{
// 			APIVersion: dataSolution.DseConfig.APIVersion,
// 			Kind:       "DSOrgConfig",
// 			Metadata: map[string]interface{}{
// 				"name":       "data-solutions-confluent-platform-org-config",
// 				"generation": 1,
// 			},
// 			Spec: map[string]interface{}{
// 				"solutionType": dataSolution.DefinedEntity.DefinedEntity.Name,
// 				"data": map[string]string{
// 					"LicenseType": "No License",
// 				},
// 			},
// 		}
// 		dsOrgConfig, err := vcdClient.CreateDsOrgConfig(orgId, cfg)
// 		if err != nil {
// 			return diag.Errorf("error submitting Data Solution Org Configuration: %s", err)
// 		}
// 		util.Logger.Printf("[TRACE] Created Data Solution Org Configuration with ID %s", dsOrgConfig.DefinedEntity.DefinedEntity.ID)

// 		acl := &types.DefinedEntityAccess{
// 			Tenant:        types.OpenApiReference{ID: orgId},
// 			GrantType:     "MembershipAccessControlGrant",
// 			AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
// 			MemberID:      orgId,
// 		}

// 		// Access control for "DSE Org Config"
// 		_, err = dsOrgConfig.DefinedEntity.SetAccessControl(acl)
// 		if err != nil {
// 			return diag.Errorf("error setting access right for Data Solution Org Config with ID '%s': %s", dsOrgConfig.DefinedEntity.DefinedEntity.ID, err)
// 		}
// 	} */
// 	////// Only for Confluent Platform (at least for now)

// 	dsoName := d.Get("data_solutions_package_name").(string)
// 	err = provisionAccessToEntities(vcdClient, dataSolution, dsoName, orgIds)
// 	if err != nil {
// 		return diag.Errorf("error providing Access: %s", err)
// 	}

// 	// Publish each available Data Solution Instance template to each given tenant

// 	err = provisionAccessToTemplates(dataSolution, orgIds)
// 	if err != nil {
// 		return diag.Errorf("error provisioning Template Access for Data Solution %s: %s", dataSolution.DataSolution.Kind, err)
// 	}

// 	d.SetId(dataSolution.RdeId())

// 	return resourceVcdDsePublishRead(ctx, d, meta)
// }

func resourceVcdDsePublishUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdDsePublishRead(ctx, d, meta)
}

func resourceVcdDsePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdDsePublishDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	util.Logger.Printf("[TRACE] DSE publishing started")
	/* vcdClient := meta.(*VCDClient)

	// Publish operation does 3 things for given tenants:
	// 1. Publishes DSE right bundle to tenant
	// 2. Set access control for a given tenant
	// 3. Set access controls for all instance templates

	// rightBundleName := d.Get("dse_right_bundle_name").(string)

	// // UnPublish rights bundle. Not using the Terraform rights bundle step
	// rightsBundle, err := vcdClient.Client.GetRightsBundleByName(rightBundleName)
	// if err != nil {
	// 	return diag.Errorf("error retrieving right bundle %s: %s", rightBundleName, err)
	// }

	// orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))
	// orgOpenApiReferences := convertSliceOfStringsToOpenApiReferenceIds(orgIds)
	// err = rightsBundle.UnpublishTenants(orgOpenApiReferences)
	// if err != nil {
	// 	return diag.Errorf("error unpublishing %s from tenants '%s': %s",
	// 		rightBundleName, strings.Join(orgIds, ","), err)
	// }

	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving DSE Configuration: %s", err)
	}

	allInstanceTemplates, err := dataSolution.GetAllInstanceTemplatesBySolutionType(dataSolution.DataSolution.Spec.SolutionType)
	if err != nil {
		return diag.Errorf("error retrieving template instances for Solution Type '%s': %s", dataSolution.DataSolution.Spec.SolutionType, err)
	}

	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))

	for _, instanceTemplate := range allInstanceTemplates {
		for _, orgId := range orgIds {

			// POST https://HOST/cloudapi/1.0.0/entities/urn:vcloud:entity:vmware:dsInstanceTemplate:c9882fc0-d8ce-4ec6-9c5e-1b4c5d405cca/accessControls
			//{
			// 	"tenant": {
			// 	  "id": "urn:vcloud:org:12c2bea9-380d-4586-92cf-52385d4228fc"
			// 	},
			// 	"grantType": "MembershipAccessControlGrant",
			// 	"accessLevelId": "urn:vcloud:accessLevel:ReadOnly",
			// 	"memberId": "urn:vcloud:org:12c2bea9-380d-4586-92cf-52385d4228fc"
			//}

			acl := &types.DefinedEntityAccess{
				Tenant:        types.OpenApiReference{ID: orgId},
				GrantType:     "MembershipAccessControlGrant",
				AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
				MemberID:      orgId,
				// ObjectId:      instanceTemplate.RdeId(),
			}
			// Modify DSE entry access
			_, err = instanceTemplate.DefinedEntity.SetAccessControl(acl)
			if err != nil {
				return diag.Errorf("error setting access right for Org '%s': %s", orgId, err)
			}
		}
	} */

	return nil
}

func resourceVcdDsePublishImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func publishRightsBundleToTenants(vcdClient *VCDClient, rightsBundleName string, tenantIds []string) error {
	rightsBundle, err := vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		return fmt.Errorf("error retrieving Rights Bundle %s: %s", rightsBundleName, err)
	}

	if len(tenantIds) == 0 {
		return nil
	}

	orgOpenApiReferences := convertSliceOfStringsToOpenApiReferenceIds(tenantIds)
	err = rightsBundle.PublishTenants(orgOpenApiReferences)
	if err != nil {
		return fmt.Errorf("error publishing %s to tenants '%s': %s",
			rightsBundleName, strings.Join(tenantIds, ","), err)
	}

	return nil
}

/* func provisionAccessToEntities(vcdClient *VCDClient, dataSolutionEntity *govcd.DataSolution, dsoEntityName string, tenantIds []string) error {
	dataSolutionsEntity, err := vcdClient.GetDataSolutionByName(dsoEntityName)
	if err != nil {
		return fmt.Errorf("error retrieving Data Solution entity: %s", err)
	}

	// loop over all given organizations and adding their access to DSE
	for _, orgId := range tenantIds {
		acl := &types.DefinedEntityAccess{
			Tenant:        types.OpenApiReference{ID: orgId},
			GrantType:     "MembershipAccessControlGrant",
			AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
			MemberID:      orgId,
		}

		// Access control for "VCD Data Solutions" itself
		_, err = dataSolutionsEntity.DefinedEntity.SetAccessControl(acl)
		if err != nil {
			return fmt.Errorf("error setting access right for main Data Solution Entry and Org '%s': %s", orgId, err)
		}

		// Access control for a particular Data Solution in question
		// Modify DSE entry access
		_, err = dataSolutionEntity.DefinedEntity.SetAccessControl(acl)
		if err != nil {
			return fmt.Errorf("error setting access right for Org '%s': %s", orgId, err)
		}
	}

	return nil
} */

/* func provisionAccessToTemplates(dataSolutionEntity *govcd.DataSolution, tenantIds []string) error {
	allInstanceTemplates, err := dataSolutionEntity.GetAllInstanceTemplatesBySolutionType(dataSolutionEntity.DseConfig.Spec.SolutionType)
	if err != nil {
		return fmt.Errorf("error retrieving template instances for Solution Type '%s': %s", dataSolutionEntity.DseConfig.Spec.SolutionType, err)
	}

	for _, instanceTemplate := range allInstanceTemplates {
		for _, orgId := range tenantIds {
			acl := &types.DefinedEntityAccess{
				Tenant:        types.OpenApiReference{ID: orgId},
				GrantType:     "MembershipAccessControlGrant",
				AccessLevelID: "urn:vcloud:accessLevel:ReadOnly",
				MemberID:      orgId,
			}
			// Modify DSE entry access
			_, err = instanceTemplate.DefinedEntity.SetAccessControl(acl)
			if err != nil {
				return fmt.Errorf("error setting access right for Org '%s': %s", orgId, err)
			}
		}
	}

	return nil
}
*/
