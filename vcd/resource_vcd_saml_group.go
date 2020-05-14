package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceVcdOrgSamlGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdOrgSamlGroupCreate,
		Read:   resourceVcdOrgSamlGroupRead,
		Update: resourceVcdOrgSamlGroupUpdate,
		Delete: resourceVcdOrgSamlGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdOrgSamlGroupImport,
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

			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "SAML group name",
			},
			"role": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name to assign",
			},
		},
	}
}

/*
https://192.168.1.160/api/admin/org/c9e54447-4012-40fa-9f82-3039b3828a7e/groups

{"name":"group1","link":[],"vCloudExtension":[],"providerType":"SAML","role":{"vCloudExtension":[],"href":"https://192.168.1.160/api/admin/role/cbf95450-d6e0-3154-9537-2249481c012d","type":"application/vnd.vmware.admin.role+xml","link":[]}}
*/

// resourceVcdOrgSamlGroupCreate
func resourceVcdOrgSamlGroupCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdOrgSamlGroupUpdate
func resourceVcdOrgSamlGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdOrgSamlGroupRead
func resourceVcdOrgSamlGroupRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdOrgSamlGroupDelete
func resourceVcdOrgSamlGroupDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdOrgSamlGroupImport
func resourceVcdOrgSamlGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
