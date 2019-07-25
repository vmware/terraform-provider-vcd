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
				Description: "vCD organization in which the LB Application Profile is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the LB Application Profile is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the LB Application Profile is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique LB Application Profile name",
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Protocol type used to send requests to the server. One of 'tcp', " +
					"'udp', 'http' org 'https'",
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
					"should be redirected. Only applies for types 'http' and 'https'",
			},
			"persistence_mechanism": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Persistence mechanism for the profile. One of 'cookie', " +
					"'ssl-sessionid', 'sourceip'",
			},
			"cookie_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "Used to uniquely identify the session the first time a client " +
					"accesses the site. The load balancer refers to this cookie when connecting " +
					"subsequent requests in the session, so that they all go to the same virtual " +
					"server. Only applies for persistence_mechanism 'cookie'",
			},
			"cookie_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "The mode by which the cookie should be inserted. One of 'insert', " +
					"'prefix', or 'appsession'",
			},
			"expiration": &schema.Schema{
				Type:        schema.TypeInt,
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
			// TODO https://github.com/terraform-providers/terraform-provider-vcd/issues/258
			// This will not give much use without SSL certs being available. The only method to
			// make use of it is by manually attaching certificates.
			"enable_pool_side_ssl": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				Description: "Enable to define the certificate, CAs, or CRLs used to authenticate" +
					" the load balancer from the server side",
			},
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

	LBProfile, err := getLBAppProfileType(d)
	if err != nil {
		return fmt.Errorf("unable to create load balancer application profile type: %s", err)
	}

	createdPool, err := edgeGateway.CreateLBAppProfile(LBProfile)
	if err != nil {
		return fmt.Errorf("error creating new load balancer application profile: %s", err)
	}

	// We store the values once again because response include pool member IDs
	err = setLBAppProfileData(d, createdPool)
	if err != nil {
		return err
	}
	d.SetId(createdPool.Id)
	return nil
}

func resourceVcdLBAppProfileRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBProfile, err := edgeGateway.ReadLBAppProfileByID(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer application profile with Id %s: %s", d.Id(), err)
	}

	return setLBAppProfileData(d, readLBProfile)
}

func resourceVcdLBAppProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBProfileConfig, err := getLBAppProfileType(d)
	if err != nil {
		return fmt.Errorf("unable to create load balancer application profile type for update: %s", err)
	}

	updatedLBProfile, err := edgeGateway.UpdateLBAppProfile(updateLBProfileConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer application profile with Id %s: %s", d.Id(), err)
	}

	if err := setLBAppProfileData(d, updatedLBProfile); err != nil {
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

	err = edgeGateway.DeleteLBAppProfileByID(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting load balancer application profile: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceVcdLBAppProfileImport is responsible for importing the resource.
// The d.Id() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
//
// Example import path (id): org.vdc.edge-gw.existing-app-profile
func resourceVcdLBAppProfileImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw.existing-app-profile")
	}
	orgName, vdcName, edgeName, appProfileName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBProfile, err := edgeGateway.ReadLBAppProfileByName(appProfileName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer application profile with name %s: %s",
			d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", appProfileName)

	d.SetId(readLBProfile.Id)
	return []*schema.ResourceData{d}, nil
}

func getLBAppProfileType(d *schema.ResourceData) (*types.LbAppProfile, error) {
	LBProfile := &types.LbAppProfile{
		Name:                          d.Get("name").(string),
		Template:                      d.Get("type").(string),
		SslPassthrough:                d.Get("enable_ssl_passthrough").(bool),
		InsertXForwardedForHttpHeader: d.Get("insert_x_forwarded_http_header").(bool),
		ServerSslEnabled:              d.Get("enable_pool_side_ssl").(bool),
	}

	if d.Get("http_redirect_url").(string) != "" {
		LBProfile.HttpRedirect = &types.LbAppProfileHttpRedirect{
			To: d.Get("http_redirect_url").(string),
		}
	}

	if d.Get("persistence_mechanism").(string) != "" {
		LBProfile.Persistence = &types.LbAppProfilePersistence{
			Method:     d.Get("persistence_mechanism").(string),
			CookieName: d.Get("cookie_name").(string),
			CookieMode: d.Get("cookie_mode").(string),
			Expire:     d.Get("expiration").(int),
		}
	}

	return LBProfile, nil
}

func setLBAppProfileData(d *schema.ResourceData, LBProfile *types.LbAppProfile) error {
	d.Set("name", LBProfile.Name)
	d.Set("type", LBProfile.Template)
	d.Set("enable_ssl_passthrough", LBProfile.SslPassthrough)
	d.Set("insert_x_forwarded_http_header", LBProfile.InsertXForwardedForHttpHeader)
	d.Set("enable_pool_side_ssl", LBProfile.ServerSslEnabled)
	// Questionable field. UI has it, but does not send it. NSX documentation has it, but it is
	// never returned, nor shown
	// d.Set("expiration", LBProfile.Expire)

	if LBProfile.Persistence != nil {
		d.Set("persistence_mechanism", LBProfile.Persistence.Method)
		d.Set("cookie_name", LBProfile.Persistence.CookieName)
		d.Set("cookie_mode", LBProfile.Persistence.CookieMode)
		d.Set("expiration", LBProfile.Persistence.Expire)
	} else {
		d.Set("persistence_mechanism", "")
		d.Set("cookie_name", "")
		d.Set("cookie_mode", "")
		d.Set("expiration", "")
	}

	if LBProfile.HttpRedirect != nil {
		d.Set("http_redirect_url", LBProfile.HttpRedirect.To)
	} else { // We still want to make sure it is empty
		d.Set("http_redirect_url", "")
	}

	return nil
}
