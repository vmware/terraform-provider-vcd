package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type resourceRef struct {
	name         string
	resourceType string
	id           string
	href         string
	parent       string
	importId     bool
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
			// * vapp_vm (parent: vApp)
			// * catalogItem (catalog)
			// * mediaItem (catalog)
			// * NSX-T edge gateway (when belonging to a VDC group)
			// * all edge gateway objects (NAT, firewall, lb)
			// When the parent is org or vdc, they are taken from the regular fields above
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the parent to the resources being retrieved",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the Info",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Which resource we should list",
			},
			"list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Holds the list of requested resources",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"list_mode": {
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
			"import_file_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File where to store the import info - Only used with 'import' list mode ",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Optional regular expression filter on the resource names - Only the matching resources will be fetched",
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"name_id_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "  ",
				Description: "Separator for name_id combination",
			},
		},
	}
}

func getSiteAssociationList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)

	siteAssociationList, err := client.VCDClient.Client.QueryAllSiteAssociations(nil, nil)
	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, site := range siteAssociationList {
		items = append(items, resourceRef{
			name:     site.AssociatedSiteName,
			id:       site.Href,
			href:     site.Href,
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, resType, nil, items)
}

func getOrgAssociationList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)

	orgAssociationList, err := client.VCDClient.Client.QueryAllOrgAssociations(nil, nil)
	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, org := range orgAssociationList {
		items = append(items, resourceRef{
			name:     org.OrgName,
			id:       org.Href,
			href:     org.Href,
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, resType, nil, items)
}

func getOrgList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)

	orgList, err := client.VCDClient.GetOrgList()
	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, org := range orgList.Org {
		items = append(items, resourceRef{
			name:     org.Name,
			id:       "urn:vcloud:org:" + extractUuid(org.HREF),
			href:     org.HREF,
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, resType, nil, items)
}

func getPvdcList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	pvdcList, err := client.QueryProviderVdcs()
	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, pvdc := range pvdcList {
		items = append(items, resourceRef{
			name:         pvdc.Name,
			id:           "urn:vcloud:providervdc:" + extractUuid(pvdc.HREF),
			href:         pvdc.HREF,
			parent:       "",
			importId:     false,
			resourceType: "vcd_provider_vdc",
		})
	}
	return genericResourceList(d, "vcd_provider_vdc", nil, items)
}

func getVdcGroups(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}
	vdcGroups, err := org.GetAllVdcGroups(nil)
	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, vdcg := range vdcGroups {
		items = append(items, resourceRef{
			name:         vdcg.VdcGroup.Name,
			id:           vdcg.VdcGroup.Id,
			href:         "",
			parent:       org.AdminOrg.Name,
			importId:     false,
			resourceType: "vcd_vdc_group",
		})
	}
	return genericResourceList(d, "vcd_vdc_group", []string{org.AdminOrg.Name}, items)
}

func externalNetworkList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	if !client.VCDClient.Client.IsSysAdmin {
		return []string{}, fmt.Errorf("external network list requires system administrator privileges")
	}
	externalNetworks, err := client.GetExternalNetworks()

	if err != nil {
		return list, err
	}
	var items []resourceRef
	for _, en := range externalNetworks.ExternalNetworkReference {
		externalNetwork := govcd.NewExternalNetwork(&client.Client)
		externalNetwork.ExternalNetwork.HREF = en.HREF
		err = externalNetwork.Refresh()
		if err != nil {
			return []string{}, err
		}

		items = append(items, resourceRef{
			name:     externalNetwork.ExternalNetwork.Name,
			id:       externalNetwork.ExternalNetwork.ID,
			href:     externalNetwork.ExternalNetwork.HREF,
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, "vcd_external_network", nil, items)
}

func rightsList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}

	rights, err := org.GetAllRights(nil)
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, right := range rights {
		items = append(items, resourceRef{
			name:     right.Name,
			id:       right.ID,
			href:     "",
			parent:   org.AdminOrg.Name,
			importId: false,
		})
	}
	return genericResourceList(d, "vcd_right", []string{org.AdminOrg.Name}, items)
}

func rolesList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}

	roles, err := org.GetAllRoles(nil)
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, role := range roles {
		items = append(items, resourceRef{
			name:     role.Role.Name,
			id:       role.Role.ID,
			href:     "",
			parent:   org.AdminOrg.Name,
			importId: false,
		})
	}
	return genericResourceList(d, "vcd_role", []string{org.AdminOrg.Name}, items)

}

func globalRolesList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)
	globalRoles, err := client.Client.GetAllGlobalRoles(nil)
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, globalRole := range globalRoles {
		items = append(items, resourceRef{
			name:     globalRole.GlobalRole.Name,
			id:       globalRole.GlobalRole.Id,
			href:     "",
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, "vcd_global_role", nil, items)
}

func libraryCertificateList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	adminOrg, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}

	var ancestors []string
	parent := ""
	var certificates []*govcd.Certificate
	if isSysOrg(adminOrg) {
		certificates, err = client.Client.GetAllCertificatesFromLibrary(nil)
	} else {
		certificates, err = adminOrg.GetAllCertificatesFromLibrary(nil)
		ancestors = []string{adminOrg.AdminOrg.Name}
		parent = adminOrg.AdminOrg.Name
	}
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, certificate := range certificates {
		items = append(items, resourceRef{
			name:     certificate.CertificateLibrary.Alias,
			id:       certificate.CertificateLibrary.Id,
			href:     "",
			parent:   parent,
			importId: false,
		})

	}
	return genericResourceList(d, "vcd_certificate_library", ancestors, items)
}

func rightsBundlesList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	rightsBundles, err := client.Client.GetAllRightsBundles(nil)
	if err != nil {
		return list, err
	}
	var items []resourceRef

	for _, bundle := range rightsBundles {
		items = append(items, resourceRef{
			name:     bundle.RightsBundle.Name,
			id:       bundle.RightsBundle.Id,
			href:     "",
			parent:   "",
			importId: false,
		})
	}
	return genericResourceList(d, "vcd_rights_bundle", nil, items)
}

func catalogList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)
	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
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
	return genericResourceList(d, resType, []string{org.AdminOrg.Name}, items)
}

// catalogItemList finds either catalogItem or mediaItem
func catalogItemList(d *schema.ResourceData, meta interface{}, wantResource string) (list []string, err error) {
	client := meta.(*VCDClient)

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
			resourceType := wantResource
			wanted := false
			catalogItem, err := catalog.GetCatalogItemByHref(reference.HREF)
			if err != nil {
				return list, err
			}
			entity := types.Reference{
				HREF: catalogItem.CatalogItem.Entity.HREF,
				ID:   catalogItem.CatalogItem.Entity.ID,
				Name: reference.Name,
			}
			switch wantResource {
			case "vcd_catalog_item":
				entity.HREF = catalogItem.CatalogItem.HREF
				entity.ID = catalogItem.CatalogItem.ID
				wanted = true
			case "vcd_catalog_media":
				wanted = catalogItem.CatalogItem.Entity.Type == types.MimeMediaItem
			case "vcd_catalog_vapp_template":
				wanted = catalogItem.CatalogItem.Entity.Type == types.MimeVAppTemplate
			}
			if wanted {
				items = append(items, resourceRef{
					name:         reference.Name,
					id:           entity.ID,
					href:         entity.HREF,
					parent:       catalogName,
					importId:     false,
					resourceType: resourceType,
				})
			}
		}
	}
	return genericResourceList(d, "vcd_catalog_item", []string{org.AdminOrg.Name, catalogName}, items)
}

// vappTemplateList finds all vApp Templates
func vappTemplateList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)
	org, err := client.GetOrg(d.Get("org").(string))
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

	templates, err := catalog.QueryVappTemplateList()
	if err != nil {
		return list, err
	}

	var items []resourceRef

	for _, template := range templates {
		items = append(items, resourceRef{
			name:     template.Name,
			id:       extractUuid(template.HREF),
			href:     template.HREF,
			parent:   catalogName,
			importId: false,
		})
	}

	return genericResourceList(d, "vcd_catalog_vapp_template", []string{org.Org.Name, catalogName}, items)
}

func vdcList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)

	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, vdc := range org.AdminOrg.Vdcs.Vdcs {
		items = append(items, resourceRef{
			name:   vdc.Name,
			id:     vdc.ID,
			href:   vdc.HREF,
			parent: org.AdminOrg.Name,
		})
	}
	return genericResourceList(d, resType, []string{org.AdminOrg.Name}, items)
}

func orgUserList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	org, err := client.GetAdminOrg(firstNonEmpty(d.Get("org").(string), d.Get("parent").(string)))
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, user := range org.AdminOrg.Users.User {
		items = append(items, resourceRef{
			name:   user.Name,
			id:     user.ID,
			href:   user.HREF,
			parent: org.AdminOrg.Name,
		})
	}
	return genericResourceList(d, "vcd_org_user", []string{org.AdminOrg.Name}, items)
}

func networkList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	vdcName, err := getVdcName(client, d)
	if err != nil {
		return list, err
	}
	wantedType := d.Get("resource_type").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), vdcName)
	if err != nil {
		return list, err
	}

	networkType := ""
	networks, err := vdc.GetNetworkList()
	if err != nil {
		return list, err
	}
	var items []resourceRef

	resourceType := ""
	trueResourceType := ""
	for _, net := range networks {
		switch net.LinkType {
		case 0:
			networkType = "direct"
		case 1:
			networkType = "routed"
		case 2:
			networkType = "isolated"
		}
		resourceType = "network"
		if wantedType != "network" {
			resourceType = "vcd_network_" + networkType
		}
		trueResourceType = "vcd_network_" + networkType
		if wantedType != resourceType {
			continue
		}
		network, err := vdc.GetOrgVdcNetworkByHref(net.HREF)
		if err != nil {
			return []string{}, err
		}
		items = append(items, resourceRef{
			name:         network.OrgVDCNetwork.Name,
			id:           network.OrgVDCNetwork.ID,
			href:         network.OrgVDCNetwork.HREF,
			parent:       vdcName,
			resourceType: trueResourceType,
			importId:     false,
		})
	}

	return genericResourceList(d, resourceType, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

// orgNetworkListV2 uses OpenAPI endpoint to query Org VDC networks and return their list
func orgNetworkListV2(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)
	vdcName, err := getVdcName(client, d)
	if err != nil {
		return list, err
	}
	wantedType := d.Get("resource_type").(string)
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), vdcName)
	if err != nil {
		return list, err
	}

	orgVdcNetworkList, err := vdc.GetAllOpenApiOrgVdcNetworks(nil)
	if err != nil {
		return list, err
	}
	var items []resourceRef

	var resourceType string
	for _, net := range orgVdcNetworkList {
		trueResourceType := ""
		switch net.OpenApiOrgVdcNetwork.NetworkType {
		case types.OrgVdcNetworkTypeRouted:
			resourceType = "vcd_network_routed_v2"
		case types.OrgVdcNetworkTypeIsolated:
			resourceType = "vcd_network_isolated_v2"
		case types.OrgVdcNetworkTypeOpaque: // Used for Imported
			resourceType = "vcd_nsxt_network_imported"
		}

		trueResourceType = resourceType
		// Skip undesired network types
		if wantedType != resourceType {
			continue
		}
		href, err := client.Client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointOrgVdcNetworks, net.OpenApiOrgVdcNetwork.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, resourceRef{
			name:         net.OpenApiOrgVdcNetwork.Name,
			id:           net.OpenApiOrgVdcNetwork.ID,
			href:         href.Path,
			parent:       vdcName,
			importId:     false,
			resourceType: trueResourceType,
		})
	}

	return genericResourceList(d, resourceType, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func getVdcName(client *VCDClient, d *schema.ResourceData) (string, error) {
	vdcName := d.Get("parent").(string)
	if vdcName == "" {
		vdcName = d.Get("vdc").(string)
	}
	if vdcName == "" {
		vdcName = client.Vdc
	}
	if vdcName == "" {
		return "", fmt.Errorf("VDC name not given either as 'vdc' or 'parent' field")
	}
	return vdcName, nil
}

func getEdgeGatewayList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)

	vdcName, err := getVdcName(client, d)
	if err != nil {
		return nil, fmt.Errorf("VDC name not given either as 'vdc' or 'parent' field")
	}
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), vdcName)
	if err != nil {
		return list, err
	}

	var items []resourceRef
	edgeGatewayList, err := vdc.QueryEdgeGatewayList()
	if err != nil {
		return list, err
	}
	for _, ert := range edgeGatewayList {

		edgeGateway, err := vdc.GetEdgeGatewayByName(ert.Name, false)
		if err != nil {
			return []string{}, err
		}
		items = append(items, resourceRef{
			name:   edgeGateway.EdgeGateway.Name,
			id:     edgeGateway.EdgeGateway.ID,
			href:   edgeGateway.EdgeGateway.HREF,
			parent: vdc.Vdc.Name,
		})
	}
	return genericResourceList(d, resType, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func distributedSwitchList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	vCenterName := d.Get("parent").(string)
	if vCenterName == "" {
		return nil, fmt.Errorf("the 'parent' field must contain the vCenter name to retrieve the distributed switches")
	}

	vCenter, err := client.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}

	dSwitches, err := client.GetAllVcenterDistributedSwitches(vCenter.VSphereVCenter.VcId, nil)

	if err != nil {
		return nil, fmt.Errorf("error retrieving distributed switchs for vCenter '%s': %s", vCenterName, err)
	}

	var items []resourceRef
	for _, dsw := range dSwitches {
		items = append(items, resourceRef{
			name: dsw.BackingRef.Name,
			id:   dsw.BackingRef.ID,
		})
	}
	return genericResourceList(d, "vcd_distributed_switch", []string{vCenter.VSphereVCenter.Name}, items)
}

func transportZoneList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	nsxtManagerName := d.Get("parent").(string)
	if nsxtManagerName == "" {
		return nil, fmt.Errorf("the 'parent' field must contain the NSX-T manager name to retrieve the transport zones")
	}

	managers, err := client.QueryNsxtManagerByName(nsxtManagerName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T manager '%s': %s", nsxtManagerName, err)
	}
	if len(managers) == 0 {
		return nil, fmt.Errorf("no NSX-T manager '%s' found", nsxtManagerName)
	}
	if len(managers) > 1 {
		return nil, fmt.Errorf("more than one NSX-T manager found with name '%s'", nsxtManagerName)
	}

	manager := managers[0]
	managerId := "urn:vcloud:nsxtmanager:" + extractUuid(manager.HREF)
	transportZones, err := client.GetAllNsxtTransportZones(managerId, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving transport zones for NSX-T manager '%s': %s", nsxtManagerName, err)
	}

	var items []resourceRef
	for _, tz := range transportZones {
		if tz.AlreadyImported {
			continue
		}
		items = append(items, resourceRef{
			name: tz.Name,
			id:   tz.Id,
		})
	}
	return genericResourceList(d, "vcd_nsxt_transport_zone", []string{manager.Name}, items)
}

func importablePortGroupList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	vCenterName := d.Get("parent").(string)
	if vCenterName == "" {
		return nil, fmt.Errorf("the 'parent' field must contain the vCenter name to retrieve the distributed switches")
	}

	vCenter, err := client.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}
	var params = make(url.Values)
	params.Set("filter", fmt.Sprintf("virtualCenter.id==%s", vCenter.VSphereVCenter.VcId))
	pgroups, err := client.GetAllVcenterImportableDvpgs(params)
	if err != nil {
		return nil, fmt.Errorf("error retrieving importable port groups for vCenter '%s': %s", vCenterName, err)
	}

	var items []resourceRef
	for _, pg := range pgroups {
		items = append(items, resourceRef{
			name: pg.VcenterImportableDvpg.BackingRef.Name,
			id:   pg.VcenterImportableDvpg.BackingRef.ID,
		})
	}
	return genericResourceList(d, "vcd_importable_port_group", []string{vCenter.VSphereVCenter.Name}, items)
}

func networkPoolList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	networkPools, err := client.QueryNetworkPools()
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, np := range networkPools {

		items = append(items, resourceRef{
			name: np.Name,
			id:   extractUuid(np.HREF),
			href: np.HREF,
		})
	}
	return genericResourceList(d, "vcd_network_pool", nil, items)
}

func nsxtManagerList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	managers, err := client.QueryNsxtManagers()
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, mgr := range managers {

		items = append(items, resourceRef{
			name: mgr.Name,
			id:   extractUuid(mgr.HREF),
			href: mgr.HREF,
		})

	}
	return genericResourceList(d, "vcd_nsxt_manager", nil, items)
}

func vcenterList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	vcenters, err := client.GetAllVCenters(nil)
	if err != nil {
		return list, err
	}

	var items []resourceRef
	for _, vc := range vcenters {

		href, err := vc.GetVimServerUrl()
		if err != nil {
			return nil, err
		}
		items = append(items, resourceRef{
			name: vc.VSphereVCenter.Name,
			id:   vc.VSphereVCenter.VcId,
			href: href,
		})

	}
	return genericResourceList(d, "vcd_vcenter", nil, items)
}

func getNsxtEdgeGatewayList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)

	// A NSX-T edge gateway could belong to either a VDC or a VDC group
	// The "parent" field could refer to either of them
	parentName := d.Get("parent").(string)
	vdcName := d.Get("vdc").(string)
	var nsxtEdgeGatewayList []*govcd.NsxtEdgeGateway
	var vdcGroup *govcd.VdcGroup
	var vdc *govcd.Vdc
	var items []resourceRef

	adminOrg, err := client.GetAdminOrgFromResource(d)
	if err != nil {
		return list, err
	}

	var ancestors = []string{adminOrg.AdminOrg.Name}
	if parentName != "" {
		// we first try to get a group VDC using the parent name
		vdcGroup, err = adminOrg.GetVdcGroupByName(parentName)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				// if we haven't found a group VDC, we try a VDC using the parent name
				vdc, err = adminOrg.GetVDCByName(parentName, false)
				if err != nil {
					return nil, fmt.Errorf("neither a VDC or a VDC group found with name '%s'", parentName)
				}
			} else {
				return nil, fmt.Errorf("error retrieving VDC group '%s': %s", parentName, err)
			}
		}
	}
	if vdcGroup != nil {
		ancestors = append(ancestors, parentName)
		nsxtEdgeGatewayList, err = vdcGroup.GetAllNsxtEdgeGateways(nil)
	} else {
		if vdc == nil {
			if vdcName == "" {
				vdcName = client.Vdc
			}
			if vdcName == "" {
				return nil, fmt.Errorf("no VDC name provided")
			}
			vdc, err = adminOrg.GetVDCByName(vdcName, false)
			if err != nil {
				return list, fmt.Errorf("error retrieving VDC")
			}
		}
		parentName = vdcName
		ancestors = append(ancestors, vdcName)
		nsxtEdgeGatewayList, err = vdc.GetAllNsxtEdgeGateways(nil)
	}
	if err != nil {
		return list, err
	}
	for _, nsxtEdgeGateway := range nsxtEdgeGatewayList {

		items = append(items, resourceRef{
			name:   nsxtEdgeGateway.EdgeGateway.Name,
			id:     nsxtEdgeGateway.EdgeGateway.ID,
			href:   "",
			parent: parentName,
		})
	}
	return genericResourceList(d, "vcd_nsxt_edgegateway", ancestors, items)
}

func diskList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	client := meta.(*VCDClient)
	vdcName, err := getVdcName(client, d)
	if err != nil {
		return list, err
	}
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), vdcName)
	if err != nil {
		return list, err
	}

	var items []resourceRef

	disks, err := vdc.QueryDisks("*")
	if err != nil {
		return list, err
	}
	for _, diskRef := range *disks {
		items = append(items, resourceRef{
			name:   diskRef.Name,
			id:     extractUuid(diskRef.HREF),
			href:   diskRef.HREF,
			parent: vdc.Vdc.Name,
		})
	}
	return genericResourceList(d, "vcd_independent_disk", []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func vappList(d *schema.ResourceData, meta interface{}, resType string) (list []string, err error) {
	client := meta.(*VCDClient)
	vdcName, err := getVdcName(client, d)
	if err != nil {
		return list, err
	}
	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), vdcName)
	if err != nil {
		return list, err
	}

	var items []resourceRef

	for _, resourceEntities := range vdc.Vdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == "application/vnd.vmware.vcloud.vApp+xml" {
				items = append(items, resourceRef{
					name:   resourceReference.Name,
					id:     resourceReference.ID,
					href:   resourceReference.HREF,
					parent: vdc.Vdc.Name,
				})
			}
		}
	}
	return genericResourceList(d, resType, []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func vmList(d *schema.ResourceData, meta interface{}, vmType typeOfVm) (list []string, err error) {
	client := meta.(*VCDClient)

	org, vdc, err := client.GetOrgAndVdc(d.Get("org").(string), d.Get("vdc").(string))
	if err != nil {
		return list, err
	}

	vappName := d.Get("parent").(string)
	vms, err := vdc.QueryVmList(types.VmQueryFilterOnlyDeployed)
	if err != nil {
		return nil, err
	}
	var items []resourceRef
	for _, vm := range vms {
		if vmType == standaloneVmType && !vm.AutoNature {
			continue
		}
		if vmType == vappVmType && vm.AutoNature {
			continue
		}
		if vappName != "" && vappName != vm.ContainerName {
			continue
		}
		items = append(items, resourceRef{
			name:     vm.Name,
			id:       "urn:vcloud:vm:" + extractUuid(vm.HREF),
			href:     vm.HREF,
			parent:   vm.ContainerName,           // name of the hidden vApp
			importId: vmType == standaloneVmType, // import should use entity ID rather than name
		})
	}
	if vmType == vappVmType {
		return genericResourceList(d, "vcd_vapp_vm", []string{org.Org.Name, vdc.Vdc.Name, vappName}, items)
	}
	return genericResourceList(d, "vcd_vm", []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func genericResourceList(d *schema.ResourceData, resType string, ancestors []string, refs []resourceRef) (list []string, err error) {
	listMode := d.Get("list_mode").(string)
	nameIdSeparator := d.Get("name_id_separator").(string)
	importFile := d.Get("import_file_name").(string)
	nameRegex := d.Get("name_regex").(string)
	var importData strings.Builder
	importData.WriteString(fmt.Sprintf("# Generated by vcd_resource_list - %s\n", time.Now().Format(time.RFC3339)))
	var reName *regexp.Regexp
	if nameRegex != "" {
		reName, err = regexp.Compile(nameRegex)
		if err != nil {
			return nil, fmt.Errorf("[vcd_resource_list - %s] error compiling regular expression given with 'name_regex' '%s': %s",
				d.Get("name").(string), nameRegex, err)
		}
	}
	for _, ref := range refs {
		resourceType := resType
		if ref.resourceType != "" {
			resourceType = ref.resourceType
		}
		if reName != nil {
			// If the regular expression doesn't match, the resource is skipped from the list
			if reName.FindString(ref.name) == "" {
				continue
			}
		}
		switch listMode {
		case "name":
			list = append(list, ref.name)
		case "id":
			list = append(list, ref.id)
		case "name_id":
			list = append(list, ref.name+nameIdSeparator+ref.id)
		case "hierarchy":
			if ref.parent != "" {
				list = append(list, strings.Join(ancestors, nameIdSeparator)+
					nameIdSeparator+ref.parent+
					nameIdSeparator+ref.name)
			} else {
				list = append(list, strings.Join(ancestors, nameIdSeparator)+nameIdSeparator+ref.name)
			}
		case "href":
			list = append(list, ref.href)
		case "import":
			identifier := ref.name
			if ref.importId {
				identifier = ref.id
			}
			list = append(list, fmt.Sprintf("terraform import %s.%s %s%s%s",
				resourceType,
				ref.name,
				strings.Join(ancestors, ImportSeparator),
				ImportSeparator,
				identifier))

			ancestorsText := ""
			if len(ancestors) > 0 {
				ancestorsText = strings.Join(ancestors, ImportSeparator) + ImportSeparator
			}
			importData.WriteString(fmt.Sprintf("# Import directive for %s %s%s \n", resourceType, ancestorsText, ref.name))
			importData.WriteString("import {\n")
			importData.WriteString(fmt.Sprintf("  to = %s.%s-%s\n", resourceType, ref.name, idTail(ref.id)))
			if len(ancestors) > 0 {
				importData.WriteString(fmt.Sprintf("  id = \"%s%s%s\"\n",
					strings.Join(ancestors, ImportSeparator), ImportSeparator, identifier))
			} else {
				importData.WriteString(fmt.Sprintf("  id = \"%s\"\n", identifier))
			}
			importData.WriteString("}\n\n")
		}
	}

	if importFile != "" && listMode == "import" {
		err = os.WriteFile(importFile, []byte(importData.String()), 0600)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func idTail(id string) string {
	if id == "" {
		return ""
	}
	reTail := regexp.MustCompile(`([a-zA-z0-9]+)$`)
	return reTail.FindString(id)
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
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}
	lbServerPoolList, err := edgeGateway.GetLbServerPools()
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway LB server pools '%s': %s ", d.Get("parent").(string), err)
	}
	var items []resourceRef
	for _, service := range lbServerPoolList {
		items = append(items, resourceRef{
			name:   service.Name,
			id:     service.ID,
			href:   "",
			parent: edgeGateway.EdgeGateway.Name,
		})
	}

	return genericResourceList(d, "vcd_lb_server_pool", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbServiceMonitorList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}

	var items []resourceRef
	lbServiceMonitorList, err := edgeGateway.GetLbServiceMonitors()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB service monitor list: %s ", err)
	}
	for _, sm := range lbServiceMonitorList {
		items = append(items, resourceRef{
			name:   sm.Name,
			id:     sm.ID,
			href:   sm.URL,
			parent: edgeGateway.EdgeGateway.Name,
		})
	}
	return genericResourceList(d, "vcd_lb_service_monitor", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbVirtualServerList(d *schema.ResourceData, meta interface{}) (list []string, err error) {

	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}
	var items []resourceRef
	lbVirtualServerList, err := edgeGateway.GetLbVirtualServers()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB virtual server list: %s ", err)
	}
	for _, vs := range lbVirtualServerList {
		items = append(items, resourceRef{
			name:   vs.Name,
			id:     vs.ID,
			href:   "",
			parent: edgeGateway.EdgeGateway.Name,
		})
	}
	return genericResourceList(d, "vcd_lb_virtual_server", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func nsxvFirewallList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}

	var items []resourceRef
	fwRuleList, err := edgeGateway.GetAllNsxvFirewallRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving NSXV firewall rule list: %s ", err)
	}
	for _, fw := range fwRuleList {
		items = append(items, resourceRef{
			name:   fw.Name,
			id:     fw.ID,
			href:   "",
			parent: edgeGateway.EdgeGateway.Name,
		})
	}
	return genericResourceList(d, "vcd_nsxv_firewall_rule", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbAppRuleList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}

	var items []resourceRef
	ruleList, err := edgeGateway.GetLbAppRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB app rule list: %s ", err)
	}
	for _, fw := range ruleList {
		items = append(items, resourceRef{
			name:   fw.Name,
			id:     fw.ID,
			href:   "",
			parent: edgeGateway.EdgeGateway.Name,
		})
	}
	return genericResourceList(d, "vcd_lb_app_rule", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func lbAppProfileList(d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}

	var items []resourceRef
	profiles, err := edgeGateway.GetLbAppProfiles()
	if err != nil {
		return list, fmt.Errorf("error retrieving LB app profile list: %s ", err)
	}
	for _, fw := range profiles {
		items = append(items, resourceRef{
			name:   fw.Name,
			id:     fw.ID,
			href:   "",
			parent: edgeGateway.EdgeGateway.Name,
		})
	}
	return genericResourceList(d, "vcd_lb_app_profile", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func ipsetList(d *schema.ResourceData, meta interface{}) (list []string, err error) {

	client := meta.(*VCDClient)

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
			name:   ipSet.Name,
			id:     ipSet.ID,
			href:   "",
			parent: vdc.Vdc.Name,
		})
	}
	return genericResourceList(d, "vcd_ipset", []string{org.Org.Name, vdc.Vdc.Name}, items)
}

func nsxvNatRuleList(natType string, d *schema.ResourceData, meta interface{}) (list []string, err error) {
	orgName, vdcName, _, _, edgeGateway, err := getEdgeGatewayDetails(d, meta)
	if err != nil {
		return list, fmt.Errorf("error retrieving edge gateway '%s': %s ", d.Get("parent").(string), err)
	}

	var items []resourceRef
	rules, err := edgeGateway.GetNsxvNatRules()
	if err != nil {
		return list, fmt.Errorf("error retrieving NSXV NAT rule list: %s ", err)
	}
	for _, rule := range rules {
		if rule.Action == natType {
			items = append(items, resourceRef{
				name:     "",
				id:       rule.ID,
				href:     "",
				parent:   edgeGateway.EdgeGateway.Name,
				importId: true,
			})
		}
	}
	return genericResourceList(d, "vcd_lb_app_profile", []string{orgName, vdcName, edgeGateway.EdgeGateway.Name}, items)
}

func getResourcesList() ([]string, error) {
	var list []string
	resources := globalResourceMap
	for resource := range resources {
		list = append(list, resource)
	}
	// Returns the list of resources in alphabetical order, to keep a consistent state
	sort.Strings(list)
	return list, nil
}

func datasourceVcdResourceListRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	requested := d.Get("resource_type").(string)
	var err error
	var list []string
	switch requested {
	// Note: do not try to get the data sources list, as it would result in a circular reference
	case "resource", "resources":
		list, err = getResourcesList()
	case "vcd_site_association":
		list, err = getSiteAssociationList(d, meta, "vcd_site_association")
	case "vcd_org_association":
		list, err = getOrgAssociationList(d, meta, "vcd_org_association")
	case "vcd_org", "org", "orgs":
		list, err = getOrgList(d, meta, "vcd_org")
	case "vcd_org_ldap", "vcd_org_saml":
		list, err = getOrgList(d, meta, requested)
	case "vcd_provider_vdc", "provider_vdc":
		list, err = getPvdcList(d, meta)
	case "vcd_distributed_switch":
		list, err = distributedSwitchList(d, meta)
	case "vcd_nsxt_transport_zone":
		list, err = transportZoneList(d, meta)
	case "vcd_importable_port_group":
		list, err = importablePortGroupList(d, meta)
	case "vcd_network_pool":
		list, err = networkPoolList(d, meta)
	case "vcd_vcenter":
		list, err = vcenterList(d, meta)
	case "vcd_nsxt_manager":
		list, err = nsxtManagerList(d, meta)
	case "vcd_external_network", "external_network", "external_networks":
		list, err = externalNetworkList(d, meta)
	case "vcd_org_vdc", "vdc", "vdcs":
		list, err = vdcList(d, meta, "vcd_org_vdc")
	case "vcd_vdc_group":
		list, err = getVdcGroups(d, meta)
	case "vcd_org_vdc_access_control":
		list, err = vdcList(d, meta, "vcd_org_vdc_access_control")
	case "vcd_catalog", "catalog", "catalogs", "vcd_subscribed_catalog":
		list, err = catalogList(d, meta, "vcd_catalog")
	case "vcd_catalog_access_control":
		list, err = catalogList(d, meta, "vcd_catalog_access_control")
	case "vcd_catalog_item", "catalog_item", "catalog_items", "catalogitem", "catalogitems":
		list, err = catalogItemList(d, meta, "vcd_catalog_item")
	case "vcd_catalog_vapp_template", "vapp_template":
		list, err = vappTemplateList(d, meta)
	case "vcd_catalog_media", "catalog_media", "media_items", "mediaitems", "mediaitem":
		list, err = catalogItemList(d, meta, "vcd_catalog_media")
	case "vcd_independent_disk", "disk", "disks":
		list, err = diskList(d, meta)
	case "vcd_vapp", "vapp", "vapps", "vcd_cloned_vapp":
		list, err = vappList(d, meta, "vcd_vapp")
	case "vcd_vapp_access_control":
		list, err = vappList(d, meta, "vcd_vapp_access_control")
	case "vcd_vapp_vm", "vapp_vm", "vapp_vms":
		list, err = vmList(d, meta, vappVmType)
	case "vcd_vm", "standalone_vm":
		list, err = vmList(d, meta, standaloneVmType)
	case "vcd_all_vm", "vm", "vms":
		list, err = vmList(d, meta, "all")
	case "vcd_org_user", "org_user", "user", "users":
		list, err = orgUserList(d, meta)
	case "vcd_edgegateway", "edge_gateway", "edge", "edgegateway":
		list, err = getEdgeGatewayList(d, meta, "vcd_edgegateway")
	case "vcd_edgegateway_settings":
		list, err = getEdgeGatewayList(d, meta, "vcd_edgegateway_settings")
	case "vcd_nsxt_edgegateway", "nsxt_edge_gateway", "nsxt_edge", "nsxt_edgegateway":
		list, err = getNsxtEdgeGatewayList(d, meta)
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
	case "vcd_network_routed_v2", "vcd_network_isolated_v2", "vcd_nsxt_network_imported":
		list, err = orgNetworkListV2(d, meta)
	case "vcd_right", "rights":
		list, err = rightsList(d, meta)
	case "vcd_rights_bundle", "rights_bundle":
		list, err = rightsBundlesList(d, meta)
	case "vcd_role", "roles":
		list, err = rolesList(d, meta)
	case "vcd_global_role", "global_roles":
		list, err = globalRolesList(d, meta)
	case "vcd_library_certificate":
		list, err = libraryCertificateList(d, meta)

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
