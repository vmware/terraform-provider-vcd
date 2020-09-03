module github.com/vmware/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/terraform-plugin-sdk v1.8.0
	github.com/vmware/go-vcloud-director/v2 v2.9.0-alpha.5
	github.com/vmware/terraform-provider-vcd/v3 v3.0.0-20200828124223-85100ed1058c
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/vbauzysvmware/go-vcloud-director/v2 v2.2.0-alpha.3.0.20200903112805-b4f59f9e898b
