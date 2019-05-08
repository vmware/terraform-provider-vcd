package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdExternalNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdExternalNetworkCreate,
		Delete: resourceVcdExternalNetworkDelete,
		Read:   resourceVcdExternalNetworkRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ip_scope": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "A list of IP scopes for the network",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							Description:  "Gateway of the network",
							ValidateFunc: validation.SingleIP(),
						},
						"netmask": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							Description:  "Network mask",
							ValidateFunc: validation.SingleIP(),
						},
						"dns1": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Description:  "Primary DNS server",
							ValidateFunc: validation.SingleIP(),
						},
						"dns2": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Description:  "Secondary DNS server",
							ValidateFunc: validation.SingleIP(),
						},
						"dns_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Dns suffix",
						},
						"static_ip_pool": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    true,
							Description: "IP ranges used for static pool allocation in the network",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										Description:  "Start address of the IP range",
										ValidateFunc: validation.SingleIP(),
									},
									"end_address": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										Description:  "End address of the IP range",
										ValidateFunc: validation.SingleIP(),
									},
								},
							},
						},
					},
				},
			},
			"vsphere_networks": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "A list of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcenter": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The vCenter server name",
						},
						"vsphere_network": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The name of the port group",
						},
						"type": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							Description:  "The vSphere port group type. One of: DV_PORTGROUP (distributed virtual port group), NETWORK",
							ValidateFunc: validation.StringInSlice([]string{"DV_PORTGROUP", "NETWORK"}, false),
						},
					},
				},
			},
			"retain_net_info_across_deployments": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
		},
	}
}

// Creates a new external network from a resource definition
func resourceVcdExternalNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] external network creation initiated")

	vcdClient := meta.(*VCDClient)

	params, err := getExternalNetworkInput(d, vcdClient)
	if err != nil {
		return err
	}

	task, err := govcd.CreateExternalNetwork(vcdClient.VCDClient, params)
	if err != nil {
		log.Printf("[DEBUG] Error creating external network: %#v", err)
		return fmt.Errorf("error creating external network: %#v", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		log.Printf("[DEBUG] Error waiting for external network to finish: %#v", err)
		return fmt.Errorf("error waiting for external network to finish: %#v", err)
	}

	d.SetId(d.Get("name").(string))
	log.Printf("[TRACE] external network created: %#v", task)
	return resourceVcdExternalNetworkRead(d, meta)
}

// Fetches information about an existing external network for a data definition
func resourceVcdExternalNetworkRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] external network read initiated")

	vcdClient := meta.(*VCDClient)

	externalNetwork, err := govcd.GetExternalNetwork(vcdClient.VCDClient, d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error fetching external network details %#v", err)
	}

	log.Printf("[TRACE] external network read completed: %#v", externalNetwork.ExternalNetwork)
	return nil
}

// Deletes a external network, optionally removing all objects in it as well
func resourceVcdExternalNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] external network delete started")

	vcdClient := meta.(*VCDClient)

	externalNetwork, err := govcd.GetExternalNetwork(vcdClient.VCDClient, d.Id())
	if err != nil {
		log.Printf("[DEBUG] Error fetching external network details %#v", err)
		return fmt.Errorf("error fetching external network details %#v", err)
	}

	err = externalNetwork.DeleteWait()
	if err != nil {
		log.Printf("[DEBUG] Error removing external network %#v", err)
		return fmt.Errorf("error removing external network %#v", err)
	}

	log.Printf("[TRACE] external network delete completed: %#v", externalNetwork)
	return nil
}

// helper for transforming the resource input into the ExternalNetwork structure
// any cast operations or default values should be done here so that the create method is simple
func getExternalNetworkInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.ExternalNetwork, error) {
	params := &types.ExternalNetwork{
		Name:        d.Get("name").(string),
		Xmlns:       types.XMLNamespaceExtension,
		XmlnsVCloud: types.XMLNamespaceVCloud,
		Configuration: &types.NetworkConfiguration{
			Xmlns:                          types.XMLNamespaceVCloud,
			RetainNetInfoAcrossDeployments: d.Get("retain_net_info_across_deployments").(bool),
		},
	}

	ipScopesConfigurations := d.Get("ip_scope").([]interface{})

	var ipScopes []*types.IPScope
	for _, ipScopeValues := range ipScopesConfigurations {
		ipScopeConfiguration := ipScopeValues.(map[string]interface{})
		ipRanges := []*types.IPRange{}
		for _, ipRangeItem := range ipScopeConfiguration["static_ip_pool"].([]interface{}) {
			ipRange := ipRangeItem.(map[string]interface{})
			ipRanges = append(ipRanges, &types.IPRange{
				StartAddress: ipRange["start_address"].(string),
				EndAddress:   ipRange["end_address"].(string),
			})
		}

		ipScope := &types.IPScope{
			Gateway: ipScopeConfiguration["gateway"].(string),
			Netmask: ipScopeConfiguration["netmask"].(string),
			IPRanges: &types.IPRanges{
				IPRange: ipRanges,
			},
		}

		if ipScopeConfiguration["dns1"] != nil && "" != ipScopeConfiguration["dns1"].(string) {
			ipScope.DNS1 = ipScopeConfiguration["dns1"].(string)
		}

		if ipScopeConfiguration["dns2"] != nil && "" != ipScopeConfiguration["dns2"].(string) {
			ipScope.DNS2 = ipScopeConfiguration["dns2"].(string)
		}

		if ipScopeConfiguration["dns_suffix"] != nil && "" != ipScopeConfiguration["dns_suffix"].(string) {
			ipScope.DNSSuffix = ipScopeConfiguration["dns_suffix"].(string)
		}

		ipScopes = append(ipScopes, ipScope)
	}

	params.Configuration.IPScopes = &types.IPScopes{
		IPScope: ipScopes,
	}

	var portGroups []*types.VimObjectRef
	for _, vsphereNetwork := range d.Get("vsphere_networks").([]interface{}) {
		portGroup := vsphereNetwork.(map[string]interface{})
		portGroupConfiguration := &types.VimObjectRef{
			//MoRef:         portGroup["vsphere_network"].(string),
			VimObjectType: strings.ToUpper(portGroup["type"].(string)),
		}

		vCenterHref, err := GetVcenterHref(vcdClient.VCDClient, portGroup["vcenter"].(string))
		if err != nil {
			return &types.ExternalNetwork{}, fmt.Errorf("unable to find vCenter %s (%#v)", portGroup["vcenter"].(string), err)
		}
		portGroupConfiguration.VimServerRef = &types.Reference{
			HREF: fmt.Sprintf("%s/admin/extension/vimServer/%s", vcdClient.Client.VCDHREF.String(), vCenterHref),
		}

		var portGroupRecord []*types.PortGroupRecordType
		if portGroup["type"].(string) == "NETWORK" {
			portGroupRecord, err = govcd.QueryNetworkPortGroup(vcdClient.VCDClient, portGroup["vsphere_network"].(string))
		} else {
			portGroupRecord, err = govcd.QueryDistributedPortGroup(vcdClient.VCDClient, portGroup["vsphere_network"].(string))
		}
		if err != nil {
			return &types.ExternalNetwork{},
				fmt.Errorf("unable to find port group %s (%#v)", portGroup["vcenter"].(string), err)
		}

		if len(portGroupRecord) == 1 {
			portGroupConfiguration.MoRef = portGroupRecord[0].MoRef
		} else {
			return &types.ExternalNetwork{},
				fmt.Errorf("found less or more than one port groups - %d", len(portGroupRecord))
		}

		portGroups = append(portGroups, portGroupConfiguration)
	}
	params.VimPortGroupRefs = &types.VimObjectRefs{
		VimObjectRef: portGroups,
	}

	if description, ok := d.GetOk("description"); ok {
		params.Description = description.(string)
	}

	return params, nil
}

func GetVcenterHref(vcdClient *govcd.VCDClient, name string) (string, error) {
	virtualCenters, err := govcd.QueryVirtualCenters(vcdClient, fmt.Sprintf("(name==%s)", name))
	if err != nil {
		return "", err
	}
	if len(virtualCenters) == 0 || len(virtualCenters) > 1 {
		return "", fmt.Errorf("vSphere server found %d instances with name '%s' while expected one", len(virtualCenters), name)
	}
	return virtualCenters[0].HREF, nil
}
