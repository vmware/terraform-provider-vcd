Terraform vCloud Director Provider
==================

The official Terraform provider for [VMware vCloud Director](https://www.vmware.com/products/vcloud-director.html)

- Documentation of the latest binary release available at https://www.terraform.io/docs/providers/vcd/index.html
- This project is using [go-vcloud-director](https://github.com/vmware/go-vcloud-director) Golang SDK for making API calls to vCD
- Join through [VMware {code}](https://code.vmware.com/) to [![Chat](https://img.shields.io/badge/chat-on%20slack-brightgreen.svg)](https://vmwarecode.slack.com/messages/CBBBXVB16) in #vcd-terraform-dev channel 

Part of Terraform
-----------------

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.13
-	[Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Building The Provider (the modules way)
--------------------------------------

Starting with version 2.1 provider started using [Go modules](https://github.com/golang/go/wiki/Modules)
This means that it is no longer necessary to be in GOPATH.
[See more](https://github.com/golang/go/wiki/Modules#how-to-use-modules) on how to use modules
and toggle between modes.

```
$ cd ~/mydir
$ git clone https://github.com/terraform-providers/terraform-provider-vcd.git
$ cd terraform-provider-vcd/
$ make build
```

Building The Provider (the old [vendor](https://golang.org/cmd/go/#hdr-Vendor_Directories) way)
--------------------------------------

Prior to version 2.1 provider used Go vendor directory for dependency management. This method is not recommended
anymore, but can be used to build provider on Go versions < 1.11.

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-vcd`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone https://github.com/terraform-providers/terraform-provider-vcd.git
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-vcd
$ make build
```


Developing the Provider
---------------------------

Starting with terraform-provider-vcd version 2.1 Go modules are used, while `vendor` directory is left for backwards
compatibility only. This means a few things:
* The code no longer needs to stay in your `GOPATH`. It can though -
[see more](https://github.com/golang/go/wiki/Modules#how-to-use-modules) on how to use modules and toggle between modes.
* `vendor` directory is __not to be changed manually__. Always use Go modules when introducing new dependencies
and always rebuild the vendor directory using `go mod vendor` if you have changed `go.mod` or `go.sum`. Travis CI will
catch and fail if it is not done.
* When developing `terraform-provider-vcd` one often needs to add extra stuff to `go-vcloud-director`. Go modules
have a convenient [replace](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive)
directive which can allow you to redirect import path to your own version of `go-vcloud-director`.
`go.mod` can be altered:
 * You can replace your import with a forked branch like this:
 ```go
    module github.com/terraform-providers/terraform-provider-vcd/v2
    require (
    	...
    	github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2
    	)
    replace github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2 => github.com/my-git-user/go-vcloud-director/v2 v2.1.0-alpha.2    
 ```
 * You can also replace pointer to a branch with relative directory
 ```go
     module github.com/terraform-providers/terraform-provider-vcd/v2
     require (
     	...
     	github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2
     	)
     replace github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2 => ../go-vcloud-director
  ```

Using the provider
----------------------

### Installing the built provider

For a more thorough test using the Terraform client, you may want to transfer the plugin in the Terraform directory. A `make` command can do this for you:

```sh
$ make install
```

This command will build the plugin and transfer it to `$HOME/.terraform.d/plugins`, with a name that includes the version (as taken from the `./VERSION` file).

### Using the new plugin

Once you have installed the plugin as mentioned above, you can simply create a new `config.tf` as defined in [the manual](https://www.terraform.io/docs/providers/vcd/index.html) and run 

```sh
$ terraform init
$ terraform plan
$ terraform apply
```