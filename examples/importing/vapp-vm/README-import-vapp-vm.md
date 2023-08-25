# Importing cloned vApps with their VMs

This example shows how to import two vApps and their VMs using [Terraform experimental import blocks](https://developer.hashicorp.com/terraform/language/import)


## Steps

### Phase 1 - resource creation
1.1. Rename `terraform.tfvars.example` to `terraform.tfvars` and modify it to use the right VCD, credentials, and the path where the OVA can be found.
   The sample OVA used in this script is in the directory `./test-resources` at the top of this repository.
1.2. Run `terraform apply`: it will upload a vApp template, and create two cloned vApps, each with three VMs.

### Phase 2 - code generation

2.1. Check the current directory: three new files were created: `import_vms_from_vapp.tf`, `import_vms_from_template.tf`, and `import_vapps.tf`.
   These files contain the import blocks for the new resources.
2.2. Run `terraform plan -generate-config-out=generated_resources.tf`: this will create a HCL file containing the definition
   of the resources we need to import.

### Phase 3 - import

At this stage, we can move the generated files to a new directory, run the import, and deal with the newly imported
resources without fear of interference

3.1. Make a directory `ops`
3.2. copy `3.11-provider.tf`, `versions.tf`, `terraform.tfvars`, `generated_resources.tf`, and all the `import*.tf` into `ops`
3.3. Change directory to `ops`
3.4. Run `terraform init` and `terraform apply`: it will import the resources and update them.

