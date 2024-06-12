package vcd

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net"
	"strings"
)

func resourceVcdOrgVdcTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVdcTemplateCreate,
		UpdateContext: resourceVcdVdcTemplateUpdate,
		ReadContext:   resourceVcdVdcTemplateRead,
		DeleteContext: resourceVcdVdcTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVdcTemplateImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template as seen by the System administrator",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the VDC Template as seen by the System administrator",
			},
			"tenant_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template as seen by the tenants (organizations)",
			},
			"tenant_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the VDC Template as seen by the tenants (organizations)",
			},
			"provider_vdc": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A Provider VDC that the VDCs instantiated from this template will use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of Provider VDC",
						},
						"external_network_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the External network that the VDCs instantiated from this template will use",
						},
						"gateway_edge_cluster_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the Edge Cluster that the VDCs instantiated from this template will use with the NSX-T Gateway",
						},
						"services_edge_cluster_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the Edge Cluster that the VDCs instantiated from this template will use for services",
						},
					},
				},
			},
			"allocation_model": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Allocation model that the VDCs instantiated from this template will use. Must be one of: 'AllocationVApp', 'AllocationPool', 'ReservationPool' or 'Flex'}",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}, false)),
			},
			"compute_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_allocated": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationPool, ReservationPool, Flex: The maximum amount of CPU, in MHz, available to the VMs running within the VDC that is instantiated from this template. Minimum is 256MHz",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
						},
						"cpu_limit": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, ReservationPool, Flex: The limit amount of CPU, in MHz, of the VDC that is instantiated from this template. Minimum is 256MHz. 0 means unlimited",
							ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IntBetween(0, 0), validation.IntAtLeast(256))),
						},
						"cpu_guaranteed": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, AllocationPool, Flex: The percentage of the CPU guaranteed to be available to VMs running within the VDC instantiated from this template",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
						},
						"cpu_frequency_limit": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, AllocationPool, Flex: Specifies the clock frequency, in MHz, for any virtual CPU that is allocated to a VM. Minimum is 256MHz",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(256)),
						},
						"memory_allocated": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationPool, ReservationPool, Flex: The maximum amount of Memory, in MB, available to the VMs running within the VDC that is instantiated from this template",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
						},
						"memory_limit": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, ReservationPool, Flex: The limit amount of Memory, in MB, of the VDC that is instantiated from this template. Minimum is 1024MB. 0 means unlimited",
							ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IntBetween(0, 0), validation.IntAtLeast(1024))),
						},
						"memory_guaranteed": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, AllocationPool, Flex: The percentage of the Memory guaranteed to be available to VMs running within the VDC instantiated from this template",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
						},
						"elasticity": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flex only: True if compute capacity can grow or shrink based on demand",
						},
						"include_vm_memory_overhead": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Flex only: True if the instantiated VDC includes memory overhead into its accounting for admission control",
						},
					},
				},
				Description: "The compute configuration for the VDCs instantiated from this template",
			},
			"storage_profile": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Storage profiles that the VDCs instantiated from this template will use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of Provider VDC storage profile to use for the VDCs instantiated from this template",
						},
						"default": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "True if this is default storage profile for the VDCs instantiated from this template",
						},
						"limit": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Storage limit for the VDCs instantiated from this template, in Megabytes",
						},
					},
				},
			},
			"enable_fast_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If 'true', the VDCs instantiated from this template will have Fast provisioning enabled",
			},
			"thin_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If 'true', the VDCs instantiated from this template will have Thin provisioning enabled",
			},
			"edge_gateway": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "VDCs instantiated from this template will create a new Edge Gateway with the provided setup",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the Edge Gateway",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the Edge Gateway",
						},
						"ip_allocation_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          0,
							Description:      "Storage used in MB",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
						},
						"network_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the network to create with the Edge Gateway",
						},
						"network_description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the network to create with the Edge Gateway",
						},
						"network_gateway_cidr": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "CIDR of the Edge Gateway for the network created with the Edge Gateway",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
						},
						"static_ip_pool": {
							Type:        schema.TypeSet,
							MinItems:    1,
							Optional:    true,
							Description: "IP ranges used for the network created with the Edge Gateway. Only required if the 'edge_gateway' block is used",
							Elem:        networkV2IpRange,
						},
					},
				},
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "If set, specifies the Network pool for the instantiated VDCs. Otherwise it is automatically chosen",
			},
			"nic_quota": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				Description:      "Quota for the NICs of the instantiated VDCs. 0 means unlimited",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"vm_quota": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				Description:      "Quota for the VMs of the instantiated VDCs. 0 means unlimited",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"provisioned_network_quota": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				Description:      "Quota for the provisioned networks of the instantiated VDCs. 0 means unlimited",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"readable_by_org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of the Organizations that will be able to view and instantiate this VDC template",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"bindings": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Auxiliary map that links a binding with its corresponding VCD URN, for internal use of the provider only",
			},
		},
	}
}

func resourceVcdVdcTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateCreateOrUpdate(ctx, d, meta, "create")
}

func resourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcdVdcTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	err = vdcTemplate.Delete()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdVdcTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vdcTemplate, err := meta.(*VCDClient).GetVdcTemplateByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("could not import VDC Template with name %s: %s", d.Id(), err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	d.SetId(vdcTemplate.VdcTemplate.ID)
	return []*schema.ResourceData{d}, nil
}

func genericVcdVdcTemplateCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	var diags diag.Diagnostics
	edgeGatewayBindingId, servicesBindingId := "", ""
	vcdClient := meta.(*VCDClient)

	providerVdcBlocks := d.Get("provider_vdc").(*schema.Set).List()
	pvdcs := make([]*types.VMWVdcTemplateProviderVdcSpecification, len(providerVdcBlocks))

	for i, p := range providerVdcBlocks {
		pvdcBlock := p.(map[string]interface{})
		var bindings []*types.VMWVdcTemplateBinding

		for _, attribute := range []string{"external_network_id", "gateway_edge_cluster_id", "services_edge_cluster_id"} {
			if urn := pvdcBlock[attribute]; urn != "" {
				// We save the Binding IDs in the Terraform state to use them later
				bindingId := saveAndGetVdcTemplateBinding(d, attribute, urn.(string))
				bindings = append(bindings, &types.VMWVdcTemplateBinding{
					Name:  bindingId,
					Value: &types.Reference{ID: urn.(string)},
				})

				// We save the first available Edge cluster binding IDs for later
				if attribute == "gateway_edge_cluster_id" && edgeGatewayBindingId == "" {
					edgeGatewayBindingId = bindingId
				}
				if attribute == "services_edge_cluster_id" && servicesBindingId == "" {
					servicesBindingId = bindingId
				}
			}
		}
		pvdcs[i] = &types.VMWVdcTemplateProviderVdcSpecification{
			ID:      pvdcBlock["id"].(string),
			Binding: bindings,
		}
	}

	// If user sets "gateway_edge_cluster_id" inside a "provider_vdc" block, but "edge_gateway" attribute is empty,
	// we should warn the user
	if _, ok := d.GetOk("edge_gateway"); ok && edgeGatewayBindingId != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You have set a 'gateway_edge_cluster_id' attribute inside a 'provider_vdc' block, but it will not be applied if you do not set the 'edge_gateway' block",
			Detail:   "Missing 'edge_gateway' for the Edge cluster",
		})
	}

	// Gateway configuration
	var gateway *types.VdcTemplateSpecificationGatewayConfiguration
	if g, ok := d.GetOk("edge_gateway"); ok && len(g.([]interface{})) > 0 {
		gatewayBlock := g.([]interface{})[0].(map[string]interface{})

		// Static pools. Schema guarantees there's at least 1
		staticPoolBlocks := gatewayBlock["static_ip_pool"].([]interface{})
		staticPools := make([]*types.IPRange, len(staticPoolBlocks))
		for i, b := range staticPoolBlocks {
			block := b.(map[string]interface{})
			staticPools[i] = &types.IPRange{
				StartAddress: block["start_address"].(string),
				EndAddress:   block["end_address"].(string),
			}
		}

		ip, cidr, err := net.ParseCIDR(gatewayBlock["network_gateway_cidr"].(string))
		if err != nil {
			return append(diags, diag.Errorf("could not %s VDC Template: error parsing 'network_gateway_cidr': %s", operation, err)...)
		}
		prefixLength, _ := cidr.Mask.Size()

		// The Edge gateway doesn't have an ID as it does not exist yet, it will only exist when
		// the VDC template is instantiated. We make up a URN.
		binding := saveAndGetVdcTemplateBinding(d, "edge_gateway", "edge_gateway")

		gateway = &types.VdcTemplateSpecificationGatewayConfiguration{
			Gateway: &types.EdgeGateway{
				Name:        gatewayBlock["name"].(string),
				Description: gatewayBlock["description"].(string),
				Configuration: &types.GatewayConfiguration{
					GatewayInterfaces: &types.GatewayInterfaces{GatewayInterface: []*types.GatewayInterface{
						{
							Name:        binding,
							DisplayName: binding,
							Connected:   true,
							Network: &types.Reference{
								HREF: binding,
							},
							QuickAddAllocatedIpCount: gatewayBlock["ip_allocation_count"].(int),
						},
					}},
				},
			},
			Network: &types.OrgVDCNetwork{
				Name:        gatewayBlock["network_name"].(string),
				Description: gatewayBlock["network_description"].(string),
				Configuration: &types.NetworkConfiguration{
					IPScopes: &types.IPScopes{IPScope: []*types.IPScope{
						{
							Gateway:            ip.String(),
							Netmask:            fmt.Sprintf("%d.%d.%d.%d", cidr.Mask[0], cidr.Mask[1], cidr.Mask[2], cidr.Mask[3]),
							SubnetPrefixLength: &prefixLength,
							IPRanges:           &types.IPRanges{IPRange: staticPools},
						},
					}},

					FenceMode: "natRouted",
				},
			},
		}

		// We set the first Edge cluster available, that we saved during Binding ID generation
		if edgeGatewayBindingId != "" {
			gateway.Gateway.Configuration.EdgeClusterConfiguration = &types.EdgeClusterConfiguration{PrimaryEdgeCluster: &types.Reference{HREF: edgeGatewayBindingId}}
		}
	}

	spBlocks := d.Get("storage_profile").(*schema.Set).List()
	storageProfiles := make([]*types.VdcStorageProfile, len(spBlocks))
	for i, sp := range spBlocks {
		spBlock := sp.(map[string]interface{})
		storageProfile := &types.VdcStorageProfile{
			Name:    spBlock["name"].(string),
			Enabled: addrOf(true),
			Units:   "MB",
			Limit:   int64(spBlock["limit"].(int)),
			Default: spBlock["default"].(bool),
		}
		storageProfiles[i] = storageProfile
	}

	allocationModel := getVdcTemplateType(d.Get("allocation_model").(string))
	settings := types.VMWVdcTemplate{
		NetworkBackingType:   "NSX_T", // The only supported network provider
		ProviderVdcReference: pvdcs,
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		TenantName:           d.Get("tenant_name").(string),
		TenantDescription:    d.Get("tenant_description").(string),
		VdcTemplateSpecification: &types.VMWVdcTemplateSpecification{
			Type:                    allocationModel,
			NicQuota:                d.Get("nic_quota").(int),
			VmQuota:                 d.Get("vm_quota").(int),
			ProvisionedNetworkQuota: d.Get("provisioned_network_quota").(int),
			GatewayConfiguration:    gateway,
			StorageProfile:          storageProfiles,
			ThinProvision:           d.Get("thin_provisioning").(bool),
			FastProvisioningEnabled: d.Get("enable_fast_provisioning").(bool),
		},
	}

	// Validate and get the compute configuration.
	// Schema guarantees that there's exactly 1 item, so we can get it directly
	if allocationModel != types.VdcTemplateFlexType {
		for _, attribute := range []string{"compute_configuration.0.elasticity", "compute_configuration.0.include_vm_memory_overhead"} {
			if _, ok := d.GetOk(attribute); ok {
				return append(diags, diag.Errorf("could not %s the VDC Template, '%s' can only be set when 'allocation_model=Flex', but it is %s", operation, attribute, d.Get("allocation_model"))...)
			}
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.elasticity"); ok {
		settings.VdcTemplateSpecification.IsElastic = addrOf(c.(bool))
	}
	if c, ok := d.GetOk("compute_configuration.0.include_vm_memory_overhead"); ok {
		settings.VdcTemplateSpecification.IncludeMemoryOverhead = addrOf(c.(bool))
	}

	if c, ok := d.GetOk("compute_configuration.0.cpu_allocated"); !ok {
		if c != 0 && (allocationModel == types.VdcTemplateAllocationPoolType || allocationModel == types.VdcTemplateReservationPoolType || allocationModel == types.VdcTemplateFlexType) {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'cpu_allocated' must be set when 'allocation_model' is AllocationPool, ReservationPool or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_allocated"); ok {
		settings.VdcTemplateSpecification.CpuAllocationMhz = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.cpu_limit"); !ok {
		if c != 0 && allocationModel == types.VdcTemplatePayAsYouGoType || allocationModel == types.VdcTemplateFlexType {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'cpu_limit' must be set when 'allocation_model' is AllocationVApp or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_limit"); ok {
		settings.VdcTemplateSpecification.CpuLimitMhz = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.cpu_guaranteed"); !ok {
		if c != 0 && allocationModel != types.VdcTemplateReservationPoolType {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'cpu_guaranteed' must be set when 'allocation_model' is AllocationVApp, AllocationPool or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_guaranteed"); ok {
		settings.VdcTemplateSpecification.CpuGuaranteedPercentage = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.cpu_frequency_limit"); !ok {
		if c != 0 && (allocationModel == types.VdcTemplatePayAsYouGoType || allocationModel == types.VdcTemplateAllocationPoolType) {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'cpu_frequency_limit' must be set when 'allocation_model' is AllocationVApp, AllocationPool or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_frequency_limit"); ok {
		settings.VdcTemplateSpecification.CpuLimitMhzPerVcpu = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.memory_allocated"); !ok {
		if c != 0 && (allocationModel == types.VdcTemplateAllocationPoolType || allocationModel == types.VdcTemplateReservationPoolType || allocationModel == types.VdcTemplateFlexType) {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'memory_allocated' must be set when 'allocation_model' is AllocationPool, ReservationPool or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_allocated"); ok {
		settings.VdcTemplateSpecification.MemoryAllocationMB = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.memory_limit"); !ok {
		if c != 0 && (allocationModel == types.VdcTemplatePayAsYouGoType || allocationModel == types.VdcTemplateFlexType) {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'memory_limit' must be set when 'allocation_model' is AllocationVApp or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_limit"); ok {
		settings.VdcTemplateSpecification.MemoryLimitMb = c.(int)
	}

	if c, ok := d.GetOk("compute_configuration.0.memory_guaranteed"); !ok {
		if c != 0 && allocationModel != types.VdcTemplateReservationPoolType {
			return append(diags, diag.Errorf("could not %s the VDC Template, 'memory_guaranteed' must be set when 'allocation_model' is AllocationVApp, AllocationPool or Flex", operation)...)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_guaranteed"); ok {
		settings.VdcTemplateSpecification.MemoryGuaranteedPercentage = c.(int)
	}

	if networkPoolId, ok := d.GetOk("network_pool_id"); ok {
		// We need to convert ID to HREF to avoid problems in the backend
		settings.VdcTemplateSpecification.NetworkPoolReference = &types.Reference{
			ID:   networkPoolId.(string),
			HREF: fmt.Sprintf("%s/cloudapi/%s%s%s", vcdClient.Client.VCDHREF.String(), types.OpenApiPathVersion1_0_0, types.OpenApiEndpointNetworkPools, networkPoolId),
		}
	}

	// We set the first Edge cluster available, that we saved during Binding ID generation
	if servicesBindingId != "" {
		settings.VdcTemplateSpecification.NetworkProfileConfiguration = &types.VdcTemplateNetworkProfile{
			ServicesEdgeCluster: &types.Reference{HREF: servicesBindingId},
		}
	}

	var err error
	var vdcTemplate *govcd.VdcTemplate
	switch operation {
	case "create":
		_, err = vcdClient.CreateVdcTemplate(settings)
	case "update":

		vdcTemplate, err = vcdClient.GetVdcTemplateById(d.Id())
		if err != nil {
			return append(diags, diag.Errorf("could not retrieve the VDC Template to update it: %s", err)...)
		}
		_, err = vdcTemplate.Update(settings)
	default:
		return append(diags, diag.Errorf("the operation '%s' is not supported for VDC templates", operation)...)
	}
	if err != nil {
		return append(diags, diag.Errorf("could not %s the VDC Template: %s", operation, err)...)
	}

	if vdcTemplate != nil {
		orgs := d.Get("readable_by_org_ids").(*schema.Set)
		if len(orgs.List()) > 0 {
			err = vdcTemplate.SetAccess(convertSchemaSetToSliceOfStrings(orgs))
			if err != nil {
				return append(diags, diag.Errorf("could not %s VDC Template, setting access list failed: %s", operation, err)...)
			}
		}
	}

	return append(diags, resourceVcdVdcTemplateRead(ctx, d, meta)...)
}

func genericVcdVdcTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	dSet(d, "description", vdcTemplate.VdcTemplate.Description)
	dSet(d, "tenant_name", vdcTemplate.VdcTemplate.TenantName)
	dSet(d, "tenant_description", vdcTemplate.VdcTemplate.TenantDescription)

	pvdcBlock := make([]interface{}, len(vdcTemplate.VdcTemplate.ProviderVdcReference))
	for i, providerVdcRef := range vdcTemplate.VdcTemplate.ProviderVdcReference {
		p := map[string]interface{}{}
		pvdcId, err := govcd.GetUuidFromHref(providerVdcRef.HREF, true)
		if err != nil {
			return diag.Errorf("error reading VDC Template: %s", err)
		}
		p["id"] = fmt.Sprintf("urn:vcloud:providervdc:%s", pvdcId)
		for _, binding := range providerVdcRef.Binding {
			// The Binding Name is the binding URN auto-generated during create/update.
			// Each Binding Value is a Reference (we only need the ID)
			if strings.Contains(binding.Value.ID, "urn:vcloud:backingEdgeCluster") {
				// We have an Edge Cluster here, it can belong to several attributes:
				// gateway_edge_cluster_id, services_edge_cluster_id
				// We review the saved "bindings" to know where the Edge cluster belongs.
				switch binding.Value.ID {
				case getVdcTemplateBinding(d, "gateway_edge_cluster_id", binding.Name):
					p["gateway_edge_cluster_id"] = binding.Value.ID
				case getVdcTemplateBinding(d, "services_edge_cluster_id", binding.Name):
					p["services_edge_cluster_id"] = binding.Value.ID
				default:
					return diag.Errorf("the binding ID '%s' is not saved in state, hence the provider can't know whether '%s' is a Primary/Gateway or Secondary/Services edge cluster", binding.Name, binding.Value.ID)
				}
			} else {
				// We can only have one external network per PVDC, so we don't check bindings here
				p["external_network_id"] = binding.Value.ID
			}
		}
		pvdcBlock[i] = p
	}
	err = d.Set("provider_vdc", pvdcBlock)
	if err != nil {
		return diag.FromErr(err)
	}

	if vdcTemplate.VdcTemplate.VdcTemplateSpecification != nil {
		dSet(d, "allocation_model", getVdcTemplateType(vdcTemplate.VdcTemplate.VdcTemplateSpecification.Type))
		dSet(d, "enable_fast_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.FastProvisioningEnabled)
		dSet(d, "thin_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ThinProvision)
		dSet(d, "nic_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NicQuota)
		dSet(d, "vm_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.VmQuota)
		dSet(d, "provisioned_network_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ProvisionedNetworkQuota)

		if vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference != nil {
			dSet(d, "network_pool_id", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference.ID)
		}

		if len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile) > 0 {
			storageProfiles := make([]interface{}, len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile))
			for i, storageProfile := range vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile {
				sp := map[string]interface{}{}
				sp["name"] = storageProfile.Name
				sp["default"] = storageProfile.Default
				sp["limit"] = storageProfile.Limit
				storageProfiles[i] = sp
			}
			err = d.Set("storage_profile", storageProfiles)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration != nil &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway != nil &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network != nil &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Configuration != nil &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Configuration.GatewayInterfaces != nil &&
			len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Configuration.GatewayInterfaces.GatewayInterface) > 0 &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration != nil &&
			vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes != nil &&
			len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope) > 0 {
			edgeGatewayConfig := make([]interface{}, 1)
			ec := map[string]interface{}{}

			ec["name"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Name
			ec["description"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Description
			ec["ip_allocation_count"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Configuration.GatewayInterfaces.GatewayInterface[0].QuickAddAllocatedIpCount
			ec["network_name"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Name
			ec["network_description"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Description
			if vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].SubnetPrefixLength != nil {
				ec["network_gateway_cidr"] = fmt.Sprintf("%s/%d", vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].Gateway, *vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].SubnetPrefixLength)
			}
			if vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges != nil {
				ipRanges := make([]interface{}, len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges.IPRange))
				for i, ir := range vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges.IPRange {
					ipRange := map[string]interface{}{}
					ipRange["start_address"] = ir.StartAddress
					ipRange["end_address"] = ir.EndAddress
					ipRanges[i] = ipRange
				}
				ec["ip_ranges"] = ipRanges
			}

			edgeGatewayConfig[0] = ec
			err = d.Set("edge_gateway", edgeGatewayConfig)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	access, err := vdcTemplate.GetAccess()
	if err != nil {
		return diag.Errorf("could not read VDC Template, retrieving its setting access list failed: %s", err)
	}
	if access != nil && access.AccessSettings != nil {
		orgIds := make([]string, len(access.AccessSettings.AccessSetting))
		for i, setting := range access.AccessSettings.AccessSetting {
			if setting.Subject != nil {
				orgIds[i] = fmt.Sprintf("urn:vcloud:org:%s", extractUuid(setting.Subject.HREF))
			}
		}
		err = d.Set("readable_by_org_ids", orgIds)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(vdcTemplate.VdcTemplate.ID)

	return nil
}

// getVdcTemplate retrieves a VDC Template with the available information in the configuration.
func getVdcTemplate(d *schema.ResourceData, vcdClient *VCDClient) (*govcd.VdcTemplate, error) {
	var vdcTemplate *govcd.VdcTemplate
	var err error
	if d.Id() == "" {
		name := d.Get("name").(string)
		vdcTemplate, err = vcdClient.GetVdcTemplateByName(name)
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with name %s: %s", name, err)
		}
	} else {
		vdcTemplate, err = vcdClient.GetVdcTemplateById(d.Id())
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with ID %s: %s", d.Id(), err)
		}
	}
	return vdcTemplate, nil
}

// saveAndGetVdcTemplateBinding saves the given URN (example: urn:vcloud:edgecluster:...) of the given
// argument (example: gateway_edge_cluster_id) in the Terraform state, and returns the corresponding
// Binding ID that can be sent to VCD and used again to retrieve the URN on reads.
// If the URN is already saved, it returns the existing binding ID instead of using a new one.
func saveAndGetVdcTemplateBinding(d *schema.ResourceData, field, urn string) string {
	separator := "@@"
	bindings := d.Get("bindings").(map[string]interface{})

	// Search for an existing binding ID if the URN is already saved.
	// If the URN is already present, we reuse the same Binding ID
	for k, v := range bindings {
		if v == urn {
			return strings.Split(k, separator)[1]
		}
	}

	// Otherwise, generate a new one
	bindingId := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())
	bindings[fmt.Sprintf("%s%s%s", field, separator, bindingId)] = urn
	err := d.Set("bindings", bindings)
	if err != nil {
		util.Logger.Printf("[ERROR] could not save binding with URN '%s' for attribute '%s' and ID '%s': %s", urn, field, bindingId, err)
	}
	return bindingId
}

// getVdcTemplateBinding recovers the URN (example: urn:vcloud:edgecluster:...) that was saved
// in Terraform state, that corresponds to the given Binding ID (urn:vcloud:binding:...). If it's
// not found, returns an empty string
func getVdcTemplateBinding(d *schema.ResourceData, field, bindingId string) string {
	separator := "@@"
	bindings := d.Get("bindings").(map[string]interface{})

	urn, ok := bindings[fmt.Sprintf("%s%s%s", field, separator, bindingId)]
	if !ok {
		return ""
	}
	return urn.(string)
}

// getVdcTemplateType transforms the allocation models used by VDC Templates to the allocation models used by regular
// VDCs and viceversa.
func getVdcTemplateType(input string) string {
	switch input {
	case types.VdcTemplatePayAsYouGoType:
		return "AllocationVApp"
	case types.VdcTemplateAllocationPoolType:
		return "AllocationPool"
	case types.VdcTemplateReservationPoolType:
		return "ReservationPool"
	case types.VdcTemplateFlexType:
		return "Flex"
	case "AllocationVApp":
		return types.VdcTemplatePayAsYouGoType
	case "AllocationPool":
		return types.VdcTemplateAllocationPoolType
	case "ReservationPool":
		return types.VdcTemplateReservationPoolType
	case "Flex":
		return types.VdcTemplateFlexType
	}
	return ""
}
