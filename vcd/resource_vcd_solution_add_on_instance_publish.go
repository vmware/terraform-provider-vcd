package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdSolutionAddonInstancePublish() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionAddonInstancePublishCreateUpdate,
		ReadContext:   resourceVcdSolutionAddonInstancePublishRead,
		UpdateContext: resourceVcdSolutionAddonInstancePublishCreateUpdate,
		DeleteContext: resourceVcdSolutionAddonInstancePublishDelete,
		// Import is exactly the same as for Solution Add-On Instance
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionAddonInstanceImport,
		},

		Schema: map[string]*schema.Schema{
			"add_on_instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Solution Add-On Instance ID",
			},
			"org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of Organization IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"publish_to_all_tenants": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Publish Solution Add-On Instance to all tenants",
			},
		},
	}
}

func resourceVcdSolutionAddonInstancePublishCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Get("add_on_instance_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	orgIds := convertSchemaSetToSliceOfStrings(d.Get("org_ids").(*schema.Set))

	scopes, err := orgIdsToNames(vcdClient, orgIds)
	if err != nil {
		return diag.Errorf("error converting Org IDs to Names: %s", err)
	}
	publishToAll := d.Get("publish_to_all_tenants").(bool)

	if len(scopes) > 0 && publishToAll {
		return diag.Errorf("'org_ids' must be empty when 'publish_to_all_tenants' is set")
	}

	_, err = addOnInstance.Publishing(scopes, publishToAll)
	if err != nil {
		return diag.Errorf("error publishing Solution Add-On Instance %s: %s", addOnInstance.SolutionAddOnInstance.Name, err)
	}

	d.SetId(addOnInstance.RdeId())

	return resourceVcdSolutionAddonInstancePublishRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstancePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	d.Set("publish_to_all_tenants", addOnInstance.SolutionAddOnInstance.Scope.AllTenants)
	orgNames := addOnInstance.SolutionAddOnInstance.Scope.Tenants

	orgIds, err := orgNamesToIds(vcdClient, orgNames)
	if err != nil {
		return diag.Errorf("error converting Org IDs to Names: %s", err)
	}

	orgIdsSet := convertStringsToTypeSet(orgIds)
	err = d.Set("org_ids", orgIdsSet)
	if err != nil {
		return diag.Errorf("error storing Org IDs: %s", err)
	}

	dSet(d, "add_on_instance_id", addOnInstance.RdeId())

	return nil
}

func resourceVcdSolutionAddonInstancePublishDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	_, err = addOnInstance.Publishing(nil, false)
	if err != nil {
		return diag.Errorf("error unpublishing Solution Add-On Instance %s: %s", addOnInstance.SolutionAddOnInstance.Name, err)
	}

	return resourceVcdSolutionAddonInstancePublishRead(ctx, d, meta)
}

func orgIdsToNames(vcdClient *VCDClient, orgIds []string) ([]string, error) {
	existingOrgs, err := vcdClient.GetOrgList()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all the Organizations: %s", err)
	}

	if len(orgIds) == 0 {
		return []string{}, nil
	}

	orgNames := make([]string, 0)

	for _, orgId := range orgIds {
		for _, org := range existingOrgs.Org {
			if haveSameUuid(org.HREF, orgId) { // ensure that URN vs UUID formats are verified
				orgNames = append(orgNames, org.Name)
			}
		}
	}

	return orgNames, nil
}

func orgNamesToIds(vcdClient *VCDClient, orgNames []string) ([]string, error) {
	existingOrgs, err := vcdClient.GetOrgList()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all the Organizations: %s", err)
	}

	if len(existingOrgs.Org) == 0 {
		return []string{}, nil
	}

	orgUrns := make([]string, 0)

	for _, orgName := range orgNames {
		for _, org := range existingOrgs.Org {
			if org.Name == orgName {
				orgUuid := extractUuid(org.HREF)
				orgUrn := fmt.Sprintf("urn:vcloud:org:%s", orgUuid)
				orgUrns = append(orgUrns, orgUrn)
			}
		}
	}

	return orgUrns, nil
}
