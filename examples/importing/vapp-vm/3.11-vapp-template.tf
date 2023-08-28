
# Using an existing catalog
data "vcd_catalog" "cat" {
  org  = var.org
  name = var.catalog
}

# we load a vApp template containing 3 VMs into the catalog
resource "vcd_catalog_vapp_template" "multi-vm-template" {
  org        = var.org
  catalog_id = data.vcd_catalog.cat.id

  name              = "small-3VM"
  description       = "vApp template with multiple VMs"
  ova_path          = "${var.ova_path}/vapp_with_3_vms.ova"
  upload_piece_size = 5
}
