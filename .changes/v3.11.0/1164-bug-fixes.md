* Fix a bug in `ignore_metadata_changes` provider configuration block when `conflict_action = warn`, that caused
  an operation to fail immediately instead of continuing without an error when a conflict was found [GH-1164, GH-1173]
