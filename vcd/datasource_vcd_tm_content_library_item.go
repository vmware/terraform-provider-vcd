package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmContentLibraryItem() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceTmContentLibraryItemRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Content Library Item",
			},
			"content_library_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Content Library that this item belongs to",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO-8601 timestamp representing when this item was created",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Content Library Item",
			},
			"image_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual Machine Identifier (VMI) of the item. This is a ReadOnly field",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this item is published",
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this item is subscribed",
			},
			"last_successful_sync": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO-8601 timestamp representing when this item was last synced if subscribed",
			},
			"owner_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reference to the organization that the item belongs to",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of this Content Library Item",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of this item. For a subscribed library, this version is same as in publisher library",
			},
		},
	}
}

func datasourceTmContentLibraryItemRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	cl, err := vcdClient.GetContentLibraryById(d.Get("content_library_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Content Library: %s", err)
	}

	cli, err := cl.GetContentLibraryItemByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Content Library Item: %s", err)
	}

	setTmContentLibraryItemData(d, cli.ContentLibraryItem)

	return nil
}

func setTmContentLibraryItemData(d *schema.ResourceData, cli *types.ContentLibraryItem) {
	dSet(d, "content_library_id", cli.ContentLibrary.ID) // Cannot be nil
	dSet(d, "name", cli.Name)
	dSet(d, "description", cli.Description)
	dSet(d, "creation_date", cli.CreationDate)
	dSet(d, "image_identifier", cli.ImageIdentifier)
	dSet(d, "is_published", cli.IsPublished)
	dSet(d, "is_subscribed", cli.IsSubscribed)
	dSet(d, "last_successful_sync", cli.LastSuccessfulSync)
	if cli.Org != nil {
		dSet(d, "owner_org_id", cli.Org.ID)
	}
	dSet(d, "status", cli.Status)
	dSet(d, "version", cli.Version)
	d.SetId(cli.Id)
}
