package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNetworkDirectCreate,
		Read:   resourceVcdNetworkRead,
		Delete: resourceVcdNetworkDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},

			"external_network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"href": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"shared": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceVcdNetworkDirectCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	extNetName, ok := d.GetOk("external_network")
	externalNetworkName := ""
	var externalNetwork types.ExternalNetworkReference
	if ok {
		externalNetworkName = extNetName.(string)
		network, err := govcd.GetExternalNetworkByName(vcdClient.VCDClient, externalNetworkName)
		if err == nil {
			externalNetwork = *network
		} else {
			return fmt.Errorf("unable to find external network %s (%s)", externalNetworkName, err)
		}
	} // If no external network is provided, Terraform should fail the resource

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		Name:  d.Get("name").(string),
		Configuration: &types.NetworkConfiguration{
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.HREF,
				Type: externalNetwork.Type,
				Name: externalNetwork.Name,
			},
			FenceMode:                 "bridged",
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return fmt.Errorf("error: %#v", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceVcdNetworkRead(d, meta)
}
