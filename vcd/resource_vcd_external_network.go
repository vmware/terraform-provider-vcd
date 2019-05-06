package vcd

import (
	"fmt"
	"log"
	"net"
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
						"is_inherited": &schema.Schema{
							Type:        schema.TypeBool,
							Required:    true,
							Description: "True if the IP scope is inherit from parent network",
						},
						"gateway": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Gateway of the network",
							ValidateFunc: singleIP(),
						},
						"netmask": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network mask",
							ValidateFunc: singleIP(),
						},
						"dns1": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Primary DNS server",
							ValidateFunc: singleIP(),
						},
						"dns2": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Secondary DNS server",
							ValidateFunc: singleIP(),
						},
						"dns_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Dns suffix",
						},
						"ip_range": &schema.Schema{
							Type:        schema.TypeList,
							Optional:    true,
							Description: "IP ranges used for static pool allocation in the network",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Start address of the IP range",
										ValidateFunc: singleIP(),
									},
									"end": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										Description:  "End address of the IP range",
										ValidateFunc: singleIP(),
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
							Description: "The vCenter server name",
						},
						"vsphere_network": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the port group",
						},
						"type": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The vSphere port group type. One of: DV_PORTGROUP (distributed virtual port group), NETWORK",
							ValidateFunc: validatePortGroupObjectType,
						},
					},
				},
			},
			"fence_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Isolation type of the network. If ParentNetwork is specified, this property controls connectivity to the parent. One of: bridged (connected directly to the ParentNetwork), isolated (not connected to any other network), natRouted (connected to the ParentNetwork via a NAT service)",
				ValidateFunc: validateFenceMode,
			},
			"retain_net_info_across_deployments": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
			"parent_network": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Contains reference to parent network",
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
			FenceMode:                      d.Get("fence_mode").(string),
			RetainNetInfoAcrossDeployments: d.Get("retain_net_info_across_deployments").(bool),
		},
	}

	ipScopesConfigurations := d.Get("ip_scope").([]interface{})
	// This is a limitation from the vcloud package
	/*	if len(ipScopes) != 1 {
		return &types.ExternalNetwork{}, fmt.Errorf("only one ip_scope is allowed")
	}*/
	var ipScopes []*types.IPScope
	for _, ipScopeValues := range ipScopesConfigurations {
		ipScopeConfiguration := ipScopeValues.(map[string]interface{})
		ipRanges := []*types.IPRange{}
		for _, ipRangeItem := range ipScopeConfiguration["ip_range"].([]interface{}) {
			ipRange := ipRangeItem.(map[string]interface{})
			ipRanges = append(ipRanges, &types.IPRange{
				StartAddress: ipRange["start"].(string),
				EndAddress:   ipRange["end"].(string),
			})
		}

		ipScope := &types.IPScope{
			IsInherited: ipScopeConfiguration["is_inherited"].(bool),
			Gateway:     ipScopeConfiguration["gateway"].(string),
			Netmask:     ipScopeConfiguration["netmask"].(string),
			DNS1:        ipScopeConfiguration["dns1"].(string),
			DNS2:        ipScopeConfiguration["dns2"].(string),
			IPRanges: &types.IPRanges{
				IPRange: ipRanges,
			},
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

		vCenterHref, err := getVcenterHref(vcdClient.VCDClient, portGroup["vcenter"].(string))
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
				fmt.Errorf("find less or more than one port groups - %d", len(portGroupRecord))
		}

		portGroups = append(portGroups, portGroupConfiguration)
	}
	params.VimPortGroupRefs = &types.VimObjectRefs{
		VimObjectRef: portGroups,
	}

	if parentNetworkName, ok := d.GetOk("parent_network"); ok {
		name := parentNetworkName.(string)
		parentNetwork, err := govcd.GetExternalNetworkByName(vcdClient.VCDClient, name)
		if err != nil {
			return &types.ExternalNetwork{}, fmt.Errorf("unable to find parent network %s (%s)", name, err)
		}
		params.Configuration.ParentNetwork = &types.Reference{
			HREF: parentNetwork.HREF,
			Type: parentNetwork.Type,
			Name: parentNetwork.Name,
		}
	}

	if description, ok := d.GetOk("description"); ok {
		params.Description = description.(string)
	}

	return params, nil
}

func getVcenterHref(vcdClient *govcd.VCDClient, name string) (string, error) {
	virtualCenters, err := govcd.QueryVirtualCenters(vcdClient, fmt.Sprintf("(name==%s)", name))
	if err != nil {
		return "", err
	}
	if len(virtualCenters) == 0 {
		return "", fmt.Errorf("No vSphere server found with name '%s'", name)
	}
	return virtualCenters[0].HREF, nil
}

// Remove this function when the version of Terraform is updated; it was added in 0.11.6
// singleIP returns a SchemaValidateFunc which tests if the provided value
// is of type string, and in valid single IP notation
func singleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		ip := net.ParseIP(v)
		if ip == nil {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP, got: %s", k, v))
		}
		return
	}
}

func validateFenceMode(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	switch v {
	case
		"bridged",
		"isolated",
		"natRouted":
		return
	default:
		errs = append(errs, fmt.Errorf("%q must be one of {bridged, isolated, natRouted}, got: %s", key, v))
	}
	return
}

func validatePortGroupObjectType(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	switch v {
	case
		"DV_PORTGROUP",
		"NETWORK":
		return
	default:
		errs = append(errs, fmt.Errorf("%q must be one of {DV_PORTGROUP, NETWORK}, got: %s", key, v))
	}
	return
}
