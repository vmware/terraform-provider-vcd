* **New Resource:** `vcd_nsxt_distributed_firewall_rule` to manage NSX-T Distributed Firewall one by
  one. Rules will *not be created in parallel* because the API provides no direct endpoint to create
  a single rule and this functionality uses a custom-made function that abstracts the "update all"
  endpoint. [GH-1076]
* **New Data Source:** `vcd_nsxt_distributed_firewall_rule` to read NSX-T Distributed Firewall one
  by one [GH-1076]