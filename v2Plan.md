# Version 2 Plan::DRAFT

The idea is to overhaul how vApps interact with the networks, and the relationship between vApps and VMs. In vCloud Director 9, vApps become optional.

The current version of the provider tries to support v5.7 onwards. However, there are current issues with the implementation of the provider, which limits a vApp to only have 1 VM. This is visible with references of index zero VM's i.e. vApp.VM[0]. This means any action on the vApp is only applied to the first VM.

Many users of the provider, would like there to be multiple VM's per vApp, which is why the 'vcd_vapp_vm' resource was added. However, due to the way the vCloud Director API works, the only way to add a VM is to make a call vApp API. Adding multiple VM's with the the new vApp_vm resource is not efficient, as each API call is blocked by the previous call, until it has finished. 

A solution is to move the vapp_vm resource to sit as a sub-element of the vapp resource, so that Terraform can calculate all the VM resources before making the API call, of which only 1 call would be needed.

An issue with this solution, is that in vCloud Director 9 the vApp resource becomes optional rather than required. With lots of API changes between 8.2 and 9.0, it makes development of the provider very difficult with regard to design choices.

The purpose of this document is to help us decide how we are going to move forward. 

```
resource "vcd_dnat" "bar" {
	edge_gateway = "%s"
	external_ip = "%s"
	port = 7777
	internal_ip = "10.10.102.60"
	translated_port = 77
}
```

```
resource "vcd_vpn" "vpn" {
    edge_gateway        = "%s"
    name                = "west-to-east"
	description         = "Description"
	encryption_protocol = "AES256"
    mtu                 = 1400
    peer_id             = "51.179.218.226"
    peer_ip_address     = "51.179.218.226"
    local_id            = "51.179.218.225"
    local_ip_address    = "51.179.218.225"
    shared_secret       = "yZ4B8pxS5334m6ho692hjbtb7zo2vbesn7pe8ry5hyud86M433tbnnfxt6Dqn73g"
    
    peer_subnets {
        peer_subnet_name = "DMZ_WEST"
        peer_subnet_gateway = "10.0.10.1"
        peer_subnet_mask = "255.255.255.0"
    }

    peer_subnets {
        peer_subnet_name = "WEB_WEST"
        peer_subnet_gateway = "10.0.20.1"
        peer_subnet_mask = "255.255.255.0"
    }

    local_subnets {
        local_subnet_name = "DMZ_EAST"
        local_subnet_gateway = "10.0.1.1"
        local_subnet_mask = "255.255.255.0"
    }

    local_subnets {
        local_subnet_name = "WEB_EAST"
        local_subnet_gateway = "10.0.22.1"
        local_subnet_mask = "255.255.255.0"
    }
}
```

```
resource "vcd_firewall_rules" "bar" {
  edge_gateway = "%s"
	default_action = "%s"

	rule {
		description = "Test rule"
		policy = "allow"
		protocol = "any"
		destination_port = "any"
		destination_ip = "any"
		source_port = "any"
		source_ip = "any"
	}
}
```

```
resource "vcd_network" "foonet" {
	name = "foonet"
	edge_gateway = "%s"
	gateway = "10.10.102.1"
	static_ip_pool {
		start_address = "10.10.102.2"
		end_address = "10.10.102.254"
	}
}
```


```
resource "vcd_vapp" "foobar" {
  name = "foobar"

  networks = ["network1", "network2"]

  vm {
	  name          = "moo"
	  catalog_name  = "Skyscape Catalogue"
	  template_name = "Skyscape_CentOS_6_4_x64_50GB_Small_v1.0.1"
	  memory        = 1024
	  cpus          = 1
	  
	  network  {
	  	name = "network1"
	  	ip_allocation_mode = "dhcp"
	  }

	  network  {
	  	name = "network2"
	  	ip   = "10.10.102.161"
	  	ip_allocation_mode = "allocated"
	  }
  }
}
```