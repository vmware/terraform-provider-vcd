---
layout: "vcd"
page_title: "VMware Cloud Director: nsxt alb"
sidebar_current: "docs-vcd-guides-nsxt-alb"
description: |-
  Provides guidance to VMware NSX Advanced Load Balancer (Avi)
---

# VMware NSX Advanced Load Balancer (Avi)

## About 

Starting with version 10.2, VMware Cloud Director provides load balancing services by leveraging the capabilities of
VMware NSX Advanced Load Balancer. _System administrators_ can enable and configure access to load balancing services
for VDCs backed by NSX-T.

Load balancing services are associated with NSX-T Edge Gateways, which can be scoped either to an organization VDC
backed by NSX-T VDC or to a VDC group with NSX-T Data Center network provider type.

To use the virtual infrastructure provided by NSX-T Advanced Load Balancer, register your NSX-T Cloud instances with
VMware Cloud Director. Controllers serve as a central control plane for load balancing services. After registering
controllers, one can manage them directly from VMware Cloud Director.

The load balancing compute infrastructure provided by NSX Advanced Load Balancer is organized into Service Engine
Groups. Multiple Service Engine Groups can be assigned to a single NSX-T Edge Gateway.

A Service Engine Group has a unique set of compute characteristics that are defined upon creation.

## Requirements

* NSX Advanced Load Balancer is supported starting VCD versions *10.2+*.
* NSX Advanced Load Balancer configured with NSX-T, see [Avi Integration with NSX-T](https://avinetworks.com/docs/20.1/avi-nsx-t-integration/).
* Provider operations supported in Terraform provider VCD *v3.4+*. 
* Tenant operations supported in Terraform provider VCD *3.5+*. 

-> In this document, when we mention **tenants**, the term can be substituted with **organizations**.

## Resource and data source overview

The following list of resources and matching data sources exists to perform NSX-T ALB infrastructure
setup for providers:

* [vcd_nsxt_alb_controller](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_controller)
* [vcd_nsxt_alb_cloud](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_cloud)
* [vcd_nsxt_alb_service_engine_group](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_service_engine_group)

Additionally, there is a data source only to help lookup ALB Importable Clouds helping to populate 
[vcd_nsxt_alb_cloud](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_cloud):

* [vcd_nsxt_alb_importable_cloud](/providers/vmware/vcd/latest/docs/data-sources/nsxt_alb_importable_cloud)

Above resources and data sources cover infrastructure setup for providers. The next two resources 
still *require provider rights*, but help to enable ALB for tenants on particular NSX-T Edge Gateway:

* [vcd_nsxt_alb_settings](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_general_settings)
* [vcd_nsxt_alb_edgegateway_service_engine_group](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)


Finally, the remaining two resources help tenants to manage their ALB configurations:

* [vcd_nsxt_alb_pool](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_pool)
* [vcd_nsxt_alb_virtual_service](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service)

-> Examples below demonstrate a working setup, but do not cover all capabilities. More information about capabilities of
each resource are outlined in their own documentation pages.

## Infrastructure Setup example (Provider)

-> All operations in this part require Provider access.

The following snippet will do the following:

* Register NSX-T ALB Controller using
  [vcd_nsxt_alb_controller](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_controller) resource
* Look up available Clouds to import using
  [vcd_nsxt_alb_importable_cloud](/providers/vmware/vcd/latest/docs/data-sources/nsxt_alb_importable_cloud) data source
* Define NSX-T ALB Cloud in VCD using [vcd_nsxt_alb_cloud](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_cloud)
  resource
* Define a Service Engine Group
  [vcd_nsxt_alb_service_engine_group](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_service_engine_group) which
  can later be assigned to tenant Edge Gateways


```hcl
# Local variable is used to avoid direct reference and 
# cover Terraform core bug https://github.com/hashicorp/terraform/issues/29484
# Even changing NSX-T ALB Controller name in UI, plan will 
# cause to recreate all resources depending on 
# vcd_nsxt_alb_importable_cloud data source if
# this indirect reference (via local) variable is not used.
locals {
  controller_id = vcd_nsxt_alb_controller.main.id
}

# Configuration of NSX-T ALB Controller
resource "vcd_nsxt_alb_controller" "main" {
  name         = "alb-controller-1"
  description  = "my first alb controller configured via Terraform"
  url          = "https://alb-controller-url.example"
  username     = "admin"
  password     = "MY-SECRET-PASSWORD"
  license_type = "ENTERPRISE"
}

# Lookup of ALB Importable Cloud (to be referenced in vcd_nsxt_alb_cloud resource)
data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "my-importable-cloud-name"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.main.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "first-se-updated"
  description                          = "test-description"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "SHARED"
}
```

## NSX-T Edge Gateway configuration setup example (Provider)

-> All operations in this part require Provider access.

The following snippet will do two things on a specified NSX-T Edge Gateway:

* Enable ALB for that Edge Gateway
* Assign Service Engine Group to it 

````hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

resource "vcd_nsxt_alb_settings" "main" {
  org = "my-org"
  vdc = "nsxt-vdc"

  # Reference to Edge Gateway
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org = "my-org"
  vdc = "nsxt-vdc"

  edge_gateway_id         = vcd_nsxt_alb_settings.main.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
````

And that completes the work required for Providers.

## Pool and Virtual Service configuration NSX-T Edge Gateway configuration (Tenant)

This part demonstrates how Tenant can handle Pools and Virtual Services once providers have done
their part to enable ALB on NSX-T Edge Gateways. It will:

* Look up existing NSX-T Edge Gateway using
  [vcd_nsxt_edgegateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway) data source
* Look up Service Engine Groups that are available for this NSX-T Edge Gateway using
  [vcd_nsxt_alb_edgegateway_service_engine_group](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)
  data source
* Set up an ALB Pool with 3 members using [vcd_nsxt_alb_pool](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_pool)
  resource
* Expose a Virtual Service using
  [vcd_nsxt_alb_virtual_service](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service) resource which
  combines all the data

```hcl
data "vcd_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw-dainius"
}

data "vcd_nsxt_alb_edgegateway_service_engine_group" "assigned" {
  org = "my-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  # This name comes from prerequisite setup (can be looked up in the UI by tenants)
  service_engine_group_name = "assigned-service-engine-group-name"
}

resource "vcd_nsxt_alb_pool" "test" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name            = "first-pool"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  algorithm               = "LEAST_LOAD"
  default_port            = "9000"
  graceful_timeout_period = "0" # Immediately removes member from pool once disabled

  member {
    ip_address = "192.168.1.1"
  }

  member {
    enabled    = false
    ip_address = "192.168.1.7"
    ratio      = 3
    port       = 7000
  }

  member {
    ip_address = "192.168.1.8"
    ratio      = 1
    port       = 6000
  }
}

resource "vcd_nsxt_alb_virtual_service" "test" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name            = "first-virtual-service"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  pool_id                  = vcd_nsxt_alb_pool.test.id
  service_engine_group_id  = data.vcd_nsxt_alb_edgegateway_service_engine_group.assigned.service_engine_group_id
  virtual_ip_address       = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }
}
```

## References

* [VMware Cloud Director Documentation for Providers](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-1D3014BC-4792-40E8-99E1-A8F0FFC691FE.html)
* [VMware Cloud Director Documentation for Tenants](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-789FCC6A-EE14-4CAA-AB91-08841513B328.html)
* [VMware blog post introducing NSX-T ALB](https://blogs.vmware.com/cloudprovider/2020/11/embrace-next-gen-networking-security-with-nsx-t-and-vmware-cloud-director-10-2.html)
* [Feature Fridays video - NSX-T Advanced Load balancer](https://blogs.vmware.com/cloudprovider/2020/10/feature-fridays-episode-19-nsx-t-advanced-load-balancer.html)
* [Avi Integration with NSX-T](https://avinetworks.com/docs/20.1/avi-nsx-t-integration/)
