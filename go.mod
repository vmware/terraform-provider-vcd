module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0 // indirect
	github.com/aws/aws-sdk-go v1.25.3 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/vmihailenco/msgpack v4.0.1+incompatible // indirect
	github.com/vmware/go-vcloud-director/v2 v2.4.0
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb // indirect
	golang.org/x/sys v0.0.0-20190804053845-51ab0e2deafa // indirect
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/dataclouder/go-vcloud-director/v2 v2.5.0-alpha.3
