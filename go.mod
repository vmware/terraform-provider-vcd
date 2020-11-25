module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.1.0
	github.com/vmware/go-vcloud-director/v2 v2.10.0-alpha.2
)

replace github.com/vmware/go-vcloud-director/v2 => ../go-vcloud-director
