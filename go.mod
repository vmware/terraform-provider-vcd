module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/hashicorp/go-version v1.5.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.17.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.7
)

replace github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.7 => github.com/mikeletux/go-vcloud-director/v2 v2.16.0-alpha.1.0.20220609102706-31e5d2910eb6
