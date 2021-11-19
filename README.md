Terraform VMware Cloud Director Provider
==================

The official Terraform provider for [VMware Cloud Director](https://www.vmware.com/products/cloud-director.html)

- Documentation of the latest binary release available at https://registry.terraform.io/providers/vmware/vcd/latest/docs
- This project is using [go-vcloud-director](https://github.com/vmware/go-vcloud-director) Golang SDK for making API calls to vCD
- Join through [VMware {code}](https://code.vmware.com/) to [![Chat](https://img.shields.io/badge/chat-on%20slack-brightgreen.svg)](https://vmwarecode.slack.com/messages/CBBBXVB16) in #vcd-terraform-dev channel 

Part of Terraform
-----------------

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
- [Hashicorp Discuss](https://discuss.hashicorp.com/c/terraform-core/27) 

<img src="https://www.datocms-assets.com/2885/1629941242-logo-terraform-main.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html)
-	[Go](https://golang.org/doc/install) 1.14 (to build the provider plugin)

Building The Provider (the modules way)
--------------------------------------
**Note.** You *only* need to build the provider plugin if you want to *develop* it. Refer to
[documentation](https://registry.terraform.io/providers/vmware/vcd/latest/docs) for using it. Terraform will
automatically download officially released binaries of this provider plugin on the first run of `terraform init`
command.

Starting with version 2.1 provider started using [Go modules](https://github.com/golang/go/wiki/Modules)
This means that it is no longer necessary to be in GOPATH.
[See more](https://github.com/golang/go/wiki/Modules#how-to-use-modules) on how to use modules
and toggle between modes.

```
$ cd ~/mydir
$ git clone https://github.com/vmware/terraform-provider-vcd.git
$ cd terraform-provider-vcd/
$ make build
```

Developing the Provider
---------------------------

Starting with terraform-provider-vcd version 2.1 Go modules are used. This means a few things:
* The code no longer needs to stay in your `GOPATH`. It can though -
[see more](https://github.com/golang/go/wiki/Modules#how-to-use-modules) on how to use modules and toggle between modes.
* When developing `terraform-provider-vcd` one often needs to add extra stuff to `go-vcloud-director`. Go modules
have a convenient [replace](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive)
directive which can allow you to redirect import path to your own version of `go-vcloud-director`.
`go.mod` can be altered:
 * You can replace your import with a forked branch like this:
 ```go
    module github.com/vmware/terraform-provider-vcd/v2
    require (
    	...
    	github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2
    	)
    replace github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2 => github.com/my-git-user/go-vcloud-director/v2 v2.1.0-alpha.2    
 ```
 * You can also replace pointer to a branch with relative directory
 ```go
     module github.com/vmware/terraform-provider-vcd/v2
     require (
     	...
     	github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2
     	)
     replace github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2 => ../go-vcloud-director
  ```

See **CODING_GUIDELINES.md** for more advice on how to write code for this project.

Using the provider
----------------------

### Installing the built provider

For a more thorough test using the Terraform client, you may want to transfer the plugin in the Terraform directory. A `make` command can do this for you:

```sh
$ make install
```

This command will build the plugin and transfer it to `$HOME/.terraform.d/plugins`, with a name that includes the version (as taken from the `./VERSION` file).

Starting with terraform 0.13, the path where the plugin is deployed is
```
`$HOME/.terraform.d/plugins/registry.terraform.io/vmware/vcd/${VERSION}/${OS}_amd64/terraform-provider-vcd_v${VERSION}`
```

For example, on MacOS:

```
$HOME/.terraform.d/
├── checkpoint_signature
└── plugins
    ├── registry.terraform.io
    └── vmware
        └── vcd
            ├── 2.9.0
            │   └── darwin_amd64
            │       └── terraform-provider-vcd_v2.9.0
            └── 3.0.0
                └── darwin_amd64
                    └── terraform-provider-vcd_v3.0.0
```

On Linux:

```
$HOME/.terraform.d/
├── checkpoint_signature
└── plugins
    ├── registry.terraform.io
    └── vmware
        └── vcd
            ├── 2.9.0
            │   └── linux_amd64
            │       └── terraform-provider-vcd_v2.9.0
            └── 3.0.0
                └── linux_amd64
                    └── terraform-provider-vcd_v3.0.0
```


### Using the new plugin

Once you have installed the plugin as mentioned above, you can simply create a new `config.tf` as defined in [the manual](https://www.terraform.io/docs/providers/vcd/index.html) and run 

```sh
$ terraform init
$ terraform plan
$ terraform apply
```

When using terraform 0.13+, you also need to have a `terraform` block either in your script or in an adjacent `versions.tf` file,
containing.

```
terraform {
  required_providers {
    vcd = {
      source = "vmware/vcd"
    }
  }
  required_version = ">= 0.13"
}
```

In this block, the `vmware` part of the source corresponds to the directory
`$HOME/.terraform.d/plugins/registry.terraform.io/vmware` created by the command `make install`.

Note that `versions.tf` is generated when you run the `terraform 0.13upgrade` command. If you have run such command,
you need to edit the file and make sure the **`source`** path corresponds to the one installed, or remove the file
altogether if you have already the right block in your script.