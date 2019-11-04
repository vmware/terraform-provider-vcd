---
layout: "vcd"
page_title: "Provider: VMware vCloudDirector"
sidebar_current: "docs-vcd-index"
description: |-
  The VMware vCloud Director provider is used to interact with the resources supported by VMware vCloud Director. The provider needs to be configured with the proper credentials before it can be used.
---

# VMware vCloud Director Provider 2.6

The VMware vCloud Director provider is used to interact with the resources supported by VMware vCloud Director. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

~> **NOTE:** The VMware vCloud Director Provider went through a refresh at the beginning of 2019 and some semantic changes were made compared to the previously available initial version. Please check docs for *v2.0+*, *v2.1+*, *v2.2+*, *v2.4+* labels and your existing .tf configuration files carefully when shifting to this new version. 

## Supported vCD Versions

The following vCloud Director versions are supported by this provider:

* 9.0
* 9.1
* 9.5
* 9.7
* 10.0

## Example Usage

### Connecting as Org Admin

The most common - tenant - use case when you set user to organization administrator and when all resources are in a single organization. 

```hcl
# Configure the VMware vCloud Director Provider
provider "vcd" {
  user                 = "${var.vcd_user}"
  password             = "${var.vcd_pass}"
  org                  = "${var.vcd_org}"
  vdc                  = "${var.vcd_vdc}"
  url                  = "${var.vcd_url}"
  max_retry_timeout    = "${var.vcd_max_retry_timeout}"
  allow_unverified_ssl = "${var.vcd_allow_unverified_ssl}"
}

# Create a new network in organization and VDC defined above
resource "vcd_network_routed" "net" {
  # ...
}
```

### Connecting as Sys Admin

When you want to manage resources across different organizations from a single configuration.

```hcl
# Configure the VMware vCloud Director Provider
provider "vcd" {
  user                 = "administrator"
  password             = "${var.vcd_pass}"
  org                  = "System"
  url                  = "${var.vcd_url}"
  max_retry_timeout    = "${var.vcd_max_retry_timeout}"
  allow_unverified_ssl = "${var.vcd_allow_unverified_ssl}"
}

# Create a new network in some organization and VDC
resource "vcd_network_routed" "net1" {
  org = "Org1"
  vdc = "Org1VDC"

  # ...
}

# Create a new network in a different organization and VDC
resource "vcd_network_routed" "net2" {
  org = "Org2"
  vdc = "Org2VDC"

  # ...
}
```

### Connecting as Sys Admin with Default Org and VDC

When you want to manage resources across different organizations but set a default one. 

```hcl
# Configure the VMware vCloud Director Provider
provider "vcd" {
  user                 = "administrator"
  password             = "${var.vcd_pass}"
  sysorg               = "System"
  org                  = "${var.vcd_org}"                  # Default for resources
  vdc                  = "${var.vcd_vdc}"                  # Default for resources
  url                  = "${var.vcd_url}"
  max_retry_timeout    = "${var.vcd_max_retry_timeout}"
  allow_unverified_ssl = "${var.vcd_allow_unverified_ssl}"
}

# Create a new network in the default organization and VDC
resource "vcd_network_routed" "net1" {
  # ...
}

# Create a new network in a specific organization and VDC
resource "vcd_network_routed" "net2" {
  org = "OrgZ"
  vdc = "OrgZVDC"

  # ...
}
```

## Argument Reference

The following arguments are used to configure the VMware vCloud Director Provider:

* `user` - (Required) This is the username for vCloud Director API operations. Can also
  be specified with the `VCD_USER` environment variable.  
  *v2.0+* `user` may be "administrator" (set `org` or `sysorg` to "System" in this case).
  
* `password` - (Required) This is the password for vCloud Director API operations. Can
  also be specified with the `VCD_PASSWORD` environment variable.
  
* `org` - (Required) This is the vCloud Director Org on which to run API
  operations. Can also be specified with the `VCD_ORG` environment
  variable.  
  *v2.0+* `org` may be set to "System" when connection as Sys Admin is desired
  (set `user` to "administrator" in this case).  
  Note: `org` value is case sensitive.
  
* `sysorg` - (Optional; *v2.0+*) - Organization for user authentication. Can also be
   specified with the `VCD_SYS_ORG` environment variable. Set `sysorg` to "System" and
   `user` to "administrator" to free up `org` argument for setting a default organization
   for resources to use.
   
* `url` - (Required) This is the URL for the vCloud Director API endpoint. e.g.
  https://server.domain.com/api. Can also be specified with the `VCD_URL` environment variable.
  
* `vdc` - (Optional) This is the virtual datacenter within vCloud Director to run
  API operations against. If not set the plugin will select the first virtual
  datacenter available to your Org. Can also be specified with the `VCD_VDC` environment
  variable.
  
* `max_retry_timeout` - (Optional) This provides you with the ability to specify the maximum
  amount of time (in seconds) you are prepared to wait for interactions on resources managed
  by vCloud Director to be successful. If a resource action fails, the action will be retried
  (as long as it is still within the `max_retry_timeout` value) to try and ensure success.
  Defaults to 60 seconds if not set.
  Can also be specified with the `VCD_MAX_RETRY_TIMEOUT` environment variable.
  
* `maxRetryTimeout` - (Deprecated) Use `max_retry_timeout` instead.

* `allow_unverified_ssl` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your auth token. If omitted, default
  value is false. Can also be specified with the
  `VCD_ALLOW_UNVERIFIED_SSL` environment variable.

* `logging` - (Optional; *v2.0+*) Boolean that enables API calls logging from upstream library `go-vcloud-director`. 
   The logging file will record all API requests and responses, plus some debug information that is part of this 
   provider. Logging can also be activated using the `VCD_API_LOGGING` environment variable.

* `logging_file` - (Optional; *v2.0+*) The name of the log file (when `logging` is enabled). By default is 
  `go-vcloud-director` and it can also be changed using the `VCD_API_LOGGING_FILE` environment variable.
  
* `import_separator` - (Optional; *v2.5+*) The string to be used as separator with `terraform import`. By default
  it is a dot (`.`).

## Connection Cache (*2.0+*)

vCloud Director connection calls can be expensive, and if a definition file contains several resources, it may trigger 
multiple connections. There is a cache engine, disabled by default, which can be activated by the `VCD_CACHE` 
environment variable. When enabled, the provider will not reconnect, but reuse an active connection for up to 20 
minutes, and then connect again.
