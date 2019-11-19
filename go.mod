module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/hashicorp/terraform-plugin-sdk v1.3.0
	github.com/vmware/go-vcloud-director/v2 v2.5.0-alpha.6
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.5.0-alpha.4.0.20191119131702-2c1c2c2ae71c
