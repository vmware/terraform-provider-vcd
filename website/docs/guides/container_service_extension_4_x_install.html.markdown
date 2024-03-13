---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension 4.2 installation"
sidebar_current: "docs-vcd-guides-cse-4-x-install"
description: |-
  Provides guidance on configuring VCD to be able to install and use Container Service Extension 4.2
---

# Container Service Extension 4.2 installation

## About

This guide describes the required steps to configure VCD to install the Container Service Extension (CSE) 4.2, that
will allow tenant users to deploy **Tanzu Kubernetes Grid Multi-cloud (TKGm)** clusters on VCD using Terraform or the UI.

To know more about CSE 4.2, you can visit [the documentation][cse_docs].

## Pre-requisites

-> Please read also the pre-requisites section in the [CSE documentation][cse_docs].

In order to complete the steps described in this guide, please be aware:

* CSE 4.2 is supported from VCD v10.4.2 or above, as specified in the [Product Interoperability Matrix][product_matrix].
  Please check that the target VCD appliance matches the criteria.
* Terraform provider needs to be v3.12.0 or above.
* Both CSE Server and the Bootstrap clusters require outbound Internet connectivity.
* CSE 4.2 makes use of [ALB](/providers/vmware/vcd/latest/docs/guides/nsxt_alb) capabilities.

## Installation process

-> To install CSE 4.2, this guide will make use of the example Terraform configuration located [here](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.2/install).
You can check it, customise it to your needs and apply. However, reading this guide first is recommended to understand what it does and how to use it.

The installation process is split in two independent steps that must be run one after the other:

* The first step installs the same elements as the _"Configure Settings for CSE Server"_ section in UI wizard, that is, creates the
  [RDE Interfaces][rde_interface], [RDE Types][rde_type], [RDE Interface Behaviors][rde_interface_behavior] and the [RDE][rde] that
  are required for the CSE Server to work, in addition to a new [Role][role], new [VM Sizing Policies][sizing] and a CSE Administrator [User][user] that will be
  referenced later on in the second step.
* The second step configures the remaining resources, like [Organizations][org], [VDCs][vdc], [Catalogs and OVAs][catalog], Networks, and the CSE Server [VM][vm].

The reason for such as split is that the CSE Administrator created during the first step is used to configure a new `provider` block in
the second one, so that it can provision a valid [API token][api_token]. This operation must be done separately as a `provider` block
can't log in with a user created in the same run.

### Step 1: Configure Settings for CSE Server

-> This step of the installation refers to the [step 1 of the example Terraform configuration][step1].

This step will create the same elements as the _"Configure Settings for CSE Server"_ section in UI wizard. The subsections
below can be helpful to understand all the building blocks that are described in the proposed example of Terraform configuration.

In the directory there is also a file named `terraform.tfvars.example`, which needs to be renamed to `terraform.tfvars`
and its values to be set to the correct ones. In general, for this specific step, the proposed HCL files (`.tf`) should not be
modified and be applied as they are.

#### RDE Interfaces, Types and Behaviors

CSE 4.2 requires a set of Runtime Defined Entity items, such as [Interfaces][rde_interface], [Types][rde_type] and [Behaviors][rde_interface_behavior].
In the [step 1 configuration][step1] you can find the following:

* The required `VCDKEConfig` [RDE Interface][rde_interface] and [RDE Type][rde_type]. These two resources specify the schema of the **CSE Server
  configuration** that will be instantiated with a [RDE][rde].

* The required `capvcd` [RDE Interface][rde_interface] and `capvcdCluster` [RDE Type][rde_type].
  These two resources specify the schema of the [TKGm clusters][tkgm_docs].

* The required [RDE Interface Behaviors][rde_interface_behavior] used to retrieve critical information from the [TKGm clusters][tkgm_docs],
  for example, the resulting **Kubeconfig**.

#### RDE (CSE Server configuration / VCDKEConfig)

The CSE Server configuration lives in a [Runtime Defined Entity][rde] that uses the `VCDKEConfig` [RDE Type][rde_type].
To customise it, the [step 1 configuration][step1] asks for the following variables that you can set in `terraform.tfvars`:

* `vcdkeconfig_template_filepath` references a local file that defines the `VCDKEConfig` [RDE][rde] contents.
  It should be a JSON file with template variables that Terraform can interpret, like
  [the RDE template file for CSE 4.2](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.2/entities/vcdkeconfig.json.template)
  used in the step 1 configuration, that can be rendered correctly with the Terraform built-in function `templatefile`.
  (Note: In `terraform.tfvars.example` the path for the CSE 4.2 RDE contents is already provided).
* `capvcd_version`: The version for CAPVCD. Must be **"1.5.0"** for CSE 4.2.0, or **"1.6.0"** for CSE 4.2.1.
  (Note: Do not confuse with the version of the `capvcdCluster` [RDE Type][rde_type], which **must be "1.3.0"** for CSE 4.2.X, and cannot be changed through a variable).
* `cpi_version`: The version for CPI (Cloud Provider Interface). Must be **"1.5.0"** for CSE 4.2.0, or **"1.6.0"** for CSE 4.2.1.
* `csi_version`: The version for CSI (Cloud Storage Interface). Must be **"1.5.0"** for CSE 4.2.0, or **"1.6.0"** for CSE 4.2.1.
* `rde_projector_version`: The version for the RDE Projector. The default value is **"0.7.0"** for CSE 4.2.X.
* `github_personal_access_token`: Create this one [here](https://github.com/settings/tokens),
  this will avoid installation errors caused by GitHub rate limiting, as the TKGm cluster creation process requires downloading
  some Kubernetes components from GitHub.
  The token should have the `public_repo` scope for classic tokens and `Public Repositories` for fine-grained tokens.
* `http_proxy`: Address of your HTTP proxy server. Optional in the step 1 configuration.
* `https_proxy`: Address of your HTTPS proxy server. Optional in the step 1 configuration.
* `no_proxy`: A list of comma-separated domains without spaces that indicate the targets that must **not** go through the configured proxy. Optional in the step 1 configuration.
* `syslog_host`: Domain where to send the system logs. Optional in the step 1 configuration.
* `syslog_port`: Port where to send the system logs. Optional in the step 1 configuration.
* `node_startup_timeout`: A node will be considered unhealthy and remediated if joining the cluster takes longer than this timeout (seconds, defaults to 900 in the step 1 configuration).
* `node_not_ready_timeout`: A newly joined node will be considered unhealthy and remediated if it cannot host workloads for longer than this timeout (seconds, defaults to 300 in the step 1 configuration).
* `node_unknown_timeout`: A healthy node will be considered unhealthy and remediated if it is unreachable for longer than this timeout (seconds, defaults to 300 in the step 1 configuration).
* `max_unhealthy_node_percentage`: Remediation will be suspended when the number of unhealthy nodes exceeds this percentage.
  (100% means that unhealthy nodes will always be remediated, while 0% means that unhealthy nodes will never be remediated). Defaults to 100 in the step 1 configuration.
* `container_registry_url`: URL from where TKG clusters will fetch container images, useful for VCD appliances that are completely isolated from Internet. Defaults to "projects.registry.vmware.com" in the step 1 configuration.
* `bootstrap_vm_certificates`: Certificate(s) to allow the ephemeral VM (created during cluster creation) to authenticate with.
  For example, when pulling images from a container registry. Optional in the step 1 configuration.
* `k8s_cluster_certificates`: Certificate(s) to allow clusters to authenticate with.
  For example, when pulling images from a container registry. Optional in the step 1 configuration.

#### Rights, Roles and VM Sizing Policies

CSE 4.2 requires a set of new [Rights Bundles][rights_bundle], [Roles][role] and [VM Sizing Policies][sizing] that are also created
in this step of the [step 1 configuration][step1]. Nothing should be customised here, except for the "CSE Administrator"
account to be created, where you can provide a username of your choice (`cse_admin_username`) and its password (`cse_admin_password`).

This account will be used in the next step to provision an [API Token][api_token] to deploy the CSE Server.

Once all variables are reviewed and set, you can start the installation with `terraform apply`. When it finishes successfully, you can continue with the **step 2**.

### Step 2: Create the infrastructure and deploy the CSE Server

-> This step of the installation refers to the Terraform configuration present [here][step2].

~> Make sure that the previous step is successfully completed.

This step will create all the remaining elements to install CSE 4.2 in VCD. You can read the subsequent sections
to have a better understanding of the building blocks that are described in the [step 2 Terraform configuration][step2].

In this [configuration][step2] you can also find a file named `terraform.tfvars.example` that needs to be updated with correct values and renamed to `terraform.tfvars`
and change the values present there to the correct ones. You can also modify the proposed resources to better suit your needs.

#### Organizations

The [step 2 configuration][step2] will create two new [Organizations][org], as specified in the [CSE documentation][cse_docs]:

* A Solutions [Organization][org], which will host all provider-scoped items, such as the CSE Server.
  It should only be accessible to the CSE Administrator and Providers.
* A Tenant [Organization][org], which will host the [TKGm clusters][tkgm_docs] for the users of this tenant to consume them.

-> If you already have these two [Organizations][org] created and you want to use them instead,
you can leverage customising the [step 2 configuration][step2] to use the Organization [data source][org_d] to fetch them.

#### VDCs

The [step 2 configuration][step2] will create two [VDCs][vdc], one for the Solutions Organization and another one for the Tenant Organization.

You need to specify the following values in `terraform.tfvars`:

* `provider_vdc_name`: This is used to fetch an existing [Provider VDC][provider_vdc], that will be used to create the two VDCs.
  If you are going to use more than one [Provider VDC][provider_vdc], please consider modifying the step 2 configuration.
  In UI, [Provider VDCs][provider_vdc] can be found in the Provider view, inside _Cloud Resources_ menu.
* `nsxt_edge_cluster_name`: This is used to fetch an existing [Edge Cluster][edge_cluster], that will be used to create the two VDCs.
  If you are going to use more than one [Edge Cluster][edge_cluster], please consider modifying the step 2 configuration.
  In UI, [Edge Clusters][edge_cluster] can be found in the NSX-T manager web UI.
* `network_pool_name`: This references an existing Network Pool, which is used to create both VDCs.
  If you are going to use more than one Network Pool, please consider modifying the step 2 configuration.

In the [step 2 configuration][step2] the Tenant Organization's VDC has all the required VM Sizing Policies from the first step assigned,
with the `TKG small` being the default one. You can customise it to make any other TKG policy the default one.

You can also leverage changing the storage profiles and other parameters to fit the requirements of your organization. Also,
if you already have usable [VDCs][vdc], you can change the configuration to fetch them instead.

#### Catalog and OVAs

The [step 2 configuration][step2] will create two catalogs:

* A catalog to host CSE Server OVA files, only accessible to CSE Administrators. This catalog will allow CSE Administrators to organise and manage
  all the CSE Server OVAs that are required to run and upgrade the CSE Server.
* A catalog to host TKGm OVA files, only accessible to CSE Administrators but shared as read-only to tenants, that can use them to create [TKGm clusters][tkgm_docs].

Then it will upload the required OVAs to them. The OVAs can be specified in `terraform.tfvars`:

* `tkgm_ova_folder`: This will reference the path to the TKGm OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
* `tkgm_ova_files`: This will reference the file names of the TKGm OVAs, like `[ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc.ova, ubuntu-2004-kube-v1.24.11+vmware.1-tkg.1-2ccb2a001f8bd8f15f1bfbc811071830.ova]`.
* `cse_ova_folder`: This will reference the path to the CSE OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
* `cse_ova_file`: This will reference the file name of the CSE OVA, like `VMware_Cloud_Director_Container_Service_Extension-4.2.1.ova`.

-> To download the required OVAs, please refer to the [CSE documentation][cse_docs]. 
You can also check the [Product Interoperability Matrix][product_matrix] to confirm the appropriate version of TKGm.

~> Both CSE Server and TKGm OVAs are heavy. Please take into account that the upload process could take more than 30 minutes, depending
on upload speed. You can tune the `upload_piece_size` to speed up the upload. Another option would be uploading them manually in the UI.
In case you're using a pre-uploaded OVA, leverage the [vcd_catalog_vapp_template][catalog_vapp_template_ds] data source (instead of the resource).

#### Networking

The [step 2 configuration][step2] prepares a basic networking layout that will make CSE 4.2 work. However, it is
recommended that you review the code and adapt the different parts to your needs, specially for the resources like `vcd_nsxt_firewall`.

The configuration will create the following:

* A [Provider Gateway][provider_gateway] per Organization. You can learn more about Provider Gateways [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-E6BAC24B-9628-495A-BA67-6DE6C5CF70F2.html).
  In this configuration we just expose some static IPs to the two Organizations, so they can consume them.
* An [Edge Gateway][edge_gateway] per Organization. You can learn more about Edge Gateways [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-45C0FEDF-84F2-4487-8DB8-3BC281EB25CD.html).
  In this configuration we create two that act as a router for each Organization that we created.
* Configure ALB with a shared Service Engine Group. You can learn more about Advanced Load Balancers [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-92A0563D-A272-4958-B732-9C35901D9DB8.html).
  In this setup, we provide a virtual service pool that CSE Server uses to provide load balancing capabilities to the [TKGm clusters][tkgm_docs].
* A [Routed network][routed_network] per Organization. You can learn more about Routed networks [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-74C4D27F-9E2A-4EB2-BBE1-CDD45C80E270.html).
  In this setup, we just provide a routed network per organization, so the CSE Server is inside its own network, isolated from the [TKGm clusters][tkgm_docs] network.
* Two [SNAT rules][nat_rule] that will allow outbound access. Feel free to adjust or replace these rules with other ways of providing outbound access.

~> SNAT rules is just a proposal to give the CSE Server and the clusters outbound access. Please review the [step 2 configuration][step2]
first.

In order to create all the items listed above, the [step 2 configuration][step2] asks for the following variables that you can customise in `terraform.tfvars`:

* `nsxt_manager_name`: It is the name of an existing [NSX-T Manager][nsxt_manager], which is needed in order to create the [Provider Gateways][provider_gateway].
  If you are going to use more than one [NSX-T Manager][nsxt_manager], please consider modifying the step 2 configuration.
  In UI, [NSX-T Managers][nsxt_manager] can be found in the Provider view, inside _Infrastructure Resources > NSX-T_.
* `solutions_nsxt_tier0_router_name`: It is the name of an existing [Tier-0 Router][nsxt_tier0_router], which is needed in order to create the [Provider Gateway][provider_gateway] in the Solutions Organization.
  In UI, [Tier-0 Routers][nsxt_tier0_router] can be found in the NSX-T manager web UI.
* `solutions_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Solutions Organization.
* `solutions_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
* `solutions_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Solutions Organization.
  At least one IP is required. You can check the minimum amount of IPs required in [CSE documentation][cse_docs].
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
* `tenant_nsxt_tier0_router_name`: It is the name of an existing [Tier-0 Router][nsxt_tier0_router], which is needed in order to create the [Provider Gateway][provider_gateway] in the Tenant Organization.
  In UI, [Tier-0 Routers][nsxt_tier0_router] can be found in the NSX-T manager web UI.
* `tenant_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Tenant Organization.
* `tenant_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
* `tenant_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Tenant Organization.
  At least one IP is required. You can check the minimum amount of IPs required in [CSE documentation][cse_docs].
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
* `alb_controller_url`: URL of an existing ALB controller that will be created in VCD side. See the [ALB guide][alb] for more info.
* `alb_controller_username`: Username to access the ALB controller. See the [ALB guide][alb] for more info.
* `alb_controller_password`: Password of the username used to access the ALB controller. See the [ALB guide][alb] for more info.
* `alb_importable_cloud_name`: Name of the existing ALB Cloud defined in the ALB controller that will be imported to create an ALB Cloud in VCD. See the [ALB guide][alb] for more info.
* `solutions_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed_network] that will be created in the Solutions Organization.
* `solutions_routed_network_prefix_length`: The prefix length of the [Routed network][routed_network] that will be created in the Solutions Organization.
* `solutions_routed_network_ip_pool_start_address`: The [Routed network][routed_network] that will be created in the Solutions Organization will have a pool of usable IPs, this field
  defines the first usable IP.
* `solutions_routed_network_ip_pool_end_address`: The [Routed network][routed_network] that will be created in the Solutions Organization will have a pool of usable IPs, this field
  defines the end usable IP.
* `solutions_snat_external_ip`: This is used to create a SNAT rule on the Solutions Edge Gateway to provide Internet connectivity to the CSE Server. The external IP should be one available IP of the Solutions
  [Provider Gateway][provider_gateway].
* `solutions_snat_internal_network_cidr`: This is used to create a SNAT rule on the Solutions Edge Gateway to provide Internet connectivity to the CSE Server. The subnet should correspond to the Solutions
  Organization [Routed network][routed_network].
* `solutions_routed_network_dns`: DNS Server for the Solutions Organization [Routed network][routed_network]. It can be left blank if it's not needed.
* `tenant_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed_network] that will be created in the Tenant Organization.
* `tenant_routed_network_prefix_length`: The prefix length of the [Routed network][routed_network] that will be created in the Tenant Organization.
* `tenant_routed_network_ip_pool_start_address`: The [Routed network][routed_network] that will be created in the Tenant Organization will have a pool of usable IPs, this field
  defines the first usable IP.
* `tenant_routed_network_ip_pool_end_address`: The [Routed network][routed_network] that will be created in the Tenant Organization will have a pool of usable IPs, this field
  defines the end usable IP.
* `tenant_snat_external_ip`: This is used to create a SNAT rule on the Tenant Edge Gateway to provide Internet connectivity to the clusters. The external IP should be one available IP of the Tenant
  [Provider Gateway][provider_gateway].
* `tenant_snat_internal_network_cidr`: This is used to create a SNAT rule on the Tenant Edge Gateway to provide Internet connectivity to the clusters. The subnet should correspond to the Tenant
  Organization [Routed network][routed_network].
* `tenant_routed_network_dns`: DNS Server for the Tenant Organization [Routed network][routed_network]. It can be left blank if it's not needed.

If you wish to have a different networking setup, please modify the [step 2 configuration][step2].

#### CSE Server

There is also a set of resources created by the [step 2 configuration][step2] that correspond to the CSE Server vApp.
The generated VM makes use of the uploaded CSE OVA and some required guest properties:

* `cse_admin_username`: This must be the same CSE Administrator user created in the first step.
* `cse_admin_password`: This must be the same CSE Administrator user's password created in the first step.
* `cse_admin_api_token_file`: This specifies the path where the API token is saved and consumed.

#### UI plugin installation

-> If the old CSE 3.x UI plugin is installed, you will need to remove it before installing the new one.

The final resource created by the [step 2 configuration][step2] is the [`vcd_ui_plugin`][ui_plugin] resource.

This resource is **optional**, it will be only created if the variable `k8s_container_clusters_ui_plugin_path` is not empty,
so you can leverage whether your tenant users or system administrators will need it or not. It can be useful for troubleshooting,
or if your tenant users are not familiar with Terraform, they will be still able to create and manage their clusters
with the UI.

If you decide to install it, `k8s_container_clusters_ui_plugin_path` should point to the
[Kubernetes Container Clusters UI plug-in 4.2][cse_docs] ZIP file that you can download in the [CSE documentation][cse_docs].

### Final considerations

#### Verifying the CSE Server

To validate that the CSE Server is working correctly, you can either do it programmatically with a [DNAT rule][nat_rule] that maps
one available IP to the CSE Server, or using the UI:

* With a [DNAT rule][nat_rule] you would be able to connect to the CSE Server VM through `ssh` and the credentials that are stored in the `terraform.tfstate` file,
  with a resource similar to this:
```
resource "vcd_nsxt_nat_rule" "solutions_nat" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id

  name        = "CSE Server DNAT rule"
  rule_type   = "DNAT"
  description = "CSE Server DNAT rule"

  external_address = "One available IP from Solutions Provider Gateway"
  internal_address = "CSE Server IP"
  logging          = true
}
```

* Using the UI, you can go to the CSE Server VM and open a **web console**. The credentials to login are shown in _Guest customization properties_ > _Edit_.

Once you gain access to the CSE Server, you can check the `cse.log` file, the configuration file or check Internet connectivity.
If something does not work, please check the **Troubleshooting** section below.

## Troubleshooting

To evaluate the correctness of the setup, you can check the _"Verifying that the setup works"_ section above.

-> You can visit [the CSE documentation][cse_docs] to learn how to monitor the logs and troubleshoot possible problems.

The most common issues are:

* Lack of Internet connectivity in CSE Server:
  * Verify that the IPs specified in your Provider Gateways are correct.
  * Verify that the IPs specified in your Edge Gateways are correct.
  * Verify that your Firewall setup is not blocking outbound connectivity.
  * Verify that the Routed network has the DNS correctly set and working.

* OVA upload is taking too long:
  * Verify your Internet connectivity is not having any issues.
  * OVAs are quite big, you could tune `upload_piece_size` to speed up the upload process.
  * If upload fails, or you need to re-upload it, you can do a `terraform apply -replace=vcd_catalog_vapp_template.cse_ova`.
  * Verify that there's not a huge latency between your VCD and the place where Terraform configuration is run.

* Cluster creation is failing:
  * Please visit the [CSE documentation][cse_docs] to learn how to monitor the logs and troubleshoot possible problems.

## Upgrade from CSE 4.1 to 4.2.0

In this section you can find the required steps to update from CSE 4.1 to 4.2.0.

~> This section assumes that the old CSE 4.1 installation was done with Terraform by following the 4.1 guide steps.
Also, you need to meet [the pre-requisites criteria](#pre-requisites).

### Create the new RDE elements

Create a new version of the [RDE Types][rde_type] that were used in 4.1. This will allow them to co-exist with the old ones,
so we can perform a smooth upgrade.

```hcl
resource "vcd_rde_type" "capvcdcluster_type_v130" {
  # Same attributes as 4.1, except for:
  version = "1.3.0" # New version
  # New schema:
  schema_url = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension/v4.2/schemas/capvcd-type-schema-v1.3.0.json"
  # Behaviors need to be created before any RDE Type
  depends_on = [vcd_rde_interface_behavior.capvcd_behavior]
}
```

## Upgrade from CSE 4.2.0 to 4.2.1

In this section you can find the required steps to update from CSE 4.2.0 to 4.2.1.

Change the `VCDKEConfig` [RDE][rde] to update the `capvcd_version`, `cpi_version` and `csi_version` (follow [the instructions
in the section below](#upgrade-the-vcdkeconfig-rde-cse-server-configuration) to know how to upgrade this configuration):

```hcl
resource "vcd_rde" "vcdkeconfig_instance" {
  # ...omitted
  input_entity = templatefile(var.vcdkeconfig_template_filepath, {
    # ...omitted
    capvcd_version = "1.3.0" # It was 1.2.0 in 4.2.0
    cpi_version    = "1.6.0" # It was 1.5.0 in 4.2.0
    csi_version    = "1.6.0" # It was 1.5.0 in 4.2.0
  })
}
```

The Kubernetes Clusters Right bundle and Kubernetes Cluster Author role need to have the right to view and manage IP Spaces:

```hcl
resource "vcd_role" "cse_admin_role" {
  name = "CSE Admin Role"
  # ...omitted
  rights = [
    "API Tokens: Manage",
    # ...omitted
    "IP Spaces: Allocate",
    "Private IP Spaces: View",
    "Private IP Spaces: Manage",
  ]
}

resource "vcd_rights_bundle" "k8s_clusters_rights_bundle" {
  name = "Kubernetes Clusters Rights Bundle"
  # ...omitted
  rights = [
    "API Tokens: Manage",
    # ...omitted
    "IP Spaces: Allocate",
    "Private IP Spaces: View",
    "Private IP Spaces: Manage",
  ]
}

resource "vcd_global_role" "k8s_cluster_author" {
  name = "Kubernetes Cluster Author"
  # ...omitted
  rights = [
    "API Tokens: Manage",
    # ...omitted
    "IP Spaces: Allocate",
    "Private IP Spaces: View",
    "Private IP Spaces: Manage",
  ]
}
```

### Upgrade the VCDKEConfig RDE (CSE Server configuration)

With the new [RDE Types][rde_type] in place, you need to perform an upgrade of the existing `VCDKEConfig` [RDE][rde], which
stores the CSE Server configuration. By using the v3.12.0 of the VCD Terraform Provider, you can do this update without forcing
a replacement:

```hcl
resource "vcd_rde" "vcdkeconfig_instance" {
  # Same values as before, except:
  input_entity = templatefile(var.vcdkeconfig_template_filepath, {
    # Same values as before, except:
    rde_projector_version = "0.7.0"
  })
}
```

You can find the meaning of these values in the section ["RDE (CSE Server configuration / VCDKEConfig)"](#rde-cse-server-configuration--vcdkeconfig).
Please notice that you need to upgrade the CAPVCD, CPI and CSI versions. The new values are stated in the same section.

### Upload the new CSE 4.2 OVA

You need to upload the new CSE 4.2 OVA to the `cse_catalog` that already hosts the CSE 4.1 one.
To download the required OVAs, please refer to the [CSE documentation][cse_docs].

```hcl
resource "vcd_catalog_vapp_template" "cse_ova_4_2" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization that already exists from 4.1
  catalog_id = vcd_catalog.cse_catalog.id          # References the CSE Catalog that already exists from 4.1

  name        = "VMware_Cloud_Director_Container_Service_Extension-4.2.0"
  description = "VMware_Cloud_Director_Container_Service_Extension-4.2.0"
  ova_path    = "VMware_Cloud_Director_Container_Service_Extension-4.2.0.ova"
}
```

### Update CSE Server

To update the CSE Server, just change the referenced OVA:

```hcl
resource "vcd_vapp_vm" "cse_server_vm" {
  # All values remain the same, except:
  vapp_template_id = vcd_catalog_vapp_template.cse_ova_4_2.id # Reference the 4.2 OVA
}
```

This will re-deploy the VM with the new CSE 4.2 Server.

## Update CSE Server Configuration

To make changes to the existing server configuration, you should be able to locate the [`vcd_rde`][rde] resource named `vcdkeconfig_instance`
in the [step 2 configuration][step2] that was created during the installation process. To update its configuration, you can
**change the variable values that are referenced**. For this, you can review the **"CSE Server"** section in the Installation process to
see how this can be done.

After variables are changed, the CSE Server VM needs to be rebooted. You can trigger a reboot changing the CSE Server configuration
as follows:

```hcl
resource "vcd_vapp_vm" "cse_server_vm" {
  # ...
  power_on = false # Trigger a power off
  # ...
}
```

Then change to power on again:

```hcl
resource "vcd_vapp_vm" "cse_server_vm" {
  # ...
  power_on = true # Trigger a power on
  # ...
}
```

This must be done as a 2-step operation.

### Patch version upgrade

To upgrade the CSE Server appliance, first you need to upload a new CSE Server OVA to the CSE catalog and then replace
the reference to the [vApp Template][catalog_vapp_template] in the CSE Server VM.

In the [step 2 configuration][step2], you can find the `cse_ova` [vApp Template][catalog_vapp_template] and the 
`cse_server_vm` [VM][vm] that were applied during the installation process.
Then you can create a new `vcd_catalog_vapp_template` and modify `cse_server_vm` to reference it:

```hcl
# Uploads a new CSE Server OVA. In the example below, we upload version 4.2.1
resource "vcd_catalog_vapp_template" "new_cse_ova" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization
  catalog_id = vcd_catalog.cse_catalog.id          # References the CSE Catalog

  name        = "VMware_Cloud_Director_Container_Service_Extension-4.2.1"
  description = "VMware_Cloud_Director_Container_Service_Extension-4.2.1"
  ova_path    = "/home/bob/cse/VMware_Cloud_Director_Container_Service_Extension-4.2.1.ova"
}

# ...

# Update the vApp Template reference to update the CSE Server
resource "vcd_vapp_vm" "cse_server_vm" {
  # ...
  vapp_template_id = vcd_catalog_vapp_template.new_cse_ova.id # Change to the new OVA version
}
```

## Working with Kubernetes clusters

Please read the specific guide on that topic [here][cse_cluster_management_guide].

## Uninstall CSE

~> Before uninstalling CSE, make sure you mark all clusters for deletion.

Once all clusters are removed in the background by CSE Server, you may destroy the remaining infrastructure with Terraform command.

[alb]: /providers/vmware/vcd/latest/docs/guides/nsxt_alb
[api_token]: /providers/vmware/vcd/latest/docs/resources/api_token
[catalog]: /providers/vmware/vcd/latest/docs/resources/catalog
[catalog_vapp_template_ds]: /providers/vmware/vcd/latest/docs/data-sources/catalog_vapp_template
[cse_cluster_management_guide]: /providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_cluster_management
[cse_docs]: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
[edge_cluster]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_edge_cluster
[edge_gateway]: /providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway
[global_role]: /providers/vmware/vcd/latest/docs/resources/global_role
[nat_rule]: /providers/vmware/vcd/latest/docs/resources/nsxt_nat_rule
[nsxt_manager]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_manager
[nsxt_tier0_router]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_tier0_router
[org]: /providers/vmware/vcd/latest/docs/resources/org
[org_d]: /providers/vmware/vcd/latest/docs/data-sources/org
[product_matrix]: https://interopmatrix.vmware.com/Interoperability?col=659,&row=0
[provider_gateway]: /providers/vmware/vcd/latest/docs/resources/external_network_v2
[provider_vdc]: /providers/vmware/vcd/latest/docs/data-sources/provider_vdc
[rights_bundle]: /providers/vmware/vcd/latest/docs/resources/rights_bundle
[rde]: /providers/vmware/vcd/latest/docs/resources/rde
[rde_interface]: /providers/vmware/vcd/latest/docs/resources/rde_interface
[rde_type]: /providers/vmware/vcd/latest/docs/resources/rde_type
[rde_interface_behavior]: /providers/vmware/vcd/latest/docs/resources/rde_interface_behavior
[rde_type_behavior_acl]: /providers/vmware/vcd/latest/docs/resources/rde_type_behavior_acl
[role]: /providers/vmware/vcd/latest/docs/resources/role
[routed_network]: /providers/vmware/vcd/latest/docs/resources/network_routed_v2
[sizing]: /providers/vmware/vcd/latest/docs/resources/vm_sizing_policy
[step1]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.2/install/step1
[step2]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.2/install/step2
[tkgm_docs]: https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html
[user]: /providers/vmware/vcd/latest/docs/resources/org_user
[ui_plugin]: /providers/vmware/vcd/latest/docs/resources/ui_plugin
[catalog_vapp_template]: /providers/vmware/vcd/latest/docs/resources/catalog_vapp_template
[vdc]: /providers/vmware/vcd/latest/docs/resources/org_vdc
[vm]: /providers/vmware/vcd/latest/docs/resources/vapp_vm
