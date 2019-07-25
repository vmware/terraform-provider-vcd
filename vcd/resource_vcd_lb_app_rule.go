package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdLBAppRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdLBAppRuleCreate,
		Read:   resourceVcdLBAppRuleRead,
		Update: resourceVcdLBAppRuleUpdate,
		Delete: resourceVcdLBAppRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdLBAppRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the LB Application Rule is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the LB Application Rule is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the LB Application Rule is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique LB Application Rule name",
			},
			"script": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The script for the application rule. Note - you may find HEREDOC " +
					"useful to pass multiline strings",
			},
		},
	}
}

func resourceVcdLBAppRuleCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	LBRule, err := getLBAppRuleType(d)
	if err != nil {
		return fmt.Errorf("unable to create load balancer application rule type: %s", err)
	}

	createdPool, err := edgeGateway.CreateLbAppRule(LBRule)
	if err != nil {
		return fmt.Errorf("error creating new load balancer application rule: %s", err)
	}

	err = setLBAppRuleData(d, createdPool)
	if err != nil {
		return err
	}
	d.SetId(createdPool.Id)
	return resourceVcdLBAppRuleRead(d, meta)
}

func resourceVcdLBAppRuleRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBRule, err := edgeGateway.GetLbAppRuleById(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer application rule with Id %s: %s", d.Id(), err)
	}

	return setLBAppRuleData(d, readLBRule)
}

func resourceVcdLBAppRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBRuleConfig, err := getLBAppRuleType(d)
	if err != nil {
		return fmt.Errorf("could not create load balancer application rule type for update: %s", err)
	}

	updatedLBRule, err := edgeGateway.UpdateLbAppRule(updateLBRuleConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer application rule with Id %s: %s", d.Id(), err)
	}

	if err := setLBAppRuleData(d, updatedLBRule); err != nil {
		return err
	}

	return nil
}

func resourceVcdLBAppRuleDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteLbAppRuleById(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting load balancer application rule: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceVcdLBAppRuleImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the Id field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known Id of object.
//
// Example resource name (_resource_name_): vcd_lb_app_rule.my-test-app-rule
// Example import path (_the_id_string_): org.vdc.edge-gw.existing-app-rule
func resourceVcdLBAppRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw.existing-app-rule")
	}
	orgName, vdcName, edgeName, appRuleName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBRule, err := edgeGateway.GetLbAppRuleByName(appRuleName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer application rule with name %s: %s",
			d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", appRuleName)

	d.SetId(readLBRule.Id)
	return []*schema.ResourceData{d}, nil
}

// getLBAppRuleType converts Terraform resource data into types.LbAppRule type for API request.
func getLBAppRuleType(d *schema.ResourceData) (*types.LbAppRule, error) {
	lbAppRule := &types.LbAppRule{
		Name:   d.Get("name").(string),
		Script: d.Get("script").(string),
	}

	return lbAppRule, nil
}

// setLBAppRuleData sets name and script API fields. API output returns a single string terminated
// by newline ("\n") for each line of script. If a user wants to set multiline script Terraform's
// HEREDOC syntax is helpful.
// This terraform configuration
// script = <<-EOT
//   acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page
//   acl other_page2 url_beg / other2 redirect location https://www.other2.com/ ifother_page2
//   acl hello payload(0,6) -m bin 48656c6c6f0a
//   EOT
// is rendered as such API call
// <script>acl en req.fhdr(accept-language),language(es;fr;en) -m str en\nuse_backend english if en</script>
func setLBAppRuleData(d *schema.ResourceData, LBRule *types.LbAppRule) error {
	d.Set("script", LBRule.Script)
	d.Set("name", LBRule.Name)
	return nil
}
