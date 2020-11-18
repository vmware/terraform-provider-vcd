package vcd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

type resourceRef struct {
	name string
	id   string
	href string
}

func datasourceVcdResourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdResourceListRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			// Parent will be needed for:
			// * VM (parent: vApp)
			// * catalogItem (catalog)
			// * mediaItem (catalog)
			// * all edge gateway objects (NAT, firewall, lb)
			// When the parent is org or vdc, they are taken from the regular fields above
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the parent to the resources being retrieved",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the Info",
			},
			"resource_type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Which resource we should list",
			},
			"list": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Holds the list of requested resources",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"list_mode": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "How the list should be built",
				ValidateFunc: validation.StringInSlice([]string{
					"name",      // The list will contain only the entity name
					"id",        // The list will contain only the entity ID
					"href",      // The list will contain only the entity HREF
					"import",    // The list will contain the terraform import command
					"name_id",   // The list will contain name + ID for each item
					"hierarchy", // The list will contain parent names + resource name for each item
				}, true),
			},
			"name_id_separator": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "  ",
				Description: "Separator for name_id combination",
			},
		},
	}
}

func orgList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	orgList, err := GetOrgList(client.VCDClient)
	if err != nil {
		return list, err
	}
	for _, org := range orgList.Org {

		adminOrg, err := client.GetAdminOrgByName(org.Name)
		if err != nil {
			return []string{}, err
		}
		switch listMode {
		case "name", "hierarchy":
			list = append(list, org.Name)
		case "id":
			list = append(list, adminOrg.AdminOrg.ID)
		case "name_id":
			list = append(list, org.Name+nameIdSeparator+adminOrg.AdminOrg.ID)
		case "href":
			list = append(list, org.HREF)
		case "import":
			list = append(list, fmt.Sprintf("terraform import vcd_org.%s %s", org.Name, org.Name))
		}
	}
	return list, err
}

func externalNetworkList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	externalNetworks, err := client.GetExternalNetworks()

	if err != nil {
		return list, err
	}
	for _, en := range externalNetworks.ExternalNetworkReference {
		externalNetwork := govcd.NewExternalNetwork(&client.Client)
		externalNetwork.ExternalNetwork.HREF = en.HREF
		err = externalNetwork.Refresh()
		if err != nil {
			return []string{}, err
		}
		switch listMode {
		case "name", "hierarchy":
			list = append(list, en.Name)
		case "id":
			list = append(list, externalNetwork.ExternalNetwork.ID)
		case "name_id":
			list = append(list, en.Name+nameIdSeparator+externalNetwork.ExternalNetwork.ID)
		case "href":
			list = append(list, en.HREF)
		case "import":
			list = append(list, fmt.Sprintf("terraform import vcd_external_network.%s %s", en.Name, en.Name))
		}
	}
	return list, err
}

func catalogList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, err := client.GetAdminOrg(d.Get("org").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef

	for _, catRef := range org.AdminOrg.Catalogs.Catalog {
		catalog, err := org.GetCatalogByHref(catRef.HREF)
		if err != nil {
			return []string{}, err
		}
		items = append(items, resourceRef{
			name: catRef.Name,
			id:   catalog.Catalog.ID,
			href: catalog.Catalog.HREF,
		})
	}
	return genericResourceList("vcd_catalog", listMode, nameIdSeparator, []string{org.AdminOrg.Name}, items)
}

// catalogItemList finds either catalogItem or mediaItem
func catalogItemList(d *schema.ResourceData, meta interface{}, wantMedia bool) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, err := client.GetAdminOrg(d.Get("org").(string))
	if err != nil {
		return list, err
	}
	catalogName := d.Get("parent").(string)
	if catalogName == "" {
		return list, fmt.Errorf(`no catalog name (as "parent") given`)
	}
	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return list, err
	}
	var items []resourceRef

	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, reference := range catalogItems.CatalogItem {
			wanted := true
			catalogItem, err := catalog.GetCatalogItemByHref(reference.HREF)
			if err != nil {
				return list, err
			}
			if catalogItem.CatalogItem.Entity.Type == "application/vnd.vmware.vcloud.media+xml" {
				wanted = wantMedia
			}

			if wanted {
				items = append(items, resourceRef{
					name: reference.Name,
					id:   reference.ID,
					href: reference.HREF,
				})
			}

		}
	}
	return genericResourceList("vcd_catalog_item", listMode, nameIdSeparator, []string{org.AdminOrg.Name, catalogName}, items)
}

func vdcList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, err := client.GetAdminOrg(d.Get("org").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, vdc := range org.AdminOrg.Vdcs.Vdcs {
		items = append(items, resourceRef{
			name: vdc.Name,
			id:   vdc.ID,
			href: vdc.HREF,
		})
	}
	return genericResourceList("vcd_org_vdc", listMode, nameIdSeparator, []string{org.AdminOrg.Name}, items)
}

func orgUserList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, err := client.GetAdminOrg(d.Get("org").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, user := range org.AdminOrg.Users.User {
		items = append(items, resourceRef{
			name: user.Name,
			id:   user.ID,
			href: user.HREF,
		})
	}
	return genericResourceList("vcd_org_user", listMode, nameIdSeparator, []string{org.AdminOrg.Name}, items)
}

func networkList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	wantedType := d.Get("resource_type").(string)
	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}

	networkType := ""
	networkList, err := vdc.GetNetworkList()
	if err != nil {
		return list, err
	}
	for _, net := range networkList {
		switch net.LinkType {
		case 0:
			networkType = "direct"
		case 1:
			networkType = "routed"
		case 2:
			networkType = "isolated"
		}
		resourceName := "network"
		if wantedType != "network" {
			resourceName = "network_" + networkType
		}
		if wantedType != resourceName {
			continue
		}
		network, err := vdc.GetOrgVdcNetworkByHref(net.HREF)
		if err != nil {
			return []string{}, err
		}
		switch listMode {
		case "name":
			list = append(list, network.OrgVDCNetwork.Name)
		case "id":
			list = append(list, network.OrgVDCNetwork.ID)
		case "name_id":
			list = append(list, network.OrgVDCNetwork.Name+nameIdSeparator+network.OrgVDCNetwork.ID)
		case "hierarchy":
			list = append(list, org.Org.Name+nameIdSeparator+vdc.Vdc.Name+nameIdSeparator+network.OrgVDCNetwork.Name)
		case "href":
			list = append(list, network.OrgVDCNetwork.HREF)
		case "import":
			list = append(list, fmt.Sprintf("terraform import vcd_network_%s.%s %s%s%s%s%s",
				networkType, network.OrgVDCNetwork.Name,
				org.Org.Name,
				ImportSeparator,
				vdc.Vdc.Name,
				ImportSeparator,
				network.OrgVDCNetwork.Name))
		}
	}

	return list, nil
}

func edgeGatewayList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef
	edgeGatewayList, err := GetEdgeGatewayRecordsType(&client.Client, vdc, false)
	if err != nil {
		return list, err
	}
	for _, ert := range edgeGatewayList.EdgeGatewayRecord {

		edgeGateway, err := vdc.GetEdgeGatewayByName(ert.Name, false)
		if err != nil {
			return []string{}, err
		}
		items = append(items, resourceRef{
			name: edgeGateway.EdgeGateway.Name,
			id:   edgeGateway.EdgeGateway.ID,
			href: edgeGateway.EdgeGateway.HREF,
		})
	}
	return genericResourceList("vcd_edgegateway", listMode, nameIdSeparator, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func vappList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef

	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				items = append(items, resourceRef{
					name: resourceReference.Name,
					id:   resourceReference.ID,
					href: resourceReference.HREF,
				})
			}
		}
	}
	return genericResourceList("vcd_vapp", listMode, nameIdSeparator, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func vappVmList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}
	vappName := d.Get("parent").(string)
	if vappName == "" {
		return list, fmt.Errorf(`vApp name (as "parent") is required for VM lists`)
	}
	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return list, fmt.Errorf("error retrieving vApp '%s': %s ", vappName, err)
	}

	var items []resourceRef
	for _, vm := range vapp.VApp.Children.VM {
		items = append(items, resourceRef{
			name: vm.Name,
			id:   vm.ID,
			href: vm.HREF,
		})
	}
	return genericResourceList("vcd_vapp_vm", listMode, nameIdSeparator, []string{org.Org.Name, vdc.Vdc.Name, vappName}, items)
}

func genericResourceList(resType, listMode, nameIdSeparator string, ancestors []string, refs []resourceRef) (list []string, err error) {

	for _, ref := range refs {
		switch listMode {
		case "name":
			list = append(list, ref.name)
		case "id":
			list = append(list, ref.id)
		case "name_id":
			list = append(list, ref.name+nameIdSeparator+ref.id)
		case "hierarchy":
			list = append(list, strings.Join(ancestors, nameIdSeparator)+nameIdSeparator+ref.name)
		case "href":
			list = append(list, "")
		case "import":
			list = append(list, fmt.Sprintf("terraform import %s.%s %s%s%s",
				resType,
				ref.name,
				strings.Join(ancestors, ImportSeparator),
				ImportSeparator,
				ref.name))
		}
	}
	return list, nil
}

func getEdgeGatewayDetails(d *schema.ResourceData, meta interface{}) (orgName string, vdcName string, listMode string, separator string, egw *govcd.EdgeGateway, err error) {
	client := meta.(*VCDClient)

	listMode = d.Get("list_mode").(string)
	separator = d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return "", "", "", "", nil, err
	}
	edgeGatewayName := d.Get("parent").(string)
	if edgeGatewayName == "" {
		return "", "", "", "", nil, fmt.Errorf(`edge gateway name (as "parent") is required for this task`)
	}
	edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
	if err != nil {
		return "", "", "", "", nil, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGatewayName, err)
	}
	return org.Org.Name, vdc.Vdc.Name, listMode, separator, edgeGateway, nil
}

func lbServerPoolList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}
	lbServerPoolList, err := edgeGateway.GetLbServerPools()
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway LB server pools '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}
	var items []resourceRef
	for _, service := range lbServerPoolList {
		items = append(items, resourceRef{
			name: service.Name,
			id:   service.ID,
			href: "",
		})
	}

	return genericResourceList("vcd_lb_server_pool", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbServiceMonitorList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}

	var items []resourceRef
	lbServiceMonitorList, err := edgeGateway.GetLbServiceMonitors()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB service monitor list: %s ", err)
	}
	for _, sm := range lbServiceMonitorList {
		items = append(items, resourceRef{
			name: sm.Name,
			id:   sm.ID,
			href: sm.URL,
		})
	}
	return genericResourceList("vcd_lb_service_monitor", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbVirtualServerList(d *schema.ResourceData, meta interface{}) (list []string, err error) {

	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}
	var items []resourceRef
	lbVirtualServerList, err := edgeGateway.GetLbVirtualServers()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB virtual server list: %s ", err)
	}
	for _, vs := range lbVirtualServerList {
		items = append(items, resourceRef{
			name: vs.Name,
			id:   vs.ID,
			href: "",
		})
	}
	return genericResourceList("vcd_lb_virtual_server", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func nsxvFirewallList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}

	var items []resourceRef
	fwRuleList, err := edgeGateway.GetAllNsxvFirewallRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving NSXV firewall rule list: %s ", err)
	}
	for _, fw := range fwRuleList {
		items = append(items, resourceRef{
			name: fw.Name,
			id:   fw.ID,
			href: "",
		})
	}
	return genericResourceList("vcd_nsxv_firewall_rule", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbAppRuleList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}

	var items []resourceRef
	ruleList, err := edgeGateway.GetLbAppRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB app rule list: %s ", err)
	}
	for _, fw := range ruleList {
		items = append(items, resourceRef{
			name: fw.Name,
			id:   fw.ID,
			href: "",
		})
	}
	return genericResourceList("vcd_lb_app_rule", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbAppProfileList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}

	var items []resourceRef
	profiles, err := edgeGateway.GetLbAppProfiles()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB app profile list: %s ", err)
	}
	for _, fw := range profiles {
		items = append(items, resourceRef{
			name: fw.Name,
			id:   fw.ID,
			href: "",
		})
	}
	return genericResourceList("vcd_lb_app_profile", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func ipsetList(d *schema.ResourceData, meta interface{}) (list []string, err error) {

	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}

	var items []resourceRef

	ipSets, err := vdc.GetAllNsxvIpSets()
	// we only fail on errors other than an empty list
	if err != nil && !govcd.IsNotFound(err) {
		return list, err
	}

	for _, ipSet := range ipSets {
		items = append(items, resourceRef{
			name: ipSet.Name,
			id:   ipSet.ID,
			href: "",
		})
	}
	return genericResourceList("vcd_ipset", listMode, nameIdSeparator, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func nsxvNatRuleList(natType string, d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, listMode, separator, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", edgeGateway.EdgeGateway.Name, err)
	}

	var items []resourceRef
	rules, err := edgeGateway.GetNsxvNatRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving NSXV NAT rule list: %s ", err)
	}
	for _, rule := range rules {
		if rule.Action == natType {
			items = append(items, resourceRef{
				name: "",
				id:   rule.ID,
				href: "",
			})
		}
	}
	return genericResourceList("vcd_lb_app_profile", listMode, separator, []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func getResourcesList() ([]string, error) {
	var list []string
	resources := GlobalResourceMap
	for resource := range resources {
		list = append(list, resource)
	}
	// Returns the list of resources in alphabetical order, to keep a consistent state
	sort.Strings(list)
	return list, nil
}

func datasourceVcdResourceListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	requested := d.Get("resource_type")
	var err error
	var list []string
	switch requested {
	// Note: do not try to get the data sources list, as it would result in a circular reference
	case "resource", "resources":
		list, err = getResourcesList()
	case "vcd_org", "org", "orgs":
		list, err = orgList(d, meta)
	case "vcd_external_network", "external_network", "external_networks":
		list, err = externalNetworkList(d, meta)
	case "vcd_org_vdc", "vdc", "vdcs":
		list, err = vdcList(d, meta)
	case "vcd_catalog", "catalog", "catalogs":
		list, err = catalogList(d, meta)
	case "vcd_catalog_item", "catalog_item", "catalog_items", "catalogitem", "catalogitems":
		list, err = catalogItemList(d, meta, false)
	case "vcd_catalog_media", "catalog_media", "media_items", "mediaitems", "mediaitem":
		list, err = catalogItemList(d, meta, true)
	case "vcd_vapp", "vapp", "vapps":
		list, err = vappList(d, meta)
	case "vcd_vapp_vm", "vapp_vm", "vapp_vms":
		list, err = vappVmList(d, meta)
	case "vcd_org_user", "org_user", "user", "users":
		list, err = orgUserList(d, meta)
	case "vcd_edgegateway", "edge_gateway", "edge", "edgegateway":
		list, err = edgeGatewayList(d, meta)
	case "vcd_lb_server_pool", "lb_server_pool":
		list, err = lbServerPoolList(d, meta)
	case "vcd_lb_service_monitor", "lb_service_monitor":
		list, err = lbServiceMonitorList(d, meta)
	case "vcd_lb_virtual_server", "lb_virtual_server":
		list, err = lbVirtualServerList(d, meta)
	case "vcd_lb_app_rule", "lb_app_rule":
		list, err = lbAppRuleList(d, meta)
	case "vcd_lb_app_profile", "lb_app_profile":
		list, err = lbAppProfileList(d, meta)
	case "vcd_nsxv_firewall_rule", "nsxv_firewall_rule":
		list, err = nsxvFirewallList(d, meta)
	case "vcd_ipset", "ipset":
		list, err = ipsetList(d, meta)
	case "vcd_nsxv_dnat", "nsxv_dnat":
		list, err = nsxvNatRuleList("dnat", d, meta)
	case "vcd_nsxv_snat", "nsxv_snat":
		list, err = nsxvNatRuleList("snat", d, meta)
	case "vcd_network_isolated", "vcd_network_direct", "vcd_network_routed",
		"network", "networks", "network_direct", "network_routed", "network_isolated":
		list, err = networkList(d, meta)

		//// place holder to remind of what needs to be implemented
		//	case "edgegateway_vpn",
		//		"vapp_network",
		//		"independent_disk",
		//		"inserted_media":
		//		list, err = []string{"not implemented yet"}, nil
	default:
		return diag.FromErr(fmt.Errorf("unhandled resource type '%s'", requested))
	}

	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("list", list)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("name").(string))

	return diag.Diagnostics{}
}
