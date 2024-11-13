package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmRegionZone = "TM Region Zone"

func datasourceVcdTmRegionZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceVcdTmRegionZoneRead,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent Region ID for %s", labelTmRegionZone),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmRegionZone),
			},
		},
	}
}

func resourceVcdTmRegionZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	region, err := vcdClient.GetRegionById(d.Get("region_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelTmRegion, err)
	}

	getZone := func() func(name string) (*govcd.Zone, error) {
		return func(name string) (*govcd.Zone, error) {
			return region.GetZoneByName(name)
		}
	}()

	c := dsCrudConfig[*govcd.Zone, types.Zone]{
		entityLabel:    labelTmRegionZone,
		getEntityFunc:  getZone,
		stateStoreFunc: setZoneData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setZoneData(d *schema.ResourceData, z *govcd.Zone) error {
	if z == nil {
		return fmt.Errorf("nil Zone ")
	}
	d.SetId(z.Zone.ID)
	// IMPLEMENT
	return nil
}
