package vcd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmEdgeClusterQos = "TM Edge Cluster QoS"

func resourceVcdTmEdgeClusterQos() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmEdgeClusterQosCreate,
		ReadContext:   resourceVcdTmEdgeClusterQosRead,
		UpdateContext: resourceVcdTmEdgeClusterQosUpdate,
		DeleteContext: resourceVcdTmEdgeClusterQosDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmEdgeClusterQosImport,
		},

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
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelTmEdgeCluster),
				ValidateFunc: emptyOrIsIntAndAtLeast(1),
				RequiredWith: []string{"ingress_burst_size_bytes"},
			},
			"ingress_burst_size_bytes": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("Ingress burst size bytes for %s", labelTmEdgeCluster),
				ValidateFunc: emptyOrIsIntAndAtLeast(1),
				RequiredWith: []string{"ingress_committed_bandwidth_mbps"},
			},
			"egress_committed_bandwidth_mbps": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelTmEdgeCluster),
				ValidateFunc: emptyOrIsIntAndAtLeast(1),
				RequiredWith: []string{"egress_burst_size_bytes"},
			},
			"egress_burst_size_bytes": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("Ingress burst size bytes for %s", labelTmEdgeCluster),
				ValidateFunc: emptyOrIsIntAndAtLeast(1),
				RequiredWith: []string{"egress_committed_bandwidth_mbps"},
			},
		},
	}
}

func resourceVcdTmEdgeClusterQosCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// The Edge Cluster must already be existing that is handled by 'vcd_tm_edge_cluster'
	// This is not a "real" entity creation, rather a lookup of existing one
	lookupEdgeClusterById := func(_ *types.TmEdgeCluster) (*govcd.TmEdgeCluster, error) {
		return vcdClient.GetTmEdgeClusterById(d.Get("edge_cluster_id").(string))
	}

	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:      labelTmEdgeClusterQos,
		getTypeFunc:      getTmEdgeClusterQosType,
		stateStoreFunc:   setTmEdgeClusterQosData,
		createFunc:       lookupEdgeClusterById,
		resourceReadFunc: resourceVcdTmEdgeClusterQosRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmEdgeClusterQosUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:      labelTmEdgeClusterQos,
		getTypeFunc:      getTmEdgeClusterQosType,
		getEntityFunc:    vcdClient.GetTmEdgeClusterById,
		resourceReadFunc: resourceVcdTmEdgeClusterQosRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmEdgeClusterQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:    labelTmEdgeClusterQos,
		getEntityFunc:  vcdClient.GetTmEdgeClusterById,
		stateStoreFunc: setTmEdgeClusterQosData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmEdgeClusterQosDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:   labelTmEdgeClusterQos,
		getEntityFunc: vcdClient.GetTmEdgeClusterById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmEdgeClusterQosImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	ec, err := vcdClient.GetTmEdgeClusterByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by Name '%s': %s", labelTmEdgeClusterQos, d.Id(), err)
	}

	d.SetId(ec.TmEdgeCluster.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmEdgeClusterQosType(vcdClient *VCDClient, d *schema.ResourceData) (*types.TmEdgeCluster, error) {
	// Only the QoS configuration is updatable, everything else is read-only
	t := &types.TmEdgeCluster{
		DefaultQosConfig: types.TmEdgeClusterDefaultQosConfig{
			// IngressProfile: &types.TmEdgeClusterQosProfile{
			// 	CommittedBandwidthMbps: mustStrToInt(d.Get("ingress_committed_bandwidth_mbps").(string)),
			// 	BurstSizeBytes:         mustStrToInt(d.Get("ingress_burst_size_bytes").(string)),
			// 	Type:                   "DEFAULT",
			// },
			// EgressProfile: &types.TmEdgeClusterQosProfile{
			// 	CommittedBandwidthMbps: mustStrToInt(d.Get("egress_committed_bandwidth_mbps").(string)),
			// 	BurstSizeBytes:         mustStrToInt(d.Get("egress_burst_size_bytes").(string)),
			// 	Type:                   "DEFAULT",
			// },
		},
	}

	// Ingress setup
	// Only initialize IngressProfile type if at least one of the fields is set
	ingressCommittedBandwidthMbps := d.Get("ingress_committed_bandwidth_mbps").(string)
	ingressBurstSizeBytes := d.Get("ingress_burst_size_bytes").(string)
	if ingressCommittedBandwidthMbps != "" || ingressBurstSizeBytes != "" {
		t.DefaultQosConfig.IngressProfile = &types.TmEdgeClusterQosProfile{Type: "DEFAULT"}

		if ingressCommittedBandwidthMbps != "" {
			intIngressCommittedBandwidthMbps := mustStrToInt(ingressCommittedBandwidthMbps)
			t.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps = intIngressCommittedBandwidthMbps
		}

		if ingressBurstSizeBytes != "" {
			intIngressBurstSizeBytes := mustStrToInt(ingressBurstSizeBytes)
			t.DefaultQosConfig.IngressProfile.BurstSizeBytes = intIngressBurstSizeBytes
		}
	}

	// Egress setup
	// Only initialize EgressProfile type if at least one of the fields is set
	egressCommittedBandwidthMbps := d.Get("egress_committed_bandwidth_mbps").(string)
	egressBurstSizeBytes := d.Get("egress_burst_size_bytes").(string)
	if egressCommittedBandwidthMbps != "" || egressBurstSizeBytes != "" {
		t.DefaultQosConfig.EgressProfile = &types.TmEdgeClusterQosProfile{Type: "DEFAULT"}

		if egressCommittedBandwidthMbps != "" {
			integressCommittedBandwidthMbps := mustStrToInt(egressCommittedBandwidthMbps)
			t.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps = integressCommittedBandwidthMbps
		}

		if egressBurstSizeBytes != "" {
			integressBurstSizeBytes := mustStrToInt(egressBurstSizeBytes)
			t.DefaultQosConfig.EgressProfile.BurstSizeBytes = integressBurstSizeBytes
		}
	}

	return t, nil
}

func setTmEdgeClusterQosData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmEdgeCluster) error {
	if t == nil || t.TmEdgeCluster == nil {
		return fmt.Errorf("empty %s received", labelTmEdgeCluster)
	}

	d.SetId(t.TmEdgeCluster.ID)
	dSet(d, "edge_cluster_id", t.TmEdgeCluster.ID)

	dSet(d, "region_id", "")
	if t.TmEdgeCluster.RegionRef != nil {
		dSet(d, "region_id", t.TmEdgeCluster.RegionRef.ID)
	}

	dSet(d, "ingress_committed_bandwidth_mbps", nil)
	if t.TmEdgeCluster.DefaultQosConfig.IngressProfile != nil {
		strValue := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)
		dSet(d, "ingress_committed_bandwidth_mbps", strValue)
	}
	dSet(d, "ingress_burst_size_bytes", nil)
	if t.TmEdgeCluster.DefaultQosConfig.IngressProfile != nil {
		strValue := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes)
		dSet(d, "ingress_burst_size_bytes", strValue)
	}

	dSet(d, "egress_committed_bandwidth_mbps", nil)
	if t.TmEdgeCluster.DefaultQosConfig.EgressProfile != nil {
		strValue := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)
		dSet(d, "egress_committed_bandwidth_mbps", strValue)
	}

	dSet(d, "egress_burst_size_bytes", nil)
	if t.TmEdgeCluster.DefaultQosConfig.EgressProfile != nil {
		strValue := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes)
		dSet(d, "ingress_burst_size_bytes", strValue)
	}

	return nil
}
