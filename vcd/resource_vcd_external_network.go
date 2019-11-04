package vcd

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdExternalNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdExternalNetworkCreate,
		Delete: resourceVcdExternalNetworkDelete,
		Read:   resourceVcdExternalNetworkRead,
		Importer: &schema.ResourceImporter{
			State: resourceVcdExternalNetworkImport,
		},
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
							Description: "DNS suffix",
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
			"vsphere_network": &schema.Schema{
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
						"name": &schema.Schema{
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

// resourceVcdExternalNetworkCreate creates a new external network from a resource definition
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

func setExternalNetworkData(d *schema.ResourceData, extNetRes StringMap) error {
	_ = d.Set("name", extNetRes["name"])
	_ = d.Set("description", extNetRes["description"])
	_ = d.Set("retain_net_info_across_deployments", extNetRes["retain_net_info_across_deployments"])

	err := d.Set("ip_scope", extNetRes["ip_scope"])
	if err != nil {
		return err
	}

	err = d.Set("vsphere_network", extNetRes["vsphere_network"])
	if err != nil {
		return err
	}

	log.Printf("[TRACE] external network read completed: %#v", extNetRes)
	return nil
}

// resourceVcdExternalNetworkRead fetches information about an existing external network
func resourceVcdExternalNetworkRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] external network read initiated")

	vcdClient := meta.(*VCDClient)

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	extNeRes, ID, err := getExternalNetworkResource(vcdClient.VCDClient, identifier)

	if err != nil {
		return fmt.Errorf("error fetching external network (%s) details %s", identifier, err)
	}
	err = setExternalNetworkData(d, extNeRes)
	if err != nil {
		return err
	}
	d.SetId(ID)

	return nil
}

// resourceVcdExternalNetworkDelete deletes an external network, optionally removing all objects in it as well
func resourceVcdExternalNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] external network delete started")

	vcdClient := meta.(*VCDClient)

	externalNetwork, err := vcdClient.GetExternalNetworkByNameOrId(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Error fetching external network details %s", err)
		return fmt.Errorf("error fetching external network details %s", err)
	}

	err = externalNetwork.DeleteWait()
	if err != nil {
		log.Printf("[DEBUG] Error removing external network %#v", err)
		return fmt.Errorf("error removing external network %s", err)
	}

	log.Printf("[TRACE] external network delete completed: %#v", externalNetwork)
	return nil
}

// getExternalNetworkInput is an helper for transforming the resource input into the ExternalNetwork structure
// any cast operations or default values should be done here so that the create method is simple
func getExternalNetworkInput(d *schema.ResourceData, vcdClient *VCDClient) (*types.ExternalNetwork, error) {
	params := &types.ExternalNetwork{
		Name: d.Get("name").(string),
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

		if ipScopeConfiguration["dns1"] != nil && ipScopeConfiguration["dns1"].(string) != "" {
			ipScope.DNS1 = ipScopeConfiguration["dns1"].(string)
		}

		if ipScopeConfiguration["dns2"] != nil && ipScopeConfiguration["dns2"].(string) != "" {
			ipScope.DNS2 = ipScopeConfiguration["dns2"].(string)
		}

		if ipScopeConfiguration["dns_suffix"] != nil && ipScopeConfiguration["dns_suffix"].(string) != "" {
			ipScope.DNSSuffix = ipScopeConfiguration["dns_suffix"].(string)
		}

		ipScopes = append(ipScopes, ipScope)
	}

	params.Configuration.IPScopes = &types.IPScopes{
		IPScope: ipScopes,
	}

	var portGroups []*types.VimObjectRef
	for _, vsphereNetwork := range d.Get("vsphere_network").([]interface{}) {
		portGroup := vsphereNetwork.(map[string]interface{})
		portGroupConfiguration := &types.VimObjectRef{
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
			portGroupRecord, err = govcd.QueryNetworkPortGroup(vcdClient.VCDClient, portGroup["name"].(string))
		} else {
			portGroupRecord, err = govcd.QueryDistributedPortGroup(vcdClient.VCDClient, portGroup["name"].(string))
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

// getExternalNetworkResource retrieves an external network and returns an interface map corresponding to the resource
// Input: vcdClient , external network Identifier (either name or ID)
// output: StringMap (representing the resource), external network ID, error
func getExternalNetworkResource(vcdClient *govcd.VCDClient, extNetIdentifier string) (StringMap, string, error) {
	var extNetRes = make(StringMap)

	externalNetwork, err := vcdClient.GetExternalNetworkByNameOrId(extNetIdentifier)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching external network details %s", err)
	}

	// Although the resource allows an array of vCenters,
	// the current implementation of external network only records one.
	vcenterHref := ""

	if externalNetwork.ExternalNetwork.VimPortGroupRef != nil && externalNetwork.ExternalNetwork.VimPortGroupRef.VimServerRef != nil {
		vcenterHref = externalNetwork.ExternalNetwork.VimPortGroupRef.VimServerRef.HREF
	} else {
		return nil, "", fmt.Errorf("error retrieving VC HREF : %s", err)
	}
	virtualCenters, err := govcd.QueryVirtualCenters(vcdClient, fmt.Sprintf("(href==%s)", vcenterHref))
	if err != nil {
		return nil, "", err
	}
	if len(virtualCenters) == 0 {
		return nil, "", fmt.Errorf("no virtual centers found with HREF %s", vcenterHref)
	}

	var ipScopes []StringMap
	for _, ips := range externalNetwork.ExternalNetwork.Configuration.IPScopes.IPScope {
		ipScope := StringMap{
			"gateway":    ips.Gateway,
			"dns1":       ips.DNS1,
			"dns2":       ips.DNS2,
			"dns_suffix": ips.DNSSuffix,
			"netmask":    ips.Netmask,
		}
		var stIpPool []StringMap
		for _, ipr := range ips.IPRanges.IPRange {
			ipRange := StringMap{
				"start_address": ipr.StartAddress,
				"end_address":   ipr.EndAddress,
			}
			stIpPool = append(stIpPool, ipRange)
		}
		ipScope["static_ip_pool"] = stIpPool
		ipScopes = append(ipScopes, ipScope)
	}

	portGroupMoRef := externalNetwork.ExternalNetwork.VimPortGroupRef.MoRef

	portGroups, err := govcd.QueryPortGroups(vcdClient,
		fmt.Sprintf("(moref==%s;portgroupType==%s)",
			url.QueryEscape(portGroupMoRef),
			url.QueryEscape(externalNetwork.ExternalNetwork.VimPortGroupRef.VimObjectType)))
	if err != nil {
		return StringMap{}, "", fmt.Errorf("error retrieving port group %s: %s", portGroupMoRef, err)
	}

	portGroupName := ""
	for _, pg := range portGroups {
		if portGroupName != "" {
			return StringMap{}, "", fmt.Errorf("more than one portgroup found for moref %s", portGroupMoRef)
		}
		portGroupName = pg.Name
	}

	extNetRes["vsphere_network"] = []StringMap{
		StringMap{
			"name":    portGroupName,
			"vcenter": virtualCenters[0].Name,
			"type":    externalNetwork.ExternalNetwork.VimPortGroupRef.VimObjectType,
		},
	}

	extNetRes["ip_scope"] = ipScopes
	extNetRes["name"] = externalNetwork.ExternalNetwork.Name
	extNetRes["description"] = externalNetwork.ExternalNetwork.Description
	extNetRes["retain_net_info_across_deployments"] = externalNetwork.ExternalNetwork.Configuration.RetainNetInfoAcrossDeployments

	return extNetRes, externalNetwork.ExternalNetwork.ID, nil
}

// resourceVcdExternalNetworkImport is responsible for importing the resource.
// The d.ID() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
// For this resource, the import path is just the external network name.
//
// Example import path (id): externalNetworkName
// Example import command:   terraform import vcd_external_network.externalNetworkName externalNetworkName
func resourceVcdExternalNetworkImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	vcdClient := meta.(*VCDClient)
	extNetRes, ID, err := getExternalNetworkResource(vcdClient.VCDClient, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error fetching external network details %s", err)
	}

	err = setExternalNetworkData(d, extNetRes)
	if err != nil {
		return nil, err
	}
	d.SetId(ID)
	return []*schema.ResourceData{d}, nil
}
