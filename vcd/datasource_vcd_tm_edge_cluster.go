package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmEdgeCluster = "TM Edge Cluster"
const labelTmEdgeClusterSync = "TM Edge Cluster Sync"

func datasourceVcdTmEdgeCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceVcdTmEdgeClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name %s", labelTmEdgeCluster),
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Region ID of  %s", labelTmEdgeCluster),
			},
			"sync_before_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf(" %s", labelTmEdgeCluster),
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Node count in %s", labelTmEdgeCluster),
			},
			"org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Org count %s", labelTmEdgeCluster),
			},
			"vpc_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("VPC count %s", labelTmEdgeCluster),
			},
			"average_cpu_usage_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: fmt.Sprintf("Average CPU Usage percentage of %s ", labelTmEdgeCluster),
			},
			"average_memory_usage_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: fmt.Sprintf("Average Memory Usage percentage of %s ", labelTmEdgeCluster),
			},
			"health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Health status of %s", labelTmEdgeCluster),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelTmEdgeCluster),
			},
			"deployment_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Deployment type of %s", labelTmEdgeCluster),
			},
		},
	}
}

func resourceVcdTmEdgeClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if d.Get("sync_before_read").(bool) {
		err := vcdClient.TmSyncEdgeClusters()
		if err != nil {
			return diag.Errorf("error syncing %s before lookup: %s", labelTmEdgeClusterSync, err)
		}
	}

	c := dsReadConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:    labelTmEdgeCluster,
		getEntityFunc:  vcdClient.GetTmEdgeClusterByName,
		stateStoreFunc: setTmEdgeClusterData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setTmEdgeClusterData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmEdgeCluster) error {
	if t == nil || t.TmEdgeCluster == nil {
		return fmt.Errorf("empty %s received", labelTmEdgeCluster)
	}

	d.SetId(t.TmEdgeCluster.ID)
	dSet(d, "status", t.TmEdgeCluster.Status)
	dSet(d, "health_status", t.TmEdgeCluster.HealthStatus)
	dSet(d, "region_id", t.TmEdgeCluster.RegionRef.ID)
	dSet(d, "deployment_type", t.TmEdgeCluster.DeploymentType)
	dSet(d, "node_count", t.TmEdgeCluster.NodeCount)
	dSet(d, "org_count", t.TmEdgeCluster.OrgCount)
	dSet(d, "vpc_count", t.TmEdgeCluster.VpcCount)
	dSet(d, "average_cpu_usage_percentage", t.TmEdgeCluster.AvgCPUUsagePercentage)
	dSet(d, "average_memory_usage_percentage", t.TmEdgeCluster.AvgMemoryUsagePercentage)

	return nil
}
