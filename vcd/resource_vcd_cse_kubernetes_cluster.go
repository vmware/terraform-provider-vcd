package vcd

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"text/template"
	"time"
)

//go:embed cse/4.2/capvcd.tmpl
var capvcdTemplate string

//go:embed cse/4.2/default_storage_class.tmpl
var defaultStorageClass string

func resourceVcdCseKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCseKubernetesClusterCreate,
		ReadContext:   resourceVcdCseKubernetesRead,
		UpdateContext: resourceVcdCseKubernetesUpdate,
		DeleteContext: resourceVcdCseKubernetesDelete,
		Schema: map[string]*schema.Schema{
			"runtime": {
				Type:         schema.TypeString,
				Default:      "tkg",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"tkg"}, false), // May add others in future releases of CSE
				Description:  "The Kubernetes runtime for the cluster. Only 'tkg' (Tanzu Kubernetes Grid) is supported",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Kubernetes cluster",
			},
			"ova_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the vApp Template that corresponds to a Kubernetes template OVA",
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
				Description: "The name of organization that will own this Kubernetes cluster, optional if defined at provider " +
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
			"control_plane": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"machine_count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3, // As suggested in UI
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
						"disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      20, // As suggested in UI
							ForceNew:     true,
							ValidateFunc: IsIntAndAtLeast(20),
							Description:  "Disk size for the control plane nodes",
						},
						"sizing_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "VM Sizing policy for the control plane nodes",
						},
						"placement_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "VM Placement policy for the control plane nodes",
						},
						"storage_profile": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Storage profile for the control plane nodes",
						},
						"ip": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "IP for the control plane",
						},
					},
				},
			},
			"node_pool": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"machine_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1, // As suggested in UI
							Description:  "The number of nodes that this node pool has. Must be higher than 0",
							ValidateFunc: IsIntAndAtLeast(1),
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     20, // As suggested in UI
							ForceNew:    true,
							Description: "Disk size for the control plane nodes",
						},
						"sizing_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "VM Sizing policy for the control plane nodes",
						},
						"placement_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "VM Placement policy for the control plane nodes",
						},
						"vgpu_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "vGPU policy for the control plane nodes",
						},
						"storage_profile": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Storage profile for the control plane nodes",
						},
					},
				},
			},
			"storage_class": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_profile_id": {
							Required:    true,
							Type:        schema.TypeString,
							Description: "ID of the storage profile to use for the storage class",
						},
						"name": {
							Required:    true,
							Type:        schema.TypeString,
							Description: "Name to give to this storage class",
						},
						"reclaim_policy": {
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"delete", "retain"}, false),
							Description:  "'delete' deletes the volume when the PersistentVolumeClaim is deleted. 'retain' does not, and the volume can be manually reclaimed",
						},
						"filesystem": {
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"ext4", "xfs"}, false),
							Description:  "Filesystem of the storage class, can be either 'ext4' or 'xfs'",
						},
					},
				},
			},
			"pods_cidr": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "100.96.0.0/11", // As suggested in UI
				Description: "CIDR that the Kubernetes pods will use",
			},
			"services_cidr": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "100.64.0.0/13", // As suggested in UI
				Description: "CIDR that the Kubernetes services will use",
			},
			"virtual_ip_subnet": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Virtual IP subnet for the cluster",
			},
			"auto_repair_on_errors": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "If errors occur before the Kubernetes cluster becomes available, and this argument is 'true', CSE Server will automatically attempt to repair the cluster",
			},
			"node_health_check": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "After the Kubernetes cluster becomes available, nodes that become unhealthy will be remediated according to unhealthy node conditions and remediation rules",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the cluster, can be 'provisioning', 'provisioned' or 'error'. Useful to check whether the Kubernetes cluster is in a stable status",
			},
			"raw_cluster_rde_json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The raw JSON that describes the cluster configuration inside the Runtime Defined Entity",
			},
		},
	}
}

func resourceVcdCseKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	clusterDetails, err := createClusterInfoDto(d, vcdClient)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s': %s", clusterDetails.Name, err)
	}

	entityMap, err := getCseKubernetesClusterEntityMap(d, clusterDetails)
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s': %s", clusterDetails.Name, err)
	}

	_, err = clusterDetails.RdeType.CreateRde(types.DefinedEntity{
		EntityType: clusterDetails.RdeType.DefinedEntityType.ID,
		Name:       clusterDetails.Name,
		Entity:     entityMap,
	}, &govcd.TenantContext{
		OrgId:   clusterDetails.Org.AdminOrg.ID,
		OrgName: clusterDetails.Org.AdminOrg.Name,
	})
	if err != nil {
		return diag.Errorf("could not create Kubernetes cluster with name '%s': %s", clusterDetails.Name, err)
	}

	return resourceVcdCseKubernetesRead(ctx, d, meta)
}

func getCseKubernetesClusterEntityMap(d *schema.ResourceData, clusterDetails *clusterInfoDto) (StringMap, error) {
	name := d.Get("name").(string)

	storageClass := "{}"
	if clusterDetails.StorageProfile != nil {
		storageClassEmpty := template.Must(template.New(name + "_StorageClass").Parse(defaultStorageClass))
		storageClassName := d.Get("storage_class.0.name").(string)
		reclaimPolicy := d.Get("storage_class.0.reclaim_policy").(string)
		filesystem := d.Get("storage_class.0.filesystem").(string)

		buf := &bytes.Buffer{}
		if err := storageClassEmpty.Execute(buf, map[string]string{
			"FileSystem":     filesystem,
			"Name":           storageClassName,
			"StorageProfile": clusterDetails.StorageProfile.Name,
			"ReclaimPolicy":  reclaimPolicy,
		}); err != nil {
			return nil, fmt.Errorf("could not generate a correct storage class JSON block: %s", err)
		}
		storageClass = buf.String()
	}

	capvcdEmpty := template.Must(template.New(name).Parse(capvcdTemplate))
	buf := &bytes.Buffer{}
	if err := capvcdEmpty.Execute(buf, map[string]string{
		"Name":                       name,
		"Org":                        clusterDetails.Org.AdminOrg.Name,
		"VcdUrl":                     clusterDetails.VcdUrl.String(),
		"Vdc":                        clusterDetails.Vdc.Vdc.Name,
		"Delete":                     "false",
		"ForceDelete":                "false",
		"AutoRepairOnErrors":         d.Get("auto_repair_on_errors").(string),
		"DefaultStorageClassOptions": storageClass,
		"ApiToken":                   d.Get("api_token").(string),
		"CapiYaml":                   getCapiYamlPlaintext(d, clusterDetails),
	}); err != nil {
		return nil, fmt.Errorf("could not generate a correct CAPVCD JSON: %s", err)
	}

	result := map[string]interface{}{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return nil, fmt.Errorf("could not generate a correct CAPVCD JSON: %s", err)
	}

	return result, nil
}

func getCapiYamlPlaintext(d *schema.ResourceData, clusterDetails *clusterInfoDto) string {
	return ""
}

func resourceVcdCseKubernetesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rde, err := vcdClient.GetRdeById(d.Id())
	if err != nil {
		return diag.Errorf("could not read Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}
	jsonEntity, err := jsonToCompactString(rde.DefinedEntity.Entity)
	if err != nil {
		return diag.Errorf("could not save the cluster '%s' raw RDE contents into state: %s", rde.DefinedEntity.ID, err)
	}
	dSet(d, "raw_cluster_rde_json", jsonEntity)

	status, ok := rde.DefinedEntity.Entity["status"].(StringMap)
	if !ok {
		return diag.Errorf("could not read the 'status' JSON object of the Kubernetes cluster with ID '%s'", d.Id())
	}

	vcdKe, ok := status["vcdKe"].(StringMap)
	if !ok {
		return diag.Errorf("could not read the 'status.vcdKe' JSON object of the Kubernetes cluster with ID '%s'", d.Id())
	}

	dSet(d, "state", vcdKe["state"])
	d.SetId(rde.DefinedEntity.ID)
	return nil
}

func resourceVcdCseKubernetesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceVcdCseKubernetesDelete deletes a CSE Kubernetes cluster. To delete a Kubernetes cluster, one must send
// the flags "markForDelete" and "forceDelete" back to true, so the CSE Server is able to delete all cluster elements
// and perform a cleanup. Hence, this function sends these properties and waits for deletion.
func resourceVcdCseKubernetesDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rde, err := vcdClient.GetRdeById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve the Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}

	spec, ok := rde.DefinedEntity.Entity["spec"].(StringMap)
	if !ok {
		return diag.Errorf("could not delete the cluster, JSON object 'spec' is not correct in the RDE")
	}

	spec["markForDelete"] = true
	spec["forceDelete"] = true
	rde.DefinedEntity.Entity["spec"] = spec

	err = rde.Update(*rde.DefinedEntity)
	if err != nil {
		return diag.Errorf("could not delete the cluster '%s': %s", rde.DefinedEntity.ID, err)
	}

	// TODO: Add a timeout
	deletionComplete := false
	for !deletionComplete {
		_, err = vcdClient.GetRdeById(d.Id())
		if err != nil {
			if govcd.IsNotFound(err) {
				deletionComplete = true
			}
			return diag.Errorf("could not check whether the cluster '%s' is deleted: %s", d.Id(), err)
		}
		time.Sleep(30 * time.Second)
	}
	return nil
}

// clusterInfoDto is a helper struct that contains all the required elements to successfully create and manage
// a Kubernetes cluster using CSE.
type clusterInfoDto struct {
	Name           string
	VcdUrl         url.URL
	Org            *govcd.AdminOrg
	Vdc            *govcd.Vdc
	VAppTemplate   *govcd.VAppTemplate
	Network        *govcd.OrgVDCNetwork
	RdeType        *govcd.DefinedEntityType
	StorageProfile *types.VdcStorageProfile
}

// createClusterInfoDto creates and returns a clusterInfoDto object by obtaining all the required information
// from th input Terraform resource data.
func createClusterInfoDto(d *schema.ResourceData, vcdClient *VCDClient) (*clusterInfoDto, error) {
	result := &clusterInfoDto{}

	name := d.Get("name").(string)
	result.Name = name

	org, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the cluster Organization: %s", err)
	}
	result.Org = org

	vdcId := d.Get("vdc_id").(string)
	vdc, err := org.GetVDCById(vdcId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the VDC with ID '%s': %s", vdcId, err)
	}
	result.Vdc = vdc

	vAppTemplateId := d.Get("ova_id").(string)
	vAppTemplate, err := vcdClient.GetVAppTemplateById(vAppTemplateId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Kubernetes OVA with ID '%s': %s", vAppTemplateId, err)
	}
	result.VAppTemplate = vAppTemplate

	networkId := d.Get("network_id").(string)
	network, err := vdc.GetOrgVdcNetworkById(networkId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Org VDC Network with ID '%s': %s", networkId, err)
	}
	result.Network = network

	rdeTypeId := d.Get("capvcd_rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve RDE Type with ID '%s': %s", rdeTypeId, err)
	}
	result.RdeType = rdeType

	var storageProfile *types.VdcStorageProfile
	if _, isStorageClassSet := d.GetOk("storage_class"); isStorageClassSet {
		storageProfileId := d.Get("storage_class.0.storage_profile_id").(string)
		storageProfile, err = vcdClient.GetStorageProfileById(storageProfileId)
		if err != nil {
			return nil, fmt.Errorf("could not get a Storage Profile with ID '%s': %s", storageProfileId, err)
		}
	}
	result.StorageProfile = storageProfile

	result.VcdUrl = vcdClient.VCDClient.Client.VCDHREF
	return result, nil
}
