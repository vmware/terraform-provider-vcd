package vcd

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdOrg() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdOrgRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization name for lookup",
			},
			"full_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"is_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "True if this organization is enabled (allows login and all other operations).",
			},
			"deployed_vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. (0 = unlimited)",
			},
			"stored_vm_quota": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. (0 = unlimited)",
			},
			"can_publish_catalogs": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if this organization is allowed to share catalogs.",
			},
			"delay_after_power_on_seconds": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Specifies this organization's default for virtual machine boot delay after power on.",
			},
		},
	}
}

func datasourceVcdOrgRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	identifier := d.Get("name").(string)
	log.Printf("Reading Org with id %s", d.Get("name").(string))
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)

	if err != nil {
		log.Printf("Org with id %s not found. Setting ID to nothing", identifier)
		d.SetId("")
		return nil
	}
	log.Printf("Org with id %s found", identifier)
	d.SetId(adminOrg.AdminOrg.ID)
	return setOrgData(d, adminOrg)
}
