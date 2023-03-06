---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension v4.0"
sidebar_current: "docs-vcd-guides-cse"
description: |-
  Provides guidance on configuring VCD to be able to install and use Container Service Extension v4.0
---

# Container Service Extension v4.0

## About

This guide describes the required steps to configure VCD to install the Container Service Extension (CSE) v4.0, that
will allow tenant users to deploy **Tanzu Kubernetes Grid Multi-cloud (TKGm)** clusters on VCD using Terraform.

To know more about CSE v4.0, you can visit [the documentation][cse_docs].

## Pre-requisites

-> Please read also the pre-requisites section in the [CSE documentation][cse_docs].

In order to complete the steps described in this guide, please be aware:

* CSE v4.0 is supported from VCD v10.4.0 or above, make sure your VCD appliance matches the criteria.
* Terraform provider needs to be v3.9.0 or above.
* Both CSE Server and the Bootstrap clusters require outbound internet connectivity.
* CSE v4.0 makes use of [ALB](/providers/vmware/vcd/latest/docs/guides/nsxt_alb) capabilities.

## Installation process

-> To install CSE v4.0, this guide will make use of the ready-to-use Terraform configuration located [here](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/install).
You can check it, customise it to your needs and apply. However, reading this guide first is recommended to understand what it does and how to use it.

The installation process is split in two independent steps that should be run separately:

- The first step creates the [Runtime Defined Entity Interfaces][rde_interface] and [Types][rde_type] that are required for CSE to work, a new [Role][role]
  and a CSE Administrator [User][user] that will be referenced later on in second step.
- The second step will configure remaining resources, like [Organizations][org], [VDCs][vdc], [Catalogs][catalog], Networks and [VMs][vm].

The reason for such as split is that Providers require to generate an [API token][api_token]
for the CSE Administrator user. This operation needs to be done outside the Terraform realm for security reasons, and it's
up to the Providers to decide the most ideal way to generate such a token for its CSE Administrator in their particular scenarios.

### Step 1: Create RDEs and the CSE Administrator user

-> This step of the installation refers to the Terraform configuration present [here][step1].

In the [given configuration][step1] you can find a file named `terraform.tfvars.example`, you need to rename it to `terraform.tfvars`
and change the values present there to the ones that fit with your needs.

This step will create the following:

- The required `VCDKEConfig` [RDE Interface][rde_interface] and [RDE Type][rde_type].
- The required `capvcdCluster` [RDE Type][rde_type].
- The **CSE Admin [Role][role]**, that specifies the required rights for the CSE Administrator to manage provider-sided elements of VCD.
- The **CSE Administrator [User][user]** that will administrate the CSE Server and other aspects of VCD that are directly related to CSE.
  Feel free to add more attributes like `description` or `full_name` if needed.

Once reviewed and applied with `terraform apply`, one **must login with the created CSE Administrator user to
generate an API token** that will be used in the next step.

### Step 2: Install CSE

-> This step of the installation refers to the Terraform configuration present [here][step2].

~> Be sure that previous step is successfully completed and the API token for the CSE Administrator user was created.

This step will create all the remaining elements to install CSE v4.0 in VCD. You can read subsequent sections
to have a better understanding of the building blocks that are described in the [proposed Terraform configuration][step2].

In this [configuration][step2] you can also find a file named `terraform.tfvars.example`, you need to rename it to `terraform.tfvars`
and change the values present there to the correct ones. You can also modify the proposed resources so they fit better to your needs.

#### Organizations

The [proposed configuration][step2] will create two new [Organizations][org], as specified in the [CSE documentation][cse_docs]:

- A Solutions [Organization][org]: This [Organization][org] will host all provider-scoped items, such as the CSE Server.
  It should only be accessible to the CSE Administrator and Providers.
- A Cluster [Organization][org]: This [Organization][org] will host the TKGm clusters for the users of this tenant to consume them.

If you already have these two [Organizations][org] created and you want to use them instead, you can leverage customising the [proposed configuration][step2]
to use the Organization [data source][org_d] to fetch them.

#### VM Sizing Policies

The [proposed configuration][step2] will create four VM Sizing Policies:

- `TKG extra_large`: 8 CPUs, 32GB RAM.
- `TKG large`: 4 CPUs, 16GB RAM.
- `TKG medium`: 2 CPUs, 8GB RAM.
- `TKG small`: 2 CPU, 4GB RAM.

These VM Sizing Policies should be applied as they are. Nothing should be changed here. They will be assigned to the Cluster
Organization's VDC to be able to dimension the created TKGm clusters (see section below).

#### VDCs

The [proposed configuration][step2] will create two [VDCs][vdc], one for the Solutions Organization and another one for the Cluster Organization.

You need to specify the following values in `terraform.tfvars`:

- `provider_vdc_name`: This is used to fetch the [Provider VDC][provider_vdc] used to create the two VDCs. If you are going to use more than
one Provider VDC, please consider modifying the proposed configuration.
- `nsxt_edge_cluster_name`: This is used to fetch the [Edge Cluster][edge_cluster] used to create the two VDCs. If you are going to use more than
one Edge Cluster, please consider modifying the proposed configuration.
- `network_pool_name`: This is used to create both VDCs. If you are going to use more than
one Network pool, please consider modifying the proposed configuration.

The Cluster Organization's VDC has all the VM Sizing Policies assigned, with the `TKG small` being the default one.
You can customise the `default_compute_policy_id` to make any other TKG policy the default one.

You can also leverage changing the storage profiles and other parameters to fit the requirements of your organization.

#### Catalog and OVAs

The [proposed configuration][step2] will create two catalogs:

- A catalog to host CSE OVA files, only accessible to CSE Administrators.
- A catalog to host TKGm OVA files, only accessible to CSE Administrators but shared as read-only to tenants.

Then it will upload the required OVAs to them. The OVAs can be specified in `terraform.tfvars`:

- `tkgm_ova_folder`: This will reference the path to the TKGm OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
- `tkgm_ova_file`: This will reference the file name of the TKGm OVA, like `ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933.ova`.
- `cse_ova_folder`: This will reference the path to the CSE OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
- `cse_ova_file`: This will reference the file name of the CSE OVA, like `VMware_Cloud_Director_Container_Service_Extension-4.0.1.ova`.

-> To download the required OVAs, please refer to the [CSE documentation][cse_docs].

If you need to upload more than one OVA, please modify the [proposed configuration][step2].

### "Kubernetes Cluster Author" global role

Apart from the role to administrate the CSE Server created in [step 1][step1], we also need a [Global Role][global_role]
for the TKGm clusters consumers (it would be similar to the concept of "vApp Author" but for TKGm clusters).

In order to create this [Global Role][global_role], the [proposed configuration][step2] first
creates a new [Rights Bundle][rights_bundle] and publishes it to all the tenants, then creates the [Global Role][global_role].

### Networking

The [proposed configuration][step2] configures a basic networking layout that will make CSE v4.0 work. However, it is
recommended that you review the code and adapt the different parts to your needs, specially for the resources like `vcd_nsxt_firewall`.

The configuration will create the following:

- A [Provider Gateway][provider_gateway] per Organization.
- An [Edge Gateway][edge_gateway] per Organization.
- Configure ALB with a shared Service Engine Group.
- A [Routed network][routed_network] per Organization.

In order to do so, the [proposed configuration][step2] asks for the following variables that you can customise in `terraform.tfvars`:

- `nsxt_manager_name`: It is required to create the [Provider Gateways][provider_gateway]. If you are going to use more than
  one [NSX-T Manager][nsxt_manager], please consider modifying the proposed configuration. 
- `nsxt_tier0_router_name`: It is required to create the [Provider Gateways][provider_gateway]. If you are going to use more than
  one [Tier-0 Router][nsxt_tier0_router], please consider modifying the proposed configuration.
- `solutions_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Solutions Organization.
- `solutions_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
- `solutions_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Solutions Organization.
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
- `cluster_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Cluster Organization.
- `cluster_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
- `cluster_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Cluster Organization.
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
- `alb_controller_url`: URL of the ALB controller that will be used. See the [ALB guide][alb] for more info.
- `alb_controller_username`: Username to access the ALB controller. See the [ALB guide][alb] for more info.
- `alb_controller_password`: Password of the username used to access the ALB controller. See the [ALB guide][alb] for more info.
- `alb_importable_cloud_name`: Name of the ALB Cloud defined in the ALB controller that will be imported to create an ALB Cloud in VCD. See the [ALB guide][alb] for more info.
- `solutions_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed] of the Solutions Organization.
- `solutions_routed_network_prefix_length`: The prefix length of the [Routed network][routed] of the Solutions Organization.
- `solutions_routed_network_ip_pool_start_address`: The [Routed network][routed] for the Solutions Organization has a pool of usable IPs, this field
  defines the first usable IP.
- `solutions_routed_network_ip_pool_end_address`: The [Routed network][routed] for the Solutions Organization has a pool of usable IPs, this field
  defines the end usable IP.
- `solutions_routed_network_advertised_subnet`: This enables route advertisement on the specified subnet, which should correspond to the Solutions
  Organization [Routed network][routed].
- `solutions_routed_network_dns`: DNS Server for the Solutions Organization [Routed network][routed]. It can be left blank if it's not needed.
- `cluster_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed] of the Cluster Organization.
- `cluster_routed_network_prefix_length`: The prefix length of the [Routed network][routed] of the Cluster Organization.
- `cluster_routed_network_ip_pool_start_address`: The [Routed network][routed] for the Cluster Organization has a pool of usable IPs, this field
  defines the first usable IP.
- `cluster_routed_network_ip_pool_end_address`: The [Routed network][routed] for the Cluster Organization has a pool of usable IPs, this field
  defines the end usable IP.
- `cluster_routed_network_advertised_subnet`: This enables route advertisement on the specified subnet, which should correspond to the Cluster
  Organization [Routed network][routed].
- `cluster_routed_network_dns`: DNS Server for the Cluster Organization [Routed network][routed]. It can be left blank if it's not needed.

If you wish to have a different networking setup, please modify the [proposed configuration][step2].

### CSE Server

The final set of resources created by the [proposed configuration][step2] correspond to the CSE Server vApp.
The generated VM makes use of the uploaded CSE OVA and some required guest properties.

In order to do so, the [configuration][step2] asks for the following variables that you can customise in `terraform.tfvars`:

- `vcdkeconfig_template_filepath`: This references a local file that corresponds with the `VCDKEConfig` [RDE][rde] contents specified as a JSON template.
  You can find this template [here](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/entities/vcdkeconfig-template.json).
  (Note: In `terraform.tfvars.example` the correct path is already provided).
- `capvcd_version`: The version for CAPVCD schema. It should be "1.1.0" for CSE v4.0.
- `cpi_version`: The version for CPI. It should be "1.2.0" for CSE v4.0.
- `csi_version`: The version for CSI. It should be "1.3.0" for CSE v4.0.
- `github_personal_access_token`: Create this one [here](https://github.com/settings/tokens), this will avoid installation errors caused by GitHub rate limiting.
- `cse_admin_user`: This should reference the CSE Administrator [User][user] that was created in Step 1.
- `cse_admin_api_token`: This should be the API token that you created for the CSE Administrator after Step 1.

### Final considerations

To evaluate the correctness of the setup, you can look up the CSE logs present in the CSE Server VM.
You can visit [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html)
to learn how to monitor the logs and troubleshoot possible problems.

## CSE upgrade process

Coming soon

## Cluster operations

Coming soon

### Create a cluster

Coming soon

```hcl
data "template_file" "k8s_cluster_yaml_template" {
  template = file("${path.module}/capvcd_templates/v1.22.9.yaml")
  vars = {
    cluster_name = var.k8s_cluster_name

    vcd_url     = replace(var.vcd_api_endpoint, "/api", "")
    org         = vcd_org.cluster_organization.name
    vdc         = vcd_org_vdc.cluster_vdc.name
    org_network = vcd_network_routed_v2.cluster_routed_network.name

    base64_username  = base64encode(var.k8s_cluster_user)
    base64_api_token = base64encode(var.k8s_cluster_api_token)

    ssh_public_key                   = ""
    control_plane_machine_count      = 1
    control_plane_sizing_policy_name = vcd_vm_sizing_policy.default_policy.name
    control_plane_sizing_policy_name = ""
    control_plane_placement_policy   = ""
    control_plane_storage_profile    = ""
    control_plane_disk_size          = "20Gi"

    worker_storage_policy   = ""
    worker_sizing_policy = vcd_vm_sizing_policy.default_policy.name
    worker_placement_policy = ""
    worker_machine_count    = 1
    worker_disk_size        = "20Gi"

    catalog_name = vcd_catalog.cse_catalog.name
    tkgm_ova     = replace(var.tkgm_ova_name, ".ova", "")

    pods_cidr     = "100.96.0.0/11"
    services_cidr = "100.64.0.0/13"
  }
}

# sample_cluster.yaml file should convert \n into \\n and " into \" first
data "template_file" "rde_k8s_cluster_instance_template" {
  template = file("${path.module}/entities/k8s_cluster.json")
  vars = {
    vcd_url   = replace(var.vcd_api_endpoint, "/api", "")
    name      = var.k8s_cluster_name
    org       = vcd_org.cluster_organization.name
    vdc       = vcd_org_vdc.cluster_vdc.name
    capi_yaml = replace(replace(data.template_file.k8s_cluster_yaml_template.rendered, "\n", "\\n"), "\"", "\\\"")

    delete                = false # Make this true to delete the cluster
    resolve_on_destroy    = false # Make this true to forcefully delete the cluster
    auto_repair_on_errors = false
  }
}

resource "vcd_rde" "k8s_cluster_instance" {
  org              = vcd_org.cluster_organization.name
  name             = var.k8s_cluster_name
  rde_type_vendor  = vcd_rde_type.capvcd_cluster_type.vendor
  rde_type_nss     = vcd_rde_type.capvcd_cluster_type.nss
  rde_type_version = vcd_rde_type.capvcd_cluster_type.version
  resolve          = false # MUST be false as it is resolved by CSE server
  resolve_on_destroy     = true  # MUST be true as it won't be resolved by Terraform
  input_entity     = data.template_file.rde_k8s_cluster_instance_template.rendered

  depends_on = [
    vcd_vapp_vm.cse_server_vm, vcd_catalog_vapp_template.tkgm_ova
  ]
}

output "computed_k8s_cluster_id" {
  value = vcd_rde.k8s_cluster_instance.id
}

output "computed_k8s_cluster_capvcdyaml" {
  value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["spec"]["capiYaml"]
}
```

### Retrieve a cluster Kubeconfig

Coming soon

```hcl
# output "kubeconfig" {  
#   value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["status"]["capvcd"]["private"]["kubeConfig"]
# }
```

### Upgrade a cluster

Coming soon

### Delete a cluster

Coming soon

~> Don't remove the resource from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

## Uninstall CSE

Before uninstalling CSE, make sure you perform an update operation to mark all clusters for deletion.

~> Don't remove the K8s cluster resources from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

Once all clusters are removed in the background by CSE Server, you may destroy the remaining infrastructure.

[alb]: https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/nsxt_alb
[api_token]: https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-A1B3B2FA-7B2C-4EE1-9D1B-188BE703EEDE.html
[catalog]: /providers/vmware/vcd/latest/docs/resources/catalog
[cse_docs]: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
[edge_cluster]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_edge_cluster
[edge_gateway]: /providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway
[global_role]: /providers/vmware/vcd/latest/docs/resources/global_role
[nsxt_manager]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_manager
[nsxt_tier0_router]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_tier0_router
[org]: /providers/vmware/vcd/latest/docs/resources/org
[org_d]: /providers/vmware/vcd/latest/docs/data-sources/org
[provider_gateway]: /providers/vmware/vcd/latest/docs/resources/external_network_v2
[provider_vdc]: /providers/vmware/vcd/latest/docs/data-sources/provider_vdc
[rights_bundle]: /providers/vmware/vcd/latest/docs/resources/rights_bundle
[rde]: /providers/vmware/vcd/latest/docs/resources/rde
[rde_interface]: /providers/vmware/vcd/latest/docs/resources/rde_interface
[rde_type]: /providers/vmware/vcd/latest/docs/resources/rde_type
[role]: /providers/vmware/vcd/latest/docs/resources/role
[routed_network]: /providers/vmware/vcd/latest/docs/resources/network_routed_v2
[sizing]: /providers/vmware/vcd/latest/docs/resources/vm_sizing_policy
[step1]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/install/step1
[step2]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/install/step2
[user]: /providers/vmware/vcd/latest/docs/resources/org_user
[vdc]: /providers/vmware/vcd/latest/docs/resources/org_vdc
[vm]: /providers/vmware/vcd/latest/docs/resources/vapp_vm