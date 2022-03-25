module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.15.0-alpha.8
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/mikeletux/go-vcloud-director/v2 v2.15.0-alpha.1.0.20220325091848-7d836411215a
