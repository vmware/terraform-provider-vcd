module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/vmware/go-vcloud-director/v2 v2.10.0-alpha.4
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/Didainius/go-vcloud-director/v2 v2.9.1-0.20201126153022-0f4d1019926b
