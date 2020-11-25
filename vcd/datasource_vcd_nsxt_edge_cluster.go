package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeCluster() *schema.Resource {
	return &schema.Resource{
		Read: datasourceNsxtEdgeCluster,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Edge Cluster",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NSX-T Edge Cluster",
			},
			"node_count": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of nodes in NSX-T Edge Cluster",
			},
			"node_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Node type of NSX-T Edge Cluster",
			},
			"deployment_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Deployment type of NSX-T Edge Cluster",
			},
		},
	}
}

func datasourceNsxtEdgeCluster(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	nsxtEdgeClusterName := d.Get("name").(string)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	nsxtEdgeCluster, err := vdc.GetNsxtEdgeClusterByName(nsxtEdgeClusterName)
	if err != nil {
		return fmt.Errorf("could not find NSX-T Edge Cluster by name '%s': %s", nsxtEdgeClusterName, err)
	}

	_ = d.Set("description", nsxtEdgeCluster.NsxtEdgeCluster.Description)
	_ = d.Set("node_count", nsxtEdgeCluster.NsxtEdgeCluster.NodeCount)
	_ = d.Set("node_type", nsxtEdgeCluster.NsxtEdgeCluster.NodeType)
	_ = d.Set("deployment_type", nsxtEdgeCluster.NsxtEdgeCluster.DeploymentType)

	d.SetId(nsxtEdgeCluster.NsxtEdgeCluster.ID)

	return nil
}
