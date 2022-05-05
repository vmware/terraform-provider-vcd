module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.16.0-alpha.1
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/mikeletux/go-vcloud-director/v2 v2.16.0-alpha.1.0.20220505133738-465d3803d077
