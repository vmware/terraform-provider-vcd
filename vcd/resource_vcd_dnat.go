package vcd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdDNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdDNATCreate,
		Delete: resourceVcdDNATDelete,
		Read:   resourceVcdDNATRead,
		Update: resourceVcdDNATUpdate,

		Schema: map[string]*schema.Schema{
			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ext", "org"}, false),
			},
			"external_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"translated_port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"internal_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "TCP", // keep back compatibility as was hardcoded previously
			},
			"icmp_sub_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVcdDNATCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	portString := getPortString(d.Get("port").(int))
	translatedPortString := portString // default
	if d.Get("translated_port").(int) > 0 {
		translatedPortString = getPortString(d.Get("translated_port").(int))
	}

	protocol := d.Get("protocol").(string)
	icmpSubType := d.Get("icmp_sub_type").(string)
	if strings.ToUpper(protocol) != "ICMP" {
		icmpSubType = ""
	}

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	networkName := d.Get("network_name").(string)
	networkType := d.Get("network_type").(string)

	var natRule *types.NatRule

	if networkName != "" && networkType == "org" {
		orgVdcNetwork, err := getOrgVdcNetwork(d, vcdClient, networkName)
		if err != nil {
			return fmt.Errorf("unable to find orgVdcNetwork: %s, err: %s", networkName, err)
		}

		natRule, err = edgeGateway.AddDNATRule(govcd.NatRule{NetworkHref: orgVdcNetwork.HREF,
			ExternalIP: d.Get("external_ip").(string), ExternalPort: portString,
			InternalIP: d.Get("internal_ip").(string), InternalPort: translatedPortString,
			Protocol: protocol, IcmpSubType: icmpSubType, Description: d.Get("description").(string)})

		if err != nil {
			return fmt.Errorf("error creating DNAT rule: %#v", err)
		}
	} else if networkName != "" && networkType == "ext" {
		externalNetwork, err := govcd.GetExternalNetwork(vcdClient.VCDClient, networkName)
		if err != nil {
			return fmt.Errorf("unable to find external network: %s, err: %s", networkName, err)
		}

		natRule, err = edgeGateway.AddDNATRule(govcd.NatRule{NetworkHref: externalNetwork.ExternalNetwork.HREF,
			ExternalIP: d.Get("external_ip").(string), ExternalPort: portString,
			InternalIP: d.Get("internal_ip").(string), InternalPort: translatedPortString,
			Protocol: protocol, IcmpSubType: icmpSubType, Description: d.Get("description").(string)})
		if err != nil {
			return fmt.Errorf("error creating DNAT rule: %#v", err)
		}
	} else {
		// TODO remove when major release is done
		_, _ = fmt.Fprint(GetTerraformStdout(), "WARNING: This resource will require network_name and network_type in the next major version \n")
		task, err := edgeGateway.AddNATPortMapping("DNAT",
			d.Get("external_ip").(string),
			portString,
			d.Get("internal_ip").(string),
			translatedPortString, protocol,
			icmpSubType)
		if err != nil {
			return fmt.Errorf("error setting DNAT rules: %#v", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error completing tasks: %#v", err)
		}
	}

	if networkName != "" {
		d.SetId(natRule.ID)
	} else {
		d.SetId(d.Get("external_ip").(string) + ":" + portString + " > " + d.Get("internal_ip").(string) + ":" + translatedPortString)
	}
	return nil
}

func resourceVcdDNATRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")

	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// Terraform refresh won't work if Rule was edit in advanced edge gateway UI. vCD API uses <tag> elements to map edge gtw IDs
	// and UI will reset the tag element on update.

	var found bool

	networkName := d.Get("network_name")
	if nil != networkName && networkName.(string) != "" {
		natRule, err := edgeGateway.GetNatRule(d.Id())
		if err != nil {
			log.Printf("rule %s (stored in <tag> in Advanced GW case) not found: %s. Removing from state.", d.Id(), err)
			d.SetId("")
		}

		portInt, _ := strconv.Atoi(natRule.GatewayNatRule.OriginalPort)
		translatedPortInt, _ := strconv.Atoi(natRule.GatewayNatRule.TranslatedPort)

		d.Set("description", natRule.Description)
		d.Set("external_ip", natRule.GatewayNatRule.OriginalIP)
		d.Set("port", portInt)
		d.Set("internal_ip", natRule.GatewayNatRule.TranslatedIP)
		d.Set("translated_port", translatedPortInt)
		d.Set("protocol", natRule.GatewayNatRule.Protocol)
		d.Set("icmp_sub_type", natRule.GatewayNatRule.IcmpSubType)
		d.Set("network_name", natRule.GatewayNatRule.Interface.Name)

		orgVdcNetwork, _ := getOrgVdcNetwork(d, vcdClient, natRule.GatewayNatRule.Interface.Name)
		if orgVdcNetwork != nil {
			d.Set("network_type", "org")
		} else {
			externalNetwork, _ := govcd.GetExternalNetwork(vcdClient.VCDClient, natRule.GatewayNatRule.Interface.Name)
			if externalNetwork != nil && externalNetwork != (&govcd.ExternalNetwork{}) {
				d.Set("network_type", "ext")
			} else {
				return fmt.Errorf("didn't find external network or org VCD network with name: %s", natRule.GatewayNatRule.Interface.Name)
			}
		}

		found = true
	} else {
		// TODO remove when major release is done
		for _, r := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "DNAT" &&
				r.GatewayNatRule.OriginalIP == d.Get("external_ip").(string) &&
				r.GatewayNatRule.OriginalPort == getPortString(d.Get("port").(int)) {
				found = true
				d.Set("internal_ip", r.GatewayNatRule.TranslatedIP)
			}
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceVcdDNATDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Multiple VCD components need to run operations on the Edge Gateway, as
	// the edge gatway will throw back an error if it is already performing an
	// operation we must wait until we can aquire a lock on the client
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	portString := getPortString(d.Get("port").(int))
	translatedPortString := portString // default
	if d.Get("translated_port").(int) > 0 {
		translatedPortString = getPortString(d.Get("translated_port").(int))
	}

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	if d.Get("network_name").(string) != "" {
		err = edgeGateway.RemoveNATRule(d.Id())
		if err != nil {
			return fmt.Errorf("error deleting SNAT rule: %#v", err)
		}
	} else {
		// this for back compatibility when network name and network type isn't provided - TODO remove with major release
		task, err := edgeGateway.RemoveNATPortMapping("DNAT",
			d.Get("external_ip").(string),
			portString,
			d.Get("internal_ip").(string),
			translatedPortString)
		if err != nil {
			return fmt.Errorf("error setting DNAT rules: %#v", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error completing tasks: %#v", err)
		}
	}
	return nil
}

func resourceVcdDNATUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	// Update supports only when network name and network type provided
	networkName := d.Get("network_name")
	if networkName == nil || networkName.(string) == "" {
		return fmt.Errorf("update works only when network_name and network_type is provided and rule created using them \n")
	}

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	natRule, err := edgeGateway.GetNatRule(d.Id())
	if err != nil {
		log.Printf(" rule %s not found: %s. Removing from state.", d.Id(), err)
		d.SetId("")
	}

	natRule.GatewayNatRule.OriginalIP = d.Get("external_ip").(string)
	natRule.GatewayNatRule.OriginalPort = getPortString(d.Get("port").(int))
	natRule.GatewayNatRule.TranslatedIP = d.Get("internal_ip").(string)
	natRule.GatewayNatRule.TranslatedPort = getPortString(d.Get("translated_port").(int))
	natRule.GatewayNatRule.Protocol = d.Get("protocol").(string)
	natRule.GatewayNatRule.IcmpSubType = d.Get("icmp_sub_type").(string)
	natRule.Description = d.Get("description").(string)

	if d.Get("network_type").(string) == "org" {
		orgVdcNetwork, err := getOrgVdcNetwork(d, vcdClient, d.Get("network_name").(string))
		if orgVdcNetwork == nil || err != nil {
			return fmt.Errorf("unable to find orgVdcNetwork: %s, err: %s", networkName, err)
		}
		natRule.GatewayNatRule.Interface.Name = orgVdcNetwork.Name
		natRule.GatewayNatRule.Interface.HREF = orgVdcNetwork.HREF

	} else if d.Get("network_type").(string) == "ext" {
		externalNetwork, _ := govcd.GetExternalNetwork(vcdClient.VCDClient, d.Get("network_name").(string))
		if externalNetwork == nil || externalNetwork == (&govcd.ExternalNetwork{}) || err != nil {
			return fmt.Errorf("unable to find external network: %s, err: %s", networkName, err)
		}
		natRule.GatewayNatRule.Interface.Name = externalNetwork.ExternalNetwork.Name
		natRule.GatewayNatRule.Interface.HREF = externalNetwork.ExternalNetwork.HREF
	} else {
		return fmt.Errorf("network_type isn't provided or not `ext` or `org` ")
	}

	_, err = edgeGateway.UpdateNatRule(natRule)
	if err != nil {
		return fmt.Errorf("unable to update nat Rule: err: %s", err)
	}

	return resourceVcdSNATRead(d, meta)
}

func getOrgVdcNetwork(d *schema.ResourceData, vcdClient *VCDClient, networkname string) (*types.OrgVDCNetwork, error) {

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.FindVDCNetwork(networkname)
	if err != nil {
		log.Printf("[DEBUG] Network doesn't exist: " + networkname)
		return nil, err
	}

	return network.OrgVDCNetwork, nil
}
