* Added `prefix_length` field to `resourceVcdVappNetwork` to support creating IPv6 vApp networks [GH-1007]
* Deprecated `netmask` in `resourceVcdVappNetwork` [GH-1007]
* Removed `netmask` field's `Default` value which, if not provided before, will result in a Terraform error. The user would then need to add a `"netmask" = "255.255.255.0"` to their existing vApp networks [GH-1007] 
