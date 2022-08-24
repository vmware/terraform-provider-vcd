package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"log"
	"net/url"
)

func datasourceVcdVmPlacementPolicy() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdVmPlacementPolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the VM Placement Policy",
			},
			"provider_vdc_id": {
				Type:     schema.TypeString,
				Required: true,
				Description: "ID of the Provider VDC to which the VM Placement Policy belongs",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Description of the VM Placement Policy",
			},
			"vm_groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Collection of VMs with similar host requirements",
			},
			"logical_vm_groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "One or more Logical VM Groups defined in this VM Placement policy. There is an AND relationship among all the entries fetched to this attribute",
			},
		},
	}
}

func datasourceVcdVmPlacementPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	pVdcId := d.Get("provider_vdc_id").(string)
	log.Printf("[TRACE] VM Placement Policy read initiated: %s in pVDC with ID %s", policyName, pVdcId)

	vcdClient := meta.(*VCDClient)

	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	var policy *govcd.VdcComputePolicy
	var err error
	if d.Id() != "" {
		policy, err = vcdClient.Client.GetVdcComputePolicyById(d.Id())
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM Placement Policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM Placement Policy %s, err: %s. Removing from tfstate", policyName, err)
		}
	}

	// The secondary method of retrieval is from name
	if d.Id() == "" {
		if policyName == "" {
			return diag.Errorf("both Placement Policy name and ID are empty")
		}
		if pVdcId == "" {
			return diag.Errorf("both Provider VDC ID and Placement Policy ID are empty")
		}

		method = "name"
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name==%s;pvdc==%s",policyName, pVdcId))
		filteredPoliciesByName, err := vcdClient.Client.GetAllVdcComputePolicies(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM Placement policy %s of Provider VDC %s. Removing from tfstate.", policyName, pVdcId)
			d.SetId("")
			return diag.Errorf("unable to find VM Placement policy %s of Provider VDC %s, err: %s. Removing from tfstate", policyName, pVdcId, err)
		}
		if len(filteredPoliciesByName) != 1 {
			log.Printf("[DEBUG] Unable to find VM Placement policy %s of Provider VDC %s. Found Policies by name: %d. Removing from tfstate.", policyName, pVdcId, len(filteredPoliciesByName))
			d.SetId("")
			return diag.Errorf("[DEBUG] Unable to find VM Placement policy %s of Provider VDC %s, err: %s. Found Policies by name: %d. Removing from tfstate", policyName, pVdcId, govcd.ErrorEntityNotFound, len(filteredPoliciesByName))
		}
		policy = filteredPoliciesByName[0]
		d.SetId(policy.VdcComputePolicy.ID)
	}

	// Fix coverity warning
	if policy == nil {
		return diag.Errorf("[datasourceVcdVmPlacementPolicyRead] error defining VM Placement Policy")
	}
	util.Logger.Printf("[TRACE] [get VM Placement Policy] Retrieved by %s\n", method)
	return setVmPlacementPolicy(ctx, d, *policy.VdcComputePolicy)
}

// TODO: Probably we should move this function to the Resource when it is created, to follow same code style as other resource-datasource pairs.
// setVmPlacementPolicy sets object state from *govcd.VdcComputePolicy
func setVmPlacementPolicy(_ context.Context, d *schema.ResourceData, policy types.VdcComputePolicy) diag.Diagnostics {

	dSet(d, "name", policy.Name)
	dSet(d, "description", policy.Description)
	dSet(d, "vm_groups", policy.NamedVMGroups) // FIXME: Flatten the structure
	dSet(d, "logical_vm_groups", policy.LogicalVMGroupReferences) // FIXME: Flatten the structure

	log.Printf("[TRACE] VM Placement Policy read completed: %s", policy.Name)
	return nil
}