module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/terraform-plugin-sdk v1.8.0
	github.com/vmware/go-vcloud-director/v2 v2.8.0-alpha.3
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/dataclouder/go-vcloud-director/v2 v2.8.0-alpha.4.0.20200505180518-fb2a507882d9

// replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
