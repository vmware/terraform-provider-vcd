module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/terraform-plugin-sdk v1.3.0
	github.com/vmware/go-vcloud-director/v2 v2.6.0-alpha.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.6.0-alpha.1.0.20200117164355-f6409474a9e9
