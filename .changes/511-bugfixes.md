* `nat_enabled` and `firewall_enabled` were incorrectly added to `vcd_vapp_network` and would collide with the depending resources. 
Now moved to respective resources `vcd_vapp_nat_rules` and `vcd_vapp_firewall_rules`.
