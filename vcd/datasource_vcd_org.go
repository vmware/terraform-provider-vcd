package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os"
	"path"
)

func datasourceVcdOrg() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOrgRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization name for lookup",
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this organization is enabled (allows login and all other operations).",
			},
			"deployed_vm_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of virtual machines that can be deployed simultaneously by a member of this organization. (0 = unlimited)",
			},
			"stored_vm_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of virtual machines in vApps or vApp templates that can be stored in an undeployed state by a member of this organization. (0 = unlimited)",
			},
			"can_publish_catalogs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this organization is allowed to share catalogs.",
			},
			"can_publish_external_catalogs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this organization is allowed to publish external catalogs.",
			},
			"can_subscribe_external_catalogs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this organization is allowed to subscribe to external catalogs.",
			},
			"number_of_catalogs": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of catalogs, owned or shared, available to this organization",
			},
			"list_of_catalogs": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of catalogs, owned or shared, available to this organization",
			},
			"number_of_vdcs": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of VDCs, owned or shared, available to this organization",
			},
			"list_of_vdcs": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of VDCs, owned or shared, available to this organization",
			},
			"association_data_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the file to be filled with association data for this Org",
			},
			"vapp_lease": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_runtime_lease_in_sec": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "How long vApps can run before they are automatically stopped (in seconds)",
						},
						"power_off_on_runtime_lease_expiration": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "When true, vApps are powered off when the runtime lease expires. " +
								"When false, vApps are suspended when the runtime lease expires",
						},
						"maximum_storage_lease_in_sec": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "How long stopped vApps are available before being automatically cleaned up (in seconds)",
						},
						"delete_on_storage_lease_expiration": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "If true, storage for a vApp is deleted when the vApp's lease expires. " +
								"If false, the storage is flagged for deletion, but not deleted.",
						},
					},
				},
			},
			"vapp_template_lease": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_storage_lease_in_sec": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "How long vApp templates are available before being automatically cleaned up (in seconds)",
						},
						"delete_on_storage_lease_expiration": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "If true, storage for a vAppTemplate is deleted when the vAppTemplate lease expires. " +
								"If false, the storage is flagged for deletion, but not deleted",
						},
					},
				},
			},
			"delay_after_power_on_seconds": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Specifies this organization's default for virtual machine boot delay after power on.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for organization metadata",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("Organization"),
		},
	}
}

func datasourceVcdOrgRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	identifier := d.Get("name").(string)
	log.Printf("Reading Org with id %s", identifier)
	adminOrg, err := vcdClient.VCDClient.GetAdminOrgByNameOrId(identifier)

	if err != nil {
		log.Printf("Org with ID %s not found. Setting ID to nothing", identifier)
		d.SetId("")
		return diag.Errorf("org %s not found: %s", identifier, err)
	}
	log.Printf("Org with id %s found", identifier)
	d.SetId(adminOrg.AdminOrg.ID)

	diags := setOrgData(d, vcdClient, adminOrg)
	if diags != nil && diags.HasError() {
		return diags
	}
	associationDataFile := d.Get("association_data_file").(string)
	if associationDataFile != "" {

		associationRawData, err := adminOrg.GetOrgRawAssociationData()
		if err != nil {
			return diag.Errorf("error getting organization association data: %s", err)
		}
		err = os.WriteFile(path.Clean(associationDataFile), associationRawData, 0600)
		if err != nil {
			return diag.Errorf("error writing organization association data: %s", err)
		}
	}

	// This must be checked at the end as setOrgData can throw Warning diagnostics
	if len(diags) > 0 {
		return diags
	}
	return diags
}
