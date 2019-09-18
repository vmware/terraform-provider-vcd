package vcd

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkRouted() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkRoutedCreate,
		Read:   resourceVcdNetworkRead,
		Delete: resourceVcdNetworkDeleteLocked,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"netmask": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "255.255.255.0",
			},

			"gateway": &schema.Schema{
				Type:     schema.TypeString,
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
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"dhcp_pool": &schema.Schema{
				Type:     schema.TypeSet,
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
				Set: resourceVcdNetworkIPAddressHash,
			},
			"static_ip_pool": &schema.Schema{
				Type:     schema.TypeSet,
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
				Set: resourceVcdNetworkIPAddressHash,
			},
		},
	}
}

func resourceVcdNetworkRoutedCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	// Collecting components
	egName, ok := d.GetOk("edge_gateway")
	var edgeGateway govcd.EdgeGateway

	edgeGatewayName := ""
	if ok {
		edgeGatewayName = egName.(string)
		eg, err := vdc.FindEdgeGateway(edgeGatewayName)
		if err == nil {
			edgeGateway = eg
		} else {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}
	}

	gwName, ok := d.GetOk("gateway")
	gatewayName := ""
	if ok {
		gatewayName = gwName.(string)
	}

	ipRanges := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  d.Get("name").(string),

		EdgeGateway: &types.Reference{
			HREF: edgeGateway.EdgeGateway.HREF,
		},
		Configuration: &types.NetworkConfiguration{
			FenceMode: "natRouted",
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

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	network, err := vdc.FindVDCNetwork(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("error finding network: %#v", err)
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok {
		task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.(*schema.Set).List())
		if err != nil {
			return fmt.Errorf("error adding DHCP pool: %#v", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}

	}

	d.SetId(d.Get("name").(string))

	return resourceVcdNetworkRead(d, meta)
}

func resourceVcdNetworkRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.FindVDCNetwork(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}

	d.Set("name", network.OrgVDCNetwork.Name)
	d.Set("href", network.OrgVDCNetwork.HREF)
	if c := network.OrgVDCNetwork.Configuration; c != nil {
		d.Set("fence_mode", c.FenceMode)
		if c.IPScopes != nil {
			d.Set("gateway", c.IPScopes.IPScope[0].Gateway)
			d.Set("netmask", c.IPScopes.IPScope[0].Netmask)
			d.Set("dns1", c.IPScopes.IPScope[0].DNS1)
			d.Set("dns2", c.IPScopes.IPScope[0].DNS2)
		}
	}

	return nil
}

func resourceVcdNetworkDeleteLocked(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	return resourceVcdNetworkDelete(d, meta)
}

func resourceVcdNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.FindVDCNetwork(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error finding network: %#v", err)
	}

	task, err := network.Delete()
	if err != nil {
		return fmt.Errorf("error deleting network: %#v", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	return nil
}

func resourceVcdNetworkIPAddressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["start_address"].(string))))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["end_address"].(string))))

	return hashcode.String(buf.String())
}
