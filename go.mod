module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.1.0
	github.com/vmware/go-vcloud-director/v2 v2.9.0
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/dataclouder/go-vcloud-director 6166a2c7a616f66d817e49e32527f861d2079b9d
