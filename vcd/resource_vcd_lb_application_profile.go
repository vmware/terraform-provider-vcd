package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdLBAppProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdLBAppProfileCreate,
		Read:   resourceVcdLBAppProfileRead,
		Update: resourceVcdLBAppProfileUpdate,
		Delete: resourceVcdLBAppProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdLBAppProfileImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the Application Profile is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the Application Profile is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the Application Profile is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique Application Profile name",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "Protocol type used to send requests to the server. One of 'TCP', " +
					"'UDP', 'HTTP' org 'HTTPS'",
			},
			"enable_ssl_passthrough": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Enable SSL authentication to be passed through to the virtual " +
					"server. Otherwise SSL authentication takes place at the destination address.",
			},
			"http_redirect_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "The URL to which traffic that arrives at the destination address " +
					"should be redirected. Only applies for types HTTP and HTTPS",
			},
			"persistence_mechanism": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Persistence mechanism for the profile. One of 'COOKIE', ''",
			},
			"cookie_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "Used to uniquely identify the session the first time a client " +
					"accesses the site. The load balancer refers to this cookie when connecting " +
					"subsequent requests in the session, so that they all go to the same virtual " +
					"server. Only applies for persistence_mechanism 'COOKIE'",
			},
			"cookie_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "The mode by which the cookie should be inserted. One of 'Insert', " +
					"'PREFIX', or 'APPSESSION'",
			},
			"expiration": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Length of time in seconds that persistence stays in effect",
			},
			"insert_x_forwarded_http_header": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Enables 'X-Forwarded-For' header for identifying the originating IP" +
					" address of a client connecting to a Web server through the load balancer. " +
					"Only applies for types HTTP and HTTPS",
			},
			// This will not give much use without SSL certs being available
			// "enable_pool_side_ssl": &schema.Schema{
			// 	Type:        schema.TypeBool,
			// 	Default:     false,
			// 	Optional:    true,
			// 	Description: ". Only applies for type HTTPS",
			// },
		},
	}
}

func resourceVcdLBAppProfileCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	LBProfile, err := expandLBProfile(d)
	if err != nil {
		return fmt.Errorf("unable to expand load balancer application profile: %s", err)
	}

	createdPool, err := edgeGateway.CreateLBAppProfile(LBProfile)
	if err != nil {
		return fmt.Errorf("error creating new load balancer application profile: %s", err)
	}

	// We store the values once again because response include pool member IDs
	flattenLBProfile(d, createdPool)
	d.SetId(createdPool.ID)
	return nil
}

func resourceVcdLBAppProfileRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBProfile, err := edgeGateway.ReadLBAppProfile(&types.LBAppProfile{ID: d.Id()})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer application profile with ID %s: %s", d.Id(), err)
	}

	return flattenLBProfile(d, readLBProfile)
}

func resourceVcdLBAppProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBProfileConfig, err := expandLBProfile(d)
	if err != nil {
		return fmt.Errorf("could not expand load balancer application profile for update: %s", err)
	}

	updatedLBProfile, err := edgeGateway.UpdateLBAppProfile(updateLBProfileConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer application profile with ID %s: %s", d.Id(), err)
	}

	if err := flattenLBProfile(d, updatedLBProfile); err != nil {
		return err
	}

	return nil
}

func resourceVcdLBAppProfileDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteLBAppProfile(&types.LBAppProfile{ID: d.Id()})
	if err != nil {
		return fmt.Errorf("error deleting load balancer application profile: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdLBAppProfileImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way my-org.my-org-vdc.my-edge-gw.existing-app-profile")
	}
	orgName, vdcName, edgeName, poolName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBProfile, err := edgeGateway.ReadLBAppProfile(&types.LBAppProfile{Name: poolName})
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer application profile with name %s: %s", d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", poolName)

	d.SetId(readLBProfile.ID)
	return []*schema.ResourceData{d}, nil
}

func expandLBProfile(d *schema.ResourceData) (*types.LBAppProfile, error) {
	LBProfile := &types.LBAppProfile{
		Name:     d.Get("name").(string),
		Template: d.Get("type").(string),
		Persistence: &types.LBAppProfilePersistence{
			Method:     d.Get("persistence_mechanism").(string),
			CookieName: d.Get("cookie_name").(string),
			CookieMode: d.Get("cookie_mode").(string),
		},
		SSLPassthrough:                d.Get("enable_ssl_passthrough").(bool),
		InsertXForwardedForHTTPHeader: d.Get("insert_x_forwarded_http_header").(bool),
		ServerSSLEnabled:              d.Get("enable_pool_side_ssl").(bool),
		HTTPRedirect: &types.LBAppProfileHTTPRedirect{
			To: d.Get("http_redirect_url").(string),
		},
	}

	return LBProfile, nil
}

func flattenLBProfile(d *schema.ResourceData, LBProfile *types.LBAppProfile) error {
	d.Set("name", LBProfile.Name)
	d.Set("type", LBProfile.Template)
	d.Set("enable_ssl_passthrough", LBProfile.SSLPassthrough)
	d.Set("http_redirect_url", LBProfile.HTTPRedirect.To)
	d.Set("insert_x_forwarded_http_header", LBProfile.InsertXForwardedForHTTPHeader)
	d.Set("persistence_mechanism", LBProfile.Persistence.Method)
	d.Set("cookie_name", LBProfile.Persistence.CookieName)
	d.Set("cookie_mode", LBProfile.Persistence.CookieMode)
	d.Set("enable_pool_side_ssl", LBProfile.ServerSSLEnabled)
	return nil
}
