module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.12.6
	github.com/vmware/go-vcloud-director/v2 v2.4.0-alpha.5
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.4.0-alpha.3.0.20190826091546-05025c24e752

//replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
