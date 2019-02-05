Terraform Provider
==================

NOTE: We are in the process of planning v2 of the provider which will not be backwards compatible. See the [v2Plan](https://github.com/terraform-providers/terraform-provider-vcd/blob/master/v2Plan.md) for further details.

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.10 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-vcd`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-vcd
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-vcd
$ make build
```


Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.10+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-vcd
...
```

See TESTING.md for details on how to test.

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