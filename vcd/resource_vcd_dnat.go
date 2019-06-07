package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdDNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdDNATCreate,
		Delete: resourceVcdDNATDelete,
		Read:   resourceVcdDNATRead,

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
				ForceNew: true,
			},
			"network_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ext",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ext", "org"}, false),
			},
			"external_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"translated_port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"internal_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "TCP", // keep back compatibility as was hardcoded previously
			},
			"icmp_sub_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "any",
			},
		},
	}
}

func resourceVcdDNATCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Multiple VCD components need to run operations on the Edge Gateway, as
	// the edge gateway will throw back an error if it is already performing an
	// operation we must wait until we can acquire a lock on the client
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	portString := getPortString(d.Get("port").(int))
	translatedPortString := portString // default
	if d.Get("translated_port").(int) > 0 {
		translatedPortString = getPortString(d.Get("translated_port").(int))
	}

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	protocol := d.Get("protocol").(string)
	icmpSubType := d.Get("icmp_sub_type").(string)
	if strings.ToUpper(protocol) != "ICMP" {
		icmpSubType = ""
	}

	var orgVdcnetwork *types.OrgVDCNetwork
	providedNetworkName := d.Get("network_name")
	if nil != providedNetworkName && "" != providedNetworkName.(string) {
		orgVdcnetwork, err = getNetwork(d, vcdClient, providedNetworkName.(string))
	}
	// TODO enable when external network supported
	/*else {
		_, _ = fmt.Fprint(GetTerraformStdout(), "WARNING: This resource will require network_name and network_type in the next major version \n")
	}*/
	if err != nil {
		return fmt.Errorf("unable to find orgVdcnetwork: %s, err: %s", providedNetworkName.(string), err)
	}

	if nil != providedNetworkName && providedNetworkName != "" {
		task, err := edgeGateway.AddNATPortMappingWithUplink(orgVdcnetwork, "DNAT",
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

	} else {
		// TODO remove when major release is done
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

	if nil != providedNetworkName && "" != providedNetworkName {
		d.SetId(orgVdcnetwork.Name + ":" + d.Get("external_ip").(string) + ":" + portString + " > " + d.Get("internal_ip").(string) + ":" + translatedPortString)
	} else {
		d.SetId(d.Get("external_ip").(string) + ":" + portString + " > " + d.Get("internal_ip").(string) + ":" + translatedPortString)
	}
	return nil
}

func resourceVcdDNATRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	e, err := vcdClient.GetEdgeGatewayFromResource(d)

	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	var found bool

	providedNetworkName := d.Get("network_name")
	if nil != providedNetworkName && providedNetworkName.(string) != "" {
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "DNAT" && r.GatewayNatRule.Interface.Name == d.Get("network_name") &&
				r.GatewayNatRule.OriginalIP == d.Get("external_ip").(string) &&
				r.GatewayNatRule.OriginalPort == getPortString(d.Get("port").(int)) {
				found = true
				d.Set("internal_ip", r.GatewayNatRule.TranslatedIP)
			}
		}
	} else {
		// TODO remove when major release is done
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
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

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)

	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
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
	return nil
}

func getNetwork(d *schema.ResourceData, vcdClient *VCDClient, networkname string) (*types.OrgVDCNetwork, error) {

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return &types.OrgVDCNetwork{}, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.FindVDCNetwork(networkname)
	if err != nil {
		log.Printf("[DEBUG] Network doesn't exist: " + networkname)
		return nil, err
	}

	return network.OrgVDCNetwork, nil
}
