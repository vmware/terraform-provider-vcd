package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmEdgeClusterQos() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmEdgeClusterQosRead,

		Schema: map[string]*schema.Schema{
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of %s", labelTmEdgeCluster),
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Region ID of  %s", labelTmEdgeCluster),
			},
			"ingress_committed_bandwidth_mbps": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelTmEdgeCluster),
			},
			"ingress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelTmEdgeCluster),
			},
			"egress_committed_bandwidth_mbps": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelTmEdgeCluster),
			},
			"egress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Egress burst size bytes for %s", labelTmEdgeCluster),
			},
		},
	}
}

func datasourceVcdTmEdgeClusterQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := dsReadConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:              labelTmEdgeClusterQos,
		stateStoreFunc:           setTmEdgeClusterQosData,
		overrideDefaultNameField: "edge_cluster_id", // pass the value of this field to getEntityFunc
		getEntityFunc:            vcdClient.GetTmEdgeClusterById,
	}
	return readDatasource(ctx, d, meta, c)
}
