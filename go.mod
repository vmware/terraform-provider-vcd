module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/hashicorp/go-version v1.5.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.17.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.9
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/adambarreiro/go-vcloud-director/v2 v2.15.0-alpha.1.0.20220628122003-68ecbad5a0b3
