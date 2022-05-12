module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/hashicorp/go-version v1.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.14.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.3
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/mikeletux/go-vcloud-director/v2 v2.16.0-alpha.1.0.20220512090507-5d5c4bca93c3
