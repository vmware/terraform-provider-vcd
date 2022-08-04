package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdLBAppProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdLBAppProfileRead,
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
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the LB Application Profile is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LB Application Profile name for lookup",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Protocol type used to send requests to the server. One of 'TCP', " +
					"'UDP', 'HTTP' org 'HTTPS'",
			},
			"enable_ssl_passthrough": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Enable SSL authentication to be passed through to the virtual " +
					"server. Otherwise SSL authentication takes place at the destination address.",
			},
			"http_redirect_url": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The URL to which traffic that arrives at the destination address " +
					"should be redirected. Only applies for types HTTP and HTTPS",
			},
			"persistence_mechanism": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Persistence mechanism for the profile. One of 'cookie', " +
					"'ssl-sessionid', 'sourceip'",
			},
			"cookie_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Used to uniquely identify the session the first time a client " +
					"accesses the site. The load balancer refers to this cookie when connecting " +
					"subsequent requests in the session, so that they all go to the same virtual " +
					"server. Only applies for persistence_mechanism 'cookie'",
			},
			"cookie_mode": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The mode by which the cookie should be inserted. One of 'insert', " +
					"'prefix', or 'appsession'",
			},
			"expiration": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Length of time in seconds that persistence stays in effect",
			},
			"insert_x_forwarded_http_header": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Enables 'X-Forwarded-For' header for identifying the originating IP" +
					" address of a client connecting to a Web server through the load balancer. " +
					"Only applies for types HTTP and HTTPS",
			},
			// TODO https://github.com/vmware/terraform-provider-vcd/issues/258
			// This will not give much use without SSL certs being available
			"enable_pool_side_ssl": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Enable to define the certificate, CAs, or CRLs used to authenticate" +
					" the load balancer from the server side",
			},
		},
	}
}

func datasourceVcdLBAppProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBAppProfile, err := edgeGateway.GetLbAppProfileByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("unable to find load balancer application profile with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBAppProfile.ID)
	err = setLBAppProfileData(d, readLBAppProfile)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
