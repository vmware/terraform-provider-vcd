
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
