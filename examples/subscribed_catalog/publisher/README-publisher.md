# Publishing a catalog

This example shows how to publish a catalog containing 10 vApp templates and 20 Media items.

Before running `terraform apply`, you need to modify `terraform.tfvars` to use the right VCD, credentials, and the path where the OVA and ISO can be found. (It is recommended to choose small items for this test)

After `terraform apply` runs successfully, you need to run `terraform refresh` to see the contents updated.

Please copy the output of `publishing_url` to use it in the subscriber example.
