package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdOpenApiSecurityTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOpenApiSecurityTagCreate,
		ReadContext:   resourceVcdOpenApiSecurityTagRead,
		UpdateContext: resourceVcdOpenApiSecurityTagCreate,
		DeleteContext: resourceVcdOpenApiSecurityTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdOpenApiSecurityTagImport,
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Security tag name to be created",
			},
			"vm_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of VM IDs that the security tags is going to be tied to",
				MinItems:    1, // If vm_ids has nothing, the tag will be removed. We enforce to have at least 1 to avoid that behavior
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdOpenApiSecurityTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	securityTagName := d.Get("name").(string)

	securityTag := &types.SecurityTag{
		Tag:      securityTagName,
		Entities: convertSchemaSetToSliceOfStrings(d.Get("vm_ids").(*schema.Set)),
	}
	err := vcdClient.UpdateSecurityTag(securityTag)
	if err != nil {
		return diag.Errorf("error when setting up security tags - %s", err)
	}

	d.SetId(securityTagName)

	return resourceVcdOpenApiSecurityTagRead(ctx, d, meta)
}

func resourceVcdOpenApiSecurityTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	securityTagName := d.Id()
	taggedEntities, err := vcdClient.GetSecurityTaggedEntities(fmt.Sprintf("tag==%s", securityTagName))
	if err != nil {
		return diag.Errorf("error retrieving tagged entities - %s", err)
	}

	readEntities := make([]string, len(taggedEntities))
	for i, entity := range taggedEntities {
		readEntities[i] = entity.ID
	}

	d.Set("vm_ids", convertStringsToTypeSet(readEntities))
	return nil
}

func resourceVcdOpenApiSecurityTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	securityTagName := d.Id()

	securityTag := &types.SecurityTag{
		Tag:      securityTagName,
		Entities: []string{},
	}
	err := vcdClient.UpdateSecurityTag(securityTag)
	if err != nil {
		return diag.Errorf("error when deleting security tag - %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdOpenApiSecurityTagImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// TBD
	return nil, nil
}
