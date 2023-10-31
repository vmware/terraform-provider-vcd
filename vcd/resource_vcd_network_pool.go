package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
)

const (
	networkProviderVcenter       = "vCenter"
	networkProviderNsxtManager   = "NSX-T manager"
	backingTypeTransportZone     = "Transport Zone"
	backingTypePortGroup         = "Port Group"
	backingTypeDistributedSwitch = "Distributed Switch"
)

var (
	// supportedNetworkPoolTypes defines the network pool that we can create and modify
	supportedNetworkPoolTypes = []string{
		types.NetworkPoolGeneveType,    // GENEVE
		types.NetworkPoolVlanType,      // VLAN
		types.NetworkPoolPortGroupType, // PORTGROUP_BACKED
	}
	// networkProviders defines the appropriate network provider for each type of network Pool
	networkProviders = map[string]string{
		types.NetworkPoolGeneveType:    networkProviderNsxtManager,
		types.NetworkPoolVlanType:      networkProviderVcenter,
		types.NetworkPoolPortGroupType: networkProviderVcenter,
		types.NetworkPoolVxlanType:     networkProviderVcenter, // NSX-V backed. Read-only
	}
)

func resourceNetworkPoolBacking(origin string) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    origin == "resource",
				Computed:    true,
				Description: "Backing name",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Backing ID",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Backing Type (one of 'Transport Zone', 'Port Group', 'Distributed Switch')",
			},
		},
	}
}

var resourceNetworkPoolVlanIdRange = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_id": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Start of the IDs range",
		},
		"end_id": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "End of the IDs range",
		},
	},
}

func resourceVcdNetworkPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkPoolCreate,
		ReadContext:   resourceNetworkPoolRead,
		UpdateContext: resourceNetworkPoolUpdate,
		DeleteContext: resourceNetworkPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkPoolImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of network pool.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Type of the network pool (one of `GENEVE`, `VLAN`, `PORTGROUP_BACKED`)",
				ValidateFunc: validation.StringInSlice(supportedNetworkPoolTypes, false),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the network pool",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the network pool",
			},
			"promiscuous_mode": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the network pool is in promiscuous mode",
			},
			"total_backings_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of backings",
			},
			"used_backings_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of used backings",
			},
			"network_provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the network provider (either vCenter or NSX-T manager)",
			},
			"network_provider_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the network provider",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of network provider",
			},
			"backing_components_use_constraint": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  types.BackingUseExplicit,
				ValidateFunc: validation.StringInSlice([]string{
					string(types.BackingUseExplicit),
					string(types.BackingUseWhenOnlyOne),
					string(types.BackingUseFirstAvailable)}, false),
				Description: "Defines how backing components are accepted",
			},
			"backing": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The components used by the network pool",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transport_zone": {
							Type:          schema.TypeList,
							Optional:      true,
							Computed:      true,
							MaxItems:      1,
							ForceNew:      true,
							Description:   "Transport Zone Backing",
							ConflictsWith: []string{"backing.0.port_group", "backing.0.distributed_switch"},
							Elem:          resourceNetworkPoolBacking("resource"),
						},
						"port_group": {
							Type:          schema.TypeSet,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							Description:   "Backing port groups",
							ConflictsWith: []string{"backing.0.distributed_switch", "backing.0.transport_zone"},
							Elem:          resourceNetworkPoolBacking("resource"),
						},
						"distributed_switch": {
							Type:          schema.TypeList,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							MaxItems:      1,
							Description:   "Backing distributed switches",
							ConflictsWith: []string{"backing.0.port_group", "backing.0.transport_zone"},
							Elem:          resourceNetworkPoolBacking("resource"),
						},
						"range_id": {
							Type:          schema.TypeList,
							Optional:      true,
							Description:   "Distributed Switch ID ranges (used with VLAN)",
							ConflictsWith: []string{"backing.0.port_group", "backing.0.transport_zone"},
							Elem:          resourceNetworkPoolVlanIdRange,
						},
					},
				},
			},
		},
	}
}

func resourceNetworkPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPoolName := d.Get("name").(string)
	networkPoolDescription := d.Get("description").(string)
	backingComponentsUseConstraint := d.Get("backing_components_use_constraint").(string)

	networkPoolType := d.Get("type").(string)
	networkPoolProviderId := d.Get("network_provider_id").(string)
	networkPoolProviderType := networkProviderVcenter
	if strings.Contains(networkPoolProviderId, "urn:vcloud:nsxtmanager") {
		networkPoolProviderType = networkProviderNsxtManager
	}

	expectedProviderType := networkProviders[networkPoolType]
	if networkPoolProviderType != expectedProviderType {
		return diag.Errorf("expected network provider type for network pool '%s' is '%s', but given '%s'",
			networkPoolType, expectedProviderType, networkPoolProviderType)
	}
	var networkPoolProvider types.OpenApiReference

	if networkPoolProviderType == networkProviderVcenter {
		vCenter, err := vcdClient.GetVCenterById(networkPoolProviderId)
		if err != nil {
			return diag.Errorf("[network pool create] error retrieving vCenter with ID '%s': %s", networkPoolProviderId, err)
		}
		networkPoolProvider.Name = vCenter.VSphereVCenter.Name
		networkPoolProvider.ID = vCenter.VSphereVCenter.VcId
	} else {
		managers, err := vcdClient.QueryNsxtManagers()
		if err != nil {
			return diag.Errorf("[network pool create] error retrieving list of NSX-T managers: %s", err)
		}
		var manager *types.QueryResultNsxtManagerRecordType

		bareId := extractUuid(networkPoolProviderId)
		for _, m := range managers {
			if bareId == extractUuid(m.HREF) {
				manager = m
				break
			}
		}
		if manager == nil {
			return diag.Errorf("[network pool create] NSX-T manager with ID '%s' not found", networkPoolProviderId)
		}
		networkPoolProvider.Name = manager.Name
		networkPoolProvider.ID = "urn:vcloud:nsxtmanager:" + extractUuid(manager.HREF)
	}

	if networkPoolProvider.ID == "" {
		return diag.Errorf("[network pool create] no suitable network provider (%s) found from ID '%s'", networkPoolProviderType, networkPoolProviderId)
	}
	backing, err := getNetworkPoolBacking(networkPoolType, d)
	if err != nil {
		return diag.Errorf("[network pool create] error fetching network pool backing data: %s", err)
	}
	if networkPoolType != types.NetworkPoolVlanType && backing != nil && len(backing.VlanIdRanges.Values) > 0 {
		return diag.Errorf("[network pool create] only network pools of type '%s' need range IDs", types.NetworkPoolVlanType)
	}
	var networkPool *govcd.NetworkPool
	switch networkPoolType {
	case types.NetworkPoolGeneveType:
		transportZoneName := ""
		if backing != nil {
			transportZoneName = backing.TransportZoneRef.Name
		}
		networkPool, err = vcdClient.CreateNetworkPoolGeneve(
			networkPoolName,
			networkPoolDescription,
			networkPoolProvider.Name,
			transportZoneName,
			types.BackingUseConstraint(backingComponentsUseConstraint))
	case types.NetworkPoolVlanType:
		var dsName string
		var ranges []types.VlanIdRange
		if backing != nil {
			for _, ds := range backing.VdsRefs {
				dsName = ds.Name
				break
			}
			ranges = append(ranges, backing.VlanIdRanges.Values...)
		}
		networkPool, err = vcdClient.CreateNetworkPoolVlan(
			networkPoolName,
			networkPoolDescription,
			networkPoolProvider.Name,
			dsName,
			ranges,
			types.BackingUseConstraint(backingComponentsUseConstraint))
	case types.NetworkPoolPortGroupType:
		var pgNames []string
		if backing != nil {
			for _, pg := range backing.PortGroupRefs {
				pgNames = append(pgNames, pg.Name)
			}
		}
		if len(pgNames) == 0 {
			if backingComponentsUseConstraint != string(types.BackingUseExplicit) {
				pgNames = append(pgNames, "")
			} else {
				return diag.Errorf("no port group names detected in 'backing' structure")
			}
		}
		networkPool, err = vcdClient.CreateNetworkPoolPortGroup(
			networkPoolName,
			networkPoolDescription,
			networkPoolProvider.Name,
			pgNames,
			types.BackingUseConstraint(backingComponentsUseConstraint))
	}

	if err != nil {
		return diag.Errorf("[network pool create] error creating network pool '%s': %s", networkPoolName, err)
	}
	d.SetId(networkPool.NetworkPool.Id)
	return resourceNetworkPoolRead(ctx, d, meta)
}

// resourceNetworkPoolUpdate updates the network pool
// The only fields that can be updated are name, description, and range IDs (for VLAN type)
// everything else is either read-only or ForceNew.
func resourceNetworkPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPoolName := d.Get("name").(string)
	networkPoolType := d.Get("type").(string)
	networkPoolDescription := d.Get("description").(string)

	networkPool, err := vcdClient.GetNetworkPoolById(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChanges("backing.0.range_id") {
		backing, err := getNetworkPoolBacking(networkPoolType, d)
		if err != nil {
			return diag.Errorf("error getting backing info: %s", err)
		}
		if networkPool.NetworkPool.PoolType != types.NetworkPoolVlanType && len(backing.VlanIdRanges.Values) > 0 {
			return diag.Errorf("[network pool update] only network pools of type '%s' need range IDs", types.NetworkPoolVlanType)
		}
		networkPool.NetworkPool.Backing.VlanIdRanges = backing.VlanIdRanges
	}

	networkPool.NetworkPool.Name = networkPoolName
	networkPool.NetworkPool.Description = networkPoolDescription
	err = networkPool.Update()
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceNetworkPoolRead(ctx, d, meta)
}

func resourceNetworkPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericNetworkPoolRead(ctx, d, meta, "resource")
}

func genericNetworkPoolRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPool, err := vcdClient.GetNetworkPoolById(d.Id())
	if err != nil {
		if origin == "datasource" {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if networkPool.NetworkPool.Name == "" {
		return diag.Errorf("[network pool read] found empty network pool")
	}
	dSet(d, "name", networkPool.NetworkPool.Name)
	dSet(d, "type", networkPool.NetworkPool.PoolType)
	dSet(d, "description", networkPool.NetworkPool.Description)
	dSet(d, "status", networkPool.NetworkPool.Status)
	dSet(d, "promiscuous_mode", networkPool.NetworkPool.PromiscuousMode)
	dSet(d, "total_backings_count", networkPool.NetworkPool.TotalBackingsCount)
	dSet(d, "used_backings_count", networkPool.NetworkPool.UsedBackingsCount)

	networkProviderType, ok := networkProviders[networkPool.NetworkPool.PoolType]
	if !ok {
		return diag.Errorf("[network pool read] no provider type found for pool type '%s'", networkPool.NetworkPool.PoolType)
	}
	dSet(d, "network_provider_type", networkProviderType)
	dSet(d, "network_provider_id", networkPool.NetworkPool.ManagingOwnerRef.ID)
	dSet(d, "network_provider_name", networkPool.NetworkPool.ManagingOwnerRef.Name)
	var backing = make(map[string]any)

	switch networkPool.NetworkPool.PoolType {
	case types.NetworkPoolGeneveType:
		backing["transport_zone"] = []any{
			map[string]any{
				"name": networkPool.NetworkPool.Backing.TransportZoneRef.Name,
				"id":   networkPool.NetworkPool.Backing.TransportZoneRef.ID,
				"type": backingTypeTransportZone,
			},
		}
	case types.NetworkPoolVlanType:
		var dSwitches []any
		for _, ds := range networkPool.NetworkPool.Backing.VdsRefs {
			dSwitches = append(dSwitches, map[string]any{
				"name": ds.Name,
				"id":   ds.ID,
				"type": backingTypeDistributedSwitch,
			})
		}
		backing["distributed_switch"] = dSwitches
		var ranges []any
		for _, r := range networkPool.NetworkPool.Backing.VlanIdRanges.Values {
			ranges = append(ranges, map[string]any{
				"start_id": r.StartId,
				"end_id":   r.EndId,
			})
		}
		backing["range_id"] = ranges
	case types.NetworkPoolPortGroupType:
		var pGroups []any
		for _, pg := range networkPool.NetworkPool.Backing.PortGroupRefs {
			pGroups = append(pGroups, map[string]any{
				"name": pg.Name,
				"id":   pg.ID,
				"type": backingTypePortGroup,
			})
		}
		backing["port_group"] = pGroups
	}
	err = d.Set("backing", []any{backing})
	if err != nil {
		return diag.Errorf("[network pool read] error setting backing : %s", err)
	}

	d.SetId(networkPool.NetworkPool.Id)

	return nil
}

func resourceNetworkPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPoolName := d.Get("name").(string)

	networkPool, err := vcdClient.GetNetworkPoolById(d.Id())
	if err != nil {
		return diag.Errorf("[network pool delete] network pool '%s' not found: %s", networkPoolName, err)
	}
	err = networkPool.Delete()
	if err != nil {
		return diag.Errorf("[network pool delete] error deleting network pool '%s': %s", networkPoolName, err)
	}
	return nil
}

func getNetworkPoolBacking(networkPoolType string, d *schema.ResourceData) (*types.NetworkPoolBacking, error) {
	rawBacking := d.Get("backing")
	if rawBacking == nil {
		return nil, nil
	}
	var backing types.NetworkPoolBacking

	rawList := rawBacking.([]any)
	if len(rawList) == 0 || rawList[0] == nil {
		return nil, nil
	}
	backingElement := rawList[0].(map[string]any)
	for name, value := range backingElement {
		switch name {
		case "transport_zone":
			tzRawList := value.([]any)
			if len(tzRawList) > 0 {
				if networkPoolType != types.NetworkPoolGeneveType {
					return nil, fmt.Errorf("transport_zone specified with invalid type %s", networkPoolType)
				}
				tzMap := tzRawList[0].(map[string]any)
				backing.TransportZoneRef.Name = tzMap["name"].(string)
			}
		case "port_group":
			pgRawList := value.(*schema.Set)
			for _, m := range pgRawList.List() {
				if networkPoolType != types.NetworkPoolPortGroupType {
					return nil, fmt.Errorf("port_group specified with invalid type %s", networkPoolType)
				}
				pgMap := m.(map[string]any)
				backing.PortGroupRefs = append(backing.PortGroupRefs, types.OpenApiReference{Name: pgMap["name"].(string)})
			}
		case "distributed_switch":
			dsRawList := value.([]any)
			for _, m := range dsRawList {
				if networkPoolType != types.NetworkPoolVlanType {
					return nil, fmt.Errorf("distributed_switch specified with invalid type %s", networkPoolType)
				}
				dsMap := m.(map[string]any)
				backing.VdsRefs = append(backing.VdsRefs, types.OpenApiReference{Name: dsMap["name"].(string)})
			}
		case "range_id":
			ridRawList := value.([]any)
			for _, m := range ridRawList {
				ridMap := m.(map[string]any)
				backing.VlanIdRanges.Values = append(backing.VlanIdRanges.Values, types.VlanIdRange{
					StartId: ridMap["start_id"].(int),
					EndId:   ridMap["end_id"].(int),
				})
			}
		}
	}
	if len(backing.VdsRefs) > 0 {
		if len(backing.VlanIdRanges.Values) == 0 {
			return nil, fmt.Errorf("[getNetworkPoolBacking] distributed_switch selected but no range IDs were indicated")
		}
	}
	// Note: an empty backing block is acceptable, as the system will try to fetch the first available backing
	return &backing, nil
}

func resourceNetworkPoolImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	identifier := d.Id()
	if identifier == "" {
		return nil, fmt.Errorf("[network pool import] no identifier given. The name or the ID of the network pool should be given")
	}

	var nPool *govcd.NetworkPool
	var err error
	if extractUuid(identifier) != "" {
		nPool, err = vcdClient.GetNetworkPoolById(identifier)
	} else {
		nPool, err = vcdClient.GetNetworkPoolByName(identifier)
	}
	if err != nil {
		return nil, fmt.Errorf("[network pool import] error retrieving network pool '%s'", identifier)
	}
	d.SetId(nPool.NetworkPool.Id)
	dSet(d, "name", nPool.NetworkPool.Name)

	return []*schema.ResourceData{d}, nil
}
