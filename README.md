Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

![Terraform](https://rawgithub.com/hashicorp/terraform/master/website/source/assets/images/logo-hashicorp.svg)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-$PROVIDER_NAME`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-$PROVIDER_NAME
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-$PROVIDER_NAME
$ make build
```

Using the provider
----------------------
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-$PROVIDER_NAME
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

The acceptance tests will run against your own vCloud Air setup, using the enviroment variables you should have set below:

```sh
export VCD_USER=***********
export VCD_PASSWORD=************
export VCD_ORG=**********
export VCD_EXTERNAL_IP=xxx.xxx.xxx.xxx
export VCD_URL=https://api.vcd.xxxxxxxx.xxxxxxxx.com/api
export VCD_EDGE_GATEWAY=xxxxxxxxx
export VCD_VDC="xxxxxxxx"
```

Pulling in the 'Go vCloud Air' (govcloudair) Library
--------------------------------------------------------

If you've also been adding in functionality to the govcloudair library, you need to ensure your changes have been merged and then run:

```sh
$ make fetchvendor
```

This will pull in the latest version, including your changes, and you will be able to develop against the updated library.


