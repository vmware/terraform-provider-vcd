package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdSNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdSNATCreate,
		Delete: resourceVcdSNATDelete,
		Read:   resourceVcdSNATRead,

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

			"internal_ip": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVcdSNATCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	// Multiple VCD components need to run operations on the Edge Gateway, as
	// the edge gatway will throw back an error if it is already performing an
	// operation we must wait until we can aquire a lock on the client
	lockParentEdgeGtw(d)
	defer unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// TODO add support of external network
	var orgVdcnetwork *types.OrgVDCNetwork
	providedNetworkName := d.Get("network_name")
	if nil != providedNetworkName && "" != providedNetworkName.(string) {
		orgVdcnetwork, err = getNetwork(d, vcdClient, providedNetworkName.(string))
	}
	// TODO enable when external network supported
	/*else {
		_, _ = fmt.Fprint(GetTerraformStdout(), "WARNING: this resource will require network_name and network_type in the next major version \n")
	}*/
	if err != nil {
		return fmt.Errorf("unable to find orgVdcnetwork: %s, err: %s", providedNetworkName.(string), err)
	}

	if nil != providedNetworkName && providedNetworkName != "" {
		task, err := edgeGateway.AddNATRule(orgVdcnetwork, "SNAT",
			d.Get("external_ip").(string), d.Get("internal_ip").(string))
		if err != nil {
			return fmt.Errorf("error setting SNAT rules: %#v", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
	} else {
		// TODO remove when major release is done
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

	if nil != providedNetworkName && "" != providedNetworkName {
		d.SetId(orgVdcnetwork.Name + ":" + d.Get("internal_ip").(string))
	} else {
		// TODO remove when major release is done
		d.SetId(d.Get("internal_ip").(string))
	}
	return nil
}

func resourceVcdSNATRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	e, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	var found bool

	providedNetworkName := d.Get("network_name")
	if nil != providedNetworkName && providedNetworkName.(string) != "" {
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "SNAT" && r.GatewayNatRule.TranslatedIP == d.Get("internal_ip").(string) && r.GatewayNatRule.Interface.Name == d.Get("network_name").(string) {
				found = true
				d.Set("external_ip", r.GatewayNatRule.OriginalIP)
			}
		}
	} else { // TODO remove after network_name becomes mandatory
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
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

	// Multiple VCD components need to run operations on the Edge Gateway, as
	// the edge gatway will throw back an error if it is already performing an
	// operation we must wait until we can aquire a lock on the client
	lockParentEdgeGtw(d)
	defer unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	task, err := edgeGateway.RemoveNATMapping("SNAT", d.Get("internal_ip").(string),
		d.Get("external_ip").(string),
		"")
	if err != nil {
		return fmt.Errorf("error setting SNAT rules: %#v", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}

	return nil
}
