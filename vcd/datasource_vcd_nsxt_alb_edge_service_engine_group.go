package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbServiceEngineGroupAssignment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbServiceEngineGroupAssignmentRead,

		Schema: map[string]*schema.Schema{
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
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which ALB Service Engine Group should be located",
			},
			"service_engine_group_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which ALB Service Engine Group should be located",
			},
			"max_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of virtual services to be used in this Service Engine Group",
			},
			"reserved_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of reserved virtual services for this Service Engine Group",
			},
			"deployed_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of reserved deployed virtual services for this Service Engine Group",
			},
		},
	}
}

func datasourceVcdAlbServiceEngineGroupAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeGatewayId := d.Get("edge_gateway_id").(string)
	serviceEngineGroupId := d.Get("service_engine_group_id")

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s;serviceEngineGroupRef.id==%s", edgeGatewayId, serviceEngineGroupId))

	edgeAlbServiceEngineAssignments, err := vcdClient.GetAllAlbServiceEngineGroupAssignments(queryParams)
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group Assignment: %s", err)
	}

	if len(edgeAlbServiceEngineAssignments) == 0 {
		return diag.Errorf("%s", govcd.ErrorEntityNotFound)
	}

	if len(edgeAlbServiceEngineAssignments) > 1 {
		return diag.Errorf("more than one Service Engine Group assignment found (%d)", len(edgeAlbServiceEngineAssignments))
	}

	setNsxtAlbServiceEngineGroupAssignmentData(d, edgeAlbServiceEngineAssignments[0].NsxtAlbServiceEngineGroupAssignment)

	d.SetId(edgeAlbServiceEngineAssignments[0].NsxtAlbServiceEngineGroupAssignment.ID)

	return nil
}
