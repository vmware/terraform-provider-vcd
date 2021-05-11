package vcd

import (
	"context"
	"errors"
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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP set name",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which IP Set is located",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP set description",
			},
			"ip_addresses": {
				Optional:    true,
				Type:        schema.TypeSet,
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
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet := getNsxtIpSetType(d)

	createdFwGroup, err := vdc.CreateNsxtFirewallGroup(ipSet)
	if err != nil {
		return diag.Errorf("error creating NSX-T IP Set '%s': %s", ipSet.Name, err)
	}

	d.SetId(createdFwGroup.NsxtFirewallGroup.ID)

	return resourceVcdNsxtIpSetRead(ctx, d, meta)
}

func resourceVcdNsxtIpSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T IP Set: %s", err)
	}

	updateIpSet := getNsxtIpSetType(d)
	// Inject existing ID for update
	updateIpSet.ID = d.Id()

	_, err = ipSet.Update(updateIpSet)
	if err != nil {
		return diag.Errorf("error updating NSX-T IP Set '%s': %s", ipSet.NsxtFirewallGroup.Name, err)
	}

	return resourceVcdNsxtIpSetRead(ctx, d, meta)
}

func resourceVcdNsxtIpSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting NSX-T IP Set with ID '%s': %s", d.Id(), err)
	}

	err = setNsxtIpSetData(d, ipSet.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("error setting NSX-T IP Set: %s", err)
	}

	return nil
}

func resourceVcdNsxtIpSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T IP Set: %s", err)
	}

	err = ipSet.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T IP Set: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtIpSetImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.ip_set_name")
	}
	orgName, vdcName, edgeGatewayName, ipSetName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", orgName, err)
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
	}

	if !vdc.IsNsxt() {
		return nil, errors.New("IP Sets are only supported by NSX-T VDCs")
	}

	edgeGateway, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("unable to find NSX-T Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	ipSet, err := edgeGateway.GetNsxtFirewallGroupByName(ipSetName, types.FirewallGroupTypeIpSet)
	if err != nil {
		return nil, fmt.Errorf("unable to find IP Set '%s': %s", edgeGatewayName, err)
	}

	if !ipSet.IsIpSet() {
		return nil, fmt.Errorf("firewall group '%s' is not a IP Set, but '%s'",
			ipSet.NsxtFirewallGroup.Name, ipSet.NsxtFirewallGroup.Type)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(ipSet.NsxtFirewallGroup.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtIpSetData(d *schema.ResourceData, ipSetType *types.NsxtFirewallGroup) error {
	_ = d.Set("name", ipSetType.Name)
	_ = d.Set("description", ipSetType.Description)

	ipSetInterface := convertToTypeSet(ipSetType.IpAddresses)
	ipSetSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), ipSetInterface)

	err := d.Set("ip_addresses", ipSetSet)
	if err != nil {
		return fmt.Errorf("error settings 'ip_addresses': %s", err)
	}

	return nil
}

func getNsxtIpSetType(d *schema.ResourceData) *types.NsxtFirewallGroup {
	ipSet := &types.NsxtFirewallGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EdgeGatewayRef: &types.OpenApiReference{
			ID: d.Get("edge_gateway_id").(string),
		},
		Type: types.FirewallGroupTypeIpSet,
	}

	if ipAddresses, isSet := d.GetOk("ip_addresses"); isSet {
		ipSet.IpAddresses = convertSchemaSetToSliceOfStrings(ipAddresses.(*schema.Set))
	}

	return ipSet
}
