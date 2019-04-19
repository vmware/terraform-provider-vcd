package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

	// Creating a loop to offer further protection from the edge gateway erroring
	// due to being busy eg another person is using another client so wouldn't be
	// constrained by out lock. If the edge gateway reurns with a busy error, wait
	// 3 seconds and then try again. Continue until a non-busy error or success

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	var orgVdcnetwork *types.OrgVDCNetwork
	providedNetworkName := d.Get("network_name")
	if nil != providedNetworkName && "" != providedNetworkName {
		orgVdcnetwork, err = getNetwork(d, vcdClient, providedNetworkName.(string))
	}
	if err != nil {
		return fmt.Errorf("unable to find orgVdcnetwork: %s, err: %s", providedNetworkName.(string), err)
	}

	if nil != providedNetworkName && providedNetworkName != "" {
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {

			task, err := edgeGateway.AddNATPortMappingWithUplink(orgVdcnetwork, "SNAT",
				d.Get("external_ip").(string), "any", d.Get("internal_ip").(string),
				"any", "any", "")
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error setting SNAT rules: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
	} else {
		// TODO remove when major release is done
		err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
			task, err := edgeGateway.AddNATMapping("SNAT", d.Get("internal_ip").(string),
				d.Get("external_ip").(string))
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error setting SNAT rules: %#v", err))
			}
			return resource.RetryableError(task.WaitTaskCompletion())
		})
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
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
	if nil != providedNetworkName && providedNetworkName != "" {
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "SNAT" && r.GatewayNatRule.OriginalIP == d.Id() {
				found = true
				d.Set("external_ip", r.GatewayNatRule.TranslatedIP)
			}
		}
	} else {
		for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
			if r.RuleType == "SNAT" && r.GatewayNatRule.OriginalIP == d.Id() && r.GatewayNatRule.Interface.Name == d.Get("network_name").(string) {
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
	vcdClient.Mutex.Lock()
	defer vcdClient.Mutex.Unlock()

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {
		task, err := edgeGateway.RemoveNATMapping("SNAT", d.Get("internal_ip").(string),
			d.Get("external_ip").(string),
			"")
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error setting SNAT rules: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return err
	}

	return nil
}
