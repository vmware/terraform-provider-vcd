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

	vcUuid, err := govcd.GetUuidFromHref(vcs[0].HREF, true)
	if err != nil {
		return fmt.Errorf("error getting UUID from HREF '%s'", vcs[0].HREF)
	}
	d.SetId("urn:vcloud:vimserver:" + vcUuid)

	return nil
}
