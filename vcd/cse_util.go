package vcd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"strconv"
	"strings"
	"text/template"
)

// tkgVersionBundle is a type that contains all the versions of the components of
// a Kubernetes cluster that can be obtained with the vApp Template name, downloaded
// from VMware Customer connect:
// https://customerconnect.vmware.com/downloads/details?downloadGroup=TKG-240&productId=1400
type tkgVersionBundle struct {
	EtcdVersion       string
	CoreDnsVersion    string
	TkgVersion        string
	TkrVersion        string
	KubernetesVersion string
}

// getTkgVersionBundleFromVAppTemplateName returns a tkgVersionBundle with the details of
// all the Kubernetes cluster components versions given a valid vApp Template name, that should
// correspond to a Kubernetes template. If it is not a valid vApp Template, returns an error.
func getTkgVersionBundleFromVAppTemplateName(ovaName string) (tkgVersionBundle, error) {
	versionsMap := map[string]map[string]string{
		"v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc": {
			"tkg":     "v2.2.0",
			"etcd":    "v3.5.6_vmware.9",
			"coreDns": "v1.9.3_vmware.8",
		},
		"v1.27.5+vmware.1-tkg.1-0eb96d2f9f4f705ac87c40633d4b69st": {
			"tkg":     "v2.4.0",
			"etcd":    "v3.5.7_vmware.6",
			"coreDns": "v1.10.1_vmware.7",
		},
		"v1.26.8+vmware.1-tkg.1-b8c57a6c8c98d227f74e7b1a9eef27st": {
			"tkg":     "v2.4.0",
			"etcd":    "v3.5.6_vmware.20",
			"coreDns": "v1.10.1_vmware.7",
		},
		"v1.26.8+vmware.1-tkg.1-0edd4dafbefbdb503f64d5472e500cf8": {
			"tkg":     "v2.3.1",
			"etcd":    "v3.5.6_vmware.20",
			"coreDns": "v1.9.3_vmware.16",
		},
	}

	result := tkgVersionBundle{}

	if strings.Contains(ovaName, "photon") {
		return result, fmt.Errorf("the vApp Template '%s' uses Photon, and it is not supported", ovaName)
	}

	cutPosition := strings.LastIndex(ovaName, "kube-")
	if cutPosition < 0 {
		return result, fmt.Errorf("the vApp Template '%s' is not a Kubernetes template OVA", ovaName)
	}
	parsedOvaName := strings.ReplaceAll(ovaName, ".ova", "")[cutPosition+len("kube-"):]
	if _, ok := versionsMap[parsedOvaName]; !ok {
		return result, fmt.Errorf("the Kubernetes OVA '%s' is not supported", parsedOvaName)
	}

	// The map checking above guarantees that all splits and replaces will work
	result.KubernetesVersion = strings.Split(parsedOvaName, "-")[0]
	result.TkrVersion = strings.ReplaceAll(strings.Split(parsedOvaName, "-")[0], "+", "---") + "-" + strings.Split(parsedOvaName, "-")[1]
	result.TkgVersion = versionsMap[parsedOvaName]["tkg"]
	result.EtcdVersion = versionsMap[parsedOvaName]["etcd"]
	result.CoreDnsVersion = versionsMap[parsedOvaName]["coreDns"]
	return result, nil
}

// createClusterDto is a helper struct that contains all the required elements to successfully create a Kubernetes cluster using CSE.
type createClusterDto struct {
	Name            string
	VcdUrl          string
	Org             *govcd.AdminOrg
	VdcName         string
	OvaName         string
	CatalogName     string
	NetworkName     string
	RdeType         *govcd.DefinedEntityType
	UrnToNamesCache map[string]string // Maps unique IDs with their resource names (example: Compute policy ID with its name)
	VCDKEConfig     struct {
		MaxUnhealthyNodesPercentage string
		NodeStartupTimeout          string
		NodeNotReadyTimeout         string
		NodeUnknownTimeout          string
		ContainerRegistryUrl        string
	}
	TkgVersion tkgVersionBundle
	Owner      string
	ApiToken   string
}

// getClusterCreateDto creates and returns a createClusterDto object by obtaining all the required information
// from the Terraform resource data and the target VCD.
func getClusterCreateDto(d *schema.ResourceData, vcdClient *VCDClient, vcdKeConfigRdeTypeVersion, capvcdClusterRdeTypeVersion string) (*createClusterDto, error) {
	result := &createClusterDto{}
	result.UrnToNamesCache = map[string]string{"": ""} // Initialize with a "zero" entry, used when there's no ID set in the Terraform schema

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
	result.VdcName = vdc.Vdc.Name

	vAppTemplateId := d.Get("ova_id").(string)
	vAppTemplate, err := vcdClient.GetVAppTemplateById(vAppTemplateId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Kubernetes OVA with ID '%s': %s", vAppTemplateId, err)
	}
	result.OvaName = vAppTemplate.VAppTemplate.Name

	tkgVersions, err := getTkgVersionBundleFromVAppTemplateName(vAppTemplate.VAppTemplate.Name)
	if err != nil {
		return nil, err
	}
	result.TkgVersion = tkgVersions

	catalogName, err := vAppTemplate.GetCatalogName()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the CatalogName of the OVA '%s': %s", vAppTemplateId, err)
	}
	result.CatalogName = catalogName

	networkId := d.Get("network_id").(string)
	network, err := vdc.GetOrgVdcNetworkById(networkId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Org VDC NetworkName with ID '%s': %s", networkId, err)
	}
	result.NetworkName = network.OrgVDCNetwork.Name

	rdeType, err := vcdClient.GetRdeType("vmware", "capvcdCluster", capvcdClusterRdeTypeVersion)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve RDE Type vmware:capvcdCluster:'%s': %s", capvcdClusterRdeTypeVersion, err)
	}
	result.RdeType = rdeType

	// Builds a map that relates storage profiles IDs (the schema uses them to build a healthy Terraform dependency graph)
	// with their corresponding names (the cluster YAML and CSE in general uses names only).
	// Having this map minimizes the amount of queries to VCD, specially when building the set of node pools,
	// as there can be a lot of them.
	if _, isStorageClassSet := d.GetOk("default_storage_class"); isStorageClassSet {
		storageProfileId := d.Get("default_storage_class.0.storage_profile_id").(string)
		storageProfile, err := vcdClient.GetStorageProfileById(storageProfileId)
		if err != nil {
			return nil, fmt.Errorf("could not get a Storage Profile with ID '%s' for the Storage Class: %s", storageProfileId, err)
		}
		result.UrnToNamesCache[storageProfileId] = storageProfile.Name
	}
	controlPlaneStorageProfileId := d.Get("control_plane.0.storage_profile_id").(string)
	if _, ok := result.UrnToNamesCache[controlPlaneStorageProfileId]; !ok { // Only query if not already present
		storageProfile, err := vcdClient.GetStorageProfileById(controlPlaneStorageProfileId)
		if err != nil {
			return nil, fmt.Errorf("could not get a Storage Profile with ID '%s' for the Control Plane: %s", controlPlaneStorageProfileId, err)
		}
		result.UrnToNamesCache[controlPlaneStorageProfileId] = storageProfile.Name
	}
	for _, nodePoolRaw := range d.Get("node_pool").(*schema.Set).List() {
		nodePool := nodePoolRaw.(map[string]interface{})
		nodePoolStorageProfileId := nodePool["storage_profile_id"].(string)
		if _, ok := result.UrnToNamesCache[nodePoolStorageProfileId]; !ok { // Only query if not already present
			storageProfile, err := vcdClient.GetStorageProfileById(nodePoolStorageProfileId)
			if err != nil {
				return nil, fmt.Errorf("could not get a Storage Profile with ID '%s' for the Node Pool: %s", controlPlaneStorageProfileId, err)
			}
			result.UrnToNamesCache[nodePoolStorageProfileId] = storageProfile.Name
		}
	}

	// Builds a map that relates Compute Policies IDs (the schema uses them to build a healthy Terraform dependency graph)
	// with their corresponding names (the cluster YAML and CSE in general uses names only).
	// Having this map minimizes the amount of queries to VCD, specially when building the set of node pools,
	// as there can be a lot of them.
	if controlPlaneSizingPolicyId, isSet := d.GetOk("control_plane.0.sizing_policy_id"); isSet {
		computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(controlPlaneSizingPolicyId.(string))
		if err != nil {
			return nil, fmt.Errorf("could not get a Sizing Policy with ID '%s' for the Control Plane: %s", controlPlaneStorageProfileId, err)
		}
		result.UrnToNamesCache[controlPlaneSizingPolicyId.(string)] = computePolicy.VdcComputePolicyV2.Name
	}
	if controlPlanePlacementPolicyId, isSet := d.GetOk("control_plane.0.placement_policy_id"); isSet {
		if _, ok := result.UrnToNamesCache[controlPlanePlacementPolicyId.(string)]; !ok { // Only query if not already present
			computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(controlPlanePlacementPolicyId.(string))
			if err != nil {
				return nil, fmt.Errorf("could not get a Placement Policy with ID '%s' for the Control Plane: %s", controlPlaneStorageProfileId, err)
			}
			result.UrnToNamesCache[controlPlanePlacementPolicyId.(string)] = computePolicy.VdcComputePolicyV2.Name
		}
	}
	for _, nodePoolRaw := range d.Get("node_pool").(*schema.Set).List() {
		nodePool := nodePoolRaw.(map[string]interface{})
		if nodePoolSizingPolicyId, isSet := nodePool["sizing_policy_id"]; isSet {
			if _, ok := result.UrnToNamesCache[nodePoolSizingPolicyId.(string)]; !ok { // Only query if not already present
				computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(nodePoolSizingPolicyId.(string))
				if err != nil {
					return nil, fmt.Errorf("could not get a Sizing Policy with ID '%s' for the Node Pool: %s", controlPlaneStorageProfileId, err)
				}
				result.UrnToNamesCache[nodePoolSizingPolicyId.(string)] = computePolicy.VdcComputePolicyV2.Name
			}
		}
		if nodePoolPlacementPolicyId, isSet := nodePool["placement_policy_id"]; isSet {
			if _, ok := result.UrnToNamesCache[nodePoolPlacementPolicyId.(string)]; !ok { // Only query if not already present
				computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(nodePoolPlacementPolicyId.(string))
				if err != nil {
					return nil, fmt.Errorf("could not get a Placement Policy with ID '%s' for the Node Pool: %s", controlPlaneStorageProfileId, err)
				}
				result.UrnToNamesCache[nodePoolPlacementPolicyId.(string)] = computePolicy.VdcComputePolicyV2.Name
			}
		}
		if nodePoolVGpuPolicyId, isSet := nodePool["vgpu_policy_id"]; isSet {
			if _, ok := result.UrnToNamesCache[nodePoolVGpuPolicyId.(string)]; !ok { // Only query if not already present
				computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(nodePoolVGpuPolicyId.(string))
				if err != nil {
					return nil, fmt.Errorf("could not get a Placement Policy with ID '%s' for the Node Pool: %s", controlPlaneStorageProfileId, err)
				}
				result.UrnToNamesCache[nodePoolVGpuPolicyId.(string)] = computePolicy.VdcComputePolicyV2.Name
			}
		}
	}

	rdes, err := vcdClient.GetRdesByName("vmware", "VCDKEConfig", vcdKeConfigRdeTypeVersion, "vcdKeConfig")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", vcdKeConfigRdeTypeVersion, err)
	}
	if len(rdes) != 1 {
		return nil, fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}

	// Obtain some required elements from the CSE Server configuration (aka VCDKEConfig), so we don't have
	// to deal with it again.
	type vcdKeConfigType struct {
		Profiles []struct {
			K8Config struct {
				Mhc struct {
					MaxUnhealthyNodes   int `json:"maxUnhealthyNodes:omitempty"`
					NodeStartupTimeout  int `json:"nodeStartupTimeout:omitempty"`
					NodeNotReadyTimeout int `json:"nodeNotReadyTimeout:omitempty"`
					NodeUnknownTimeout  int `json:"nodeUnknownTimeout:omitempty"`
				} `json:"mhc:omitempty"`
			} `json:"K8Config:omitempty"`
			ContainerRegistryUrl string `json:"containerRegistryUrl,omitempty"`
		} `json:"profiles,omitempty"`
	}

	var vcdKeConfig vcdKeConfigType
	rawData, err := json.Marshal(rdes[0].DefinedEntity.Entity)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawData, &vcdKeConfig)
	if err != nil {
		return nil, err
	}

	if len(vcdKeConfig.Profiles) != 1 {
		return nil, fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(vcdKeConfig.Profiles))
	}

	result.VCDKEConfig.MaxUnhealthyNodesPercentage = strconv.Itoa(vcdKeConfig.Profiles[0].K8Config.Mhc.MaxUnhealthyNodes)
	result.VCDKEConfig.NodeStartupTimeout = strconv.Itoa(vcdKeConfig.Profiles[0].K8Config.Mhc.NodeStartupTimeout)
	result.VCDKEConfig.NodeNotReadyTimeout = strconv.Itoa(vcdKeConfig.Profiles[0].K8Config.Mhc.NodeNotReadyTimeout)
	result.VCDKEConfig.NodeUnknownTimeout = strconv.Itoa(vcdKeConfig.Profiles[0].K8Config.Mhc.NodeUnknownTimeout)
	result.VCDKEConfig.ContainerRegistryUrl = fmt.Sprintf("%s/tkg", vcdKeConfig.Profiles[0].ContainerRegistryUrl)

	owner, ok := d.GetOk("owner")
	if !ok {
		sessionInfo, err := vcdClient.Client.GetSessionInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting the owner of the cluster: %s", err)
		}
		owner = sessionInfo.User.Name
	}
	result.Owner = owner.(string)

	apiToken, err := govcd.GetTokenFromFile(d.Get("api_token_file").(string))
	if err != nil {
		return nil, fmt.Errorf("API token file could not be parsed or found: %s\nPlease check that the format is the one that 'vcd_api_token' resource uses", err)
	}
	result.ApiToken = apiToken.RefreshToken

	result.VcdUrl = strings.Replace(vcdClient.VCDClient.Client.VCDHREF.String(), "/api", "", 1)
	return result, nil
}

// generateCapiYaml generates the YAML string that is required during Kubernetes cluster creation, to be embedded
// in the CAPVCD cluster JSON payload. This function picks data from the Terraform schema and the createClusterDto to
// populate several Go templates and build a final YAML.
func generateCapiYaml(d *schema.ResourceData, clusterDetails *createClusterDto) (string, error) {
	// This YAML snippet contains special strings, such as "%,", that render wrong using the Go template engine
	sanitizedTemplate := strings.NewReplacer("%", "%%").Replace(cseClusterYamlTemplate)
	capiYamlEmpty := template.Must(template.New(clusterDetails.Name + "_CapiYaml").Parse(sanitizedTemplate))

	nodePoolYaml, err := generateNodePoolYaml(d, clusterDetails)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	args := map[string]string{
		"ClusterName":                 clusterDetails.Name,
		"TargetNamespace":             clusterDetails.Name + "-ns",
		"TkrVersion":                  clusterDetails.TkgVersion.TkrVersion,
		"TkgVersion":                  clusterDetails.TkgVersion.TkgVersion,
		"UsernameB64":                 base64.StdEncoding.EncodeToString([]byte(clusterDetails.Owner)),
		"ApiTokenB64":                 base64.StdEncoding.EncodeToString([]byte(clusterDetails.ApiToken)),
		"PodCidr":                     d.Get("pods_cidr").(string),
		"ServiceCidr":                 d.Get("services_cidr").(string),
		"VcdSite":                     clusterDetails.VcdUrl,
		"Org":                         clusterDetails.Org.AdminOrg.Name,
		"OrgVdc":                      clusterDetails.VdcName,
		"OrgVdcNetwork":               clusterDetails.NetworkName,
		"Catalog":                     clusterDetails.CatalogName,
		"VAppTemplate":                clusterDetails.OvaName,
		"ControlPlaneSizingPolicy":    clusterDetails.UrnToNamesCache[d.Get("control_plane.0.sizing_policy_id").(string)],
		"ControlPlanePlacementPolicy": clusterDetails.UrnToNamesCache[d.Get("control_plane.0.placement_policy_id").(string)],
		"ControlPlaneStorageProfile":  clusterDetails.UrnToNamesCache[d.Get("control_plane.0.storage_profile_id").(string)],
		"ControlPlaneDiskSize":        fmt.Sprintf("%dGi", d.Get("control_plane.0.disk_size_gi").(int)),
		"ControlPlaneMachineCount":    strconv.Itoa(d.Get("control_plane.0.machine_count").(int)),
		"DnsVersion":                  clusterDetails.TkgVersion.CoreDnsVersion,
		"EtcdVersion":                 clusterDetails.TkgVersion.EtcdVersion,
		"ContainerRegistryUrl":        clusterDetails.VCDKEConfig.ContainerRegistryUrl,
		"KubernetesVersion":           clusterDetails.TkgVersion.KubernetesVersion,
		"SshPublicKey":                d.Get("ssh_public_key").(string),
	}

	if _, ok := d.GetOk("control_plane.0.ip"); ok {
		args["ControlPlaneEndpoint"] = d.Get("control_plane.0.ip").(string)
	}
	if _, ok := d.GetOk("virtual_ip_subnet"); ok {
		args["VirtualIpSubnet"] = d.Get("virtual_ip_subnet").(string)
	}

	if d.Get("node_health_check").(bool) {
		args["MaxUnhealthyNodePercentage"] = fmt.Sprintf("%s%%", clusterDetails.VCDKEConfig.MaxUnhealthyNodesPercentage) // With the 'percentage' suffix, it is doubled to render the template correctly
		args["NodeStartupTimeout"] = fmt.Sprintf("%ss", clusterDetails.VCDKEConfig.NodeStartupTimeout)                   // With the 'second' suffix
		args["NodeUnknownTimeout"] = fmt.Sprintf("%ss", clusterDetails.VCDKEConfig.NodeUnknownTimeout)                   // With the 'second' suffix
		args["NodeNotReadyTimeout"] = fmt.Sprintf("%ss", clusterDetails.VCDKEConfig.NodeNotReadyTimeout)                 // With the 'second' suffix
	}

	if err := capiYamlEmpty.Execute(buf, args); err != nil {
		return "", fmt.Errorf("could not generate a correct CAPI YAML: %s", err)
	}

	prettyYaml := fmt.Sprintf("%s\n%s", nodePoolYaml, buf.String())

	// This encoder is used instead of a standard json.Marshal as the YAML contains special
	// characters that are not encoded properly, such as '<'.
	buf.Reset()
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err = enc.Encode(prettyYaml)
	if err != nil {
		return "", fmt.Errorf("could not encode the CAPI YAML into JSON: %s", err)
	}

	return strings.Trim(strings.TrimSpace(buf.String()), "\""), nil
}
