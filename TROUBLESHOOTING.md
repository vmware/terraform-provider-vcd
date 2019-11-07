
## Upgrading Org resources to 2.5

If you have resources that were created with earlier versions, in rare cases they may not work correctly in 2.5+, due to
a few bugs in the handling of the resource ID and the default values for VM quotas.

Running a plan on such resource, terraform would want to re-deploy the resource, which is a consequence of the bug fix
that now gives the correct ID to the resource.

In this scenario, the safest approach is to remove the resource from terraform state and import it, using these steps.
Let's assume your org `my-org` was created in 2.4.

1. `terraform state list` (it will show `vcd_org.my-org`)
2. `terraform state rm vcd_org.my-org`
3. `terraform import vcd_org.my-org my-org`

At this point, the org will have the correct information.

## Upgrading Catalog and Catalog Items to 2.5

Similar to the Org situation, the catalog and catalog item resources have changed their internal ID from storing their name to storing their ID.
If, during a `terraform plan`, you see that the resource should be created again, you can import the entity again. For example, with a catalog mycat
that was created in 2.4:

1. `terraform state list` (it will show `vcd_catalog.mycat`)
2. `terraform state rm vcd_catalog.mycat`
3. `terraform import vcd_catalog.mycat my-org.mycat`

## Upgrading Network resources to 2.5

In a similar way, the resources `network_routed`, `network_isolated`, and `network_direct` may show some surprise when
upgrading from 2.4 to 2.5. In addition to using the ID as resource identifier, instead of the name, the new version
provides more details that were previously hidden. Consequently, you may see a request of re-creation for such
resources.

A possible solution is to delete the resource from the state file and import it using the new plugin:

1. `terraform state list` (it will show `vcd_network_routed.my-net`)
2. `terraform state rm vcd_network_routed.my-net`
3. `terraform import vcd_network_routed.mynet my-org.my-vdc.my-net`


## Migrating to new NSX-V API based NAT (vcd_nsxv_dnat and vcd_nsxv_snat) and firewall (vcd_nsxv_firewall_rule) resources

Version 2.5 introduced new resources for NAT and firewall rules. They only work with advanced edge
gateways (NSX-V). Previous resources (vcd_dnat, vcd_snat and vcd_firewall_rules) used vCD API
which is not recommended anymore as it may cause configuration problems and lacks features. It is
recommended to start using newer resources to avoid problems and access new features (including data
sources).

### For migrating DNAT and SNAT rules similar operation to the one above can be used:

1. `terraform state list` (it will show `vcd_dnat.my-dnat`, `vcd_snat.my-snat`)
2. `terraform state rm vcd_dnat.my-dnat`
3. `terraform import vcd_nsxv_dnat.my-dnat my-org-name.my-vdc-name.my-edge-gw-name.my-dnat-rule-id`

Note. NAT rule ID can be checked in the UI

### For migrating firewall rules 
For migrating firewall rules there should be one resource per one firewall rule. Similar process as
for NAT can be used, but firewall rules do not show real ID (only their number) in the UI. For this
reason there is an extra `list@` helper in the import command. Read more about it in `import` section
of the [`vcd_nsxv_firewall_rule`](https://www.terraform.io/docs/providers/vcd/r/nsxv_firewall_rule.html#importing)
resource.
