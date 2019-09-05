module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/hashicorp/terraform v0.12.6
	github.com/vmware/go-vcloud-director/v2 v2.4.0-alpha.6
)

// Temporary workaround for git.apache.org outage
replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
