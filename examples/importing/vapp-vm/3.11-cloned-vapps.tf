
# The first vApp is built from a vApp template
resource "vcd_cloned_vapp" "vapp_from_template" {
  org           = var.org
  vdc           = var.vdc
  name          = "ClonedVAppFromTemplate"
  description   = "Example cloned vApp from Template"
  power_on      = true
  source_id     = vcd_catalog_vapp_template.multi-vm-template.id
  source_type   = "template"
  delete_source = false
}

# The second vApp is cloned from the first vApp
resource "vcd_cloned_vapp" "vapp_from_vapp" {
  org           = var.org
  vdc           = var.vdc
  name          = "ClonedVAppFromVapp"
  description   = "Example cloned vApp from Template"
  power_on      = true
  source_id     = vcd_cloned_vapp.vapp_from_template.id
  source_type   = "vapp"
  delete_source = false
}

# This data source creates the import blocks for the vApps
data "vcd_resource_list" "import_vapps" {
  resource_type    = "vcd_vapp"
  name             = "all_vapps"
  parent           = var.vdc
  list_mode        = "import"
  import_file_name = "import-vapps.tf"
  # We want to avoid creating import blocks for all the vApps in the VDC: thus we filter for names of the vApps just created.
  name_regex = "(ClonedVAppFromTemplate|ClonedVAppFromVapp)"
  # We need to force a dependency, as this data source should be able to find the recently created vApps
  depends_on = [vcd_cloned_vapp.vapp_from_vapp, vcd_cloned_vapp.vapp_from_template]
}

# This data source creates the import blocks for the VMs in the first vApp
data "vcd_resource_list" "import_vms_from_template" {
  resource_type    = "vcd_vapp_vm"
  name             = "import_vms_from_template"
  parent           = vcd_cloned_vapp.vapp_from_template.name
  list_mode        = "import"
  import_file_name = "import-vms_from_template.tf"
}

# This data source creates the import blocks for the VMs in the second vApp
data "vcd_resource_list" "import_vms_from_vapp" {
  resource_type    = "vcd_vapp_vm"
  name             = "import_vms_from_vapp"
  parent           = vcd_cloned_vapp.vapp_from_vapp.name
  list_mode        = "import"
  import_file_name = "import-vms_from_vapp.tf"
}

output "import_vapps" {
  value = data.vcd_resource_list.import_vapps.list
}

output "import_vms_from_template" {
  value = data.vcd_resource_list.import_vms_from_template.list
}

output "import_vms_from_vapp" {
  value = data.vcd_resource_list.import_vms_from_vapp.list
}
