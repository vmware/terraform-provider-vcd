# Importing shared catalog items

This example shows how to import a shared catalog and its contents using [Terraform experimental import blocks](https://developer.hashicorp.com/terraform/language/import)


## Steps

### Phase 1 - create

1.1. **PREREQUISITE:** before running this example, you must run the one in `../catalog-sharing`

### Phase 2 - code generation

2.1. Run `terraform init` and `terraform apply`
2.2. Check the current directory: three new files were created: `import-catalog.tf`, `import-vapp-templates.tf`, and `import-media.tf`.
   These files contain the import blocks for the new resources.
2.3. Run `terraform plan -generate-config-out=generated_resources.tf`: this will create a HCL file containing the definition
   of the resources we need to import.

### Phase 3 - import

At this stage, we can run the import, and deal with the newly imported resources without fear of interference.

3.1. Run `terraform apply`: it will import the resources and update them.

