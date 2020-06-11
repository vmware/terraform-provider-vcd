package vcd

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVappNetworkCreate,
		Read:   resourceVappNetworkRead,
		Update: resourceVappNetworkUpdate,
		Delete: resourceVappNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdVappNetworkImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network name",
				// we can't change network name as this results in ID (HREF) change
			},
			"vapp_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp to use",
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
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "255.255.255.0",
				Description: "Netmask address for a subnet. Default is 255.255.255.0",
			},
			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Gateway of the network",
			},
			"dns1": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Primary DNS server",
			},
			"dns2": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Secondary DNS server",
			},
			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS suffix",
			},

			"guest_vlan_allowed": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if Network allows guest VLAN tagging",
			},
			"org_network_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "org network name to which vapp network is connected",
			},
			"firewall_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "firewall service enabled or disabled",
			},
			"nat_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "NAT service enabled or disabled",
			},
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
			"dhcp_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
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

						"enabled": &schema.Schema{
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
					},
				},
				Set: resourceVcdDhcpPoolHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
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
				Set: resourceVcdNetworkStaticIpPoolHash,
			},
		},
	}
}

func resourceVappNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	vappNetworkName := d.Get("name").(string)
	vappNetworkSettings := &govcd.VappNetworkSettings{
		Name:               vappNetworkName,
		Description:        d.Get("description").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		NatEnabled:         takeBoolPointer(d.Get("nat_enabled").(bool)),
		FirewallEnabled:    takeBoolPointer(d.Get("firewall_enabled").(bool)),
		RetainIpMacEnabled: takeBoolPointer(d.Get("retain_ip_mac_enabled").(bool)),
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	expandDhcpPool(d, vappNetworkSettings)

	var orgVdcNetwork *types.OrgVDCNetwork
	if networkId, ok := d.GetOk("org_network_name"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkId.(string), true)
		if err != nil {
			return err
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}
	vAppNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return fmt.Errorf("error creating vApp network. %s", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vappNetworkName {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		return fmt.Errorf("didn't find vApp network: %s", vappNetworkName)
	}

	// Parsing UUID from 'https://bos1-vcloud-static-170-210.eng.vmware.com/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset' or similar
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF, false)
	if err != nil {
		return fmt.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	return resourceVappNetworkRead(d, meta)
}

func expandDhcpPool(d *schema.ResourceData, vappNetworkSettings *govcd.VappNetworkSettings) {
	if dhcp, ok := d.GetOk("dhcp_pool"); ok && len(dhcp.(*schema.Set).List()) > 0 {
		for _, item := range dhcp.(*schema.Set).List() {
			data := item.(map[string]interface{})
			vappNetworkSettings.DhcpSettings = &govcd.DhcpSettings{
				IsEnabled:        data["enabled"].(bool),
				DefaultLeaseTime: data["default_lease_time"].(int),
				MaxLeaseTime:     data["max_lease_time"].(int),
				IPRange: &types.IPRange{StartAddress: data["start_address"].(string),
					EndAddress: data["end_address"].(string)}}
		}
	}
}

func resourceVappNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return genericVappNetworkRead(d, meta, "resource")
}

func genericVappNetworkRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding Vapp: %s", err)
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return fmt.Errorf("error getting vApp networks: %s", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	var networkId string
	vappNetworkName := d.Get("name").(string)
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.Link != nil {
			networkId, err = govcd.GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return fmt.Errorf("unable to get network ID from HREF: %s", err)
			}
			// Check name as well to support old resource IDs that are names and datasources that have names provided by the user
			if d.Id() == networkId || networkConfig.NetworkName == vappNetworkName {
				vAppNetwork = networkConfig
				break
			}
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		if origin == "resource" {
			log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[VAPP network read] %s : %s", govcd.ErrorEntityNotFound, vappNetworkName)
	}

	// needs to set for datasource. Do not set always as keep back compatibility when ID was name.
	if d.Id() == "" {
		d.SetId(normalizeId("urn:vcloud:network:", networkId))
	}
	_ = d.Set("description", vAppNetwork.Description)
	if config := vAppNetwork.Configuration; config != nil {
		if config.IPScopes != nil {
			_ = d.Set("gateway", config.IPScopes.IPScope[0].Gateway)
			_ = d.Set("netmask", config.IPScopes.IPScope[0].Netmask)
			_ = d.Set("dns1", config.IPScopes.IPScope[0].DNS1)
			_ = d.Set("dns2", config.IPScopes.IPScope[0].DNS2)
			_ = d.Set("dns_suffix", config.IPScopes.IPScope[0].DNSSuffix)
		}
		if config.Features != nil && config.Features.DhcpService != nil {
			transformed := schema.NewSet(resourceVcdDhcpPoolHash, []interface{}{})
			newValues := map[string]interface{}{
				"enabled":            config.Features.DhcpService.IsEnabled,
				"max_lease_time":     config.Features.DhcpService.MaxLeaseTime,
				"default_lease_time": config.Features.DhcpService.DefaultLeaseTime,
			}
			if config.Features.DhcpService.IPRange != nil {
				newValues["start_address"] = config.Features.DhcpService.IPRange.StartAddress
				// when only start address provided, API returns end address same as start address
				if config.Features.DhcpService.IPRange.StartAddress != config.Features.DhcpService.IPRange.EndAddress {
					newValues["end_address"] = config.Features.DhcpService.IPRange.EndAddress
				}
			}
			transformed.Add(newValues)
			err = d.Set("dhcp_pool", transformed)
			if err != nil {
				return fmt.Errorf("[vApp network DHCP pool read] set issue: %s", err)
			}
		}

		if config.IPScopes != nil && config.IPScopes.IPScope[0].IPRanges != nil {
			staticIpRanges := schema.NewSet(resourceVcdNetworkStaticIpPoolHash, []interface{}{})
			for _, ipRange := range config.IPScopes.IPScope[0].IPRanges.IPRange {
				newValues := map[string]interface{}{
					"start_address": ipRange.StartAddress,
					"end_address":   ipRange.EndAddress,
				}
				staticIpRanges.Add(newValues)
			}
			err = d.Set("static_ip_pool", staticIpRanges)
			if err != nil {
				return fmt.Errorf("[vApp network static pool read] set issue: %s", err)
			}
		}

		// TODO adjust when we have option to switch between API versions or upgrade the default version
		// API does not return GuestVlanAllowed if API client version is 27.0 (default at the moment) therefore we rely
		// on updating statefile only if the field was returned. In API v31.0 - the field is returned.
		if config.GuestVlanAllowed != nil {
			err = d.Set("guest_vlan_allowed", *config.GuestVlanAllowed)
			if err != nil {
				return err
			}
		}
		if config.ParentNetwork != nil {
			_ = d.Set("org_network_name", config.ParentNetwork.Name)
		} else {
			_ = d.Set("org_network_name", nil)
		}
		if config.Features != nil && config.Features.FirewallService != nil {
			_ = d.Set("firewall_enabled", vAppNetwork.Configuration.Features.FirewallService.IsEnabled)
		} else {
			_ = d.Set("firewall_enabled", nil)
		}
		if config.Features != nil && config.Features.NatService != nil {
			_ = d.Set("nat_enabled", config.Features.NatService.IsEnabled)
		} else {
			_ = d.Set("nat_enabled", nil)
		}
		_ = d.Set("retain_ip_mac_enabled", config.RetainNetInfoAcrossDeployments)
	}
	return nil
}

func resourceVappNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return err
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		ID:                 d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		NatEnabled:         takeBoolPointer(d.Get("nat_enabled").(bool)),
		FirewallEnabled:    takeBoolPointer(d.Get("firewall_enabled").(bool)),
		RetainIpMacEnabled: takeBoolPointer(d.Get("retain_ip_mac_enabled").(bool)),
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	expandDhcpPool(d, vappNetworkSettings)

	var orgVdcNetwork *types.OrgVDCNetwork
	if networkName, ok := d.GetOk("org_network_name"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkName.(string), true)
		if err != nil {
			return err
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}

	_, err = vapp.UpdateNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return fmt.Errorf("error creating vApp network. %s", err)
	}
	return resourceVappNetworkRead(d, meta)
}

func resourceVappNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp: %s", err)
	}

	_, err = vapp.RemoveNetwork(d.Id())
	if err != nil {
		return fmt.Errorf("error removing vApp network: %s", err)
	}

	d.SetId("")

	return nil
}

// resourceVcdVappNetworkImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_network.network_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name.network-name
func resourceVcdVappNetworkImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("[vApp network import] resource name must be specified as org-name.vdc-name.vapp-name.network-name")
	}
	orgName, vdcName, vappName, networkName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] error retrieving vapp %s: %s", vappName, err)
	}
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] error retrieving vApp network configuration %s: %s", networkName, err)
	}

	vappNetworkToImport := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			vappNetworkToImport = networkConfig
			break
		}
	}

	if vappNetworkToImport == (types.VAppNetworkConfiguration{}) {
		return nil, fmt.Errorf("didn't find vApp network: %s", networkName)
	}

	if isVappOrgNetwork(&vappNetworkToImport) {
		return nil, fmt.Errorf("found vApp org network, not vApp network: %s", networkName)
	}

	networkId, err := govcd.GetUuidFromHref(vappNetworkToImport.Link.HREF, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get network ID from HREF: %s", err)
	}

	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	if vcdClient.Org != orgName {
		_ = d.Set("org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		_ = d.Set("vdc", vdcName)
	}
	_ = d.Set("name", networkName)
	_ = d.Set("vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}

func resourceVcdDhcpPoolHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["enabled"].(bool)))
	buf.WriteString(fmt.Sprintf("%d-", m["max_lease_time"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["default_lease_time"].(int)))
	if m["start_address"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["start_address"].(string))))
	}
	if m["end_address"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["end_address"].(string))))
	}
	return hashcode.String(buf.String())
}

// Allows to identify if vApp Org network and not vApp network
func isVappOrgNetwork(networkConfig *types.VAppNetworkConfiguration) bool {
	return !isVappNetwork(networkConfig)
}
