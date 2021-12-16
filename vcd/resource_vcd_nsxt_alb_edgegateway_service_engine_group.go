package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbEdgeGatewayServiceEngineGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupCreate,
		UpdateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupUpdate,
		ReadContext:   resourceVcdAlbEdgeGatewayServiceEngineGroupRead,
		DeleteContext: resourceVcdAlbEdgeGatewayServiceEngineGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupImport,
		},

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
				Description: "Service Engine Group ID to attach to this NSX-T Edge Gateway",
			},
			"service_engine_group_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service Engine Group Name which is attached to NSX-T Edge Gateway",
			},
			"max_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of virtual services to be used in this Service Engine Group",
			},
			"reserved_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Number of reserved virtual services for this Service Engine Group",
			},
			"deployed_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of deployed virtual services for this Service Engine Group",
			},
		},
	}
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignmentConfig := getAlbServiceEngineGroupAssignmentType(d)
	edgeAlbServiceEngineGroupAssignment, err := vcdClient.CreateAlbServiceEngineGroupAssignment(edgeAlbServiceEngineGroupAssignmentConfig)
	if err != nil {
		return diag.Errorf("error creating ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	d.SetId(edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return resourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}
	edgeAlbServiceEngineGroupAssignmentConfig := getAlbServiceEngineGroupAssignmentType(d)
	// Add correct ID for update
	edgeAlbServiceEngineGroupAssignmentConfig.ID = edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID
	updatedEdgeAlbServiceEngineGroupAssignment, err := edgeAlbServiceEngineGroupAssignment.Update(edgeAlbServiceEngineGroupAssignmentConfig)
	if err != nil {
		return diag.Errorf("error updating ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	d.SetId(updatedEdgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return resourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group assignment: %s", err)
	}
	setAlbServiceEngineGroupAssignmentData(d, edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment)
	d.SetId(edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return nil
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	err = edgeAlbServiceEngineGroupAssignment.Delete()
	if err != nil {
		return diag.Errorf("error deleting ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}
	return nil
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Service Engine Group assignment import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name.se-group-name")
	}
	orgName, vdcName, edgeName, seGroupName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("this resource is only supported for NSX-T Edge Gateways")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	seGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentByName(seGroupName)
	if err != nil {
		return nil, fmt.Errorf("errorr retrieving Servce Engine Group assignment to Edge Gateway with Name '%s': %s",
			seGroupName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(seGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return []*schema.ResourceData{d}, nil
}

func setAlbServiceEngineGroupAssignmentData(d *schema.ResourceData, t *types.NsxtAlbServiceEngineGroupAssignment) {
	dSet(d, "edge_gateway_id", t.GatewayRef.ID)
	dSet(d, "service_engine_group_id", t.ServiceEngineGroupRef.ID)
	dSet(d, "service_engine_group_name", t.ServiceEngineGroupRef.Name)
	dSet(d, "max_virtual_services", t.MaxVirtualServices)
	dSet(d, "reserved_virtual_services", t.MinVirtualServices)
	dSet(d, "deployed_virtual_services", t.NumDeployedVirtualServices)
}

func getAlbServiceEngineGroupAssignmentType(d *schema.ResourceData) *types.NsxtAlbServiceEngineGroupAssignment {
	edgeAlbServiceEngineAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: d.Get("edge_gateway_id").(string)},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: d.Get("service_engine_group_id").(string)},
	}

	// Max Virtual Services and Reserved Virtual Services only work with SHARED Service Engine Group, but validation
	// enforcement is left for VCD API.
	if maxServicesInterface, isSet := d.GetOk("max_virtual_services"); isSet {
		edgeAlbServiceEngineAssignmentConfig.MaxVirtualServices = takeIntPointer(maxServicesInterface.(int))
	}

	if reservedServicesInterface, isSet := d.GetOk("reserved_virtual_services"); isSet {
		edgeAlbServiceEngineAssignmentConfig.MinVirtualServices = takeIntPointer(reservedServicesInterface.(int))
	}

	return edgeAlbServiceEngineAssignmentConfig
}

// validateEdgeGatewayIdParent validates if specified field `edge_gateway_id` exists in defined Org and VDC
func validateEdgeGatewayIdParent(d *schema.ResourceData, vcdClient *VCDClient) error {
	org, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf("error retrieving Org and VDC")
	}

	_, err = vcdClient.GetNsxtEdgeGatewayFromResourceById(d, "edge_gateway_id")
	if err != nil {
		return fmt.Errorf("unable to locate NSX-T Edge Gateway with ID '%s' in Org '%s' and VDC '%s': %s",
			d.Get("edge_gateway_id").(string), org.Org.Name, vdc.Vdc.Name, err)
	}

	return nil
}
