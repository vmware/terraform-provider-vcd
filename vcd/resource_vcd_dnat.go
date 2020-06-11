package vcd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// deprecated in favor of vcd_nsxv_dnat
func resourceVcdDNAT() *schema.Resource {
	return &schema.Resource{
		Create:             resourceVcdDNATCreate,
		Delete:             resourceVcdDNATDelete,
		Read:               resourceVcdDNATRead,
		Update:             resourceVcdDNATUpdate,
		DeprecationMessage: "vcd_dnat is deprecated. It should only be used for non-advanced edge gateways. Use vcd_nsxv_dnat instead.",

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
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tcp", // keep back compatibility as was hardcoded previously
				ValidateFunc: validateCase("lower"),
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

	if (networkName != "" && networkType == "") || (networkName == "" && networkType != "") {
		return fmt.Errorf("network_type and network_name are used together")
	}

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
		if !vcdClient.Client.IsSysAdmin {
			return fmt.Errorf("functionality requires system administrator privileges")
		}

		externalNetwork, err := vcdClient.GetExternalNetworkByNameOrId(networkName)
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
		_, _ = fmt.Fprint(getTerraformStdout(), "WARNING: This resource will require network_name and network_type in the next major version \n")
		//lint:ignore SA1019 Preserving back compatibility until removal
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
	if networkName != nil && networkName.(string) != "" {
		natRule, err := edgeGateway.GetNatRule(d.Id())
		if err != nil {
			log.Printf("[DEBUG] rule %s (stored in <tag> in Advanced GW case) not found: %s. Removing from state.", d.Id(), err)
			d.SetId("")
			return nil
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

		orgVdcNetwork, err := getOrgVdcNetwork(d, vcdClient, natRule.GatewayNatRule.Interface.Name)
		if orgVdcNetwork != nil {
			d.Set("network_type", "org")
		} else {
			log.Printf("[DEBUG] didn't find org VDC network with name: %s, %#v", natRule.GatewayNatRule.Interface.Name, err)
		}

		_, extNetwErr := vcdClient.GetExternalNetworkByNameOrId(natRule.GatewayNatRule.Interface.Name)
		if extNetwErr == nil {
			d.Set("network_type", "ext")
		} else {
			log.Printf("[DEBUG] didn't find external network with name: %s, %s", natRule.GatewayNatRule.Interface.Name, extNetwErr)
		}

		if orgVdcNetwork != nil && extNetwErr == nil {
			return fmt.Errorf("found external network or org VCD network with same name: %s", natRule.GatewayNatRule.Interface.Name)
		}

		if orgVdcNetwork == nil && extNetwErr != nil {
			return fmt.Errorf("issue updating resource state. Didn't find external network or org VCD network with name: %s", natRule.GatewayNatRule.Interface.Name)
		}

		return nil
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
		log.Printf("[INFO] Removing from state.")
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
			return fmt.Errorf("error deleting DNAT rule: %#v", err)
		}
	} else {
		// this for back compatibility when network name and network type isn't provided - TODO remove with major release
		//lint:ignore SA1019 Preserving back compatibility until removal
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
	networkName, ok := d.GetOk("network_name")
	if !ok || networkName == "" {
		return fmt.Errorf("update works only when network_name and network_type is provided and rule created using them")
	}

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	natRule, err := edgeGateway.GetNatRule(d.Id())
	if err != nil {
		log.Printf("[DEBUG] rule %s not found: %s. Removing from state.", d.Id(), err)
		d.SetId("")
		return nil
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
		externalNetwork, err := vcdClient.GetExternalNetworkByName(d.Get("network_name").(string))
		if err != nil {
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

	network, err := vdc.GetOrgVdcNetworkByName(networkname, false)
	if err != nil {
		log.Printf("[DEBUG] Network doesn't exist: " + networkname)
		return nil, err
	}

	return network.OrgVDCNetwork, nil
}
