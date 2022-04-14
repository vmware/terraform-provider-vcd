* `vcd_org_user` resource and data source have now `is_external` attribute to support the importing of LDAP users into the Organization [GH-798]
* `vcd_org_user` resource does not have a default value for `deployed_vm_quota` and `stored_vm_quota`. Local users will have unlimited quota by default, imported from LDAP will have no quota [GH-798]
* `vcd_org_user` resource and data source have now `group_names` attribute to list group names if the user comes from an LDAP group [GH-798]
* `vcd_org_group` resource and data source have now `user_names` attribute to list user names if the user was imported from LDAP [GH-798]
