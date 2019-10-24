module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.13

require (
	github.com/hashicorp/terraform v0.12.8
	github.com/vmware/go-vcloud-director/v2 v2.4.0-rc.2
)

replace github.com/vmware/go-vcloud-director/v2 => github.com/vbauzysvmware/go-vcloud-director/v2 v2.0.0-20191024101442-d15926123585
