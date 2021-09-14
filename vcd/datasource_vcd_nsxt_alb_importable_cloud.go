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
			"controller_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Controller ID",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Importable Cloud Name",
			},
			"already_imported": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flags if the NSX-T ALB Importable Cloud is already imported",
			},
			"network_pool_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool name of NSX-T ALB Importable Cloud",
			},
			"network_pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool ID NSX-T ALB Importable Cloud",
			},
			"transport_zone_name": &schema.Schema{
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

	d.Set("already_imported", albImportableCloud.NsxtAlbImportableCloud.AlreadyImported)
	d.Set("network_pool_name", albImportableCloud.NsxtAlbImportableCloud.NetworkPoolRef.Name)
	d.Set("network_pool_id", albImportableCloud.NsxtAlbImportableCloud.NetworkPoolRef.ID)
	d.Set("transport_zone_name", albImportableCloud.NsxtAlbImportableCloud.TransportZoneName)
	d.SetId(albImportableCloud.NsxtAlbImportableCloud.ID)

	return nil
}
