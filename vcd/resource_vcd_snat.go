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
		DeprecationMessage: "This resource will require network_name in the next major version",
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
	if networkName := d.Get("network_name").(string); networkName != "" {
		orgVdcnetwork, err = getNetwork(d, vcdClient, networkName)
	}
	if orgVdcnetwork == nil || orgVdcnetwork == (&types.OrgVDCNetwork{}) {
		return fmt.Errorf("unable to find orgVdcnetwork: %s, err: %s", d.Get("network_name").(string), err)
	}

	err = retryCall(vcdClient.MaxRetryTimeout, func() *resource.RetryError {

		task, err := edgeGateway.AddNATPortMappingWithUplink(orgVdcnetwork, "SNAT",
			d.Get("external_ip").(string), "any", d.Get("internal_ip").(string),
			"any", "any", "")
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error setting SNAT rules: %#v", err))
		}
		return resource.RetryableError(task.WaitTaskCompletion())
	})
	if err != nil {
		return err
	}

	d.SetId(orgVdcnetwork.Name + ":" + d.Get("internal_ip").(string))
	return nil
}

func resourceVcdSNATRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	e, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	var found bool

	for _, r := range e.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
		if r.RuleType == "SNAT" && r.GatewayNatRule.Interface.Name == d.Get("network_name") &&
			r.GatewayNatRule.OriginalIP == d.Id() {
			found = true
			d.Set("external_ip", r.GatewayNatRule.TranslatedIP)
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
