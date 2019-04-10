module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.11.13
	github.com/vmware/go-vcloud-director/v2 v2.2.0-alpha.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/megalord/go-vcloud-director/v2 v2.0.0-20190410164604-f4845c5f159c
