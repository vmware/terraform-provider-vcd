module github.com/vmware/terraform-provider-vcd/v3

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.12 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/vmware/go-vcloud-director/v2 v2.10.0-alpha.2
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/dataclouder/go-vcloud-director/v2 v2.5.0-alpha.4.0.20201118110526-71c1863d0e59
