package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Role.",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role description",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this role is read-only",
			},
			"rights": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "list of rights assigned to this role",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	roleName := d.Get("name").(string)
	orgName := d.Get("org").(string)

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return diag.Errorf("[role create] error retrieving Org %s: %s", orgName, err)
	}

	// Check rights early, so that we can show a friendly error message when there are missing implied rights
	inputRights, err := getRights(vcdClient, org, "role create", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}

	role, err := org.CreateRole(&types.Role{
		Name:        roleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
	})
	if err != nil {
		return diag.Errorf("[role create] error creating role %s: %s", roleName, err)
	}
	if len(inputRights) > 0 {
		err = role.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[role create] error adding rights to role %s: %s", roleName, err)
		}
	}

	d.SetId(role.Role.ID)
	return genericRoleRead(ctx, d, meta, "resource", "create")
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericRoleRead(ctx, d, meta, "resource", "read")
}

func genericRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	roleName := d.Get("name").(string)
	orgName := d.Get("org").(string)
	identifier := d.Id()

	var role *govcd.Role
	var err error

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return diag.Errorf("[role %s-%s] error retrieving Org %s: %s", operation, origin, orgName, err)
	}

	if identifier == "" {
		role, err = org.GetRoleByName(roleName)
	} else {
		role, err = org.GetRoleById(identifier)
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("%s", err)
	}

	d.SetId(role.Role.ID)
	dSet(d, "description", role.Role.Description)
	dSet(d, "bundle_key", role.Role.BundleKey)
	dSet(d, "read_only", role.Role.ReadOnly)

	rights, err := role.GetRights(nil)
	if err != nil {
		return diag.Errorf("[role %s-%s] error while querying role rights: %s", operation, origin, err)

	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[role %s-%s] error setting rights for role %s: %s", operation, origin, roleName, err)
		}
	}
	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	roleName := d.Get("name").(string)
	orgName := d.Get("org").(string)

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return diag.Errorf("[role update] error retrieving Org %s: %s", orgName, err)
	}

	role, err := org.GetRoleById(d.Id())
	if err != nil {
		return diag.Errorf("[role update] error retrieving role %s: %s", roleName, err)
	}

	if d.HasChange("name") || d.HasChange("description") {
		role.Role.Name = roleName
		role.Role.Description = d.Get("description").(string)
		_, err = role.Update()
		if err != nil {
			return diag.Errorf("[role update] error updating role %s: %s", roleName, err)
		}
	}

	inputRights, err := getRights(vcdClient, org, "role update", d)
	if err != nil {
		return diag.Errorf("%s", err)
	}

	if len(inputRights) > 0 {
		err = role.UpdateRights(inputRights)
		if err != nil {
			return diag.Errorf("[role update] error updating role %s rights: %s", roleName, err)
		}
	} else {
		currentRights, err := role.GetRights(nil)
		if err != nil {
			return diag.Errorf("[role update] error retrieving role %s rights: %s", roleName, err)
		}
		if len(currentRights) > 0 {
			err = role.RemoveAllRights()
			if err != nil {
				return diag.Errorf("[role update] error removing role %s rights: %s", roleName, err)
			}
		}
	}
	return genericRoleRead(ctx, d, meta, "resource", "update")
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	roleName := d.Get("name").(string)
	orgName := d.Get("org").(string)

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return diag.Errorf("[role delete] error retrieving Org %s: %s", orgName, err)
	}
	var role *govcd.Role
	identifier := d.Id()
	if identifier != "" {
		role, err = org.GetRoleById(identifier)
	} else {
		role, err = org.GetRoleByName(roleName)
	}
	if err != nil {
		return diag.Errorf("[role delete] error retrieving role %s: %s", roleName, err)
	}
	err = role.Delete()
	if err != nil {
		return diag.Errorf("[role delete] error deleting role %s: %s", roleName, err)
	}
	return nil
}

// getRights will collect the list of rights of a rights collection (role, global role, rights bundle)
// and check whether the necessary implied rights are included.
// Calling resources should provide a client and optionally an Org (role)
// The "label" identifies the calling resource and operation and it is used to form error messages
func getRights(client *VCDClient, org *govcd.AdminOrg, label string, d *schema.ResourceData) ([]types.OpenApiReference, error) {
	var inputRights []types.OpenApiReference

	if client == nil {
		return nil, fmt.Errorf("[getRights - %s] client was empty", label)
	}
	rights := d.Get("rights").(*schema.Set).List()

	var right *types.Right
	var err error

	for _, r := range rights {
		rn := r.(string)
		if org != nil {
			right, err = org.GetRightByName(rn)
		} else {
			right, err = client.Client.GetRightByName(rn)
		}
		if err != nil {
			return nil, fmt.Errorf("[%s] error retrieving right %s: %s", label, rn, err)
		}
		inputRights = append(inputRights, types.OpenApiReference{Name: rn, ID: right.ID})
	}

	missingImpliedRights, err := govcd.FindMissingImpliedRights(&client.Client, inputRights)
	if err != nil {
		return nil, fmt.Errorf("[role create] error inspecting implied rights: %s", err)
	}

	if len(missingImpliedRights) > 0 {
		message := "The rights set for this role require the following implied rights to be added:"
		rightsList := ""
		for _, right := range missingImpliedRights {
			rightsList += fmt.Sprintf("\"%s\",\n", right.Name)
		}
		return nil, fmt.Errorf("%s\n%s", message, rightsList)
	}
	return inputRights, nil
}

func resourceVcdRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.role-name")
	}
	orgName, roleName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[role import] error retrieving org %s: %s", orgName, err)
	}

	role, err := org.GetRoleByName(roleName)
	if err != nil {
		return nil, fmt.Errorf("[role import] error retrieving role %s: %s", roleName, err)
	}
	dSet(d, "org", orgName)
	dSet(d, "name", roleName)
	dSet(d, "description", role.Role.Description)
	dSet(d, "bundle_key", role.Role.BundleKey)
	d.SetId(role.Role.ID)
	return []*schema.ResourceData{d}, nil
}
