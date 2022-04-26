package vcd

import (
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdVcenter() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcenterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of vCenter.",
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter version",
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter hostname",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "vCenter version",
			},
			// In UI this field is called `connection`, but it is a reserved field in Terrraform
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter connection state",
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
	setVcenterData(d, vcs[0])

	return nil
}

func setVcenterData(d *schema.ResourceData, vc *types.QueryResultVirtualCenterRecordType) {
	dSet(d, "vcenter_version", vc.VcVersion)
	// vc.Url is in format `https://XXXX.com/sdk` while UI shows hostname only so we extract it
	// The error should not be a reason to fail datasource if it is invalid so it is just logged
	host, err := url.Parse(vc.Url)
	if err != nil {
		log.Printf("[DEBUG] [vCenter read] - could not parse vCenter URL '%s': %s", vc.Url, err)
	}
	dSet(d, "vcenter_host", host.Host)
	dSet(d, "status", vc.Status)
	dSet(d, "is_enabled", vc.IsEnabled)
	dSet(d, "connection_status", vc.ListenerState)
}
