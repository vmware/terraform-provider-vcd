package vcd

import (
	"context"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdTmContentLibrary() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmContentLibraryRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Content Library",
			},
			"storage_policy_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of Region Storage Policy or VDC Storage Policy IDs used by this Content Library",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auto_attach": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "For Tenant Content Libraries this field represents whether this Content Library should be " +
					"automatically attached to all current and future namespaces in the tenant organization",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO-8601 timestamp representing when this Content Library was created",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Content Library",
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this Content Library is shared with other Organziations",
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this Content Library is subscribed from an external published library",
			},
			"library_type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The type of content library, can be either PROVIDER (Content Library that is scoped to a " +
					"provider) or TENANT (Content Library that is scoped to a tenant organization)",
			},
			"owner_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reference to the Organization that the Content Library belongs to",
			},
			"subscription_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A block representing subscription settings of a Content Library",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subscription url of this Content Library",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Password to use to authenticate with the publisher",
						},
						"need_local_copy": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to eagerly download content from publisher and store it locally",
						},
					},
				},
			},
			"version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Version number of this Content library",
			},
		},
	}
}

func datasourceVcdTmContentLibraryRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	cl, err := vcdClient.GetContentLibraryByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error getting Content Library: %s", err)
	}

	err = setTmContentLibraryData(d, cl.ContentLibrary)
	if err != nil {
		return diag.Errorf("error storing Content Library data: %s", err)
	}

	d.SetId(cl.ContentLibrary.ID)

	return nil
}

func setTmContentLibraryData(d *schema.ResourceData, cl *types.ContentLibrary) error {
	dSet(d, "name", cl.Name)
	dSet(d, "auto_attach", cl.AutoAttach)
	dSet(d, "creation_date", cl.CreationDate)
	dSet(d, "description", cl.Description)
	dSet(d, "is_shared", cl.IsShared)
	dSet(d, "is_subscribed", cl.IsSubscribed)
	dSet(d, "library_type", cl.LibraryType)
	dSet(d, "version_number", cl.VersionNumber)
	if cl.Org != nil {
		dSet(d, "owner_org_id", cl.Org.ID)
	}

	sps := make([]string, len(cl.StoragePolicies))
	for i, sp := range cl.StoragePolicies {
		sps[i] = sp.ID
	}
	err := d.Set("storage_policy_ids", sps)
	if err != nil {
		return err
	}

	if cl.SubscriptionConfig != nil {
		err = d.Set("subscription_config", []interface{}{
			map[string]interface{}{
				"subscription_url": cl.SubscriptionConfig.SubscriptionUrl,
				"password":         cl.SubscriptionConfig.Password,
				"need_local_copy":  cl.SubscriptionConfig.NeedLocalCopy,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
