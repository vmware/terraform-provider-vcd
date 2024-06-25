package vcd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

const confluenceLicenseTypeWithLicense = "With License"
const confluenceLicenseTypeNoLicense = "No License"

func resourceVcdDsePublish() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdDsePublishCreate,
		ReadContext:   resourceVcdDsePublishRead,
		UpdateContext: resourceVcdDsePublishUpdate,
		DeleteContext: resourceVcdDsePublishDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdDseRegistryConfigurationImport,
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
			"confluent_license_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // This setting cannot be modified after publishing
				Description:  fmt.Sprintf("Only for 'Confluent Platform'. One of 'No License' or 'With License'", confluenceLicenseTypeNoLicense, confluenceLicenseTypeWithLicense),
				ValidateFunc: validation.StringInSlice([]string{confluenceLicenseTypeNoLicense, confluenceLicenseTypeWithLicense}, false),
			},
			"confluent_license_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true, // This setting cannot be modified after publishing
				Description: "Only for 'Confluent Platform'. Required if 'confluent_license_type = With License'",
			},
		},
	}
}

func resourceVcdDsePublishCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution publishing started")
	vcdClient := meta.(*VCDClient)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving DSE Configuration: %s", err)
	}

	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))

	// Data Solution "Confluent Platform" has a custom baked menu for choosing licensing type
	// that cannot be dynamically defined.
	if dataSolution.Name() == "Confluent Platform" {
		licenseType := d.Get("confluent_license_type").(string)
		licenseKey := d.Get("confluent_license_key").(string)

		err = createConfluentOrgConfig(vcdClient, dataSolution, orgIds, licenseType, licenseKey)
		if err != nil {
			return diag.FromErr(err)
		}
	}

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
	util.Logger.Printf("[TRACE] Data Solution update publishing started")
	vcdClient := meta.(*VCDClient)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving DSE Configuration: %s", err)
	}

	old, new := d.GetChange("org_ids")
	oldOrgIds := convertSchemaSetToSliceOfStrings(old.(*schema.Set))
	newOrgIds := convertSchemaSetToSliceOfStrings(new.(*schema.Set))

	unpublishOrgIds := sliceDifference(oldOrgIds, newOrgIds)
	publishOrgIds := sliceDifference(newOrgIds, oldOrgIds)

	err = dataSolution.Unpublish(unpublishOrgIds)
	if err != nil {
		return diag.Errorf("error unpublishing Data Solution '%s' to Orgs with IDs '%s': %s",
			dataSolution.Name(), strings.Join(unpublishOrgIds, ","), err)
	}

	// Data Solution "Confluent Platform" has a custom baked menu for choosing licensing type
	// that cannot be dynamically defined.
	if dataSolution.Name() == "Confluent Platform" {
		licenseType := d.Get("confluent_license_type").(string)
		licenseKey := d.Get("confluent_license_key").(string)

		err = createConfluentOrgConfig(vcdClient, dataSolution, publishOrgIds, licenseType, licenseKey)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = dataSolution.Publish(publishOrgIds)
	if err != nil {
		return diag.Errorf("error publishing Data Solution '%s' to Orgs with IDs '%s': %s",
			dataSolution.Name(), strings.Join(publishOrgIds, ","), err)
	}

	return resourceVcdDsePublishRead(ctx, d, meta)
}

func resourceVcdDsePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution Publishing read started")
	vcdClient := meta.(*VCDClient)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution: %s", err)
	}

	allAcls, err := dataSolution.GetAllAccessControls(nil)
	if err != nil {
		return diag.Errorf("error retrieving Access Controls for Data Solution '%s': %s", dataSolution.Name(), err)
	}

	orgIds := make([]string, 0)
	for _, acl := range allAcls {
		// Skipping "System" access rights that are created automatically and are not managed for
		// tenants
		if acl.Tenant.Name == "System" {
			continue
		}
		orgIds = append(orgIds, acl.Tenant.ID)
	}

	orgIdSet := convertStringsToTypeSet(orgIds)
	err = d.Set("org_ids", orgIdSet)
	if err != nil {
		return diag.Errorf("error storing 'org_ids': %s", err)
	}

	return nil
}

func resourceVcdDsePublishDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution unpublishing started")
	vcdClient := meta.(*VCDClient)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution: %s", err)
	}

	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))
	err = dataSolution.Unpublish(orgIds)
	if err != nil {
		return diag.Errorf("error unpublishing Data Solution '%s' to Orgs with IDs '%s': %s",
			dataSolution.Name(), strings.Join(orgIds, ","), err)
	}

	return nil
}

// sliceDifference returns the elements that are present in 'first' slice but not available in 'second'
func sliceDifference(first, second []string) []string {
	result := make([]string, 0)
	for _, firstElement := range first {
		if slices.Contains(second, firstElement) {
			continue
		}
		result = append(result, firstElement)
	}

	return result
}

// licensetype: "No License" or "With License".
// privateSecureData.LicenseKey must have license key set when licensetype=="With License"
func createConfluentOrgConfig(vcdClient *VCDClient, dataSolution *govcd.DataSolution, orgIds []string, licenseType, licenseKey string) error {
	// validation
	if licenseType == "" {
		return fmt.Errorf("'confluent_license_type' must be specified for 'Confluent Platform'")
	}

	if licenseType == "With License" && licenseKey == "" {
		return fmt.Errorf("'confluent_license_key' must be specified if using type '%s'", confluenceLicenseTypeWithLicense)
	}

	for _, orgId := range orgIds {
		cfg := &types.DataSolutionOrgConfig{
			APIVersion: "vcloud.vmware.com/v1alpha1",
			Kind:       "DSOrgConfig",
			Metadata: map[string]interface{}{
				"name":       "data-solutions-confluent-platform-org-config",
				"generation": 1,
			},
			Spec: map[string]interface{}{
				"solutionType": dataSolution.DefinedEntity.DefinedEntity.Name,
				"data": map[string]string{
					"LicenseType": licenseType,
				},
			},
		}

		if licenseType == "With License" {
			cfg.Spec["privateSecureData"] = map[string]string{
				"LicenseKey": licenseKey,
			}
		}

		dsOrgConfig, err := vcdClient.CreateDataSolutionOrgConfig(orgId, cfg)
		if err != nil {
			return fmt.Errorf("error submitting Data Solution Org Configuration: %s", err)
		}
		util.Logger.Printf("[TRACE] Created Org Config for Data Solution 'Confluent Platform' %s", dsOrgConfig.RdeId())

		err = dsOrgConfig.DefinedEntity.Resolve()
		if err != nil {
			return fmt.Errorf("error resolving Data Solution Org Config for 'Confluent Platform': %s", err)
		}

	}

	return nil
}

// {
// 	"name": "urn:vcloud:org:04aa3a27-8f23-45c4-acff-b2427431dbf2",
// 	"entity": {
// 	  "apiVersion": "vcloud.vmware.com/v1alpha1",
// 	  "kind": "DSOrgConfig",
// 	  "metadata": {
// 		"name": "data-solutions-confluent-platform-org-config",
// 		"generation": 1,
// 		"annotations": {
// 		  "vcd-ds-meta/secure": ""
// 		}
// 	  },
// 	  "spec": {
// 		"solutionType": "Confluent Platform",
// 		"privateData": {},
// 		"data": {
// 		  "LicenseType": "With License"
// 		},
// 		"privateSecureData": {
// 		  "LicenseKey": "LICENSE-KEY-FAKE"
// 		}
// 	  }
// 	}
//   }
