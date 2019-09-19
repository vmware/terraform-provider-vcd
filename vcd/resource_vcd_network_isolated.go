package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkIsolated() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkIsolatedCreate,
		Read:   resourceVcdNetworkIsolatedRead,
		Delete: resourceVcdNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNetworkIsolatedImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "255.255.255.0",
			},
			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},

			"dns1": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "8.8.8.8",
			},

			"dns2": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "8.8.4.4",
			},

			"dns_suffix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"href": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"default_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Default:  3600,
							Optional: true,
						},

						"max_lease_time": &schema.Schema{
							Type:     schema.TypeInt,
							Default:  7200,
							Optional: true,
						},
					},
				},
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceVcdNetworkIsolatedCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	gatewayName := d.Get("gateway").(string)
	networkName := d.Get("name").(string)

	ipRanges := expandIPRange(d.Get("static_ip_pool").([]interface{}))

	dhcpPool := d.Get("dhcp_pool").([]interface{})

	var dhcpPoolService []*types.DhcpPoolService

	if len(dhcpPool) > 0 {
		for _, pool := range dhcpPool {

			//fmt.Printf("%#v\n",pool)
			poolMap := pool.(map[string]interface{})

			var poolService types.DhcpPoolService

			poolService.IsEnabled = true
			poolService.DefaultLeaseTime = poolMap["default_lease_time"].(int)
			poolService.MaxLeaseTime = poolMap["max_lease_time"].(int)
			poolService.LowIPAddress = poolMap["start_address"].(string)
			poolService.HighIPAddress = poolMap["end_address"].(string)
			dhcpPoolService = append(dhcpPoolService, &poolService)
		}
	}

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  networkName,
		Configuration: &types.NetworkConfiguration{
			FenceMode: "isolated",
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     d.Get("netmask").(string),
					DNS1:        d.Get("dns1").(string),
					DNS2:        d.Get("dns2").(string),
					DNSSuffix:   d.Get("dns_suffix").(string),
					IPRanges:    &ipRanges,
				}},
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}
	var services *types.GatewayFeatures
	if len(dhcpPoolService) > 0 {
		services = &types.GatewayFeatures{
			GatewayDhcpService: &types.GatewayDhcpService{
				IsEnabled: true,
				Pool:      dhcpPoolService},
		}
	} else {
		services = &types.GatewayFeatures{
			GatewayDhcpService: &types.GatewayDhcpService{
				IsEnabled: false,
				Pool:      []*types.DhcpPoolService{}},
		}
	}
	orgVDCNetwork.ServiceConfig = services

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, true)
	if err != nil {
		return fmt.Errorf("error retrieving network %s after creation", networkName)
	}
	d.SetId(network.OrgVDCNetwork.ID)

	return resourceVcdNetworkIsolatedRead(d, meta)
}

func resourceVcdNetworkIsolatedRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	identifier := d.Id()

	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	network, err := vdc.GetOrgVdcNetworkByNameOrId(identifier, false)
	if err != nil {
		log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", identifier)
		return fmt.Errorf("[network isolated read] error looking for %s %s", identifier, err)
	}

	_ = d.Set("name", network.OrgVDCNetwork.Name)
	_ = d.Set("href", network.OrgVDCNetwork.HREF)
	if c := network.OrgVDCNetwork.Configuration; c != nil {
		_ = d.Set("fence_mode", c.FenceMode)
		if c.IPScopes != nil {
			_ = d.Set("gateway", c.IPScopes.IPScope[0].Gateway)
			_ = d.Set("netmask", c.IPScopes.IPScope[0].Netmask)
			_ = d.Set("dns1", c.IPScopes.IPScope[0].DNS1)
			_ = d.Set("dns2", c.IPScopes.IPScope[0].DNS2)
			_ = d.Set("dnd_suffix", c.IPScopes.IPScope[0].DNSSuffix)
		}
	}
	_ = d.Set("shared", network.OrgVDCNetwork.IsShared)

	staticIpPool := getStaticIpPool(network)
	if len(staticIpPool) > 0 {
		err := d.Set("static_ip_pool", staticIpPool)
		if err != nil {
			return fmt.Errorf("[network isolated read] %s", err)
		}
	}
	dhcpPool := getDhcpPool(network)
	if len(dhcpPool) > 0 {
		err := d.Set("dhcp_pool", dhcpPool)
		if err != nil {
			return fmt.Errorf("[network isolated read] %s", err)
		}
	}

	d.SetId(network.OrgVDCNetwork.ID)
	return nil
}

func getDhcpPool(network *govcd.OrgVDCNetwork) []StringMap {
	var dhcpPool []StringMap
	if network.OrgVDCNetwork.ServiceConfig == nil ||
		network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService == nil ||
		len(network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService.Pool) == 0 {
		return dhcpPool
	}
	for _, service := range network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService.Pool {
		if service.IsEnabled {
			dhcp := StringMap{
				"start_address":      service.LowIPAddress,
				"end_address":        service.HighIPAddress,
				"default_lease_time": service.DefaultLeaseTime,
				"max_lease_time":     service.MaxLeaseTime,
			}
			dhcpPool = append(dhcpPool, dhcp)
		}
	}

	return dhcpPool
}

// resourceVcdNetworkIsolatedImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_network_isolated.my-network
// Example import path (_the_id_string_): org.vdc.my-network
func resourceVcdNetworkIsolatedImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[network import] resource name must be specified as org.vdc.network")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[network import] unable to find VDC %s: %s ", vdcName, err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[network import] error retrieving network %s: %s", networkName, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	d.SetId(network.OrgVDCNetwork.ID)
	return []*schema.ResourceData{d}, nil
}
