
# create import blocks for catalog
data "vcd_resource_list" "catalog" {
  org           = var.org
  name          = "list-catalog"
  resource_type = "vcd_catalog"
  list_mode     = "import"
  # We want to avoid creating import blocks for all the catalogs in the organization: thus we filter for names of the catalog that was just created.
  name_regex       = var.catalog
  import_file_name = "import-catalog.tf"
}

# create import blocks for vApp templates
data "vcd_resource_list" "templates" {
  org              = var.org
  name             = "list-templates"
  parent           = var.catalog
  resource_type    = "vcd_catalog_vapp_template"
  list_mode        = "import"
  import_file_name = "import-vapp-templates.tf"
}

# create import blocks for media items
data "vcd_resource_list" "media" {
  org              = var.org
  name             = "list-media"
  parent           = var.catalog
  resource_type    = "vcd_catalog_media"
  list_mode        = "import"
  import_file_name = "import-media.tf"
}
