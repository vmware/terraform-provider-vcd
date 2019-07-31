module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.12.0
	github.com/vmware/go-vcloud-director/v2 v2.3.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.3.2-0.20190731113052-87a5af498696
