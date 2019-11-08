module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/vmware/go-vcloud-director/v2 v2.5.0-alpha.2
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.5.0-alpha.2.0.20191108124622-a7699d7679d0
