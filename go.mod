module github.com/terraform-providers/terraform-provider-vcd/v2

go 1.12

require (
	github.com/apparentlymart/go-cidr v0.0.0-20170418151526-7e4b007599d4
	github.com/apparentlymart/go-rundeck-api v0.0.0-20160826143032-f6af74d34d1e
	github.com/aws/aws-sdk-go v1.8.34
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d
	github.com/blang/semver v3.5.1+incompatible
	github.com/davecgh/go-spew v1.1.0
	github.com/fsouza/go-dockerclient v0.0.0-20160427172547-1d4f4ae73768
	github.com/go-ini/ini v1.23.1
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce
	github.com/hashicorp/go-cleanhttp v0.0.0-20170211013415-3573b8b52aa7
	github.com/hashicorp/go-getter v0.0.0-20170207215532-c3d66e76678d
	github.com/hashicorp/go-multierror v0.0.0-20150916205742-d30f09973e19
	github.com/hashicorp/go-plugin v0.0.0-20170217162722-f72692aebca2
	github.com/hashicorp/go-uuid v0.0.0-20160120003506-36289988d83c
	github.com/hashicorp/go-version v0.0.0-20161031182605-e96d38404026
	github.com/hashicorp/hcl v0.0.0-20170504190234-a4b07c25de5f
	github.com/hashicorp/hil v0.0.0-20170512213305-fac2259da677
	github.com/hashicorp/logutils v0.0.0-20150609070431-0dc08b1671f3
	github.com/hashicorp/terraform v0.10.0
	github.com/hashicorp/yamux v0.0.0-20160720233140-d1caa6c97c9f
	github.com/jmespath/go-jmespath v0.0.0-20160803190731-bd40a432e4c7
	github.com/mitchellh/copystructure v0.0.0-20161013195342-5af94aef99f5
	github.com/mitchellh/go-homedir v0.0.0-20161203194507-b8bc1bf76747
	github.com/mitchellh/hashstructure v0.0.0-20160209213820-6b17d669fac5
	github.com/mitchellh/mapstructure v0.0.0-20170307201123-53818660ed49
	github.com/mitchellh/reflectwalk v0.0.0-20161003174516-92573fe8d000
	github.com/rancher/go-rancher v0.0.0-20170407040943-ec24b7f12fca
	github.com/satori/go.uuid v0.0.0-20160927100844-b061729afc07
	github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2
	golang.org/x/crypto v0.0.0-20170808112155-b176d7def5d7
	golang.org/x/net v0.0.0-20170809000501-1c05540f6879
	k8s.io/kubernetes v1.6.1
)

replace github.com/vmware/go-vcloud-director/v2 v2.1.0-alpha.2 => github.com/Didainius/go-vcloud-director/v2 v2.1.0-alpha.2

