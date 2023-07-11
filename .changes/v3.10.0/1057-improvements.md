* Add `ignore_metadata_changes` argument to the Provider configuration to be able to specify metadata entries that should **not**
  be managed by Terraform when using `metadata_entry` configuration blocks [GH-1057]
* Add `ignore_metadata_changes_error_level` argument to the Provider configuration to be able to choose what to do if
  a `metadata_entry` managed by Terraform matches with the criteria specified in `ignore_metadata_changes` blocks [GH-1057]
