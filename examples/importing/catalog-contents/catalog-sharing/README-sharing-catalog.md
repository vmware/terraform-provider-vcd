# Importing shared catalog items

This example creates a catalog, and populates it with several items.
It is the pre-requisite to run the actual importing in `../catalog-using`.


## Steps

### Phase 1 - resource creation
1.1. Rename `terraform.tfvars.example` to `terraform.tfvars` and modify it to use the right VCD, credentials, and the path where the OVA and ISO can be found.
   The sample OVA and ISO used in this script is in the directory `./test-resources` at the top of this repository.
1.2. Run `terraform apply`: it will  create the catalog and upload several vApp templates and media items. It will also create a new user.

### Phase 2 - import

2.1. Change directory to `../catalog-using`, and run the example using the user just created here.

