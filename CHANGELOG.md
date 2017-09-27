## 1.0.1 (Unreleased)

IMPROVEMENTS:

* `vcd_vapp` - Fixes an issue with Networks in vApp templates being required, also introduced in 0.1.2 ([#38](https://github.com/terraform-providers/terraform-provider-vcd/issues/38))

FEATURES:

* `vcd_vapp` - Add support for defining shared vcd_networks ([#46](https://github.com/terraform-providers/terraform-provider-vcd/pull/46))
* `vcd_vapp` - Added options to configure dhcp lease times ([#47](https://github.com/terraform-providers/terraform-provider-vcd/pull/47))


## 1.0.0 (August 17, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Fixes an issue with storage profiles introduced in 0.1.2 ([#39](https://github.com/terraform-providers/terraform-provider-vcd/issues/39))

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: Deprecate `maxRetryTimeout` in favour of `max_retry_timeout` ([#40](https://github.com/terraform-providers/terraform-provider-vcd/issues/40))

## 0.1.3 (August 09, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Setting the computer name regardless of init script ([#31](https://github.com/terraform-providers/terraform-provider-vcd/issues/31))
* `vcd_vapp` - Fixes the power_on issue introduced in 0.1.2 ([#33](https://github.com/terraform-providers/terraform-provider-vcd/issues/33))
* `vcd_vapp` - Fixes issue with allocated IP address affecting tf plan ([#17](https://github.com/terraform-providers/terraform-provider-vcd/issues/17) & [#29](https://github.com/terraform-providers/terraform-provider-vcd/issues/29))
* `vcd_vapp_vm` - Setting the computer name regardless of init script ([#31](https://github.com/terraform-providers/terraform-provider-vcd/issues/31))
* `vcd_firewall_rules` - corrected typo in docs ([#35](https://github.com/terraform-providers/terraform-provider-vcd/issues/35))


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
