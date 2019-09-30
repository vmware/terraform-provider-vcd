module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/terraform v0.12.8
	github.com/vmware/go-vcloud-director/v2 v2.4.0-alpha.10
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.4.0-alpha-2.0.20190930093433-8e9b0560a763
