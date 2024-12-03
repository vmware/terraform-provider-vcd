package vcd

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"sort"
	"time"
)

func resourceVcdCseKubernetesCluster() *schema.Resource {
	// This regular expression matches strings with at most 31 characters, composed only by lowercase alphanumeric characters or '-',
	// that must start with an alphabetic character, and end with an alphanumeric.
	// This is used for any "name" property in CSE, like cluster name, worker pool name or storage class name.
	kubernetesNameRegex := regexp.MustCompile(`^[a-z](?:[a-z0-9-]{0,29}[a-z0-9])?$`)

	return &schema.Resource{
		CreateContext: resourceVcdCseKubernetesClusterCreate,
		ReadContext:   resourceVcdCseKubernetesRead,
		UpdateContext: resourceVcdCseKubernetesUpdate,
		DeleteContext: resourceVcdCseKubernetesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCseKubernetesImport,
		},
		Schema: map[string]*schema.Schema{
			"cse_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"4.1.0", "4.1.1", "4.2.0", "4.2.1", "4.2.2", "4.2.3"}, false),
				Description:  "The CSE version to use",
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// This custom diff function allows to correctly compare versions.
					oldVersion, err := semver.NewVersion(oldValue)
					if err != nil {
						return false
					}
					newVersion, err := semver.NewVersion(newValue)
					if err != nil {
						return false
					}
					return oldVersion.Equal(newVersion)
				},
				DiffSuppressOnRefresh: true,
			},
			"runtime": {
				Type:         schema.TypeString,
				Optional:     true,
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
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(kubernetesNameRegex, "name must contain only lowercase alphanumeric characters or '-',"+
					"start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters")),
			},
			"kubernetes_template_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the vApp Template that corresponds to a Kubernetes template OVA",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true, // Gets the Provider org if not set
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
			"owner": {
				Type:        schema.TypeString,
				Optional:    true, // Gets the Provider user if not set
				ForceNew:    true,
				Description: "The user that creates the cluster and owns the API token specified in 'api_token'. It must have the 'Kubernetes Cluster Author' role. If not specified, it assumes it's the user from the provider configuration",
			},
			"api_token_file": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false, // It's only used on creation, so we do not care about updates
				Description: "A file generated by 'vcd_api_token' resource, that stores the API token used to create and manage the cluster, owned by the user specified in 'owner'. Be careful about this file, as it contains sensitive information",
			},
			"ssh_public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The SSH public key used to login into the cluster nodes",
			},
			"control_plane": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Defines the control plane for the cluster",
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
						"disk_size_gi": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          20, // As suggested in UI
							ForceNew:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(20)),
							Description:      "Disk size, in Gibibytes (Gi), for the control plane nodes. Must be at least 20",
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
						"storage_profile_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Storage profile for the control plane nodes",
						},
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true, // IP can be auto-assigned if left-empty
							ForceNew:     true,
							Description:  "IP for the control plane. It will be automatically assigned during cluster creation if left empty",
							ValidateFunc: checkEmptyOrSingleIP(),
						},
					},
				},
			},
			"worker_pool": {
				// This is a list because TypeSet tries to replace the whole block when we just change a sub-attribute like "machine_count",
				// that would cause the worker pool to be deleted and then re-created, which is not allowed in CSE.
				// On the other hand, with TypeList the updates on sub-attributes work as expected but in exchange
				// we need to be careful on reads to guarantee that order is respected.
				Type:        schema.TypeList,
				Required:    true,
				Description: "Defines a worker pool for the cluster",
				Elem: &schema.Resource{
					// Ideally, all of these sub-attributes should have ForceNew: true except for "machine_count", as
					// they can't be changed. However, this doesn't work well, so we check this at runtime.
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of this worker pool. Must be unique",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(kubernetesNameRegex, "name must contain only lowercase alphanumeric characters or '-',"+
								"start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters")),
						},
						"machine_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          1, // As suggested in UI
							Description:      "The number of nodes that this worker pool has. Must be higher than or equal to 0. Ignored if 'autoscaler_max_replicas' and 'autoscaler_min_replicas' are set",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
						},
						"disk_size_gi": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          20, // As suggested in UI
							Description:      "Disk size, in Gibibytes (Gi), for this worker pool",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(20)),
						},
						"sizing_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "VM Sizing policy for this worker pool",
						},
						"placement_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "VM Placement policy for this worker pool",
						},
						"vgpu_policy_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "vGPU policy for this worker pool",
						},
						"storage_profile_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Storage profile for this worker pool",
						},
						"autoscaler_max_replicas": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum replicas for the autoscaling capabilities of this worker pool. Requires 'autoscaler_min_replicas'",
						},
						"autoscaler_min_replicas": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Minimum replicas for the autoscaling capabilities of this worker pool. Requires 'autoscaler_max_replicas'",
						},
					},
				},
			},
			"default_storage_class": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Defines the default storage class for the cluster",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_profile_id": {
							Required:    true,
							ForceNew:    true,
							Type:        schema.TypeString,
							Description: "ID of the storage profile to use for the storage class",
						},
						"name": {
							Required:    true,
							ForceNew:    true,
							Type:        schema.TypeString,
							Description: "Name to give to this storage class",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringMatch(kubernetesNameRegex, "name must contain only lowercase alphanumeric characters or '-',"+
								"start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters")),
						},
						"reclaim_policy": {
							Required:     true,
							ForceNew:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"delete", "retain"}, false),
							Description:  "Reclaim policy. Possible values are: `delete` deletes the volume when the `PersistentVolumeClaim` is deleted; `retain` does not delete, and the volume can be manually reclaimed",
						},
						"filesystem": {
							Required:     true,
							ForceNew:     true,
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
				Optional:    true,
				Description: "Virtual IP subnet for the cluster",
			},
			"auto_repair_on_errors": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true, // CSE Server turns this off after the cluster is successfully provisioned
				Description: "If errors occur before the Kubernetes cluster becomes available, and this argument is 'true', CSE Server will automatically attempt to repair the cluster",
			},
			"node_health_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "After the Kubernetes cluster becomes available, nodes that become unhealthy will be remediated according to unhealthy node conditions and remediation rules",
			},
			"operations_timeout_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
				Description: "The time, in minutes, to wait for the cluster operations to be successfully completed. For example, during cluster creation, it should be in `provisioned`" +
					"state before the timeout is reached, otherwise the operation will return an error. For cluster deletion, this timeout" +
					"specifies the time to wait until the cluster is completely deleted. Setting this argument to `0` means to wait indefinitely",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of Kubernetes installed in this cluster",
			},
			"tkg_product_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of TKG installed in this cluster",
			},
			"capvcd_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of CAPVCD used by this cluster",
			},
			"cluster_resource_set_bindings": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The cluster resource set bindings of this cluster",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpi_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the Cloud Provider Interface used by this cluster",
			},
			"csi_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the Container Storage Interface used by this cluster",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the cluster, can be 'provisioning', 'provisioned', 'deleting' or 'error'. Useful to check whether the Kubernetes cluster is in a stable status",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contents of the kubeconfig of the Kubernetes cluster, only available when 'state=provisioned'",
				Sensitive:   true,
			},
			"supported_upgrades": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of vApp Template names that can be used to upgrade the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"events": {
				Type:        schema.TypeList, // Order matters here, as they're ordered by date
				Computed:    true,
				Description: "A list of events that happened during the Kubernetes cluster lifecycle",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of the event",
						},
						"resource_id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "ID of the resource that caused the event",
						},
						"type": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Type of the event, either 'event' or 'error'",
						},
						"occurred_at": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "When the event happened",
						},
						"details": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Details of the event",
						},
					},
				},
			},
		},
	}
}

func resourceVcdCseKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cseVersion, err := semver.NewSemver(d.Get("cse_version").(string))
	if err != nil {
		return diag.Errorf("the introduced 'cse_version=%s' is not valid: %s", d.Get("cse_version"), err)
	}

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("could not create a Kubernetes cluster in the target Organization: %s", err)
	}

	apiTokenFile := d.Get("api_token_file").(string)
	if apiTokenFile == "" {
		return diag.Errorf("the API token file 'is required during Kubernetes cluster creation")
	}
	apiToken, err := govcd.GetTokenFromFile(apiTokenFile)
	if err != nil {
		return diag.Errorf("could not read the API token from the file '%s': %s", apiTokenFile, err)
	}

	owner := d.Get("owner").(string)
	if owner == "" {
		session, err := vcdClient.Client.GetSessionInfo()
		if err != nil {
			return diag.Errorf("could not get an Owner for the Kubernetes cluster. 'owner' is not set and cannot get one from the Provider configuration: %s", err)
		}
		owner = session.User.Name
		if owner == "" {
			return diag.Errorf("could not get an Owner for the Kubernetes cluster. 'owner' is not set and cannot get one from the Provider configuration")
		}
	}

	creationData := govcd.CseClusterSettings{
		CseVersion:              *cseVersion,
		Name:                    d.Get("name").(string),
		OrganizationId:          org.Org.ID,
		VdcId:                   d.Get("vdc_id").(string),
		NetworkId:               d.Get("network_id").(string),
		KubernetesTemplateOvaId: d.Get("kubernetes_template_id").(string),
		ControlPlane: govcd.CseControlPlaneSettings{
			MachineCount:      d.Get("control_plane.0.machine_count").(int),
			DiskSizeGi:        d.Get("control_plane.0.disk_size_gi").(int),
			SizingPolicyId:    d.Get("control_plane.0.sizing_policy_id").(string),
			PlacementPolicyId: d.Get("control_plane.0.placement_policy_id").(string),
			StorageProfileId:  d.Get("control_plane.0.storage_profile_id").(string),
			Ip:                d.Get("control_plane.0.ip").(string),
		},
		Owner:              owner,
		ApiToken:           apiToken.RefreshToken,
		NodeHealthCheck:    d.Get("node_health_check").(bool),
		PodCidr:            d.Get("pods_cidr").(string),
		ServiceCidr:        d.Get("services_cidr").(string),
		SshPublicKey:       d.Get("ssh_public_key").(string),
		VirtualIpSubnet:    d.Get("virtual_ip_subnet").(string),
		AutoRepairOnErrors: d.Get("auto_repair_on_errors").(bool),
	}

	workerPoolsAttr := d.Get("worker_pool").([]interface{})
	workerPools := make([]govcd.CseWorkerPoolSettings, len(workerPoolsAttr))
	for i, w := range workerPoolsAttr {
		workerPool := w.(map[string]interface{})
		workerPools[i] = govcd.CseWorkerPoolSettings{
			Name:              workerPool["name"].(string),
			DiskSizeGi:        workerPool["disk_size_gi"].(int),
			SizingPolicyId:    workerPool["sizing_policy_id"].(string),
			PlacementPolicyId: workerPool["placement_policy_id"].(string),
			VGpuPolicyId:      workerPool["vgpu_policy_id"].(string),
			StorageProfileId:  workerPool["storage_profile_id"].(string),
		}
		autoscalerMaxReplicas := workerPool["autoscaler_max_replicas"].(int)
		autoscalerMinReplicas := workerPool["autoscaler_min_replicas"].(int)

		if autoscalerMaxReplicas > 0 && autoscalerMinReplicas <= 0 {
			return diag.Errorf("Worker Pool '%s' 'autoscaler_max_replicas=%d' requires 'autoscaler_min_replicas=%d' to be higher than 0", workerPools[i].Name, autoscalerMaxReplicas, autoscalerMinReplicas)
		}
		if autoscalerMinReplicas > 0 && autoscalerMaxReplicas <= 0 {
			return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' requires 'autoscaler_max_replicas=%d' to be higher than 0", workerPools[i].Name, autoscalerMinReplicas, autoscalerMaxReplicas)
		}
		if autoscalerMinReplicas > autoscalerMaxReplicas {
			return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' should not be higher than 'autoscaler_max_replicas=%d'", workerPools[i].Name, autoscalerMinReplicas, autoscalerMaxReplicas)
		}
		if autoscalerMaxReplicas > 0 && autoscalerMinReplicas > 0 {
			if workerPool["machine_count"].(int) != 0 {
				return diag.Errorf("Worker Pool '%s' 'machine_count=%d' should be set to 0 when 'autoscaler_min_replicas=%d'/'autoscaler_max_replicas=%d'", workerPools[i].Name, workerPool["machine_count"], autoscalerMinReplicas, autoscalerMaxReplicas)
			}
			workerPools[i].Autoscaler = &govcd.CseWorkerPoolAutoscaler{
				MaxSize: autoscalerMaxReplicas,
				MinSize: autoscalerMinReplicas,
			}
		} else {
			workerPools[i].MachineCount = workerPool["machine_count"].(int)
		}
	}
	creationData.WorkerPools = workerPools

	if _, ok := d.GetOk("default_storage_class"); ok {
		creationData.DefaultStorageClass = &govcd.CseDefaultStorageClassSettings{
			StorageProfileId: d.Get("default_storage_class.0.storage_profile_id").(string),
			Name:             d.Get("default_storage_class.0.name").(string),
			ReclaimPolicy:    d.Get("default_storage_class.0.reclaim_policy").(string),
			Filesystem:       d.Get("default_storage_class.0.filesystem").(string),
		}
	}

	cluster, err := org.CseCreateKubernetesCluster(creationData, time.Duration(d.Get("operations_timeout_minutes").(int))*time.Minute)
	if err != nil && cluster == nil {
		return diag.Errorf("Kubernetes cluster creation failed: %s", err)
	}

	// If we get here, it means we got either a successful created cluster, a timeout or a cluster in "error" state.
	// Either way, from this point we should go to the Update logic as the cluster is definitely present in VCD, so we store the ID.
	// Also, we need to set the ID to be able to distinguish this cluster from all the others that may have the same name and RDE Type.
	// We could use some other ways of filtering, but ID is the only accurate one.
	// If the cluster can't be created due to errors, users should delete it and retry, like in UI.
	d.SetId(cluster.ID)

	if cluster.State != "provisioned" {
		return diag.Errorf("Kubernetes cluster creation finished, but it is not in 'provisioned' state: %s", err)
	}

	return resourceVcdCseKubernetesRead(ctx, d, meta)
}

func resourceVcdCseKubernetesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vcdClient := meta.(*VCDClient)
	// The ID must be already set for the read to be successful. We can't rely on the name as there can be
	// many clusters with the same name in the same org.
	cluster, err := vcdClient.CseGetKubernetesClusterById(d.Id())
	if err != nil {
		return diag.Errorf("could not read Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}

	warns, err := saveClusterDataToState(d, vcdClient, cluster, "resource")
	if err != nil {
		return diag.Errorf("could not save Kubernetes cluster data into Terraform state: %s", err)
	}
	for _, warning := range warns {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  warning.Error(),
		})
	}

	if len(diags) > 0 {
		return diags
	}
	return nil
}

// resourceVcdCseKubernetesUpdate updates the Kubernetes clusters. Note that re-creating the CAPI YAML and sending it
// back will break everything, so we must patch the YAML piece by piece.
func resourceVcdCseKubernetesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Some arguments don't require changes in the backend
	if !d.HasChangesExcept("operations_timeout_minutes") {
		return nil
	}

	vcdClient := meta.(*VCDClient)
	cluster, err := vcdClient.CseGetKubernetesClusterById(d.Id())
	if err != nil {
		return diag.Errorf("could not get Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}
	payload := govcd.CseClusterUpdateInput{}
	if d.HasChange("worker_pool") {
		oldPools, newPools := d.GetChange("worker_pool")
		existingPools := map[string]bool{}

		// Fetch the already existing worker pools that have been modified
		changePoolsPayload := map[string]govcd.CseWorkerPoolUpdateInput{}
		for _, o := range oldPools.([]interface{}) {
			oldPool := o.(map[string]interface{})
			for _, n := range newPools.([]interface{}) {
				newPool := n.(map[string]interface{})
				if oldPool["name"].(string) == newPool["name"].(string) {
					if oldPool["disk_size_gi"] != newPool["disk_size_gi"] {
						return diag.Errorf("'disk_size_gi' of Worker Pool '%s' cannot be changed", oldPool["name"])
					}
					if oldPool["sizing_policy_id"] != newPool["sizing_policy_id"] {
						return diag.Errorf("'sizing_policy_id' of Worker Pool '%s' cannot be changed", oldPool["name"])
					}
					if oldPool["placement_policy_id"] != newPool["placement_policy_id"] {
						return diag.Errorf("'placement_policy_id' of Worker Pool '%s' cannot be changed", oldPool["name"])
					}
					if oldPool["vgpu_policy_id"] != newPool["vgpu_policy_id"] {
						return diag.Errorf("'vgpu_policy_id' of Worker Pool '%s' cannot be changed", oldPool["name"])
					}
					if oldPool["storage_profile_id"] != newPool["storage_profile_id"] {
						return diag.Errorf("'storage_profile_id' of Worker Pool '%s' cannot be changed", oldPool["name"])
					}
					autoscalerMaxReplicas := newPool["autoscaler_max_replicas"].(int)
					autoscalerMinReplicas := newPool["autoscaler_min_replicas"].(int)

					if autoscalerMaxReplicas > 0 && autoscalerMinReplicas <= 0 {
						return diag.Errorf("Worker Pool '%s' 'autoscaler_max_replicas=%d' requires 'autoscaler_min_replicas=%d' to be higher than 0", newPool["name"], autoscalerMaxReplicas, autoscalerMinReplicas)
					}
					if autoscalerMinReplicas > 0 && autoscalerMaxReplicas <= 0 {
						return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' requires 'autoscaler_max_replicas=%d' to be higher than 0", newPool["name"], autoscalerMinReplicas, autoscalerMaxReplicas)
					}
					if autoscalerMinReplicas > autoscalerMaxReplicas {
						return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' should not be higher than 'autoscaler_max_replicas=%d'", newPool["name"], autoscalerMinReplicas, autoscalerMaxReplicas)
					}
					wpUpdateInput := govcd.CseWorkerPoolUpdateInput{}
					if autoscalerMaxReplicas > 0 && autoscalerMinReplicas > 0 {
						if newPool["machine_count"].(int) != 0 {
							return diag.Errorf("Worker Pool '%s' 'machine_count=%d' should be set to 0 when 'autoscaler_min_replicas=%d'/'autoscaler_max_replicas=%d'", newPool["name"], newPool["machine_count"], autoscalerMinReplicas, autoscalerMaxReplicas)
						}
						wpUpdateInput.Autoscaler = &govcd.CseWorkerPoolAutoscaler{
							MaxSize: autoscalerMaxReplicas,
							MinSize: autoscalerMinReplicas,
						}
					} else {
						wpUpdateInput.MachineCount = newPool["machine_count"].(int)
					}
					changePoolsPayload[newPool["name"].(string)] = wpUpdateInput
					existingPools[newPool["name"].(string)] = true // Register this pool as not new
				}
			}
		}
		payload.WorkerPools = &changePoolsPayload

		// Check that no Worker Pools are deleted
		for _, o := range oldPools.([]interface{}) {
			oldPool := o.(map[string]interface{})
			if _, ok := existingPools[oldPool["name"].(string)]; !ok {
				return diag.Errorf("the Worker Pool '%s' can't be deleted, but you can scale it to 0", oldPool["name"].(string))
			}
		}

		// Fetch the worker pools that are brand new
		var addPoolsPayload []govcd.CseWorkerPoolSettings
		for _, n := range newPools.([]interface{}) {
			newPool := n.(map[string]interface{})
			if _, ok := existingPools[newPool["name"].(string)]; !ok {
				wp := govcd.CseWorkerPoolSettings{
					Name:              newPool["name"].(string),
					DiskSizeGi:        newPool["disk_size_gi"].(int),
					SizingPolicyId:    newPool["sizing_policy_id"].(string),
					PlacementPolicyId: newPool["placement_policy_id"].(string),
					VGpuPolicyId:      newPool["vgpu_policy_id"].(string),
					StorageProfileId:  newPool["storage_profile_id"].(string),
				}
				autoscalerMaxReplicas := newPool["autoscaler_max_replicas"].(int)
				autoscalerMinReplicas := newPool["autoscaler_min_replicas"].(int)

				if autoscalerMaxReplicas > 0 && autoscalerMinReplicas <= 0 {
					return diag.Errorf("Worker Pool '%s' 'autoscaler_max_replicas=%d' requires 'autoscaler_min_replicas=%d' to be higher than 0", wp.Name, autoscalerMaxReplicas, autoscalerMinReplicas)
				}
				if autoscalerMinReplicas > 0 && autoscalerMaxReplicas <= 0 {
					return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' requires 'autoscaler_max_replicas=%d' to be higher than 0", wp.Name, autoscalerMinReplicas, autoscalerMaxReplicas)
				}
				if autoscalerMinReplicas > autoscalerMaxReplicas {
					return diag.Errorf("Worker Pool '%s' 'autoscaler_min_replicas=%d' should not be higher than 'autoscaler_max_replicas=%d'", wp.Name, autoscalerMinReplicas, autoscalerMaxReplicas)
				}
				if autoscalerMaxReplicas > 0 && autoscalerMinReplicas > 0 {
					if newPool["machine_count"].(int) != 0 {
						return diag.Errorf("Worker Pool '%s' 'machine_count=%d' should be set to 0 when 'autoscaler_min_replicas=%d'/'autoscaler_max_replicas=%d'", wp.Name, newPool["machine_count"].(int), autoscalerMinReplicas, autoscalerMaxReplicas)
					}
					wp.Autoscaler = &govcd.CseWorkerPoolAutoscaler{
						MaxSize: autoscalerMaxReplicas,
						MinSize: autoscalerMinReplicas,
					}
				} else {
					wp.MachineCount = newPool["machine_count"].(int)
				}
				addPoolsPayload = append(addPoolsPayload, wp)
			}
		}
		payload.NewWorkerPools = &addPoolsPayload
	}
	if d.HasChange("control_plane") {
		controlPlane := govcd.CseControlPlaneUpdateInput{}
		for _, controlPlaneAttr := range d.Get("control_plane").([]interface{}) {
			c := controlPlaneAttr.(map[string]interface{})
			controlPlane.MachineCount = c["machine_count"].(int)
		}
		payload.ControlPlane = &controlPlane
	}
	if d.HasChange("kubernetes_template_id") {
		payload.KubernetesTemplateOvaId = addrOf(d.Get("kubernetes_template_id").(string))
	}

	if d.HasChange("node_health_check") {
		payload.NodeHealthCheck = addrOf(d.Get("node_health_check").(bool))
	}

	if d.HasChanges("auto_repair_on_errors") {
		payload.AutoRepairOnErrors = addrOf(d.Get("auto_repair_on_errors").(bool))
	}

	err = cluster.Update(payload, true)
	if err != nil {
		return diag.Errorf("Kubernetes cluster update failed: %s", err)
	}

	return resourceVcdCseKubernetesRead(ctx, d, meta)
}

// resourceVcdCseKubernetesDelete deletes a CSE Kubernetes cluster. To delete a Kubernetes cluster, one must send
// the flags "markForDelete" and "forceDelete" back to true, so the CSE Server is able to delete all cluster elements
// and perform a cleanup. Hence, this function sends an update of just these two properties and waits for the cluster RDE
// to be gone.
func resourceVcdCseKubernetesDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	cluster, err := vcdClient.CseGetKubernetesClusterById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return nil // The cluster is gone, nothing to do
		}
		return diag.FromErr(err)
	}
	err = cluster.Delete(time.Duration(d.Get("operations_timeout_minutes").(int)) * time.Minute)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdCseKubernetesImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	cluster, err := vcdClient.CseGetKubernetesClusterById(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving Kubernetes cluster with ID '%s': %s", d.Id(), err)
	}

	warns, err := saveClusterDataToState(d, vcdClient, cluster, "import")
	if err != nil {
		return nil, fmt.Errorf("failed importing Kubernetes cluster '%s': %s", cluster.ID, err)
	}
	for _, warn := range warns {
		// We can't do much here as Import does not support Diagnostics
		logForScreen(cluster.ID, fmt.Sprintf("got a warning during import: %s", warn))
	}

	return []*schema.ResourceData{d}, nil
}

// saveClusterDataToState reads the received RDE contents and sets the Terraform arguments and attributes.
// Returns a slice of warnings first and an error second.
func saveClusterDataToState(d *schema.ResourceData, vcdClient *VCDClient, cluster *govcd.CseKubernetesCluster, origin string) ([]error, error) {
	var warnings []error

	dSet(d, "name", cluster.Name)
	dSet(d, "cse_version", cluster.CseVersion.Original())
	dSet(d, "runtime", "tkg") // Only one supported
	dSet(d, "vdc_id", cluster.VdcId)
	dSet(d, "network_id", cluster.NetworkId)
	dSet(d, "cpi_version", cluster.CpiVersion.Original())
	dSet(d, "csi_version", cluster.CsiVersion.Original())
	dSet(d, "capvcd_version", cluster.CapvcdVersion.Original())
	dSet(d, "kubernetes_version", cluster.KubernetesVersion.Original())
	dSet(d, "tkg_product_version", cluster.TkgVersion.Original())
	dSet(d, "pods_cidr", cluster.PodCidr)
	dSet(d, "services_cidr", cluster.ServiceCidr)
	dSet(d, "kubernetes_template_id", cluster.KubernetesTemplateOvaId)
	dSet(d, "ssh_public_key", cluster.SshPublicKey)
	dSet(d, "virtual_ip_subnet", cluster.VirtualIpSubnet)
	dSet(d, "auto_repair_on_errors", cluster.AutoRepairOnErrors)
	dSet(d, "node_health_check", cluster.NodeHealthCheck)

	// The data source does not have the attribute "org", so we cannot set it
	if origin != "datasource" {
		// If the Org was set, it needs to be refreshed (it should not change, though)
		// We also set it always during imports.
		if _, ok := d.GetOk("org"); ok || origin == "import" {
			if cluster.OrganizationId != "" {
				org, err := vcdClient.GetOrgById(cluster.OrganizationId)
				if err != nil {
					return nil, fmt.Errorf("could not set 'org' argument: %s", err)
				}
				dSet(d, "org", org.Org.Name)
			}
		}
	}

	// If the Owner was set, it needs to be refreshed (it should not change, though).
	// If the origin is a data source or import, we always need to set this one.
	if _, ok := d.GetOk("owner"); ok || origin == "datasource" || origin == "import" {
		dSet(d, "owner", cluster.Owner)
	}

	err := d.Set("cluster_resource_set_bindings", cluster.ClusterResourceSetBindings)
	if err != nil {
		return nil, err
	}

	workerPoolBlocks := make([]map[string]interface{}, len(cluster.WorkerPools))
	for i, workerPool := range cluster.WorkerPools {
		workerPoolBlocks[i] = map[string]interface{}{
			"machine_count":       workerPool.MachineCount,
			"name":                workerPool.Name,
			"vgpu_policy_id":      workerPool.VGpuPolicyId,
			"sizing_policy_id":    workerPool.SizingPolicyId,
			"placement_policy_id": workerPool.PlacementPolicyId,
			"storage_profile_id":  workerPool.StorageProfileId,
			"disk_size_gi":        workerPool.DiskSizeGi,
		}
		if workerPool.Autoscaler != nil {
			workerPoolBlocks[i]["autoscaler_max_replicas"] = workerPool.Autoscaler.MaxSize
			workerPoolBlocks[i]["autoscaler_min_replicas"] = workerPool.Autoscaler.MinSize
		}
	}
	// The "worker_pool" argument is a TypeList, not a TypeSet (check the Schema comments for context),
	// so we need to guarantee order. We order them by name, which is unique.
	sort.SliceStable(workerPoolBlocks, func(i, j int) bool {
		return workerPoolBlocks[i]["name"].(string) < workerPoolBlocks[j]["name"].(string)
	})

	err = d.Set("worker_pool", workerPoolBlocks)
	if err != nil {
		return nil, err
	}

	err = d.Set("control_plane", []map[string]interface{}{
		{
			"machine_count":       cluster.ControlPlane.MachineCount,
			"ip":                  cluster.ControlPlane.Ip,
			"sizing_policy_id":    cluster.ControlPlane.SizingPolicyId,
			"placement_policy_id": cluster.ControlPlane.PlacementPolicyId,
			"storage_profile_id":  cluster.ControlPlane.StorageProfileId,
			"disk_size_gi":        cluster.ControlPlane.DiskSizeGi,
		},
	})
	if err != nil {
		return nil, err
	}

	if cluster.DefaultStorageClass != nil {
		err = d.Set("default_storage_class", []map[string]interface{}{{
			"storage_profile_id": cluster.DefaultStorageClass.StorageProfileId,
			"name":               cluster.DefaultStorageClass.Name,
			"reclaim_policy":     cluster.DefaultStorageClass.ReclaimPolicy,
			"filesystem":         cluster.DefaultStorageClass.Filesystem,
		}})
		if err != nil {
			return nil, err
		}
	}

	dSet(d, "state", cluster.State)

	supportedUpgrades, err := cluster.GetSupportedUpgrades(true)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the supported upgrades for the Kubernetes cluster with ID '%s': %s", cluster.ID, err)
	}
	supportedUpgradesNames := make([]string, len(supportedUpgrades))
	for i, upgrade := range supportedUpgrades {
		supportedUpgradesNames[i] = upgrade.Name
	}
	err = d.Set("supported_upgrades", supportedUpgradesNames)
	if err != nil {
		return nil, err
	}

	events := make([]interface{}, len(cluster.Events))
	for i, event := range cluster.Events {
		events[i] = map[string]interface{}{
			"resource_id": event.ResourceId,
			"name":        event.Name,
			"occurred_at": event.OccurredAt.String(),
			"details":     event.Details,
			"type":        event.Type,
		}
	}
	err = d.Set("events", events)
	if err != nil {
		return nil, err
	}

	if cluster.State == "provisioned" {
		kubeconfig, err := cluster.GetKubeconfig(false)
		if err != nil {
			return nil, fmt.Errorf("error getting Kubeconfig for the Kubernetes cluster with ID '%s': %s", cluster.ID, err)
		}
		dSet(d, "kubeconfig", kubeconfig)
	} else {
		warnings = append(warnings, fmt.Errorf("the Kubernetes cluster with ID '%s' is in '%s' state, meaning that "+
			"the Kubeconfig cannot be retrieved and "+
			"some attributes could be unavailable", cluster.ID, cluster.State))
	}

	d.SetId(cluster.ID)
	return warnings, nil
}
