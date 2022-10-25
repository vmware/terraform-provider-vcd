package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtEdgeCluster,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				ConflictsWith: []string{"vdc_id", "vdc_group_id", "provider_vdc_id"},
				Deprecated:    "This field is deprecated in favor of 'owner_id' which accepts IDs of VDC, VDC Group and Provider VDC",
			},
			"vdc_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of VDC, VDC Group or Provider VDC",
				ConflictsWith: []string{"vdc", "vdc_group_id", "provider_vdc_id"},
			},
			"vdc_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of VDC, VDC Group or Provider VDC",
				ConflictsWith: []string{"vdc", "vdc_id", "provider_vdc_id"},
			},
			"provider_vdc_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of VDC, VDC Group or Provider VDC",
				ConflictsWith: []string{"vdc", "vdc_id", "vdc_group_id"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Edge Cluster",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NSX-T Edge Cluster",
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of nodes in NSX-T Edge Cluster",
			},
			"node_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Node type of NSX-T Edge Cluster",
			},
			"deployment_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Deployment type of NSX-T Edge Cluster",
			},
		},
	}
}

func datasourceNsxtEdgeCluster(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	nsxtEdgeClusterName := d.Get("name").(string)

	vdcId := d.Get("vdc_id").(string)
	vdcGroupId := d.Get("vdc_group_id").(string)
	pVdcId := d.Get("provider_vdc_id").(string)

	// Using Raw filter query parameters instead of SDK based functions because filtering is
	// conditional based on filter type.
	queryParams := url.Values{}
	switch {
	case vdcId != "":
		queryParams.Add("filter", fmt.Sprintf("orgVdcId==%s", vdcId))
	case vdcGroupId != "":
		queryParams.Add("filter", fmt.Sprintf("vdcGroupId==%s", vdcGroupId))
	case pVdcId != "":
		queryParams.Add("filter", fmt.Sprintf("pvdcId==%s", pVdcId))
	default:
		// The original filtering option
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf(errorRetrievingOrgAndVdc, err)
		}
		queryParams.Add("filter", fmt.Sprintf("orgVdcId==%s", vdc.Vdc.ID))
	}

	allEdgeClusters, err := vcdClient.GetAllNsxtEdgeClusters(queryParams)
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Clusters: %s", err)
	}

	nameFilteredNsxtEdgeClusters := filterNsxtEdgeClusters(nsxtEdgeClusterName, allEdgeClusters)

	if len(nameFilteredNsxtEdgeClusters) == 0 {
		return diag.Errorf("%s: no NSX-T Edge Cluster with name '%s' and filter '%s'", govcd.ErrorEntityNotFound, nsxtEdgeClusterName, queryParams.Get("filter"))
	}

	if len(nameFilteredNsxtEdgeClusters) > 1 {
		return diag.Errorf("more than one (%d) NSX-T Edge Cluster with name '%s' and filter '%s'", len(nameFilteredNsxtEdgeClusters), nsxtEdgeClusterName, queryParams.Get("filter"))
	}

	dSet(d, "description", nameFilteredNsxtEdgeClusters[0].NsxtEdgeCluster.Description)
	dSet(d, "node_count", nameFilteredNsxtEdgeClusters[0].NsxtEdgeCluster.NodeCount)
	dSet(d, "node_type", nameFilteredNsxtEdgeClusters[0].NsxtEdgeCluster.NodeType)
	dSet(d, "deployment_type", nameFilteredNsxtEdgeClusters[0].NsxtEdgeCluster.DeploymentType)

	d.SetId(nameFilteredNsxtEdgeClusters[0].NsxtEdgeCluster.ID)

	return nil
}

func filterNsxtEdgeClusters(name string, allNnsxtEdgeCluster []*govcd.NsxtEdgeCluster) []*govcd.NsxtEdgeCluster {
	filteredNsxtEdgeClusters := make([]*govcd.NsxtEdgeCluster, 0)
	for index, nsxtEdgeCluster := range allNnsxtEdgeCluster {
		if allNnsxtEdgeCluster[index].NsxtEdgeCluster.Name == name {
			filteredNsxtEdgeClusters = append(filteredNsxtEdgeClusters, nsxtEdgeCluster)
		}
	}

	return filteredNsxtEdgeClusters
}
