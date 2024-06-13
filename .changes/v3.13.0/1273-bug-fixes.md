* Fix `vcd_catalog_media` resource so it doesn't wait indefinitely to the upload task to reach 100% progress,
  by checking also its status, to decide that the upload is complete or aborted [GH-1273]