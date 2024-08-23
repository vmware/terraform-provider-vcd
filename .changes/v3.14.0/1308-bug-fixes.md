* Fix [Issue 1307](https://github.com/vmware/terraform-provider-vcd/issues/1307) in `vcd_vapp_vm`
  and `vcd_vm` resources where `firmware=efi` field wouldn't be applied for template based
  VMs with `firmware=bios` on creation [GH-1308]
