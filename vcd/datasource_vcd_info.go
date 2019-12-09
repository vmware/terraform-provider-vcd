package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdInfo() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdInfoRead,
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
			// Parent may be needed for:
			// * VM (parent: vApp)
			// * catalogItem (catalog)
			// * mediaItem (catalog)
			//
			// When the parent is org or vdc, they are taken from the regular fields above
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the parent to the resources being retrieved. Will search all if omitted",
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

	for _, catRef := range org.AdminOrg.Catalogs.Catalog {
		catalog, err := org.GetCatalogByHref(catRef.HREF)
		if err != nil {
			return []string{}, err
		}
		switch listMode {
		case "name":
			list = append(list, catRef.Name)
		case "id":
			list = append(list, catalog.Catalog.ID)
		case "name_id":
			list = append(list, catRef.Name+nameIdSeparator+catalog.Catalog.ID)
		case "hierarchy":
			list = append(list, org.AdminOrg.Name+nameIdSeparator+catRef.Name)
		case "href":
			list = append(list, catRef.HREF)
		case "import":
			list = append(list, fmt.Sprintf("terraform import vcd_catalog.%s %s%s%s", catRef.Name,
				org.AdminOrg.Name, ImportSeparator, catRef.Name))
		}
	}
	return list, nil
}

func vdcList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, err := client.GetAdminOrg(d.Get("org").(string))
	if err != nil {
		return list, err
	}

	for _, vdc := range org.AdminOrg.Vdcs.Vdcs {
		switch listMode {
		case "name":
			list = append(list, vdc.Name)
		case "id":
			list = append(list, vdc.ID)
		case "name_id":
			list = append(list, vdc.Name+nameIdSeparator+vdc.ID)
		case "hierarchy":
			list = append(list, org.AdminOrg.Name+nameIdSeparator+vdc.Name)
		case "href":
			list = append(list, vdc.HREF)
		case "import":
			list = append(list, fmt.Sprintf("terraform import vcd_org_vdc.%s %s%s%s", vdc.Name,
				org.AdminOrg.Name, ImportSeparator, vdc.Name))
		}
	}
	return list, nil
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
		case "tree":
			list = append(list, network.OrgVDCNetwork.Name+nameIdSeparator+network.OrgVDCNetwork.ID)
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

func vappList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}
	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				switch listMode {
				case "name":
					list = append(list, resourceReference.Name)
				case "id":
					list = append(list, resourceReference.ID)
				case "name_id":
					list = append(list, resourceReference.Name+nameIdSeparator+resourceReference.ID)
				case "hierarchy":
					list = append(list, org.Org.Name+nameIdSeparator+vdc.Vdc.Name+nameIdSeparator+resourceReference.Name)
				case "href":
					list = append(list, resourceReference.HREF)
				case "import":
					list = append(list, fmt.Sprintf("terraform import vcd_vapp.%s %s%s%s%s%s",
						resourceReference.Name, org.Org.Name, ImportSeparator, vdc.Vdc.Name,
						ImportSeparator, resourceReference.Name))
				}
			}
		}
	}
	return list, nil
}

func vappVmList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	listMode := d.Get("list_mode").(string)
	parent := d.Get("parent").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}
	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				vapp, err := vdc.GetVAppByHref(resourceReference.HREF)
				if err != nil {
					return []string{}, err
				}
				if parent == "" || parent == vapp.VApp.Name {
					for _, vm := range vapp.VApp.Children.VM {
						switch listMode {
						case "name":
							list = append(list, vm.Name)
						case "id":
							list = append(list, vm.ID)
						case "name_id":
							list = append(list, vm.Name+nameIdSeparator+vm.ID)
						case "hierarchy":
							list = append(list, org.Org.Name+nameIdSeparator+vdc.Vdc.Name+nameIdSeparator+resourceReference.Name+nameIdSeparator+vm.Name)
						case "href":
							list = append(list, vm.HREF)
						case "import":
							list = append(list, fmt.Sprintf("terraform import vcd_vapp_vm.%s %s%s%s%s%s%s%s",
								vm.Name,
								org.Org.Name, ImportSeparator,
								vdc.Vdc.Name, ImportSeparator,
								resourceReference.Name, ImportSeparator, vm.Name))
						}
					}
				}
			}
		}
	}
	return list, nil
}

func getDataSourceList() (list []string, err error) {
	p := Provider()
	for _, r := range p.DataSources() {
		list = append(list, r.Name)
	}
	return
}

func getResourcesList() (list []string, err error) {
	p := Provider()
	for _, r := range p.Resources() {
		list = append(list, r.Name)
	}
	return
}

func datasourceVcdInfoRead(d *schema.ResourceData, meta interface{}) error {

	requested := d.Get("resource_type")
	var err error
	var list []string
	switch requested {
	case "data_source", "data_sources", "datasource", "datasources":
		list, err = getDataSourceList()
	case "resource", "resources":
		list, err = getResourcesList()
	case "org", "orgs":
		list, err = orgList(d, meta)
	case "external_network", "external_networks":
		list, err = externalNetworkList(d, meta)
	case "vdc", "vdcs":
		list, err = vdcList(d, meta)
	case "catalog", "catalogs":
		list, err = catalogList(d, meta)
	case "vapp", "vapps":
		list, err = vappList(d, meta)
	case "vapp_vm", "vapp_vms":
		list, err = vappVmList(d, meta)
	case "network", "networks", "network_direct", "network_routed", "network_isolated":
		list, err = networkList(d, meta)
	/*
		// place holder to remind of what needs to be implemented
			case "edgegateway", "edgegateway_vpn",
				"lb_app_rule", "lb_app_profile", "lb_server_pool", "lb_virtual_server",
				"dnat", "snat", "firewall_rules",
				"nsxv_firewall_rule",
				"nsxv_snat",
				"ipset",
				"org_user",
				"vapp_network",
				"independent_disk",
				"catalog_media", "inserted_media", "catalog_item":
				list, err = []string{"not implemented yet"}, nil
	*/
	default:
		return fmt.Errorf("unhandled resource type %s", requested)
	}

	if err != nil {
		return err
	}
	err = d.Set("list", list)
	fmt.Printf("%#v\n", list)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))

	return nil
}
