module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.12.0
	github.com/vmware/go-vcloud-director/v2 v2.4.0-alpha-2
)

// replace github.com/vmware/go-vcloud-director/v2 => github.com/dataclouder/go-vcloud-director/v2 v2.4.0-alpha-3

// replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
