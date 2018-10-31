package vcd

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

// DEPRECATED: use vcd_network_routed instead
func resourceVcdNetwork() *schema.Resource {
	newRes := resourceVcdNetworkRouted()
	// DeprecationMessage requires a version of Terraform newer than what
	// we currently use in the vendor directory
	// newRes.DeprecationMessage = "Deprecated. Use vcd_network_routed instead",
	return newRes
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
