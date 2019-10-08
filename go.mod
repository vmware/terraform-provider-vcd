module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/terraform v0.12.8
	github.com/vmware/go-vcloud-director/v2 v2.4.0-alpha.11
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/vbauzysvmware/go-vcloud-director/v2 v2.2.0-alpha.3.0.20191007135458-9117fe79dc45
