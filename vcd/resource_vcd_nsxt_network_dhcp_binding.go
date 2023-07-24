package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtDhcpBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtDhcpBindingCreate,
		ReadContext:   resourceVcdNsxtDhcpBindingRead,
		UpdateContext: resourceVcdNsxtDhcpBindingUpdate,
		DeleteContext: resourceVcdNsxtDhcpBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtDhcpBindingImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of DHCP binding",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address of the DHCP binding",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MAC address of the DHCP binding",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of DHCP binding",
			},
			"binding_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Binding type 'IPV4' or 'IPV6'",
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The DNS server IPs to be assigned . 2 values maximum.",
				MaxItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lease_time": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Lease time in seconds. Minimum value is 60 seconds",
				ValidateFunc: validation.IntAtLeast(60),
			},
			"dhcp_v4_config": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"dhcp_v6_config"},
				Description:   "IPv4 specific DHCP Binding configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway_ip_address": {
							Optional:    true,
							Type:        schema.TypeString,
							Description: "IPv4 gateway address",
						},
						"hostname": {
							Optional:    true,
							Type:        schema.TypeString,
							Description: "Hostname for the DHCP client",
						},
					},
				},
			},
			"dhcp_v6_config": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"dhcp_v4_config"},
				Description:   "IPv6 specific DHCP Binding configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sntp_servers": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of SNTP servers",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"domain_names": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of domain names",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxtDhcpBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentOrgNetwork(d)
	defer vcdClient.unLockParentOrgNetwork(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding create] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding create] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	if !orgVdcNet.IsDhcpEnabled() {
		return diag.Errorf("[NSX-T DHCP binding create] DHCP is not enabled for Org VDC network with ID '%s'. Please use 'vcd_nsxt_network_dhcp resource to enable DHCP", orgNetworkId)
	}

	// get DHCP dhcpBindingConfig config
	dhcpBindingConfig, err := getOpenApiOrgVdcNetworkDhcpBindingType(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding create] error getting DHCP binding configuration: %s", err)
	}

	createdDhcpBinding, err := orgVdcNet.CreateOpenApiOrgVdcNetworkDhcpBinding(dhcpBindingConfig)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding create] error creating DHCP binding for Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	d.SetId(createdDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)

	return resourceVcdNsxtDhcpBindingRead(ctx, d, meta)
}

func resourceVcdNsxtDhcpBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentOrgNetwork(d)
	defer vcdClient.unLockParentOrgNetwork(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding update] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding update] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	dhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding update] error retrieving DHCP binding with ID '%s' for Org VDC network with ID '%s': %s", d.Id(), orgNetworkId, err)
	}

	// get DHCP binding config
	dhcpBindingConfig, err := getOpenApiOrgVdcNetworkDhcpBindingType(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding update] error getting DHCP binding configuration: %s", err)
	}
	dhcpBindingConfig.ID = d.Id()

	_, err = dhcpBinding.Update(dhcpBindingConfig)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding update] error updating DHCP binding for Org VDC network with ID '%s', binding Name '%s', ID '%s': %s", orgNetworkId,
			dhcpBindingConfig.Name, dhcpBindingConfig.ID, err)
	}

	return resourceVcdNsxtDhcpBindingRead(ctx, d, meta)
}

func resourceVcdNsxtDhcpBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding read] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	dhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding read] error retrieving DHCP binding with ID '%s' for Org VDC network with ID '%s': %s", d.Id(), orgNetworkId, err)
	}

	if err := setOpenApiOrgVdcNetworkDhcpBindingData(d, dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding); err != nil {
		return diag.Errorf("[NSX-T DHCP binding read] error setting DHCP binding data: %s", err)
	}
	return nil
}

func resourceVcdNsxtDhcpBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentOrgNetwork(d)
	defer vcdClient.unLockParentOrgNetwork(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding delete] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding delete] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	dhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingById(d.Id())
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding delete] error retrieving DHCP binding with ID '%s' for Org VDC network with ID '%s': %s", d.Id(), orgNetworkId, err)
	}

	err = dhcpBinding.Delete()
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding delete] error deleting DHCP binding with ID '%s' for Org VDC network with ID '%s': %s", d.Id(), orgNetworkId, err)
	}

	return nil
}

func resourceVcdNsxtDhcpBindingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-org-vdc-group-name.org_network_name.my-binding-name")
	}
	orgName, vdcOrVdcGroupName, orgVdcNetworkName, bindingName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("[NSX-T DHCP binding import] DHCP configuration is only supported for NSX-T networks: %s", err)
	}

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(orgVdcNetworkName)
	if err != nil {
		return nil, fmt.Errorf("[NSX-T DHCP binding import] error retrieving Org VDC network with name '%s': %s", orgVdcNetworkName, err)
	}

	if !orgVdcNet.IsDhcpEnabled() {
		return nil, fmt.Errorf("[NSX-T DHCP binding import] DHCP is not enabled for Org VDC network with name '%s'", orgVdcNetworkName)
	}

	dhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingByName(bindingName)
	if err != nil {
		return nil, fmt.Errorf("[NSX-T DHCP binding import] error retrieving DHCP binding with name '%s' for Org VDC network with name '%s': %s", bindingName, orgVdcNetworkName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "org_network_id", orgVdcNet.OpenApiOrgVdcNetwork.ID)
	d.SetId(dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)

	return []*schema.ResourceData{d}, nil
}

func getOpenApiOrgVdcNetworkDhcpBindingType(d *schema.ResourceData) (*types.OpenApiOrgVdcNetworkDhcpBinding, error) {
	bindingConfig := &types.OpenApiOrgVdcNetworkDhcpBinding{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		BindingType: d.Get("binding_type").(string),
		MacAddress:  d.Get("mac_address").(string),
		IpAddress:   d.Get("ip_address").(string),
		LeaseTime:   addrOf(d.Get("lease_time").(int)),
	}

	dnsServers, ok := d.GetOk("dns_servers")
	if ok {
		bindingConfig.DnsServers = convertTypeListToSliceOfStrings(dnsServers.([]interface{}))
	}

	// IPv4 configuration specifics
	dhcpV4Config := d.Get("dhcp_v4_config").([]interface{})
	if len(dhcpV4Config) > 0 {

		dhcpConfig := dhcpV4Config[0].(map[string]interface{})
		bindingConfig.DhcpV4BindingConfig = &types.DhcpV4BindingConfig{
			GatewayIPAddress: dhcpConfig["gateway_ip_address"].(string),
			HostName:         dhcpConfig["hostname"].(string),
		}
	}

	// IPv6 configuration specifics
	dhcpV6Config := d.Get("dhcp_v6_config").([]interface{})
	if len(dhcpV6Config) > 0 {
		dhcpConfig := dhcpV6Config[0].(map[string]interface{})
		domainNames := convertSchemaSetToSliceOfStrings(dhcpConfig["domain_names"].(*schema.Set))
		sntpServers := convertSchemaSetToSliceOfStrings(dhcpConfig["sntp_servers"].(*schema.Set))

		bindingConfig.DhcpV6BindingConfig = &types.DhcpV6BindingConfig{
			DomainNames: domainNames,
			SntpServers: sntpServers,
		}
	}

	return bindingConfig, nil
}

func setOpenApiOrgVdcNetworkDhcpBindingData(d *schema.ResourceData, dhcpBinding *types.OpenApiOrgVdcNetworkDhcpBinding) error {
	dSet(d, "name", dhcpBinding.Name)
	dSet(d, "description", dhcpBinding.Description)
	dSet(d, "binding_type", dhcpBinding.BindingType)
	dSet(d, "mac_address", dhcpBinding.MacAddress)
	dSet(d, "ip_address", dhcpBinding.IpAddress)
	if dhcpBinding.LeaseTime != nil {
		dSet(d, "lease_time", *dhcpBinding.LeaseTime)
	}

	if len(dhcpBinding.DnsServers) > 0 {
		err := d.Set("dns_servers", dhcpBinding.DnsServers)
		if err != nil {
			return fmt.Errorf("error setting DNS servers: %s", err)
		}
	}

	// IPv4 configuration specifics
	if dhcpBinding.DhcpV4BindingConfig != nil {
		dhcpV4Config := make(map[string]string)
		dhcpV4Config["gateway_ip_address"] = dhcpBinding.DhcpV4BindingConfig.GatewayIPAddress
		dhcpV4Config["hostname"] = dhcpBinding.DhcpV4BindingConfig.HostName
		err := d.Set("dhcp_v4_config", []map[string]string{dhcpV4Config})
		if err != nil {
			return fmt.Errorf("error setting DHCPv4 configuration: %s", err)
		}
	}

	// IPv6 configuration specifics
	if dhcpBinding.DhcpV6BindingConfig != nil {
		dhcpV6Config := make(map[string]interface{})
		dhcpV6Config["domain_names"] = dhcpBinding.DhcpV6BindingConfig.DomainNames
		dhcpV6Config["sntp_servers"] = dhcpBinding.DhcpV6BindingConfig.SntpServers
		err := d.Set("dhcp_v6_config", []map[string]interface{}{dhcpV6Config})
		if err != nil {
			return fmt.Errorf("error setting DHCPv6 configuration: %s", err)
		}
	}

	return nil
}
