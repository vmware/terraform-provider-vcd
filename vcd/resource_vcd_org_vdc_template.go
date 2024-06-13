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
							Description: "ID of the Edge Cluster that the VDCs instantiated from this template will use with the Edge Gateway",
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
				Description:      "Allocation model that the VDCs instantiated from this template will use. Must be one of: 'AllocationVApp', 'AllocationPool', 'ReservationPool' or 'Flex'",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"AllocationVApp", "AllocationPool", "ReservationPool", "Flex"}, false)),
			},
			"compute_configuration": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The compute configuration for the VDCs instantiated from this template",
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
							Computed:         true, // Is set implicitly when allocation model is  ReservationPool
							Description:      "AllocationVApp, ReservationPool, Flex: The limit amount of CPU, in MHz, of the VDC that is instantiated from this template. Minimum is 256MHz. 0 means unlimited",
							ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IntBetween(0, 0), validation.IntAtLeast(256))),
						},
						"cpu_guaranteed": {
							Type:             schema.TypeInt,
							Optional:         true,
							Description:      "AllocationVApp, AllocationPool, Flex: The percentage of the CPU guaranteed to be available to VMs running within the VDC instantiated from this template",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
						},
						"cpu_speed": {
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
							Computed:    true, // Is set implicitly when allocation model is AllocationVApp
							Description: "Flex only: True if compute capacity can grow or shrink based on demand",
						},
						"include_vm_memory_overhead": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true, // Is set implicitly when allocation model is AllocationPool
							Description: "Flex only: True if the instantiated VDC includes memory overhead into its accounting for admission control",
						},
					},
				},
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
							Description: "Storage limit for the VDCs instantiated from this template, in Megabytes. 0 means unlimited",
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
							Description:      "Allocated IPs for the Edge Gateway",
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
							Description:      "CIDR of the Edge Gateway for the created network",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
						},
						"static_ip_pool": {
							Type:        schema.TypeSet,
							MinItems:    1,
							MaxItems:    1, // Due to a bug in VCD
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
				Description: "If set, specifies the Network pool for the instantiated VDCs. Otherwise, it is automatically chosen",
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
	vcdClient := meta.(*VCDClient)

	providerVdcBlocks := d.Get("provider_vdc").(*schema.Set).List()
	pvdcs := make([]*types.VMWVdcTemplateProviderVdcSpecification, len(providerVdcBlocks))

	// Generate unique UUIDs as bindings. Each binding "describes a section" in the Provider VDC binding section.
	// For example, "urn:vcloud:binding:foo" would describe "External network" for each Provider VDC, whereas
	// "urn:vcloud:binding:bar" would describe the "Gateway Edge cluster" for each PVDC.
	bindingNs := "urn:vcloud:binding:"
	externalNetworkBindingId := fmt.Sprintf("%s%s", bindingNs, uuid.NewString())
	edgeGatewayBindingId := fmt.Sprintf("%s%s", bindingNs, uuid.NewString())
	servicesBindingId := fmt.Sprintf("%s%s", bindingNs, uuid.NewString())
	gatewayEdgeClusterSet, servicesEdgeClusterSet := false, false
	for i, p := range providerVdcBlocks {
		pvdcBlock := p.(map[string]interface{})
		var bindings []*types.VMWVdcTemplateBinding

		if pvdcBlock["external_network_id"] != "" {
			bindings = append(bindings, &types.VMWVdcTemplateBinding{
				Name:  externalNetworkBindingId,
				Value: &types.Reference{ID: pvdcBlock["external_network_id"].(string)},
			})
		}
		if pvdcBlock["gateway_edge_cluster_id"] != "" {
			gatewayEdgeClusterSet = true
			bindings = append(bindings, &types.VMWVdcTemplateBinding{
				Name:  edgeGatewayBindingId,
				Value: &types.Reference{ID: fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", pvdcBlock["gateway_edge_cluster_id"].(string))},
			})
		}
		if pvdcBlock["services_edge_cluster_id"] != "" {
			servicesEdgeClusterSet = true
			bindings = append(bindings, &types.VMWVdcTemplateBinding{
				Name:  servicesBindingId,
				Value: &types.Reference{ID: fmt.Sprintf("urn:vcloud:backingEdgeCluster:%s", pvdcBlock["services_edge_cluster_id"].(string))},
			})
		}
		pvdcs[i] = &types.VMWVdcTemplateProviderVdcSpecification{
			ID:      pvdcBlock["id"].(string),
			Binding: bindings,
		}
	}

	// If user sets "gateway_edge_cluster_id" inside a "provider_vdc" block, but "edge_gateway" attribute is empty,
	// we should throw an error, as subsequent plans won't be able to recover the Gateway Edge cluster field and will
	// cause unwanted updates-in-place until solved.
	if _, ok := d.GetOk("edge_gateway"); !ok && gatewayEdgeClusterSet {
		return diag.Errorf("You have set a 'gateway_edge_cluster_id' attribute inside a 'provider_vdc' block, but no 'edge_gateway' block was found. " +
			"Either remove the 'gateway_edge_cluster_id' or add the 'edge_gateway' block to the configuration")
	}

	// Gateway configuration
	var gateway *types.VdcTemplateSpecificationGatewayConfiguration
	if g, ok := d.GetOk("edge_gateway"); ok && len(g.([]interface{})) > 0 {
		gatewayBlock := g.([]interface{})[0].(map[string]interface{})

		// Static pools. Schema guarantees there's at least 1
		staticPoolBlocks := gatewayBlock["static_ip_pool"].(*schema.Set).List()
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
			return diag.Errorf("could not %s VDC Template: error parsing 'network_gateway_cidr': %s", operation, err)
		}
		prefixLength, _ := cidr.Mask.Size()

		gateway = &types.VdcTemplateSpecificationGatewayConfiguration{
			Gateway: &types.EdgeGateway{
				Name:        gatewayBlock["name"].(string),
				Description: gatewayBlock["description"].(string),
				Configuration: &types.GatewayConfiguration{
					GatewayInterfaces: &types.GatewayInterfaces{GatewayInterface: []*types.GatewayInterface{
						{
							Name:        edgeGatewayBindingId,
							DisplayName: edgeGatewayBindingId,
							Connected:   true,
							Network: &types.Reference{
								HREF: edgeGatewayBindingId,
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
		if gatewayEdgeClusterSet {
			gateway.Gateway.Configuration.EdgeClusterConfiguration = &types.EdgeClusterConfiguration{PrimaryEdgeCluster: &types.Reference{HREF: edgeGatewayBindingId}}
		}
	}

	// Storage profiles
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

	// Get the compute configuration. There are too many combinations to perform "smart validations" here,
	// and given that VCD behaves well when input is incorrect, we skip them.
	// Schema guarantees that there's exactly 1 item, so we can retrieve every field it directly.
	if c, ok := d.GetOk("compute_configuration.0.elasticity"); ok {
		if allocationModel != types.VdcTemplatePayAsYouGoType && allocationModel != types.VdcTemplateFlexType {
			return diag.Errorf("could not %s the VDC Template, 'elasticity' can only be set when 'allocation_model' is AllocationVApp or Flex', but it is %s", operation, d.Get("allocation_model"))
		}
		settings.VdcTemplateSpecification.IsElastic = addrOf(c.(bool))
	}
	if c, ok := d.GetOk("compute_configuration.0.include_vm_memory_overhead"); ok {
		if allocationModel != types.VdcTemplateAllocationPoolType && allocationModel != types.VdcTemplateReservationPoolType && allocationModel != types.VdcTemplateFlexType {
			return diag.Errorf("could not %s the VDC Template, 'include_vm_memory_overhead' can only be set when 'allocation_model' is AllocationPool, ReservationPool or Flex', but it is %s", operation, d.Get("allocation_model"))
		}
		settings.VdcTemplateSpecification.IncludeMemoryOverhead = addrOf(c.(bool))
	}

	if c, ok := d.GetOk("compute_configuration.0.cpu_allocated"); ok {
		settings.VdcTemplateSpecification.CpuAllocationMhz = c.(int)
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_limit"); ok {
		settings.VdcTemplateSpecification.CpuLimitMhz = c.(int)
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_guaranteed"); ok {
		settings.VdcTemplateSpecification.CpuGuaranteedPercentage = c.(int)
	}
	if c, ok := d.GetOk("compute_configuration.0.cpu_speed"); ok {
		if allocationModel != types.VdcTemplateAllocationPoolType {
			settings.VdcTemplateSpecification.CpuLimitMhzPerVcpu = c.(int)
		} else {
			settings.VdcTemplateSpecification.VCpuInMhz = c.(int)
		}
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_allocated"); ok {
		settings.VdcTemplateSpecification.MemoryAllocationMB = c.(int)
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_limit"); ok {
		settings.VdcTemplateSpecification.MemoryLimitMb = c.(int)
	}
	if c, ok := d.GetOk("compute_configuration.0.memory_guaranteed"); ok {
		settings.VdcTemplateSpecification.MemoryGuaranteedPercentage = c.(int)
	}

	// Network pool. If not specified, we need to populate the AutomaticNetworkPoolReference field with an empty object
	// (it's not a bool for some reason).
	if networkPoolId, ok := d.GetOk("network_pool_id"); ok {
		// We need to convert ID to HREF to avoid problems in the backend.
		settings.VdcTemplateSpecification.NetworkPoolReference = &types.Reference{
			ID:   networkPoolId.(string),
			HREF: fmt.Sprintf("%s/cloudapi/%s%s%s", vcdClient.Client.VCDHREF.String(), types.OpenApiPathVersion1_0_0, types.OpenApiEndpointNetworkPools, networkPoolId),
		}
		settings.VdcTemplateSpecification.AutomaticNetworkPoolReference = nil
	} else {
		settings.VdcTemplateSpecification.NetworkPoolReference = nil
		settings.VdcTemplateSpecification.AutomaticNetworkPoolReference = &types.AutomaticNetworkPoolReference{}
	}

	if servicesEdgeClusterSet {
		settings.VdcTemplateSpecification.NetworkProfileConfiguration = &types.VdcTemplateNetworkProfile{
			ServicesEdgeCluster: &types.Reference{HREF: servicesBindingId},
		}
	}

	// The create and update operations are almost identical
	var err error
	var vdcTemplate *govcd.VdcTemplate
	switch operation {
	case "create":
		vdcTemplate, err = vcdClient.CreateVdcTemplate(settings)
	case "update":
		vdcTemplate, err = vcdClient.GetVdcTemplateById(d.Id())
		if err != nil {
			return diag.Errorf("could not retrieve the VDC Template to update it: %s", err)
		}
		vdcTemplate, err = vdcTemplate.Update(settings)
	default:
		return diag.Errorf("the operation '%s' is not supported for VDC templates", operation)
	}
	if err != nil {
		return diag.Errorf("could not %s the VDC Template: %s", operation, err)
	}

	if vdcTemplate != nil {
		orgs := d.Get("readable_by_org_ids").(*schema.Set)
		if len(orgs.List()) > 0 {
			err = vdcTemplate.SetAccess(convertSchemaSetToSliceOfStrings(orgs))
			if err != nil {
				return diag.Errorf("could not %s VDC Template, setting access list failed: %s", operation, err)
			}
		}
	}

	return resourceVcdVdcTemplateRead(ctx, d, meta)
}

func genericVcdVdcTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		return diag.FromErr(err)
	}
	if vdcTemplate.VdcTemplate.NetworkBackingType != "NSX_T" {
		return diag.Errorf("could not read VDC Template with ID '%s', '%s' network provider is not supported", vdcTemplate.VdcTemplate.ID, vdcTemplate.VdcTemplate.NetworkBackingType)
	}
	if vdcTemplate.VdcTemplate.VdcTemplateSpecification == nil {
		return diag.Errorf("could not read VDC Template with ID '%s', its specification is nil", vdcTemplate.VdcTemplate.ID)
	}

	// Retrieve the Binding name for the Services Edge cluster, if present. This will be used later, when reading Provider
	// VDC information, to put the correct data into the Terraform state.
	servicesEdgeClusterBindingName := ""
	if vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkProfileConfiguration != nil &&
		vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkProfileConfiguration.ServicesEdgeCluster != nil {
		servicesEdgeClusterBindingName = vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkProfileConfiguration.ServicesEdgeCluster.HREF
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	dSet(d, "description", vdcTemplate.VdcTemplate.Description)
	dSet(d, "tenant_name", vdcTemplate.VdcTemplate.TenantName)
	dSet(d, "tenant_description", vdcTemplate.VdcTemplate.TenantDescription)
	dSet(d, "allocation_model", getVdcTemplateType(vdcTemplate.VdcTemplate.VdcTemplateSpecification.Type))
	dSet(d, "enable_fast_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.FastProvisioningEnabled)
	dSet(d, "thin_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ThinProvision)
	dSet(d, "nic_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NicQuota)
	dSet(d, "vm_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.VmQuota)
	dSet(d, "provisioned_network_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ProvisionedNetworkQuota)

	// Compute block
	compute := map[string]interface{}{
		"cpu_allocated":              vdcTemplate.VdcTemplate.VdcTemplateSpecification.CpuAllocationMhz,
		"cpu_limit":                  vdcTemplate.VdcTemplate.VdcTemplateSpecification.CpuLimitMhz,
		"cpu_guaranteed":             vdcTemplate.VdcTemplate.VdcTemplateSpecification.CpuGuaranteedPercentage,
		"memory_allocated":           vdcTemplate.VdcTemplate.VdcTemplateSpecification.MemoryAllocationMB,
		"memory_limit":               vdcTemplate.VdcTemplate.VdcTemplateSpecification.MemoryLimitMb,
		"memory_guaranteed":          vdcTemplate.VdcTemplate.VdcTemplateSpecification.MemoryGuaranteedPercentage,
		"elasticity":                 vdcTemplate.VdcTemplate.VdcTemplateSpecification.IsElastic != nil && *vdcTemplate.VdcTemplate.VdcTemplateSpecification.IsElastic,
		"include_vm_memory_overhead": vdcTemplate.VdcTemplate.VdcTemplateSpecification.IncludeMemoryOverhead != nil && *vdcTemplate.VdcTemplate.VdcTemplateSpecification.IncludeMemoryOverhead,
	}
	if vdcTemplate.VdcTemplate.VdcTemplateSpecification.Type != types.VdcTemplateAllocationPoolType {
		compute["cpu_speed"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.CpuLimitMhzPerVcpu
	} else {
		compute["cpu_speed"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.VCpuInMhz
	}
	err = d.Set("compute_configuration", []interface{}{compute})
	if err != nil {
		return diag.FromErr(err)
	}

	// Network pool, optional
	if vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference != nil {
		dSet(d, "network_pool_id", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference.ID)
	}

	// Storage profiles
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

	// Edge gateway configuration, optional.
	gatewayConfiguration := vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration
	var edgeGatewayConfig []interface{}
	edgeGatewayEdgeClusterBindingName := ""
	if gatewayConfiguration != nil && gatewayConfiguration.Gateway != nil && gatewayConfiguration.Network != nil &&
		gatewayConfiguration.Gateway.Configuration != nil && gatewayConfiguration.Gateway.Configuration.GatewayInterfaces != nil &&
		len(gatewayConfiguration.Gateway.Configuration.GatewayInterfaces.GatewayInterface) > 0 &&
		gatewayConfiguration.Network.Configuration != nil && gatewayConfiguration.Network.Configuration.IPScopes != nil &&
		len(gatewayConfiguration.Network.Configuration.IPScopes.IPScope) > 0 {

		ec := map[string]interface{}{}
		ec["name"] = gatewayConfiguration.Gateway.Name
		ec["description"] = gatewayConfiguration.Gateway.Description
		ec["ip_allocation_count"] = gatewayConfiguration.Gateway.Configuration.GatewayInterfaces.GatewayInterface[0].QuickAddAllocatedIpCount
		ec["network_name"] = gatewayConfiguration.Network.Name
		ec["network_description"] = gatewayConfiguration.Network.Description
		if gatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].SubnetPrefixLength != nil {
			ec["network_gateway_cidr"] = fmt.Sprintf("%s/%d", gatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].Gateway, *vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].SubnetPrefixLength)
		}
		if gatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges != nil {
			ipRanges := make([]interface{}, len(gatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges.IPRange))
			for i, ir := range gatewayConfiguration.Network.Configuration.IPScopes.IPScope[0].IPRanges.IPRange {
				ipRange := map[string]interface{}{}
				ipRange["start_address"] = ir.StartAddress
				ipRange["end_address"] = ir.EndAddress
				ipRanges[i] = ipRange
			}
			ec["static_ip_pool"] = ipRanges
		}

		// Retrieve the Binding name for the Gateway Edge cluster, if present. This will be used later, when reading Provider
		// VDC information, to put the correct data into the Terraform state.
		if gatewayConfiguration.Gateway.Configuration.EdgeClusterConfiguration != nil &&
			gatewayConfiguration.Gateway.Configuration.EdgeClusterConfiguration.PrimaryEdgeCluster != nil {
			edgeGatewayEdgeClusterBindingName = gatewayConfiguration.Gateway.Configuration.EdgeClusterConfiguration.PrimaryEdgeCluster.HREF
		}

		edgeGatewayConfig = append(edgeGatewayConfig, ec)
	}
	err = d.Set("edge_gateway", edgeGatewayConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	// Provider VDC information must be handled at the end, because we need the Binding names for the
	// Gateway and Services edge clusters, which are obtained above.
	pvdcBlock := make([]interface{}, len(vdcTemplate.VdcTemplate.ProviderVdcReference))
	for i, providerVdcRef := range vdcTemplate.VdcTemplate.ProviderVdcReference {
		p := map[string]interface{}{}
		pvdcId, err := govcd.GetUuidFromHref(providerVdcRef.HREF, true)
		if err != nil {
			return diag.Errorf("error reading VDC Template: %s", err)
		}
		p["id"] = fmt.Sprintf("urn:vcloud:providervdc:%s", pvdcId)
		for _, binding := range providerVdcRef.Binding {
			if binding.Value == nil {
				continue
			}
			switch binding.Name {
			// We do the ReplaceAll as the UUIDs from vcd_nsxt_edge_cluster (or other API calls) come without any namespace, but the VDC Template uses
			// "urn:vcloud:backingEdgeCluster:". If we don't remove them, Terraform would ask for a replacement of the whole block in subsequent plans.
			case edgeGatewayEdgeClusterBindingName:
				p["gateway_edge_cluster_id"] = strings.ReplaceAll(binding.Value.ID, "urn:vcloud:backingEdgeCluster:", "")
			case servicesEdgeClusterBindingName:
				p["services_edge_cluster_id"] = strings.ReplaceAll(binding.Value.ID, "urn:vcloud:backingEdgeCluster:", "")
			default:
				if strings.HasPrefix(binding.Value.ID, "urn:vcloud:network:") {
					p["external_network_id"] = binding.Value.ID
				}
			}
		}
		pvdcBlock[i] = p
	}
	err = d.Set("provider_vdc", pvdcBlock)
	if err != nil {
		return diag.FromErr(err)
	}

	// Access settings is only available for System administrators
	if meta.(*VCDClient).Client.IsSysAdmin {
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
