package vcd

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"fmt"
	"log"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var networkV2IpScope = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": &schema.Schema{
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Gateway of the network",
			ValidateFunc: validation.IsIPAddress,
		},
		"prefix_length": &schema.Schema{
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Network mask",
			ValidateFunc: validation.IntAtLeast(1),
		},
		"enabled": &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "If subnet is enabled",
		},
		"static_ip_pool": &schema.Schema{
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "IP ranges used for static pool allocation in the network",
			Elem:        networkV2IpRange,
		},
		"dns1": &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Primary DNS server",
			ValidateFunc: validation.IsIPAddress,
			// Only NSX-V allows configuring DNS
			ConflictsWith: []string{"nsxt_network"},
		},
		"dns2": &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Secondary DNS server",
			ValidateFunc: validation.IsIPAddress,
			// Only NSX-V allows configuring DNS
			ConflictsWith: []string{"nsxt_network"},
		},
		"dns_suffix": &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "DNS suffix",
			// Only NSX-V allows configuring DNS
			ConflictsWith: []string{"nsxt_network"},
		},
	},
}

var networkV2IpRange = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": &schema.Schema{
			Type:         schema.TypeString,
			Required:     true,
			Description:  "Start address of the IP range",
			ValidateFunc: validation.IsIPAddress,
		},
		"end_address": &schema.Schema{
			Type:         schema.TypeString,
			Required:     true,
			Description:  "End address of the IP range",
			ValidateFunc: validation.IsIPAddress,
		},
	},
}

var networkV2NsxtNetwork = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"nsxt_manager_id": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "ID of NSX-T manager",
		},
		// NSX-T Tier 0 router backed external network
		"nsxt_tier0_router_id": &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Description:  "ID of NSX-T Tier-0 router (for T0 gateway backed external network)",
			ExactlyOneOf: []string{"nsxt_network.0.nsxt_tier0_router_id", "nsxt_network.0.nsxt_segment_name"},
		},
		// NSX-T segment backed external network (VCD 10.3+)
		"nsxt_segment_name": &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Description:  "Name of NSX-T segment (for NSX-T segment backed external network)",
			ExactlyOneOf: []string{"nsxt_network.0.nsxt_tier0_router_id", "nsxt_network.0.nsxt_segment_name"},
		},
	},
}

var networkV2VsphereNetwork = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"vcenter_id": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The vCenter server name",
		},
		"portgroup_id": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the port group",
		},
	},
}

func resourceVcdExternalNetworkV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdExternalNetworkV2Create,
		Update: resourceVcdExternalNetworkV2Update,
		Delete: resourceVcdExternalNetworkV2Delete,
		Read:   resourceVcdExternalNetworkV2Read,
		Importer: &schema.ResourceImporter{
			State: resourceVcdExternalNetworkV2Import,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network description",
			},
			"ip_scope": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A set of IP scopes for the network",
				Elem:        networkV2IpScope,
			},
			"vsphere_network": &schema.Schema{
				Type:         schema.TypeSet,
				Optional:     true,
				ExactlyOneOf: []string{"vsphere_network", "nsxt_network"},

				ForceNew:    true,
				Description: "A set of port groups that back this network. Each referenced DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.",
				Elem:        networkV2VsphereNetwork,
			},
			"nsxt_network": &schema.Schema{
				Type:         schema.TypeList,
				Optional:     true,
				ExactlyOneOf: []string{"vsphere_network", "nsxt_network"},
				MaxItems:     1,
				ForceNew:     true,
				Description:  "Reference to NSX-T Tier-0 router or Segment and manager",
				Elem:         networkV2NsxtNetwork,
			},
		},
	}
}

func resourceVcdExternalNetworkV2Create(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 creation initiated")

	err := externalNetworkV2ValidateApiVersionFeatures(d, vcdClient)
	if err != nil {
		return err
	}

	netType, err := getExternalNetworkV2Type(vcdClient, d, "")
	if err != nil {
		return fmt.Errorf("could not get network data: %s", err)
	}

	extNet, err := govcd.CreateExternalNetworkV2(vcdClient.VCDClient, netType)
	if err != nil {
		return fmt.Errorf("error applying data: %s", err)
	}

	// Only store ID and leave all the rest to "READ"
	d.SetId(extNet.ExternalNetwork.ID)

	return resourceVcdExternalNetworkV2Read(d, meta)
}

func resourceVcdExternalNetworkV2Update(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] update network V2 creation initiated")

	extNet, err := govcd.GetExternalNetworkV2ById(vcdClient.VCDClient, d.Id())
	if err != nil {
		return fmt.Errorf("could not find external network V2 by ID '%s': %s", d.Id(), err)
	}

	var knownNsxtSegmentId string
	if extNet.ExternalNetwork.NetworkBackings.Values[0].BackingTypeValue == types.ExternalNetworkBackingTypeNsxtSegment {
		knownNsxtSegmentId = extNet.ExternalNetwork.NetworkBackings.Values[0].BackingID
	}

	netType, err := getExternalNetworkV2Type(vcdClient, d, knownNsxtSegmentId)
	if err != nil {
		return fmt.Errorf("could not get network data: %s", err)
	}

	netType.ID = extNet.ExternalNetwork.ID
	extNet.ExternalNetwork = netType

	_, err = extNet.Update()
	if err != nil {
		return fmt.Errorf("error updating external network V2: %s", err)
	}

	return resourceVcdExternalNetworkV2Read(d, meta)
}

func resourceVcdExternalNetworkV2Read(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 read initiated")

	extNet, err := govcd.GetExternalNetworkV2ById(vcdClient.VCDClient, d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not find external network V2 by ID '%s': %s", d.Id(), err)
	}

	return setExternalNetworkV2Data(d, extNet.ExternalNetwork, vcdClient)
}

func resourceVcdExternalNetworkV2Delete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] external network V2 creation initiated")

	extNet, err := govcd.GetExternalNetworkV2ById(vcdClient.VCDClient, d.Id())
	if err != nil {
		return fmt.Errorf("could not find external network V2 by ID '%s': %s", d.Id(), err)
	}

	return extNet.Delete()
}

// resourceVcdExternalNetworkV2Import is responsible for importing the resource.
// The d.ID() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
// For this resource, the import path is just the external network name.
//
// Example import path (id): externalNetworkName
// Example import command:   terraform import vcd_external_network_v2.externalNetworkResourceName externalNetworkName
func resourceVcdExternalNetworkV2Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	extNetRes, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error fetching external network V2 details %s", err)
	}

	d.SetId(extNetRes.ExternalNetwork.ID)

	setExternalNetworkV2Data(d, extNetRes.ExternalNetwork, vcdClient)

	return []*schema.ResourceData{d}, nil
}

func getExternalNetworkV2Type(vcdClient *VCDClient, d *schema.ResourceData, knownNsxtSegmentId string) (*types.ExternalNetworkV2, error) {
	networkBackings, err := getExternalNetworkV2BackingType(vcdClient, d, knownNsxtSegmentId)
	if err != nil {
		return nil, fmt.Errorf("error getting network backing type: %s", err)
	}
	subnetSlice := getSubnetsType(d)

	newExtNet := &types.ExternalNetworkV2{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Subnets:         types.ExternalNetworkV2Subnets{Values: subnetSlice},
		NetworkBackings: networkBackings,
	}

	return newExtNet, nil
}

func getExternalNetworkV2BackingType(vcdClient *VCDClient, d *schema.ResourceData, knownNsxtSegmentId string) (types.ExternalNetworkV2Backings, error) {
	var backings types.ExternalNetworkV2Backings
	// var backing types.ExternalNetworkV2Backing
	// Network backings
	nsxtNetworkSlice := d.Get("nsxt_network").([]interface{})
	nsxvNetwork := d.Get("vsphere_network").(*schema.Set)

	switch {
	// NSX-T network defined. Can only be one.
	case len(nsxtNetworkSlice) > 0:
		nsxtNetworkStrings := convertToStringMap(nsxtNetworkSlice[0].(map[string]interface{}))

		var backingId string
		var backingType string

		switch {
		// External network backed by NSX-T Tier 0 router
		case nsxtNetworkStrings["nsxt_tier0_router_id"] != "":
			backingId = nsxtNetworkStrings["nsxt_tier0_router_id"]
			backingType = types.ExternalNetworkBackingTypeNsxtTier0Router
		// External network backed by NSX-T Segment
		case nsxtNetworkStrings["nsxt_segment_name"] != "":
			// for create operation NSX-T Segment ID can be looked up using nsxt_segment_name because it is not yet
			// consumed
			if knownNsxtSegmentId == "" {
				// bareNsxtManagerUuid
				bareNsxtManagerUuid := extractUuid(nsxtNetworkStrings["nsxt_manager_id"])
				filter := map[string]string{"nsxTManager": bareNsxtManagerUuid}
				nsxtImportableSwitch, err := vcdClient.GetFilteredNsxtImportableSwitchesByName(filter, nsxtNetworkStrings["nsxt_segment_name"])
				if err != nil {
					return types.ExternalNetworkV2Backings{}, fmt.Errorf("unable to find NSX-T logical switch: %s", err)
				}
				backingId = nsxtImportableSwitch.NsxtImportableSwitch.ID
			} else {
				// for update operation the existing NSX-T Segment ID must be fed in because consumed NSX-T segment can
				// not be looked up anymore
				backingId = knownNsxtSegmentId
			}

			backingType = types.ExternalNetworkBackingTypeNsxtSegment
		}

		backing := types.ExternalNetworkV2Backing{
			BackingID:        backingId, // Tier 0 router or NSX-T Importable Switch ID
			BackingTypeValue: backingType,
			NetworkProvider: types.NetworkProvider{
				ID: nsxtNetworkStrings["nsxt_manager_id"], // NSX-T manager
			},
		}
		backings.Values = append(backings.Values, backing)
	// NSX-V network defined. Can be multiple blocks
	case len(nsxvNetwork.List()) > 0:
		nsxvNetworkSlice := nsxvNetwork.List()

		for nsxvNetworkIndex := range nsxvNetworkSlice {

			nsxvNetworkStrings := convertToStringMap(nsxvNetworkSlice[nsxvNetworkIndex].(map[string]interface{}))

			// Lookup portgroup type to avoid user passing it because it was already present in datasource
			pgType, err := getPortGroupTypeById(vcdClient, nsxvNetworkStrings["portgroup_id"], nsxvNetworkStrings["vcenter_id"])

			// For standard vSwitch portgroups VCD reports the type to be "NETWORK", but OpenAPI external network
			// requires parameter "PORTGROUP".
			if pgType == types.ExternalNetworkBackingTypeNetwork {
				pgType = "PORTGROUP"
			}

			if err != nil {
				return types.ExternalNetworkV2Backings{}, fmt.Errorf("error validating portgroup type: %s", err)
			}

			backing := types.ExternalNetworkV2Backing{
				BackingID:   nsxvNetworkStrings["portgroup_id"],
				BackingType: pgType,
				NetworkProvider: types.NetworkProvider{
					ID: nsxvNetworkStrings["vcenter_id"],
				},
			}

			backings.Values = append(backings.Values, backing)
		}
	}

	// external network API has changed field from types.ExternalNetworkV2Backings.BackingType to
	// types.ExternalNetworkV2Backings.BackingTypeValue in version 35.0 (VCD 10.2). This resource must still support
	// both versions therefore field translation is required.
	//
	// The main code branch works with newer field values, but VCD 10.1 is still supported at this time
	// This loop translates requests to still work with pre VCD 10.2 versions.
	// TO-DO remove this translation with VCD 10.1 EOL
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") { // types.ExternalNetworkV2Backings.BackingType
		for index := range backings.Values {
			backings.Values[index].BackingType = backings.Values[index].BackingTypeValue
			backings.Values[index].BackingTypeValue = ""
		}
	}

	return backings, nil
}

func getPortGroupTypeById(vcdClient *VCDClient, portGroupId, vCenterId string) (string, error) {
	var pgType string

	// Lookup portgroup_type
	pgs, err := govcd.QueryPortGroups(vcdClient.VCDClient, "moref=="+portGroupId)
	if err != nil {
		return "", fmt.Errorf("error validating portgroup '%s' type: %s", portGroupId, err)
	}

	for _, pg := range pgs {
		if pg.MoRef == portGroupId && haveSameUuid(pg.Vc, vCenterId) {
			pgType = pg.PortgroupType
		}
	}
	if pgType == "" {
		return "", fmt.Errorf("could not find portgroup type for '%s'", portGroupId)
	}

	return pgType, nil
}

func getSubnetsType(d *schema.ResourceData) []types.ExternalNetworkV2Subnet {
	subnets := d.Get("ip_scope").(*schema.Set)
	subnetSlice := make([]types.ExternalNetworkV2Subnet, len(subnets.List()))
	for subnetIndex, subnet := range subnets.List() {
		subnetMap := subnet.(map[string]interface{})

		subnet := types.ExternalNetworkV2Subnet{
			Gateway:      subnetMap["gateway"].(string),
			DNSSuffix:    subnetMap["dns_suffix"].(string),
			DNSServer1:   subnetMap["dns1"].(string),
			DNSServer2:   subnetMap["dns2"].(string),
			PrefixLength: subnetMap["prefix_length"].(int),
			Enabled:      subnetMap["enabled"].(bool),
		}
		// Loop over IP ranges (static IP pools)
		subnet.IPRanges = types.ExternalNetworkV2IPRanges{Values: processIpRangesInMap(subnetMap)}

		subnetSlice[subnetIndex] = subnet
	}
	return subnetSlice
}

func processIpRangesInMap(subnetMap map[string]interface{}) []types.ExternalNetworkV2IPRange {
	staticIpRange := subnetMap["static_ip_pool"].(*schema.Set)
	return processIpRanges(staticIpRange)
}

func processIpRanges(staticIpPool *schema.Set) []types.ExternalNetworkV2IPRange {
	subnetRng := make([]types.ExternalNetworkV2IPRange, len(staticIpPool.List()))
	for rangeIndex, subnetRange := range staticIpPool.List() {
		subnetRangeStr := convertToStringMap(subnetRange.(map[string]interface{}))
		oneRange := types.ExternalNetworkV2IPRange{
			StartAddress: subnetRangeStr["start_address"],
			EndAddress:   subnetRangeStr["end_address"],
		}
		subnetRng[rangeIndex] = oneRange
	}
	return subnetRng
}

func setExternalNetworkV2Data(d *schema.ResourceData, net *types.ExternalNetworkV2, vcdClient *VCDClient) error {
	_ = d.Set("name", net.Name)
	_ = d.Set("description", net.Description)

	// Loop over all subnets (known as ip_scope in UI)
	subnetSlice := make([]interface{}, len(net.Subnets.Values))
	for i, subnet := range net.Subnets.Values {
		subnetMap := make(map[string]interface{})
		subnetMap["gateway"] = subnet.Gateway
		subnetMap["prefix_length"] = subnet.PrefixLength
		subnetMap["dns1"] = subnet.DNSServer1
		subnetMap["dns2"] = subnet.DNSServer2
		subnetMap["dns_suffix"] = subnet.DNSSuffix
		subnetMap["enabled"] = subnet.Enabled

		// Gather all IP ranges  (known as static_ip_pool in UI)
		if len(subnet.IPRanges.Values) > 0 {
			ipRangeSlice := make([]interface{}, len(subnet.IPRanges.Values))
			for ii, ipRange := range subnet.IPRanges.Values {
				ipRangeMap := make(map[string]interface{})
				ipRangeMap["start_address"] = ipRange.StartAddress
				ipRangeMap["end_address"] = ipRange.EndAddress

				ipRangeSlice[ii] = ipRangeMap
			}
			ipRangeSet := schema.NewSet(schema.HashResource(networkV2IpRange), ipRangeSlice)
			subnetMap["static_ip_pool"] = ipRangeSet
		}
		subnetSlice[i] = subnetMap
	}

	subnetSet := schema.NewSet(schema.HashResource(networkV2IpScope), subnetSlice)
	err := d.Set("ip_scope", subnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'ip_scope' block: %s", err)
	}

	// external network API has changed field from types.ExternalNetworkV2Backings.BackingType to
	// types.ExternalNetworkV2Backings.BackingTypeValue in version 35.0 (VCD 10.2). This resource must still support
	// both versions therefore field translation is required.
	//
	// The main code branch works with newer field values, but VCD 10.1 is still supported at this time
	// This loop translates requests to still work with pre VCD 10.2 versions.
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") { // types.ExternalNetworkV2Backings.BackingType
		for index := range net.NetworkBackings.Values {
			net.NetworkBackings.Values[index].BackingTypeValue = net.NetworkBackings.Values[index].BackingType
		}
	}

	// Switch on first value of backing ID. If it is NSX-T - it can be only one block (limited by schema).
	// NSX-V can have more than one
	switch net.NetworkBackings.Values[0].BackingTypeValue {
	// Some versions of VCD behave strangely in API. They do accept a parameter of types.ExternalNetworkBackingTypeNetwork
	// as it was always the case, but in response they do return "PORTGROUP".
	case types.ExternalNetworkBackingDvPortgroup, types.ExternalNetworkBackingTypeNetwork, "PORTGROUP":
		backingInterface := make([]interface{}, len(net.NetworkBackings.Values))
		for backingIndex := range net.NetworkBackings.Values {
			backing := net.NetworkBackings.Values[backingIndex]
			backingMap := make(map[string]interface{})
			backingMap["vcenter_id"] = backing.NetworkProvider.ID
			backingMap["portgroup_id"] = backing.BackingID

			backingInterface[backingIndex] = backingMap

		}
		backingSet := schema.NewSet(schema.HashResource(networkV2VsphereNetwork), backingInterface)
		err := d.Set("vsphere_network", backingSet)
		if err != nil {
			return fmt.Errorf("error setting 'vsphere_network' block: %s", err)
		}

	case types.ExternalNetworkBackingTypeNsxtTier0Router, types.ExternalNetworkBackingTypeNsxtVrfTier0Router:
		backingInterface := make([]interface{}, 1)
		backing := net.NetworkBackings.Values[0]
		backingMap := make(map[string]interface{})
		backingMap["nsxt_manager_id"] = backing.NetworkProvider.ID
		backingMap["nsxt_tier0_router_id"] = backing.BackingID

		backingInterface[0] = backingMap
		err := d.Set("nsxt_network", backingInterface)
		if err != nil {
			return fmt.Errorf("error setting 'nsxt_network' block: %s", err)
		}
	case types.ExternalNetworkBackingTypeNsxtSegment:
		backingInterface := make([]interface{}, 1)
		backing := net.NetworkBackings.Values[0]
		backingMap := make(map[string]interface{})
		backingMap["nsxt_manager_id"] = backing.NetworkProvider.ID
		backingMap["nsxt_segment_name"] = backing.Name

		backingInterface[0] = backingMap
		err := d.Set("nsxt_network", backingInterface)
		if err != nil {
			return fmt.Errorf("error setting 'nsxt_network' block: %s", err)
		}

	default:
		return fmt.Errorf("unrecognized network backing type: %s", net.NetworkBackings.Values[0].BackingType)
	}

	return nil
}

func externalNetworkV2ValidateApiVersionFeatures(d *schema.ResourceData, vcdClient *VCDClient) error {
	// Only VCD 10.3.0+ supports NSX-T Segment backed external network
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.0") && d.Get("nsxt_network.0.nsxt_segment_name").(string) != "" {
		return fmt.Errorf("NSX-T Segment backed External Network is only supported in VCD 10.3.0+")
	}
	return nil
}
