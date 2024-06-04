* Resource and data source `vcd_nsxt_edgegateway` add new field `read_limit_unused_ip_count` that
  can limit number of IPs to count for `used_ip_count` and `unused_ip_count` as it may exhaust
  compute resources (issue [#1270](https://github.com/vmware/terraform-provider-vcd/issues/1270))
  [GH-1275]
