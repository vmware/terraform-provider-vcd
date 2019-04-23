module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.11.13
	github.com/kr/pretty v0.1.0
	github.com/vmware/go-vcloud-director/v2 v2.2.0-alpha.2
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.2.0-alpha.2.0.20190423125913-051bb7c0f61c
