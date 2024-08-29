package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdIpSpaceUplink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIpSpaceUplinkCreate,
		ReadContext:   resourceVcdIpSpaceUplinkRead,
		UpdateContext: resourceVcdIpSpaceUplinkUpdate,
		DeleteContext: resourceVcdIpSpaceUplinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdIpSpaceUplinkImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tenant facing name for IP Space Uplink",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP Space Uplink description",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "External Network ID",
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IP Space ID",
			},
			"ip_space_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Space Type",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Space Status",
			},
			"associated_interface_ids": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "A set of Tier-0 interfaces to associate to this uplink",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdIpSpaceUplinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space Uplink creation initiated")
	vcdClient.lockParentExternalNetwork(d)
	defer vcdClient.unlockParentExternalNetwork(d)

	ipSpaceUplinkConfig := getIpSpaceUplinkType(d)
	createdIpSpaceUplink, err := vcdClient.CreateIpSpaceUplink(ipSpaceUplinkConfig)
	if err != nil {
		return diag.Errorf("error creating IP Space Uplink: %s", err)
	}

	d.SetId(createdIpSpaceUplink.IpSpaceUplink.ID)

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcdClient.Client.WaitForRouteAdvertisementTasks()
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVcdIpSpaceUplinkRead(ctx, d, meta)
}

func resourceVcdIpSpaceUplinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space Uplink update initiated")
	vcdClient.lockParentExternalNetwork(d)
	defer vcdClient.unlockParentExternalNetwork(d)

	ipSpaceUplinkConfig := getIpSpaceUplinkType(d)
	ipSpaceUplink, err := vcdClient.GetIpSpaceUplinkById(d.Id())
	if err != nil {
		return diag.Errorf("error finding IP Space Uplink by ID '%s': %s", d.Id(), err)
	}

	ipSpaceUplinkConfig.ID = d.Id()
	_, err = ipSpaceUplink.Update(ipSpaceUplinkConfig)
	if err != nil {
		return diag.Errorf("error updating IP Space Uplink: %s", err)
	}

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcdClient.Client.WaitForRouteAdvertisementTasks()
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVcdIpSpaceUplinkRead(ctx, d, meta)
}

func resourceVcdIpSpaceUplinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space Uplink read initiated")

	ipSpaceUplink, err := vcdClient.GetIpSpaceUplinkById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}

		// For user convenience - removing the uplink configuration from state if the parent
		// External Network is no longer found as it is possible to remove External network itself
		// and all the uplinks are automatically removed
		_, err2 := govcd.GetExternalNetworkV2ById(vcdClient.VCDClient, d.Get("external_network_id").(string))
		if govcd.ContainsNotFound(err2) {
			d.SetId("")
			return nil
		}

		return diag.Errorf("error finding IP Space by ID '%s': %s", d.Id(), err)
	}

	err = setIpSpaceUplinkData(d, ipSpaceUplink.IpSpaceUplink)
	if err != nil {
		return diag.Errorf("error storing IP Space Uplink state: %s", err)
	}

	return nil
}

func resourceVcdIpSpaceUplinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space Uplink deletion initiated")
	vcdClient.lockParentExternalNetwork(d)
	defer vcdClient.unlockParentExternalNetwork(d)

	ipSpaceUplink, err := vcdClient.GetIpSpaceUplinkById(d.Id())
	if err != nil {
		return diag.Errorf("error finding IP Space Uplink by ID '%s': %s", d.Id(), err)
	}

	err = ipSpaceUplink.Delete()
	if err != nil {
		return diag.Errorf("error deleting IP Space Uplink by ID '%s': %s", d.Id(), err)
	}

	// Operations on IP Space related entities trigger a separate task
	// 'ipSpaceUplinkRouteAdvertisementSync' which is better to finish before any other operations
	// as it might cause an error: busy completing an operation IP_SPACE_UPLINK_ROUTE_ADVERTISEMENT_SYNC
	// Sleeping a few seconds because the task is not immediately seen sometimes.
	time.Sleep(3 * time.Second)
	err = vcdClient.Client.WaitForRouteAdvertisementTasks()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVcdIpSpaceUplinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] IP Space Uplink import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as external-network-name.uplink-name")
	}

	externalNetworkName := resourceURI[0]
	ipSpaceUplinkName := resourceURI[1]

	vcdClient := meta.(*VCDClient)

	extNetRes, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, externalNetworkName)
	if err != nil {
		return nil, fmt.Errorf("error fetching external network V2 details %s", err)
	}

	ipSpaceUplink, err := vcdClient.GetIpSpaceUplinkByName(extNetRes.ExternalNetwork.ID, ipSpaceUplinkName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Space Uplink by Name '%s' in External Network '%s': %s",
			ipSpaceUplinkName, extNetRes.ExternalNetwork.ID, err)
	}

	d.SetId(ipSpaceUplink.IpSpaceUplink.ID)

	return []*schema.ResourceData{d}, nil
}

func getIpSpaceUplinkType(d *schema.ResourceData) *types.IpSpaceUplink {
	result := &types.IpSpaceUplink{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		ExternalNetworkRef: &types.OpenApiReference{ID: d.Get("external_network_id").(string)},
		IPSpaceRef:         &types.OpenApiReference{ID: d.Get("ip_space_id").(string)},
	}

	if _, ok := d.GetOk("associated_interface_ids"); ok {
		associatedInterfaceSlice := convertSchemaSetToSliceOfStrings(d.Get("associated_interface_ids").(*schema.Set))
		associatedInterfaces := make([]types.IpSpaceUplinkInterface, len(associatedInterfaceSlice))
		for i, v := range associatedInterfaceSlice {
			associatedInterfaces[i].ID = v
		}

		result.Interfaces = associatedInterfaces
	}
	// else {
	// 	result.Interfaces = []types.IpSpaceUplinkInterface{}
	// }

	return result
}

func setIpSpaceUplinkData(d *schema.ResourceData, ipSpaceUplink *types.IpSpaceUplink) error {
	dSet(d, "name", ipSpaceUplink.Name)
	dSet(d, "description", ipSpaceUplink.Description)

	if ipSpaceUplink.ExternalNetworkRef != nil {
		dSet(d, "external_network_id", ipSpaceUplink.ExternalNetworkRef.ID)
	}

	if ipSpaceUplink.IPSpaceRef != nil {
		dSet(d, "ip_space_id", ipSpaceUplink.IPSpaceRef.ID)
	}

	dSet(d, "ip_space_type", ipSpaceUplink.IPSpaceType)
	dSet(d, "status", ipSpaceUplink.Status)

	if len(ipSpaceUplink.Interfaces) > 0 {
		ids := make([]string, len(ipSpaceUplink.Interfaces))
		for i, v := range ipSpaceUplink.Interfaces {
			ids[i] = v.ID
		}

		associatedInterfaceSet := convertStringsToTypeSet(ids)
		err := d.Set("associated_interface_ids", associatedInterfaceSet)
		if err != nil {
			return fmt.Errorf("error storing 'associated_interface_ids' to schema: %s", err)
		}

	} else {
		dSet(d, "associated_interface_ids", nil)
	}

	return nil
}
