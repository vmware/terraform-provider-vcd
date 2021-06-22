module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.4
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.12.0-alpha.6
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.12.0-alpha.6.0.20210622123245-528082b16c9d
