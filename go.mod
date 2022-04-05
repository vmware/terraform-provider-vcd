module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/kr/pretty v0.2.1
	github.com/vmware/go-vcloud-director/v2 v2.15.0-alpha.10
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.15.0-alpha.7.0.20220405101106-8aeae0e28cc9
