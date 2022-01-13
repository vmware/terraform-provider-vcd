module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/kr/pretty v0.2.1
	github.com/vbauzysvmware/go-vcloud-director v2.0.0+incompatible // indirect
	github.com/vmware/go-vcloud-director/v2 v2.14.0
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/vbauzysvmware/go-vcloud-director/v2 v2.0.0-20220113111923-6cd0742957ad
