package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxtManager() *schema.Resource {
	return &schema.Resource{
		Read: datasourceNsxtManagerRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T manager.",
			},
		},
	}
}

func datasourceNsxtManagerRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	nsxtManagerName := d.Get("name").(string)

	nsxtManager, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
	if err != nil {
		return fmt.Errorf("could not find NSX-T manager by name '%s': %s", nsxtManagerName, err)
	}

	if len(nsxtManager) == 0 {
		return fmt.Errorf("%s found %d NSX-T managers with name '%s'",
			govcd.ErrorEntityNotFound, len(nsxtManager), nsxtManagerName)
	}

	if len(nsxtManager) > 1 {
		return fmt.Errorf("found %d NSX-T managers with name '%s'", len(nsxtManager), nsxtManagerName)
	}

	id := extractUuid(nsxtManager[0].HREF)
	d.SetId("urn:vcloud:nsxtmanager:" + id)

	return nil
}
