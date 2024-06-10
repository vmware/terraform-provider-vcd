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
	"strings"
)

func resourceVcdVdcTemplate() *schema.Resource {
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
			// TODO: Missing CPU, memory and so on
			"storage_profile": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "Storage profiles that the VDCs instantiated from this template will use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of VDC storage profile",
						},
						"default": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "True if this is default storage profile for this VDC. The default storage profile is used when an object that can specify a storage profile is created with no storage profile specified.",
						},
						"storage_used_in_mb": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Storage used in MB",
						},
					},
				},
			},
			"enable_fast_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If 'true', the VDCs instantiated from this template will have Fast provisioning enabled",
			},
			"thin_provisioning": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If 'true', the VDCs instantiated from this template will have Thin provisioning enabled",
			},
			"edge_gateway": {
				Type:        schema.TypeList,
				Required:    true,
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
							Required:    true,
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
						"gateway_cidr": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "CIDR of the Edge Gateway",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
						},
					},
				},
				RequiredWith: []string{"edge_gateway_static_ip_pool"},
			},
			"edge_gateway_static_ip_pool": {
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "IP ranges used for the network created with the Edge Gateway. Only required if the 'edge_gateway' block is used",
				Elem:         networkV2IpRange,
				RequiredWith: []string{"edge_gateway"},
			},
			"network_pool_id": {
				Type:        schema.TypeInt,
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
			"view_and_instantiate_org_ids": {
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
	// Checks:
	// NSX-V edge clusters are used when type=NSX_V
	// NSX-T edge clusters are used when type=NSX_V
	// edge_gateway_static_ip_pool is present if and only if the edge_gateway block is present
	edgeGatewayBindingId, servicesBindingId := "", ""
	vcdClient := meta.(*VCDClient)

	providerBlocks := d.Get("provider_vdc").(*schema.Set).List()
	pvdcs := make([]*types.VMWVdcTemplateProviderVdcSpecification, len(providerBlocks))

	for i, p := range providerBlocks {
		pvdcBlock := p.(map[string]interface{})

		// External network binding. This one is Required
		externalNetworkBinding := saveAndGetVdcTemplateBinding(d, "external_network_id", pvdcBlock["external_network_id"].(string))
		bindings := []*types.VMWVdcTemplateBinding{
			{
				Name:  externalNetworkBinding,
				Value: &types.Reference{ID: pvdcBlock["external_network_id"].(string)},
			},
		}
		// We generate the bindings for the Edge Clusters and save them into the state + prepare the API payload.
		for _, attribute := range []string{"gateway_edge_cluster_id", "services_edge_cluster_id"} {
			if ecId, ok := pvdcBlock[attribute]; ok {
				bindingId := saveAndGetVdcTemplateBinding(d, attribute, ecId.(string))
				bindings = append(bindings, &types.VMWVdcTemplateBinding{
					Name:  bindingId,
					Value: &types.Reference{ID: ecId.(string)},
				})

				// We save this binding ID for later
				if attribute == "nsxt_gateway_edge_cluster_id" {
					edgeGatewayBindingId = bindingId
				}
				if attribute == "services_edge_cluster_id" {
					servicesBindingId = bindingId
				}
			}
		}
		pvdcs[i] = &types.VMWVdcTemplateProviderVdcSpecification{
			ID:      pvdcBlock["id"].(string),
			Binding: bindings,
		}
	}
	// Gateway information
	var gateway *types.VdcTemplateSpecificationGatewayConfiguration
	if g, ok := d.GetOk("gateway"); ok && len(g.([]interface{})) > 0 {

		gatewayBlock := g.([]interface{})[0].(map[string]interface{})

		binding := saveAndGetVdcTemplateBinding(d, "gateway", "")
		gatewayIp := gatewayBlock["gateway_cidr"].(string)
		gatewayPrefix := gatewayBlock["gateway_cidr"].(string)
		gatewayNetmask := gatewayBlock["gateway_cidr"].(int)

		staticPoolBlocks := d.Get("edge_gateway_static_ip_pool").(*schema.Set).List()
		if len(staticPoolBlocks) == 0 {
			return diag.Errorf("at least one static IP pool is required when 'gateway' block is specified")
		}
		staticPools := make([]*types.IPRange, len(staticPoolBlocks))
		for i, b := range staticPoolBlocks {
			block := b.(map[string]interface{})
			staticPools[i] = &types.IPRange{
				StartAddress: block["start_address"].(string),
				EndAddress:   block["end_address"].(string),
			}
		}

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
							Gateway:            gatewayIp,
							Netmask:            gatewayPrefix,
							SubnetPrefixLength: &gatewayNetmask,
							IPRanges:           &types.IPRanges{IPRange: staticPools},
						},
					}},
					FenceMode: "natRouted",
				},
				IsShared: false,
			},
		}

		// TODO: What with multiple pvdcs?
		if edgeGatewayBindingId != "" {
			gateway.Gateway.Configuration.EdgeClusterConfiguration = &types.EdgeClusterConfiguration{PrimaryEdgeCluster: &types.Reference{HREF: edgeGatewayBindingId}}
		}
	}

	storageProfiles := []types.VdcStorageProfile{{}}

	_, err := vcdClient.CreateVdcTemplate(types.VMWVdcTemplate{
		NetworkBackingType:   "NSX_T", // The only supported network provider
		ProviderVdcReference: pvdcs,
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		TenantName:           d.Get("tenant_name").(string),
		TenantDescription:    d.Get("tenant_description").(string),
		VdcTemplateSpecification: &types.VMWVdcTemplateSpecification{
			Type:                    types.VdcTemplateFlexType,
			NicQuota:                d.Get("nic_quota").(int),
			VmQuota:                 d.Get("vm_quota").(int),
			ProvisionedNetworkQuota: d.Get("provisioned_network_quota").(int),
			GatewayConfiguration:    gateway,
			StorageProfile:          storageProfiles,
			IsElastic:               false,
			IncludeMemoryOverhead:   true,
			ThinProvision:           true,
			FastProvisioningEnabled: true,
			NetworkPoolReference: &types.Reference{
				HREF: "",
				ID:   "",
				Type: "",
				Name: "",
			},
			// TODO: This should be conditional + What happens with multiple PVDCS??
			NetworkProfileConfiguration: &types.VdcTemplateNetworkProfile{
				ServicesEdgeCluster: &types.Reference{HREF: servicesBindingId},
			},
			CpuAllocationMhz:           addrOf(0),
			CpuLimitMhzPerVcpu:         addrOf(1000),
			CpuLimitMhz:                addrOf(0),
			MemoryAllocationMB:         addrOf(0),
			MemoryLimitMb:              addrOf(0),
			CpuGuaranteedPercentage:    addrOf(20),
			MemoryGuaranteedPercentage: addrOf(20),
		},
	})
	if err != nil {
		return diag.Errorf("could not create the VDC Template: %s", err)
	}

	return resourceVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdVdcTemplateRead(ctx, d, meta)
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

func genericVcdVdcTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	dSet(d, "network_provider_type", vdcTemplate.VdcTemplate.NetworkBackingType)

	pvdcBlock := make([]interface{}, len(vdcTemplate.VdcTemplate.ProviderVdcReference))
	for i, providerVdcRef := range vdcTemplate.VdcTemplate.ProviderVdcReference {
		p := map[string]interface{}{}
		p["id"] = providerVdcRef.ID
		for _, binding := range providerVdcRef.Binding {
			// The Binding Name is the binding URN auto-generated during create/update.
			// Each Binding Value is a Reference (we only need the ID)
			if strings.Contains(binding.Value.ID, "urn:vcloud:external") {
				// We can only have one external network per PVDC, so we don't check bindings here
				p["external_network_id"] = binding.Value.ID
			}
			if strings.Contains(binding.Value.ID, "urn:vcloud:backingEdgeCluster") {
				// We have an Edge Cluster here, it can belong to several attributes:
				// nsxt_gateway_edge_cluster_id, nsxv_primary_edge_cluster_id, nsxt_services_edge_cluster_id or nsxv_secondary_edge_cluster_id
				// We review the saved "bindings" to know where the Edge cluster belongs.
				switch binding.Value.ID {
				case getVdcTemplateBinding(d, "nsxt_gateway_edge_cluster_id", binding.Name):
					p["nsxt_gateway_edge_cluster_id"] = binding.Value.ID
				case getVdcTemplateBinding(d, "nsxv_primary_edge_cluster_id", binding.Name):
					p["nsxv_primary_edge_cluster_id"] = binding.Value.ID
				case getVdcTemplateBinding(d, "nsxt_services_edge_cluster_id", binding.Name):
					p["nsxt_services_edge_cluster_id"] = binding.Value.ID
				case getVdcTemplateBinding(d, "nsxv_secondary_edge_cluster_id", binding.Name):
					p["nsxv_secondary_edge_cluster_id"] = binding.Value.ID
				default:
					return diag.Errorf("the binding ID '%s' is not saved in state, hence the provider can't know whether '%s' is a Primary/Gateway or Secondary/Services edge cluster", binding.Name, binding.Value.ID)
				}
			}
		}
		pvdcBlock[i] = p
	}
	err = d.Set("provider_vdc", pvdcBlock)
	if err != nil {
		return diag.FromErr(err)
	}

	if vdcTemplate.VdcTemplate.VdcTemplateSpecification != nil {
		dSet(d, "allocation_model", vdcTemplate.VdcTemplate.VdcTemplateSpecification.Type)
		dSet(d, "enable_fast_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.FastProvisioningEnabled)
		dSet(d, "thin_provisioning", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ThinProvision)
		dSet(d, "nics_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NicQuota)
		dSet(d, "provisioned_networks_quota", vdcTemplate.VdcTemplate.VdcTemplateSpecification.ProvisionedNetworkQuota)

		if vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference != nil {
			dSet(d, "network_pool_id", vdcTemplate.VdcTemplate.VdcTemplateSpecification.NetworkPoolReference.ID)
		}

		if len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile) > 0 {
			storageProfiles := make([]interface{}, len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile))
			for i, storageProfile := range vdcTemplate.VdcTemplate.VdcTemplateSpecification.StorageProfile {
				sp := map[string]interface{}{}
				sp["id"] = storageProfile.ID
				sp["default"] = storageProfile.Default
				sp["storage_used_in_mb"] = storageProfile.StorageUsedMB
				storageProfiles[i] = sp
			}
			err = d.Set("storage_profile", storageProfiles)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration != nil && vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway != nil && vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network != nil {
			edgeGatewayConfig := make([]interface{}, 1)
			ec := map[string]interface{}{}

			ec["name"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Name
			ec["description"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Description
			// ec["ip_allocation_count"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.
			ec["network_name"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Name
			ec["network_description"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Description
			// ec["gateway_cidr"] = vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Gateway.Configuration.

			edgeGatewayConfig[0] = ec
			err = d.Set("edge_gateway", edgeGatewayConfig)
			if err != nil {
				return diag.FromErr(err)
			}

			// Revisit
			staticIpPool := make([]interface{}, len(vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope))
			for i, ipScope := range vdcTemplate.VdcTemplate.VdcTemplateSpecification.GatewayConfiguration.Network.Configuration.IPScopes.IPScope {
				pool := map[string]interface{}{}
				pool["start_address"] = ipScope.IPRanges.IPRange[0].StartAddress
				pool["end_address"] = ipScope.IPRanges.IPRange[0].EndAddress
				edgeGatewayConfig[i] = pool
			}
			err = d.Set("edge_gateway_static_ip_pool", staticIpPool)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// dSet(d, "view_and_instantiate_org_ids", ????)
	dSet(d, "vdc_template_system_name", vdcTemplate.VdcTemplate.Name)
	dSet(d, "vdc_template_tenant_name", vdcTemplate.VdcTemplate.TenantName)
	dSet(d, "vdc_template_system_description", vdcTemplate.VdcTemplate.Description)
	dSet(d, "vdc_template_tenant_description", vdcTemplate.VdcTemplate.TenantDescription)

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
// argument (example: nsxt_gateway_edge_cluster_id) in the Terraform state, and returns the corresponding
// Binding ID that can be sent to VCD and used again to retrieve the URN on reads.
func saveAndGetVdcTemplateBinding(d *schema.ResourceData, field, urn string) string {
	bindings := d.Get("bindings").(map[string]interface{})
	id := fmt.Sprintf("urn:vcloud:binding:%s", uuid.NewString())
	bindings[fmt.Sprintf("%s_%s", field, id)] = urn
	err := d.Set("bindings", bindings)
	if err != nil {
		util.Logger.Printf("[ERROR] could not save binding with URN '%s' for attribute '%s' and ID '%s': %s", urn, field, id, err)
	}
	return id
}

// getVdcTemplateBinding recovers the URN (example: urn:vcloud:edgecluster:...) that was saved
// in Terraform state, that corresponds to the given Binding ID (urn:vcloud:binding:...). If it's
// not found, returns an empty string
func getVdcTemplateBinding(d *schema.ResourceData, field, bindingId string) string {
	bindings := d.Get("bindings").(map[string]interface{})
	urn, ok := bindings[fmt.Sprintf("%s_%s", field, bindingId)]
	if !ok {
		return ""
	}
	return urn.(string)
}
