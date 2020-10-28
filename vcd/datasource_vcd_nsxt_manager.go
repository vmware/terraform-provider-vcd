package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	nsxtManagers, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
	if err != nil {
		return fmt.Errorf("could not find NSX-T manager by name '%s': %s", nsxtManagerName, err)
	}

	if len(nsxtManagers) == 0 {
		return fmt.Errorf("%s found %d NSX-T managers with name '%s'",
			govcd.ErrorEntityNotFound, len(nsxtManagers), nsxtManagerName)
	}

	if len(nsxtManagers) > 1 {
		return fmt.Errorf("found %d NSX-T managers with name '%s'", len(nsxtManagers), nsxtManagerName)
	}

	// We try to keep IDs clean
	id := extractUuid(nsxtManagers[0].HREF)
	urn, err := govcd.BuildUrnWithUuid("urn:vcloud:nsxtmanager:", id)
	if err != nil {
		return fmt.Errorf("could not construct URN from id '%s': %s", id, err)
	}
	d.SetId(urn)

	return nil
}
