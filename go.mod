module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.11.13
	github.com/vmware/go-vcloud-director/v2 v2.2.0-alpha.3
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.2.0-alpha.1.0.20190503105237-3e428d5a3843
