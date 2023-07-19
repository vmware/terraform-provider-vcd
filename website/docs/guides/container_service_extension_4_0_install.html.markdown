---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension v4.0 installation"
sidebar_current: "docs-vcd-guides-cse-4-0-install"
description: |-
  Provides guidance on configuring VCD to be able to install and use Container Service Extension v4.0
---

# Container Service Extension v4.0 installation

## About

This guide describes the required steps to configure VCD to install the Container Service Extension (CSE) v4.0, that
will allow tenant users to deploy **Tanzu Kubernetes Grid Multi-cloud (TKGm)** clusters on VCD using Terraform or the UI.

To know more about CSE v4.0, you can visit [the documentation][cse_docs].

## Pre-requisites

-> Please read also the pre-requisites section in the [CSE documentation][cse_docs].

In order to complete the steps described in this guide, please be aware:

* CSE v4.0 is supported from VCD v10.4.0 or above, make sure your VCD appliance matches the criteria.
* Terraform provider needs to be v3.10.0 or above.
* Both CSE Server and the Bootstrap clusters require outbound Internet connectivity.
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

- The required `VCDKEConfig` [RDE Interface][rde_interface] and [RDE Type][rde_type]. These two resources specify the schema of the CSE Server
  configuration (called "VCDKEConfig") that will be instantiated in next step with a [RDE][rde].
- The required `capvcdCluster` [RDE Type][rde_type]. Its version is specified by the `capvcd_rde_version` variable, that **must be "1.1.0" for CSE v4.0**.
  This resource specifies the schema of the [TKGm clusters][tkgm_docs].
- The **CSE Admin [Role][role]**, that specifies the required rights for the CSE Administrator to manage provider-side elements of VCD.
- The **CSE Administrator [User][user]** that will administrate the CSE Server and other aspects of VCD that are directly related to CSE.
  Feel free to add more attributes like `description` or `full_name` if needed.

Once reviewed and applied with `terraform apply`, one **must login with the created CSE Administrator user to
generate an API token** that will be used in the next step. In UI, the API tokens can be generated in the CSE Administrator
user preferences in the top right, then go to the API tokens section, add a new one.
Or you can visit `/provider/administration/settings/user-preferences` at your VCD URL as CSE Administrator.

### Step 2: Install CSE

-> This step of the installation refers to the Terraform configuration present [here][step2].

~> Be sure that previous step is successfully completed and the API token for the CSE Administrator user was created.

This step will create all the remaining elements to install CSE v4.0 in VCD. You can read subsequent sections
to have a better understanding of the building blocks that are described in the [proposed Terraform configuration][step2].

In this [configuration][step2] you can also find a file named `terraform.tfvars.example`, you need to rename it to `terraform.tfvars`
and change the values present there to the correct ones. You can also modify the proposed resources so they fit better to your needs.

#### Organizations

The [proposed configuration][step2] will create two new [Organizations][org], as specified in the [CSE documentation][cse_docs]:

- A Solutions [Organization][org], which will host all provider-scoped items, such as the CSE Server.
  It should only be accessible to the CSE Administrator and Providers.
- A Tenant [Organization][org], which will host the [TKGm clusters][tkgm_docs] for the users of this tenant to consume them.

-> If you already have these two [Organizations][org] created and you want to use them instead,
you can leverage customising the [proposed configuration][step2] to use the Organization [data source][org_d] to fetch them.

#### VM Sizing Policies

The [proposed configuration][step2] will create four VM Sizing Policies:

- `TKG extra-large`: 8 CPU, 32GB memory.
- `TKG large`: 4 CPU, 16GB memory.
- `TKG medium`: 2 CPU, 8GB memory.
- `TKG small`: 2 CPU, 4GB memory.

These VM Sizing Policies should be applied as they are, so nothing should be changed here as these are the exact same
VM Sizing Policies created during CSE installation in UI. They will be assigned to the Tenant
Organization's VDC to be able to dimension the created [TKGm clusters][tkgm_docs] (see section below).

#### VDCs

The [proposed configuration][step2] will create two [VDCs][vdc], one for the Solutions Organization and another one for the Tenant Organization.

You need to specify the following values in `terraform.tfvars`:

- `provider_vdc_name`: This is used to fetch an existing [Provider VDC][provider_vdc], that will be used to create the two VDCs.
  If you are going to use more than one [Provider VDC][provider_vdc], please consider modifying the proposed configuration.
  In UI, [Provider VDCs][provider_vdc] can be found in the Provider view, inside _Cloud Resources_ menu.
- `nsxt_edge_cluster_name`: This is used to fetch an existing [Edge Cluster][edge_cluster], that will be used to create the two VDCs.
  If you are going to use more than one [Edge Cluster][edge_cluster], please consider modifying the proposed configuration.
  In UI, [Edge Clusters][edge_cluster] can be found in the NSX-T manager web UI.
- `network_pool_name`: This references an existing Network Pool, which is used to create both VDCs.
  If you are going to use more than one Network Pool, please consider modifying the proposed configuration.

In the [proposed configuration][step2] the Tenant Organization's VDC has all the required VM Sizing Policies assigned, with the `TKG small` being the default one.
You can customise it to make any other TKG policy the default one.

You can also leverage changing the storage profiles and other parameters to fit the requirements of your organization. Also,
if you already have usable [VDCs][vdc], you can change the configuration to fetch them instead.

#### Catalog and OVAs

The [proposed configuration][step2] will create two catalogs:

- A catalog to host CSE Server OVA files, only accessible to CSE Administrators. This catalog will allow CSE Administrators to organise and manage
  all the CSE Server OVAs that are required to run and upgrade the CSE Server.
- A catalog to host TKGm OVA files, only accessible to CSE Administrators but shared as read-only to tenants, that can use them to create [TKGm clusters][tkgm_docs].

Then it will upload the required OVAs to them. The OVAs can be specified in `terraform.tfvars`:

- `tkgm_ova_folder`: This will reference the path to the TKGm OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
- `tkgm_ova_file`: This will reference the file name of the TKGm OVA, like `ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933.ova`.
- `cse_ova_folder`: This will reference the path to the CSE OVA, as an absolute or relative path. It should **not** end with a trailing `/`.
- `cse_ova_file`: This will reference the file name of the CSE OVA, like `VMware_Cloud_Director_Container_Service_Extension-4.0.1.ova`.

-> To download the required OVAs, please refer to the [CSE documentation][cse_docs].

~> Both CSE Server and TKGm OVAs are heavy. Please take into account that the upload process could take more than 30 minutes, depending
on upload speed. You can tune the `upload_piece_size` to speed up the upload. Another option would be uploading them manually in the UI.
In case you're using a pre-uploaded OVA, leverage the [vcd_catalog_vapp_template][catalog_vapp_template_ds] data source (instead of the resource).

If you need to upload more than one OVA, please modify the [proposed configuration][step2].

### "Kubernetes Cluster Author" global role

Apart from the role to manage the CSE Server created in [step 1][step1], we also need a [Global Role][global_role]
for the [TKGm clusters][tkgm_docs] consumers (it would be similar to the concept of "vApp Author" but for [TKGm clusters][tkgm_docs]).

In order to create this [Global Role][global_role], the [proposed configuration][step2] first
creates a new [Rights Bundle][rights_bundle] and publishes it to all the tenants, then creates the [Global Role][global_role].

### Networking

The [proposed configuration][step2] prepares a basic networking layout that will make CSE v4.0 work. However, it is
recommended that you review the code and adapt the different parts to your needs, specially for the resources like `vcd_nsxt_firewall`.

The configuration will create the following:

- A [Provider Gateway][provider_gateway] per Organization. You can learn more about Provider Gateways [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-E6BAC24B-9628-495A-BA67-6DE6C5CF70F2.html).
  In this configuration we just expose some static IPs to the two Organizations, so they can consume them.
- An [Edge Gateway][edge_gateway] per Organization. You can learn more about Edge Gateways [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-45C0FEDF-84F2-4487-8DB8-3BC281EB25CD.html).
  In this configuration we create two that act as a router for each Organization that we created.
- Configure ALB with a shared Service Engine Group. You can learn more about Advanced Load Balancers [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-92A0563D-A272-4958-B732-9C35901D9DB8.html).
  In this setup, we provide a virtual service pool that CSE Server uses to provide load balancing capabilities to the [TKGm clusters][tkgm_docs].
- A [Routed network][routed_network] per Organization. You can learn more about Routed networks [here](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-74C4D27F-9E2A-4EB2-BBE1-CDD45C80E270.html).
  In this setup, we just provide a routed network per organization, so the CSE Server is inside its own network, isolated from the [TKGm clusters][tkgm_docs] network.
- Two [SNAT rules][nat_rule] that will allow outbound access. Feel free to adjust or replace these rules with other ways of providing outbound access.

~> SNAT rules is just a proposal to give the CSE Server and the clusters outbound access. Please review the [proposed configuration][step2]
first.

In order to create all the items listed above, the [proposed configuration][step2] asks for the following variables that you can customise in `terraform.tfvars`:

- `nsxt_manager_name`: It is the name of an existing [NSX-T Manager][nsxt_manager], which is needed in order to create the [Provider Gateways][provider_gateway].
  If you are going to use more than one [NSX-T Manager][nsxt_manager], please consider modifying the proposed configuration.
  In UI, [NSX-T Managers][nsxt_manager] can be found in the Provider view, inside _Infrastructure Resources > NSX-T_.
- `solutions_nsxt_tier0_router_name`: It is the name of an existing [Tier-0 Router][nsxt_tier0_router], which is needed in order to create the [Provider Gateway][provider_gateway] in the Solutions Organization.
  In UI, [Tier-0 Routers][nsxt_tier0_router] can be found in the NSX-T manager web UI.
- `solutions_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Solutions Organization.
- `solutions_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
- `solutions_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Solutions Organization.
  At least one IP is required. You can check the minimum amount of IPs required in [CSE documentation][cse_docs].
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
- `tenant_nsxt_tier0_router_name`: It is the name of an existing [Tier-0 Router][nsxt_tier0_router], which is needed in order to create the [Provider Gateway][provider_gateway] in the Tenant Organization.
  In UI, [Tier-0 Routers][nsxt_tier0_router] can be found in the NSX-T manager web UI.
- `tenant_provider_gateway_gateway_ip`: The gateway IP of the [Provider Gateway][provider_gateway] that will be used by the Tenant Organization.
- `tenant_provider_gateway_gateway_prefix_length`: Prefix length for the mentioned [Provider Gateway][provider_gateway].
- `tenant_provider_gateway_static_ip_ranges`: This is a list IP ranges that will be used by the [Provider Gateway][provider_gateway] that serves the Tenant Organization.
  At least one IP is required. You can check the minimum amount of IPs required in [CSE documentation][cse_docs].
  Each element of the list should be a 2-tuple like `[first IP, last IP]`. For example, a valid value
  for this attribute would be:
  ```
  solutions_provider_gateway_static_ip_ranges = [
    ["10.20.30.170", "10.20.30.170"], # A single IP ending in 170
    ["10.20.30.180", "10.20.30.182"], # A range of three IPs ending in 180,181,182
  ]
  ```
- `alb_controller_url`: URL of an existing ALB controller that will be created in VCD side. See the [ALB guide][alb] for more info.
- `alb_controller_username`: Username to access the ALB controller. See the [ALB guide][alb] for more info.
- `alb_controller_password`: Password of the username used to access the ALB controller. See the [ALB guide][alb] for more info.
- `alb_importable_cloud_name`: Name of the existing ALB Cloud defined in the ALB controller that will be imported to create an ALB Cloud in VCD. See the [ALB guide][alb] for more info.
- `solutions_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed_network] that will be created in the Solutions Organization.
- `solutions_routed_network_prefix_length`: The prefix length of the [Routed network][routed_network] that will be created in the Solutions Organization.
- `solutions_routed_network_ip_pool_start_address`: The [Routed network][routed_network] that will be created in the Solutions Organization will have a pool of usable IPs, this field
  defines the first usable IP.
- `solutions_routed_network_ip_pool_end_address`: The [Routed network][routed_network] that will be created in the Solutions Organization will have a pool of usable IPs, this field
  defines the end usable IP.
- `solutions_snat_external_ip`: This is used to create a SNAT rule on the Solutions Edge Gateway to provide Internet connectivity to the CSE Server. The external IP should be one available IP of the Solutions
  [Provider Gateway][provider_gateway].
- `solutions_snat_internal_network_cidr`: This is used to create a SNAT rule on the Solutions Edge Gateway to provide Internet connectivity to the CSE Server. The subnet should correspond to the Solutions
  Organization [Routed network][routed_network].
- `solutions_routed_network_dns`: DNS Server for the Solutions Organization [Routed network][routed_network]. It can be left blank if it's not needed.
- `tenant_routed_network_gateway_ip`: The gateway IP of the [Routed network][routed_network] that will be created in the Tenant Organization.
- `tenant_routed_network_prefix_length`: The prefix length of the [Routed network][routed_network] that will be created in the Tenant Organization.
- `tenant_routed_network_ip_pool_start_address`: The [Routed network][routed_network] that will be created in the Tenant Organization will have a pool of usable IPs, this field
  defines the first usable IP.
- `tenant_routed_network_ip_pool_end_address`: The [Routed network][routed_network] that will be created in the Tenant Organization will have a pool of usable IPs, this field
  defines the end usable IP.
- `tenant_snat_external_ip`: This is used to create a SNAT rule on the Tenant Edge Gateway to provide Internet connectivity to the clusters. The external IP should be one available IP of the Tenant
  [Provider Gateway][provider_gateway].
- `tenant_snat_internal_network_cidr`: This is used to create a SNAT rule on the Tenant Edge Gateway to provide Internet connectivity to the clusters. The subnet should correspond to the Tenant
  Organization [Routed network][routed_network].
- `tenant_routed_network_dns`: DNS Server for the Tenant Organization [Routed network][routed_network]. It can be left blank if it's not needed.

If you wish to have a different networking setup, please modify the [proposed configuration][step2].

### CSE Server

There is also a set of resources created by the [proposed configuration][step2] that correspond to the CSE Server vApp.
The generated VM makes use of the uploaded CSE OVA and some required guest properties.

In order to do so, the [configuration][step2] asks for the following variables that you can customise in `terraform.tfvars`:

- `vcdkeconfig_template_filepath` references a local file that defines the `VCDKEConfig` [RDE][rde] contents. It should be a JSON template, like
  [the one used in the configuration](https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/entities/vcdkeconfig-template.json).
  (Note: In `terraform.tfvars.example` the correct path is already provided).
- `capvcd_version`: The version for CAPVCD. It should be "1.0.0" for CSE v4.0.
- `capvcd_rde_version`: The version for the CAPVCD [RDE Type][rde_type]. It should be the same version used in Step 1.
- `cpi_version`: The version for CPI. It should be "1.2.0" for CSE v4.0.
- `csi_version`: The version for CSI. It should be "1.3.0" for CSE v4.0.
- `github_personal_access_token`: Create this one [here](https://github.com/settings/tokens),
  this will avoid installation errors caused by GitHub rate limiting, as the TKGm cluster creation process requires downloading
  some Kubernetes components from GitHub.
  The token should have the `public_repo` scope for classic tokens and `Public Repositories` for fine-grained tokens.
- `cse_admin_user`: This should reference the CSE Administrator [User][user] that was created in Step 1.
- `cse_admin_api_token`: This should be the API token that you created for the CSE Administrator after Step 1.

### UI plugin installation

The final resource created by the [proposed configuration][step2] is the [`vcd_ui_plugin`][ui_plugin] resource.

This resource is optional, it will be only created if the variable `k8s_container_clusters_ui_plugin_path` is not empty,
so you can leverage whether your tenant users or system administrators will need it or not. It can be useful for troubleshooting,
or if your tenant users are not familiar with Terraform, they will be still able to create and manage their clusters
with the UI.

If you decide to install it, `k8s_container_clusters_ui_plugin_path` should point to the
[Kubernetes Container Clusters UI plug-in v4.0][cse_docs] ZIP file that you can download in the [CSE documentation][cse_docs].

-> If the old CSE 3.x plugin is installed, you will need to remove it also.

### Final considerations

#### Verifying the CSE Server

To validate that the CSE Server is working correctly, you can either do it programmatically with a [DNAT rule][nat_rule] that maps
one available IP to the CSE Server, or using the UI:

- With a [DNAT rule][nat_rule] you would be able to connect to the CSE Server VM through `ssh` and the credentials that are stored in the `terraform.tfstate` file,
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

- Using the UI, you can go to the CSE Server VM and open a **web console**. The credentials to login are shown in _Guest customization properties_ > _Edit_.

Once you gain access to the CSE Server, you can check the `cse.log` file, the configuration file or check Internet connectivity.
If something does not work, please check the **Troubleshooting** section below.

#### Troubleshooting

To evaluate the correctness of the setup, you can check the _"Verifying that the setup works"_ section above.

-> You can visit [the CSE documentation][cse_docs] to learn how to monitor the logs and troubleshoot possible problems.

The most common issues are:

- Lack of Internet connectivity in CSE Server:
  - Verify that the IPs specified in your Provider Gateways are correct.
  - Verify that the IPs specified in your Edge Gateways are correct.
  - Verify that your Firewall setup is not blocking outbound connectivity.
  - Verify that the Routed network has the DNS correctly set and working.

- OVA upload is taking too long:
  - Verify your Internet connectivity is not having any issues.
  - OVAs are quite big, you could tune `upload_piece_size` to speed up the upload process.
  - If upload fails, or you need to re-upload it, you can do a `terraform apply -replace=vcd_catalog_vapp_template.cse_ova`.
  - Verify that there's not a huge latency between your VCD and the place where Terraform configuration is run.

- Cluster creation is failing:
  - Please visit the [CSE documentation][cse_docs] to learn how to monitor the logs and troubleshoot possible problems.

## Update CSE Server

### Update Configuration

To make changes to the existing server configuration, you should be able to locate the [`vcd_rde`][rde] resource named `vcdkeconfig_instance`
in the [proposed configuration][step2] that was created during the installation process. To update its configuration, you can
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

In the [proposed configuration][step2], you can find the `cse_ova` [vApp Template][catalog_vapp_template] and the 
`cse_server_vm` [VM][vm] that were applied during the installation process.
Then you can create a new `vcd_catalog_vapp_template` and modify `cse_server_vm` to reference it:

```hcl
# Uploads a new CSE Server OVA. In the example below, we upload version 4.0.2
resource "vcd_catalog_vapp_template" "new_cse_ova" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization
  catalog_id = vcd_catalog.cse_catalog.id          # References the CSE Catalog

  name        = "VMware_Cloud_Director_Container_Service_Extension-4.0.2"
  description = "VMware_Cloud_Director_Container_Service_Extension-4.0.2"
  ova_path    = "/home/bob/cse/VMware_Cloud_Director_Container_Service_Extension-4.0.2.ova"
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
[api_token]: https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-A1B3B2FA-7B2C-4EE1-9D1B-188BE703EEDE.html
[catalog]: /providers/vmware/vcd/latest/docs/resources/catalog
[catalog_vapp_template_ds]: /providers/vmware/vcd/latest/docs/data-sources/catalog_vapp_template
[cse_cluster_management_guide]: /providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0_cluster_management
[cse_docs]: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
[edge_cluster]: /providers/vmware/vcd/latest/docs/data-sources/nsxt_edge_cluster
[edge_gateway]: /providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway
[global_role]: /providers/vmware/vcd/latest/docs/resources/global_role
[nat_rule]: /providers/vmware/vcd/latest/docs/resources/nsxt_nat_rule
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
[tkgm_docs]: https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html
[user]: /providers/vmware/vcd/latest/docs/resources/org_user
[ui_plugin]: /providers/vmware/vcd/latest/docs/resources/ui_plugin
[catalog_vapp_template]: /providers/vmware/vcd/latest/docs/resources/catalog_vapp_template
[vdc]: /providers/vmware/vcd/latest/docs/resources/org_vdc
[vm]: /providers/vmware/vcd/latest/docs/resources/vapp_vm
