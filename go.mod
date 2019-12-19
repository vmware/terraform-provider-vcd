module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/terraform-plugin-sdk v1.3.0
	github.com/vmware/go-vcloud-director/v2 v2.5.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/vbauzysvmware/go-vcloud-director/v2 v2.2.0-alpha.3.0.20191219102849-f40088ca976a
