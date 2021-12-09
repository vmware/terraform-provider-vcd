package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbPoolCreate,
		ReadContext:   resourceVcdAlbPoolRead,
		UpdateContext: resourceVcdAlbPoolUpdate,
		DeleteContext: resourceVcdAlbPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbPoolImport,
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
				Description: "Edge gateway ID in which ALB Pool should be created",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of ALB Pool",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Boolean value if ALB Pool is enabled or not (default true)",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of ALB Pool",
			},
			"algorithm": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Algorithm for choosing pool members (default LEAST_CONNECTIONS)",
				// Default is LEAST_CONNECTIONS even if no value is sent
				Default: "LEAST_CONNECTIONS",
			},
			"default_port": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Default Port defines destination server port used by the traffic sent to the member (default 80)",
				// Default even if no value is sent
				Default: 80,
			},
			//
			"graceful_timeout_period": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum time in minutes to gracefully disable pool member (default 1)",
				// Default even if no value is sent
				Default: 1,
			},
			"member": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        nsxtAlbPoolMember,
				Description: "ALB Pool Members",
			},
			"health_monitor": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     nsxtAlbPoolHealthMonitor,
			},
			"persistence_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     nsxtAlbPoolPersistenceProfile,
			},
			"ca_certificate_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of root certificate IDs to use when validating certificates presented by pool members",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cn_check_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean flag if common name check of the certificate should be enabled",
			},
			"domain_names": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of domain names which will be used to verify common names",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"passive_monitoring_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Monitors if the traffic is accepted by node (default true)",
			},
			// Read only information
			"associated_virtual_service_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IDs of associated virtual services",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"associated_virtual_services": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Names of associated virtual services",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"member_count": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of members in the pool",
			},
			"up_member_count": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of members in the pool serving the traffic",
			},
			"enabled_member_count": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of enabled members in the pool",
			},
			"health_message": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Health message",
			},
		},
	}
}

var nsxtAlbPoolMember = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"enabled": &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Defines if pool member is accepts traffic (default 'true')",
		},
		"ip_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "IP address of pool member",
		},
		"port": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "Member port",
			ValidateFunc: validation.IntAtLeast(1),
			Default:      nil,
		},
		"ratio": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1, // Such value is set even if it is not sent
			Description:  "Ratio of selecting eligible servers in the pool",
			ValidateFunc: validation.IntAtLeast(1),
		},
		"marked_down_by": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Marked down by one or more health monitors",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"health_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Health status",
		},
		"detailed_health_message": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Detailed health message",
		},
	},
}

var nsxtAlbPoolHealthMonitor = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "Type of health monitor",
		},
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"system_defined": &schema.Schema{
			Type:     schema.TypeBool,
			Computed: true,
		},
	},
}

var nsxtAlbPoolPersistenceProfile = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "System generated name of persistence profile",
		},
		"type": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "Type of persistence strategy",
		},
		"value": &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Value of attribute based on persistence type",
		},
	},
}

func resourceVcdAlbPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albPoolConfig, err := getNsxtAlbPoolType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T ALB Pool type: %s", err)
	}
	createdAlbPool, err := vcdClient.CreateNsxtAlbPool(albPoolConfig)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Pool: %s", err)
	}

	d.SetId(createdAlbPool.NsxtAlbPool.ID)

	return resourceVcdAlbPoolRead(ctx, d, meta)
}

func resourceVcdAlbPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albPool, err := vcdClient.GetAlbPoolById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Pool: %s", err))
	}

	updatePoolConfig, err := getNsxtAlbPoolType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T ALB Pool type: %s", err)
	}
	updatePoolConfig.ID = d.Id()

	_, err = albPool.Update(updatePoolConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating NSX-T ALB Pool: %s", err))
	}

	return resourceVcdAlbPoolRead(ctx, d, meta)
}

func resourceVcdAlbPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albPool, err := vcdClient.GetAlbPoolById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Pool: %s", err))
	}

	err = setNsxtAlbPoolData(d, albPool.NsxtAlbPool)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Pool data: %s", err)
	}
	d.SetId(albPool.NsxtAlbPool.ID)
	return nil
}

func resourceVcdAlbPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albPool, err := vcdClient.GetAlbPoolById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Pool: %s", err))
	}

	err = albPool.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T ALB Pool: %s", err)
	}

	return nil
}

func resourceVcdAlbPoolImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Pool import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name.pool_name")
	}
	orgName, vdcName, edgeName, poolName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("ALB Pools are only supported on NSX-T please use 'vcd_lb_server_pool' for NSX-V load balancers")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	albPools, err := vcdClient.GetAllAlbPools(edge.EdgeGateway.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T ALB Pool '%s': %s", poolName, err)
	}

	if len(albPools) < 1 {
		return nil, fmt.Errorf("ALB Pool with name '%s' in Edge Gateway '%s' not found", poolName, edgeName)
	}

	if len(albPools) > 1 {
		return nil, fmt.Errorf("found more than one (%d) ALB Pool with name '%s' in Edge Gateway '%s'",
			len(albPools), poolName, edgeName)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)

	d.SetId(albPools[0].NsxtAlbPool.ID)

	return []*schema.ResourceData{d}, nil
}

// getNsxtAlbPoolType is the main function for getting *types.NsxtAlbPool for API request. It nests multiple smaller
// functions for smaller types.
func getNsxtAlbPoolType(d *schema.ResourceData) (*types.NsxtAlbPool, error) {
	albPoolConfig := &types.NsxtAlbPool{
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		Enabled:                  takeBoolPointer(d.Get("enabled").(bool)),
		GatewayRef:               types.OpenApiReference{ID: d.Get("edge_gateway_id").(string)},
		Algorithm:                d.Get("algorithm").(string),
		DefaultPort:              takeIntPointer(d.Get("default_port").(int)),
		GracefulTimeoutPeriod:    takeIntPointer(d.Get("graceful_timeout_period").(int)),
		PassiveMonitoringEnabled: takeBoolPointer(d.Get("passive_monitoring_enabled").(bool)),
	}

	poolMembers, err := getNsxtAlbPoolMembersType(d)
	if err != nil {
		return nil, fmt.Errorf("error defining pool members: %s", err)
	}
	albPoolConfig.Members = poolMembers

	persistenceProfile, err := getNsxtAlbPoolPersistenceProfileType(d)
	if err != nil {
		return nil, fmt.Errorf("error defining persistence profile: %s", err)
	}
	albPoolConfig.PersistenceProfile = persistenceProfile

	healthMonitors, err := getNsxtAlbPoolHealthMonitorType(d)
	if err != nil {
		return nil, fmt.Errorf("error defining health monitors: %s", err)
	}
	albPoolConfig.HealthMonitors = healthMonitors

	caCertificateRefs, commonNameCheckEnabled, domainNames := getCertificateTypes(d)
	albPoolConfig.CaCertificateRefs = caCertificateRefs
	albPoolConfig.CommonNameCheckEnabled = takeBoolPointer(commonNameCheckEnabled)
	albPoolConfig.DomainNames = domainNames

	return albPoolConfig, nil
}

// setNsxtAlbPoolData is the main function for storing type details into statefile. It nests multiple smaller functions
// for separate TypeSet and TypeList blocks.
func setNsxtAlbPoolData(d *schema.ResourceData, albPool *types.NsxtAlbPool) error {
	dSet(d, "name", albPool.Name)
	dSet(d, "description", albPool.Description)
	dSet(d, "edge_gateway_id", albPool.GatewayRef.ID)
	dSet(d, "enabled", albPool.Enabled)
	dSet(d, "algorithm", albPool.Algorithm)
	dSet(d, "default_port", albPool.DefaultPort)
	dSet(d, "graceful_timeout_period", albPool.GracefulTimeoutPeriod)
	dSet(d, "passive_monitoring_enabled", albPool.PassiveMonitoringEnabled)

	err := setNsxtAlbPoolMemberData(d, albPool.Members)
	if err != nil {
		return fmt.Errorf("error storing ALB Pool Members: %s", err)
	}

	err = setNsxtAlbPoolPersistenceProfileData(d, albPool.PersistenceProfile)
	if err != nil {
		return fmt.Errorf("error storing ALB Pool Persistence Profile: %s", err)
	}

	err = setNsxtAlbPoolHealthMonitorData(d, albPool.HealthMonitors)
	if err != nil {
		return fmt.Errorf("error storing ALB Pool Health Monitors: %s", err)
	}

	err = setCertificateData(d, albPool)
	if err != nil {
		return fmt.Errorf("error storing ALB Pool Certificate data: %s", err)
	}

	// Computed only variables below
	dSet(d, "member_count", albPool.MemberCount)
	dSet(d, "up_member_count", albPool.UpMemberCount)
	dSet(d, "enabled_member_count", albPool.EnabledMemberCount)
	dSet(d, "health_message", albPool.HealthMessage)
	dSet(d, "name", albPool.Name)

	associatedVirtualServiceIds := extractIdsFromOpenApiReferences(albPool.VirtualServiceRefs)
	associatedVirtualServiceNames := extractNamesFromOpenApiReferences(albPool.VirtualServiceRefs)
	associatedVirtualServiceIdsSet := convertStringsTotTypeSet(associatedVirtualServiceIds)
	associatedVirtualServiceNameSet := convertStringsTotTypeSet(associatedVirtualServiceNames)
	err = d.Set("associated_virtual_service_ids", associatedVirtualServiceIdsSet)
	if err != nil {
		return fmt.Errorf("error setting 'associated_virtual_service_ids': %s", err)
	}
	err = d.Set("associated_virtual_services", associatedVirtualServiceNameSet)
	if err != nil {
		return fmt.Errorf("error setting 'associated_virtual_service_ids': %s", err)
	}

	return nil
}

func getCertificateTypes(d *schema.ResourceData) ([]types.OpenApiReference, bool, []string) {
	certificatedIds := convertSchemaSetToSliceOfStrings(d.Get("ca_certificate_ids").(*schema.Set))
	certOpenApiRefs := convertSliceOfStringsToOpenApiReferenceIds(certificatedIds)

	cnCheckEnabled := d.Get("cn_check_enabled").(bool)

	domainNames := convertSchemaSetToSliceOfStrings(d.Get("domain_names").(*schema.Set))

	return certOpenApiRefs, cnCheckEnabled, domainNames
}

func getNsxtAlbPoolMembersType(d *schema.ResourceData) ([]types.NsxtAlbPoolMember, error) {
	members := d.Get("member").(*schema.Set)
	memberSlice := make([]types.NsxtAlbPoolMember, len(members.List()))
	for memberIndex, memberDefinition := range members.List() {
		memberMap := memberDefinition.(map[string]interface{})

		member := types.NsxtAlbPoolMember{
			Enabled:   memberMap["enabled"].(bool),
			IpAddress: memberMap["ip_address"].(string),
			Ratio:     takeIntPointer(memberMap["ratio"].(int)),
			Port:      memberMap["port"].(int),
		}

		memberSlice[memberIndex] = member
	}
	return memberSlice, nil
}

func getNsxtAlbPoolHealthMonitorType(d *schema.ResourceData) ([]types.NsxtAlbPoolHealthMonitor, error) {
	healthMonitors := d.Get("health_monitor").(*schema.Set)
	healthMonitorSlice := make([]types.NsxtAlbPoolHealthMonitor, len(healthMonitors.List()))

	for hmIndex, healthMonitor := range healthMonitors.List() {
		healthMonitorMap := healthMonitor.(map[string]interface{})
		singleHealthMonitor := types.NsxtAlbPoolHealthMonitor{
			Type: healthMonitorMap["type"].(string),
		}
		healthMonitorSlice[hmIndex] = singleHealthMonitor
	}

	return healthMonitorSlice, nil
}

func getNsxtAlbPoolPersistenceProfileType(d *schema.ResourceData) (*types.NsxtAlbPoolPersistenceProfile, error) {
	if _, isSet := d.GetOk("persistence_profile"); !isSet {
		util.Logger.Printf("[NSX-T ALB Pool Create] Persistence Profile is not set")
		return nil, nil
	}

	persistenceProfileSlice := d.Get("persistence_profile").([]interface{})
	if len(persistenceProfileSlice) < 1 {
		util.Logger.Printf("[NSX-T ALB Pool Create] Persistence Profile has 0 elements")
		return nil, nil
	}

	persistenceProfileMap := persistenceProfileSlice[0].(map[string]interface{})

	persistenceProfile := &types.NsxtAlbPoolPersistenceProfile{}
	persistenceProfile.Type = persistenceProfileMap["type"].(string)
	persistenceProfile.Value = persistenceProfileMap["value"].(string)

	return persistenceProfile, nil
}

func setNsxtAlbPoolMemberData(d *schema.ResourceData, members []types.NsxtAlbPoolMember) error {
	// Loop over all subnets (known as ip_scope in UI)
	memberSlice := make([]interface{}, len(members))
	for i, member := range members {
		memberMap := make(map[string]interface{})
		memberMap["enabled"] = member.Enabled
		memberMap["ip_address"] = member.IpAddress
		memberMap["port"] = member.Port

		if member.Ratio != nil {
			memberMap["ratio"] = *member.Ratio
		}

		memberMap["marked_down_by"] = convertStringsTotTypeSet(member.MarkedDownBy)
		memberMap["health_status"] = member.HealthStatus
		memberMap["detailed_health_message"] = member.DetailedHealthMessage

		memberSlice[i] = memberMap
	}
	subnetSet := schema.NewSet(schema.HashResource(nsxtAlbPoolMember), memberSlice)
	err := d.Set("member", subnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'member' block: %s", err)
	}
	return nil
}

func setNsxtAlbPoolHealthMonitorData(d *schema.ResourceData, healthMonitors []types.NsxtAlbPoolHealthMonitor) error {
	memberSlice := make([]interface{}, len(healthMonitors))
	for i, healthMonitor := range healthMonitors {
		hm := make(map[string]interface{})
		hm["name"] = healthMonitor.Name
		hm["type"] = healthMonitor.Type
		hm["system_defined"] = healthMonitor.SystemDefined

		memberSlice[i] = hm
	}
	subnetSet := schema.NewSet(schema.HashResource(nsxtAlbPoolHealthMonitor), memberSlice)
	err := d.Set("health_monitor", subnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'member' block: %s", err)
	}
	return nil
}

func setNsxtAlbPoolPersistenceProfileData(d *schema.ResourceData, persistenceProfile *types.NsxtAlbPoolPersistenceProfile) error {
	if persistenceProfile == nil {
		return nil
	}

	persistenceProfileSlice := make([]interface{}, 1)
	persistenceProfileMap := make(map[string]interface{})

	if persistenceProfile != nil {
		persistenceProfileMap["name"] = persistenceProfile.Name
		persistenceProfileMap["type"] = persistenceProfile.Type
		persistenceProfileMap["value"] = persistenceProfile.Value
	}
	persistenceProfileSlice[0] = persistenceProfileMap

	err := d.Set("persistence_profile", persistenceProfileSlice)
	if err != nil {
		return fmt.Errorf("error setting 'persistence_profile' block: %s", err)
	}
	return nil
}

func setCertificateData(d *schema.ResourceData, albPool *types.NsxtAlbPool) error {
	if albPool.CaCertificateRefs != nil {
		certIds := extractIdsFromOpenApiReferences(albPool.CaCertificateRefs)
		certIdSet := convertStringsTotTypeSet(certIds)
		err := d.Set("ca_certificate_ids", certIdSet)
		if err != nil {
			return fmt.Errorf("error setting 'ca_certificate_ids': %s", err)
		}
	} else {
		dSet(d, "ca_certificate_ids", nil)
	}

	if albPool.CommonNameCheckEnabled != nil {
		dSet(d, "cn_check_enabled", *albPool.CommonNameCheckEnabled)
	} else {
		dSet(d, "cn_check_enabled", false)
	}

	if albPool.DomainNames != nil {
		domainNameSet := convertStringsTotTypeSet(albPool.DomainNames)
		err := d.Set("domain_names", domainNameSet)
		if err != nil {
			return fmt.Errorf("error setting 'domain_names': %s", err)
		}
	} else {
		dSet(d, "domain_names", nil)
	}

	return nil
}
