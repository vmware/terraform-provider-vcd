package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
)

func resourceVcdSNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdSNATCreate,
		Delete: resourceVcdSNATDelete,
		Read:   resourceVcdSNATRead,
		Update: resourceVcdSNATUpdate,

		Schema: map[string]*schema.Schema{
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
			"edge_gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ext",
				ValidateFunc: validation.StringInSlice([]string{"ext", "org"}, false),
			},
			"external_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"internal_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVcdSNATCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

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

		natRule, err = edgeGateway.AddSNATRule(orgVdcNetwork.HREF, d.Get("external_ip").(string),
			d.Get("internal_ip").(string), d.Get("description").(string))
		if err != nil {
			return fmt.Errorf("error creating SNAT rule: %#v", err)
		}
	} else if networkName != "" && networkType == "ext" {
		externalNetwork, err := govcd.GetExternalNetwork(vcdClient.VCDClient, networkName)
		if err != nil {
			return fmt.Errorf("unable to find external network: %s, err: %s", networkName, err)
		}

		natRule, err = edgeGateway.AddSNATRule(externalNetwork.ExternalNetwork.HREF, d.Get("external_ip").(string),
			d.Get("internal_ip").(string), d.Get("description").(string))
		if err != nil {
			return fmt.Errorf("error creating SNAT rule: %#v", err)
		}
	} else {
		_, _ = fmt.Fprint(GetTerraformStdout(), "WARNING: this resource will require network_name and network_type in the next major version \n")
		// TODO remove when major release is done
		// this for back compatibility  when network name and network type isn't provided - this assign rule only for first external network
		task, err := edgeGateway.AddNATMapping("SNAT", d.Get("internal_ip").(string),
			d.Get("external_ip").(string))
		if err != nil {
			return fmt.Errorf("error setting SNAT rules: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
	}

	if networkName != "" {
		d.SetId(natRule.ID)
	} else {
		// TODO remove when major release is done
		d.SetId(d.Get("internal_ip").(string))
	}
	return nil
}

func resourceVcdSNATRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// Terraform refresh won't work if Rule was edit in advanced edge gateway UI. vCD API uses tag elements to map edge gtw IDs
	// and UI will reset the tag element on update.

	var found bool

	networkName := d.Get("network_name")
	if nil != networkName && networkName.(string) != "" {
		natRule, err := edgeGateway.GetNatRule(d.Id())
		if err != nil {
			log.Printf("rule %s (stored in <tag> in Advanced GW case) not found: %s. Removing from state.", d.Id(), err)
			d.SetId("")
		}

		d.Set("description", natRule.Description)
		d.Set("external_ip", natRule.GatewayNatRule.OriginalIP)
		d.Set("internal_ip", natRule.GatewayNatRule.TranslatedIP)
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
	} else { // TODO remove after network_name becomes mandatory
		for _, r := range edgeGateway.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "SNAT" && r.GatewayNatRule.OriginalIP == d.Get("internal_ip").(string) {
				found = true
				d.Set("external_ip", r.GatewayNatRule.TranslatedIP)
			}
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceVcdSNATDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

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
		task, err := edgeGateway.RemoveNATMapping("SNAT", d.Get("internal_ip").(string),
			d.Get("external_ip").(string),
			"")
		if err != nil {
			return fmt.Errorf("error deleting SNAT rule: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}

	}
	return nil
}

func resourceVcdSNATUpdate(d *schema.ResourceData, meta interface{}) error {

	vcdClient := meta.(*VCDClient)

	// Update supports only when network name and network type provided
	networkName := d.Get("network_name")
	if nil == networkName || networkName.(string) == "" {
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
	natRule.GatewayNatRule.TranslatedIP = d.Get("internal_ip").(string)
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
