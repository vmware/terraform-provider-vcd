# Subscribing to a published catalog

This example shows how to subscribe to a catalog that has been published in a different VCD.

**IMPORTANT**: before using this example, you need to run the corresponding "publisher" example, or have a published catalog already available.

Before running `terraform apply`, you need to modify `terraform.tfvars` to use the right VCD and credentials. Also, you need to fill the variable `publishing_url` with the one obtained from the "publisher" example.

After `terraform apply` runs successfully, you need to run `terraform refresh` to see the contents updated. You should see the same number of vApp templates and media items as for the "publisher" example.

At this point, you can remove the comment before and after the data sources, and run another `terraform apply`: this will
create data sources out of some subscribed catalog items, vApp templates, and media items.
