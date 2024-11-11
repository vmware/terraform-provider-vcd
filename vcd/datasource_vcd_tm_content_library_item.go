package vcd

import (
	"context"
	"fmt"
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
				Description: fmt.Sprintf("Name of the %s", labelTmContentLibraryItem),
			},
			"content_library_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("ID of the Content Library that this %s belongs to", labelTmContentLibraryItem),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelTmContentLibraryItem),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The description of the %s", labelTmContentLibraryItem),
			},
			"image_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Virtual Machine Identifier (VMI) of the %s. This is a ReadOnly field", labelTmContentLibraryItem),
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is published", labelTmContentLibraryItem),
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is subscribed", labelTmContentLibraryItem),
			},
			"last_successful_sync": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was last synced if subscribed", labelTmContentLibraryItem),
			},
			"owner_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The reference to the organization that the %s belongs to", labelTmContentLibraryItem),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of this %s", labelTmContentLibraryItem),
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The version of this %s. For a subscribed library, this version is same as in publisher library", labelTmContentLibraryItem),
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
	d.SetId(cli.ID)
}
