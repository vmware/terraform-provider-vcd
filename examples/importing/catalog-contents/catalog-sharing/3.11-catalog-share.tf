
# This is the user that will use the shared catalog in "../catalog-sharing" example
resource "vcd_org_user" "cat-author1" {
  org            = var.org
  name           = var.new_user
  password       = var.new_password
  role           = "Catalog Author"
  take_ownership = false
}

# New catalog to be created
resource "vcd_catalog" "catalog-Org1" {
  org              = var.org
  name             = var.catalog
  delete_force     = true
  delete_recursive = true
}

# 3 vApp templates to be upload into the new catalog
resource "vcd_catalog_vapp_template" "sample_vt" {
  count      = 3
  org        = var.org
  catalog_id = vcd_catalog.catalog-Org1.id

  name              = "sample-vt-${count.index}"
  description       = "Sample vapp template sample-vt-${count.index}"
  ova_path          = "${var.ova_path}/test_vapp_template.ova"
  upload_piece_size = 5
}

# 3 media items to be upload into the new catalog
resource "vcd_catalog_media" "sample_media" {
  count      = 3
  org        = var.org
  catalog_id = vcd_catalog.catalog-Org1.id

  name              = "sample-media-${count.index}"
  description       = "Sample media item sample-media-${count.index}"
  media_path        = "${var.ova_path}/test.iso"
  upload_piece_size = 5
}

# The catalog will be shared with the user cat-author1
resource "vcd_catalog_access_control" "AC-Catalog" {
  org        = var.org
  catalog_id = vcd_catalog.catalog-Org1.id

  shared_with_everyone = false

  shared_with {
    user_id      = vcd_org_user.cat-author1.id
    access_level = "FullControl"
  }
}
