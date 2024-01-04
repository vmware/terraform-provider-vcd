package vcd

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdCseKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCseKubernetesClusterCreate,
		ReadContext:   resourceVcdCseKubernetesRead,
		UpdateContext: resourceVcdCseKubernetesUpdate,
		DeleteContext: resourceVcdCseKubernetesDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Kubernetes cluster",
			},
			"capvcd_rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The CAPVCD RDE Type ID",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization that will own this Runtime Defined Entity, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VDC that hosts the Kubernetes cluster",
			},
			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the network that the Kubernetes cluster will use",
			},
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API token used to create and manage the cluster. The owner must have the 'Kubernetes Cluster Author' role",
			},
			"ssh_public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The SSH public key used to login into the cluster nodes",
			},
			"control_plane_machine_count": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The number of nodes that the control plane has. Must be an odd number and higher than 0",
				ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
					value, ok := v.(int)
					if !ok {
						return diag.Errorf("could not parse int value '%v' for control plane nodes", v)
					}
					if value < 1 || value%2 == 0 {
						return diag.Errorf("number of control plane nodes must be odd and higher than 0, but it was '%d'", value)
					}
					return nil
				},
			},
			"worker_machine_count": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The number of worker nodes, where the workloads are run",
				ValidateFunc: IsIntAndAtLeast(1),
			},
		},
	}
}

func resourceVcdCseKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	name := d.Get("name").(string)

	capvcdRdeTypeId := d.Get("capvcd_rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(capvcdRdeTypeId)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s', could not retrieve CAPVCD RDE Type with ID '%s': %s", name, capvcdRdeTypeId, err)
	}

	tenantContext := govcd.TenantContext{}
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s', error retrieving Org: %s", name, err)
	}
	tenantContext.OrgId = org.Org.ID
	tenantContext.OrgName = org.Org.Name

	err = validateCseKubernetesCluster(d)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s', error validating the payload: %s", name, err)
	}

	_, err = rdeType.CreateRde(types.DefinedEntity{
		EntityType: rdeType.DefinedEntityType.ID,
		Name:       name,
		Entity:     nil,
	}, &tenantContext)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s': %s", name, err)
	}

	return resourceVcdCseKubernetesRead(ctx, d, meta)
}

func resourceVcdCseKubernetesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rde, err := vcdClient.GetRdeById(d.Id())
	if err != nil {
		return diag.Errorf("could not read Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}
	status, ok := rde.DefinedEntity.Entity["status"].(map[string]interface{})
	if !ok {
		return diag.Errorf("could not read the status of the Kubernetes cluster with ID '%s'", d.Id())
	}
	dSet(d, "asd", status[""])

	return nil
}

func resourceVcdCseKubernetesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdCseKubernetesDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func validateCseKubernetesCluster(d *schema.ResourceData) error {
	return nil
}
