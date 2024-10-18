package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const confluentLicenseTypeWithLicense = "With License"
const confluentLicenseTypeNoLicense = "No License"

var globalDataSolutionPublishLockKey = "vcd_dse_solution_publish"

func resourceVcdDsePublish() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdDsePublishCreate,
		ReadContext:   resourceVcdDsePublishRead,
		DeleteContext: resourceVcdDsePublishDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdDsePublishImport,
		},

		Schema: map[string]*schema.Schema{
			"data_solution_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of Data Solution",
			},
			// Note. This resource publishes single DSE Solution to a single tenant as opposed to
			// 'vcd_solution_add_on_instance_publish.org_ids' that can publish to many tenants at
			// once. The reason this is done because publishing Solution Add-On Instance is a lot
			// simpler than DSE Solution.
			// Publishing DSE does many things:
			// * Publishes rights bundle
			// * Publishes access for that particular Data Solution (this defines if UI shows whether a Data Solution is published to a particular tenant)
			// * Publishes access for Data Solutions Operator (DSO)
			// * Publishes all Data Solution Instance Templates
			// Additionally, for "Confluent Platform" it can provision licensing data which is
			// custom for each tenant.
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A tenant ID that will get the Data Solution Published",
			},
			"dso_acl_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ACL ID for Data Solutions Operator",
			},
			"template_acl_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Solution Template ACL IDs provisioned to the tenant",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"confluent_license_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // This setting cannot be modified after publishing
				Description:  fmt.Sprintf("Only for 'Confluent Platform'. One of '%s' or '%s'", confluentLicenseTypeNoLicense, confluentLicenseTypeWithLicense),
				ValidateFunc: validation.StringInSlice([]string{confluentLicenseTypeNoLicense, confluentLicenseTypeWithLicense}, false),
			},
			"confluent_license_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true, // This setting cannot be modified after publishing
				Description: "Only for 'Confluent Platform'. Required if 'confluent_license_type = With License'",
			},
			"ds_org_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data Solution Org Configuration ID",
			},
		},
	}
}

func resourceVcdDsePublishCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution publishing started")
	vcdClient := meta.(*VCDClient)

	// The operations are quick, but performing them concurrently on multiple entries
	// poses a risk of hitting VCD database concurency errors
	vcdMutexKV.kvLock(globalDataSolutionPublishLockKey)
	defer vcdMutexKV.kvUnlock(globalDataSolutionPublishLockKey)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution Configuration: %s", err)
	}

	orgId := d.Get("org_id").(string)
	// Data Solution "Confluent Platform" has a custom baked menu for choosing licensing type
	// that cannot be dynamically defined.
	if dataSolution.Name() == "Confluent Platform" {
		licenseType := d.Get("confluent_license_type").(string)
		licenseKey := d.Get("confluent_license_key").(string)

		dsOrgConfigId, err := createConfluentOrgConfig(vcdClient, dataSolution, orgId, licenseType, licenseKey)
		if err != nil {
			return diag.FromErr(err)
		}
		dSet(d, "ds_org_config_id", dsOrgConfigId)
	}

	acl, dsoAcl, templateAcls, err := dataSolution.Publish(orgId)
	if err != nil {
		return diag.Errorf("error publishing Data Solution '%s' to Org with IDs '%s': %s",
			dataSolution.Name(), orgId, err)
	}
	dSet(d, "dso_acl_id", dsoAcl.Id)

	aclIds := make([]string, len(templateAcls))
	for aclIndex, acl := range templateAcls {
		aclIds[aclIndex] = acl.Id
	}
	err = d.Set("template_acl_ids", convertStringsToTypeSet(aclIds))
	if err != nil {
		return diag.Errorf("error storing Solution Template ACL IDs: %s", err)
	}

	// The main ACL is what matters when UI is showing whether a particular Data Solution is published to a tenant or not
	// For reference - publishing does more than that:
	// * Publishes rights bundle "vmware:dataSolutionsRightsBundle" which is created upon Solution
	// Add-On Instatiation
	// * Publishes access for that particular Data Solution (this defines if UI shows whether a Data
	// Solution is published to a particular tenant)
	// * Publishes access for Data Solutions Operator (DSO)
	// * Publishes all Data Solution Instance Templates

	d.SetId(acl.Id)

	return nil
}

func resourceVcdDsePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdDsePublishRead(ctx, d, meta, "resource")
}

func genericVcdDsePublishRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution Publishing read started with origin %s", origin)
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	// Check if the data solution provided exists at all
	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) && origin == "resource" {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving Data Solution: %s", err)
	}

	var acl *types.DefinedEntityAccess

	if origin == "resource" {
		acl, err = dataSolution.GetAccessControlById(d.Id())
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}

			return diag.Errorf("error retrieving Data Solution ACL by ID: %s", err)
		}
	} else {
		allAcls, err := dataSolution.GetAllAccessControlsForTenant(orgId)
		if err != nil {
			return diag.Errorf("error retrieving Data Solution ACL for Org %s: %s", orgId, err)
		}

		if len(allAcls) < 1 {
			return diag.Errorf("no Data Solution %s ACL for Org %s was found", dataSolution.Name(), orgId)
		}

		if len(allAcls) > 1 {
			return diag.Errorf("found more than 1 (%d) Data Solution %s ACLs for Org %s was found",
				len(allAcls), dataSolution.Name(), orgId)
		}
		acl = allAcls[0]
	}

	if dataSolution.Name() == "Confluent Platform" {
		dsOrgConfig, err := dataSolution.GetDataSolutionOrgConfigForTenant(orgId)
		if err != nil {
			return diag.Errorf("error retrieving Data Solution %s Org Configuration for Org %s: %s",
				dataSolution.Name(), orgId, err)
		}

		dSet(d, "ds_org_config_id", dsOrgConfig.RdeId())
		if dsOrgConfig.DataSolutionOrgConfig != nil && dsOrgConfig.DataSolutionOrgConfig.Spec != nil {
			specData := dsOrgConfig.DataSolutionOrgConfig.Spec["data"]
			if specData != nil {
				specDataMap := specData.(map[string]interface{})
				licenseType := specDataMap["LicenseType"]
				licenseTypeString := licenseType.(string)
				dSet(d, "confluent_license_type", licenseTypeString)
			}
		}
	}

	// Get Default DSO ACL ID
	defaultDso, err := vcdClient.GetDataSolutionByName(defaultDsoName)
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", defaultDsoName, err)
	}
	dsoAcls, err := defaultDso.GetAllAccessControlsForTenant(orgId)
	if err != nil {
		return diag.Errorf("error retrieving %s ACLs: %s", defaultDsoName, err)
	}
	if len(dsoAcls) > 1 {
		dSet(d, "dso_acl_id", dsoAcls[0].Id)
	} else {
		dSet(d, "dso_acl_id", "")
	}

	// Read Template ACL IDs
	allAclIds := make([]string, 0)
	allTemplates, err := dataSolution.GetAllInstanceTemplates()
	if err != nil {
		return diag.Errorf("error retrieving all Solution Instance Templates: %s", err)
	}
	for _, template := range allTemplates {
		acls, err := template.GetAllAccessControlsForTenant(orgId)
		if err != nil {
			return diag.Errorf("error retrieving ACL for Solution Instance Template: %s", err)
		}
		for _, acl := range acls {
			allAclIds = append(allAclIds, acl.Id)
		}
	}

	err = d.Set("template_acl_ids", convertStringsToTypeSet(allAclIds))
	if err != nil {
		return diag.Errorf("error storing Solution Template ACL IDs: %s", err)
	}

	d.SetId(acl.Id)

	return nil
}

func resourceVcdDsePublishDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution unpublishing started")
	vcdClient := meta.(*VCDClient)

	dataSolution, err := vcdClient.GetDataSolutionById(d.Get("data_solution_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution: %s", err)
	}

	// Confluent Platform requires cleanup
	if dataSolution.Name() == "Confluent Platform" {
		dsOrgConfig, err := dataSolution.GetDataSolutionOrgConfigForTenant(d.Get("org_id").(string))
		if err != nil {
			return diag.Errorf("error retrieving Data Solution Org Config for 'Confluent Platform': %s", err)
		}

		err = dsOrgConfig.Delete()
		if err != nil {
			return diag.Errorf("error deleting Data Solution Org Config for 'Confluent Platform': %s", err)
		}
	}

	acl, err := dataSolution.GetAccessControlById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Data Solution ACL by ID: %s", err)
	}

	err = dataSolution.DefinedEntity.DeleteAccessControl(acl)
	if err != nil {
		return diag.Errorf("error deleting Data Solution ACL: %s", err)
	}

	return nil
}

func resourceVcdDsePublishImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as \"data solution name\".org-name")
	}
	dataSolutionName, orgName := resourceURI[0], resourceURI[1]
	util.Logger.Printf("[TRACE] Data Solution publishing import started. Data Solution Name '%s', Org Name '%s'", dataSolutionName, orgName)

	dataSolution, err := vcdClient.GetDataSolutionByName(dataSolutionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DSE Configuration: %s", err)
	}

	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org: %s", err)
	}

	mainAcl, err := dataSolution.GetAllAccessControlsForTenant(org.Org.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Data Solution '%s' ACL for Org '%s': %s", dataSolutionName, orgName, err)
	}

	dSet(d, "data_solution_id", dataSolution.RdeId())
	dSet(d, "org_id", org.Org.ID)
	d.SetId(mainAcl[0].Id)

	return []*schema.ResourceData{d}, nil
}

// licensetype: "No License" or "With License".
// privateSecureData.LicenseKey must have license key set when licensetype=="With License"
func createConfluentOrgConfig(vcdClient *VCDClient, dataSolution *govcd.DataSolution, orgId string, licenseType, licenseKey string) (string, error) {
	// validation
	if licenseType == "" {
		return "", fmt.Errorf("'confluent_license_type' must be specified for 'Confluent Platform'")
	}

	if licenseType == "With License" && licenseKey == "" {
		return "", fmt.Errorf("'confluent_license_key' must be specified if using type '%s'", confluentLicenseTypeWithLicense)
	}

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
		return "", fmt.Errorf("error creating Data Solution Org Configuration: %s", err)
	}
	util.Logger.Printf("[TRACE] Created Org Config for Data Solution '%s' %s",
		dataSolution.Name(), dsOrgConfig.RdeId())

	err = dsOrgConfig.DefinedEntity.Resolve()
	if err != nil {
		return dsOrgConfig.RdeId(), fmt.Errorf("error resolving Data Solution Org Config for '%s': %s",
			dataSolution.Name(), err)
	}

	return dsOrgConfig.RdeId(), nil
}
