
provider "vcd" {
  user                 = "administrator"
  password             = "ca$hc0w"
  org                  = "au"
  url                  = "https://bos1-vcd-sp-static-200-22.eng.vmware.com/api"
  vdc                  = "au-vdc"
  max_retry_timeout    = "60"
  allow_unverified_ssl = "true"
}

resource "vcd_vapp" "test-tf-2" {
  name          = "test-tf-2"
}

resource "vcd_vapp_vm" "web" {
	vapp_name = "test-tf-2"
	name = "my-first-vm"
	catalog_name = "Re"
	template_name = "phot"
	network_name = "testNetwork"
  ip = "10.10.0.153"
  accept_all_eulas = "true"

	memory = 1024
	cpus = 1

}

resource "vcd_org" "test"{
  name = "test"
  full_name = "test"
  is_enabled = "true"
  network_name = "testNetwork"
}
