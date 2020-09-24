package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdVcenter() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcenterRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of vCenter.",
			},
		},
	}
}

func datasourceVcenterRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	vCenterName := d.Get("name").(string)

	vcs, err := govcd.QueryVirtualCenters(vcdClient.VCDClient, "name=="+vCenterName)
	if err != nil {
		return fmt.Errorf("error occured while querying vCenters: %s", err)
	}

	if len(vcs) == 0 {
		return fmt.Errorf("%s: could not identify single vCenter. Got %d with name '%s'",
			govcd.ErrorEntityNotFound, len(vcs), vCenterName)
	}

	if len(vcs) > 1 {
		return fmt.Errorf("could not identify single vCenter. Got %d with name '%s'",
			len(vcs), vCenterName)
	}

	uuid := extractUuid(vcs[0].HREF)
	urn, err := govcd.BuildUrnWithUuid("urn:vcloud:vimserver:", uuid)
	if err != nil {
		return fmt.Errorf("could not build URN for ID '%s': %s", uuid, err)
	}

	d.SetId(urn)

	return nil
}
