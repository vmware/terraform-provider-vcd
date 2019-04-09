package vcd

import (
	"fmt"
	"log"
	"net"

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
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_inherited": &schema.Schema{
							Type:     schema.TypeBool,
							Required: true,
						},
						"gateway": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: singleIP(),
						},
						"netmask": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: singleIP(),
						},
						"dns1": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: singleIP(),
						},
						"dns2": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: singleIP(),
						},
						"ip_range": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: singleIP(),
									},
									"end": &schema.Schema{
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: singleIP(),
									},
								},
							},
						},
					},
				},
				Description: "",
			},
			"vim_port_group": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vim_server": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"mo_ref": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"vim_object_type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"fence_mode": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "",
			},
			"retain_net_info_across_deployments": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "",
			},
			"parent_network": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "",
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

	externalNetwork, err := govcd.GetExternalNetworkByName2(vcdClient.VCDClient, d.Id())
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

	externalNetwork, err := govcd.GetExternalNetworkByName2(vcdClient.VCDClient, d.Id())
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
		Xmlns:       "http://www.vmware.com/vcloud/extension/v1.5",
		XmlnsVCloud: "http://www.vmware.com/vcloud/v1.5",
		Configuration: &types.NetworkConfiguration{
			Xmlns:                          "http://www.vmware.com/vcloud/v1.5",
			FenceMode:                      d.Get("fence_mode").(string),
			RetainNetInfoAcrossDeployments: d.Get("retain_net_info_across_deployments").(bool),
		},
	}

	ipScopes := d.Get("ip_scope").([]interface{})
	// This is a limitation from the vcloud package
	if len(ipScopes) != 1 {
		return &types.ExternalNetwork{}, fmt.Errorf("only one ip_scope is allowed")
	}
	ipScope := ipScopes[0].(map[string]interface{})
	ipRanges := []*types.IPRange{}
	for _, ipRangeItem := range ipScope["ip_range"].([]interface{}) {
		ipRange := ipRangeItem.(map[string]interface{})
		ipRanges = append(ipRanges, &types.IPRange{
			StartAddress: ipRange["start"].(string),
			EndAddress:   ipRange["end"].(string),
		})
	}
	params.Configuration.IPScopes = &types.IPScopes{
		IPScope: types.IPScope{
			IsInherited: ipScope["is_inherited"].(bool),
			Gateway:     ipScope["gateway"].(string),
			Netmask:     ipScope["netmask"].(string),
			DNS1:        ipScope["dns1"].(string),
			DNS2:        ipScope["dns2"].(string),
			IPRanges: &types.IPRanges{
				IPRange: ipRanges,
			},
		},
	}

	vimPortGroups := []*types.VimObjectRef{}
	for _, vimPortGroupItem := range d.Get("vim_port_group").([]interface{}) {
		vimPortGroup := vimPortGroupItem.(map[string]interface{})
		vimPortGroups = append(vimPortGroups, &types.VimObjectRef{
			VimServerRef: &types.Reference{
				HREF: fmt.Sprintf("%s/admin/extension/vimServer/%s", vcdClient.Client.VCDHREF.String(), vimPortGroup["vim_server"].(string)),
			},
			MoRef:         vimPortGroup["mo_ref"].(string),
			VimObjectType: vimPortGroup["vim_object_type"].(string),
		})
	}
	params.VimPortGroupRefs = &types.VimObjectRefs{
		VimObjectRef: vimPortGroups,
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
