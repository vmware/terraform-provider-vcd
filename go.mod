module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/hashicorp/go-version v1.5.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.17.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.7
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.15.0-alpha.7.0.20220603131433-5255108fe50c

//replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
