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

func resourceVcdAlbServiceEngineGroupAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbServiceEngineGroupAssignmentCreate,
		UpdateContext: resourceVcdAlbServiceEngineGroupAssignmentUpdate,
		ReadContext:   resourceVcdAlbServiceEngineGroupAssignmentRead,
		DeleteContext: resourceVcdAlbServiceEngineGroupAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdEdgeAlbServiceEngineGroupImport,
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
				Description: "Edge Gateway ID in which ALB Service Engine Group should be located",
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
				Description: "Number of reserved deployed virtual services for this Service Engine Group",
			},
		},
	}
}

func resourceVcdAlbServiceEngineGroupAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineAssigmentConfig := getNsxtAlbServiceEngineGroupAssignmentType(d)
	edgeAlbServiceEngineAssigment, err := vcdClient.CreateAlbServiceEngineGroupAssignment(edgeAlbServiceEngineAssigmentConfig)
	if err != nil {
		return diag.Errorf("error creating ALB Service Engine Group Assignment: %s", err)
	}

	d.SetId(edgeAlbServiceEngineAssigment.NsxtAlbServiceEngineGroupAssignment.ID)

	return resourceVcdAlbServiceEngineGroupAssignmentRead(ctx, d, meta)
}

func resourceVcdAlbServiceEngineGroupAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group Assignment: %s", err)
	}
	edgeAlbServiceEngineAssigmentConfig := getNsxtAlbServiceEngineGroupAssignmentType(d)
	// Inject correct ID for update
	edgeAlbServiceEngineAssigmentConfig.ID = edgeAlbServiceEngineAssignment.NsxtAlbServiceEngineGroupAssignment.ID
	updatedEdgeAlbServiceEngineAssigment, err := edgeAlbServiceEngineAssignment.Update(edgeAlbServiceEngineAssigmentConfig)
	if err != nil {
		return diag.Errorf("error updating ALB Service Engine Group Assignment: %s", err)
	}

	d.SetId(updatedEdgeAlbServiceEngineAssigment.NsxtAlbServiceEngineGroupAssignment.ID)

	return resourceVcdAlbServiceEngineGroupAssignmentRead(ctx, d, meta)
}

func resourceVcdAlbServiceEngineGroupAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeAlbServiceEngineAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group Assignment: %s", err)
	}
	setNsxtAlbServiceEngineGroupAssignmentData(d, edgeAlbServiceEngineAssignment.NsxtAlbServiceEngineGroupAssignment)

	d.SetId(edgeAlbServiceEngineAssignment.NsxtAlbServiceEngineGroupAssignment.ID)

	return nil
}

func resourceVcdAlbServiceEngineGroupAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group Assignment: %s", err)
	}

	err = edgeAlbServiceEngineAssignment.Delete()
	if err != nil {
		return diag.Errorf("error deleting ALB Service Engine Group Assignment: %s", err)
	}
	return nil
}

func resourceVcdEdgeAlbServiceEngineGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Service Engine Group Assignment import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name.se-group-name")
	}
	orgName, vdcName, edgeName, seGroupName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("this resource is only supported for NSX-T Edge Gateways")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	seGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentByName(seGroupName)
	if err != nil {
		return nil, fmt.Errorf("errorr retrieving Servce Engine Group Assignment with Name '%s': %s", seGroupName, err)
	}

	_ = d.Set("edge_gateway_id", edge.EdgeGateway.ID)

	d.SetId(seGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtAlbServiceEngineGroupAssignmentData(d *schema.ResourceData, t *types.NsxtAlbServiceEngineGroupAssignment) {
	_ = d.Set("edge_gateway_id", t.GatewayRef.ID)
	_ = d.Set("service_engine_group_id", t.ServiceEngineGroupRef.ID)
	_ = d.Set("max_virtual_services", t.MaxVirtualServices)
	_ = d.Set("reserved_virtual_services", t.MinVirtualServices)
	_ = d.Set("deployed_virtual_services", t.NumDeployedVirtualServices)
}

func getNsxtAlbServiceEngineGroupAssignmentType(d *schema.ResourceData) *types.NsxtAlbServiceEngineGroupAssignment {
	edgeAlbServiceEngineAssigmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: d.Get("edge_gateway_id").(string)},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: d.Get("service_engine_group_id").(string)},
	}

	if maxServicesInterface, isSet := d.GetOk("max_virtual_services"); isSet {
		edgeAlbServiceEngineAssigmentConfig.MaxVirtualServices = takeIntPointer(maxServicesInterface.(int))
	}

	if reservedServicesInterface, isSet := d.GetOk("reserved_virtual_services"); isSet {
		edgeAlbServiceEngineAssigmentConfig.MinVirtualServices = takeIntPointer(reservedServicesInterface.(int))
	}

	return edgeAlbServiceEngineAssigmentConfig
}
