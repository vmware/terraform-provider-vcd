package vcd

import (
	"context"
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
			"backing": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "The components used by the network pool",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transport_zone": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Transport Zone Backing",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"port_groups": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Backing port groups",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"distributed_switches": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Backing distributed switches",
							Elem:        resourceNetworkPoolBacking("resource"),
						},
						"range_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Distributed Switch ID ranges (used with VLAN)",
							Elem:        resourceNetworkPoolVlanIdRange,
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

	networkPoolType := d.Get("type").(string)
	networkPoolProviderId := d.Get("network_provider_id").(string)
	networkPoolProviderType := networkProviderVcenter
	if strings.Contains(networkPoolProviderId, "urn:vcloud:nsxtmanager") {
		networkPoolProviderType = networkProviderNsxtManager
	}

	var networkPoolProvider types.OpenApiReference

	if networkPoolProviderType == networkProviderVcenter {
		vCenter, err := vcdClient.GetVCenterById(networkPoolProviderId)
		if err != nil {
			return diag.Errorf("error retrieving vCenter with ID '%s': %s", networkPoolProviderId, err)
		}
		networkPoolProvider.Name = vCenter.VSphereVCenter.Name
		networkPoolProvider.ID = vCenter.VSphereVCenter.VcId
	} else {
		managers, err := vcdClient.QueryNsxtManagers()
		if err != nil {
			return diag.Errorf("error retrieving list of NSX-T managers: %s", err)
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
			return diag.Errorf("NSX-T manager with ID '%s' not found", networkPoolProviderId)
		}
		networkPoolProvider.Name = manager.Name
		networkPoolProvider.ID = "urn:vcloud:nsxtmanager:" + extractUuid(manager.HREF)
	}

	if networkPoolProvider.ID == "" {
		return diag.Errorf("no suitable network provider (%s) found from ID '%s'", networkPoolProviderType, networkPoolProviderId)
	}
	backing, err := getNetworkPoolBacking(d)
	if err != nil {
		return diag.Errorf("error fetching network pool backing data: %s", err)
	}
	var networkPool *govcd.NetworkPool
	switch networkPoolType {
	case types.NetworkPoolGeneveType:
		//transportZoneName := d.Get("backing.0.transport_zone.0.name").(string)
		transportZoneName := ""
		if backing != nil {
			transportZoneName = backing.TransportZoneRef.Name
		}
		networkPool, err = vcdClient.CreateNetworkPoolGeneve(
			networkPoolName,
			networkPoolDescription,
			networkPoolProvider.Name,
			transportZoneName)
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
			ranges)
	case types.NetworkPoolPortGroupType:
		var pgName string
		if backing != nil {
			for _, pg := range backing.PortGroupRefs {
				pgName = pg.Name
				break
			}
		}
		networkPool, err = vcdClient.CreateNetworkPoolPortGroup(
			networkPoolName,
			networkPoolDescription,
			networkPoolProvider.Name,
			pgName)
	}

	if err != nil {
		return diag.Errorf("error creating network pool '%s': %s", networkPoolName, err)
	}
	d.SetId(networkPool.NetworkPool.Id)
	return resourceNetworkPoolRead(ctx, d, meta)
}

func resourceNetworkPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented yet")
}

func resourceNetworkPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericNetworkPoolRead(ctx, d, meta, "resource")
}

func genericNetworkPoolRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPoolName := d.Get("name").(string)

	networkPool, err := vcdClient.GetNetworkPoolByName(networkPoolName)
	if err != nil {
		if origin == "datasource" {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}

	if networkPool.NetworkPool.Name == "" {
		return diag.Errorf("found empty network pool")
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
		return diag.Errorf("no provider type found for pool type '%s'", networkPool.NetworkPool.PoolType)
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
		backing["distributed_switches"] = dSwitches
		var ranges []any
		for _, r := range networkPool.NetworkPool.Backing.VlanIdRanges.Values {
			ranges = append(ranges, map[string]any{
				"start_id": r.StartId,
				"end_id":   r.EndId,
			})
		}
		backing["range_ids"] = ranges
	case types.NetworkPoolPortGroupType:
		var pGroups []any
		for _, pg := range networkPool.NetworkPool.Backing.PortGroupRefs {
			pGroups = append(pGroups, map[string]any{
				"name": pg.Name,
				"id":   pg.ID,
				"type": backingTypePortGroup,
			})
		}
		backing["port_groups"] = pGroups
	}
	err = d.Set("backing", []any{backing})
	if err != nil {
		return diag.Errorf("error setting backing : %s", err)
	}

	d.SetId(networkPool.NetworkPool.Id)

	return nil
}

func resourceNetworkPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	networkPoolName := d.Get("name").(string)

	networkPool, err := vcdClient.GetNetworkPoolByName(networkPoolName)
	if err != nil {
		return diag.Errorf("network pool '%s' not found: %s", networkPoolName, err)
	}
	err = networkPool.Delete()
	if err != nil {
		return diag.Errorf("error deleting network pool '%s': %s", networkPoolName, err)
	}
	return nil
}

func getNetworkPoolBacking(d *schema.ResourceData) (*types.NetworkPoolBacking, error) {
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
				tzMap := tzRawList[0].(map[string]any)
				backing.TransportZoneRef.Name = tzMap["name"].(string)
			}
		case "port_group":
			pgRawList := value.([]any)
			//pgMap := value.([]map[string]any)
			for _, m := range pgRawList {
				pgMap := m.(map[string]any)
				backing.PortGroupRefs = append(backing.PortGroupRefs, types.OpenApiReference{Name: pgMap["name"].(string)})
			}
		case "distributed_switches":
			dsRawList := value.([]any)
			//dsMap := dsRawList.(map[string]any)
			for _, m := range dsRawList {
				dsMap := m.(map[string]any)
				backing.VdsRefs = append(backing.VdsRefs, types.OpenApiReference{Name: dsMap["name"].(string)})
			}
		case "range_ids":
			ridRawList := value.([]any)
			//ridMap := value.([]map[string]any)
			for _, m := range ridRawList {
				ridMap := m.(map[string]any)
				backing.VlanIdRanges.Values = append(backing.VlanIdRanges.Values, types.VlanIdRange{
					StartId: ridMap["start_id"].(int),
					EndId:   ridMap["end_id"].(int),
				})
			}
		}
	}
	return &backing, nil

}

// TODO: remove before opening PR
/*
{
  "name": "TestVCD.Test_CreateNetworkPoolGeneve",
  "description": "test network pool geneve",
  "poolType": "GENEVE",
  "managingOwnerRef": {
    "name": "nsxManager1",
    "id": "urn:vcloud:nsxtmanager:74f10a3e-0fb3-4631-b35e-e548848c64a4"
  },
  "backing": {
    "vlanIdRanges": {
      "values": null
    },
    "transportZoneRef": {
      "name": "nsx-overlay-transportzone",
      "id": "/infra/sites/default/enforcement-points/default/transport-zones/1b3a2f36-bfd1-443e-a0f6-4de01abc963e"
    },
    "providerRef": {
      "name": "nsxManager1",
      "id": "urn:vcloud:nsxtmanager:74f10a3e-0fb3-4631-b35e-e548848c64a4"
    }
  }
}

{
  "name": "TestVCD.Test_CreateNetworkPoolPortgroup",
  "description": "test network pool port group",
  "poolType": "PORTGROUP_BACKED",
  "managingOwnerRef": {
    "name": "vc1",
    "id": "urn:vcloud:vimserver:1ed6e7c0-5761-4850-9b6b-c49fb5e0bd89"
  },
  "backing": {
    "vlanIdRanges": {
      "values": null
    },
    "portGroupRefs": [
      {
        "name": "TestbedPG",
        "id": "dvportgroup-29"
      }
    ],
    "transportZoneRef": {},
    "providerRef": {
      "name": "vc1",
      "id": "urn:vcloud:vimserver:1ed6e7c0-5761-4850-9b6b-c49fb5e0bd89"
    }
  }
}

{
  "name": "TestVCD.Test_CreateNetworkPoolVlan",
  "description": "test network pool VLAN",
  "poolType": "VLAN",
  "managingOwnerRef": {
    "name": "vc1",
    "id": "urn:vcloud:vimserver:1ed6e7c0-5761-4850-9b6b-c49fb5e0bd89"
  },
  "backing": {
    "vlanIdRanges": {
      "values": [
        {
          "startId": 1,
          "endId": 100
        },
        {
          "startId": 201,
          "endId": 300
        }
      ]
    },
    "vdsRefs": [
      {
        "name": "TestbedDVS",
        "id": "dvs-27"
      }
    ],
    "transportZoneRef": {},
    "providerRef": {
      "name": "vc1",
      "id": "urn:vcloud:vimserver:1ed6e7c0-5761-4850-9b6b-c49fb5e0bd89"
    }
  }
}

*/
