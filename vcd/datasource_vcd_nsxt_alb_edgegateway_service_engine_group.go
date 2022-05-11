package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbEdgeGatewayServiceEngineGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbEdgeGatewayServiceEngineGroupRead,

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
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which ALB Service Engine Group should be located",
			},
			// Following ID reference practice there could be `service_engine_group_id` field used to reference outcome
			// of `vcd_nsxt_alb_service_engine_group` resource or data source. However - using the mentioned resource or
			// data source would require Provider level access and this would break workflow when tenant needs to
			// reference service engine group ID in `vcd_nsxt_alb_virtual_service`. Because of that
			// `service_engine_group_id` and `service_engine_group_name` are both supported but only one required.
			"service_engine_group_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				Description:  "Service Engine Group Name which is attached to NSX-T Edge Gateway",
				ExactlyOneOf: []string{"service_engine_group_name", "service_engine_group_id"},
			},
			"service_engine_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				Description:  "Service Engine Group ID which is attached to NSX-T Edge Gateway",
				ExactlyOneOf: []string{"service_engine_group_name", "service_engine_group_id"},
			},
			"max_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of virtual services to be used in this Service Engine Group",
			},
			"reserved_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of reserved virtual services for this Service Engine Group",
			},
			"deployed_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of reserved deployed virtual services for this Service Engine Group",
			},
		},
	}
}

func datasourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeGatewayId := d.Get("edge_gateway_id").(string)
	serviceEngineGroupName := d.Get("service_engine_group_name").(string)
	serviceEngineGroupId := d.Get("service_engine_group_id").(string)

	var err error
	var edgeAlbServiceEngineAssignment *govcd.NsxtAlbServiceEngineGroupAssignment

	switch {
	// When `service_engine_group_name` lookup field is presented
	case serviceEngineGroupName != "":
		// This will filter service engine groups by assigned NSX-T Edge Gateway ID and additionally filter by Name on client
		// side
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s", edgeGatewayId))
		edgeAlbServiceEngineAssignment, err = vcdClient.GetFilteredAlbServiceEngineGroupAssignmentByName(serviceEngineGroupName, queryParams)
		if err != nil {
			return diag.Errorf("error retrieving Service Engine Group assignment to NSX-T Edge Gateway: %s", err)
		}
	// When `id` lookup field is presented
	case serviceEngineGroupId != "":
		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s;serviceEngineGroupRef.id==%s", edgeGatewayId, serviceEngineGroupId))

		edgeAlbServiceEngineAssignments, err := vcdClient.GetAllAlbServiceEngineGroupAssignments(queryParams)
		if err != nil {
			return diag.Errorf("error reading ALB Service Engine Group assignment to Edge Gateway: %s", err)
		}

		if len(edgeAlbServiceEngineAssignments) == 0 {
			return diag.FromErr(govcd.ErrorEntityNotFound)
		}

		if len(edgeAlbServiceEngineAssignments) > 1 {
			return diag.Errorf("more than one Service Engine Group assignment to Edge Gateway found (%d)", len(edgeAlbServiceEngineAssignments))
		}

		// Exactly one Service Engine Group assignment is found
		edgeAlbServiceEngineAssignment = edgeAlbServiceEngineAssignments[0]
	default:
		return diag.Errorf("Name or ID must be specified for Service Engine Group assignment data source")
	}

	setAlbServiceEngineGroupAssignmentData(d, edgeAlbServiceEngineAssignment.NsxtAlbServiceEngineGroupAssignment)
	d.SetId(edgeAlbServiceEngineAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return nil
}
