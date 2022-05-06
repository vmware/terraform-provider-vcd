package vcd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVmSizingPolicy() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceVmSizingPolicyCreate,
		DeleteContext: resourceVmSizingPolicyDelete,
		ReadContext:   resourceVmSizingPolicyRead,
		UpdateContext: resourceVmSizingPolicyUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVmSizingPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Deprecated:  "Unneeded property, which was included by mistake",
				Optional:    true,
				Description: "The name of organization to use - Deprecated and unneeded: will be ignored if used ",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cpu": {
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"speed_in_mhz": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the vCPU speed of a core in MHz.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"count": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the number of vCPUs configured for a VM. This is a VM hardware configuration. When a tenant assigns the VM sizing policy to a VM, this count becomes the configured number of vCPUs for the VM.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"cores_per_socket": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "The number of cores per socket for a VM. This is a VM hardware configuration. The number of vCPUs that is defined in the VM sizing policy must be divisible by the number of cores per socket. If the number of vCPUs is not divisible by the number of cores per socket, the number of cores per socket becomes invalid.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"reservation_guarantee": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines how much of the CPU resources of a VM are reserved. The allocated CPU for a VM equals the number of vCPUs times the vCPU speed in MHz. The value of the attribute ranges between 0 and one. Value of 0 CPU reservation guarantee defines no CPU reservation. Value of 1 defines 100% of CPU reserved.",
							ValidateFunc: IsFloatAndBetween(0, 1),
						},
						"limit_in_mhz": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the CPU limit in MHz for a VM. If not defined in the VDC compute policy, CPU limit is equal to the vCPU speed multiplied by the number of vCPUs.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"shares": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the number of CPU shares for a VM. Shares specify the relative importance of a VM within a virtual data center. If a VM has twice as many shares of CPU as another VM, it is entitled to consume twice as much CPU when these two virtual machines are competing for resources. If not defined in the VDC compute policy, normal shares are applied to the VM.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
					},
				},
			},
			"memory": {
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_mb": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the memory configured for a VM in MB. This is a VM hardware configuration. When a tenant assigns the VM sizing policy to a VM, the VM receives the amount of memory defined by this attribute.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"reservation_guarantee": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the reserved amount of memory that is configured for a VM. The value of the attribute ranges between 0 and one. Value of 0 memory reservation guarantee defines no memory reservation. Value of 1 defines 100% of memory reserved.",
							ValidateFunc: IsFloatAndBetween(0, 1),
						},
						"limit_in_mb": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the memory limit in MB for a VM. If not defined in the VM sizing policy, memory limit is equal to the allocated memory for the VM.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
						"shares": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Description:  "Defines the number of memory shares for a VM. Shares specify the relative importance of a VM within a virtual data center. If a VM has twice as many shares of memory as another VM, it is entitled to consume twice as much memory when these two virtual machines are competing for resources. If not defined in the VDC compute policy, normal shares are applied to the VM.",
							ValidateFunc: IsIntAndAtLeast(0),
						},
					},
				},
			},
		},
	}
}

func resourceVmSizingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy creation initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}
	params, err := getVmZingingPolicyInput(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating VM sizing policy: %#v", params)

	createdVmSizingPolicy, err := vcdClient.Client.CreateVdcComputePolicy(params)
	if err != nil {
		log.Printf("[DEBUG] Error VM sizing policy: %s", err)
		return diag.Errorf("error VM sizing policy: %s", err)
	}

	d.SetId(createdVmSizingPolicy.VdcComputePolicy.ID)
	log.Printf("[TRACE] VM sizing policy created: %#v", createdVmSizingPolicy.VdcComputePolicy)

	return resourceVmSizingPolicyRead(ctx, d, meta)
}

// resourceVcdVmAffinityRuleRead reads a resource VM affinity rule
func resourceVmSizingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmSizingPolicyRead(ctx, d, meta)
}

// Fetches information about an existing VM sizing policy for a data definition
func genericVcdVmSizingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy read initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	// The method variable stores the information about how we found the rule, for logging purposes
	method := "id"

	var policy *govcd.VdcComputePolicy
	var err error
	if d.Id() != "" {
		policy, err = vcdClient.Client.GetVdcComputePolicyById(d.Id())
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM sizing policy %s, err: %s. Removing from tfstate", policyName, err)
		}
	}

	// The secondary method of retrieval is from name
	if d.Id() == "" {
		if policyName == "" {
			return diag.Errorf("both name and ID are empty")
		}
		method = "name"
		queryParams := url.Values{}
		queryParams.Add("filter", "name=="+policyName)
		filteredPoliciesByName, err := vcdClient.Client.GetAllVdcComputePolicies(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate.", policyName)
			d.SetId("")
			return diag.Errorf("unable to find VM sizing policy %s, err: %s. Removing from tfstate", policyName, err)
		}
		if len(filteredPoliciesByName) != 1 {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s . Found Policies by name: %d. Removing from tfstate.", policyName, len(filteredPoliciesByName))
			d.SetId("")
			return diag.Errorf("[DEBUG] Unable to find VM sizing policy %s, err: %s. Found Policies by name: %d. Removing from tfstate", policyName, govcd.ErrorEntityNotFound, len(filteredPoliciesByName))
		}
		policy = filteredPoliciesByName[0]
		d.SetId(policy.VdcComputePolicy.ID)
	}

	// Fix coverity warning
	if policy == nil {
		return diag.Errorf("[genericVcdVmSizingPolicyRead] error defining sizing policy")
	}
	util.Logger.Printf("[TRACE] [get VM sizing policy] Retrieved by %s\n", method)
	return setVmSizingPolicy(ctx, d, *policy.VdcComputePolicy)
}

// setVmSizingPolicy sets object state from *govcd.VdcComputePolicy
func setVmSizingPolicy(ctx context.Context, d *schema.ResourceData, policy types.VdcComputePolicy) diag.Diagnostics {

	dSet(d, "name", policy.Name)
	dSet(d, "description", policy.Description)

	var cpuList []map[string]interface{}
	cpuMap := make(map[string]interface{})

	if policy.CPUShares != nil {
		cpuMap["shares"] = strconv.Itoa(*policy.CPUShares)
	}
	cpuFieldProvided := false
	if policy.CPUShares != nil {
		cpuMap["shares"] = strconv.Itoa(*policy.CPUShares)
		cpuFieldProvided = true
	}
	if policy.CPULimit != nil {
		cpuMap["limit_in_mhz"] = strconv.Itoa(*policy.CPULimit)
		cpuFieldProvided = true
	}
	if policy.CPUCount != nil {
		cpuMap["count"] = strconv.Itoa(*policy.CPUCount)
		cpuFieldProvided = true
	}
	if policy.CPUSpeed != nil {
		cpuMap["speed_in_mhz"] = strconv.Itoa(*policy.CPUSpeed)
		cpuFieldProvided = true
	}
	if policy.CoresPerSocket != nil {
		cpuMap["cores_per_socket"] = strconv.Itoa(*policy.CoresPerSocket)
		cpuFieldProvided = true
	}
	if policy.CPUReservationGuarantee != nil {
		cpuMap["reservation_guarantee"] = strconv.FormatFloat(*policy.CPUReservationGuarantee, 'f', -1, 64)
		cpuFieldProvided = true
	}
	if cpuFieldProvided {
		cpuList = append(cpuList, cpuMap)
		err := d.Set("cpu", cpuList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	var memoryList []map[string]interface{}
	memoryMap := make(map[string]interface{})
	memoryFieldProvided := false
	if policy.Memory != nil {
		memoryMap["size_in_mb"] = strconv.Itoa(*policy.Memory)
		memoryFieldProvided = true
	}
	if policy.MemoryLimit != nil {
		memoryMap["limit_in_mb"] = strconv.Itoa(*policy.MemoryLimit)
		memoryFieldProvided = true
	}
	if policy.MemoryShares != nil {
		memoryMap["shares"] = strconv.Itoa(*policy.MemoryShares)
		memoryFieldProvided = true
	}
	if policy.MemoryReservationGuarantee != nil {
		memoryMap["reservation_guarantee"] = strconv.FormatFloat(*policy.MemoryReservationGuarantee, 'f', -1, 64)
		memoryFieldProvided = true
	}
	if memoryFieldProvided {
		memoryList = append(memoryList, memoryMap)
		err := d.Set("memory", memoryList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] VM sizing policy read completed: %s", policy.Name)
	return nil
}

//resourceVmSizingPolicyUpdate function updates resource with found configurations changes
func resourceVmSizingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy update initiated: %s", policyName)

	vcdClient := meta.(*VCDClient)

	policy, err := vcdClient.Client.GetVdcComputePolicyById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM sizing policy %s", policyName)
		return diag.Errorf("unable to find VM sizing policy %s, error:  %s", policyName, err)
	}

	changedPolicy, err := getUpdatedVmSizingPolicyInput(d, policy)
	if err != nil {
		log.Printf("[DEBUG] Error updating VM sizing policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM sizing policy %s, err: %s", policyName, err)
	}

	_, err = changedPolicy.Update()
	if err != nil {
		log.Printf("[DEBUG] Error updating VM sizing policy %s with error %s", policyName, err)
		return diag.Errorf("error updating VM sizing policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM sizing policy update completed: %s", policyName)
	return resourceVmSizingPolicyRead(ctx, d, meta)
}

// Deletes a VM sizing policy
func resourceVmSizingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyName := d.Get("name").(string)
	log.Printf("[TRACE] VM sizing policy delete started: %s", policyName)

	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("functionality requires System administrator privileges")
	}

	policy, err := vcdClient.Client.GetVdcComputePolicyById(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Unable to find VM sizing policy %s. Removing from tfstate", policyName)
		d.SetId("")
		return nil
	}

	err = policy.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing VM sizing policy %s, err: %s", policyName, err)
		return diag.Errorf("error removing VM sizing policy %s, err: %s", policyName, err)
	}

	log.Printf("[TRACE] VM sizing policy delete completed: %s", policyName)
	return nil
}

// helper for transforming the resource input into the VdcComputePolicy structure
func getUpdatedVmSizingPolicyInput(d *schema.ResourceData, policy *govcd.VdcComputePolicy) (*govcd.VdcComputePolicy, error) {
	if d.HasChange("name") {
		policy.VdcComputePolicy.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		policy.VdcComputePolicy.Description = d.Get("description").(string)
	}

	if d.HasChange("cpu") {
		return nil, fmt.Errorf("only name and description are updatable for VM sizing policy")
	}
	if d.HasChange("memory") {
		return nil, fmt.Errorf("only name and description are updatable for VM sizing policy")
	}

	return policy, nil
}

// helper for transforming the resource input into the VdcComputePolicy structure
// any cast operations or default values should be done here so that the create method is simple
func getVmZingingPolicyInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcComputePolicy, error) {

	params := &types.VdcComputePolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	cpuPart := d.Get("cpu").([]interface{})
	if len(cpuPart) == 1 {
		var err error
		params, err = getCpuInput(cpuPart, params)
		if err != nil {
			return nil, err
		}
	}

	memoryPart := d.Get("memory").([]interface{})
	if len(memoryPart) == 1 {
		var err error
		params, err = getMemoryInput(memoryPart, params)
		if err != nil {
			return nil, err
		}
	}
	return params, nil
}

func getCpuInput(cpuPart []interface{}, params *types.VdcComputePolicy) (*types.VdcComputePolicy, error) {
	cpuMap := cpuPart[0].(map[string]interface{})

	speedInMhz := cpuMap["speed_in_mhz"].(string)
	if speedInMhz != "" {
		convertedNumber, err := strconv.Atoi(speedInMhz)
		if err != nil {
			return nil, fmt.Errorf("value `%s` speed_in_mhz is not number. err: %s", speedInMhz, err)
		}
		params.CPUSpeed = &convertedNumber
	}
	limitInMhz := cpuMap["limit_in_mhz"].(string)
	if limitInMhz != "" {
		convertedNumber, err := strconv.Atoi(limitInMhz)
		if err != nil {
			return nil, fmt.Errorf("value `%s` limit_in_mhz is not number. err: %s", limitInMhz, err)
		}
		params.CPULimit = &convertedNumber
	}
	shares := cpuMap["shares"].(string)
	if shares != "" {
		convertedNumber, err := strconv.Atoi(shares)
		if err != nil {
			return nil, fmt.Errorf("value `%s` shares is not number. err: %s", shares, err)
		}
		params.CPUShares = &convertedNumber
	}
	count := cpuMap["count"].(string)
	if count != "" {
		convertedNumber, err := strconv.Atoi(count)
		if err != nil {
			return nil, fmt.Errorf("value `%s` count is not number. err: %s", count, err)
		}
		params.CPUCount = &convertedNumber
	}
	coresPerSocket := cpuMap["cores_per_socket"].(string)
	if coresPerSocket != "" {
		convertedNumber, err := strconv.Atoi(coresPerSocket)
		if err != nil {
			return nil, fmt.Errorf("value `%s` cores_per_socket is not number. err: %s", coresPerSocket, err)
		}
		params.CoresPerSocket = &convertedNumber
	}
	cpuReservation := cpuMap["reservation_guarantee"].(string)
	if cpuReservation != "" {
		convertedNumber, err := strconv.ParseFloat(cpuReservation, 64)
		if err != nil {
			return nil, fmt.Errorf("value `%s` reservation_guarantee is not number. err: %s", cpuReservation, err)
		}
		params.CPUReservationGuarantee = &convertedNumber
	}
	return params, nil
}

func getMemoryInput(memoryPart []interface{}, params *types.VdcComputePolicy) (*types.VdcComputePolicy, error) {
	memoryMap := memoryPart[0].(map[string]interface{})
	sizeInMb := memoryMap["size_in_mb"].(string)
	if sizeInMb != "" {
		convertedNumber, err := strconv.Atoi(sizeInMb)
		if err != nil {
			return nil, fmt.Errorf("value `%s` size_in_mb is not number. err: %s", sizeInMb, err)
		}
		params.Memory = &convertedNumber
	}
	limitInMb := memoryMap["limit_in_mb"].(string)
	if limitInMb != "" {
		convertedNumber, err := strconv.Atoi(limitInMb)
		if err != nil {
			return nil, fmt.Errorf("value `%s` limit_in_mb is not number. err: %s", limitInMb, err)
		}
		params.MemoryLimit = &convertedNumber
	}
	shares := memoryMap["shares"].(string)
	if shares != "" {
		convertedNumber, err := strconv.Atoi(shares)
		if err != nil {
			return nil, fmt.Errorf("value `%s` shares is not number. err: %s", shares, err)
		}
		params.MemoryShares = &convertedNumber
	}
	memoryReservation := memoryMap["reservation_guarantee"].(string)
	if memoryReservation != "" {
		convertedNumber, err := strconv.ParseFloat(memoryReservation, 64)
		if err != nil {
			return nil, fmt.Errorf("value `%s` reservation_guarantee is not number. err: %s", memoryReservation, err)
		}
		params.MemoryReservationGuarantee = &convertedNumber
	}
	return params, nil
}

var errHelpVmSizingPolicyImport = fmt.Errorf(`resource id must be specified in one of these formats:
'vm-sizing-policy-name', 'vm-sizing-policy-id' or 'list@' to get a list of VM sizing policies with their IDs`)

// resourceVmSizingPolicyImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vm_sizing_policy.my_existing_policy_name
// Example import path (_the_id_string_): my_existing_vm_sizing_policy_id
// Example list path (_the_id_string_): list@
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVmSizingPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing VM sizing policy resource with provided id %s", d.Id())

	if len(resourceURI) != 1 {
		return nil, errHelpVmSizingPolicyImport
	}
	if strings.Contains(d.Id(), "list@") {

		return listVmSizingPoliciesForImport(meta)
	} else {
		policyId := resourceURI[0]
		return getVmSizingPolicy(d, meta, policyId)
	}
}

func getVmSizingPolicy(d *schema.ResourceData, meta interface{}, policyId string) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	var vmSizingPolicy *govcd.VdcComputePolicy
	var err error
	vmSizingPolicy, err = vcdClient.Client.GetVdcComputePolicyById(policyId)
	if err != nil {
		queryParams := url.Values{}
		queryParams.Add("filter", "name=="+policyId)
		vmSizingPolicies, err := vcdClient.Client.GetAllVdcComputePolicies(queryParams)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VM sizing policy %s", policyId)
			return nil, fmt.Errorf("unable to find VM sizing policy %s, err: %s", policyId, err)
		}
		if len(vmSizingPolicies) != 1 {
			log.Printf("[DEBUG] Unable to find unique VM sizing policy %s", policyId)
			return nil, fmt.Errorf("unable to find unique VM sizing policy %s, err: %s", policyId, err)
		}
		vmSizingPolicy = vmSizingPolicies[0]
	}

	dSet(d, "name", vmSizingPolicy.VdcComputePolicy.Name)
	d.SetId(vmSizingPolicy.VdcComputePolicy.ID)

	return []*schema.ResourceData{d}, nil
}

func listVmSizingPoliciesForImport(meta interface{}) ([]*schema.ResourceData, error) {

	vcdClient := meta.(*VCDClient)
	var err error

	buf := new(bytes.Buffer)
	_, err = fmt.Fprintln(buf, "Retrieving all VM sizing policies")
	if err != nil {
		logForScreen("vcd_vm_sizing_policy", fmt.Sprintf("error writing to buffer: %s", err))
	}
	policies, err := vcdClient.Client.GetAllVdcComputePolicies(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve VM sizing policies: %s", err)
	}

	writer := tabwriter.NewWriter(buf, 0, 8, 1, '\t', tabwriter.AlignRight)

	_, err = fmt.Fprintln(writer, "No\tID\tName\t")
	if err != nil {
		logForScreen("vcd_vm_sizing_policy", fmt.Sprintf("error writing to buffer: %s", err))
	}
	_, err = fmt.Fprintln(writer, "--\t--\t----\t")
	if err != nil {
		logForScreen("vcd_vm_sizing_policy", fmt.Sprintf("error writing to buffer: %s", err))
	}

	for index, policy := range policies {
		// If we don't skip the auto generated policies, we also get in the list the ones that are
		// created and assigned to a VDC by default
		if policy.VdcComputePolicy.IsAutoGenerated {
			continue
		}
		_, err = fmt.Fprintf(writer, "%d\t%s\t%s \n", index+1, policy.VdcComputePolicy.ID, policy.VdcComputePolicy.Name)
		if err != nil {
			logForScreen("vcd_vm_sizing_policy", fmt.Sprintf("error writing to buffer: %s", err))
		}
	}
	err = writer.Flush()
	if err != nil {
		logForScreen("vcd_vm_sizing_policy", fmt.Sprintf("error flushing buffer: %s", err))
	}

	return nil, fmt.Errorf("resource was not imported! %s\n%s", errHelpVmSizingPolicyImport, buf.String())
}
