package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtIpSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtIpSetCreate,
		ReadContext:   resourceVcdNsxtIpSetRead,
		UpdateContext: resourceVcdNsxtIpSetUpdate,
		DeleteContext: resourceVcdNsxtIpSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtIpSetImport,
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
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Deprecated in favor of `edge_gateway_id`. IP Sets will inherit VDC from parent Edge Gateway.",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of VDC or VDC Group",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP Set name",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Edge Gateway name in which IP Set is located",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP Set description",
			},
			"ip_addresses": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of IP address, CIDR, IP range objects",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdNsxtIpSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "create")
	if err != nil {
		return diag.FromErr(err)
	}

	var ipSet *types.NsxtFirewallGroup
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
		ipSet = getNsxtIpSetType(d, parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
		ipSet = getNsxtIpSetType(d, d.Get("edge_gateway_id").(string))
	}

	createdFwGroup, err := nsxtEdgeGateway.CreateNsxtFirewallGroup(ipSet)
	if err != nil {
		return diag.Errorf("[nsxt ip set create] error creating NSX-T IP Set '%s': %s", ipSet.Name, err)
	}

	dSet(d, "edge_gateway_id", nsxtEdgeGateway.EdgeGateway.ID)
	d.SetId(createdFwGroup.NsxtFirewallGroup.ID)

	return resourceVcdNsxtIpSetRead(ctx, d, meta)
}

func getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient *VCDClient, d *schema.ResourceData, action string) (string, *govcd.NsxtEdgeGateway, error) {
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return "", nil, fmt.Errorf("[nsxt ip set create] error retrieving Org: %s", err)
	}

	// Lookup Edge Gateway to know parent VDC or VDC Group
	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return "", nil, fmt.Errorf("[nsxt ip set %s] error retrieving Edge Gateway structure: %s", action, err)
	}
	if anyEdgeGateway.IsNsxv() {
		return "", nil, fmt.Errorf("[nsxt ip set %s] NSX-V edge gateway not supported", action)
	}

	nsxtEdgeGateway, err := anyEdgeGateway.GetNsxtEdgeGateway()
	if err != nil {
		return "", nil, fmt.Errorf("[nsxt ip set %s] could not retrieve NSX-T Edge Gateway with ID '%s': %s", action, d.Id(), err)
	}

	return anyEdgeGateway.EdgeGateway.OwnerRef.ID, nsxtEdgeGateway, nil
}

func resourceVcdNsxtIpSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "update")
	if err != nil {
		return diag.FromErr(err)
	}

	var updateIpSet *types.NsxtFirewallGroup
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
		updateIpSet = getNsxtIpSetType(d, parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
		updateIpSet = getNsxtIpSetType(d, d.Get("edge_gateway_id").(string))
	}

	ipSet, err := nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt ip set update] error getting NSX-T IP Set: %s", err)
	}

	// Inject existing ID for update
	updateIpSet.ID = d.Id()

	_, err = ipSet.Update(updateIpSet)
	if err != nil {
		return diag.Errorf("[nsxt ip set update] error updating NSX-T IP Set '%s': %s", ipSet.NsxtFirewallGroup.Name, err)
	}

	return resourceVcdNsxtIpSetRead(ctx, d, meta)
}

func resourceVcdNsxtIpSetRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt ip set create] error retrieving Org: %s", err)
	}

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "read")
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var ipSet *govcd.NsxtFirewallGroup
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vdcGroup, err := adminOrg.GetVdcGroupById(parentEdgeGatewayOwnerId)
		if err != nil {
			return diag.Errorf("[nsxt ip set resource read] error finding VDC Group by id '%s': %s", parentEdgeGatewayOwnerId, err)
		}

		ipSet, err = vdcGroup.GetNsxtFirewallGroupById(d.Id())
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("[nsxt ip set resource read] error getting NSX-T IP Set with ID '%s': %s", d.Id(), err)
		}
	} else {
		ipSet, err = nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("[nsxt ip set resource read] error getting NSX-T IP Set with ID '%s': %s", d.Id(), err)
		}
	}

	err = setNsxtIpSetData(d, ipSet.NsxtFirewallGroup, parentEdgeGatewayOwnerId)
	if err != nil {
		return diag.Errorf("[nsxt ip set resource read] error setting NSX-T IP Set: %s", err)
	}

	return nil
}

func resourceVcdNsxtIpSetDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "delete")
	if err != nil {
		return diag.FromErr(err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	ipSet, err := nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt ip set resource delete] error getting NSX-T IP Set: %s", err)
	}

	err = ipSet.Delete()
	if err != nil {
		return diag.Errorf("[nsxt ip set resource delete] error deleting NSX-T IP Set: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtIpSetImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.ip_set_name or" +
			"as org-name.vdc-group-name.edge_gateway_name.ip_set_name")
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, ipSetName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	// define an interface type to match VDC and VDC Groups
	var vdcOrVdcGroup vdcOrVdcGroupHandler
	var adminOrg *govcd.AdminOrg
	_, vdcOrVdcGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcOrVdcGroupName)
	if govcd.ContainsNotFound(err) {
		adminOrg, err = vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("[nsxt ip set resource import] error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrVdcGroup, err = adminOrg.GetVdcGroupByName(vdcOrVdcGroupName)
		if err != nil {
			return nil, fmt.Errorf("[nsxt ip set resource import] error finding VDC or VDC Group by name '%s': %s", vdcOrVdcGroupName, err)
		}
	}

	// Lookup Edge Gateway to know parent VDC or VDC Group
	nsxtEdgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("[nsxt ip set import] error retrieving Edge Gateway structure: %s", err)
	}

	var ipSet *govcd.NsxtFirewallGroup
	parentEdgeGatewayOwnerId := nsxtEdgeGateway.EdgeGateway.OwnerRef.ID
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vdcGroup, err := adminOrg.GetVdcGroupById(parentEdgeGatewayOwnerId)
		if err != nil {
			return nil, fmt.Errorf("[nsxt ip set resource import] error finding VDC Group by id '%s': %s", parentEdgeGatewayOwnerId, err)
		}

		ipSet, err = vdcGroup.GetNsxtFirewallGroupByName(ipSetName, types.FirewallGroupTypeIpSet)
		if err != nil {
			return nil, fmt.Errorf("[nsxt ip set resource import] error getting NSX-T IP Set with ID '%s': %s", d.Id(), err)
		}
	} else {
		ipSet, err = nsxtEdgeGateway.GetNsxtFirewallGroupByName(ipSetName, types.FirewallGroupTypeIpSet)
		if err != nil {
			return nil, fmt.Errorf("[nsxt ip set resource import] unable to find IP Set '%s': %s", ipSetName, err)
		}
	}

	if !ipSet.IsIpSet() {
		return nil, fmt.Errorf("[nsxt ip set resource import] firewall group '%s' is not a IP Set, but '%s'",
			ipSet.NsxtFirewallGroup.Name, ipSet.NsxtFirewallGroup.Type)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", nsxtEdgeGateway.EdgeGateway.ID)
	d.SetId(ipSet.NsxtFirewallGroup.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtIpSetData(d *schema.ResourceData, ipSetType *types.NsxtFirewallGroup, parentVdcOrVdcGroupId string) error {
	dSet(d, "name", ipSetType.Name)
	dSet(d, "description", ipSetType.Description)

	ipSetSet := convertStringsTotTypeSet(ipSetType.IpAddresses)

	err := d.Set("ip_addresses", ipSetSet)
	if err != nil {
		return fmt.Errorf("error settings 'ip_addresses': %s", err)
	}

	dSet(d, "owner_id", parentVdcOrVdcGroupId)
	return nil
}

func getNsxtIpSetType(d *schema.ResourceData, ownerId string) *types.NsxtFirewallGroup {
	ipSet := &types.NsxtFirewallGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: ownerId},
		Type:        types.FirewallGroupTypeIpSet,
	}

	if ipAddresses, isSet := d.GetOk("ip_addresses"); isSet {
		ipSet.IpAddresses = convertSchemaSetToSliceOfStrings(ipAddresses.(*schema.Set))
	}

	return ipSet
}
