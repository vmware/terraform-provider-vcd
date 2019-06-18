
resource "vcd_external_network" "{{.ExternalNetwork}}" {
  name        = "{{.ExternalNetwork}}"
  description = "external net datacloud "

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.ExternalNetworkPortGroup}}"
    type    = "{{.ExternalNetworkPortGroupType}}"
  }

  ip_scope {
    gateway    = "{{.MainGateway}}"
    netmask    = "{{.MainNetmask}}"
    dns1       = "{{.MainDns1}}"
    dns2       = "{{.MainDns2}}"
    dns_suffix = "{{.Org}}.org"

    static_ip_pool {
      start_address = "{{.ExternalNetworkStaticStartIp}}"
      end_address   = "{{.ExternalNetworkStaticEndIp}}"
    }
  }
  retain_net_info_across_deployments = "false"
}


resource "vcd_org" "{{.Org}}" {
  name              = "{{.Org}}"
  full_name         = "{{.Org}}"
  is_enabled        = "true"
  stored_vm_quota   = 50
  deployed_vm_quota = 50
  delete_force      = "true"
  delete_recursive  = "true"
}

resource "vcd_org_vdc" "{{.Vdc}}" {
  name = "{{.Vdc}}"
  org  = "${vcd_org.{{.Org}}.name}"

  allocation_model  = "AllocationVApp"
  provider_vdc_name = "{{.ProviderVdc}}"
  network_pool_name = "{{.NetworkPool}}"
  network_quota     = 50

  compute_capacity {
    cpu {
      limit = 0
    }

    memory {
      limit = 0
    }
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    enabled = true
    limit   = 0
    default = true
  }

#_SECOND_STORAGE_PROFILE_

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}


resource "vcd_edgegateway" "{{.EdgeGateway}}" {
  org                     = "${vcd_org.{{.Org}}.name}"
  vdc                     = "${vcd_org_vdc.{{.Vdc}}.name}"
  name                    = "{{.EdgeGateway}}"
  description             = "{{.Org}} edge gateway"
  configuration           = "compact"
  default_gateway_network = "${vcd_external_network.{{.ExternalNetwork}}.name}"
  advanced                = true

  external_networks = ["${vcd_external_network.{{.ExternalNetwork}}.name}"]
}

resource "vcd_catalog" "{{.Catalog}}" {
  org         = "${vcd_org.{{.Org}}.name}"
  name        = "{{.Catalog}}"
  description = "{{.Org}} catalog"

  delete_force     = "true"
  delete_recursive = "true"
  depends_on       = ["vcd_org_vdc.{{.Vdc}}"]
}

resource "vcd_catalog_item" "{{.CatalogItem}}" {
  org     = "${vcd_org.{{.Org}}.name}"
  catalog = "${vcd_catalog.{{.Catalog}}.name}"

  name                 = "{{.CatalogItem}}"
  description          = "{{.CatalogItem}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = 5
  show_upload_progress = "true"
}

resource "vcd_catalog_media" "{{.MediaTestName}}" {
  org     = "${vcd_org.{{.Org}}.name}"
  catalog = "${vcd_catalog.{{.Catalog}}.name}"

  name                 = "{{.MediaTestName}}"
  description          = "{{.MediaTestName}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = 5
  show_upload_progress = "true"
}

# Optional networks will be added only if the
# corresponding names are set in the configuration file

