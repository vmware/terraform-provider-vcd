package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmContentLibraryItem = "Content Library Item"

func resourceVcdTmContentLibraryItem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmContentLibraryItemCreate,
		ReadContext:   resourceVcdTmContentLibraryItemRead,
		UpdateContext: resourceVcdTmContentLibraryItemUpdate,
		DeleteContext: resourceVcdTmContentLibraryItemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmContentLibraryItemImport,
		},

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

func resourceVcdTmContentLibraryItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	clId := d.Get("content_library_id").(string)
	cl, err := vcdClient.GetContentLibraryById(clId)
	if err != nil {
		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:      labelTmContentLibraryItem,
		getTypeFunc:      getContentLibraryItemType,
		stateStoreFunc:   setContentLibraryItemData,
		createFunc:       cl.CreateContentLibraryItem,
		resourceReadFunc: resourceVcdTmContentLibraryItemRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmContentLibraryItemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	clId := d.Get("content_library_id").(string)
	cl, err := vcdClient.GetContentLibraryById(clId)
	if err != nil {
		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:      labelTmContentLibraryItem,
		getTypeFunc:      getContentLibraryItemType,
		getEntityFunc:    cl.GetContentLibraryItemById,
		resourceReadFunc: resourceVcdTmContentLibraryItemRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmContentLibraryItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	clId := d.Get("content_library_id").(string)
	cl, err := vcdClient.GetContentLibraryById(clId)
	if err != nil {
		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:    labelTmContentLibraryItem,
		getEntityFunc:  cl.GetContentLibraryItemById,
		stateStoreFunc: setContentLibraryItemData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmContentLibraryItemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	clId := d.Get("content_library_id").(string)
	cl, err := vcdClient.GetContentLibraryById(clId)
	if err != nil {
		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:   labelTmContentLibraryItem,
		getEntityFunc: cl.GetContentLibraryItemById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmContentLibraryItemImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	id := strings.Split(d.Id(), ImportSeparator)
	if len(id) != 2 {
		return nil, fmt.Errorf("ID syntax should be \"Content Library name\".\"Content Library Item name\", where '.' is a customisable import separator")
	}

	cl, err := vcdClient.GetContentLibraryByName(id[0])
	if err != nil {
		return nil, fmt.Errorf("error getting Content Library with name '%s' for import: %s", id[0], err)
	}

	cli, err := cl.GetContentLibraryItemByName(id[1])
	if err != nil {
		return nil, fmt.Errorf("error getting Content Library Item with name '%s': %s", id[1], err)
	}

	d.SetId(cli.ContentLibraryItem.Id)
	dSet(d, "content_library_id", cl.ContentLibrary.Id)
	return []*schema.ResourceData{d}, nil
}

func getContentLibraryItemType(d *schema.ResourceData) (*types.ContentLibraryItem, error) {
	t := &types.ContentLibraryItem{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	return t, nil
}

func setContentLibraryItemData(d *schema.ResourceData, cli *govcd.ContentLibraryItem) error {
	if cli == nil || cli.ContentLibraryItem == nil {
		return fmt.Errorf("cannot save state for nil Content Library Item")
	}

	d.SetId(cli.ContentLibraryItem.Id)
	dSet(d, "name", cli.ContentLibraryItem.Name)

	return nil
}
