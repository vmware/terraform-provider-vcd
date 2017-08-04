## 0.1.3 (Unreleased)

IMPROVEMENTS:

* `vcd_vapp` - Setting the computer name regardless of init script [GH-31]
* `vcd_vapp` - Fixes the power_on issue introduced in 0.1.2 [GH-33]
* `vcd_vapp_vm` - Setting the computer name regardless of init script [GH-31]


## 0.1.2 (August 03, 2017)

IMPROVEMENTS:

* Possibility to add OVF parameters to a vApp ([#1](https://github.com/terraform-providers/terraform-provider-vcd/pull/1))
* Added storage profile support  ([#23](https://github.com/terraform-providers/terraform-provider-vcd/pull/23))

## 0.1.1 (June 28, 2017)

FEATURES:

* **New VM Resource**: `vcd_vapp_vm` ([#9](https://github.com/terraform-providers/terraform-provider-vcd/issues/9))
* **New VPN Resource**: `vcd_edgegateway_vpn`

IMPROVEMENTS:

* resource/vcd_dnat: Added a new (optional) param translated_port ([#14](https://github.com/terraform-providers/terraform-provider-vcd/issues/14))

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
