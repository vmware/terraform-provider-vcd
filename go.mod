module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.15.0-alpha.2
)

//replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.14.0-rc.3.0.20220224195422-2a5f827ef552

//replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
//replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.12.1-0.20211018060826-c7f8ab32330e
replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
