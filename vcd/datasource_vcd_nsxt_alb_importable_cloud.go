package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbImportableCloud() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbImportableCloudRead,

		Schema: map[string]*schema.Schema{
			"controller_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Controller ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Importable Cloud Name",
			},
			"already_imported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flags if the NSX-T ALB Importable Cloud is already imported",
			},
			"network_pool_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool name of NSX-T ALB Importable Cloud",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool ID NSX-T ALB Importable Cloud",
			},
			"transport_zone_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Transport zone name NSX-T ALB Importable Cloud",
			},
		},
	}
}

func datasourceVcdAlbImportableCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albImportableCloud, err := vcdClient.GetAlbImportableCloudByName(
		d.Get("controller_id").(string), d.Get("name").(string))

	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Importable Cloud by Name '%s': %s",
			d.Get("name").(string), err)
	}

	dSet(d, "already_imported", albImportableCloud.NsxtAlbImportableCloud.AlreadyImported)
	dSet(d, "network_pool_name", albImportableCloud.NsxtAlbImportableCloud.NetworkPoolRef.Name)
	dSet(d, "network_pool_id", albImportableCloud.NsxtAlbImportableCloud.NetworkPoolRef.ID)
	dSet(d, "transport_zone_name", albImportableCloud.NsxtAlbImportableCloud.TransportZoneName)
	d.SetId(albImportableCloud.NsxtAlbImportableCloud.ID)

	return nil
}
