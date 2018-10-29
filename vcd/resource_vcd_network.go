package vcd

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/govcd"
	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
	"log"
	"strings"
)

// DEPRECATED: use vcd_network_routed instead
func resourceVcdNetwork() *schema.Resource {
	return &schema.Resource{
		// DeprecationMessage requires a version of Terraform newer than what
		// we currently use in the vendor directory
		// DeprecationMessage: "Deprecated. Use vcd_network_routed instead",
		Create: resourceVcdNetworkDeprecatedCreate,
		Read:   resourceVcdNetworkRead,
		Delete: resourceVcdNetworkDelete,

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
			},
			"vdc": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
			"fence_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "natRouted",
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
				Required: true,
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

func resourceVcdNetworkDeprecatedCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] CLIENT: %#v", vcdClient)
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	// Collecting components
	egName, ok := d.GetOk("edge_gateway")
	var edgeGateway govcd.EdgeGateway
	fenceMode := "natRouted"

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
		Configuration: &types.NetworkConfiguration{
			FenceMode: fenceMode,
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     d.Get("netmask").(string),
					DNS1:        d.Get("dns1").(string),
					DNS2:        d.Get("dns2").(string),
					DNSSuffix:   d.Get("dns_suffix").(string),
					IPRanges:    &ipRanges,
				},
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}

	orgVDCNetwork.EdgeGateway = &types.Reference{
		HREF: edgeGateway.EdgeGateway.HREF,
	}

	util.Logger.Printf("[INFO] NETWORK: %#v", orgVDCNetwork)
	util.Logger.Printf("[INFO] NETWORK: %#v", orgVDCNetwork.Configuration)
	util.Logger.Printf("[INFO] NETWORK: %#v", orgVDCNetwork.Configuration.IPScopes)

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		return resource.RetryableError(vdc.CreateOrgVDCNetworkWait(orgVDCNetwork))
	})
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	// err = vdc.Refresh()
	// if err != nil {
	// 	return fmt.Errorf("error refreshing VDC: %#v", err)
	// }

	network, err := vdc.FindVDCNetwork(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("error finding network: %#v", err)
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.(*schema.Set).List())
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error adding DHCP pool: %#v", err))
			}

			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return fmt.Errorf(errorCompletingTask, err)
		}

	}

	d.SetId(d.Get("name").(string))

	return resourceVcdNetworkRead(d, meta)
}

func resourceVcdNetworkRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[DEBUG] VCD Client configuration: %#v", vcdClient)
	//log.Printf("[DEBUG] VCD Client configuration: %#v", vcdClient.OrgVdc)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	// err = vdc.Refresh()
	// if err != nil {
	// 	return fmt.Errorf("error refreshing VDC: %#v", err)
	// }

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
			d.Set("gateway", c.IPScopes.IPScope.Gateway)
			d.Set("netmask", c.IPScopes.IPScope.Netmask)
			d.Set("dns1", c.IPScopes.IPScope.DNS1)
			d.Set("dns2", c.IPScopes.IPScope.DNS2)
		}
	}

	return nil
}

func resourceVcdNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	// err = vdc.Refresh()
	// if err != nil {
	// 	return fmt.Errorf("error refreshing VDC: %#v", err)
	// }

	network, err := vdc.FindVDCNetwork(d.Id())
	if err != nil {
		return fmt.Errorf("error finding network: %#v", err)
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := network.Delete()
		if err != nil {
			return resource.RetryableError(
				fmt.Errorf("error Deleting Network: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
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
