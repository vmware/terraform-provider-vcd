---
layout: "vcd"
page_title: "VMware Cloud Director: VM Guest Customization"
sidebar_current: "docs-vcd-guides-vm-guest-customization"
description: |-
  Provides guidance on VM Guest Customization.
---


# VM Guest Customization

Guest Customization allows users to supply custom configuration for new VMs which is inevitable in
cloud environments. Unfortunately there is no single mechanism for this task and one must pick one of
many available - [`VMware Guest
Customization`](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-BB682E4D-DCD7-4936-A665-0B0FBD6F0EB5.html),
[`cloud-init`](https://cloud-init.io/), [`Ignition`](https://coreos.github.io/ignition/),
[`Talos`](https://www.talos.dev/docs/v0.13/virtualized-platforms/vmware/#update-settings-for-the-worker-nodes),
[`Packer`](https://www.packer.io/), etc. To make matters more challenging each OS might require a different
approach. The goal of this page is to give pointers and make it easier to set up Guest Customization.

~> This page is expected to grow over time. If you have a good working example for this page, we
would happily accept a Pull Request with documentation how to use it.

## Guest Customization using Ignition files

The main point is that **Ignition** configuration can be supplied using
[`guest_properties`](/providers/vmware/vcd/latest/docs/resources/vapp_vm#guest_properties) key value
map in [`vcd_vapp_vm`](/providers/vmware/vcd/latest/docs/resources/vapp_vm) or
[`vcd_vm`](/providers/vmware/vcd/latest/docs/resources/vm) resources.

[Ignition documentation](https://docs.fedoraproject.org/en-US/fedora-coreos/provisioning-vmware/)
mentions two required fields for guest properties to enable customization:

* `guestinfo.ignition.config.data.encoding` - the encoding of the Ignition configuration. Other
  options are available, but `base64` is recommended.
* `guestinfo.ignition.config.data` - the content of the Ignition configuration, encoded according to
  the format above.


### Example using Ignition configuration

This example is based by using officially provided OVA image of **Fedora CoreOS** (FCOS)
(`34.20210919.3.0 stable` at the time of testing).

There is a lot more to read about producing Ignition files that can be read in [official
docs](https://docs.fedoraproject.org/en-US/fedora-coreos/producing-ign/), but in this case we will
just pick a quick example with minimal configuration in JSON.

This Ignition configuration will create user `core` with password `asdf123` (hashed using `mkpasswd` or other compatible
hash generator) and set hostname to `core1` in guest. Store these contents in `ignition.json`
```json 
{
    "ignition": {
        "version": "3.3.0"
    },
    "storage": {
        "files": [
            {
                "path": "/etc/hostname",
                "mode": 420,
                "overwrite": true,
                "contents": {
                    "source": "data:,core1"
                }
            }
        ]
    },
    "passwd": {
        "users": [
            {
                "name": "core",
                "passwordHash": "PNQHqA7xZADWs"
            }
        ]
    }
}
```
To supply it to Guest VM using Terraform provider VCD one must read contents and encode it.

```hcl
resource "vcd_vm" "customized" {
  name = "fedora-coreos-customized"

  catalog_name  = "my-catalog-name"
  template_name = "fedora-coreos"
  cpus          = 2
  memory        = 2048

  # power_on = false

  network {
    name               = "routed-net"
    type               = "org"
    ip_allocation_mode = "POOL"
  }

  guest_properties = {
    "guestinfo.ignition.config.data.encoding" = "base64"
    "guestinfo.ignition.config.data"          = base64encode(file("ignition.json"))
  }
}
```

After applying the configuration one should be able to login using `core` user and `asdf123`
password. In addition the hostname should be set to `core1. [Official Ignition
docs](https://docs.fedoraproject.org/en-US/fedora-coreos/producing-ign/) have a lot more
configuration options.

## Guest Customization using CloudInit (OVF datasource)

Configuration can be passed to CloudInit using
[`guest_properties`](/providers/vmware/vcd/latest/docs/resources/vapp_vm#guest_properties) key value
map in [`vcd_vapp_vm`](/providers/vmware/vcd/latest/docs/resources/vapp_vm) or
[`vcd_vm`](/providers/vmware/vcd/latest/docs/resources/vm) resources.

More about [CloudInit](https://cloudinit.readthedocs.io/en/latest/) and
[OVF](https://cloudinit.readthedocs.io/en/latest/topics/datasources/ovf.html)
[datasource](https://cloudinit.readthedocs.io/en/latest/topics/datasources.html)

### Example using CloudInit configuration (Ubuntu 22.04 LTS)

[Ubuntu 22.04 LTS OVA
Image](https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.ova)
from [Ubuntu Cloud images repository](https://cloud-images.ubuntu.com/) used.

Prepare a script to run using Cloud Init. In this example it is just a PoC (note the `sudo` command
to run as `root` as by default commands are run as `ubuntu` user).

Store it in `script.sh` (or change user-data references in HCL script). Shebang `#!/bin/sh` is a
required so that CloudInit recognizes it is a shell script. There are [other possible
formats](https://cloudinit.readthedocs.io/en/latest/topics/format.html#user-data-script) for
supplying CloudInit user data.

```sh
#!/bin/sh
echo "Command executed as 'ubuntu' user" >> /tmp/setup.log
sudo echo "Command executed as 'root' user" >> /tmp/setup.log
```

VM definition might be different, but key here is `guest_properties`. It will supply needed data for
CloudInit (including the base64 encoded shell script in file `script.sh` as required by CloudInit)

```hcl
resource "vcd_vapp_vm" "guest-vm" {
  name      = var.guest_hostname
  vapp_name = vcd_vapp.terminal.name

  catalog_name  = var.catalog_name
  template_name = var.template_name

  cpus   = 2
  memory = 1024

  network {
    name               = vcd_vapp_org_network.direct.org_network_name
    type               = "org"
    ip_allocation_mode = "POOL"
  }

  guest_properties = {
    "instance-id" = var.guest_hostname
    "hostname"    = var.guest_hostname
    "public-keys" = var.guest-ssh-public-key
    "user-data"   = base64encode(file("script.sh"))
  }
}
```