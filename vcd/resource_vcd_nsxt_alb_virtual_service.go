package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbVirtualService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbVirtualServiceCreate,
		ReadContext:   resourceVcdAlbVirtualServiceRead,
		UpdateContext: resourceVcdAlbVirtualServiceUpdate,
		DeleteContext: resourceVcdAlbVirtualServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbVirtualServiceImport,
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
				Description: "Name of ALB Virtual Service",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of ALB Virtual Service",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID in which ALB Pool should be created",
			},
			"service_engine_group_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID in which ALB Pool should be created",
			},
			"ca_certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A set of root certificate IDs to use when validating certificates presented by pool members",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Virtual Service is enabled or disabled (default true)",
			},
			"virtual_ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Virtual IP address (VIP) for Virtual Service",
			},
			"application_profile_type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "HTTP, HTTPS, L4, L4_TLS",
			},
			"service_port": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     nsxtAlbVirtualServicePort,
			},
		},
	}
}

var nsxtAlbVirtualServicePort = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_port": &schema.Schema{
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Starting port in the range",
		},
		"end_port": &schema.Schema{
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Starting port in the range",
		},
		"ssl_enabled": &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Starting port in the range",
		},
		"type": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "One of 'TCP_PROXY', 'TCP_FAST_PATH', 'UDP_FAST_PATH'",
		},
	},
}

func resourceVcdAlbVirtualServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albVirtualServiceConfig, err := getNsxtAlbVirtualServiceType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T ALB Virtual Service type: %s", err)
	}
	createdAlbVirtualService, err := vcdClient.CreateNsxtAlbVirtualService(albVirtualServiceConfig)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Virtual Service: %s", err)
	}

	d.SetId(createdAlbVirtualService.NsxtAlbVirtualService.ID)

	return resourceVcdAlbVirtualServiceRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	updateVirtualServiceConfig, err := getNsxtAlbVirtualServiceType(d)
	if err != nil {
		return diag.Errorf("error getting NSX-T ALB Virtual Service type: %s", err)
	}
	updateVirtualServiceConfig.ID = d.Id()

	_, err = albVirtualService.Update(updateVirtualServiceConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating NSX-T ALB Virtual Service: %s", err))
	}

	return resourceVcdAlbVirtualServiceRead(ctx, d, meta)
}

func resourceVcdAlbVirtualServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	albVirtualService, err := vcdClient.GetAlbVirtualServiceById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	err = setNsxtAlbVirtualServiceData(d, albVirtualService.NsxtAlbVirtualService)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Virtual Service data: %s", err)
	}
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)
	return nil
}

func resourceVcdAlbVirtualServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	albPool, err := vcdClient.GetAlbVirtualServiceById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T ALB Virtual Service: %s", err))
	}

	err = albPool.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T ALB Virtual Service: %s", err)
	}

	return nil
}

func resourceVcdAlbVirtualServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Virtual Service import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name.virtual_service_name")
	}
	orgName, vdcName, edgeName, virtualServiceName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("ALB Virtual Services are only supported on NSX-T. Please use 'vcd_lb_virtual_server' for NSX-V load balancers")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	albVirtualService, err := vcdClient.GetAlbVirtualServiceByName(edge.EdgeGateway.ID, virtualServiceName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T ALB Virtual Service '%s': %s", virtualServiceName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)

	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtAlbVirtualServiceType(d *schema.ResourceData) (*types.NsxtAlbVirtualService, error) {
	albVirtualServiceConfig := &types.NsxtAlbVirtualService{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		ApplicationProfile:    types.NsxtAlbVirtualServiceApplicationProfile{Type: d.Get("application_profile_type").(string)},
		Enabled:               takeBoolPointer(d.Get("enabled").(bool)),
		GatewayRef:            types.OpenApiReference{ID: d.Get("edge_gateway_id").(string)},
		ServiceEngineGroupRef: types.OpenApiReference{ID: d.Get("service_engine_group_id").(string)},
		LoadBalancerPoolRef:   types.OpenApiReference{ID: d.Get("pool_id").(string)},
		VirtualIpAddress:      d.Get("virtual_ip_address").(string),
	}
	// Certificate must only be set if it is specified as the API throws error
	if d.Get("ca_certificate_id").(string) != "" {
		albVirtualServiceConfig.CertificateRef = &types.OpenApiReference{ID: d.Get("ca_certificate_id").(string)}
	}

	servicePorts, err := getNsxtAlbVirtualServicePortType(d)
	if err != nil {
		return nil, fmt.Errorf("error getting virtual service port definition: %s", err)
	}
	albVirtualServiceConfig.ServicePorts = servicePorts

	return albVirtualServiceConfig, nil
}

func getNsxtAlbVirtualServicePortType(d *schema.ResourceData) ([]types.NsxtAlbVirtualServicePort, error) {
	servicePortSet := d.Get("service_port").(*schema.Set)
	servicePortSlice := make([]types.NsxtAlbVirtualServicePort, len(servicePortSet.List()))

	for hmIndex, healthMonitor := range servicePortSet.List() {
		servicePortMap := healthMonitor.(map[string]interface{})
		singleVirtualServicePort := types.NsxtAlbVirtualServicePort{
			PortStart:  takeIntPointer(servicePortMap["start_port"].(int)),
			SslEnabled: takeBoolPointer(servicePortMap["ssl_enabled"].(bool)),
			TcpUdpProfile: &types.NsxtAlbVirtualServicePortTcpUdpProfile{
				SystemDefined: true,
				Type:          servicePortMap["type"].(string),
			},
		}

		// End port is optional for single ports
		if servicePortMap["end_port"].(int) != 0 {
			singleVirtualServicePort.PortEnd = takeIntPointer(servicePortMap["end_port"].(int))
		}

		servicePortSlice[hmIndex] = singleVirtualServicePort
	}

	return servicePortSlice, nil
}

func setNsxtAlbVirtualServiceData(d *schema.ResourceData, albVirtualService *types.NsxtAlbVirtualService) error {
	dSet(d, "name", albVirtualService.Name)
	dSet(d, "description", albVirtualService.Description)
	dSet(d, "enabled", &albVirtualService.Enabled)
	dSet(d, "edge_gateway_id", albVirtualService.GatewayRef.ID)
	dSet(d, "pool_id", albVirtualService.LoadBalancerPoolRef.ID)
	dSet(d, "service_engine_group_id", albVirtualService.ServiceEngineGroupRef.ID)
	dSet(d, "virtual_ip_address", albVirtualService.VirtualIpAddress)

	dSet(d, "application_profile_type", albVirtualService.ApplicationProfile.Type)

	// Optional fields
	if albVirtualService.CertificateRef != nil {
		dSet(d, "ca_certificate_id", albVirtualService.CertificateRef.ID)
	}

	err := setNsxtAlbVirtualServicePortData(d, albVirtualService.ServicePorts)
	if err != nil {
		return err
	}

	return nil
}

func setNsxtAlbVirtualServicePortData(d *schema.ResourceData, ports []types.NsxtAlbVirtualServicePort) error {
	portSlice := make([]interface{}, len(ports))
	for i, port := range ports {
		portMap := make(map[string]interface{})
		if port.PortStart != nil {
			portMap["start_port"] = *port.PortStart
		}

		if port.PortEnd != nil && *port.PortStart != *port.PortEnd {
			portMap["end_port"] = *port.PortEnd
		}
		if port.SslEnabled != nil {
			portMap["ssl_enabled"] = *port.SslEnabled
		}

		if port.TcpUdpProfile != nil {
			portMap["type"] = port.TcpUdpProfile.Type
		}

		portSlice[i] = portMap
	}
	subnetSet := schema.NewSet(schema.HashResource(nsxtAlbVirtualServicePort), portSlice)
	err := d.Set("service_port", subnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'service_port' block: %s", err)
	}
	return nil
}
