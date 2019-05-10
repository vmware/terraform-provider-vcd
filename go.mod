module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.11.13
	github.com/vmware/go-vcloud-director/v2 v2.2.0-beta.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.2.0-alpha.1.0.20190510093434-537d7f55c5a2
