package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkIsolated() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkIsolatedCreate,
		Read:   resourceVcdNetworkIsolatedRead,
		Update: resourceVcdNetworkIsolatedUpdate,
		Delete: resourceVcdNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNetworkIsolatedImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name for this network",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description for the network",
			},
			"netmask": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "255.255.255.0",
				Description:  "The netmask for the new network",
				ValidateFunc: validation.IsIPAddress,
			},
			"gateway": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The gateway for this network",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns1": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "First DNS server to use",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns2": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Second DNS server to use",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A FQDN for the virtual machines on this network",
			},

			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hyper Reference",
			},

			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},

			"dhcp_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"end_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"default_lease_time": &schema.Schema{
							Type:        schema.TypeInt,
							Default:     3600,
							Optional:    true,
							Description: "The default DHCP lease time to use",
						},

						"max_lease_time": &schema.Schema{
							Type:        schema.TypeInt,
							Default:     7200,
							Optional:    true,
							Description: "The maximum DHCP lease time to use",
						},
					},
				},
				Set: resourceVcdNetworkIsolatedDhcpPoolHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"end_address": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},
					},
				},
				Set: resourceVcdNetworkStaticIpPoolHash,
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
	netMask := d.Get("netmask").(string)
	dns1 := d.Get("dns1").(string)
	dns2 := d.Get("dns2").(string)

	ipRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	dhcpPool := d.Get("dhcp_pool").(*schema.Set).List()

	var dhcpPoolService []*types.DhcpPoolService

	if len(dhcpPool) > 0 {
		for _, pool := range dhcpPool {

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
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        networkName,
		Description: d.Get("description").(string),
		Configuration: &types.NetworkConfiguration{
			FenceMode: "isolated",
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     netMask,
					DNS1:        dns1,
					DNS2:        dns2,
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
		return fmt.Errorf("error: %s", err)
	}

	return resourceVcdNetworkIsolatedRead(d, meta)
}

func resourceVcdNetworkIsolatedRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdNetworkIsolatedRead(d, meta, "resource")
}

func genericVcdNetworkIsolatedRead(d *schema.ResourceData, meta interface{}, origin string) error {
	var network *govcd.OrgVDCNetwork
	var err error

	switch origin {
	case "resource", "datasource":
		// From the resource creation or data source, we need to retrieve the network from scratch
		vcdClient := meta.(*VCDClient)

		network, err = getNetwork(d, vcdClient, origin == "datasource", "isolated")

		if err != nil {
			if origin == "resource" {
				networkName := d.Get("name").(string)
				log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", networkName)
				d.SetId("")
				return nil
			}
			return fmt.Errorf("[network isolated read] error looking for network: %s", err)
		}
	case "resource-update":
		// From update, we get the network directly
		network = meta.(*govcd.OrgVDCNetwork)
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
			_ = d.Set("dns_suffix", c.IPScopes.IPScope[0].DNSSuffix)
		}
	}
	_ = d.Set("shared", network.OrgVDCNetwork.IsShared)

	staticIpPool := getStaticIpPool(network)
	if len(staticIpPool) > 0 {
		newSet := &schema.Set{
			F: resourceVcdNetworkStaticIpPoolHash,
		}
		for _, element := range staticIpPool {
			newSet.Add(element)
		}
		err := d.Set("static_ip_pool", newSet.List())
		if err != nil {
			return fmt.Errorf("[isolated network read] static_ip set %s", err)
		}
	}
	dhcpPool := getDhcpPool(network)
	if len(dhcpPool) > 0 {
		newSet := &schema.Set{
			F: resourceVcdNetworkIsolatedDhcpPoolHash,
		}
		for _, element := range dhcpPool {
			newSet.Add(element)
		}
		err := d.Set("dhcp_pool", newSet.List())
		if err != nil {
			return fmt.Errorf("[isolated network read] dhcp set %s", err)
		}
	}
	_ = d.Set("description", network.OrgVDCNetwork.Description)

	d.SetId(network.OrgVDCNetwork.ID)
	return nil
}

func getDhcpPool(network *govcd.OrgVDCNetwork) []map[string]interface{} {
	var dhcpPool []map[string]interface{}
	if network.OrgVDCNetwork.ServiceConfig == nil ||
		network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService == nil ||
		len(network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService.Pool) == 0 {
		return dhcpPool
	}
	for _, service := range network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService.Pool {
		if service.IsEnabled {
			dhcp := map[string]interface{}{
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

// resourceVcdNetworkIsolatedDhcpPoolHash computes a hash for a DHCP pool
func resourceVcdNetworkIsolatedDhcpPoolHash(v interface{}) int {
	return genericResourceVcdNetworkDhcpPoolHash(v, "isolated")
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
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdNetworkIsolatedImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[isolated network import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[isolated network import] unable to find VDC %s: %s ", vdcName, err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[isolated network import] error retrieving Org VDC network %s: %s", networkName, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	d.SetId(network.OrgVDCNetwork.ID)
	return []*schema.ResourceData{d}, nil
}

func resourceVcdNetworkIsolatedUpdate(d *schema.ResourceData, meta interface{}) error {
	var (
		vcdClient          = meta.(*VCDClient)
		networkName        = d.Get("name").(string)
		networkDescription = d.Get("description").(string)
		isShared           = d.Get("shared").(bool)
		dns1               = d.Get("dns1").(string)
		dns2               = d.Get("dns2").(string)
		dnsSuffix          = d.Get("dns_suffix").(string)
		dhcpPool           = d.Get("dhcp_pool").(*schema.Set).List()
		identifier         = d.Id()
		ipRanges           types.IPRanges
		dhcpPoolService    []*types.DhcpPoolService
	)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	if identifier == "" {
		identifier = networkName
	}

	network, err := vdc.GetOrgVdcNetworkByNameOrId(identifier, false)
	if err != nil {
		return fmt.Errorf("[isolated network update] error looking for %s: %s", identifier, err)
	}

	if d.HasChange("static_ip_pool") {
		ipRanges, err = expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
		if err != nil {
			return fmt.Errorf("[isolated network update] error expanding static IP pool: %s", err)
		}
		network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges = &ipRanges
	}

	if d.HasChange("dhcp_pool") {
		if len(dhcpPool) > 0 {
			for _, pool := range dhcpPool {

				poolMap := pool.(map[string]interface{})

				var poolService types.DhcpPoolService

				poolService.IsEnabled = true
				poolService.DefaultLeaseTime = poolMap["default_lease_time"].(int)
				poolService.MaxLeaseTime = poolMap["max_lease_time"].(int)
				poolService.LowIPAddress = poolMap["start_address"].(string)
				poolService.HighIPAddress = poolMap["end_address"].(string)
				dhcpPoolService = append(dhcpPoolService, &poolService)
			}
			network.OrgVDCNetwork.ServiceConfig.GatewayDhcpService.Pool = dhcpPoolService
		}
	}

	network.OrgVDCNetwork.Name = networkName
	network.OrgVDCNetwork.Description = networkDescription
	network.OrgVDCNetwork.IsShared = isShared

	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1 = dns1
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS2 = dns2
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS2 = dnsSuffix

	err = network.Update()
	if err != nil {
		return fmt.Errorf("error updating isolated network: %s", err)
	}

	// The update returns already a network. No need to retrieve it twice
	return genericVcdNetworkIsolatedRead(d, network, "resource-update")
}
