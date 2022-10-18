* VM Creation code refactored which should result in identifiable parts creation for all types of
  VMs. Behind the scenes, there are 4 different types of VMs with respective different API calls as
  listed below [GH-901]
  * `vcd_vapp_vm` built from vApp template 
  * `vcd_vm` built from vApp template 
  * `vcd_vapp_vm` built without vApp template (empty VM) 
  * `vcd_vm` built without vApp template (empty VM) 
