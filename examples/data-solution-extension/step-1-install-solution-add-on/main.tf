provider "vcd" {
  user                 = var.vcd_admin
  password             = var.vcd_password
  auth_type            = "integrated"
  url                  = var.vcd_url
  sysorg               = var.vcd_sysorg
  org                  = var.vcd_solutions_org
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director.log"
}

resource "vcd_catalog" "solution_add_ons" {
  org = var.vcd_solutions_org

  name             = "solution_add_ons"
  description      = "Catalog hoss Data Solution Add-Ons"
  delete_recursive = true
  delete_force     = true
}

data "vcd_org_vdc" "solutions_vdc" {
  org  = var.vcd_solutions_org
  name = var.vcd_solutions_vdc
}

data "vcd_network_routed_v2" "solutions" {
  org  = var.vcd_solutions_org
  vdc  = var.vcd_solutions_vdc
  name = var.vcd_solutions_vdc_routed_network
}

data "vcd_storage_profile" "solutions" {
  org  = var.vcd_solutions_org
  vdc  = var.vcd_solutions_vdc
  name = var.vcd_solutions_vdc_storage_profile_name
}

resource "vcd_solution_landing_zone" "slz" {
  org = var.vcd_solutions_org

  catalog {
    id = vcd_catalog.solution_add_ons.id
  }

  vdc {
    id         = data.vcd_org_vdc.solutions_vdc.id
    is_default = true

    org_vdc_network {
      id         = data.vcd_network_routed_v2.solutions.id
      is_default = true
    }

    compute_policy {
      id         = data.vcd_org_vdc.solutions_vdc.default_compute_policy_id
      is_default = true
    }

    storage_policy {
      id         = data.vcd_storage_profile.solutions.id
      is_default = true
    }
  }
}

resource "vcd_catalog_media" "dse14" {
  org        = var.vcd_solutions_org
  catalog_id = vcd_catalog.solution_add_ons.id

  name              = basename(var.vcd_dse_add_on_iso_path)
  description       = "DSE Solution Add-On"
  media_path        = var.vcd_dse_add_on_iso_path
  upload_piece_size = 10
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id        = vcd_catalog_media.dse14.catalog_item_id
  add_on_path            = var.vcd_dse_add_on_iso_path
  auto_trust_certificate = true

  depends_on = [vcd_solution_landing_zone.slz]
}

resource "vcd_solution_add_on_instance" "dse14" {
  add_on_id   = vcd_solution_add_on.dse14.id
  accept_eula = true
  name        = "dse14-instance"

  input = {
    delete-previous-uiplugin-versions = true
  }

  delete_input = {
    force-delete = true
  }
}

data "vcd_org" "dse-consumer" {
  name = var.vcd_tenant_org
}

resource "vcd_solution_add_on_instance_publish" "public" {
  add_on_instance_id     = vcd_solution_add_on_instance.dse14.id
  org_ids                = [data.vcd_org.dse-consumer.id]
  publish_to_all_tenants = false
}
