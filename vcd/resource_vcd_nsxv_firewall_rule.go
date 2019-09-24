package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVcdNsxvFirewall() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvFirewallCreate,
		Read:   resourceVcdNsxvFirewallRead,
		Update: resourceVcdNsxvFirewallUpdate,
		Delete: resourceVcdNsxvFirewallDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvFirewallImport,
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
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Firewall rule name",
			},

			"source": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Optional: true,
							Type:     schema.TypeBool,
						},
						"source_ip": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"source_network_id": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},

			"destination": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Optional: true,
							Type:     schema.TypeBool,
						},
						"source_ip": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"source_network_id": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
		},
	}
}

// resourceVcdNsxvFirewallCreate
func resourceVcdNsxvFirewallCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvFirewallUpdate
func resourceVcdNsxvFirewallUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvFirewallRead
func resourceVcdNsxvFirewallRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvFirewallDelete
func resourceVcdNsxvFirewallDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvFirewallImport
func resourceVcdNsxvFirewallImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
