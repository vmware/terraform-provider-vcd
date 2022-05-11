package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdSecurityTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdOpenApiSecurityTagCreateUpdate,
		ReadContext:   resourceVcdOpenApiSecurityTagRead,
		UpdateContext: resourceVcdOpenApiSecurityTagCreateUpdate,
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

func resourceVcdOpenApiSecurityTagCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	securityTagName := d.Get("name").(string)

	securityTag := &types.SecurityTag{
		Tag:      securityTagName,
		Entities: convertSchemaSetToSliceOfStrings(d.Get("vm_ids").(*schema.Set)),
	}
	_, err = org.UpdateSecurityTag(securityTag)
	if err != nil {
		return diag.Errorf("error when setting up security tags - %s", err)
	}

	d.SetId(securityTagName) // Security tags don't have a real ID. That's why we use the name as ID here.

	return resourceVcdOpenApiSecurityTagRead(ctx, d, meta)
}

func resourceVcdOpenApiSecurityTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	securityTagName := d.Id()

	taggedEntities, err := org.GetAllSecurityTaggedEntitiesByName(securityTagName)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("[DEBUG] Unable to find entities with security tag name: %s. Removing from tfstate", securityTagName)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving tagged entities - %s", err)
	}

	readEntities := make([]string, len(taggedEntities))
	for i, entity := range taggedEntities {
		readEntities[i] = entity.ID
	}

	err = d.Set("vm_ids", convertStringsToTypeSet(readEntities))
	if err != nil {
		return diag.Errorf("could not set vm_ids field: %s", err)
	}
	return nil
}

func resourceVcdOpenApiSecurityTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	securityTagName := d.Id()

	securityTag := &types.SecurityTag{
		Tag:      securityTagName,
		Entities: []string{},
	}
	_, err = org.UpdateSecurityTag(securityTag)
	if err != nil {
		return diag.Errorf("error when deleting security tag - %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdOpenApiSecurityTagImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog")
	}

	orgName, securityTag := resourceURI[0], resourceURI[1]
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrg, orgName)
	}

	taggedEntities, err := org.GetAllSecurityTaggedEntitiesByName(securityTag)
	if err != nil {
		return nil, err
	}

	readEntities := make([]string, len(taggedEntities))
	for i, entity := range taggedEntities {
		readEntities[i] = entity.ID
	}

	dSet(d, "org", orgName)
	dSet(d, "name", securityTag)
	err = d.Set("vm_ids", convertStringsToTypeSet(readEntities))
	if err != nil {
		return nil, fmt.Errorf("could not set vm_ids field: %s", err)
	}
	d.SetId(securityTag)

	return []*schema.ResourceData{d}, nil
}
